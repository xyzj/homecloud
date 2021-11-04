package main

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

//go:embed static
var stat embed.FS

type sliceFlag []string

func (f *sliceFlag) String() string {
	return fmt.Sprintf("%v", []string(*f))
}

func (f *sliceFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

var (
	port  = flag.Int("http", 8019, "set http port")
	ports = flag.Int("https", 0, "set https port")
	cert  = flag.String("cert", "", "cert file path")
	key   = flag.String("key", "", "key file path")
	auth  = flag.Bool("auth", false, "enable basicauth")
	debug = flag.Bool("debug", false, "run in debug mode")
	html  = flag.String("html", "", "something like nginx")
	// 帮助信息
	help = flag.Bool("help", false, "print help message and exit.")
	dirs sliceFlag
	wtv  sliceFlag
)

type byModTime []os.FileInfo

func (fis byModTime) Len() int {
	return len(fis)
}

func (fis byModTime) Swap(i, j int) {
	fis[i], fis[j] = fis[j], fis[i]
}

func (fis byModTime) Less(i, j int) bool {
	return fis[i].ModTime().After(fis[j].ModTime())
}

var (
	pageWebTV    string
	extreplacer  = strings.NewReplacer(".webp", "", ".dur", "", ".smi", "", ".png", "", ".jpg", "", ".zh-Hans", "", ".zh-Hant", "", ".vtt", "", ".en", "", ".en-US", "")
	namereplacer = strings.NewReplacer("#", "", "%", "")
	subTypes     = []string{".zh-Hant", ".zh-Hans", ".en", ".en-US"}
)

// smi2Vtt smi转vtt
func smi2Vtt(in, out string) error {
	defer os.Remove(in)
	bIn, err := ioutil.ReadFile(in)
	if err != nil {
		return err
	}
	ss := strings.Split(string(bIn), "\n")
	// 分析语言
	var language string
	for _, v := range ss {
		if strings.HasPrefix(v, "-->") {
			break
		}
		if strings.HasPrefix(v, ".zh") {
			language = v[1:strings.Index(v, " {")]
			if strings.Contains(v, "自动翻译") || language != "zh-Hant" {
				continue
			}
			break
		}
	}
	// 分析主体
	var bOut bytes.Buffer
	// bOut.WriteString("WEBVTT\r\n\r\n")
	var text string
	var tStart, tEnd int64
	for _, v := range ss {
		v = strings.TrimSpace(v)
		idxTime1 := strings.Index(v, "Start=")
		idxClass := strings.Index(v, "class=")
		if idxTime1 == -1 || idxClass == -1 || !strings.Contains(v, "class='"+language+"'") {
			continue
		}
		idxTime2 := strings.Index(v, ">")
		idxText := strings.LastIndex(v, ">")
		if v[idxText+1:] == "" {
			continue
		}
		if v[idxText+1:] == "&nbsp;" {
			tEnd, _ = strconv.ParseInt(v[idxTime1+6:idxTime2], 10, 0)
			// tEnd = gopsu.String2Int(v[idxTime1+6:idxTime2], 10)
		} else {
			tStart, _ = strconv.ParseInt(v[idxTime1+6:idxTime2], 10, 0)
			// tStart = gopsu.String2Int(v[idxTime1+6:idxTime2], 10)
			text = v[idxText+1:]
		}
		if tStart >= 0 && tEnd > 0 && len(text) > 0 {
			bOut.WriteString(fmt.Sprintf("%02d:%02d:%02d.%03d --> %02d:%02d:%02d.%03d\r\n",
				tStart/1000/60/60,
				tStart/1000/60%60,
				tStart/1000%60,
				tStart%1000,
				tEnd/1000/60/60,
				tEnd/1000/60%60,
				tEnd/1000%60,
				tEnd%1000))
			bOut.WriteString(text + "\r\n")
			tStart = -1
			tEnd = 0
			text = ""
		}
	}
	if bOut.Len() == 0 {
		return nil
	}
	return ioutil.WriteFile(out, []byte("WEBVTT\r\n\r\n"+bOut.String()), 0644)
}

func runVideojs(url, dst string) gin.HandlerFunc {
	srcdir := "/v-" + url + "/"
	b, _ := stat.ReadFile("static/tpl/webtv.html")
	webtpl := unsafeString(b)
	return func(c *gin.Context) {
		subdir := c.Param("sub")
		name := c.Param("name")
		dst = filepath.Join(dst, subdir)
		flist, err := ioutil.ReadDir(dst)
		if err != nil {
			c.String(200, "wrong way")
			return
		}
		var playlist, playitem string
		var thumblocker sync.WaitGroup
		if c.Param("order") != "name" {
			sort.Sort(byModTime(flist))
		}
		// _, showdur := c.Params.Get("dur")
		var fileext, filesrc, filethumb, fileSmi, fileVtt, filename, filebase string
		if subdir != "" {
			srcdir += subdir + "/"
		}
		for _, f := range flist {
			if f.IsDir() || !strings.Contains(f.Name(), name) || strings.HasPrefix(f.Name(), ".fuse") {
				continue
			}
			filename = namereplacer.Replace(f.Name())
			if f.Name() != filename {
				os.Rename(filepath.Join(dst, f.Name()), filepath.Join(dst, filename))
			}
			fileext = strings.ToLower(filepath.Ext(filename))
			filebase = strings.TrimSuffix(filename, fileext)
			if strings.HasSuffix(filebase, "f133") ||
				strings.HasSuffix(filebase, "f242") ||
				strings.HasSuffix(filebase, "f251") ||
				strings.HasSuffix(filebase, "f250") ||
				strings.HasSuffix(filebase, "f140") {
				continue
			}
			switch fileext {
			case ".wav", ".m4a": // 音频
				filesrc = filepath.Join(dst, filename)
				playitem, _ = sjson.Set("", "name", filename)
				playitem, _ = sjson.Set(playitem, "sources.0.src", srcdir+filename)
				playitem, _ = sjson.Set(playitem, "sources.0.type", "audio/"+fileext[1:])
				playlist, _ = sjson.Set(playlist, "pl.-1", gjson.Parse(playitem).Value())
			case ".mp4", ".mkv", ".webm": // 视频
				filesrc = filepath.Join(dst, filename)
				filethumb = filepath.Join(dst, "."+filename+".webp")
				fileSmi = filepath.Join(dst, filebase+".smi")
				fileVtt = filepath.Join(dst, "."+filename+".vtt")
				playitem, _ = sjson.Set("", "name", filename)
				playitem, _ = sjson.Set(playitem, "datetime", f.ModTime().Format("01月02日 15:04"))
				playitem, _ = sjson.Set(playitem, "sources.0.src", srcdir+filename)
				switch fileext {
				case ".mkv":
					playitem, _ = sjson.Set(playitem, "sources.0.type", "video/webm")
				default:
					playitem, _ = sjson.Set(playitem, "sources.0.type", "video/"+fileext[1:])
				}
				// 缩略图
				if !isExist(filethumb) || isExist(strings.TrimSuffix(filesrc, fileext)+".webp") {
					thumblocker.Add(1)
					go func(in, out, ext string) {
						defer thumblocker.Done()
						if isExist(strings.TrimSuffix(in, ext) + ".jpg") {
							cmd := exec.Command("mogrify", "-resize", "168x", "-quality", "80%", "-write", out, strings.Trim(in, ext)+".jpg")
							cmd.Run()
							os.Remove(strings.TrimSuffix(in, ext) + ".jpg")
						} else if isExist(strings.TrimSuffix(in, ext) + ".webp") {
							cmd := exec.Command("mogrify", "-resize", "168x", "-quality", "80%", "-write", out, strings.Trim(in, ext)+".webp")
							cmd.Run()
							os.Remove(strings.TrimSuffix(in, ext) + ".webp")
						} else {
							cmd := exec.Command("ffmpeg", "-i", in, "-ss", "00:00:03", "-s", "168:94", "-vframes", "1", out)
							cmd.Run()
						}
					}(filesrc, filethumb, fileext)
				}
				playitem, _ = sjson.Set(playitem, "thumbnail.0.src", srcdir+"."+filename+".webp")
				// 字幕
				var idx = 0
				for _, v := range subTypes {
					subraw := filepath.Join(dst, filebase+v+".vtt")
					subdst := filepath.Join(dst, "."+filename+v+".vtt")
					subsrc := srcdir + "." + filename + v + ".vtt"
					if isExist(subraw) {
						os.Rename(subraw, subdst)
					}
					if isExist(subdst) {
						playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.src", idx), subsrc)
						playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.label", idx), v[1:])
						playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.srclang", idx), v[1:])
						if idx == 0 {
							playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.default", idx), "true")
						}
						idx++
					}
				}
				if idx == 0 {
					if isExist(fileSmi) && !isExist(fileVtt) {
						smi2Vtt(fileSmi, fileVtt)
					}
					if isExist(fileVtt) {
						playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.src", idx), srcdir+"."+filename+".vtt")
						playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.label", idx), "中文")
						playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.srclang", idx), "zh")
						playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.default", idx), "false")
					}
				}
				playlist, _ = sjson.Set(playlist, "pl.-1", gjson.Parse(playitem).Value())
			case ".dur", ".png", ".jpg", ".vtt", ".webp":
				if strings.HasPrefix(filename, ".") {
					if !isExist(filepath.Join(dst, extreplacer.Replace(filename[1:]))) {
						os.Remove(filepath.Join(dst, filename))
					}
				}
			case ".jpg~":
				os.Remove(filepath.Join(dst, filename))
			}
		}
		tpl := strings.Replace(webtpl, "playlist_data_here", gjson.Parse(playlist).Get("pl").String(), 1)

		thumblocker.Wait()
		c.Header("Content-Type", "text/html")
		c.String(200, tpl)
	}
}
func unsafeString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
func isExist(p string) bool {
	if p == "" {
		return false
	}
	_, err := os.Stat(p)
	return err == nil || os.IsExist(err)
}
func main() {
	// http 静态目录
	flag.Var(&dirs, "dir", "example: -dir=name:path -dir name2:path2")
	// web gallery 目录
	flag.Var(&wtv, "wtv", "example: -wtv=name:path -wtv name2:path2")
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	if len(dirs) == 0 {
		dirs = []string{"ls:."}
	}

	if !*debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(cors.New(cors.Config{
		MaxAge:           time.Hour * 24,
		AllowAllOrigins:  true,
		AllowCredentials: true,
		AllowWildcard:    true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
	}))
	if *html != "" {
		r.Static("/html", *html)
		r.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusTemporaryRedirect, "/html/index.html")
		})
	}
	if *auth {
		r.Use(gin.BasicAuth(gin.Accounts{
			"golang":     "based",
			"thewhyofgo": "simple&fast",
		}))
	}
	// 静态资源
	r.StaticFS("/emb", http.FS(stat))
	// 静态资源
	for _, v := range dirs {
		if strings.Contains(v, ":") {
			r.StaticFS("/"+strings.Split(v, ":")[0], http.Dir(strings.Split(v, ":")[1]))
		}
	}
	// webtv
	for _, v := range wtv {
		if strings.Contains(v, ":") {
			r.GET("/v/"+strings.Split(v, ":")[0], runVideojs(strings.Split(v, ":")[0], strings.Split(v, ":")[1]))
			r.StaticFS("/v-"+strings.Split(v, ":")[0], http.Dir(strings.Split(v, ":")[1]))
		}
	}
	// startup
	if !*debug {
		println(fmt.Sprintf("=== server start on :%d ===", *port))
	}
	if *cert != "" && *key != "" && *ports > 0 {
		go func() {
			r.RunTLS(fmt.Sprintf(":%d", *ports), *cert, *key)
		}()
	}
	r.Run(fmt.Sprintf(":%d", *port))
}

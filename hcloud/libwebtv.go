package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

//go:embed static
var stat embed.FS

//go:embed tpl/webtv.html
var pageWebTV string

//go:embed tpl/ariang.html
var pageAriang string

var (
	extreplacer  = strings.NewReplacer(".webp", "", ".dur", "", ".smi", "", ".png", "", ".jpg", "", ".zh-Hans", "", ".zh-Hant", "", ".vtt", "", ".en", "", ".en-US", "")
	namereplacer = strings.NewReplacer("#", "", "%", "")
	subTypes     = []string{".zh-Hant", ".zh-Hans", ".en", ".en-US"}
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
		} else {
			tStart, _ = strconv.ParseInt(v[idxTime1+6:idxTime2], 10, 0)
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

func runVideojs(url, urldst string) gin.HandlerFunc {
	return func(c *gin.Context) {
		srcdir := "/v-" + url + "/"
		subdir := c.Param("sub")
		name := c.Param("name")
		dst := filepath.Join(urldst, subdir)
		flist, err := ioutil.ReadDir(dst)
		if err != nil {
			println(dst, err.Error())
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
				// filesrc = filepath.Join(dst, filename)
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
				if runtime.GOARCH == "amd64" {
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
				}
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
		tpl := strings.Replace(pageWebTV, "playlist_data_here", gjson.Parse(playlist).Get("pl").String(), 1)

		thumblocker.Wait()
		c.Header("Content-Type", "text/html")
		c.String(200, tpl)
	}
}

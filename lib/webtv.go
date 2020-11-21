package lib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xyzj/gopsu"
	ginmiddleware "github.com/xyzj/gopsu/gin-middleware"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type videoinfo struct {
	url    string
	format string
	try    int
}

type videoformat struct {
	audio250 bool
	audio251 bool
	audio140 bool
	video242 bool
	video133 bool
	video18  bool
}

func (vf *videoformat) Format() string {
	s := make([]string, 0)
	//"242+250/133+250/133+140/18"
	if vf.video242 {
		if vf.audio250 {
			s = append(s, "242+250")
		} else if vf.audio251 {
			s = append(s, "242+251")
		} else if vf.audio140 {
			s = append(s, "242+140")
		}
	} else if vf.video133 {
		if vf.audio250 {
			s = append(s, "133+250")
		} else if vf.audio251 {
			s = append(s, "133+251")
		} else if vf.audio140 {
			s = append(s, "133+140")
		}
	} else if vf.video18 {
		s = append(s, "18")
	}
	return strings.Join(s, "/")
}

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
	pageWebTV           string
	chanDownloadControl = make(chan *videoinfo, 100)
	extreplacer         = strings.NewReplacer(".dur", "", ".smi", "", ".png", "", ".jpg", "", ".zh-Hans", "", ".zh-Hant", "", ".vtt", "", ".en", "", ".en-US", "")
	namereplacer        = strings.NewReplacer("#", "", "%", "")
	subTypes            = []string{".zh-Hant", ".zh-Hans", ".en", ".en-US"}
)

func runVideojs(c *gin.Context) {
	dir := c.Param("dir")
	subdir := c.Param("sub")
	name := c.Param("name")
	dst, err := urlConf.GetItem("tv-" + dir)
	if err != nil {
		ginmiddleware.Page404(c)
		return
	}
	dst = filepath.Join(dst, subdir)
	flist, err := ioutil.ReadDir(dst)
	if err != nil {
		ginmiddleware.Page404(c)
		return
	}
	var playlist, playitem string
	var thumblocker sync.WaitGroup
	if c.Param("order") != "name" {
		sort.Sort(byModTime(flist))
	}
	_, showdur := c.Params.Get("dur")
	var fileext, filesrc, filethumb, filedur, fileSmi, fileVtt, filename, filebase, srcdir string
	var dur int
	srcdir = "/tv-" + dir + "/"
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
		if strings.Contains(".f242.f251.f250.f133.f140", filebase[len(filebase)-5:]) {
			continue
		}
		switch fileext {
		case ".wav", ".m4a": // 音频
			filesrc = filepath.Join(dst, filename)
			filedur = filepath.Join(dst, "."+filename+".dur")
			dur = 0
			if showdur {
				if !gopsu.IsExist(filedur) {
					go func(in, out string) {
						thumblocker.Add(1)
						defer thumblocker.Done()
						cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", "-i", in)
						b, err := cmd.CombinedOutput()
						if err == nil {
							s := bytes.Split(b, []byte("."))[0]
							ioutil.WriteFile(out, s, 0664)
						}
					}(filesrc, filedur)
				}
			}
			playitem, _ = sjson.Set("", "name", filename)
			b, err := ioutil.ReadFile(filedur)
			if err == nil {
				dur = gopsu.String2Int(string(b), 10)
				playitem, _ = sjson.Set(playitem, "duration", dur)
			}
			playitem, _ = sjson.Set(playitem, "sources.0.src", srcdir+filename)
			playitem, _ = sjson.Set(playitem, "sources.0.type", "audio/"+fileext[1:])
			playlist, _ = sjson.Set(playlist, "pl.-1", gjson.Parse(playitem).Value())
		case ".mp4", ".mkv", ".webm": // 视频
			filesrc = filepath.Join(dst, filename)
			filethumb = filepath.Join(dst, "."+filename+".jpg")
			filedur = filepath.Join(dst, "."+filename+".dur")
			fileSmi = filepath.Join(dst, filebase+".smi")
			fileVtt = filepath.Join(dst, "."+filename+".vtt")
			// 视频长度
			dur = 0
			if showdur {
				if !gopsu.IsExist(filedur) {
					go func(in, out string) {
						thumblocker.Add(1)
						defer thumblocker.Done()
						cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", "-i", in)
						b, err := cmd.CombinedOutput()
						if err == nil {
							s := bytes.Split(b, []byte("."))[0]
							ioutil.WriteFile(out, s, 0664)
						}
					}(filesrc, filedur)
				}
			}
			playitem, _ = sjson.Set("", "name", filename)
			// playitem, _ = sjson.Set(playitem, "description", f.ModTime().String())
			b, err := ioutil.ReadFile(filedur)
			if err == nil {
				dur = gopsu.String2Int(string(b), 10)
				playitem, _ = sjson.Set(playitem, "duration", dur)
			}
			playitem, _ = sjson.Set(playitem, "datetime", f.ModTime().Format("01月02日 15:04"))
			playitem, _ = sjson.Set(playitem, "sources.0.src", srcdir+filename)
			switch fileext {
			case ".mkv":
				playitem, _ = sjson.Set(playitem, "sources.0.type", "video/webm")
			default:
				playitem, _ = sjson.Set(playitem, "sources.0.type", "video/"+fileext[1:])
			}
			// 缩略图
			if !gopsu.IsExist(filethumb) || gopsu.IsExist(strings.TrimSuffix(filesrc, fileext)+".jpg") {
				go func(in, out, ext string) {
					thumblocker.Add(1)
					defer thumblocker.Done()
					if gopsu.IsExist(strings.TrimSuffix(in, ext) + ".jpg") {
						cmd := exec.Command("mogrify", "-resize", "256x", "-quality", "80%", "-write", out, strings.Trim(in, ext)+".jpg")
						cmd.Run()
						os.Remove(strings.TrimSuffix(in, ext) + ".jpg")
					} else {
						cmd := exec.Command("ffmpeg", "-i", in, "-ss", "00:00:03", "-s", "256:144", "-vframes", "1", out)
						cmd.Run()
					}
				}(filesrc, filethumb, fileext)
			}
			playitem, _ = sjson.Set(playitem, "thumbnail.0.src", srcdir+"."+filename+".jpg")
			// 字幕
			var idx = 0
			for _, v := range subTypes {
				subraw := filepath.Join(dst, filebase+v+".vtt")
				subdst := filepath.Join(dst, "."+filename+v+".vtt")
				subsrc := srcdir + "." + filename + v + ".vtt"
				if gopsu.IsExist(subraw) {
					os.Rename(subraw, subdst)
				}
				if gopsu.IsExist(subdst) {
					playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.src", idx), subsrc)
					playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.label", idx), v[1:])
					// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.kind", idx), "subtitles")
					playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.srclang", idx), v[1:])
					if idx == 0 {
						playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.default", idx), "true")
					}
					idx++
				}
			}
			if idx == 0 {
				if gopsu.IsExist(fileSmi) && !gopsu.IsExist(fileVtt) {
					Smi2Vtt(fileSmi, fileVtt)
				}
				if gopsu.IsExist(fileVtt) {
					playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.src", idx), srcdir+"."+filename+".vtt")
					playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.label", idx), "中文")
					playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.srclang", idx), "zh")
					playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.default", idx), "false")
				}
			}
			playlist, _ = sjson.Set(playlist, "pl.-1", gjson.Parse(playitem).Value())
		case ".dur", ".png", ".jpg", ".vtt":
			if strings.HasPrefix(filename, ".") {
				if !gopsu.IsExist(filepath.Join(dst, extreplacer.Replace(filename[1:]))) {
					os.Remove(filepath.Join(dst, filename))
				}
			}
		// case ".vtt":
		// 	if strings.HasPrefix(filename, ".") {
		// 		if !gopsu.IsExist(filepath.Join(dst, extreplacer.Replace(filename[1:])) + ".mp4") {
		// 			os.Remove(filepath.Join(dst, filename))
		// 		}
		// 	}
		case ".jpg~":
			os.Remove(filepath.Join(dst, filename))
		}
	}

	tpl := strings.Replace(pageWebTV, "playlist_data_here", gjson.Parse(playlist).Get("pl").String(), 1)
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)

	thumblocker.Wait()
	render.WriteString(c.Writer, tpl, nil)
}

func ydl(c *gin.Context) {
	var v string
	var ok bool
	if v, ok = c.Params.Get("v"); !ok {
		c.String(200, "need param v to set video url")
		return
	}
	chanDownloadControl <- &videoinfo{url: v, format: strings.ReplaceAll(c.Param("f"), " ", "+")}
	c.String(200, "The video file has started downloading... ")
}

func ydlb(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusOK)
		render.WriteString(c.Writer, tplydl, nil)
	case "POST":
		vlist := strings.Split(c.Param("vlist"), "\n")
		for _, vl := range vlist {
			if strings.Contains(vl, "&") {
				x := strings.Split(gopsu.TrimString(vl), "&")
				chanDownloadControl <- &videoinfo{url: x[0], format: x[1]}
			} else {
				chanDownloadControl <- &videoinfo{url: vl, format: ""}
			}
		}
		c.String(200, "These videos have been added to the download queue...")
	}
}

func downloadControl() {
	var dlock sync.WaitGroup
RUN:
	go func() {
		dlock.Add(1)
		defer func() {
			recover()
			dlock.Done()
		}()
		var scmd bytes.Buffer
		var cmd *exec.Cmd
		var shellName string
		var videoName = "%(title)s"
		for {
			select {
			case vi := <-chanDownloadControl:
				if gopsu.TrimString(vi.url) == "" || vi.try >= 3 {
					continue
				}
				shellName = "/tmp/" + gopsu.CalcCRC32String([]byte(vi.url)) + ".sh"
				if gopsu.IsExist(shellName) && vi.format == "" {
					goto DOWN
				}
				scmd.Reset()
				scmd.WriteString("#!/bin/bash\n\n")

				scmd.WriteString("youtube-dl ")
				scmd.WriteString("--proxy='http://127.0.0.1:8119' ")
				scmd.WriteString("--write-thumbnail ")
				scmd.WriteString("--write-sub --write-auto-sub --sub-lang 'en,en-US,zh-Hant' ")
				scmd.WriteString("--mark-watched ")
				// scmd.WriteString("--youtube-skip-dash-manifest ")
				scmd.WriteString("--skip-unavailable-fragments ")
				// scmd.WriteString("--abort-on-unavailable-fragment ")
				scmd.WriteString("--no-mtime ")
				scmd.WriteString("--buffer-size 64k ")
				// scmd.WriteString("--recode-video mp4 ")
				scmd.WriteString("-o '" + ydir + videoName + ".%(ext)s' ")
				// scmd.WriteString("-o '" + ydir + "%(title)s.%(ext)s' ")
				if vi.format == "" {
					vi.format = "133+140/242+250/242+251/133+250/133+251/18"
				}
				scmd.WriteString("-f '" + vi.format + "' ")
				if strings.HasPrefix(vi.url, "http") {
					scmd.WriteString(vi.url)
				} else {
					scmd.WriteString("-- " + vi.url)
				}
				scmd.WriteString(" && \\\n\\\nrm $0\n")
				ioutil.WriteFile(shellName, scmd.Bytes(), 0755)
			DOWN:
				time.Sleep(time.Second * time.Duration(rand.Int31n(5)))
				cmd = exec.Command(shellName)
				b, err := cmd.CombinedOutput()
				if err != nil {
					b = append(b, []byte("\n"+err.Error()+"\n")...)
					ioutil.WriteFile(shellName+".log", b, 0664)
				}
				if gopsu.IsExist(shellName) {
					time.Sleep(time.Second * time.Duration(rand.Int31n(5)+5))
					vi.try++
					chanDownloadControl <- vi
				} else {
					os.Remove(shellName + ".log")
				}
			}
		}
	}()
	time.Sleep(time.Second)
	dlock.Wait()
	goto RUN
}

package lib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xyzj/gopsu"
	ginmiddleware "github.com/xyzj/gopsu/gin-middleware"
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
	pageWebTV string
)

func runVideojs(c *gin.Context) {
	dir := c.Param("dir")
	subdir := c.Param("sub")
	name := c.Param("name")
	dst, err := urlConf.GetItem("tv-" + dir)
	freplacer := strings.NewReplacer(".dur", "", ".smi", "", ".png", "", ".jpg", "", ".zh-Hans", "", ".zh-Hant", "", ".vtt", "")
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
	var fileext, filesrc, filethumb, filedur, fileSmi, fileVtt, filebase, fileVttHans, fileVttHant, srcdir string
	var dur int
	srcdir = "/tv-" + dir + "/"
	if subdir != "" {
		srcdir += subdir + "/"
	}
	for _, f := range flist {
		if f.IsDir() {
			continue
		}
		if !strings.Contains(f.Name(), name) {
			continue
		}
		fileext = strings.ToLower(filepath.Ext(f.Name()))
		filebase = strings.Trim(f.Name(), fileext)
		switch fileext {
		case ".wav", ".m4a": // 音频
			filesrc = filepath.Join(dst, f.Name())
			filedur = filepath.Join(dst, "."+f.Name()+".dur")
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
			playitem, _ = sjson.Set("", "name", f.Name())
			b, err := ioutil.ReadFile(filedur)
			if err == nil {
				dur = gopsu.String2Int(string(b), 10)
				playitem, _ = sjson.Set(playitem, "duration", dur)
			}
			playitem, _ = sjson.Set(playitem, "sources.0.src", srcdir+f.Name())
			playitem, _ = sjson.Set(playitem, "sources.0.type", "audio/"+fileext[1:])
			playlist, _ = sjson.Set(playlist, "pl.-1", gjson.Parse(playitem).Value())
		case ".mp4", ".mkv": // 视频
			filesrc = filepath.Join(dst, f.Name())
			filethumb = filepath.Join(dst, "."+f.Name()+".jpg")
			filedur = filepath.Join(dst, "."+f.Name()+".dur")
			fileSmi = filepath.Join(dst, filebase+".smi")
			fileVtt = filepath.Join(dst, "."+filebase+".vtt")
			fileVttHans = filepath.Join(dst, "."+filebase+".zh-Hans.vtt")
			fileVttHant = filepath.Join(dst, "."+filebase+".zh-Hant.vtt")
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
			playitem, _ = sjson.Set("", "name", f.Name())
			// playitem, _ = sjson.Set(playitem, "description", f.ModTime().String())
			b, err := ioutil.ReadFile(filedur)
			if err == nil {
				dur = gopsu.String2Int(string(b), 10)
				playitem, _ = sjson.Set(playitem, "duration", dur)
			}
			playitem, _ = sjson.Set(playitem, "datetime", f.ModTime().Format("01月02日 15:04"))
			playitem, _ = sjson.Set(playitem, "sources.0.src", srcdir+f.Name())
			switch fileext {
			case ".mkv":
				playitem, _ = sjson.Set(playitem, "sources.0.type", "video/webm")
			default:
				playitem, _ = sjson.Set(playitem, "sources.0.type", "video/"+fileext[1:])
			}
			// 缩略图
			if !gopsu.IsExist(filethumb) {
				go func(in, out string) {
					thumblocker.Add(1)
					defer thumblocker.Done()
					if gopsu.IsExist(strings.Replace(in, ".mp4", ".jpg", -1)) {
						cmd := exec.Command("mogrify", "-resize", "256x", "-quality", "80%", "-write", out, strings.Replace(in, ".mp4", ".jpg", -1))
						cmd.Run()
						os.Remove(strings.Replace(in, ".mp4", ".jpg", -1))
					} else {
						cmd := exec.Command("ffmpeg", "-i", in, "-ss", "00:00:03", "-s", "256:144", "-vframes", "1", out)
						cmd.Run()
					}
				}(filesrc, filethumb)
			}
			playitem, _ = sjson.Set(playitem, "thumbnail.0.src", srcdir+"."+f.Name()+".jpg")
			// 字幕
			var idx = 0
			if gopsu.IsExist(fileVttHant) {
				playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.src", idx), srcdir+"."+filebase+".zh-Hant.vtt")
				playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.label", idx), "中文繁体")
				// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.kind", idx), "captions")
				playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.srclang", idx), "zh-Hant")
				playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.default", idx), "true")
				idx++
			}
			if gopsu.IsExist(fileVttHans) {
				playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.src", idx), srcdir+"."+filebase+".zh-Hans.vtt")
				playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.label", idx), "中文简体")
				// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.kind", idx), "subtitles")
				playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.srclang", idx), "zh-Hans")
				playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.default", idx), "false")
				idx++
			}
			if idx == 0 {
				if gopsu.IsExist(fileSmi) && !gopsu.IsExist(fileVtt) {
					Smi2Vtt(fileSmi, fileVtt)
				}
				if gopsu.IsExist(fileVtt) {
					playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.src", idx), srcdir+"."+f.Name()+".vtt")
					playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.label", idx), "中文")
					playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.kind", idx), "subtitles")
					playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.srclang", idx), "zh")
					playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.default", idx), "false")
				}
			}
			playlist, _ = sjson.Set(playlist, "pl.-1", gjson.Parse(playitem).Value())
		case ".dur", ".png", ".jpg":
			if strings.HasPrefix(f.Name(), ".") {
				if !gopsu.IsExist(filepath.Join(dst, freplacer.Replace(f.Name()[1:]))) {
					os.Remove(filepath.Join(dst, f.Name()))
				}
			}
		case ".vtt":
			if strings.HasPrefix(f.Name(), ".") {
				if !gopsu.IsExist(filepath.Join(dst, freplacer.Replace(f.Name()[1:])) + ".mp4") {
					os.Remove(filepath.Join(dst, f.Name()))
				}
			}
		case ".jpg~":
			os.Remove(filepath.Join(dst, f.Name()))
		}
	}

	tpl := strings.Replace(pageWebTV, "playlist_data_here", gjson.Parse(playlist).Get("pl").String(), 1)
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)

	thumblocker.Wait()
	render.WriteString(c.Writer, tpl, nil)
}

func ydl(c *gin.Context) {
	var v, f string
	var ok bool
	if v, ok = c.Params.Get("v"); !ok {
		c.String(200, "need param v to set video url")
		return
	}
	var scmd bytes.Buffer
	scmd.WriteString("#!/bin/bash\n\n")

	// scmd.WriteString("youtube-dl ")
	// scmd.WriteString("--proxy='http://127.0.0.1:8119' ")
	// scmd.WriteString("--skip-download ")
	// scmd.WriteString("--write-thumbnail ")
	// scmd.WriteString("-o '/home/xy/mm/xldown/tv/1ncoming/1news/.%(title)s.%(ext)s' ")
	// scmd.WriteString("--exec mogrify -resize 256x -quality 80% {} ")
	// scmd.WriteString(v + "\n\n")

	scmd.WriteString("youtube-dl ")
	scmd.WriteString("--proxy='http://127.0.0.1:8119' ")
	scmd.WriteString("--write-thumbnail ")
	scmd.WriteString("--mark-watched ")
	scmd.WriteString("--youtube-skip-dash-manifest ")
	scmd.WriteString("--merge-output-format mp4 ")
	scmd.WriteString("-o '/home/xy/mm/xldown/tv/1ncoming/1news/%(title)s.%(ext)s' ")
	if f, ok = c.Params.Get("f"); !ok {
		f = "'242+140/133+140' "
	}
	scmd.WriteString("-f " + f + " ")
	if strings.HasPrefix(v, "http") {
		scmd.WriteString(v + " && \\\n\\\n")
	} else {
		scmd.WriteString("-- " + v + " && \\\n\\\n")
	}

	scmd.WriteString("youtube-dl ")
	scmd.WriteString("--proxy='http://127.0.0.1:8119' ")
	scmd.WriteString("--skip-download ")
	scmd.WriteString("--write-sub --write-auto-sub --sub-lang zh-Hant/zh-Hans ")
	scmd.WriteString("-o '/home/xy/mm/xldown/tv/1ncoming/1news/.%(title)s.%(ext)s' ")
	if strings.HasPrefix(v, "http") {
		scmd.WriteString(v + " && \\\n\\\n")
	} else {
		scmd.WriteString("-- " + v + " && \\\n\\\n")
	}

	scmd.WriteString("rm $0\n")
	name := "/tmp/" + gopsu.CalcCRC32String([]byte(v)) + ".sh"
	ioutil.WriteFile(name, scmd.Bytes(), 0755)
	cmd := exec.Command(name)
	cmd.Start()
	c.String(200, "The video file has started downloading... ")
}

package lib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xyzj/gopsu"
	ginmiddleware "github.com/xyzj/gopsu/gin-middleware"
)

func init() {
	rand.Seed(time.Now().Unix())
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
	pageWebTV    string
	extreplacer  = strings.NewReplacer(".webp", "", ".dur", "", ".smi", "", ".png", "", ".jpg", "", ".zh-Hans", "", ".zh-Hant", "", ".vtt", "", ".en", "", ".en-US", "")
	namereplacer = strings.NewReplacer("#", "", "%", "")
	subTypes     = []string{".zh-Hant", ".zh-Hans", ".en", ".en-US"}
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
			filedur = filepath.Join(dst, "."+filename+".dur")
			dur = 0
			if showdur {
				if !gopsu.IsExist(filedur) {
					thumblocker.Add(1)
					go func(in, out string) {
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
			filethumb = filepath.Join(dst, "."+filename+".webp")
			filedur = filepath.Join(dst, "."+filename+".dur")
			fileSmi = filepath.Join(dst, filebase+".smi")
			fileVtt = filepath.Join(dst, "."+filename+".vtt")
			// 视频长度
			dur = 0
			if showdur {
				if !gopsu.IsExist(filedur) {
					thumblocker.Add(1)
					go func(in, out string) {
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
			if !gopsu.IsExist(filethumb) || gopsu.IsExist(strings.TrimSuffix(filesrc, fileext)+".webp") {
				thumblocker.Add(1)
				go func(in, out, ext string) {
					defer thumblocker.Done()
					if gopsu.IsExist(strings.TrimSuffix(in, ext) + ".jpg") {
						cmd := exec.Command("mogrify", "-resize", "168x", "-quality", "80%", "-write", out, strings.Trim(in, ext)+".jpg")
						cmd.Run()
						os.Remove(strings.TrimSuffix(in, ext) + ".jpg")
					} else if gopsu.IsExist(strings.TrimSuffix(in, ext) + ".webp") {
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
		case ".dur", ".png", ".jpg", ".vtt", ".webp":
			if strings.HasPrefix(filename, ".") {
				if !gopsu.IsExist(filepath.Join(dst, extreplacer.Replace(filename[1:]))) {
					os.Remove(filepath.Join(dst, filename))
				}
			}
			// os.Remove(filepath.Join(dst, filename))
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

	thumblocker.Wait()
	c.Header("Content-Type", "text/html")
	c.String(200, tpl)
	// c.Status(http.StatusOK)
	// render.WriteString(c.Writer, tpl, nil)
}

package lib

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/sjson"
	"github.com/xyzj/gopsu"
)

type videoinfo struct {
	url    string
	format string
	try    int
}

var (
	chanHTTPDownloader    = make(chan *videoinfo, 100)
	chanYoutubeDownloader = make(chan *videoinfo, 100)
)

func tdlb(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		c.Header("Content-Type", "text/html")
		c.String(200, tplth)
	case "POST":
		vlist := strings.Split(c.Param("vlist"), "\n")
		for _, vl := range vlist {
		START:
			s := strings.Split(vl, ":")
			switch s[0] {
			case "thunder":
				ss, _ := base64.StdEncoding.DecodeString(s[1][2:])
				a, _ := gopsu.GbkToUtf8(ss[2 : len(ss)-2])
				vl = string(a)
				goto START
			case "http", "https":
				if strings.Contains(vl, "www.youtube.com") {
					if strings.Contains(vl, "&") {
						x := strings.Split(gopsu.TrimString(vl), "&")
						chanYoutubeDownloader <- &videoinfo{url: x[0], format: x[1]}
					} else {
						chanYoutubeDownloader <- &videoinfo{url: vl, format: ""}
					}
				} else {
					chanHTTPDownloader <- &videoinfo{url: vl}
				}
			case "magnet":
				s, _ := sjson.SetBytes([]byte{}, "jsonrpc", "2.0")
				s, _ = sjson.SetBytes(s, "id", fmt.Sprintf("%d", time.Now().UnixNano()))
				s, _ = sjson.SetBytes(s, "method", "aria2.addUri")
				s, _ = sjson.SetBytes(s, "params.0.0", vl)
				ss := strings.ReplaceAll(base64.URLEncoding.EncodeToString(s), "=", "%3D")
				req, _ := http.NewRequest("GET", "http://127.0.0.1:60090/jsonrpc?params="+ss, strings.NewReader(""))
				resp, err := httpClientPool.Do(req)
				shellName := "/tmp/" + gopsu.CalcCRC32String([]byte(vl)) + ".aria2.log"
				if err != nil {
					ioutil.WriteFile(shellName, []byte(vl+"\n\n"+err.Error()), 0664)
				} else {
					body, _ := ioutil.ReadAll(resp.Body)
					if strings.Contains(string(body), "error") {
						ioutil.WriteFile(shellName, []byte(vl+"\n\n"+string(body)), 0664)
					}
				}
			}
			// magnet:?xt=urn:btih:6f2359c12381e22c2fc0ea0b86fb9754c0ca999d
		}
		c.String(200, "These links have been added to the download queue...")
	}
}

func httpControl() {
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
		for {
			select {
			case vi := <-chanHTTPDownloader:
				if gopsu.TrimString(vi.url) == "" || vi.try >= 3 {
					continue
				}
				shellName = "/tmp/" + gopsu.CalcCRC32String([]byte(vi.url)) + ".sh"
				if gopsu.IsExist(shellName) {
					goto DOWN
				}
				scmd.Reset()
				scmd.WriteString("#!/bin/bash\n\n")
				scmd.WriteString("# " + vi.url + "\n\n")
				scmd.WriteString("wget ")
				scmd.WriteString("-c ")
				scmd.WriteString("-t 10 ")
				scmd.WriteString("-w 20 --random-wait ")
				scmd.WriteString("--no-check-certificate ")
				scmd.WriteString("-P " + tdir + " ")
				scmd.WriteString("\"" + vi.url + "\"")
				scmd.WriteString(" && \\\n\\\nrm $0\n")
				ioutil.WriteFile(shellName, scmd.Bytes(), 0755)
			DOWN:
				time.Sleep(time.Second * time.Duration(rand.Int31n(5)+10))
				cmd = exec.Command(shellName)
				b, err := cmd.CombinedOutput()
				if err != nil {
					b = append(b, []byte("\n"+err.Error()+"\n")...)
					ioutil.WriteFile(shellName+".log", b, 0664)
				}
				time.Sleep(time.Second * time.Duration(rand.Int31n(5)+3))
				if gopsu.IsExist(shellName) {
					vi.try++
					chanHTTPDownloader <- vi
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

func youtubeControl() {
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
			case vi := <-chanYoutubeDownloader:
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
				scmd.WriteString("--continue ")
				scmd.WriteString("--write-thumbnail ")
				scmd.WriteString("--write-sub --write-auto-sub --sub-lang 'en,en-US,zh-Hant' ")
				// scmd.WriteString("--mark-watched ")
				// scmd.WriteString("--youtube-skip-dash-manifest ")
				scmd.WriteString("--skip-unavailable-fragments ")
				// scmd.WriteString("--abort-on-unavailable-fragment ")
				scmd.WriteString("--no-mtime ")
				scmd.WriteString("--buffer-size 256k ")
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
				time.Sleep(time.Second * time.Duration(rand.Int31n(5)+10))
				cmd = exec.Command(shellName)
				b, err := cmd.CombinedOutput()
				if err != nil {
					b = append(b, []byte("\n"+err.Error()+"\n")...)
					ioutil.WriteFile(shellName+".log", b, 0664)
				}
				time.Sleep(time.Second * time.Duration(rand.Int31n(5)+3))
				if gopsu.IsExist(shellName) {
					if !strings.Contains(string(b), "Unable to extract video data") {
						vi.try++
					}
					chanYoutubeDownloader <- vi
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

func ydl(c *gin.Context) {
	var v string
	var ok bool
	if v, ok = c.Params.Get("v"); !ok {
		c.String(200, "need param v to set video url")
		return
	}
	chanYoutubeDownloader <- &videoinfo{url: v, format: strings.ReplaceAll(c.Param("f"), " ", "+")}
	c.String(200, "The video file has started downloading... ")
}

func ydlb(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		c.Header("Content-Type", "text/html")
		c.String(200, tplydl)
		// c.Status(http.StatusOK)
		// render.WriteString(c.Writer, tplydl, nil)
	case "POST":
		vlist := strings.Split(c.Param("vlist"), "\n")
		for _, vl := range vlist {
			if strings.Contains(vl, "&") {
				x := strings.Split(gopsu.TrimString(vl), "&")
				chanYoutubeDownloader <- &videoinfo{url: x[0], format: x[1]}
			} else {
				chanYoutubeDownloader <- &videoinfo{url: vl, format: ""}
			}
		}
		c.String(200, "These videos have been added to the download queue...")
	}
}

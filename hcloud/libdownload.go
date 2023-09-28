package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/sjson"
	"github.com/xyzj/gopsu"
	"github.com/xyzj/gopsu/txtcode"
)

type videoinfo struct {
	url    string
	format string
	try    int
}

var (
	// chanHTTPDownloader    = make(chan *videoinfo, 100)
	chanYoutubeDownloader = make(chan videoinfo, 100)
)

func tdlb(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		c.Header("Content-Type", "text/html")
		c.String(200, tplth)
	case "POST":
		vlist := strings.Split(c.Param("vlist"), "\n")
		for _, vl := range vlist {
			if gopsu.TrimString(vl) == "" {
				continue
			}
		START:
			s := strings.Split(vl, ":")
			switch s[0] {
			case "thunder":
				ss, _ := base64.StdEncoding.DecodeString(s[1][2:])
				a, _ := txtcode.GbkToUtf8(ss[2 : len(ss)-2])
				vl = string(a)
				goto START
			case "http", "https":
				if strings.Contains(vl, "www.youtube.com") {
					vl = strings.ReplaceAll(vl, "&pp=sAQA", "")
					if strings.Contains(vl, "&&") {
						x := strings.Split(gopsu.TrimString(vl), "&&")
						chanYoutubeDownloader <- videoinfo{url: x[0], format: x[1]}
					} else {
						chanYoutubeDownloader <- videoinfo{url: vl, format: ""}
					}
				} else {
					furl := vl
					if !strings.Contains(vl, "%") {
						idx := strings.LastIndex(vl, "/")
						furl = vl[:idx+1] + url.QueryEscape(vl[idx+1:])
					}
					rpcToAria2(furl)
				}
			case "magnet":
				rpcToAria2(vl)
			}
			// magnet:?xt=urn:btih:6f2359c12381e22c2fc0ea0b86fb9754c0ca999d
		}
		c.String(200, "These links have been added to the download queue...")
	}
}

func rpcToAria2(vl string) {
	s, _ := sjson.SetBytes([]byte{}, "jsonrpc", "2.0")
	s, _ = sjson.SetBytes(s, "id", fmt.Sprintf("%d", time.Now().UnixNano()))
	s, _ = sjson.SetBytes(s, "method", "aria2.addUri")
	s, _ = sjson.SetBytes(s, "params.0.0", vl)
	ss := strings.ReplaceAll(base64.URLEncoding.EncodeToString(s), "=", "%3D")
	req, _ := http.NewRequest("GET", *aria2+"/jsonrpc?params="+ss, strings.NewReader(""))
	resp, err := httpClientPool.Do(req)
	shellName := "/tmp/" + gopsu.CalcCRC32String([]byte(vl)) + ".aria2.log"
	if err != nil {
		os.WriteFile(shellName, []byte(vl+"\n\n"+err.Error()), 0664)
	} else {
		body, _ := io.ReadAll(resp.Body)
		if strings.Contains(string(body), "error") {
			os.WriteFile(shellName, []byte(vl+"\n\n"+string(body)), 0664)
		}
	}
}

func youtubeControl() {
	// videoNameReplacer := strings.NewReplacer(
	// 	"WARNING:", "",
	// 	"Failedtodownloadm3u8information:", "",
	// 	"FailedtodownloadMPDmanifest:", "",
	// 	"Unabletodownloadwebpage:", "",
	// 	"UnabletodownloadAPIpage:", "",
	// 	"<urlopenerror[Errno0]Error>", "",
	// 	"Nostatuslinereceived-theserverhasclosedtheconnection", "",
	// 	"HTTPError429:TooManyRequests", "",
	// 	"\"", "", "'", "", "、", ";", "%", "", "\n", "", "\r", "", "；", ";", "：", ":", "（", "", "）", "", "？", "", " ", "", " ", "", "《", "<", "》", ">", "！", "", "，", ",", "。", "", "“", "", "”", "")
RUN:
	func() {
		defer func() {
			if err := recover(); err != nil {
				println(err.(error).Error())
			}
		}()
		var scmd bytes.Buffer
		var cmd *exec.Cmd
		var shellName string
		var videoName = "%(title)s"
		for vi := range chanYoutubeDownloader {
			videoName = "%(title).150B"
			if gopsu.TrimString(vi.url) == "" || vi.try >= 5 {
				continue
			}
			fname := gopsu.CalcCRC32String([]byte(vi.url))
			// shellName = "/tmp/" + fname + "-name.sh"
			// scmd.Reset()
			// scmd.WriteString("#!/bin/bash\n\n")
			// scmd.WriteString("/home/xy/xbin/yt-dlp_linux ")
			// scmd.WriteString("--proxy='http://127.0.0.1:8119' ")
			// scmd.WriteString("--get-filename ")
			// scmd.WriteString("-o '%(title)s' ")
			// scmd.WriteString("'" + vi.url + "'")
			// scmd.WriteString(" && \\\nrm $0\n")
			// os.WriteFile(shellName, scmd.Bytes(), 0755)
			// cmd = exec.Command(shellName)
			// if b, err := cmd.CombinedOutput(); err == nil {
			// 	// s := gopsu.String(b)
			// 	x := videoNameReplacer.Replace(gopsu.String(b))
			// 	if len(x) > 230 {
			// 		for k := range x {
			// 			if k >= 230 {
			// 				x = x[:k]
			// 				break
			// 			}
			// 		}
			// 	}
			// 	if len(x) > 0 {
			// 		videoName = x
			// 	}
			// }
			shellName = "/tmp/" + fname + ".sh"
			// if isExist(shellName) && vi.format == "" {
			// 	goto DOWN
			// }
			scmd.Reset()

			if runtime.GOARCH == "amd64" {
				scmd.WriteString("#!/bin/bash\n\n")
				scmd.WriteString("/usr/local/bin/yt-dlp ")
			} else {
				scmd.WriteString("#!/bin/ash\n\n")
				scmd.WriteString("/usr/bin/yt-dlp ") //python3 -m pip install -U yt-dlp
			}
			scmd.WriteString("--proxy='http://127.0.0.1:8119' ")
			scmd.WriteString("--continue ")
			// scmd.WriteString("--downloader=aria2c ")
			scmd.WriteString("--no-get-comments ")
			scmd.WriteString("--trim-filenames 55 ")
			scmd.WriteString("--write-thumbnail ")
			scmd.WriteString("--retries 10 ")
			// scmd.WriteString("--write-sub --write-auto-sub --sub-lang 'en,en-US,zh-Hant' ")
			// scmd.WriteString("--mark-watched ")
			// scmd.WriteString("--youtube-skip-dash-manifest ")
			scmd.WriteString("--skip-unavailable-fragments ")
			// scmd.WriteString("--abort-on-unavailable-fragment ")
			scmd.WriteString("--no-mtime ")
			scmd.WriteString("--buffer-size 256k ")
			// scmd.WriteString("--recode-video mp4 ")
			scmd.WriteString("-o '" + *ydir + videoName + ".%(ext)s' ")
			if vi.format == "" {
				vi.format = "242+249/133+140/18"
			}
			scmd.WriteString("-f '" + vi.format + "' ")
			if strings.HasPrefix(vi.url, "http") {
				scmd.WriteString("'" + vi.url + "'")
			} else {
				scmd.WriteString("-- " + vi.url)
			}
			scmd.WriteString(" && \\\nrm $0\n")
			os.WriteFile(shellName, scmd.Bytes(), 0755)
			// DOWN:
			time.Sleep(time.Second * time.Duration(rand.Int31n(5)+10))
			cmd = exec.Command(shellName)
			b, err := cmd.CombinedOutput()
			if err != nil {
				b = append(b, []byte("\n"+err.Error()+"\n")...)
				os.WriteFile(shellName+".log", b, 0664)
			}
			time.Sleep(time.Second * 40)
			if isExist(shellName) {
				out := strings.ToLower(string(b))
				if strings.Contains(out, "error") ||
					strings.Contains(out, "errno") ||
					strings.Contains(out, "filename too long") {
					vi.try++
				}
				chanYoutubeDownloader <- vi
			} else {
				os.Remove(shellName + ".log")
			}
		}
	}()
	time.Sleep(time.Second)
	goto RUN
}

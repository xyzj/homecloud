package lib

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/sjson"
	"github.com/xyzj/toolbox/crypto"
	"github.com/xyzj/toolbox/httpclient"
	"github.com/xyzj/toolbox/pathtool"
)

type processInfo struct {
	Pid     int
	Name    string
	CmdLine string
}
type videoinfo struct {
	url    string
	format string
	try    int
}

var chanYoutubeDownloader = make(chan videoinfo, 100)
var chanYoutubeDownloaderShell = make(chan string, 100)

// queryProcess only for linux
func queryProcess(name string) []*processInfo {
	pi := make([]*processInfo, 0)
	procs, err := os.ReadDir("/proc")
	if err != nil {
		return pi
	}
	for _, proc := range procs {
		if !proc.IsDir() {
			continue
		}
		pid, _ := strconv.ParseInt(proc.Name(), 10, 32)
		if pid == 0 {
			continue
		}
		cmd, _ := os.ReadFile("/proc/" + proc.Name() + "/cmdline")
		if len(cmd) == 0 {
			continue
		}
		cl := strings.Split(string(cmd), "\x00")
		if name != filepath.Base(cl[0]) {
			continue
		}
		pi = append(pi, &processInfo{
			Name:    name,
			Pid:     int(pid),
			CmdLine: strings.Join(cl, " "),
		})
	}
	return pi
}

func tdlb(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		c.Header("Content-Type", "text/html")
		c.String(200, tplth)
	case "POST":
		vlist := strings.Split(c.Param("vlist"), "\n")
		for _, vl := range vlist {
			if strings.TrimSpace(vl) == "" {
				continue
			}
			// START:
			s := strings.Split(vl, ":")
			switch s[0] {
			// case "thunder":
			// 	ss, _ := base64.StdEncoding.DecodeString(s[1][2:])
			// 	a, _ := toolbox.GbkToUtf8(ss[2 : len(ss)-2])
			// 	vl = string(a)
			// 	goto START
			case "http", "https":
				if strings.Contains(vl, "www.youtube.com") {
					vl = strings.ReplaceAll(vl, "&pp=sAQA", "")
					if strings.Contains(vl, "&&") {
						x := strings.Split(strings.TrimSpace(vl), "&&")
						buildYoutubeShell(&videoinfo{url: x[0], format: x[1]})
					} else {
						buildYoutubeShell(&videoinfo{url: vl, format: ""})
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
	if pi := queryProcess("aria2c"); len(pi) == 0 {
		cmd := exec.Command("/opt/bin/ssdctl", "restart", "aria2c")
		cmd.Dir = "/opt/bin"
		cmd.Run()
	}
	s, _ := sjson.SetBytes([]byte{}, "jsonrpc", "2.0")
	s, _ = sjson.SetBytes(s, "id", fmt.Sprintf("%d", time.Now().UnixNano()))
	s, _ = sjson.SetBytes(s, "method", "aria2.addUri")
	s, _ = sjson.SetBytes(s, "params.0.0", vl)
	ss := strings.ReplaceAll(base64.URLEncoding.EncodeToString(s), "=", "%3D")
	req, _ := http.NewRequest("GET", aria2curl+"/jsonrpc?params="+ss, strings.NewReader(""))
	_, body, _, err := httpclient.DoRequestWithTimeout(req, time.Second*5)
	shellName := "/tmp/" + crypto.Crc32IEEE([]byte(vl)) + ".aria2.log"
	if err != nil {
		os.WriteFile(shellName, []byte(vl+"\n\n"+err.Error()), 0o664)
	} else {
		if strings.Contains(string(body), "error") {
			os.WriteFile(shellName, []byte(vl+"\n\n"+string(body)), 0o664)
		}
	}
}

func buildYoutubeShell(vi *videoinfo) {
	var scmd bytes.Buffer
	var shellName string
	videoName := "%(title)s"
	videoName = "%(title).150B"
	if strings.TrimSpace(vi.url) == "" { // || vi.try >= 5 {
		return
	}
	fname := crypto.Crc32IEEE([]byte(vi.url))
	shellName = "/tmp/" + fname + ".sh"
	// if pathtool.IsExist(shellName) && vi.format == "" {
	// 	goto DOWN
	// }
	scmd.Reset()

	if runtime.GOARCH == "amd64" {
		scmd.WriteString("#!/bin/bash\n\n")
		scmd.WriteString("export PATH=$PATH:$HOME/.deno/bin\n\n")
		scmd.WriteString("/usr/local/bin/yt-dlp ")
	} else {
		scmd.WriteString("#!/bin/ash\n\n")
		scmd.WriteString("/usr/bin/yt-dlp ") // python3 -m pip install -U yt-dlp
	}
	// scmd.WriteString("--proxy='http://127.0.0.1:8119' ")
	scmd.WriteString("--continue ")
	if vi.format == "" {
		vi.format = "242+249/133+140/134+139/93/18"
	}
	scmd.WriteString("-f '" + vi.format + "' ")
	// scmd.WriteString("--downloader=aria2c ")
	scmd.WriteString("--no-get-comments ")
	scmd.WriteString("--trim-filenames 55 ")
	scmd.WriteString("--write-thumbnail ")
	if pathtool.IsExist("/opt/bin/www.youtube.com_cookies.txt") {
		scmd.WriteString("--cookies '/opt/bin/www.youtube.com_cookies.txt' ")
	}
	// scmd.WriteString("--retries 10 ")
	// scmd.WriteString("--write-subs --write-auto-subs --sub-langs 'en,en-US,zh-Hant,zh-Hans' ")
	// scmd.WriteString("--mark-watched ")
	// scmd.WriteString("--youtube-skip-dash-manifest ")
	scmd.WriteString("--skip-unavailable-fragments ")
	// scmd.WriteString("--abort-on-unavailable-fragment ")
	scmd.WriteString("--no-mtime ")
	scmd.WriteString("--buffer-size 512k ")
	scmd.WriteString("--sleep-requests 1 ")
	scmd.WriteString("--sleep-interval 5 ")
	// scmd.WriteString("--recode-video mp4 ")
	scmd.WriteString("-o '" + ytdir + videoName + ".%(ext)s' ")
	if strings.HasPrefix(vi.url, "http") {
		scmd.WriteString("'" + vi.url + "'")
	} else {
		scmd.WriteString("-- " + vi.url)
	}
	scmd.WriteString(" && \\\nrm $0\n")
	os.WriteFile(shellName, scmd.Bytes(), 0o755)
	chanYoutubeDownloaderShell <- shellName
}

func YoutubeControl() {
	t := time.NewTimer(time.Second * time.Duration(rand.Intn(5)+10))
	var cmd *exec.Cmd
	for range t.C {
		sh := <-chanYoutubeDownloaderShell
		cmd = exec.Command(sh)
		// cmd.Env = os.Environ()
		b, err := cmd.CombinedOutput()
		if err != nil {
			b = append(b, []byte("\n"+err.Error()+"\n")...)
			os.WriteFile(sh+".log", b, 0o664)
		}
		t.Reset(time.Second * time.Duration(rand.Intn(5)+10))
	}
}

func YoutubeControlOld() {
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
		videoName := "%(title)s"
		for vi := range chanYoutubeDownloader {
			videoName = "%(title).150B"
			if strings.TrimSpace(vi.url) == "" { // || vi.try >= 5 {
				continue
			}
			fname := crypto.Crc32IEEE([]byte(vi.url))
			shellName = "/tmp/" + fname + ".sh"
			// if pathtool.IsExist(shellName) && vi.format == "" {
			// 	goto DOWN
			// }
			scmd.Reset()

			if runtime.GOARCH == "amd64" {
				scmd.WriteString("#!/bin/bash\n\n")
				scmd.WriteString("export PATH=$PATH:$HOME/.deno/bin\n\n")
				scmd.WriteString("/usr/local/bin/yt-dlp ")
			} else {
				scmd.WriteString("#!/bin/ash\n\n")
				scmd.WriteString("/usr/bin/yt-dlp ") // python3 -m pip install -U yt-dlp
			}
			scmd.WriteString("--proxy='http://127.0.0.1:8119' ")
			scmd.WriteString("--continue ")
			if vi.format == "" {
				vi.format = "242+249/133+140/134+139/93/18"
			}
			scmd.WriteString("-f '" + vi.format + "' ")
			// scmd.WriteString("--downloader=aria2c ")
			scmd.WriteString("--no-get-comments ")
			scmd.WriteString("--trim-filenames 55 ")
			scmd.WriteString("--write-thumbnail ")
			scmd.WriteString("--cookies ")
			scmd.WriteString("'/opt/bin/www.youtube.com_cookies.txt' ")
			// scmd.WriteString("--retries 10 ")
			// scmd.WriteString("--write-subs --write-auto-subs --sub-langs 'en,en-US,zh-Hant,zh-Hans' ")
			// scmd.WriteString("--mark-watched ")
			// scmd.WriteString("--youtube-skip-dash-manifest ")
			scmd.WriteString("--skip-unavailable-fragments ")
			// scmd.WriteString("--abort-on-unavailable-fragment ")
			scmd.WriteString("--no-mtime ")
			scmd.WriteString("--buffer-size 512k ")
			scmd.WriteString("--sleep-requests 1 ")
			scmd.WriteString("--sleep-interval 5 ")
			// scmd.WriteString("--recode-video mp4 ")
			scmd.WriteString("-o '" + ytdir + videoName + ".%(ext)s' ")
			if strings.HasPrefix(vi.url, "http") {
				scmd.WriteString("'" + vi.url + "'")
			} else {
				scmd.WriteString("-- " + vi.url)
			}
			scmd.WriteString(" && \\\nrm $0\n")
			os.WriteFile(shellName, scmd.Bytes(), 0o755)
			// DOWN:
			time.Sleep(time.Second * time.Duration(rand.Int31n(5)+10))
			cmd = exec.Command(shellName)
			// cmd.Env = os.Environ()
			b, err := cmd.CombinedOutput()
			if err != nil {
				b = append(b, []byte("\n"+err.Error()+"\n")...)
				os.WriteFile(shellName+".log", b, 0o664)
			}
			time.Sleep(time.Second * time.Duration(rand.Int31n(20)+20))
			if pathtool.IsExist(shellName) {
				out := strings.ToLower(string(b))
				if strings.Contains(out, "error") ||
					strings.Contains(out, "errno") ||
					strings.Contains(out, "filename too long") {
					// vi.try++
				}
				// chanYoutubeDownloader <- vi
			} else {
				os.Remove(shellName + ".log")
			}
		}
	}()
	time.Sleep(time.Second)
	goto RUN
}

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/xyzj/gopsu"
)

var (
	mainDomain  = "wgq.shwlst.com:40001"
	debugDomain = "v4.xyzjdays.xyz"

	urlDownload    = "https://%s/cert/download/%s"
	domainList     = []string{"wlst.vip"} //, "shwlst.com"}
	linuxSSLCopy   = filepath.Join(gopsu.GetExecDir(), "sslcopy.sh")
	windowsSSLCopy = filepath.Join(gopsu.GetExecDir(), "sslcopy.bat")

	dlog *log.Logger

	enableDebug = flag.Bool("debug", false, "set if enable debug info.")
	httpClient  = &http.Client{
		Timeout: time.Duration(time.Second * 300),
		Transport: &http.Transport{
			IdleConnTimeout: time.Minute,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
)

func getExecDir() string {
	a, _ := os.Executable()
	execdir := filepath.Dir(a)
	if strings.Contains(execdir, "go-build") {
		execdir, _ = filepath.Abs(".")
	}
	return execdir
}
func downloadCert(domain string) bool {
	p := filepath.Join(getExecDir(), "domain"+".zip")
	req, _ := http.NewRequest("GET", fmt.Sprintf(urlDownload, mainDomain, domain), strings.NewReader(""))
	resp, err := httpClient.Do(req)
	if err != nil {
		dlog.Println("download error:" + err.Error())
		return false
	}
	defer resp.Body.Close()
	f, err := os.Create(p)
	if err != nil {
		dlog.Println("save error:" + err.Error())
		return false
	}
	defer f.Close()
	io.Copy(f, resp.Body)
	err = gopsu.UnZIPFile(p, "ca")
	if err != nil {
		dlog.Println("unzip file error:" + err.Error())
	}
	dlog.Println("Download success. start copy ...")
	return true
}

func renew() {
	// var oldsign, newsign string
	for _, v := range domainList {
		err := os.Remove(filepath.Join(getExecDir(), v+".zip"))
		if err != nil {
			dlog.Println("clean zip files error: " + err.Error())
		}
		// oldsign = localSign(v)
		// newsign = remoteSign(v)
		// if newsign != "1" && oldsign != newsign {
		downloadCert(v)
		// } else {
		// 	if oldsign == newsign {
		// 		dlog.Println("Same signature, no update needed.")
		// 	} else {
		// 		dlog.Println("Can not get cert file info:" + v)
		// 	}
		println("-------")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin":
		if gopsu.IsExist(linuxSSLCopy) {
			cmd = exec.Command(linuxSSLCopy)
		} else {
			dlog.Println("no sslcopy found")
		}
	case "windows":
		if gopsu.IsExist(windowsSSLCopy) {
			cmd = exec.Command(windowsSSLCopy)
		} else {
			dlog.Println("no sslcopy found")
		}
	}
	if cmd != nil {
		err := cmd.Run()
		if err != nil {
			dlog.Println("run sslcopy error:" + err.Error())
		} else {
			dlog.Println("do sslcopy done")
		}
	}
	dlog.Println("All Done.")
}

func main() {
	fd, _ := os.OpenFile("sslrenew.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	dlog = log.New(io.MultiWriter(fd, os.Stdout), "", log.LstdFlags)
	flag.Parse()
	if *enableDebug {
		mainDomain = debugDomain
		domainList = []string{"xyzjdays.xyz"}
	}
	rand.Seed(time.Now().UnixNano())
	renew()
	for {
		time.Sleep(time.Hour * 3)
		if time.Now().Hour() < 3 {
			renew()
		}
	}
}

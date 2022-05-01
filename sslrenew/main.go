package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
	enableDebug = flag.Bool("debug", false, "set if enable debug info.")
	show        = flag.String("show", "", "show crt file info")
)

var (
	mainDomain  = []string{"wgq.shwlst.com:40001"}
	debugDomain = "v4.xyzjdays.xyz"

	urlDownload    = "https://%s/cert/download/%s"
	domainList     = []string{"wlst.vip"} //, "shwlst.com"}
	linuxSSLCopy   = filepath.Join(gopsu.GetExecDir(), "sslcopy.sh")
	windowsSSLCopy = filepath.Join(gopsu.GetExecDir(), "sslcopy.bat")

	dlog *log.Logger

	httpClient = &http.Client{
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
func downloadCert(url, domain string) bool {
	p := filepath.Join(getExecDir(), domain+".zip")
	req, _ := http.NewRequest("GET", fmt.Sprintf(urlDownload, url, domain), strings.NewReader(""))
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
	err = gopsu.UnZIPFile(p, filepath.Join(getExecDir(), "ca"))
	if err != nil {
		dlog.Println("unzip file error:" + err.Error())
	}
	dlog.Println("Download " + domain + " success. start copy ...")
	return true
}

func renew() {
	// var oldsign, newsign string
	for k, v := range domainList {
		err := os.Remove(filepath.Join(getExecDir(), v+".zip"))
		if err != nil {
			dlog.Println("clean zip files error: " + err.Error())
		}
		// oldsign = localSign(v)
		// newsign = remoteSign(v)
		// if newsign != "1" && oldsign != newsign {
		downloadCert(mainDomain[k], v)
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
	flag.Parse()
	if *show != "" {
		//加载PEM格式证书到字节数组
		certPEMBlock, err := ioutil.ReadFile(*show + ".crt")
		if err != nil {
			println(err.Error())
			return
		}
		certblock, _ := pem.Decode(certPEMBlock)
		if certblock == nil {
			println("can not decode")
			return
		}
		x509crt, err := x509.ParseCertificate(certblock.Bytes)
		if err != nil {
			println(err.Error())
			return
		}
		println(x509crt.NotAfter.Format("2006-01-02 15:04:05"), fmt.Sprintf("%+v", x509crt.DNSNames))
		return
	}
	fd, _ := os.OpenFile(filepath.Join(getExecDir(), "sslrenew.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	dlog = log.New(io.MultiWriter(fd, os.Stdout), "", log.LstdFlags)
	if *enableDebug {
		mainDomain = append(mainDomain, debugDomain)
		domainList = append(domainList, "xyzjdays.xyz")
	}
	renew()
	for {
		time.Sleep(time.Hour * 3)
		if time.Now().Hour() < 3 {
			renew()
		}
	}
}

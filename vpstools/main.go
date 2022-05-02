package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"
	"unsafe"

	"github.com/xyzj/gopsu"

	ginmiddleware "github.com/xyzj/gopsu/gin-middleware"
)

var (
	version     = "0.0.0"
	goVersion   = ""
	buildDate   = ""
	platform    = ""
	author      = "Xu Yuan"              // 你的名字
	programName = "golang micro service" // 服务名称
)

type sliceFlag []string

func (f *sliceFlag) String() string {
	return fmt.Sprintf("%v", []string(*f))
}

func (f *sliceFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

var (
	port  = flag.Int("http", 2082, "set http port")
	ports = flag.Int("https", 0, "set https port")
	cert  = flag.String("cert", "", "cert file path")
	key   = flag.String("key", "", "key file path")
	auth  = flag.Bool("auth", false, "enable basicauth")
	debug = flag.Bool("debug", false, "run in debug mode")
	// 帮助信息
	help = flag.Bool("help", false, "print help message and exit")
)

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
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	go func() {
		ioutil.WriteFile(gopsu.JoinPathFromHere("cfrenew.log"), []byte(certCloudflareTools("renew")), 0664)
		time.Sleep(time.Hour * 240)
	}()
	shortconf, _ = gopsu.LoadConfig(gopsu.JoinPathFromHere("short.conf"))
	opt := &ginmiddleware.ServiceOption{
		HTTPPort:   *port,
		HTTPSPort:  *ports,
		CertFile:   *cert,
		KeyFile:    *key,
		Debug:      *debug,
		EngineFunc: routeEngine,
	}
	ginmiddleware.ListenAndServeWithOption(opt)
}

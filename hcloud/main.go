package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/xyzj/gopsu"
	"github.com/xyzj/gopsu/gocmd"

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
	html  = flag.String("html", "", "something like nginx")
	ydir  = flag.String("ydir", "", "set youtube download dir")
	aria2 = flag.String("aria2", "", "set aria2 json rpc url")
	// 帮助信息
	help = flag.Bool("help", false, "print help message and exit")
	dirs gopsu.SliceFlag
	wtv  gopsu.SliceFlag
	dav  gopsu.SliceFlag
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
	// http 静态目录
	flag.Var(&dirs, "dir", "example: -dir=name:path -dir name2:path2")
	// web gallery 目录
	flag.Var(&wtv, "wtv", "example: -wtv=name:path -wtv name2:path2")
	// webdav目录
	flag.Var(&dav, "dav", "example: -dav=name:path -dav name2:path2")
	gocmd.DefaultProgram(&gocmd.Info{Title: "home cloud", Ver: "0.1.0"}).ExecuteDefault("run")
	// flag.Parse()
	// if *help {
	// 	flag.PrintDefaults()
	// 	os.Exit(0)
	// }
	// godaemon.Start(nil)
	// 参数处理
	if *aria2 == "" {
		*aria2 = "http://127.0.0.1:2052"
	}
	if *ydir == "" {
		*ydir = "/home/xy/mm/tv/news/"
	}
	if !strings.HasSuffix(*ydir, "/") {
		*ydir += "/"
	}
	// 启动youtube下载控制
	for i := 0; i < 2; i++ {
		go youtubeControl()
	}

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

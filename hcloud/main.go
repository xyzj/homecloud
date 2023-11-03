package main

import (
	"flag"
	"fmt"
	lib "homecloud/lib"

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

var (
	port  = flag.Int("http", 2082, "set http port")
	ports = flag.Int("https", 0, "set https port")
	cert  = flag.String("cert", "", "cert file path")
	key   = flag.String("key", "", "key file path")
	debug = flag.Bool("debug", false, "run in debug mode")
)

func main() {
	gocmd.DefaultProgram(&gocmd.Info{Title: "home cloud", Ver: "0.1.0"}).ExecuteRun()
	// flag.Parse()
	// if *help {
	// 	flag.PrintDefaults()
	// 	os.Exit(0)
	// }
	// godaemon.Start(nil)
	// 启动youtube下载控制
	for i := 0; i < 2; i++ {
		go lib.YoutubeControl()
	}

	opt := &ginmiddleware.ServiceOption{
		HTTPPort:   fmt.Sprintf(":%d", *port),
		HTTPSPort:  fmt.Sprintf(":%d", *ports),
		CertFile:   *cert,
		KeyFile:    *key,
		Debug:      *debug,
		EngineFunc: lib.RouteEngine,
	}
	ginmiddleware.ListenAndServeWithOption(opt)
}

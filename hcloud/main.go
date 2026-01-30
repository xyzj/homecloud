package main

import (
	"flag"
	"fmt"
	lib "homecloud/lib"
	"time"

	gocmd "github.com/xyzj/go-cmd"
	"github.com/xyzj/toolbox/crypto"
	ginmiddleware "github.com/xyzj/toolbox/ginmiddle"
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
	gocmd.DefaultProgram(&gocmd.Info{
		Title: "home cloud",
		Ver:   "0.1.0",
		Descript: gocmd.PrintVersion(&gocmd.VersionInfo{
			Version:   version,
			GoVersion: goVersion,
			BuildDate: buildDate,
			BuildOS:   platform,
			CodeBy:    author,
			Name:      programName,
		})}).ExecuteRun()
	// flag.Parse()
	// if *help {
	// 	flag.PrintDefaults()
	// 	os.Exit(0)
	// }
	// godaemon.Start(nil)
	// 启动youtube下载控制
	go lib.YoutubeControl()
	go func() {
		t := time.NewTicker(time.Minute * 30)
		for {
			<-t.C
			lib.CheckAria2cActive()
		}
	}()
	var hport, hsport string
	if *port > 0 {
		hport = fmt.Sprintf(":%d", *port)
	}
	if *ports > 0 {
		hsport = fmt.Sprintf(":%d", *ports)
	}
	tlsc, err := crypto.TLSConfigFromFile(*cert, *key, "")
	if err != nil {
		fmt.Println("load tls config error:", err.Error())
	}
	ginmiddleware.ListenAndServeWithOption(ginmiddleware.OptHTTP(hport),
		ginmiddleware.OptHTTPS(hsport, tlsc),
		ginmiddleware.OptDebug(*debug),
		ginmiddleware.OptEngine(lib.RouteEngine()),
	)
}

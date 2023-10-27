package main

import (
	"flag"
	"os"
	"time"
	"unsafe"

	"github.com/xyzj/gopsu"
	ginmiddleware "github.com/xyzj/gopsu/gin-middleware"
	"github.com/xyzj/gopsu/gocmd"
	"github.com/xyzj/gopsu/loopfunc"
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
	port  = flag.String("http", ":2082", "set http port")
	ports = flag.String("https", "", "set https port")
	cert  = flag.String("cert", "", "cert file path")
	key   = flag.String("key", "", "key file path")
	debug = flag.Bool("debug", false, "run in debug mode")
	renew = flag.Bool("renew", false, "auto renew cert files for domain")
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
	gocmd.DefaultProgram(&gocmd.Info{Ver: "1.0.0"}).ExecuteDefault("start")

	if *renew {
		go loopfunc.LoopFunc(func(params ...interface{}) {
			for {
				os.WriteFile(gopsu.JoinPathFromHere("cfrenew.log"), []byte(certCloudflareTools("renew")), 0664)
				time.Sleep(time.Hour * 216)
			}
		}, "cert renew", nil)
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

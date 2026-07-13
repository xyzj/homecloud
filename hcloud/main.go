package main

import (
	"flag"
	"fmt"

	lib "homecloud/lib"

	"github.com/gin-gonic/gin"
	gocmd "github.com/xyzj/go-cmd"
	ginmiddleware "github.com/xyzj/toolbox/ginmiddle"
	"github.com/xyzj/toolbox/pathtool"
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
	aria2 = flag.String("aria2", "", "set aria2 json rpc url")
	ydir  = flag.String("ytdir", "", "set youtube download dir")
	wtv   pathtool.SliceFlag
)

func main() {
	gin.SetMode("release")
	flag.Var(&wtv, "wtv", "example: -wtv=name:path -wtv name2:path2")
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
	lib.SetAria2cUrl(*aria2)
	lib.SetYtdir(*ydir)
	go lib.YoutubeControl()
	go lib.CheckAria2cActive()
	var hport string
	if *port > 0 {
		hport = fmt.Sprintf(":%d", *port)
	}
	ginmiddleware.ListenAndServeWithOption(ginmiddleware.OptHTTP(hport),
		ginmiddleware.OptEngine(lib.RouteEngine(wtv...)),
		ginmiddleware.OptDebug(false),
	)
}

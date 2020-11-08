package main

import (
	"runtime"
	"time"

	lib "./lib"
	"github.com/xyzj/gopsu"
)

var (
	version     = "0.0.0"
	goVersion   = ""
	buildDate   = ""
	platform    = ""
	author      = "Xu Yuan"              // 你的名字
	programName = "golang micro service" // 服务名称
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	lib.Run(gopsu.VersionInfo(programName, version, goVersion, platform, buildDate, author))
	// 主线程把所有模块以微线程的方式启动，自己可以啥都不干，睡大觉
	for {
		time.Sleep(time.Minute)
	}
}

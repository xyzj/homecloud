package main

import (
	"flag"
	"fmt"
	"os"
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

var (
	// 启动参数
	ver         = flag.Bool("version", false, "print version info and exit.")
	enableDebug = flag.Bool("debug", false, "set if enable debug info.")
	web         = flag.Int("http", 6819, "set http port to listen on.")
	domain      = flag.String("domain", "xyzjdays.xyz", "set domain name.")
	conf        = flag.String("conf", "", "set the config file path.")
)

// 程序运行后在程序目录中生成一个版本信息文件
func writeVersionInfo() {
	s := gopsu.VersionInfo(programName, version, goVersion, platform, buildDate, author)
	lib.Version = s
	p, _ := os.Executable()
	f, _ := os.OpenFile(fmt.Sprintf("%s.ver", p), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0444)
	defer f.Close()
	f.WriteString(s + "\r\n")
}

// 处理启动参数
func parseArguments() {
	flag.Parse()

	if *ver {
		println(gopsu.VersionInfo(programName, version, goVersion, platform, buildDate, author))
		os.Exit(1)
	}
	// 使用多核cpu
	runtime.GOMAXPROCS(runtime.NumCPU())
	lib.EnableDebug = *enableDebug
	lib.DomainName = *domain

	if *conf == "" {
		println("no config file set.")
		os.Exit(21)
	}
}

func startModes() {
	// 载入除标准配置外的额外配置信息（使用相同的配置文件）
	lib.LoadExtConfigure(*conf)
	// 启动http(s)服务
	// 输入参数依次为，http端口号，日志保存天数（用于控制日志文件数量）
	lib.NewHTTPService(*web)
}

func main() {
	writeVersionInfo()
	runtime.GOMAXPROCS(runtime.NumCPU())

	parseArguments()

	startModes()
	// 主线程把所有模块以微线程的方式启动，自己可以啥都不干，睡大觉
	for {
		time.Sleep(time.Minute)
	}
}

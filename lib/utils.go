package lib

import (
	"crypto/tls"
	"embed"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	ginmiddleware "github.com/xyzj/gopsu/gin-middleware"

	"github.com/xyzj/gopsu"
)

//go:embed static
var stat embed.FS

//go:embed ca/localhost.pem
var caCert []byte

//go:embed ca/localhost-key.pem
var caKey []byte

var (
	// Version 版本信息
	Version        string
	ipCached       string
	urlConf        *gopsu.ConfData
	ydir           string
	httpClientPool = &http.Client{
		Timeout: time.Duration(time.Second * 15),
		Transport: &http.Transport{
			IdleConnTimeout:     time.Second * 15,
			MaxConnsPerHost:     10,
			MaxIdleConns:        1,
			MaxIdleConnsPerHost: 1,
			DisableCompression:  true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
)
var (
	ver       = flag.Bool("version", false, "print version info and exit.")
	help      = flag.Bool("help", false, "print help and exit.")
	debug     = flag.Bool("debug", false, "set if enable debug info.")
	vauth     = flag.Bool("vauth", false, "set if enable http auth for video group.")
	filelogon = flag.Bool("logfile", false, "set if enable log to a file.")
	web       = flag.Int("http", 0, "set http port to listen on.")
	webs      = flag.Int("https", 0, "set https port to listen on.")
	certfile  = flag.String("cert", "./ca/xyzjdays.xyz.crt", "set cert file path.")
	keyfile   = flag.String("key", "./ca/xyzjdays.xyz.key", "set key file path.")
	conf      = flag.String("conf", "", "set the config file path.")
)

// LoadExtConfigure 载入除标准配置外的自定义配置内容（可选）
func LoadExtConfigure(f string) {
	var err error
	urlConf, err = gopsu.LoadConfig(f)
	if err != nil {
		println("Load configure file error:" + err.Error())
	} else {
		ydir, _ = urlConf.GetItem("ydl_dir")
		if ydir != "" && !strings.HasSuffix(ydir, "/") {
			ydir += "/"
		}
	}
	// domainList = strings.Split(urlConf.GetItemDefault("dnspod_list", "wlst.vip,shwlst.com", "要管理的dnspod域名列表"), ",")
	// urlConf.Save()
	b, _ := stat.ReadFile("static/tpl/webtv.html")
	if len(b) > 0 {
		pageWebTV = gopsu.String(b)
	} else {
		pageWebTV = loadWebTVPage()
	}
}

func loadWebTVPage() string {
	// html页面
	b, err := ioutil.ReadFile(gopsu.JoinPathFromHere("static", "tpl", "webtv.html"))
	if err != nil {
		b, _ = stat.ReadFile("static/tpl/webtv.html")
	}
	var webPage = string(b)
	//videojs css
	b, err = ioutil.ReadFile(gopsu.JoinPathFromHere("static", "css", "video-js.min.css"))
	if err == nil {
		webPage = strings.Replace(webPage, `<link rel="stylesheet" href="/static/css/video-js.min.css" />`, "<style>"+string(b)+"</style>", 1)
	}
	//playlist-ui css
	b, err = ioutil.ReadFile(gopsu.JoinPathFromHere("static", "css", "videojs-playlist-ui.custom.css"))
	if err == nil {
		webPage = strings.Replace(webPage, `<link rel="stylesheet" href="/static/css/videojs-playlist-ui.custom.css" />`, "<style>"+string(b)+"</style>", 1)
	}
	//video.min.js
	b, err = ioutil.ReadFile(gopsu.JoinPathFromHere("static", "js", "videojs", "video.min.js"))
	if err == nil {
		webPage = strings.Replace(webPage, `<script src="/static/js/videojs/video.min.js"></script>`, "<script>"+string(b)+"</script>", 1)
	}
	//videojs.hotkeys.min.js
	b, err = ioutil.ReadFile(gopsu.JoinPathFromHere("static", "js", "videojs-hotkeys", "videojs.hotkeys.min.js"))
	if err == nil {
		webPage = strings.Replace(webPage, `<script src="/static/js/videojs-hotkeys/videojs.hotkeys.min.js"></script>`, "<script>"+string(b)+"</script>", 1)
	}
	//videojs-playlist.min.js
	b, err = ioutil.ReadFile(gopsu.JoinPathFromHere("static", "js", "videojs-playlist", "videojs-playlist.min.js"))
	if err == nil {
		webPage = strings.Replace(webPage, `<script src="/static/js/videojs-playlist/videojs-playlist.min.js"></script>`, "<script>"+string(b)+"</script>", 1)
	}
	//videojs-flash.js
	b, err = ioutil.ReadFile(gopsu.JoinPathFromHere("static", "js", "videojs-flash", "videojs-flash.min.js"))
	if err == nil {
		webPage = strings.Replace(webPage, `<script src="/static/js/videojs-flash/videojs-flash.min.js"></script>`, "<script>"+string(b)+"</script>", 1)
	}
	//videojs-playlist-ui.js
	b, err = ioutil.ReadFile(gopsu.JoinPathFromHere("static", "js", "videojs-playlist", "videojs-playlist-ui.js"))
	if err == nil {
		webPage = strings.Replace(webPage, `<script src="/static/js/videojs-playlist/videojs-playlist-ui.js"></script>`, "<script>"+string(b)+"</script>", 1)
	}
	//zh-TW.js
	b, err = ioutil.ReadFile(gopsu.JoinPathFromHere("static", "js", "videojs", "lang", "zh-TW.js"))
	if err == nil {
		webPage = strings.Replace(webPage, `<script src="/static/js/videojs/lang/zh-TW.js"></script>`, "<script>"+string(b)+"</script>", 1)
	}

	return webPage
}

// Run run
func Run(version string) {
	flag.Parse()
	Version = version
	if *ver {
		println(Version)
		os.Exit(0)
	}
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	if !gopsu.IsExist(*certfile) || !gopsu.IsExist(*keyfile) {
		os.MkdirAll(filepath.Join(gopsu.GetExecDir(), "ca"), 0755)
		*certfile = filepath.Join(gopsu.GetExecDir(), "ca", "localhost.pem")
		*keyfile = filepath.Join(gopsu.GetExecDir(), "ca", "localhost-key.pem")
		ioutil.WriteFile(*certfile, caCert, 0644)
		ioutil.WriteFile(*keyfile, caKey, 0644)
	}
	LoadExtConfigure(*conf)

	// 启动youtube下载控制
	for i := 0; i < 3; i++ {
		// go httpControl()
		go youtubeControl()
	}
	// 启动http服务
	opt := &ginmiddleware.ServiceOption{
		HTTPPort:   *web,
		HTTPSPort:  *webs,
		CertFile:   *certfile,
		KeyFile:    *keyfile,
		Debug:      *debug,
		EngineFunc: routeEngine,
	}
	ginmiddleware.ListenAndServeWithOption(opt)
}

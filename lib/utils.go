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

	"github.com/xyzj/gopsu"
)

//go:embed static
var stat embed.FS

//go:embed ca/localhost.pem
var caCert []byte

//go:embed ca/localhost-key.pem
var caKey []byte

const (
	bwhStatusURL = "https://api.64clouds.com/v1/getServiceInfo?veid=%s&api_key=%s"
	bwhAPIKey    = "yfCUSxAg5fs9DMzQntChzNkPneEsvMm5bMo+iuDt9Zr0itwcP3vSrMDOfeCovNA0igyKy2z1bKy8CxsQTYCNexa"
	bwhVeid      = "979913"
	// dnspod sslrenew token
	// dnspodID    = "141155"
	// dnspodToken = "076ba7af12e110fb5c2eebc438dae5a1"
	// cloudflare
	// cfKey  = "b6c9de4a9814d534ab16c12d99718f118fde2"
	// cfZone = "fb8a871c3737648dfd964bd625f9f685"
	// cfID   = "712df327b64333800c02511f404b3157"
)

var (
	// Version 版本信息
	Version        string
	crtFile        string
	keyFile        string
	ipCached       string
	urlConf        *gopsu.ConfData
	ydir           string
	tdir           string
	httpClientPool = &http.Client{
		Timeout: time.Duration(time.Second * 15),
		Transport: &http.Transport{
			IdleConnTimeout:     time.Second * 15,
			MaxConnsPerHost:     10,
			MaxIdleConns:        1,
			MaxIdleConnsPerHost: 1,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
)
var (
	ver    = flag.Bool("version", false, "print version info and exit.")
	help   = flag.Bool("help", false, "print help and exit.")
	debug  = flag.Bool("debug", false, "set if enable debug info.")
	vauth  = flag.Bool("vauth", false, "set if enable http auth for video group.")
	web    = flag.Int("http", 0, "set http port to listen on.")
	webs   = flag.Int("https", 0, "set https port to listen on.")
	domain = flag.String("domain", "xyzjdays.xyz", "set domain name.")
	conf   = flag.String("conf", "", "set the config file path.")
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
		tdir, _ = urlConf.GetItem("tdl_dir")
		if tdir != "" && !strings.HasSuffix(tdir, "/") {
			tdir += "/"
		}
	}
	// domainList = strings.Split(urlConf.GetItemDefault("dnspod_list", "wlst.vip,shwlst.com", "要管理的dnspod域名列表"), ",")
	// urlConf.Save()
	pageWebTV = loadWebTVPage()
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
	if *web < 1000 && *webs < 1000 {
		*web = 6819
	}
	crtFile = filepath.Join(gopsu.GetExecDir(), "ca", *domain+".crt")
	keyFile = filepath.Join(gopsu.GetExecDir(), "ca", *domain+".key")
	if !gopsu.IsExist(crtFile) || !gopsu.IsExist(keyFile) {
		crtFile = filepath.Join(gopsu.GetExecDir(), "ca", "localhost.pem")
		keyFile = filepath.Join(gopsu.GetExecDir(), "ca", "localhost-key.pem")
		os.MkdirAll(filepath.Join(gopsu.GetExecDir(), "ca"), 0755)
		ioutil.WriteFile(crtFile, caCert, 0644)
		ioutil.WriteFile(keyFile, caKey, 0644)
	}
	LoadExtConfigure(*conf)
	go NewHTTPService()
	// 启动youtube下载控制
	for i := 0; i < 3; i++ {
		// go httpControl()
		go youtubeControl()
	}

}

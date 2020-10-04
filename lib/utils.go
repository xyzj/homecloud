package lib

import (
	"flag"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/xyzj/gopsu"
)

const (
	bwhStatusURL = "https://api.64clouds.com/v1/getServiceInfo?veid=%s&api_key=%s"
	bwhAPIKey    = "yfCUSxAg5fs9DMzQntChzNkPneEsvMm5bMo+iuDt9Zr0itwcP3vSrMDOfeCovNA0igyKy2z1bKy8CxsQTYCNexa"
	bwhVeid      = "979913"
	// dnspod sslrenew token
	dnspodID    = "141155"
	dnspodToken = "076ba7af12e110fb5c2eebc438dae5a1"
)

var (
	// EnableDebug 显示debug调试信息
	EnableDebug bool
	// Version 版本信息
	Version string
	// DomainName 域名
	DomainName string

	ipCached       string
	urlConf        *gopsu.ConfData
	linuxSSLCopy   = filepath.Join(gopsu.GetExecDir(), "sslcopy.sh")
	windowsSSLCopy = filepath.Join(gopsu.GetExecDir(), "sslcopy.bat")
	domainList     = []string{"wlst.vip,shwlst.com"}
	ydir           string
)
var (
	forceHTTP = flag.Bool("forcehttp", false, "set if run as http")
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
	pageWebTV = loadWebTVPage()
}

func loadWebTVPage() string {
	var webPage string
	// html页面
	b, err := ioutil.ReadFile(gopsu.JoinPathFromHere("static", "tpl", "webtv.html"))
	if err != nil {
		return tplVideojs
	}
	webPage = string(b)
	//videojs css
	b, err = ioutil.ReadFile(gopsu.JoinPathFromHere("static", "css", "video-js.min.css"))
	if err != nil {
		return tplVideojs
	}
	webPage = strings.Replace(webPage, `<link rel="stylesheet" href="/static/css/video-js.min.css" />`, "<style>"+string(b)+"</style>", 1)
	//playlist-ui css
	b, err = ioutil.ReadFile(gopsu.JoinPathFromHere("static", "css", "videojs-playlist-ui.custom.css"))
	if err != nil {
		return tplVideojs
	}
	webPage = strings.Replace(webPage, `<link rel="stylesheet" href="/static/css/videojs-playlist-ui.custom.css" />`, "<style>"+string(b)+"</style>", 1)
	//video.min.js
	b, err = ioutil.ReadFile(gopsu.JoinPathFromHere("static", "js", "videojs", "video.min.js"))
	if err != nil {
		return tplVideojs
	}
	webPage = strings.Replace(webPage, `<script src="/static/js/videojs/video.min.js"></script>`, "<script>"+string(b)+"</script>", 1)
	//videojs.hotkeys.min.js
	b, err = ioutil.ReadFile(gopsu.JoinPathFromHere("static", "js", "videojs-hotkeys", "videojs.hotkeys.min.js"))
	if err != nil {
		return tplVideojs
	}
	webPage = strings.Replace(webPage, `<script src="/static/js/videojs-hotkeys/videojs.hotkeys.min.js"></script>`, "<script>"+string(b)+"</script>", 1)
	//videojs-playlist.min.js
	b, err = ioutil.ReadFile(gopsu.JoinPathFromHere("static", "js", "videojs-playlist", "videojs-playlist.min.js"))
	if err != nil {
		return tplVideojs
	}
	webPage = strings.Replace(webPage, `<script src="/static/js/videojs-playlist/videojs-playlist.min.js"></script>`, "<script>"+string(b)+"</script>", 1)
	//videojs-flash.js
	b, err = ioutil.ReadFile(gopsu.JoinPathFromHere("static", "js", "videojs-flash", "videojs-flash.js"))
	if err != nil {
		return tplVideojs
	}
	webPage = strings.Replace(webPage, `<script src="/static/js/videojs-flash/videojs-flash.js"></script>`, "<script>"+string(b)+"</script>", 1)
	//videojs-playlist-ui.js
	b, err = ioutil.ReadFile(gopsu.JoinPathFromHere("static", "js", "videojs-playlist", "videojs-playlist-ui.js"))
	if err != nil {
		return tplVideojs
	}
	webPage = strings.Replace(webPage, `<script src="/static/js/videojs-playlist/videojs-playlist-ui.js"></script>`, "<script>"+string(b)+"</script>", 1)
	//zh-TW.js
	b, err = ioutil.ReadFile(gopsu.JoinPathFromHere("static", "js", "videojs", "lang", "zh-TW.js"))
	if err != nil {
		return tplVideojs
	}
	webPage = strings.Replace(webPage, `<script src="/static/js/videojs/lang/zh-TW.js"></script>`, "<script>"+string(b)+"</script>", 1)

	return webPage
}

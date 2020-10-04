package lib

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/xyzj/gopsu"
	ginmiddleware "github.com/xyzj/gopsu/gin-middleware"
)

// 该文件为
// statusCheck 状态检查
func statusCheck(c *gin.Context) {
	var d = gin.H{
		"timer": gopsu.Stamp2Time(time.Now().Unix()),
		"ver":   strings.Split(Version, "\n")[1:],
	}

	switch c.Request.Method {
	case "GET":
		c.HTML(200, "status", d)
	case "POST":
		c.PureJSON(200, d)
	}
}

func test(c *gin.Context) {
	gopsu.ZIPFiles("aaa.zip", []string{".ipcache", "build.py", "hcloud.conf"}, "ca")
}

// multiRender 预置模板
func multiRender() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromString("vpsinfo", tplVpsinfo)
	r.AddFromString("uuidinfo", tpluuid)
	return r
}

// NewHTTPService NewHTTPService
func NewHTTPService(port int) {
	if !EnableDebug { // 设置框架使用的模式，默认debug模式，会有控制台输出，release模式仅有文件日志输出
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(ginmiddleware.Recovery())

	// 渲染模板
	r.HTMLRender = multiRender()

	r.Static("/static", gopsu.JoinPathFromHere("static"))

	r.GET("/", remoteIP)
	r.POST("/", remoteIP)
	r.GET("/reloadext", func(c *gin.Context) {
		urlConf.Reload()
		pageWebTV = loadWebTVPage()
		c.String(200, "done")
	})
	r.GET("/givemenewuuid4", newUUID4)
	// kod共享
	r.GET("/m/:name", func(c *gin.Context) {
		urlConf.Reload()
		n, err := urlConf.GetItem(c.Param("name"))
		if err != nil {
			c.String(200, err.Error())
			return
		}
		s := "https://kod.xyzjdays.xyz:10043/index.php?share/" + gopsu.DecodeString(n)
		c.Redirect(http.StatusTemporaryRedirect, s)
	})
	r.GET("/share/add", ginmiddleware.ReadParams(), func(c *gin.Context) {
		if !urlConf.SetItem(c.Param("src"), strings.ReplaceAll(c.Param("dst"), " ", "+"), "kod share") {
			c.String(200, "failed")
			return
		}
		urlConf.Save()
		c.String(200, "ok")
	})
	r.GET("/share/del", ginmiddleware.ReadParams(), func(c *gin.Context) {
		urlConf.DelItem(c.Param("src"))
		urlConf.Save()
		c.String(200, "ok")
	})
	r.GET("/share/query", func(c *gin.Context) {
		c.String(200, urlConf.GetAll())
	})
	// 视频播放
	r.GET("/v/:dir", ginmiddleware.ReadParams(), runVideojs)
	keys := urlConf.GetKeys()
	for _, key := range keys {
		if strings.HasPrefix(key, "tv-") {
			v, _ := urlConf.GetItem(key)
			if v == "" {
				continue
			}
			r.Static("/"+key, v)
		}
	}
	// vps相关
	g1 := r.Group("/vps")
	g1.GET("v4info", vps4info)
	// homecloud服务
	g2 := r.Group("/wt")
	g2.GET("/", ipCache)
	g2.GET("/:name", wt)
	// 公司跳转
	g3 := r.Group("/soho")
	g3.GET("/", func(c *gin.Context) {
		c.String(200, "180.167.245.233")
	})
	g3.GET("/kod", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "http://office.shwlst.com:20080")
	})
	g3.GET("/zd", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "http://office.shwlst.com:5990")
	})
	// 证书管理
	g4 := r.Group("/cert", ginmiddleware.ReadParams())
	g4.GET("/sign/:name", certSign)
	g4.GET("/download/:name", certDownload)
	g4.GET("/namesilo/:do", certNamesilo)
	g4.GET("/dnspod/:do", certDNSPod)
	g4.GET("/cloudflare/:do", certCloudflare)
	// 工具
	g5 := r.Group("/tools")
	g5.GET("/codestr", codeString)
	g5.POST("/codestr", ginmiddleware.ReadParams(), codeString)
	g5.GET("/ydl", ginmiddleware.ReadParams(), ydl)
	g5.GET("/ydlb", ydlb)
	g5.POST("/ydlb", ginmiddleware.ReadParams(), ydlb)

	r.HandleMethodNotAllowed = true
	r.NoMethod(ginmiddleware.Page405)
	r.NoRoute(ginmiddleware.Page404)

	// 在微线程中启动服务
	go func() {
		var err error
		println("Starting HTTP(S) server at :" + strconv.Itoa((port)))
		if EnableDebug || *forceHTTP { // 调试模式下使用http
			err = ginmiddleware.ListenAndServe(port, r)
		} else { // 生产模式下使用https,若设置了clientca，则会验证客户端证书
			err = ginmiddleware.ListenAndServeTLS(port, r, filepath.Join(".", "ca", DomainName+".crt"), filepath.Join(".", "ca", DomainName+".key"), "")
		}
		if err != nil {
			println("Failed start HTTP(S) server at :" + strconv.Itoa(port) + "|" + err.Error())
		}
		os.Exit(1)
	}()
	// 启动youtube下载控制
	for i := 0; i < 7; i++ {
		go downloadControl()
	}
}

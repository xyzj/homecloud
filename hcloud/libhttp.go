package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	game "github.com/xyzj/gopsu/games"
	ginmiddleware "github.com/xyzj/gopsu/gin-middleware"
)

var (
	ipCached       string
	ipCached6      string
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

// multiRender 预置模板
func multiRender() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromString("vpsinfo", tplVpsinfo)
	r.AddFromString("uuidinfo", tpluuid)
	return r
}

func routeEngine() *gin.Engine {
	r := gin.New()
	// 特殊路由处理
	r.HandleMethodNotAllowed = true
	r.NoMethod(ginmiddleware.Page405)
	r.NoRoute(ginmiddleware.Page404)
	// 允许跨域
	r.Use(cors.New(cors.Config{
		MaxAge:           time.Hour * 24,
		AllowWebSockets:  true,
		AllowCredentials: true,
		AllowWildcard:    true,
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
	}))
	// 渲染模板
	r.HTMLRender = multiRender()

	if *html != "" {
		r.Static("/html", *html)
		r.GET("/index", func(c *gin.Context) {
			c.Redirect(http.StatusTemporaryRedirect, "/html/index.html")
		})
	}
	// 主页
	r.GET("/", remoteIP)
	r.GET("/game/:game", game.GameGroup)
	// 静态资源
	r.StaticFS("/emb", http.FS(stat))

	// 工具路由
	// vps信息
	r.GET("/tools/vpsinfo", vps4info)
	// ariang
	r.GET("/tools/ariang", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, pageAriang)
	})
	// 上传
	r.GET("/xyonly/upload", gin.BasicAuth(gin.Accounts{"whoareyou": "minamoto"}), fileUploadWeb)
	r.POST("/xyonly/upload", fileUpload)
	// md5编码
	r.GET("/tools/md5", md5String)
	r.POST("/tools/md5", ginmiddleware.ReadParams(), md5String)
	// 自定义编码
	r.GET("/tools/coder", codeString)
	r.POST("/tools/coder", ginmiddleware.ReadParams(), codeString)
	// 资源下载，youtubedl，aria2
	r.GET("/tools/dl", tdlb)
	r.POST("/tools/dl", ginmiddleware.ReadParams(), tdlb)
	// 查看缓存ip
	r.GET("/tools/cachedip", func(c *gin.Context) { c.String(200, ipCached) })
	// 向cf更新home的最新ip
	r.POST("/tools/updatecf/:who/:proxied", ginmiddleware.ReadParams(), updateCFRecord)

	// 证书管理
	gcert := r.Group("/cert", ginmiddleware.ReadParams())
	gcert.GET("/view/:name", certView)
	gcert.GET("/download/:name", certDownload)
	gcert.GET("/namesilo/:do", certNamesilo)
	gcert.GET("/dnspod/:do", certDNSPod)
	gcert.GET("/cloudflare/:do", certCloudflare)

	// 静态资源路由
	println("----- enable routes: -----")
	for k, v := range dirs {
		if strings.Contains(v, ":") {
			r.StaticFS("/"+strings.Split(v, ":")[0], http.Dir(strings.Split(v, ":")[1]))
			println(fmt.Sprintf("dir %d. /%s/", k+1, strings.Split(v, ":")[0]))
		}
	}

	// webtv 路由
	var gv *gin.RouterGroup
	if *auth {
		gv = r.Group("/v", gin.BasicAuth(gin.Accounts{
			"golang":     "based",
			"thewhyofgo": "simple&fast",
		}), ginmiddleware.ReadParams())
	} else {
		gv = r.Group("/v", ginmiddleware.ReadParams())
	}
	for _, v := range wtv {
		if strings.Contains(v, ":") {
			gv.GET("/"+strings.Split(v, ":")[0], runVideojs(strings.Split(v, ":")[0], strings.Split(v, ":")[1]))
			r.StaticFS("/v-"+strings.Split(v, ":")[0], http.Dir(strings.Split(v, ":")[1]))
		}
	}
	return r
}

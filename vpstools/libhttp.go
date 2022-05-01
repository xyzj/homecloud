package main

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/xyzj/gopsu/gin-middleware"
)

var (
	ipCached       string
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
	// 主页
	r.GET("/", remoteIP)
	// 工具路由
	// vps信息
	r.GET("/tools/vpsinfo", vps4info)

	// md5编码
	r.GET("/tools/md5", md5String)
	r.POST("/tools/md5", ginmiddleware.ReadParams(), md5String)
	// 自定义编码
	r.GET("/tools/coder", codeString)
	r.POST("/tools/coder", ginmiddleware.ReadParams(), codeString)
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

	return r
}

package main

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
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
	return r
}

var listMirror = []string{
	"https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt",
	"https://gitlab.com/gfwlist/gfwlist/raw/master/gfwlist.txt",
	"https://bitbucket.org/gfwlist/gfwlist/raw/HEAD/gfwlist.txt",
	"https://pagure.io/gfwlist/raw/master/f/gfwlist.txt",
	"https://git.tuxfamily.org/gfwlist/gfwlist.git/plain/gfwlist.txt",
	"https://repo.or.cz/gfwlist.git/blob_plain/HEAD:/gfwlist.txt",
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
	r.Use(ginmiddleware.CFConnectingIP())
	// 主页
	r.GET("/", ginmiddleware.PageAbort)
	r.GET("/whoami", remoteIP)
	// 工具路由
	// vps信息
	r.GET("/tools/vpsinfo", vps4info)
	// 查看缓存ip
	r.GET("/tools/cachedip", func(c *gin.Context) { c.String(200, ipCached) })
	// 向cf更新home的最新ip
	// r.POST("/tools/updatecf/:who/:proxied", ginmiddleware.ReadParams(), updateCFRecord)
	r.GET("/7623/list/latest", func(c *gin.Context) {
		b, err := os.ReadFile("7623list.txt")
		if err != nil { // 从github拉取
			for _, line := range listMirror {
				resp, err := http.Get(line)
				if err != nil {
					continue
				}
				b, err = io.ReadAll(resp.Body)
				if err != nil {
					continue
				}
				os.WriteFile("7623list.txt", b, 0o664)
			}
		}
		c.String(200, string(b))
	})

	// 证书管理
	// gcert := r.Group("/cert", ginmiddleware.BasicAuth(), ginmiddleware.ReadParams())
	// gcert.GET("/view/:name", certView)
	// gcert.GET("/download/:name", certDownload)
	// gcert.Static("/cadir", pathtool.JoinPathFromHere("ca"))
	// gcert.GET("/namesilo/:do", certNamesilo)
	// gcert.GET("/cloudflare/:do", certCloudflare)
	// gcert.Static("cfdir", pathtool.JoinPathFromHere("caddy_data", "certificates", "acme-v02.api.letsencrypt.org-directory", "xyzjdays.xyz"))

	return r
}

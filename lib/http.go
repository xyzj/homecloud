package lib

import (
	"fmt"
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

	r.Static("/static", ".")

	r.GET("/", remoteIP)
	r.POST("/", remoteIP)
	r.GET("/test", test)
	r.GET("/givemenewuuid4", newUUID4)
	r.GET("/m/:name", movies)
	r.GET("/share/:name", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:6895%s", ipCached, strings.ReplaceAll(c.Request.URL.String(), "/share", "")))
	})
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
		c.Redirect(http.StatusTemporaryRedirect, "http://180.167.245.233:20080")
	})
	g3.GET("/zd", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "http://180.167.245.233:5990")
	})
	// 证书管理
	g4 := r.Group("/cert", ginmiddleware.ReadParams())
	g4.GET("/sign/:name", certSign)
	g4.GET("/download/:name", certDownload)
	g4.GET("/namesilo/:do", certNamesilo)
	g4.GET("/dnspod/:do", certDNSPod)

	r.HandleMethodNotAllowed = true
	r.NoMethod(ginmiddleware.Page405)
	r.NoRoute(ginmiddleware.Page404)

	// 在微线程中启动服务
	go func() {
		var err error
		println("Starting HTTP(S) server at :" + strconv.Itoa((port)))
		if EnableDebug { // 调试模式下使用http
			err = ginmiddleware.ListenAndServe(port, r)
		} else { // 生产模式下使用https,若设置了clientca，则会验证客户端证书
			go func() {
				rr := gin.New()
				rr.Use(ginmiddleware.Recovery())
				rr.Use(ginmiddleware.TLSRedirect())
				err := rr.Run(fmt.Sprintf(":%d", port-1))
				if err != nil {
					println("http server err:" + err.Error())
				}
			}()
			err = ginmiddleware.ListenAndServeTLS(port, r, filepath.Join(".", "ca", DomainName+".crt"), filepath.Join(".", "ca", DomainName+".key"), "")
		}
		if err != nil {
			println("Failed start HTTP(S) server at :" + strconv.Itoa(port) + "|" + err.Error())
		}
		os.Exit(1)
	}()
}

package lib

import (
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

// multiRender 预置模板
func multiRender() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromString("vpsinfo", TPLVpsinfo)
	r.AddFromString("uuidinfo", TPLuuid)
	return r
}

// NewHTTPService NewHTTPService
func NewHTTPService(port int) {
	if !EnableDebug { // 设置框架使用的模式，默认debug模式，会有控制台输出，release模式仅有文件日志输出
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())

	// 渲染模板
	r.HTMLRender = multiRender()
	
	r.GET("/", remoteIP)
	r.GET("/m/:name", movies)
	g1 := r.Group("/vps")
	g1.GET("v4info", vps4info)
	g2 := r.Group("/wt")
	g2.GET("/", ipCache)
	g2.GET("/:name", wt)
	r.NoMethod(ginmiddleware.Page404)
	r.NoRoute(ginmiddleware.Page404)

	// 在微线程中启动服务
	go func() {
		var err error
		println("Starting HTTP(S) server at :" + strconv.Itoa((port)))
		if EnableDebug { // 调试模式下使用http
			err = ginmiddleware.ListenAndServe(port, r)
		} else { // 生产模式下使用https,若设置了clientca，则会验证客户端证书
			err = ginmiddleware.ListenAndServeTLS(port, r, filepath.Join(gopsu.DefaultConfDir, "ca", "cert.pem"), filepath.Join(gopsu.DefaultConfDir, "ca", "key.pem"), "")
		}
		if err != nil {
			println("Failed start HTTP(S) server at :" + strconv.Itoa(port) + "|" + err.Error())
		}
	}()
}

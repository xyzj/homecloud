package lib

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
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
func NewHTTPService() {
	if !*debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	if *debug {
		r.Use(ginmiddleware.LoggerWithRolling(gopsu.GetExecDir(), "", 3))
	}
	r.Use(ginmiddleware.Recovery())
	// 渲染模板
	r.HTMLRender = multiRender()

	r.Static("/static", gopsu.JoinPathFromHere("static"))

	r.GET("/", remoteIP)
	r.GET("/reloadext", func(c *gin.Context) {
		urlConf.Reload()
		pageWebTV = loadWebTVPage()
		c.String(200, "done")
	})
	// kod共享
	r.GET("/m/:name", func(c *gin.Context) {
		urlConf.Reload()
		n, err := urlConf.GetItem(c.Param("name"))
		if err != nil {
			c.String(200, err.Error())
			return
		}
		s := "https://office.shwlst.com:20081/index.php?share/" + gopsu.DecodeString(n)
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
	g5.GET("/cachedip", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusOK)
		render.WriteString(c.Writer, `<a style="color:white";>https://</a>`+ipCached+`<a style="color:white";>:60019/v/news</a>`, nil)
	})
	g5.POST("/updatecf/:who", ginmiddleware.ReadParams(), updateCFRecord)

	r.HandleMethodNotAllowed = true
	r.NoMethod(ginmiddleware.Page405)
	r.NoRoute(ginmiddleware.Page404)

	// 在微线程中启动服务
	var wl sync.WaitGroup
	var err error
	go func() {
		wl.Add(1)
		defer wl.Done()
		if *web > 1000 {
			err = ginmiddleware.ListenAndServe(*web, r)
			if err != nil {
				println("Failed start HTTP server at :" + strconv.Itoa(*web) + "|" + err.Error())
			}
		}
	}()
	go func() {
		wl.Add(1)
		defer wl.Done()
		if *webs > 1000 && *domain != "" && gopsu.IsExist(crtFile) && gopsu.IsExist(keyFile) {
			err = ginmiddleware.ListenAndServeTLS(*webs, r, crtFile, keyFile, "")
			if err != nil {
				println("Failed start HTTPS server at :" + strconv.Itoa(*webs) + "|" + err.Error())
			}
		}
	}()
	time.Sleep(time.Second)
	wl.Wait()
	os.Exit(2)
}

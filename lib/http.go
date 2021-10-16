package lib

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/xyzj/gopsu"
	ginmiddleware "github.com/xyzj/gopsu/gin-middleware"
)

// multiRender 预置模板
func multiRender() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromString("vpsinfo", tplVpsinfo)
	r.AddFromString("uuidinfo", tpluuid)
	return r
}

func routeEngine() *gin.Engine {
	var logdir, logname string
	var logdays int
	if *filelogon {
		logdir = gopsu.GetExecDir()
		logname = "hcloud"
		logdays = 5
	}
	r := ginmiddleware.LiteEngine(logdir, logname, logdays)
	// 渲染模板
	r.HTMLRender = multiRender()
	// 功能路由
	r.StaticFS("/emb", http.FS(stat))
	r.Static("/static", gopsu.JoinPathFromHere("static"))

	// font css
	r.GET("/css", ginmiddleware.ReadParams(), getCSS)
	r.Static("/fonts", gopsu.JoinPathFromHere("fonts"))

	r.GET("/", remoteIP)
	r.GET("/reloadext", func(c *gin.Context) {
		urlConf.Reload()
		pageWebTV = loadWebTVPage()
		c.String(200, "done")
	})

	// 视频播放
	var vg *gin.RouterGroup
	if *vauth {
		vg = r.Group("/v", ginmiddleware.RateLimit(1, 1), gin.BasicAuth(gin.Accounts{
			"personal": "hwadame",
			"video":    "letmesee",
		}))
	} else {
		vg = r.Group("/v")
	}
	vg.GET("/:dir", ginmiddleware.ReadParams(), runVideojs)
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
	// 证书管理
	g4 := r.Group("/cert", ginmiddleware.ReadParams())
	g4.GET("/view/:name", certView)
	g4.GET("/download/:name", certDownload)
	g4.GET("/namesilo/:do", certNamesilo)
	g4.GET("/dnspod/:do", certDNSPod)
	g4.GET("/cloudflare/:do", certCloudflare)
	// 工具
	g5 := r.Group("/tools")
	g5.GET("/vpsinfo", vps4info)
	// 字符串编码解码
	g5.GET("/codestr", codeString)
	g5.POST("/codestr", ginmiddleware.ReadParams(), codeString)
	// md5编码解码
	g5.GET("/md5", md5String)
	g5.POST("/md5", ginmiddleware.ReadParams(), md5String)
	// youtube 下载
	// g5.GET("/ydl", ginmiddleware.ReadParams(), ydl)
	g5.GET("/ydl", ydlb)
	g5.POST("/ydl", ginmiddleware.ReadParams(), ydlb)
	// 查看缓存的ip
	g5.GET("/cachedip", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, `<a style="color:white";>https://</a>`+ipCached+`<a style="color:white";>:60019/v/news</a>`)
		// c.Status(http.StatusOK)
		// render.WriteString(c.Writer, `<a style="color:white";>https://</a>`+ipCached+`<a style="color:white";>:60019/v/news</a>`, nil)
	})
	// 向cf更新home的最新ip
	g5.POST("/updatecf/:who", ginmiddleware.ReadParams(), updateCFRecord)
	// 迅雷/bt/http下载,支持youtube
	g5.GET("/dl", tdlb)
	g5.POST("/dl", ginmiddleware.ReadParams(), tdlb)
	// aria2 web ui
	g5.Static("/aria2", gopsu.JoinPathFromHere("static", "aria2web"))
	return r
}

/*
// NewHTTPService NewHTTPService
func NewHTTPService() {
	if !*debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	//cors
	r.Use(cors.New(cors.Config{
		MaxAge:           time.Hour * 24,
		AllowAllOrigins:  true,
		AllowCredentials: true,
		AllowWildcard:    true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
	}))

	if *filelogon {
		r.Use(ginmiddleware.LoggerWithRolling(gopsu.GetExecDir(), "hcloud", 3))
	}
	r.Use(ginmiddleware.Recovery())
	// 渲染模板
	r.HTMLRender = multiRender()

	r.StaticFS("/emb", http.FS(stat))
	r.Static("/static", gopsu.JoinPathFromHere("static"))

	// font css
	r.GET("/css", ginmiddleware.ReadParams(), getCSS)
	r.Static("/fonts", gopsu.JoinPathFromHere("fonts"))

	r.GET("/", remoteIP)
	r.GET("/reloadext", func(c *gin.Context) {
		urlConf.Reload()
		pageWebTV = loadWebTVPage()
		c.String(200, "done")
	})

	// 视频播放
	var vg *gin.RouterGroup
	if *vauth {
		vg = r.Group("/v", ginmiddleware.RateLimit(1, 1), gin.BasicAuth(gin.Accounts{
			"personal": "hwadame",
			"video":    "letmesee",
		}))
	} else {
		vg = r.Group("/v")
	}
	vg.GET("/:dir", ginmiddleware.ReadParams(), runVideojs)
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
	// 字符串编码解码
	g5.GET("/codestr", codeString)
	g5.POST("/codestr", ginmiddleware.ReadParams(), codeString)
	// md5编码解码
	g5.GET("/md5", md5String)
	g5.POST("/md5", ginmiddleware.ReadParams(), md5String)
	// youtube 下载
	// g5.GET("/ydl", ginmiddleware.ReadParams(), ydl)
	g5.GET("/ydl", ydlb)
	g5.POST("/ydl", ginmiddleware.ReadParams(), ydlb)
	// 查看缓存的ip
	g5.GET("/cachedip", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, `<a style="color:white";>https://</a>`+ipCached+`<a style="color:white";>:60019/v/news</a>`)
		// c.Status(http.StatusOK)
		// render.WriteString(c.Writer, `<a style="color:white";>https://</a>`+ipCached+`<a style="color:white";>:60019/v/news</a>`, nil)
	})
	// 向cf更新home的最新ip
	g5.POST("/updatecf/:who", ginmiddleware.ReadParams(), updateCFRecord)
	// 迅雷/bt/http下载,支持youtube
	g5.GET("/dl", tdlb)
	g5.POST("/dl", ginmiddleware.ReadParams(), tdlb)
	// aria2 web ui
	g5.Static("/aria2", gopsu.JoinPathFromHere("static", "aria2web"))

	r.HandleMethodNotAllowed = true
	r.NoMethod(ginmiddleware.Page405)
	r.NoRoute(ginmiddleware.Page404)

	// 在微线程中启动服务
	var wl sync.WaitGroup
	var err error
	wl.Add(1)
	go func() {
		defer wl.Done()
		if *web > 1000 {
			err = ginmiddleware.ListenAndServe(*web, r)
			if err != nil {
				println("Failed start HTTP server at :" + strconv.Itoa(*web) + "|" + err.Error())
			}
		}
	}()
	wl.Add(1)
	go func() {
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
*/

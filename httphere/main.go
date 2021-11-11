package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type sliceFlag []string

func (f *sliceFlag) String() string {
	return fmt.Sprintf("%v", []string(*f))
}

func (f *sliceFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

var (
	port  = flag.Int("http", 8019, "set http port")
	ports = flag.Int("https", 0, "set https port")
	cert  = flag.String("cert", "", "cert file path")
	key   = flag.String("key", "", "key file path")
	auth  = flag.Bool("auth", false, "enable basicauth")
	debug = flag.Bool("debug", false, "run in debug mode")
	html  = flag.String("html", "", "something like nginx")
	// 帮助信息
	help = flag.Bool("help", false, "print help message and exit.")
	dirs sliceFlag
)

func main() {
	// http 静态目录
	flag.Var(&dirs, "dir", "example: -dir=name:path -dir name2:path2")
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	if len(dirs) == 0 {
		dirs = []string{"ls:."}
	}

	if !*debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(cors.New(cors.Config{
		MaxAge:           time.Hour * 24,
		AllowAllOrigins:  true,
		AllowCredentials: true,
		AllowWildcard:    true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
	}))
	if *html != "" {
		r.Static("/html", *html)
		r.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusTemporaryRedirect, "/html/index.html")
		})
	}
	if *auth {
		r.Use(gin.BasicAuth(gin.Accounts{
			"golang":     "based",
			"thewhyofgo": "simple&fast",
		}))
	}
	// 静态资源
	for _, v := range dirs {
		if strings.Contains(v, ":") {
			r.StaticFS("/"+strings.Split(v, ":")[0], http.Dir(strings.Split(v, ":")[1]))
		}
	}
	// startup
	if !*debug {
		println(fmt.Sprintf("=== server start up with: %s ===", strings.Join(os.Args[1:], " ")))
	}
	if *cert != "" && *key != "" && *ports > 0 {
		go func() {
			r.RunTLS(fmt.Sprintf(":%d", *ports), *cert, *key)
		}()
	}
	r.Run(fmt.Sprintf(":%d", *port))
}

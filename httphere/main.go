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
	dir   = flag.String("d", ".", "set the dir to serve")
	port  = flag.Int("p", 8060, "set http port")
	cert  = flag.String("cert", "", "cert file path")
	key   = flag.String("key", "", "key file path")
	auth  = flag.Bool("auth", false, "enable basicauth")
	debug = flag.Bool("debug", false, "run in debug mode")
	html  = flag.String("html", "", "设置前端发布的目录")
	// 帮助信息
	help = flag.Bool("help", false, "print help message and exit.")
	dirs sliceFlag
)

func main() {
	flag.Var(&dirs, "moredir", "example: -moredir=name:path -moredir name2:path2")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
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
			"thewhyofgo": "simple",
		}))
	}
	r.StaticFS("/ls", http.Dir(*dir))
	for _, v := range dirs {
		if strings.Contains(v, ":") {
			r.StaticFS("/"+strings.Split(v, ":")[0], http.Dir(strings.Split(v, ":")[1]))
		}
	}
	if !*debug {
		println(fmt.Sprintf("=== server start on :%d ===", *port))
	}
	if *cert != "" && *key != "" {
		r.RunTLS(fmt.Sprintf(":%d", *port), *cert, *key)
	}
	r.Run(fmt.Sprintf(":%d", *port))
}

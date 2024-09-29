package main

import (
	"flag"
	"hash/crc32"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xyzj/gopsu"
	ginmiddleware "github.com/xyzj/gopsu/gin-middleware"
	"github.com/xyzj/gopsu/gocmd"
	"github.com/xyzj/gopsu/pathtool"
)

var (
	version     = "0.0.0"
	goVersion   = ""
	buildDate   = ""
	platform    = ""
	author      = "Xu Yuan"
	programName = "Url short alias"
)

var port = flag.Int("http", 6820, "http port")

type namelist struct {
	Short string `json:"name"`
	Uri   string `json:"uri"`
}

func main() {
	gocmd.DefaultProgram(&gocmd.Info{
		Title: "Url short alias",
		Ver:   gopsu.VersionInfo(programName, version, goVersion, buildDate, platform, author),
	}).Execute()
	p := pathtool.JoinPathFromHere("short.d")
	cc := ginmiddleware.NewHTTPClient()
	os.Mkdir(p, 0o775)
	r := ginmiddleware.LiteEngine("", 0)
	r.Use(ginmiddleware.ReadParams())
	r.GET("/r/:alias", func(c *gin.Context) {
		a := c.Param("alias")
		b, err := os.ReadFile(filepath.Join(p, a+".uri"))
		if err != nil {
			c.String(200, err.Error())
			return
		}
		c.Redirect(307, strings.TrimSpace(string(b)))
	})
	r.GET("/s/:alias", func(c *gin.Context) {
		a := c.Param("alias")
		b, err := os.ReadFile(filepath.Join(p, a+".uri"))
		if err != nil {
			c.String(200, err.Error())
			return
		}
		req, err := http.NewRequest("GET", strings.TrimSpace(string(b)), nil)
		if err != nil {
			c.String(200, err.Error())
			return
		}

		sc, body, header, err := cc.DoRequest(req, time.Second*10)
		if err != nil {
			c.String(200, err.Error())
			return
		}
		for k, v := range header {
			c.Writer.Header().Set(k, v)
		}
		c.Writer.WriteHeader(sc)
		c.Writer.Write(body)
		c.Writer.Flush()
	})
	r.GET("/uri/add", ginmiddleware.CheckRequired("uri"), func(c *gin.Context) {
		uri := c.Param("uri")
		if uri == "" || !strings.HasPrefix(uri, "http") {
			c.String(200, "uri illegal")
			return
		}
		alias := c.Param("alias")
		if alias == "" {
			alias = strconv.FormatUint(uint64(crc32.ChecksumIEEE([]byte(uri))), 16)
		}
		a := filepath.Join(p, alias)
		if pathtool.IsExist(a) {
			c.String(200, "alias already exist")
			return
		}
		os.WriteFile(a+".uri", []byte(uri), 0o664)
		c.String(200, uri+" add as /s/"+alias)
	})
	r.GET("/uri/del", ginmiddleware.CheckRequired("alias"), func(c *gin.Context) {
		err := os.Remove(filepath.Join(p, c.Param("alias")+".uri"))
		if err != nil {
			c.String(200, err.Error())
			return
		}
		c.String(200, c.Param("alias")+" removed")
	})
	r.GET("/uri/list", func(c *gin.Context) {
		dirs, err := os.ReadDir(p)
		if err != nil {
			c.String(200, err.Error())
			return
		}
		namelists := make([]*namelist, 0)
		for _, dir := range dirs {
			if dir.IsDir() {
				continue
			}
			if filepath.Ext(dir.Name()) != ".uri" {
				continue
			}
			b, err := os.ReadFile(filepath.Join(p, dir.Name()))
			if err != nil {
				continue
			}
			namelists = append(namelists, &namelist{
				Short: strings.TrimSuffix(dir.Name(), ".uri"),
				Uri:   string(b),
			})
		}
		sort.Slice(namelists, func(i, j int) bool {
			return namelists[i].Short < namelists[j].Short
		})
		c.Writer.Header().Set("Content-Type", "text/html")
		c.Writer.WriteHeader(200)
		for _, v := range namelists {
			c.Writer.WriteString(v.Short + "<br>" + v.Uri + "<br><br>")
		}
		c.Writer.Flush()
	})
	err := ginmiddleware.ListenAndServe(*port, r)
	if err != nil {
		println(err.Error())
	}
}

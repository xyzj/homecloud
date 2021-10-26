package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

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
	dir  = flag.String("d", ".", "set the dir to serve")
	port = flag.Int("p", 8060, "set http port")
	cert = flag.String("cert", "", "cert file path")
	key  = flag.String("key", "", "key file path")
	dirs sliceFlag
)

func main() {
	flag.Var(&dirs, "moredir", "example: -moredir=name:path -moredir name2:path2")
	flag.Parse()
	r := gin.New()
	r.StaticFS("/ls", http.Dir(*dir))
	for _, v := range dirs {
		if strings.Contains(v, ":") {
			r.StaticFS("/"+strings.Split(v, ":")[0], http.Dir(strings.Split(v, ":")[1]))
		}
	}
	if *cert != "" && *key != "" {
		r.RunTLS(fmt.Sprintf(":%d", *port), *cert, *key)
	}
	r.Run(fmt.Sprintf(":%d", *port))
}

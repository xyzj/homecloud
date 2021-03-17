package main

import (
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
)

var (
	dir  = flag.String("dir", ".", "set the dir to serve")
	http = flag.Int("http", 8060, "set http port")
)

func main() {
	flag.Parse()
	r := gin.New()
	r.StaticFile("/", *dir)
	r.Run(fmt.Sprintf(":%d", *http))
}

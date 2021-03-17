package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	dir  = flag.String("d", ".", "set the dir to serve")
	port = flag.Int("p", 8060, "set http port")
)

func main() {
	flag.Parse()
	r := gin.New()
	r.StaticFS("/", http.Dir(*dir))
	r.Run(fmt.Sprintf(":%d", *port))
}

package main

import (
	"net/url"
	"os"
	"strings"
)

func main() {
	args := os.Args
	if len(args) < 3 {
		println("Useage: \n\turlparse encode \"url\"\n\turlparse decode \"url\"")
		return
	}
	switch args[1] {
	case "encode":
		u, _ := url.Parse(strings.Join(args[2:], " "))
		os.Stdout.WriteString(u.Redacted() + "\n")
		// println(u.Redacted())
	case "decode":
		u, _ := url.QueryUnescape(strings.Join(args[2:], " "))
		os.Stdout.WriteString(u + "\n")
		// println(u)
	default:
		println("unknow command")
	}
}

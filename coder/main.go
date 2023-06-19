package main

import (
	"flag"

	"github.com/xyzj/gopsu"
)

func main() {
	encode := flag.String("e", "", "input the string you want to encode")
	decode := flag.String("d", "", "input the string you want to decode")
	md := flag.String("md5", "", "count md5")
	flag.Parse()
	if *encode != "" {
		println(gopsu.CodeString(*encode))
	}
	if *decode != "" {
		println(gopsu.DecodeString(*decode))
	}
	if *md != "" {
		println(gopsu.GetMD5(*md))
	}
}

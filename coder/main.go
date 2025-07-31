package main

import (
	"flag"

	"github.com/xyzj/toolbox"
	"github.com/xyzj/toolbox/crypto"
)

func main() {
	old := flag.Bool("o", false, "use old version")
	encode := flag.Bool("e", true, "Encode the input")
	decode := flag.Bool("d", false, "Decode the input")
	md := flag.String("md5", "", "count md5")
	flag.Parse()
	if *md != "" {
		println(crypto.GetMD5(*md))
		return
	}
	if *old {
		if *decode || !*encode {
			for _, v := range flag.Args() {
				println(toolbox.DecodeString(v))
			}
			return
		} else {
			for _, v := range flag.Args() {
				println(toolbox.CodeString(v))
			}
			return
		}
	}
	if *decode {
		for _, v := range flag.Args() {
			println(crypto.DeobfuscationString(v))
		}
		return
	} else {
		for _, v := range flag.Args() {
			println(crypto.ObfuscationString(v))
		}
		return
	}
}

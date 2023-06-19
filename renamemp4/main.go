package main

import (
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		println("useage: renamemp4 oldkeywords newkeywords")
	}
	fs, err := ioutil.ReadDir(".")
	if err != nil {
		println(err.Error())
		return
	}
	oldkeywords := os.Args[1]
	newkeywords := os.Args[2]
	for _, v := range fs {
		if v.IsDir() || !strings.HasSuffix(v.Name(), ".mp4") {
			continue
		}
		if strings.Contains(v.Name(), oldkeywords) {
			os.Rename(v.Name(), strings.ReplaceAll(v.Name(), oldkeywords, newkeywords))
			println("rename", v.Name(), "to", strings.ReplaceAll(v.Name(), oldkeywords, newkeywords))
		}
	}
}

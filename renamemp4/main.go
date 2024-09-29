package main

import (
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		println("useage: renamemp4 dir")
	}
	fs, err := os.ReadDir(os.Args[1])
	if err != nil {
		println(err.Error())
		return
	}
	namerep := strings.NewReplacer("。", ",", "，", ",", " ", "_", "…", "-", "！", ",", "●", "-", "～", "~", "」", "", "「", "")
	for _, v := range fs {
		if v.IsDir() {
			continue
		}
		name := strings.ToLower(v.Name())
		name = strings.ReplaceAll(name, "] ", " ")
		name = strings.ReplaceAll(name, "]", " ")
		name = strings.ReplaceAll(name, "[", "")
		name = strings.ReplaceAll(name, " - ", "-")
		name = strings.TrimPrefix(name, "_")
		name = namerep.Replace(name)
		switch filepath.Ext(name) {
		case ".mp4", ".mkv", ".avi":
			err := os.Rename(filepath.Join(os.Args[1], v.Name()), filepath.Join(os.Args[1], name))
			if err != nil {
				println(err.Error())
				continue
			}
			println("rename " + v.Name() + " to " + name)
			// go to next
		default:
			continue
		}
	}
}

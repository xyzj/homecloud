package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

func main() {
	mergesub, err := os.ReadFile("mergesub.txt")
	if err != nil {
		println(err.Error())
		return
	}
	var args bytes.Buffer
	args.WriteString("#!/bin/bash\n\n")
	for idx, f := range strings.Split(string(mergesub), "\n") {
		file := strings.TrimSpace(f)
		if file == "" {
			continue
		}
		// 提取字幕
		args.WriteString("ffmpeg ")
		args.WriteString("-i ")
		args.WriteString("'" + file + "' ")
		args.WriteString("-map ")
		args.WriteString("0:2 ")
		args.WriteString("-y ")
		args.WriteString(fmt.Sprintf("%04d.srt ", idx))
		// 合成字幕
		args.WriteString(" && ffmpeg ")
		args.WriteString("-i ")
		args.WriteString("'" + file + "' ")
		args.WriteString("-vf ")
		args.WriteString("subtitles=" + fmt.Sprintf("%04d.srt ", idx))
		args.WriteString("-vcodec h264 ")
		args.WriteString("-acodec copy ")
		args.WriteString("-y ")
		args.WriteString("'" + "merge_" + file + "'\n\n")
	}
	os.WriteFile("mergesub.sh", args.Bytes(), 0o664)
}

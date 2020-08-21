package lib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/xyzj/gopsu"
)

// Smi2Vtt smi转vtt
func Smi2Vtt(in, out string) error {
	bIn, err := ioutil.ReadFile(in)
	if err != nil {
		return err
	}
	ss := strings.Split(string(bIn), "\n")
	// 分析语言
	var language string
	for _, v := range ss {
		if strings.HasPrefix(v, "-->") {
			break
		}
		if strings.HasPrefix(v, ".zh") {
			language = v[1:strings.Index(v, " {")]
			if strings.Contains(v, "自动翻译") || language != "zh-Hant" {
				continue
			}
			break
		}
	}
	// 分析主体
	var bOut bytes.Buffer
	// bOut.WriteString("WEBVTT\r\n\r\n")
	var text string
	var tStart, tEnd int
	for _, v := range ss {
		v = gopsu.TrimString(v)
		idxTime1 := strings.Index(v, "Start=")
		idxClass := strings.Index(v, "class=")
		if idxTime1 == -1 || idxClass == -1 || !strings.Contains(v, "class='"+language+"'") {
			continue
		}
		idxTime2 := strings.Index(v, ">")
		idxText := strings.LastIndex(v, ">")
		if v[idxText+1:] == "" {
			continue
		}
		if v[idxText+1:] == "&nbsp;" {
			tEnd = gopsu.String2Int(v[idxTime1+6:idxTime2], 10)
		} else {
			tStart = gopsu.String2Int(v[idxTime1+6:idxTime2], 10)
			text = v[idxText+1:]
		}
		if tStart >= 0 && tEnd > 0 && len(text) > 0 {
			bOut.WriteString(fmt.Sprintf("%02d:%02d:%02d.%03d --> %02d:%02d:%02d.%03d\r\n",
				tStart/1000/60/60,
				tStart/1000/60%60,
				tStart/1000%60,
				tStart%1000,
				tEnd/1000/60/60,
				tEnd/1000/60%60,
				tEnd/1000%60,
				tEnd%1000))
			bOut.WriteString(text + "\r\n")
			tStart = -1
			tEnd = 0
			text = ""
		}
	}
	if bOut.Len() == 0 {
		return nil
	}
	return ioutil.WriteFile(out, []byte("WEBVTT\r\n\r\n"+bOut.String()), 0644)
}

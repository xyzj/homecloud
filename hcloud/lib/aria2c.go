package lib

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xyzj/toolbox/httpclient"
	"github.com/xyzj/toolbox/pathtool"
)

var (
	aria2curl = "http://127.0.0.1:2085"
	ytdir     = "/mm/tv1/news/"
)

func SetAria2cUrl(curl string) {
	aria2curl = pathtool.TrimSlash(curl)
}
func SetYtdir(dir string) {
	ytdir = pathtool.AppendSlash(dir)
}
func CheckAria2cActive() {
	t := time.NewTicker(time.Minute * 30)
	for range t.C {
		s, _ := sjson.SetBytes([]byte{}, "jsonrpc", "2.0")
		s, _ = sjson.SetBytes(s, "id", fmt.Sprintf("%d", time.Now().UnixNano()))
		s, _ = sjson.SetBytes(s, "method", "aria2.tellActive")
		s, _ = sjson.SetBytes(s, "params", []string{})
		ss := strings.ReplaceAll(base64.URLEncoding.EncodeToString(s), "=", "%3D")
		req, _ := http.NewRequest("GET", aria2curl+"/jsonrpc?params="+ss, strings.NewReader(""))
		_, body, _, err := httpclient.DoRequestWithTimeout(req, time.Second*2)
		if err != nil {
			continue
		}
		if len(gjson.GetBytes(body, "result").Array()) > 0 {
			continue
		}
		// shutdown
		println("no active downloads, shutdown...")
		s, _ = sjson.SetBytes([]byte{}, "jsonrpc", "2.0")
		s, _ = sjson.SetBytes(s, "id", fmt.Sprintf("%d", time.Now().UnixNano()))
		s, _ = sjson.SetBytes(s, "method", "aria2.shutdown")
		s, _ = sjson.SetBytes(s, "params", []string{})
		ss = strings.ReplaceAll(base64.URLEncoding.EncodeToString(s), "=", "%3D")
		req, _ = http.NewRequest("GET", aria2curl+"/jsonrpc?params="+ss, strings.NewReader(""))
		httpclient.DoRequestWithTimeout(req, time.Second*2)
	}
}

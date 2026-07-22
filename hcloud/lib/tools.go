package lib

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xyzj/toolbox/crypto"
	"github.com/xyzj/toolbox/httpclient"
)

const (
	bwhStatusURL = "https://api.64clouds.com/v1/getServiceInfo?veid=%s&api_key=%s"
	bwhAPIKey    = "private_3togcVA1FRA9xdwhupifGEzo"
	bwhVeid      = "979913"
)

var (
	ipCached  string
	ipCached6 string
)

// multiRender 预置模板
func multiRender() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromString("vpsinfo", tplVpsinfo)
	return r
}

func vps4info(c *gin.Context) {
	req, _ := http.NewRequest("GET", fmt.Sprintf(bwhStatusURL, bwhVeid, bwhAPIKey), strings.NewReader(""))
	_, d, _, ex := httpclient.DoRequestWithTimeout(req, time.Second*5)
	if ex == nil {
		a := gjson.ParseBytes(d)
		c.Set("plan", a.Get("plan").String())
		c.Set("vmtype", a.Get("vm_type").String())
		c.Set("os", a.Get("os").String())
		c.Set("hostname", a.Get("hostname").String())
		c.Set("location", a.Get("node_location").String())
		c.Set("datacenter", a.Get("node_datacenter").String())
		c.Set("plan_monthly_data", a.Get("plan_monthly_data").Float()/1024.0/1024.0/1024.0)
		c.Set("data_counter", fmt.Sprintf("%.03f", a.Get("data_counter").Float()/1024.0/1024.0/1024.0))
		c.Set("ivp6", a.Get("location_ipv6_ready").String())
		c.Set("error", a.Get("error").String())
		c.Set("ipv4", a.Get("ip_addresses").Array()[0].String()+":26937")
	} else {
		c.Set("err", ex.Error())
	}
	// c.Header("Content-Type", "text/html")
	// c.String(200, tplVpsinfo, c.Keys)
	c.HTML(200, "vpsinfo", c.Keys)
}

func codeString(c *gin.Context) {
	if c.Request.Method == "POST" {
		c.String(200, crypto.ObfuscationString(c.Param("rawstr")))
		return
	}
	// web页面
	c.Header("Content-Type", "text/html")
	c.String(200, tplCodeStr)
}

func md5String(c *gin.Context) {
	if c.Request.Method == "POST" {
		c.String(200, crypto.GetMD5(c.Param("rawstr")))
		return
	}
	// web页面
	c.Header("Content-Type", "text/html")
	c.String(200, tplMD5Str)
}

// zone: fb8a871c3737648dfd964bd625f9f685
// da.xyzjdays.xyz: A 712df327b64333800c02511f404b3157
// 6.xyzjdays.xyz: AAAA e9bf2e603c7c1ec17c3d0dc7dd18d391
// curl 4: 4.ipw.cn; checkip.amazonaws.com; whatismyip.akamai.com
// curl 6: 6.ipw.cn; curlmyip.net; wgetip.com
// https://www.cnblogs.com/mainos/p/15863048.html
func updateCFRecord(c *gin.Context) {
	if c.Param("who") != "ohana" {
		c.String(403, " I don't know you")
		return
	}
	out := &bytes.Buffer{}
	// 处理ip6
	ip6 := c.Param("ip6")
	proxied6, _ := strconv.ParseBool(c.Param("proxied6"))
	if len(strings.Split(ip6, ":")) == 8 { // 合法ip6
		if ip6 != ipCached6 {
			url := "https://api.cloudflare.com/client/v4/zones/fb8a871c3737648dfd964bd625f9f685/dns_records/254bb30ffaa567af7393b0d159418956"
			var js string
			js, _ = sjson.Set(js, "type", "AAAA")
			js, _ = sjson.Set(js, "name", "d6")
			js, _ = sjson.Set(js, "content", ip6)
			js, _ = sjson.Set(js, "ttl", 1)
			js, _ = sjson.Set(js, "proxied", proxied6)
			req, _ := http.NewRequest("PUT", url, strings.NewReader(js))
			req.Header.Add("X-Auth-Email", "minamoto.xu@outlook.com")
			req.Header.Add("X-Auth-Key", "b6c9de4a9814d534ab16c12d99718f118fde2")
			req.Header.Add("Content-Type", "application/json")
			sc, b, _, err := httpclient.DoRequestWithTimeout(req, time.Second*5)
			if err != nil {
				c.String(sc, err.Error())
				return
			}
			out.Write(b)
			// c.String(200, string(b))
			ipCached6 = ip6
		} else {
			out.WriteString(fmt.Sprintf("ip6 %s not changed, nothing to do", ip6))
		}
		out.WriteString("\n\n")
	}
	// 处理ip4
	ip4 := c.Param("ip4")
	proxied, _ := strconv.ParseBool(c.Param("proxied"))
	// if ip4 == "" {
	// 	if ip := c.Request.Header.Get("CF-Connecting-IP"); ip == "" {
	// 		ip4 = c.ClientIP()
	// 	} else {
	// 		ip4 = ip
	// 	}
	// }
	if len(strings.Split(ip4, ".")) == 4 {
		if ip4 != ipCached {
			// 改da
			url := "https://api.cloudflare.com/client/v4/zones/fb8a871c3737648dfd964bd625f9f685/dns_records/712df327b64333800c02511f404b3157"
			var js string
			js, _ = sjson.Set(js, "type", "A")
			js, _ = sjson.Set(js, "name", "da")
			js, _ = sjson.Set(js, "content", ip4)
			js, _ = sjson.Set(js, "ttl", 1)
			js, _ = sjson.Set(js, "proxied", proxied)
			req, _ := http.NewRequest("PUT", url, strings.NewReader(js))
			req.Header.Add("X-Auth-Email", "minamoto.xu@outlook.com")
			req.Header.Add("X-Auth-Key", "b6c9de4a9814d534ab16c12d99718f118fde2")
			req.Header.Add("Content-Type", "application/json")
			sc, b, _, err := httpclient.DoRequestWithTimeout(req, time.Second*5)
			if err != nil {
				c.String(sc, err.Error())
				return
			}
			out.Write(b)
			// 更新cache
			ipCached = ip4
		} else {
			out.WriteString(fmt.Sprintf("ip %s not changed, nothing to do", ip4))
		}
		out.WriteString("\n\n")
	}
	if out.Len() == 0 {
		c.String(200, "ip not changed, nothing to do\n")
	} else {
		c.String(200, out.String())
	}
}

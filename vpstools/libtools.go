package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/xyzj/gopsu"
)

const (
	bwhStatusURL = "https://api.64clouds.com/v1/getServiceInfo?veid=%s&api_key=%s"
	bwhAPIKey    = "yfCUSxAg5fs9DMzQntChzNkPneEsvMm5bMo+iuDt9Zr0itwcP3vSrMDOfeCovNA0igyKy2z1bKy8CxsQTYCNexa"
	bwhVeid      = "979913"
)

var (
	shortconf *gopsu.ConfData
	skey      = "0987654321qwertyuiopQWERTYUIOPasdfghjklASDFGHJKLzxcvbnmZXCVBNM"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func codeString(c *gin.Context) {
	if c.Request.Method == "POST" {
		c.String(200, gopsu.CodeString(c.Param("rawstr")))
		return
	}
	// web页面
	c.Header("Content-Type", "text/html")
	c.String(200, tplCodeStr)
}

func md5String(c *gin.Context) {
	if c.Request.Method == "POST" {
		c.String(200, gopsu.GetMD5(c.Param("rawstr")))
		return
	}
	// web页面
	c.Header("Content-Type", "text/html")
	c.String(200, tplMD5Str)
}

func vps4info(c *gin.Context) {
	req, _ := http.NewRequest("GET", fmt.Sprintf(bwhStatusURL, bwhVeid, gopsu.DecodeString(bwhAPIKey)), strings.NewReader(""))
	resp, ex := httpClientPool.Do(req)
	if ex == nil {
		if d, ex := ioutil.ReadAll(resp.Body); ex == nil {
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
			println(ex.Error())
		}
	}
	// c.Header("Content-Type", "text/html")
	// c.String(200, tplVpsinfo, c.Keys)
	c.HTML(200, "vpsinfo", c.Keys)
}

func remoteIP(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		c.String(200, c.ClientIP())
	case "POST":
		ioutil.WriteFile(".ipcache", []byte(c.ClientIP()), 0644)
		c.String(200, "success")
	}
}

func shortURL(c *gin.Context) {
	do := c.Param("do")
	surl := c.Param("v")
	switch do {
	case "add":
		_, err := http.Get(surl)
		if err != nil {
			c.String(400, "this url can not be get "+err.Error())
			return
		}
		b := []byte(surl)
		v := fmt.Sprintf("%04x", gopsu.CountCrc16VB(&b))
		if _, err := shortconf.GetItem(v); err == nil {
			v += string(skey[int(rand.Int31n(int32(len(skey))))])
		}
		shortconf.SetItem(v, surl, "short url")
		shortconf.Save()
		c.String(200, "https://xyzjdays.xyz/%s", v)
	case "del":
		shortconf.DelItem(surl)
		shortconf.Save()
	case "show":
		c.PureJSON(200, gjson.Parse(shortconf.GetAll()).Value())
	default: // redir
		url, err := shortconf.GetItem(do)
		if err != nil {
			c.String(400, err.Error())
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

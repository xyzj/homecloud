package lib

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/xyzj/gopsu"
)

// 这个是服务接口的业务逻辑处理文件，所有接口方法可以都写在这里，或按类别分文件写
// 此处是用户管理服务的代码，供参考，所有方法都是和http.go里面的路由对应的
// Enjoy your coding

// login 登录
func movies(c *gin.Context) {
	r := c.Param("name")
	urlConf.Reload()
	switch r {
	case "all":
	default:
		if ipCached == "" {
			b, err := ioutil.ReadFile(".ipcache")
			if err != nil {
				c.String(200, err.Error())
				return
			}
			ipCached = strings.TrimSpace((string(b))
		}
		n, err := urlConf.GetItem(r)
		if err != nil {
			c.String(200, err.Error())
			return
		}
		s := fmt.Sprintf("http://%s:6895/index.php?share/"+n, ipCached)
		c.Redirect(http.StatusTemporaryRedirect, s)
	}
}

func vps4info(c *gin.Context) {

	// use golang net
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	resp, ex := client.Get(fmt.Sprintf(bwhStatusURL, bwhVeid, gopsu.DecodeString(bwhAPIKey)))
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
	c.HTML(200, "vpsinfo", c.Keys)
}

func wt(c *gin.Context) {
	if ipCached == "" {
		b, err := ioutil.ReadFile(".ipcache")
		if err != nil {
			c.String(200, err.Error())
			return
		}
		ipCached = string(b)
	}
	r := c.Param("name")
	switch r {
	case "mldonkey":
		s := fmt.Sprintf("http://%s:6893/", ipCached)
		c.Redirect(http.StatusTemporaryRedirect, s)
	case "kod":
		s := fmt.Sprintf("http://%s:6895/", ipCached)
		c.Redirect(http.StatusTemporaryRedirect, s)
	case "deluge":
		s := fmt.Sprintf("http://%s:6892/", ipCached)
		c.Redirect(http.StatusTemporaryRedirect, s)
	case "ssh":
		s := fmt.Sprintf("http://%s:6896/ssh/host/127.0.0.1", ipCached)
		c.Redirect(http.StatusTemporaryRedirect, s)
	}
}

func remoteIP(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		c.String(200, strings.Split(c.Request.RemoteAddr, ":")[0])
	case "POST":
		ioutil.WriteFile(".ipcache", []byte(strings.Split(c.Request.RemoteAddr, ":")[0]), 0644)
		// f, ex := os.OpenFile(".ipcache", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		// defer f.Close()
		// if ex == nil {
		// 	f.WriteString(strings.Split(c.Request.RemoteAddr, ":")[0])
		// }
		c.String(200, "success")
	}
}

func ipCache(c *gin.Context) {
	b, _ := ioutil.ReadFile(".ipcache")
	c.String(200, string(b))
}

package lib

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/xyzj/gopsu"
)

// 这个是服务接口的业务逻辑处理文件，所有接口方法可以都写在这里，或按类别分文件写
// 此处是用户管理服务的代码，供参考，所有方法都是和http.go里面的路由对应的
// Enjoy your coding

// login 登录
func movies(c *gin.Context) {
	urlConf.Reload()
	if ipCached == "" {
		b, err := ioutil.ReadFile(".ipcache")
		if err != nil {
			c.String(200, err.Error())
			return
		}
		ipCached = strings.TrimSpace((string(b)))
	}
	n, err := urlConf.GetItem(c.Param("name"))
	if err != nil {
		c.String(200, err.Error())
		return
	}
	s := fmt.Sprintf("http://%s:6895/index.php?share/"+n, ipCached)
	c.Redirect(http.StatusTemporaryRedirect, s)
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
		c.String(200, c.ClientIP())
	case "POST":
		ipCached = c.ClientIP()
		ioutil.WriteFile(".ipcache", []byte(c.ClientIP()), 0644)
		c.String(200, "success")
	}
}

func ipCache(c *gin.Context) {
	b, _ := ioutil.ReadFile(".ipcache")
	ipCached = string(b)
	c.String(200, string(b))
}

func newUUID4(c *gin.Context) {
	var s string
	d, _ := uuid.NewRandom()
	c.Set("newuuid4", d.String())
	var i = 1
	a, err := ioutil.ReadFile("/root/conf/v2rays.json")
	if err == nil {
		b := gjson.ParseBytes(a)
		e := b.Get("inbounds.#.settings.clients.0.id")
		e.ForEach(func(key, value gjson.Result) bool {
			c.Set("uuid"+strconv.Itoa(i), value.String())
			s += "\n" + value.String() + "\n"
			i++
			return true
		})
	}
	a, err = ioutil.ReadFile("/root/conf/v2rayfwd.json")
	if err == nil {
		b := gjson.ParseBytes(a)
		e := b.Get("inbounds.#.settings.clients.0.id")
		e.ForEach(func(key, value gjson.Result) bool {
			c.Set("uuid"+strconv.Itoa(i), value.String())
			s += "\n" + value.String() + "\n"
			i++
			return true
		})
		e = b.Get("outbounds.#.settings.vnext.0.users.0.id")
		e.ForEach(func(key, value gjson.Result) bool {
			c.Set("uuid"+strconv.Itoa(i), value.String())
			s += "\n" + value.String() + "\n"
			i++
			return true
		})
	}
	switch c.Request.Method {
	case "GET":
		c.HTML(200, "uuidinfo", c.Keys)
	case "POST":
		c.String(200, s)
	}
}

func certDownload(c *gin.Context) {
	name := c.Param("name")
	crtsrc := filepath.Join(".lego", "certificates", "_."+name+".crt")
	keysrc := filepath.Join(".lego", "certificates", "_."+name+".key")
	crtdst := filepath.Join("ca", name+".crt")
	keydst := filepath.Join("ca", name+".key")
	if gopsu.IsExist(crtsrc) && gopsu.IsExist(keysrc) {
		os.Mkdir("ca", 0775)
		_, err := gopsu.CopyFile(crtsrc, crtdst)
		if err != nil {
			c.String(200, err.Error())
			return
		}
		_, err = gopsu.CopyFile(keysrc, keydst)
		if err != nil {
			c.String(200, err.Error())
			return
		}
		err = gopsu.ZIPFiles(name+".zip", []string{crtdst, keydst}, "ca")
		if err != nil {
			c.String(200, err.Error())
			return
		}
		c.FileAttachment(name+".zip", name+".zip")
	} else {
		c.String(200, "no certificate files found")
	}
}

func certNamesilo(c *gin.Context) {
	// cmd := exec.Command(filepath.Join(".", "lego"), strings.Split("--dns namesilo --domains *.xyzjdays.xyz --email minamoto.xu@outlook.com -a run", " ")...)
	cmd := exec.Command(filepath.Join(".", "lego"))
	cmd.Env = append(cmd.Env, "NAMESILO_API_KEY=f59e74d5e3f373b9e332e9b")
	cmd.Env = append(cmd.Env, "NAMESILO_PROPAGATION_TIMEOUT=1500")
	cmd.Env = append(cmd.Env, "NAMESILO_TTL=3600")
	cmd.Env = append(cmd.Env, "NAMESILO_POLLING_INTERVAL=10")

	switch c.Param("do") {
	case "run": // 创建新证书
		cmd.Args = strings.Split("--dns namesilo --domains *.xyzjdays.xyz --email minamoto.xu@outlook.com -a run", " ")
		go func() {
			out, err := cmd.CombinedOutput()
			if err != nil {
				ioutil.WriteFile("legoerr.log", []byte(err.Error()), 0664)
				return
			}
			ioutil.WriteFile("namesilo_renew.log", out, 0664)
			if runtime.GOOS == "linux" {
				if gopsu.IsExist(linuxSSLCopy) {
					cmd = exec.Command(linuxSSLCopy)
					cmd.Run()
				}
			}
			if runtime.GOOS == "windows" {
				if gopsu.IsExist(windowsSSLCopy) {
					cmd = exec.Command(windowsSSLCopy)
					cmd.Run()
				}
			}
		}()
		c.String(200, "Done, you can try to download cert and key file 20 minutes later")
	case "renew": // 更新证书
		cmd.Args = strings.Split("--dns namesilo --domains *.xyzjdays.xyz --email minamoto.xu@outlook.com -a renew --reuse-key", " ")
		go func() {
			out, err := cmd.CombinedOutput()
			if err != nil {
				ioutil.WriteFile("legoerr.log", []byte(err.Error()), 0664)
				return
			}
			ioutil.WriteFile("namesilo_renew.log", out, 0664)
			if strings.Contains(string(out), "no renew") {
				return
			}
			if runtime.GOOS == "linux" {
				if gopsu.IsExist(linuxSSLCopy) {
					cmd = exec.Command(linuxSSLCopy)
					cmd.Run()
				}
			}
			if runtime.GOOS == "windows" {
				if gopsu.IsExist(windowsSSLCopy) {
					cmd = exec.Command(windowsSSLCopy)
					cmd.Run()
				}
			}

		}()
		c.String(200, "Done, you can try to download cert and key file 20 minutes later")
	default:
		c.String(200, "Don't understand")
	}
}

func certDNSPod(c *gin.Context) {
	switch c.Param("do") {
	case "run": // 创建新证书
	case "renew": // 更新证书
	default:
		c.String(200, "Don't understand")
	}
}

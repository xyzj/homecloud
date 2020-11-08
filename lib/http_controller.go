package lib

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/tidwall/gjson"
	"github.com/xyzj/gopsu"
)

func codeString(c *gin.Context) {
	if c.Request.Method == "POST" {
		c.String(200, gopsu.CodeString(c.Param("rawstr")))
		return
	}
	// web页面
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
	render.WriteString(c.Writer, tplCodeStr, nil)
}

// login 登录
func movies(c *gin.Context) {
	urlConf.Reload()
	// if ipCached == "" {
	// 	b, err := ioutil.ReadFile(".ipcache")
	// 	if err != nil {
	// 		c.String(200, err.Error())
	// 		return
	// 	}
	// 	ipCached = strings.TrimSpace((string(b)))
	// }
	n, err := urlConf.GetItem(c.Param("name"))
	if err != nil {
		c.String(200, err.Error())
		return
	}
	s := "https://kod.xyzjdays.xyz:10043/index.php?share/" + gopsu.DecodeString(n)
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

func remoteIP(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		c.String(200, c.ClientIP())
	case "POST":
		ioutil.WriteFile(".ipcache", []byte(c.ClientIP()), 0644)
		c.String(200, "success")
	}
}

func certSign(c *gin.Context) {
	name := c.Param("name")
	crtdst := filepath.Join("ca", name+".crt")
	keydst := filepath.Join("ca", name+".key")
	if gopsu.IsExist(crtdst) && gopsu.IsExist(keydst) {
		var b bytes.Buffer
		a, err := ioutil.ReadFile(crtdst)
		if err != nil {
			c.String(400, "load cert file sign error:"+err.Error())
		}
		b.Write(a)
		a, err = ioutil.ReadFile(keydst)
		if err != nil {
			c.String(200, "load key file sign error:"+err.Error())
		}
		b.Write(a)
		c.String(200, gopsu.GetMD5(b.String()))
	} else {
		c.String(400, "no certificate files found")
	}
}

func certDownload(c *gin.Context) {
	name := c.Param("name")
	crtdst := filepath.Join("ca", name+".crt")
	keydst := filepath.Join("ca", name+".key")

	if gopsu.IsExist(crtdst) && gopsu.IsExist(keydst) {
		os.Mkdir("ca", 0775)
		err := gopsu.ZIPFiles(name+".zip", []string{crtdst, keydst}, "")
		if err != nil {
			c.String(400, err.Error())
			return
		}
		c.FileAttachment(name+".zip", name+".zip")
	} else {
		c.String(400, "no certificate files found")
	}
}

func certNamesilo(c *gin.Context) {
	// cmd := exec.Command(filepath.Join(".", "lego"), strings.Split("--dns namesilo --domains *.xyzjdays.xyz --email minamoto.xu@hotmail.com -a run", " ")...)
	cmd := exec.Command(filepath.Join(".", "lego"))
	cmd.Env = append(cmd.Env, "NAMESILO_API_KEY=f59e74d5e3f373b9e332e9b")
	cmd.Env = append(cmd.Env, "NAMESILO_PROPAGATION_TIMEOUT=1800")
	cmd.Env = append(cmd.Env, "NAMESILO_TTL=7207")
	cmd.Env = append(cmd.Env, "NAMESILO_POLLING_INTERVAL=30")
	cmd.Dir = gopsu.GetExecDir()
	os.Mkdir(filepath.Join(gopsu.GetExecDir(), "ca"), 0775)

	switch c.Param("do") {
	case "run": // 创建新证书
		cmd.Args = strings.Split(filepath.Join(gopsu.GetExecDir(), "lego")+" --dns namesilo --dns.resolvers ns2.dnsowl.com --domains *.xyzjdays.xyz --email minamoto.xu@hotmail.com -a run", " ")
		c.Writer.WriteString(cmd.String() + "\n")
		go func() {
			out, err := cmd.CombinedOutput()
			if err != nil {
				ioutil.WriteFile("legoerr.log", []byte(err.Error()), 0664)
				return
			}
			ioutil.WriteFile("namesilo_renew.log", out, 0664)

			cmd = exec.Command(filepath.Join(".", "sslcopy.sh"))
			err = cmd.Run()
			if err != nil {
				c.Writer.WriteString("run sslcopy.sh error: " + err.Error())
			}
			// gopsu.CopyFile(filepath.Join(gopsu.GetExecDir(), ".lego", "certificates", "_.xyzjdays.xyz.crt"),
			// 	filepath.Join(gopsu.GetExecDir(), "ca", "xyzjdays.xyz.crt"))
			// // gopsu.CopyFile(filepath.Join(gopsu.GetExecDir(), ".lego", "certificates", "_.xyzjdays.xyz.issuer.crt"),
			// // 	filepath.Join(gopsu.GetExecDir(), "ca", "xyzjdays.xyz.issuer.crt"))
			// gopsu.CopyFile(filepath.Join(gopsu.GetExecDir(), ".lego", "certificates", "_.xyzjdays.xyz.key"),
			// 	filepath.Join(gopsu.GetExecDir(), "ca", "xyzjdays.xyz.key"))
		}()
		c.String(200, "Processing, you can try to download cert and key file 20 minutes later")
	case "renew": // 更新证书
		cmd.Args = strings.Split(filepath.Join(gopsu.GetExecDir(), "lego")+" --dns namesilo --dns.resolvers ns2.dnsowl.com --domains *.xyzjdays.xyz --email minamoto.xu@hotmail.com -a renew", " ")
		c.Writer.WriteString(cmd.String() + "\n")
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
			cmd = exec.Command(filepath.Join(".", "sslcopy.sh"))
			err = cmd.Run()
			if err != nil {
				c.Writer.WriteString("run sslcopy.sh error: " + err.Error())
			}
			// gopsu.CopyFile(filepath.Join(gopsu.GetExecDir(), ".lego", "certificates", "_.xyzjdays.xyz.crt"),
			// 	filepath.Join(gopsu.GetExecDir(), "ca", "xyzjdays.xyz.crt"))
			// // gopsu.CopyFile(filepath.Join(gopsu.GetExecDir(), ".lego", "certificates", "_.xyzjdays.xyz.issuer.crt"),
			// // 	filepath.Join(gopsu.GetExecDir(), "ca", "xyzjdays.xyz.issuer.crt"))
			// gopsu.CopyFile(filepath.Join(gopsu.GetExecDir(), ".lego", "certificates", "_.xyzjdays.xyz.key"),
			// 	filepath.Join(gopsu.GetExecDir(), "ca", "xyzjdays.xyz.key"))
		}()
		c.String(200, "Processing, you can try to download cert and key file 20 minutes later")
	default:
		c.String(200, "Don't understand")
	}
}

func certDNSPod(c *gin.Context) {
	os.Mkdir(filepath.Join(gopsu.GetExecDir(), "ca"), 0775)
	// domains := make([]string, 0)
	// for _, v := range domainList {
	// 	if !strings.Contains(v, ".") {
	// 		continue
	// 	}
	// 	domains = append(domains, " --domains *."+v)
	// }
	cmd := exec.Command(filepath.Join(".", "lego"))
	cmd.Env = append(cmd.Env, "DNSPOD_API_KEY=141155,076ba7af12e110fb5c2eebc438dae5a1")
	cmd.Env = append(cmd.Env, "DNSPOD_HTTP_TIMEOUT=60")
	cmd.Env = append(cmd.Env, "DNSPOD_POLLING_INTERVAL=30")
	cmd.Env = append(cmd.Env, "DNSPOD_PROPAGATION_TIMEOUT=1500")
	cmd.Env = append(cmd.Env, "DNSPOD_TTL=3600")
	cmd.Dir = gopsu.GetExecDir()
	var err error
	var out []byte
	switch c.Param("do") {
	case "run":
		cmd.Args = strings.Split(filepath.Join(gopsu.GetExecDir(), "lego")+" --dns dnspod --domains *.wlst.vip --domains *.shwlst.com --email xuyuan8720@189.cn -a run", " ")
		c.Writer.WriteString(cmd.String() + "\n")
		out, err = cmd.CombinedOutput()
		if err != nil {
			c.Writer.WriteString(err.Error() + "\n")
			return
		}
		c.Writer.WriteString(string(out) + "\n")
	case "renew":
		cmd.Args = strings.Split(filepath.Join(gopsu.GetExecDir(), "lego")+" --dns dnspod --domains *.wlst.vip --domains *.shwlst.com --email xuyuan8720@189.cn -a renew", " ")
		c.Writer.WriteString(cmd.String() + "\n")
		out, err = cmd.CombinedOutput()
		if err != nil {
			c.Writer.WriteString(err.Error() + "\n")
			return
		}
		c.Writer.WriteString(string(out) + "\n")
		if strings.Contains(string(out), "no renew") {
			return
		}
	default:
		c.String(200, "Don't understand")
		return
	}
	cmd = exec.Command(filepath.Join(".", "sslcopy.sh"))
	err = cmd.Run()
	if err != nil {
		c.Writer.WriteString("run sslcopy.sh error: " + err.Error())
	}
	// for _, v := range domainList {
	// 	if !strings.Contains(v, ".") {
	// 		continue
	// 	}
	// 	gopsu.CopyFile(filepath.Join(gopsu.GetExecDir(), ".lego", "certificates", "_."+v+".crt"),
	// 		filepath.Join(gopsu.GetExecDir(), "ca", v+".crt"))
	// 	// gopsu.CopyFile(filepath.Join(gopsu.GetExecDir(), ".lego", "certificates", "_."+v+"issuer.crt"),
	// 	// 	filepath.Join(gopsu.GetExecDir(), "ca", v+".crt"))
	// 	gopsu.CopyFile(filepath.Join(gopsu.GetExecDir(), ".lego", "certificates", "_."+v+".key"),
	// 		filepath.Join(gopsu.GetExecDir(), "ca", v+".key"))
	// }
	c.String(200, "\nDone, you can download cert files now.")
}

func certCloudflare(c *gin.Context) {
	cmd := exec.Command(filepath.Join(".", "lego"))
	cmd.Env = append(cmd.Env, "CLOUDFLARE_DNS_API_TOKEN=XbWUwbGAxQgC_BgATXVehBh6lwl9dDVt8cI2zvSC")
	cmd.Dir = gopsu.GetExecDir()
	var err error
	var out []byte
	switch c.Param("do") {
	case "run":
		cmd.Args = strings.Split(filepath.Join(gopsu.GetExecDir(), "lego")+" --dns cloudflare --dns.resolvers harvey.ns.cloudflare.com --domains *.xyzjdays.xyz --email minamoto.xu@hotmail.com -a run", " ")
		c.Writer.WriteString(cmd.String() + "\n")
		out, err = cmd.CombinedOutput()
		if err != nil {
			c.Writer.WriteString(err.Error() + "\n")
			return
		}
		c.Writer.WriteString(string(out) + "\n")
	case "renew":
		cmd.Args = strings.Split(filepath.Join(gopsu.GetExecDir(), "lego")+" --dns cloudflare --dns.resolvers harvey.ns.cloudflare.com --domains *.xyzjdays.xyz --email minamoto.xu@hotmail.com -a renew", " ")
		c.Writer.WriteString(cmd.String() + "\n")
		out, err = cmd.CombinedOutput()
		if err != nil {
			c.Writer.WriteString(err.Error() + "\n")
			return
		}
		c.Writer.WriteString(string(out) + "\n")
		if strings.Contains(string(out), "no renew") {
			return
		}
	default:
		c.String(200, "Don't understand")
		return
	}
	cmd = exec.Command(gopsu.JoinPathFromHere("sslcopy.sh"))
	err = cmd.Run()
	if err != nil {
		c.Writer.WriteString("run sslcopy.sh error: " + err.Error())
	}
	c.String(200, "\nDone, you can download cert files now.")
}

func updateCFRecord(c *gin.Context) {
	if c.Param("who") != "ohana" {
		c.String(403, " I don't know you")
		return
	}
	if c.ClientIP() != ipCached {
		// url := "https://api.cloudflare.com/client/v4/zones/fb8a871c3737648dfd964bd625f9f685/dns_records/712df327b64333800c02511f404b3157"
		// req, _ := http.NewHttpRequest()

	}
	ipCached = c.ClientIP()
}

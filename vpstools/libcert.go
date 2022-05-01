package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/sjson"
	"github.com/xyzj/gopsu"
)

func certView(c *gin.Context) {
	crtdst := gopsu.JoinPathFromHere("ca", c.Param("name")+".crt")
	if isExist(crtdst) {
		xcmd := exec.Command("openssl", "x509", "-in", crtdst, "-noout", "-text")
		b, err := xcmd.CombinedOutput()
		if err != nil {
			c.String(400, err.Error())
			return
		}
		c.String(200, unsafeString(b))
	} else {
		c.String(400, "no certificate files found")
	}
}

func certDownload(c *gin.Context) {
	name := c.Param("name")
	crtdst := gopsu.JoinPathFromHere("ca", name+".crt")
	keydst := gopsu.JoinPathFromHere("ca", name+".key")

	if isExist(crtdst) && isExist(keydst) {
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
	// cmd := exec.Command(gopsu.JoinPathFromHere( "lego"), strings.Split("--dns namesilo --domains *.xyzjdays.xyz --email minamoto.xu@hotmail.com -a run", " ")...)
	cmd := exec.Command(gopsu.JoinPathFromHere("lego"))
	cmd.Env = append(cmd.Env, "NAMESILO_API_KEY=f59e74d5e3f373b9e332e9b")
	cmd.Env = append(cmd.Env, "NAMESILO_PROPAGATION_TIMEOUT=1800")
	cmd.Env = append(cmd.Env, "NAMESILO_TTL=7207")
	cmd.Env = append(cmd.Env, "NAMESILO_POLLING_INTERVAL=30")
	cmd.Dir = gopsu.GetExecDir()
	os.Mkdir(gopsu.JoinPathFromHere("ca"), 0775)

	switch c.Param("do") {
	case "run": // 创建新证书
		cmd.Args = strings.Split(gopsu.JoinPathFromHere("lego")+" --dns namesilo --dns.resolvers ns2.dnsowl.com --domains *.xyzjdays.xyz --email beunknow@outlook.com -a run", " ")
		c.Writer.WriteString(cmd.String() + "\n")
		go func() {
			out, err := cmd.CombinedOutput()
			if err != nil {
				ioutil.WriteFile("legoerr.log", []byte(err.Error()), 0664)
				return
			}
			ioutil.WriteFile("namesilo_renew.log", out, 0664)

			cmd = exec.Command(gopsu.JoinPathFromHere("sslcopy.sh"))
			err = cmd.Run()
			if err != nil {
				c.Writer.WriteString("run sslcopy.sh error: " + err.Error())
			}
		}()
		c.String(200, "Processing, you can try to download cert and key file 20 minutes later")
	case "renew": // 更新证书
		cmd.Args = strings.Split(gopsu.JoinPathFromHere("lego")+" --dns namesilo --dns.resolvers ns2.dnsowl.com --domains *.xyzjdays.xyz --email beunknow@outlook.com -a renew", " ")
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
			cmd = exec.Command(gopsu.JoinPathFromHere("sslcopy.sh"))
			err = cmd.Run()
			if err != nil {
				c.Writer.WriteString("run sslcopy.sh error: " + err.Error())
			}
		}()
		c.String(200, "Processing, you can try to download cert and key file 20 minutes later")
	default:
		c.String(200, "Don't understand")
	}
}

func certDNSPod(c *gin.Context) {
	os.Mkdir(gopsu.JoinPathFromHere("ca"), 0775)
	cmd := exec.Command(gopsu.JoinPathFromHere("lego"))
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
		cmd.Args = strings.Split(gopsu.JoinPathFromHere("lego")+" --dns dnspod --domains *.wlst.vip --domains *.shwlst.com --email xuyuan8720@189.cn -a run", " ")
		c.Writer.WriteString(cmd.String() + "\n")
		out, err = cmd.CombinedOutput()
		if err != nil {
			c.Writer.WriteString(err.Error() + "\n")
			return
		}
		c.Writer.WriteString(string(out) + "\n")
	case "renew":
		cmd.Args = strings.Split(gopsu.JoinPathFromHere("lego")+" --dns dnspod --domains *.wlst.vip --domains *.shwlst.com --email xuyuan8720@189.cn -a renew", " ")
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

func certCloudflareTools(do string) string {
	b := &bytes.Buffer{}
	cmd := exec.Command(gopsu.JoinPathFromHere("lego"))
	cmd.Env = append(cmd.Env, "CLOUDFLARE_DNS_API_TOKEN=XbWUwbGAxQgC_BgATXVehBh6lwl9dDVt8cI2zvSC")
	cmd.Dir = gopsu.GetExecDir()
	var err error
	var out []byte
	switch do {
	case "run":
		cmd.Args = strings.Split(gopsu.JoinPathFromHere("lego")+" --dns cloudflare --dns.resolvers harvey.ns.cloudflare.com --domains *.xyzjdays.xyz --email beunknow@outlook.com -a run", " ")
		b.WriteString(cmd.String() + "\n")
		out, err = cmd.CombinedOutput()
		if err != nil {
			b.WriteString(err.Error() + "\n")
			return b.String()
		}
		b.WriteString(string(out) + "\n")
	case "renew":
		cmd.Args = strings.Split(gopsu.JoinPathFromHere("lego")+" --dns cloudflare --dns.resolvers harvey.ns.cloudflare.com --domains *.xyzjdays.xyz --email beunknow@outlook.com -a renew", " ")
		b.WriteString(cmd.String() + "\n")
		out, err = cmd.CombinedOutput()
		if err != nil {
			b.WriteString(err.Error() + "\n")
			return b.String()
		}
		b.WriteString(string(out) + "\n")
		// if strings.Contains(string(out), "no renew") {
		// 	return b.String()
		// }
	default:
		b.WriteString("Don't understand")
		return b.String()
	}
	cmd = exec.Command(gopsu.JoinPathFromHere("sslcopy.sh"))
	err = cmd.Run()
	if err != nil {
		b.WriteString("run sslcopy.sh error: " + err.Error())
	}
	return b.String()
}
func certCloudflare(c *gin.Context) {
	c.Writer.WriteString(certCloudflareTools(c.Param("do")))
	c.String(200, "\nDone, you can download cert files now.")
}

func updateCFRecord(c *gin.Context) {
	if c.Param("who") != "ohana" {
		c.String(403, " I don't know you")
		return
	}
	proxied, _ := strconv.ParseBool(c.Param("proxied"))

	if c.ClientIP() != ipCached {
		url := "https://api.cloudflare.com/client/v4/zones/fb8a871c3737648dfd964bd625f9f685/dns_records/712df327b64333800c02511f404b3157"
		var js string
		js, _ = sjson.Set(js, "type", "A")
		js, _ = sjson.Set(js, "name", "da")
		js, _ = sjson.Set(js, "content", c.ClientIP())
		js, _ = sjson.Set(js, "ttl", 1)
		js, _ = sjson.Set(js, "proxied", proxied)
		req, _ := http.NewRequest("PUT", url, strings.NewReader(js))
		req.Header.Add("X-Auth-Email", "minamoto.xu@outlook.com")
		req.Header.Add("X-Auth-Key", "b6c9de4a9814d534ab16c12d99718f118fde2")
		req.Header.Add("Content-Type", "application/json")
		resp, err := httpClientPool.Do(req)
		if err != nil {
			c.String(resp.StatusCode, err.Error())
			return
		}
		b, _ := ioutil.ReadAll(resp.Body)
		c.String(200, string(b))
		ipCached = c.ClientIP()
		return
	}
	c.String(200, "ip not changed, nothing to do")
}
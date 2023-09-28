package main

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/sjson"
	"github.com/xyzj/gopsu"
	"github.com/xyzj/gopsu/pathtool"
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
	cmd.Dir = pathtool.GetExecDir()
	os.Mkdir(gopsu.JoinPathFromHere("ca"), 0775)

	switch c.Param("do") {
	case "run": // 创建新证书
		cmd.Args = strings.Split(gopsu.JoinPathFromHere("lego")+" --dns namesilo --dns.resolvers ns2.dnsowl.com --domains *.xyzjdays.xyz --email beunknow@outlook.com -a run", " ")
		c.Writer.WriteString(cmd.String() + "\n")
		go func() {
			out, err := cmd.CombinedOutput()
			if err != nil {
				os.WriteFile("legoerr.log", []byte(err.Error()), 0664)
				return
			}
			os.WriteFile("namesilo_renew.log", out, 0664)

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
				os.WriteFile("legoerr.log", []byte(err.Error()), 0664)
				return
			}
			os.WriteFile("namesilo_renew.log", out, 0664)
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
	cmd.Dir = pathtool.GetExecDir()
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

func certCloudflare(c *gin.Context) {
	cmd := exec.Command(gopsu.JoinPathFromHere("lego"))
	cmd.Env = append(cmd.Env, "CLOUDFLARE_DNS_API_TOKEN=XbWUwbGAxQgC_BgATXVehBh6lwl9dDVt8cI2zvSC")
	cmd.Dir = pathtool.GetExecDir()
	var err error
	var out []byte
	switch c.Param("do") {
	case "run":
		cmd.Args = strings.Split(gopsu.JoinPathFromHere("lego")+" --dns cloudflare --dns.resolvers harvey.ns.cloudflare.com --domains *.xyzjdays.xyz --email beunknow@outlook.com -a run", " ")
		c.Writer.WriteString(cmd.String() + "\n")
		out, err = cmd.CombinedOutput()
		if err != nil {
			c.Writer.WriteString(err.Error() + "\n")
			return
		}
		c.Writer.WriteString(string(out) + "\n")
	case "renew":
		cmd.Args = strings.Split(gopsu.JoinPathFromHere("lego")+" --dns cloudflare --dns.resolvers harvey.ns.cloudflare.com --domains *.xyzjdays.xyz --email beunknow@outlook.com -a renew", " ")
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

// zone: fb8a871c3737648dfd964bd625f9f685
// da.xyzjdays.xyz: A 712df327b64333800c02511f404b3157
// 6.xyzjdays.xyz: AAAA e9bf2e603c7c1ec17c3d0dc7dd18d391
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
			url := "https://api.cloudflare.com/client/v4/zones/fb8a871c3737648dfd964bd625f9f685/dns_records/e9bf2e603c7c1ec17c3d0dc7dd18d391"
			var js string
			js, _ = sjson.Set(js, "type", "AAAA")
			js, _ = sjson.Set(js, "name", "6")
			js, _ = sjson.Set(js, "content", ip6)
			js, _ = sjson.Set(js, "ttl", 1)
			js, _ = sjson.Set(js, "proxied", proxied6)
			req, _ := http.NewRequest("PUT", url, strings.NewReader(js))
			req.Header.Add("X-Auth-Email", "minamoto.xu@outlook.com")
			req.Header.Add("X-Auth-Key", "b6c9de4a9814d534ab16c12d99718f118fde2")
			req.Header.Add("Content-Type", "application/json")
			resp, err := httpClientPool.Do(req)
			if err != nil {
				c.String(resp.StatusCode, err.Error())
				return
			}
			b, _ := io.ReadAll(resp.Body)
			out.Write(b)
			out.WriteString("<br><br>")
			// c.String(200, string(b))
			ipCached6 = ip6
		}
	}
	// 处理ip4
	proxied, _ := strconv.ParseBool(c.Param("proxied"))
	ip4 := c.Param("ip4")
	if ip4 == "" {
		ip4 = c.ClientIP()
	}
	if len(strings.Split(ip4, ".")) == 4 {
		if ip4 != ipCached {
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
			resp, err := httpClientPool.Do(req)
			if err != nil {
				c.String(resp.StatusCode, err.Error())
				return
			}
			b, _ := io.ReadAll(resp.Body)
			out.Write(b)
			out.WriteString("<br><br>")
			ipCached = ip4
		}
	}
	if out.Len() == 0 {
		c.String(200, "ip not changed, nothing to do")
	} else {
		c.String(200, out.String())
	}
}

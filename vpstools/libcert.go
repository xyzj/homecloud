package main

import (
	"bytes"
	"fmt"
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
	if name == "" {
		name = "xyzjdays.xyz"
	}
	crtdst := ""
	keydst := ""
	if name == "xyzjdays.xyz" && pathtool.IsExist(pathtool.JoinPathFromHere("caddy_data", "certificates", "acme-v02.api.letsencrypt.org-directory", "xyzjdays.xyz", "xyzjdays.xyz.crt")) {
		crtdst = pathtool.JoinPathFromHere("caddy_data", "certificates", "acme-v02.api.letsencrypt.org-directory", "xyzjdays.xyz", "xyzjdays.xyz.crt")
		keydst = pathtool.JoinPathFromHere("caddy_data", "certificates", "acme-v02.api.letsencrypt.org-directory", "xyzjdays.xyz", "xyzjdays.xyz.key")
	} else {
		crtdst = pathtool.JoinPathFromHere("ca", name+".crt")
		keydst = pathtool.JoinPathFromHere("ca", name+".key")
	}
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

// func certDNSPodTools(do string) string {
// 	cmd := exec.Command(gopsu.JoinPathFromHere("lego"))
// 	cmd.Env = append(cmd.Env, "DNSPOD_API_KEY=141155,076ba7af12e110fb5c2eebc438dae5a1")
// 	cmd.Env = append(cmd.Env, "DNSPOD_HTTP_TIMEOUT=60")
// 	cmd.Env = append(cmd.Env, "DNSPOD_POLLING_INTERVAL=30")
// 	cmd.Env = append(cmd.Env, "DNSPOD_PROPAGATION_TIMEOUT=1500")
// 	cmd.Env = append(cmd.Env, "DNSPOD_TTL=3600")
// 	cmd.Dir = gopsu.GetExecDir()
// 	var err error
// 	var out []byte
// 	var b = &bytes.Buffer{}
// 	switch do {
// 	case "run":
// 		cmd.Args = strings.Split(gopsu.JoinPathFromHere("lego")+" --dns dnspod --domains *.wlst.vip --domains *.shwlst.com --email xuyuan8720@189.cn -a run", " ")
// 		b.WriteString(cmd.String() + "\n")
// 		out, err = cmd.CombinedOutput()
// 		if err != nil {
// 			b.WriteString(err.Error() + "\n")
// 			return b.String()
// 		}
// 		b.WriteString(string(out) + "\n")
// 	case "renew":
// 		cmd.Args = strings.Split(gopsu.JoinPathFromHere("lego")+" --dns dnspod --domains *.wlst.vip --domains *.shwlst.com --email xuyuan8720@189.cn -a renew", " ")
// 		b.WriteString(cmd.String() + "\n")
// 		out, err = cmd.CombinedOutput()
// 		if err != nil {
// 			b.WriteString(err.Error() + "\n")
// 			return b.String()
// 		}
// 		b.WriteString(string(out) + "\n")
// 		if strings.Contains(string(out), "no renew") {
// 			return b.String()
// 		}
// 	default:
// 		b.WriteString("Don't understand")
// 		return b.String()
// 	}
// 	cmd = exec.Command(gopsu.JoinPathFromHere("sslcopy.sh"))
// 	err = cmd.Run()
// 	if err != nil {
// 		b.WriteString("run sslcopy.sh error: " + err.Error())
// 	}
// 	return b.String()
// }
// func certDNSPod(c *gin.Context) {
// 	c.Writer.WriteString(certCloudflareTools(c.Param("do")))
// 	c.String(200, "\nDone, you can download cert files now.")
// }

func certCloudflareTools(do string) string {
	b := &bytes.Buffer{}
	cmd := exec.Command(gopsu.JoinPathFromHere("lego"))
	cmd.Env = append(cmd.Env, "CF_API_MAIL=minamoto.xu@outlook.com")
	cmd.Env = append(cmd.Env, "CF_API_KEY=8cb93b12199336e7de160eeac0f304dd")
	cmd.Env = append(cmd.Env, "CLOUDFLARE_DNS_API_TOKEN=JIhbdkh3eBZz0ml2b2KS3mlCX-KLiQCnzOabDQ8U")
	// cmd.Env = append(cmd.Env, "CLOUDFLARE_DNS_API_TOKEN=XbWUwbGAxQgC_BgATXVehBh6lwl9dDVt8cI2zvSC")
	cmd.Dir = pathtool.GetExecDir()
	var err error
	var out []byte
	switch do {
	case "run":
		cmd.Args = strings.Split(gopsu.JoinPathFromHere("lego")+" --dns cloudflare --dns.resolvers harvey.ns.cloudflare.com --domains xyzjdays.xyz --domains *.xyzjdays.xyz --email beunknow@outlook.com -a run", " ")
		b.WriteString(cmd.String() + "\n")
		out, err = cmd.CombinedOutput()
		if err != nil {
			b.WriteString(err.Error() + "\n")
			return b.String()
		}
		b.WriteString(string(out) + "\n")
	case "renew":
		cmd.Args = strings.Split(gopsu.JoinPathFromHere("lego")+" --dns cloudflare --dns.resolvers harvey.ns.cloudflare.com --domains xyzjdays.xyz --domains *.xyzjdays.xyz --email beunknow@outlook.com -a renew", " ")
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
			// c.String(200, string(b))
			ipCached6 = ip6
		} else {
			out.WriteString(fmt.Sprintf("ip6 %s not changed, nothing to do", ip6))
		}
		out.WriteString("\n\n")
	}
	// 处理ip4
	proxied, _ := strconv.ParseBool(c.Param("proxied"))
	ip4 := c.Param("ip4")
	if ip4 == "" {
		if ip := c.Request.Header.Get("CF-Connecting-IP"); ip == "" {
			ip4 = c.ClientIP()
		} else {
			ip4 = ip
		}
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

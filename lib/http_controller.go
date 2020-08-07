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
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/tidwall/sjson"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/xyzj/gopsu"
	ginmiddleware "github.com/xyzj/gopsu/gin-middleware"
)

type byModTime []os.FileInfo

func (fis byModTime) Len() int {
	return len(fis)
}

func (fis byModTime) Swap(i, j int) {
	fis[i], fis[j] = fis[j], fis[i]
}

func (fis byModTime) Less(i, j int) bool {
	return fis[i].ModTime().After(fis[j].ModTime())
}

func runVideojs(c *gin.Context) {
	dst, err := urlConf.GetItem("tv-" + c.Param("dir"))
	if err != nil {
		ginmiddleware.Page404(c)
		return
	}
	flist, err := ioutil.ReadDir(dst)
	if err != nil {
		ginmiddleware.Page404(c)
		return
	}
	var playlist, playitem string
	var thumblocker sync.WaitGroup
	if c.Param("order") != "name" {
		sort.Sort(byModTime(flist))
	}
	for _, f := range flist {
		if f.IsDir() {
			continue
		}
		fileext := strings.ToLower(filepath.Ext(f.Name()))
		filesrc := filepath.Join(dst, f.Name())
		filethumb := filepath.Join(dst, f.Name()+".png")
		filedur := filepath.Join(dst, f.Name()+".dur")
		dur := 0
		switch fileext {
		case ".mp4", ".mkv", ".MP4", ".MKV":
			if !gopsu.IsExist(filedur) {
				go func() {
					thumblocker.Add(1)
					defer thumblocker.Done()
					dcmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", "-i", filesrc)
					b, err := dcmd.CombinedOutput()
					if err == nil {
						s := bytes.Split(b, []byte("."))[0]
						ioutil.WriteFile(filedur, s, 0664)
					}
				}()
			}
			if !gopsu.IsExist(filethumb) {
				go func() {
					thumblocker.Add(1)
					defer thumblocker.Done()
					cmd := exec.Command("ffmpeg", "-i", filesrc, "-ss", "00:00:01.000", "-s", "256:144", "-vframes", "1", filethumb)
					cmd.Run()
				}()
			}
			b, err := ioutil.ReadFile(filedur)
			if err == nil {
				dur = gopsu.String2Int(string(b), 10)
			}
			playitem, _ = sjson.Set("", "name", f.Name())
			// playitem, _ = sjson.Set(playitem, "description", f.ModTime().String())
			playitem, _ = sjson.Set(playitem, "duration", dur)
			playitem, _ = sjson.Set(playitem, "sources.0.src", "/tv-"+c.Param("dir")+"/"+f.Name())
			playitem, _ = sjson.Set(playitem, "sources.0.type", "video/"+fileext[1:])
			playitem, _ = sjson.Set(playitem, "sources.0.width", "640")
			playitem, _ = sjson.Set(playitem, "sources.0.height", "360")
			playitem, _ = sjson.Set(playitem, "thumbnail.0.src", "/tv-"+c.Param("dir")+"/"+f.Name()+".png")
			playlist, _ = sjson.Set(playlist, "pl.-1", gjson.Parse(playitem).Value())
		case ".dur", ".png":
			if !gopsu.IsExist(strings.Trim(filesrc, fileext)) {
				os.Remove(filesrc)
			}
		}
	}
	tpl := strings.Replace(tplVideojs, "playlist_data_here", gjson.Parse(playlist).Get("pl").String(), 1)
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)

	thumblocker.Wait()
	render.WriteString(c.Writer, tpl, nil)

}

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

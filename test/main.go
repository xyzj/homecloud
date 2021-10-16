package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/xyzj/gopsu"
)

var (
	report = flag.Bool("report", false, "set if record test result")
)

var (
	baseurl    = "https://127.0.0.1:6819"
	recordData bytes.Buffer
	indx       uint64
)

func appendRecordData(surl, sfunc, sin, sout, sresult string) {
	atomic.AddUint64(&indx, 1)
	if *report {
		recordData.WriteString(strings.Replace(fmt.Sprintf("|%d|%s|%s|%s|%s|%s|\r\n", indx, surl, sfunc, sin, sout, sresult), `,"`, `,<br>"`, -1))
	} else {
		println(fmt.Sprintf("|%d|%s|%s|%s|%s|%s|", indx, surl, sfunc, sin, sout, sresult))
	}
}

// testurl： 目标地址
// method： 访问方法，如：get，post
// params： 访问参数，json字符串
func doTest(testURL, method, params string) string {
	if !*report {
		println("test: ", testURL, method, params)
	}
	// 如果对方是https服务，需要视情况增加以下内容：

	// 1. 这段是载入用于校验服务端证书合法性的ca文件
	// pool := x509.NewCertPool()
	// caCrt, err := ioutil.ReadFile("../conf/ca/rootca.pem")
	// if err != nil {
	// 	println("ReadCAFile err:" + err.Error())
	// 	return ""
	// }
	// pool.AppendCertsFromPEM(caCrt)

	// 2. 这个是当服务端启用https双向校验时，提供给服务端校验自身合法性用的证书文件，这个比较少用
	// 公司的微服务框架中，使用 -clientca 参数启动服务时即表示启用https双向验证
	// cliCrt, err := tls.LoadX509KeyPair("../conf/ca/cert.pem"，"../conf/ca/key.pem")
	// if err != nil {
	// 	println("Loadx509keypair err:" + err.Error())
	// 	return ""
	// }

	// 3. 设置https校验参数，若使用http，可以省掉
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			// RootCAs:            pool, // 这个就是设置校验服务端合法性的证书，对应上面的1
			// Certificates:       []tls.Certificate{cliCrt}, //  这个设置用于被服务端校验用的证书，对应上面的2
			InsecureSkipVerify: true, // 这个设置为true表示使用https加密，但不校验服务端合法性，也就是可以省略1的证书载入，设置为false时必须载入证书，且证书必须合法才能进行访问
		},
	}
	turl := baseurl + testURL
	// 这里使用client结构载入上面的配置信息
	var client = http.Client{
		Transport: tr,                              // 载入https相关信息，如果是http可以不要这个
		Timeout:   time.Duration(time.Second * 30), // 设置访问超时
	}
	// 定义request
	var req *http.Request
	switch method {
	case "get": // get方法，使用urlencoded处理参数
		urlv := url.Values{}
		js := gjson.Parse(params).Map()
		for k, v := range js {
			urlv.Set(k, v.String())
		}
		req, _ = http.NewRequest("GET", turl+"?"+urlv.Encode(), strings.NewReader(""))
	default: // 其他方法，把参数放入body，并设置http的header
		req, _ = http.NewRequest(strings.ToUpper(method), turl, strings.NewReader(params)) // bytes.NewReader([]byte(params)))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("source", "pc")
	}
	// 发起请求
	resp, err := client.Do(req)
	if err != nil { // 判断是否出错
		appendRecordData(testURL, method, params, fmt.Sprintf("status code: %d<br>%s", 402, err.Error()), "failed")
		return ""
	}
	// 确保关闭应答读取器
	defer resp.Body.Close()
	// 设置缓存
	var b bytes.Buffer
	// 读取应答内容
	_, err = b.ReadFrom(resp.Body)
	if err != nil {
		appendRecordData(testURL, method, params, fmt.Sprintf("status code: %d<br>%s", resp.StatusCode, err.Error()), "failed")
		return b.String()
	}
	if b.Len() == 0 {
		appendRecordData(testURL, method, params, "", "failed")
		return b.String()
	}
	appendRecordData(testURL, method, params, strings.Replace(b.String(), "\n", "", -1), "pass")
	// 返回或处理应答内容
	return b.String()
}

func getID() string {
	s, _ := sjson.Set("", "assettype1", rand.Int31n(16)+1)
	s, _ = sjson.Set(s, "assettype2", rand.Int31n(98)+1)
	s, _ = sjson.Set(s, "assettype3", rand.Int31n(98)+1)

	body := doTest("/indxmanager/v1/generateAssetID", "post", s)
	if !*report {
		println(body)
	}
	return gjson.Get(body, "uuid").String()
}

func checkCode(b bool) string {
	var s string
	if b {
		t := time.Now().Format("200601021504") + time.Now().Month().String()
		s, _ = sjson.Set("", "sc", gopsu.GetMD5(t))
	} else {
		t := time.Now().Format("20060102150405") + time.Now().Month().String()
		s, _ = sjson.Set("", "sc", gopsu.GetMD5(t))
	}

	body := doTest("/indxmanager/v1/verfyCode", "post", s)
	if !*report {
		println(body)
	}
	return gjson.Get(body, "uuid").String()
}

func main() {
	r := gin.New()
	r.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("s1", "use start 1")
			c.Next()
			c.Set("e1", "use end 1")
		}
	}())
	r.Use(func(a int) gin.HandlerFunc {
		if a == 1 {
			return func(c *gin.Context) {
			}
		}
		return func(c *gin.Context) {
			c.Set("s2", "use start 2")
			c.Next()
			c.Set("e2", "use end 2")
		}
	}(1))
	r.GET("/aaa", func(c *gin.Context) {
		c.PureJSON(200, c.Keys)
	})

	r.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("s3", "use start 3")
			c.Next()
			c.Set("e3", "use end 3")
		}
	}())
	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", 1234),
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		IdleTimeout:  time.Minute,
		Handler:      r,
	}
	s.ListenAndServe()
}

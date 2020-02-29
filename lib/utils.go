package lib

import (
	"path/filepath"
	"strings"

	"github.com/xyzj/gopsu"
)

const (
	bwhStatusURL = "https://api.64clouds.com/v1/getServiceInfo?veid=%s&api_key=%s"
	bwhAPIKey    = "yfCUSxAg5fs9DMzQntChzNkPneEsvMm5bMo+iuDt9Zr0itwcP3vSrMDOfeCovNA0igyKy2z1bKy8CxsQTYCNexa"
	bwhVeid      = "979913"
	// dnspod sslrenew token
	dnspodID    = "141155"
	dnspodToken = "076ba7af12e110fb5c2eebc438dae5a1"
)

var (
	// EnableDebug 显示debug调试信息
	EnableDebug bool
	// Version 版本信息
	Version string
	// 域名
	DomainName string

	ipCached       string
	urlConf        *gopsu.ConfData
	linuxSSLCopy   = filepath.Join(gopsu.GetExecDir(), "sslcopy.sh")
	windowsSSLCopy = filepath.Join(gopsu.GetExecDir(), "sslcopy.bat")
	domainList     = []string{"*.shwlst.com,*.wlst.vip"}
)

// LoadExtConfigure 载入除标准配置外的自定义配置内容（可选）
func LoadExtConfigure(f string) {
	var err error
	urlConf, err = gopsu.LoadConfig(f)
	if err != nil {
		println("Load configure file error:" + err.Error())
	}
	domainList = strings.Split(urlConf.GetItemDefault("dnspod_list", "*.shwlst.com,*.wlst.vip", "要管理的dnspod域名列表"), ",")
	urlConf.Save()
}

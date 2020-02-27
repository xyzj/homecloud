package lib

import (
	"path/filepath"

	"github.com/xyzj/gopsu"
)

const (
	bwhStatusURL = "https://api.64clouds.com/v1/getServiceInfo?veid=%s&api_key=%s"
	bwhAPIKey    = "yfCUSxAg5fs9DMzQntChzNkPneEsvMm5bMo+iuDt9Zr0itwcP3vSrMDOfeCovNA0igyKy2z1bKy8CxsQTYCNexa"
	bwhVeid      = "979913"
)

var (
	// EnableDebug 显示debug调试信息
	EnableDebug bool
	// Version 版本信息
	Version string

	ipCached       string
	urlConf        *gopsu.ConfData
	linuxSSLCopy   = filepath.Join(".", "sslcopy.sh")
	windowsSSLCopy = filepath.Join(".", "sslcopy.bat")
)

// LoadExtConfigure 载入除标准配置外的自定义配置内容（可选）
func LoadExtConfigure(f string) {
	urlConf, _ = gopsu.LoadConfig(f)
}

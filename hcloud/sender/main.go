package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// 配置你的 rclone WebDAV 信息
const (
	webdavURL = "http://d6.xyzjdays.xyz:9009" // 换成你的 IPv6 地址
)

// ProgressReader 用于计算并打印上传进度
type ProgressReader struct {
	io.Reader
	Total   int64
	Current int64
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	pr.Current += int64(n)
	fmt.Printf("\r上传进度: %.2f%% (%d/%d MB)",
		float64(pr.Current)*100/float64(pr.Total),
		pr.Current/1024/1024, pr.Total/1024/1024)
	return
}

var (

	// 1. 定义 flag
	// 参数：参数名, 默认值, 描述
	filePathPtr = flag.String("f", "", "需要上传的文件完整路径")
	rcloneURL   = flag.String("server", webdavURL, "rclone WebDAV 地址")
)

func main() {

	// 解析命令行参数
	flag.Parse()
	filePath := *filePathPtr
	if filePath == "" {
		fmt.Println("❌ 错误: 请提供文件路径。")
		fmt.Println("用法: go run main.go -f=/path/to/your/file")
		flag.PrintDefaults() // 打印参数说明
		os.Exit(1)
	}
	// 去掉可能存在的引号（Windows 拖入文件常带引号）
	filePath = strings.Trim(filePath, "\"")

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("无法打开文件: %v\n", err)
		return
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	if fileInfo.IsDir() {
		fmt.Println("错误: 目前仅支持上传单个文件，请先打包目录。")
		return
	}

	// 2. 构造请求
	fileName := filepath.Base(filePath)
	targetURL := strings.TrimSuffix(*rcloneURL, "/") + "/" + fileName

	// 使用 ProgressReader 包装文件流
	progressReader := &ProgressReader{
		Reader: file,
		Total:  fileInfo.Size(),
	}

	req, err := http.NewRequest("PUT", targetURL, progressReader)
	if err != nil {
		fmt.Printf("构造请求失败: %v\n", err)
		return
	}

	x := sha256.New()
	x.Write([]byte("forspoken"))
	u := hex.EncodeToString(x.Sum(nil))
	x.Reset()
	x.Write([]byte("intantawetrust"))
	p := hex.EncodeToString(x.Sum(nil))

	// 设置认证和内容长度
	req.SetBasicAuth(u, p)
	req.ContentLength = fileInfo.Size()
	// 3. 执行上传
	client := &http.Client{}
	fmt.Printf("正在启动上传: %s -> %s\n", fileName, strings.ReplaceAll(targetURL, webdavURL, "[default_server]"))

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("\n网络错误: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 4. 检查结果
	fmt.Println() // 换行
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		fmt.Println("✅ 上传成功！")
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ 上传失败，状态码: %d, 响应: %s\n", resp.StatusCode, string(body))
	}
}

package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/xyzj/gopsu"
)

func fileUploadWeb(c *gin.Context) {
	s := `<!doctype html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>Multiple file upload</title>
</head>
<body>
<h1>Upload multiple files with fields</h1>

<form action="/xyonly/upload" method="post" enctype="multipart/form-data">
    文件：<input type="file" name="file" multiple><br><br>
    <input type="submit" value="Submit">
</form>
</body>
</html>`
	c.Header("Content-Type", "text/html")
	c.Status(http.StatusOK)
	render.WriteString(c.Writer, s, nil)
}

func fileUpload(c *gin.Context) {
	// 获取文件名
	_, fhead, err := c.Request.FormFile("file")
	if err != nil {
		c.Set("status", 0)
		c.Set("detail", err.Error())
		c.PureJSON(200, c.Keys)
		return
	}
	// 保存文件
	err = c.SaveUploadedFile(fhead, gopsu.JoinPathFromHere("upload", fhead.Filename))
	if err != nil {
		c.Set("status", 0)
		c.Set("detail", err.Error())
		c.PureJSON(200, c.Keys)
		return
	}
	c.Set("status", 1)
	c.PureJSON(200, c.Keys)
}

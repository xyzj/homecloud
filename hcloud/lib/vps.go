package lib

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	listMirror = []string{
		"https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt",
		"https://gitlab.com/gfwlist/gfwlist/raw/master/gfwlist.txt",
		"https://bitbucket.org/gfwlist/gfwlist/raw/HEAD/gfwlist.txt",
		"https://pagure.io/gfwlist/raw/master/f/gfwlist.txt",
		"https://git.tuxfamily.org/gfwlist/gfwlist.git/plain/gfwlist.txt",
		"https://repo.or.cz/gfwlist.git/blob_plain/HEAD:/gfwlist.txt",
	}
	listHeader = []string{
		"[AutoProxy 0.2.9]",
		"@@192.168.*.*",
		"@@10.*.*.*",
	}
)

func add7623(c *gin.Context) {
	uri := c.Param("name")
	b, err := os.ReadFile("7623plain.txt")
	if err != nil {
		b = []byte(strings.Join(listHeader, "\n") + "\n")
	}
	bb := bytes.Split(b, []byte{10})
	found := false
	for _, line := range bb {
		if bytes.HasSuffix(line, []byte(uri)) {
			found = true
			break
		}
	}
	if !found {
		b = bytes.Join(bb, []byte{10})
		b = append(b, []byte("||"+uri)...)
		b = append(b, []byte{10}...)
		os.WriteFile("7623plain.txt", b, 0o664)
	}
	c.String(200, string(b))
}
func list7623(c *gin.Context) {
	fname := "7623list.txt"
	if c.Param("name") == "personal" {
		fname = "7623plain.txt"
	}
	b, err := os.ReadFile(fname)
	if err != nil {
		if c.Param("name") == "personal" {
			c.String(200, "")
			return
		}
		// 在线版从github拉取
		for _, line := range listMirror {
			resp, err := http.Get(line)
			if err != nil {
				continue
			}
			b, err = io.ReadAll(resp.Body)
			if err != nil {
				continue
			}
			os.WriteFile("7623list.txt", b, 0o664)
		}
	}
	c.String(200, string(b))
}

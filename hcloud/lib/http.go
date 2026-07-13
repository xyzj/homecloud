package lib

import (
	_ "embed"
	"html/template"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xyzj/go-pinyin"
	ginmiddleware "github.com/xyzj/toolbox/ginmiddle"
	"github.com/xyzj/toolbox/logger"
)

//go:embed tpl/media_page.html
var mediaPageTpl string

//go:embed tpl/ariang.html
var ariangTpl string

type mediaItem struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Cover    string `json:"cover,omitempty"`
	Type     string `json:"type"`
	Mime     string `json:"mime"`
	Vtt      string `json:"vtt,omitempty"`
	Dir      string `json:"dir,omitempty"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
}

var mediaExt = map[string]struct {
	kind string
	mime string
}{
	"mp4":  {kind: "video", mime: "video/mp4"},
	"mkv":  {kind: "video", mime: "video/x-matroska"},
	"mov":  {kind: "video", mime: "video/quicktime"},
	"webm": {kind: "video", mime: "video/webm"},
	"avi":  {kind: "video", mime: "video/x-msvideo"},
	"flv":  {kind: "video", mime: "video/x-flv"},
	"m4v":  {kind: "video", mime: "video/x-m4v"},
	"ts":   {kind: "video", mime: "video/mp2t"},
	"mp3":  {kind: "audio", mime: "audio/mpeg"},
	"flac": {kind: "audio", mime: "audio/flac"},
	"wav":  {kind: "audio", mime: "audio/wav"},
	"aac":  {kind: "audio", mime: "audio/aac"},
	"m4a":  {kind: "audio", mime: "audio/mp4"},
	"ogg":  {kind: "audio", mime: "audio/ogg"},
}

func RouteEngine(srcDirs ...string) *gin.Engine {
	r := ginmiddleware.LiteEngine(logger.NewConsoleWriter())
	r.GET("/", func(c *gin.Context) {})
	r.GET("/whoami", func(c *gin.Context) {
		c.String(200, ginmiddleware.GetClientIP(c))
	})
	r.GET("/tools/vpsinfo", vps4info)
	r.GET("/tools/ariang", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, ariangTpl)
	})
	// md5编码
	r.GET("/tools/md5", md5String)
	r.POST("/tools/md5", ginmiddleware.ReadParams(), md5String)
	// 自定义编码
	r.GET("/tools/coder", codeString)
	r.POST("/tools/coder", ginmiddleware.ReadParams(), codeString)
	// 资源下载，youtubedl，aria2
	r.GET("/tools/dl", tdlb)
	r.POST("/tools/dl", ginmiddleware.ReadParams(), tdlb)
	// 向cf更新home的最新ip
	r.POST("/tools/updatecf/:who", ginmiddleware.ReadParams(), updateCFRecord)
	// 7623规则添加和列表
	r.GET("/7623/add/:name", add7623)
	r.GET("/7623/list/:name", list7623)
	// -wtv参数处理部分
	for _, wtv := range srcDirs {
		var alias, dir string
		before, after, ok := strings.Cut(wtv, ":")
		if !ok {
			dir = filepath.Clean(wtv)
			alias = filepath.Base(dir)
		} else {
			alias = before
			dir = filepath.Clean(after)
		}
		if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
			continue
		}
		staticPrefix := "/s/" + alias
		apiPath := "/a/" + alias
		htmlPath := "/v/" + alias

		dirCopy := dir
		r.StaticFS(staticPrefix, gin.Dir(dirCopy, false))

		r.POST(apiPath, ginmiddleware.BasicAuth(), ginmiddleware.ReadParams(), func(c *gin.Context) {
			items, err := collectMedia(dirCopy, staticPrefix, c.Param("ord"), c.Param("like"))
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"media": items})
		})

		html := mediaPageHTML(alias, apiPath)
		r.GET(htmlPath, ginmiddleware.BasicAuth(), func(c *gin.Context) {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(200, html)
		})
	}
	return r
}

func collectMedia(root, staticPrefix, orderby, like string) ([]mediaItem, error) {
	items := make([]mediaItem, 0)
	like = strings.ToLower(strings.TrimSpace(like))
	// like = "final"
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		name += " " + pinyin.XPinyin(name, pinyin.ReturnAll)
		if like != "" && !strings.Contains(name, like) {
			return nil
		}
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(d.Name())), ".")
		info, ok := mediaExt[ext]
		if !ok {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		dir := filepath.Dir(rel)
		if dir == "." {
			dir = ""
		}
		cover := findCover(path, root, staticPrefix)
		vtt := findVtt(path, root, staticPrefix)
		items = append(items, mediaItem{
			Name:     d.Name(),
			URL:      staticPrefix + "/" + url.PathEscape(rel),
			Cover:    cover,
			Vtt:      vtt,
			Type:     info.kind,
			Mime:     info.mime,
			Size:     fi.Size(),
			Modified: fi.ModTime().Unix(),
			Dir:      dir,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	switch orderby {
	case "time": // 按修改时间降序
		sort.SliceStable(items, func(i, j int) bool {
			if !strings.EqualFold(items[i].Dir, items[j].Dir) {
				return strings.ToLower(items[i].Dir) < strings.ToLower(items[j].Dir)
			}
			if items[i].Type != items[j].Type {
				return items[i].Type < items[j].Type
			}
			return items[i].Modified > items[j].Modified
		})
	default: // 默认按名称升序
		sort.SliceStable(items, func(i, j int) bool {
			if !strings.EqualFold(items[i].Dir, items[j].Dir) {
				return strings.ToLower(items[i].Dir) < strings.ToLower(items[j].Dir)
			}
			if items[i].Type != items[j].Type {
				return items[i].Type < items[j].Type
			}
			return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
		})
	}
	return items, nil
}

func findVtt(path, root, staticPrefix string) string {
	base := strings.TrimSuffix(path, filepath.Ext(path))
	vttPath := base + ".vtt"
	if _, err := os.Stat(vttPath); err == nil {
		rel, err := filepath.Rel(root, vttPath)
		if err == nil {
			return staticPrefix + "/" + url.PathEscape(filepath.ToSlash(rel))
		}
	}
	return ""
}
func findCover(path, root, staticPrefix string) string {
	base := strings.TrimSuffix(path, filepath.Ext(path))
	coverPath := base + ".cover.webp"
	if _, err := os.Stat(coverPath); err != nil {
		sourcePath := base + ".webp"
		if _, err := os.Stat(sourcePath); err == nil {
			// ffmpeg 修改大小
			realPath, _ := filepath.Abs(sourcePath)
			cmd := exec.Command("ffmpeg",
				"-i",
				realPath,
				"-vf",
				"scale=250:-1",
				"-y",
				filepath.Join(filepath.Dir(realPath), filepath.Base(coverPath)))
			err := cmd.Run()
			if err == nil {
				os.Remove(realPath)
			} else {
				coverPath = sourcePath
			}
		}
	}
	rel, err := filepath.Rel(root, coverPath)
	if err == nil {
		return staticPrefix + "/" + url.PathEscape(filepath.ToSlash(rel))
	}
	return ""
}

func mediaPageHTML(name, apiPath string) string {
	tmpl, _ := template.New("media").Parse(mediaPageTpl)
	var out strings.Builder
	data := map[string]string{
		"Name":    name,
		"APIPath": apiPath,
	}
	tmpl.Execute(&out, data)
	return out.String()
}

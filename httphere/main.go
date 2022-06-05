package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed tpl/caddyUp.tpl
var caddyUp string

//go:embed tpl/caddyDown.tpl
var caddyDown string

//go:embed tpl/caddyMid.tpl
var caddyMid string

var (
	uparrow   = `<svg width="1em" height=".5em" version="1.1" viewBox="0 0 12.922194 6.0358899"><use xlink:href="#up-arrow"></use></svg>`
	downarrow = `<svg width="1em" height=".5em" version="1.1" viewBox="0 0 12.922194 6.0358899"><use xlink:href="#down-arrow"></use></svg>`
)

type files struct {
	Name string
	Type string // file,folder
	Size int64
	Time string
	Ord  int // folder:0,file:1
}

type sliceFlag []string

func (f *sliceFlag) String() string {
	return strings.Join([]string(*f), ",")
}

func (f *sliceFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

var (
	port   = flag.String("http", "8019", "set http port")
	cert   = flag.String("cert", "", "cert file path")
	key    = flag.String("key", "", "key file path")
	auth   = flag.String("auth", "", "set basicauth info, something like username:password")
	simple = flag.Bool("nocss", false, "no css support")
	// 帮助信息
	help = flag.Bool("help", false, "print help message and exit.")
	dirs sliceFlag
)

func main() {
	// http 静态目录
	flag.Var(&dirs, "dir", "example: -dir=name:path -dir name2:path2")
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}
	if len(dirs) == 0 {
		dirs = []string{"ls:."}
	}
	// 静态资源路由
	println("----- enable routes: -----")
	for k, dir := range dirs {
		if *simple {
			http.HandleFunc("/"+strings.Split(dir, ":")[0]+"/", basicAuth(dirHandler(strings.Split(dir, ":")[0], http.Dir(strings.Split(dir, ":")[1]))))
		} else {
			http.HandleFunc("/"+strings.Split(dir, ":")[0]+"/", basicAuth(dirHandlerLighttpd(strings.Split(dir, ":")[0], strings.Split(dir, ":")[1])))
		}
		println(fmt.Sprintf("dir %d. /%s/", k+1, strings.Split(dir, ":")[0]))
	}
	// showroutes
	http.HandleFunc("/showroutes", showroutes())
	// 证书，https
	if *cert != "" && *key != "" {
		println("===== start https server on: " + *port + " =====")
		if err := http.ListenAndServeTLS(":"+*port, *cert, *key, nil); err != nil {
			println("error: " + err.Error())
		} else {
			return
		}
	}
	// http
	println("\n===== start http server on: " + *port + " =====")
	http.ListenAndServe(":"+*port, nil)
}

func showroutes() http.HandlerFunc {
	s := `<a href="%s/">%s</a><br>`
	return func(w http.ResponseWriter, r *http.Request) {
		tpl := &bytes.Buffer{}
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "text/html")
		for _, dir := range dirs {
			tpl.WriteString(fmt.Sprintf(s, url.PathEscape(strings.Split(dir, ":")[0]), "dir : "+strings.Split(dir, ":")[0]))
		}
		tpl.WriteString("<br>")
		w.Write(tpl.Bytes())
	}
}

func basicAuth(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(*auth, ":") {
			u, p, ok := r.BasicAuth()
			if !ok {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			if u+":"+p != *auth {
				w.WriteHeader(401)
				w.Write([]byte("username or password provided is not correct"))
				return
			}
		}
		f(w, r)
	}
}

// default list
func dirHandler(alias string, fs http.FileSystem) http.HandlerFunc {
	fileServer := http.StripPrefix("/"+alias+"/", http.FileServer(fs))
	return func(w http.ResponseWriter, r *http.Request) {
		fileServer.ServeHTTP(w, r)
	}
}

// lighttpd style
func dirHandlerLighttpd(alias, dir string) http.HandlerFunc {
	root, err := filepath.Abs(dir)
	if err != nil {
		root = dir
	}
	return func(w http.ResponseWriter, r *http.Request) {
		al := "/" + alias + "/"
		pa := strings.Replace(r.URL.Path, al, "", 1)
		values := r.URL.Query()
		fsort := values.Get("sort")
		if fsort == "" {
			fsort = "name"
		}
		xvalue := "?" + values.Encode()
		if xvalue == "?" {
			xvalue = ""
		}
		// order := values.Get("order")
		// if order == "" {
		// 	order = "asc"
		// }
		urlpath := `/ <a href="` + al + xvalue + `">` + alias + `</a> /`
		spa := strings.Split(pa, "/")
		for k, v := range spa {
			if v == "" {
				continue
			}
			urlpath += `<a href="` + al + strings.Join(spa[:k+1], "/") + "/" + xvalue + `">` + v + `</a>/`
		}
		y := strings.ReplaceAll(caddyUp, "__path__", r.URL.Path)
		y = strings.ReplaceAll(y, "__url_path__", urlpath)
		switch fsort {
		case "name":
			y = strings.ReplaceAll(y, "__name_arrow__", uparrow)
			y = strings.ReplaceAll(y, "__size_arrow__", "")
			y = strings.ReplaceAll(y, "__time_arrow__", "")
		case "size":
			y = strings.ReplaceAll(y, "__name_arrow__", "")
			y = strings.ReplaceAll(y, "__size_arrow__", downarrow)
			y = strings.ReplaceAll(y, "__time_arrow__", "")
		case "time":
			y = strings.ReplaceAll(y, "__name_arrow__", "")
			y = strings.ReplaceAll(y, "__size_arrow__", "")
			y = strings.ReplaceAll(y, "__time_arrow__", downarrow)
		}
		tpl := &bytes.Buffer{}
		tpl.WriteString(y)
		p := filepath.Join(root, pa)
		var err error
		var fns []os.FileInfo
		var ssdir = make([]*files, 0)
		var ssfile = make([]*files, 0)
		fn, err := os.Stat(p)
		if err != nil {
			println(err.Error())
			goto RANDER
		}
		if !fn.IsDir() { // 文件类，分类处理
			w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(fn.Name())))
			b, _ := ioutil.ReadFile(p)
			w.Write(b)
			return
		}
		fns, err = ioutil.ReadDir(p)
		if err != nil {
			println(err.Error())
			goto RANDER
		}
		for _, fn := range fns {
			if strings.HasPrefix(fn.Name(), ".") {
				continue
			}
			// url,filename,"/" or "",time,size(- &nbsp;),type(Dir,file)
			if fn.IsDir() {
				ssdir = append(ssdir, &files{
					Name: fn.Name() + "/",
					Type: "folder",
					Size: 0,
					Time: fn.ModTime().Format("2006-01-02 15:04:05"),
				})
				continue
			}
			ssfile = append(ssfile, &files{
				Name: fn.Name(),
				Type: "file",
				Size: fn.Size(),
				Time: fn.ModTime().Format("2006-01-02 15:05:05"),
			})
		}
	RANDER:
		switch fsort {
		case "name":
			sort.Slice(ssdir, func(i, j int) bool {
				return ssdir[i].Name < ssdir[j].Name
			})
			sort.Slice(ssfile, func(i, j int) bool {
				return ssfile[i].Name < ssfile[j].Name
			})
		case "size":
			sort.Slice(ssdir, func(i, j int) bool {
				return ssdir[i].Name < ssdir[j].Name
			})
			sort.Slice(ssfile, func(i, j int) bool {
				return ssfile[i].Size > ssfile[j].Size
			})
		case "time":
			sort.Slice(ssdir, func(i, j int) bool {
				return ssdir[i].Time > ssdir[j].Time
			})
			sort.Slice(ssfile, func(i, j int) bool {
				return ssfile[i].Time > ssfile[j].Time
			})
		}
		for _, v := range ssdir {
			tpl.WriteString(fmt.Sprintf(caddyMid, v.Name+xvalue, v.Type, v.Name, formatFileSize(v.Size), v.Time))
		}
		for _, v := range ssfile {
			tpl.WriteString(fmt.Sprintf(caddyMid, v.Name, v.Type, v.Name, formatFileSize(v.Size), v.Time))
		}
		tpl.WriteString(caddyDown)
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "text/html")
		w.Write(tpl.Bytes())
	}
}
func formatFileSize(fileSize int64) (size string) {
	if fileSize == 0 {
		return "--"
	}
	if fileSize < 1024 {
		return fmt.Sprintf("%d B", fileSize/1)
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%d KB", fileSize/1024)
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%d MB", fileSize/(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%d GB", fileSize/(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%d TB", fileSize/(1024*1024*1024*1024))
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%d EB", fileSize/(1024*1024*1024*1024*1024))
	}
}

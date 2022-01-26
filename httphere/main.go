package main

import (
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

var (
	tplHeader = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Index of %s</title>
<style type="text/css">
a, a:active {text-decoration: none; color: blue;}
a:visited {color: #48468F;}
a:hover, a:focus {text-decoration: underline; color: red;}
body {background-color: #F5F5F5;}
h2 {margin-bottom: 12px;}
table {margin-left: 12px;}
th, td { font: monospace; text-align: left;}
th { font-weight: bold; padding-right: 14px; padding-bottom: 3px;}
td {padding-right: 66px;}
td.s, th.s {text-align: right;}
div.list { background-color: white; border-top: 1px solid #646464; border-bottom: 1px solid #646464; padding-top: 10px; padding-bottom: 14px;}
div.foot { font: monospace; color: #787878; padding-top: 4px;}
</style>
</head>
<body>
<h2>Index of %[1]s</h2>
<div class="list">
<table summary="Directory Listing" cellpadding="0" cellspacing="0">
<thead><tr><th>Name</th><th>Last Modified</th><th>Size</th><th>Type</th></tr></thead>
<tbody>
`
	tplUp = `<tr><td><a href="../">..</a>/</td><td>&nbsp;</td><td class="s">- &nbsp;</td><td>Dir</td></tr>
	`
	tplFoot = `
</tbody>
</table>
</div>
<div class="foot">httphere by X.Yuan</div>
</body>
</html>`
	// url,filename,time,size(- &nbsp;),type(Dir,file)
	shref = `<tr><td><a href="%s">%s</a>%s</td><td>%s</td><td class="s">%s</td><td>%s</td></tr>
`
)

type sliceFlag []string

func (f *sliceFlag) String() string {
	return strings.Join([]string(*f), ",")
}

func (f *sliceFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

var (
	port     = flag.String("http", "8019", "set http port")
	cert     = flag.String("cert", "", "cert file path")
	key      = flag.String("key", "", "key file path")
	auth     = flag.String("auth", "", "set basicauth info, something like username:password")
	lighttpd = flag.Bool("lighttpd", false, "looks like lighttpd page")
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

	println("----- enable routes: -----")
	for _, dir := range dirs {
		if *lighttpd {
			http.HandleFunc("/"+strings.Split(dir, ":")[0]+"/", dirHandlerLighttpd(strings.Split(dir, ":")[0], strings.Split(dir, ":")[1]))
		} else {
			http.HandleFunc("/"+strings.Split(dir, ":")[0]+"/", dirHandler(strings.Split(dir, ":")[0], http.Dir(strings.Split(dir, ":")[1])))
		}
		println("/" + strings.Split(dir, ":")[0] + "/")
	}
	if *cert != "" && *key != "" {
		println("===== start https server on: " + *port + " =====")
		if err := http.ListenAndServeTLS(":"+*port, *cert, *key, nil); err != nil {
			println("error: " + err.Error())
		} else {
			return
		}
	}
	println("\n===== start http server on: " + *port + " =====")
	http.ListenAndServe(":"+*port, nil)
}

func dirHandler(alias string, fs http.FileSystem) http.HandlerFunc {
	fileServer := http.StripPrefix("/"+alias+"/", http.FileServer(fs))
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
		fileServer.ServeHTTP(w, r)
	}
}

func dirHandlerLighttpd(alias, dir string) http.HandlerFunc {
	root, err := filepath.Abs(dir)
	if err != nil {
		root = dir
	}
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
		al := "/" + alias + "/"
		x, _ := url.PathUnescape(r.RequestURI)
		tpl := fmt.Sprintf(tplHeader, x)
		p := strings.Split(filepath.Join(root, strings.Replace(x, al, "", 1)), "?")[0]
		var err error
		var fns []os.FileInfo
		var ssdir = make([][]string, 0)
		var ssfile = make([][]string, 0)
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
				ssdir = append(ssdir, []string{url.PathEscape(fn.Name()) + "/",
					fn.Name(),
					"/",
					fn.ModTime().Format("2006-01-02 15:04:05"),
					"- &nbsp;",
					"Dir"})
				continue
			}
			ssfile = append(ssfile, []string{url.PathEscape(fn.Name()),
				fn.Name(),
				"",
				fn.ModTime().Format("2006-01-02 15:04:05"),
				formatFileSize(fn.Size()),
				strings.Split(mime.TypeByExtension(filepath.Ext(fn.Name())), ";")[0]})
		}
		sort.Slice(ssdir, func(i int, j int) bool {
			return ssdir[i][1] < ssdir[j][1]
		})
		sort.Slice(ssfile, func(i int, j int) bool {
			return ssfile[i][1] < ssfile[j][1]
		})
		if al != r.RequestURI {
			tpl += tplUp
		}
		for _, ss := range ssdir {
			tpl += fmt.Sprintf(shref, ss[0], ss[1], ss[2], ss[3], ss[4], ss[5])
		}
		for _, ss := range ssfile {
			tpl += fmt.Sprintf(shref, ss[0], ss[1], ss[2], ss[3], ss[4], ss[5])
		}
	RANDER:
		tpl += tplFoot
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(tpl))
	}
}
func formatFileSize(fileSize int64) (size string) {
	if fileSize < 1024 {
		//return strconv.FormatInt(fileSize, 10) + "B"
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

package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

//go:embed static
var stat embed.FS

//go:embed tpl/webtv.html
var pageWebTV string

var (
	extreplacer  = strings.NewReplacer(".webp", "", ".dur", "", ".smi", "", ".png", "", ".jpg", "", ".zh-Hans", "", ".zh-Hant", "", ".vtt", "", ".en", "", ".en-US", "")
	namereplacer = strings.NewReplacer("#", "", "%", "")
	subTypes     = []string{".zh-Hant", ".zh-Hans", ".en", ".en-US"}
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
	wtvs sliceFlag
)

func main() {
	// http 静态目录
	flag.Var(&dirs, "dir", "example: -dir=name:path -dir name2:path2")
	// web gallery 目录
	flag.Var(&wtvs, "wtv", "example: -wtv=name:path -wtv name2:path2")
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}
	if len(dirs)+len(wtvs) == 0 {
		dirs = []string{"ls:."}
	}
	http.Handle("/emb/", http.StripPrefix("/emb/", http.FileServer(http.FS(stat))))
	// 静态资源路由
	println("----- enable routes: -----")
	for k, dir := range dirs {
		if *lighttpd {
			http.HandleFunc("/"+strings.Split(dir, ":")[0]+"/", basicAuth(dirHandlerLighttpd(strings.Split(dir, ":")[0], strings.Split(dir, ":")[1])))
		} else {
			http.HandleFunc("/"+strings.Split(dir, ":")[0]+"/", basicAuth(dirHandler(strings.Split(dir, ":")[0], http.Dir(strings.Split(dir, ":")[1]))))
		}
		println("dir " + strconv.Itoa(k+1) + ". /" + strings.Split(dir, ":")[0] + "/")
	}
	// wtv路由
	for k, wtv := range wtvs {
		http.HandleFunc("/"+strings.Split(wtv, ":")[0]+"/", basicAuth(runVideojs(strings.Split(wtv, ":")[0], strings.Split(wtv, ":")[1])))
		http.HandleFunc("/v-"+strings.Split(wtv, ":")[0]+"/", dirHandler("v-"+strings.Split(wtv, ":")[0], http.Dir(strings.Split(wtv, ":")[1])))
		println("wtv " + strconv.Itoa(k+1) + ". /" + strings.Split(wtv, ":")[0] + "/")
	}
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

// smi2Vtt smi转vtt
func smi2Vtt(in, out string) error {
	defer os.Remove(in)
	bIn, err := ioutil.ReadFile(in)
	if err != nil {
		return err
	}
	ss := strings.Split(string(bIn), "\n")
	// 分析语言
	var language string
	for _, v := range ss {
		if strings.HasPrefix(v, "-->") {
			break
		}
		if strings.HasPrefix(v, ".zh") {
			language = v[1:strings.Index(v, " {")]
			if strings.Contains(v, "自动翻译") || language != "zh-Hant" {
				continue
			}
			break
		}
	}
	// 分析主体
	var bOut bytes.Buffer
	// bOut.WriteString("WEBVTT\r\n\r\n")
	var text string
	var tStart, tEnd int64
	for _, v := range ss {
		v = strings.TrimSpace(v)
		idxTime1 := strings.Index(v, "Start=")
		idxClass := strings.Index(v, "class=")
		if idxTime1 == -1 || idxClass == -1 || !strings.Contains(v, "class='"+language+"'") {
			continue
		}
		idxTime2 := strings.Index(v, ">")
		idxText := strings.LastIndex(v, ">")
		if v[idxText+1:] == "" {
			continue
		}
		if v[idxText+1:] == "&nbsp;" {
			tEnd, _ = strconv.ParseInt(v[idxTime1+6:idxTime2], 10, 0)
		} else {
			tStart, _ = strconv.ParseInt(v[idxTime1+6:idxTime2], 10, 0)
			text = v[idxText+1:]
		}
		if tStart >= 0 && tEnd > 0 && len(text) > 0 {
			bOut.WriteString(fmt.Sprintf("%02d:%02d:%02d.%03d --> %02d:%02d:%02d.%03d\r\n",
				tStart/1000/60/60,
				tStart/1000/60%60,
				tStart/1000%60,
				tStart%1000,
				tEnd/1000/60/60,
				tEnd/1000/60%60,
				tEnd/1000%60,
				tEnd%1000))
			bOut.WriteString(text + "\r\n")
			tStart = -1
			tEnd = 0
			text = ""
		}
	}
	if bOut.Len() == 0 {
		return nil
	}
	return ioutil.WriteFile(out, []byte("WEBVTT\r\n\r\n"+bOut.String()), 0644)
}

func isExist(p string) bool {
	if p == "" {
		return false
	}
	_, err := os.Stat(p)
	return err == nil || os.IsExist(err)
}

// wtv
func runVideojs(url, urldst string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()
		srcdir := "/v-" + url + "/"
		subdir := values.Get("sub")
		name := values.Get("name")
		dst := filepath.Join(urldst, subdir)
		flist, err := ioutil.ReadDir(dst)
		if err != nil {
			println(dst, err.Error())
			w.Write([]byte("wrong way"))
			return
		}
		// var playlist, playitem string
		var playlist = make([]*pl, 0)
		var playitem *pl
		var thumblocker sync.WaitGroup
		if values.Get("order") != "name" {
			sort.Slice(flist, func(i int, j int) bool {
				return flist[i].ModTime().After(flist[j].ModTime())
			})
		}
		// _, showdur := c.Params.Get("dur")
		var fileext, filesrc, filethumb, fileSmi, fileVtt, filename, filebase string
		if subdir != "" {
			srcdir += subdir + "/"
		}
		for _, f := range flist {
			if f.IsDir() || !strings.Contains(f.Name(), name) || strings.HasPrefix(f.Name(), ".fuse") {
				continue
			}
			filename = namereplacer.Replace(f.Name())
			if f.Name() != filename {
				os.Rename(filepath.Join(dst, f.Name()), filepath.Join(dst, filename))
			}
			fileext = strings.ToLower(filepath.Ext(filename))
			filebase = strings.TrimSuffix(filename, fileext)
			if strings.HasSuffix(filebase, "f133") ||
				strings.HasSuffix(filebase, "f242") ||
				strings.HasSuffix(filebase, "f251") ||
				strings.HasSuffix(filebase, "f250") ||
				strings.HasSuffix(filebase, "f140") {
				continue
			}
			playitem = &pl{Sources: make([]*psrc, 1), Thumbnail: make([]*psrc, 1), TextTracks: make([]*psrc, 0)}
			switch fileext {
			case ".wav", ".m4a": // 音频
				playitem.Name = filename
				playitem.Sources[0] = &psrc{Src: srcdir + filename, Type: mime.TypeByExtension(fileext)}
				playlist = append(playlist, playitem)

				// playitem, _ = sjson.Set("", "name", filename)
				// playitem, _ = sjson.Set(playitem, "sources.0.src", srcdir+filename)
				// playitem, _ = sjson.Set(playitem, "sources.0.type", "audio/"+fileext[1:])
				// playlist, _ = sjson.Set(playlist, "pl.-1", gjson.Parse(playitem).Value())
			case ".mp4", ".mkv", ".webm": // 视频
				filesrc = filepath.Join(dst, filename)
				filethumb = filepath.Join(dst, "."+filename+".webp")
				fileSmi = filepath.Join(dst, filebase+".smi")
				fileVtt = filepath.Join(dst, "."+filename+".vtt")

				playitem.Name = filename
				playitem.DateTime = f.ModTime().Format("01月02日 15:04")
				playitem.Sources[0] = &psrc{Src: srcdir + filename, Type: mime.TypeByExtension(fileext)}

				// playitem, _ = sjson.Set("", "name", filename)
				// playitem, _ = sjson.Set(playitem, "datetime", f.ModTime().Format("01月02日 15:04"))
				// playitem, _ = sjson.Set(playitem, "sources.0.src", srcdir+filename)
				// switch fileext {
				// case ".mkv":
				// 	playitem, _ = sjson.Set(playitem, "sources.0.type", "video/webm")
				// default:
				// 	playitem, _ = sjson.Set(playitem, "sources.0.type", "video/"+fileext[1:])
				// }
				// 缩略图
				if runtime.GOARCH == "amd64" {
					if !isExist(filethumb) || isExist(strings.TrimSuffix(filesrc, fileext)+".webp") {
						thumblocker.Add(1)
						go func(in, out, ext string) {
							defer thumblocker.Done()
							if isExist(strings.TrimSuffix(in, ext) + ".jpg") {
								cmd := exec.Command("mogrify", "-resize", "168x", "-quality", "80%", "-write", out, strings.Trim(in, ext)+".jpg")
								cmd.Run()
								os.Remove(strings.TrimSuffix(in, ext) + ".jpg")
							} else if isExist(strings.TrimSuffix(in, ext) + ".webp") {
								cmd := exec.Command("mogrify", "-resize", "168x", "-quality", "80%", "-write", out, strings.Trim(in, ext)+".webp")
								cmd.Run()
								os.Remove(strings.TrimSuffix(in, ext) + ".webp")
							} else {
								cmd := exec.Command("ffmpeg", "-i", in, "-ss", "00:00:03", "-s", "168:94", "-vframes", "1", out)
								cmd.Run()
							}
						}(filesrc, filethumb, fileext)
					}
					playitem.Thumbnail[0] = &psrc{Src: srcdir + "." + filename + ".webp"}
					// playitem, _ = sjson.Set(playitem, "thumbnail.0.src", srcdir+"."+filename+".webp")
				}
				// 字幕
				var idx = 0
				for _, v := range subTypes {
					subraw := filepath.Join(dst, filebase+v+".vtt")
					subdst := filepath.Join(dst, "."+filename+v+".vtt")
					subsrc := srcdir + "." + filename + v + ".vtt"
					if isExist(subraw) {
						os.Rename(subraw, subdst)
					}
					if isExist(subdst) {
						playitem.TextTracks = append(playitem.TextTracks, &psrc{
							Src:     subsrc,
							Label:   v[1:],
							SrcLang: v[1:],
							Default: func(idx int) string {
								if idx == 0 {
									return "true"
								}
								return ""
							}(idx),
						})
						// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.src", idx), subsrc)
						// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.label", idx), v[1:])
						// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.srclang", idx), v[1:])
						// if idx == 0 {
						// 	playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.default", idx), "true")
						// }
						idx++
					}
				}
				if idx == 0 {
					if isExist(fileSmi) && !isExist(fileVtt) {
						smi2Vtt(fileSmi, fileVtt)
					}
					if isExist(fileVtt) {
						playitem.TextTracks = append(playitem.TextTracks, &psrc{
							Src:     srcdir + "." + filename + ".vtt",
							Label:   "中文",
							SrcLang: "zh",
							Default: "false",
						})
						// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.src", idx), srcdir+"."+filename+".vtt")
						// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.label", idx), "中文")
						// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.srclang", idx), "zh")
						// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.default", idx), "false")
					}
				}
				// playlist, _ = sjson.Set(playlist, "pl.-1", gjson.Parse(playitem).Value())
				playlist = append(playlist, playitem)
			case ".dur", ".png", ".jpg", ".vtt", ".webp":
				if strings.HasPrefix(filename, ".") {
					if !isExist(filepath.Join(dst, extreplacer.Replace(filename[1:]))) {
						os.Remove(filepath.Join(dst, filename))
					}
				}
			case ".jpg~":
				os.Remove(filepath.Join(dst, filename))
			}
		}
		// tpl := strings.Replace(pageWebTV, "playlist_data_here", gjson.Parse(playlist).Get("pl").String(), 1)
		b, _ := json.Marshal(playlist)
		tpl := strings.Replace(pageWebTV, "playlist_data_here", string(b), 1)
		// println(gjson.Parse(playlist).Get("pl").String())

		thumblocker.Wait()
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(tpl))
	}
}

type pl struct {
	DateTime   string  `json:"datetime,omitempty"`
	Name       string  `json:"name,omitempty"`
	Sources    []*psrc `json:"sources,omitempty"`
	Thumbnail  []*psrc `json:"thumbnail,omitempty"`
	TextTracks []*psrc `json:"textTracks,omitempty"`
}

type psrc struct {
	Src     string `json:"src,omitempty"`
	Type    string `json:"type,omitempty"`
	Label   string `json:"label,omitempty"`
	SrcLang string `json:"srclang,omitempty"`
	Default string `json:"default,omitempty"`
}

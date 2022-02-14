package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"math"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/image/webp"
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
<title>Index of __path__</title>
<style type="text/css">
a, a:active {text-decoration: none; color: blue;}
a:visited {color: #48468F;}
a:hover, a:focus {text-decoration: underline; color: red;}
body {background-color: #F5F5F5;}
h2 {margin-bottom: 12px;}
table {margin-left: 12px;}
th, td { font: 90% monospace; text-align: left;}
th { font-weight: bold; padding-right: 14px; padding-bottom: 3px;}
td {padding-right: 66px;}
td.s, th.s {text-align: right;}
div.list { background-color: white; border-top: 1px solid #646464; border-bottom: 1px solid #646464; padding-top: 10px; padding-bottom: 14px;}
div.foot { font: 90% monospace; color: #787878; padding-top: 4px;}
</style>
</head>
<body>
<h2>Index of __path__</h2>
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
		println(fmt.Sprintf("dir %d. /%s/", k+1, strings.Split(dir, ":")[0]))
	}
	// wtv路由
	for k, wtv := range wtvs {
		http.HandleFunc("/"+strings.Split(wtv, ":")[0]+"/", basicAuth(runVideojs(strings.Split(wtv, ":")[0], strings.Split(wtv, ":")[1])))
		http.HandleFunc("/v-"+strings.Split(wtv, ":")[0]+"/", dirHandler("v-"+strings.Split(wtv, ":")[0], http.Dir(strings.Split(wtv, ":")[1])))
		println(fmt.Sprintf("wtv %d. /%s/", k+1, strings.Split(wtv, ":")[0]))
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
		for _, dir := range wtvs {
			tpl.WriteString(fmt.Sprintf(s, url.PathEscape(strings.Split(dir, ":")[0]), "wtv : "+strings.Split(dir, ":")[0]))
		}
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
		x, _ := url.PathUnescape(r.RequestURI)
		tpl := strings.ReplaceAll(tplHeader, "__path__", x)
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
// func smi2Vtt(in, out string) error {
// 	defer os.Remove(in)
// 	bIn, err := ioutil.ReadFile(in)
// 	if err != nil {
// 		return err
// 	}
// 	ss := strings.Split(string(bIn), "\n")
// 	// 分析语言
// 	var language string
// 	for _, v := range ss {
// 		if strings.HasPrefix(v, "-->") {
// 			break
// 		}
// 		if strings.HasPrefix(v, ".zh") {
// 			language = v[1:strings.Index(v, " {")]
// 			if strings.Contains(v, "自动翻译") || language != "zh-Hant" {
// 				continue
// 			}
// 			break
// 		}
// 	}
// 	// 分析主体
// 	var bOut = &strings.Builder{}
// 	// bOut.WriteString("WEBVTT\r\n\r\n")
// 	var text string
// 	var tStart, tEnd int64
// 	for _, v := range ss {
// 		v = strings.TrimSpace(v)
// 		idxTime1 := strings.Index(v, "Start=")
// 		idxClass := strings.Index(v, "class=")
// 		if idxTime1 == -1 || idxClass == -1 || !strings.Contains(v, "class='"+language+"'") {
// 			continue
// 		}
// 		idxTime2 := strings.Index(v, ">")
// 		idxText := strings.LastIndex(v, ">")
// 		if v[idxText+1:] == "" {
// 			continue
// 		}
// 		if v[idxText+1:] == "&nbsp;" {
// 			tEnd, _ = strconv.ParseInt(v[idxTime1+6:idxTime2], 10, 0)
// 		} else {
// 			tStart, _ = strconv.ParseInt(v[idxTime1+6:idxTime2], 10, 0)
// 			text = v[idxText+1:]
// 		}
// 		if tStart >= 0 && tEnd > 0 && len(text) > 0 {
// 			bOut.WriteString(fmt.Sprintf("%02d:%02d:%02d.%03d --> %02d:%02d:%02d.%03d\r\n",
// 				tStart/1000/60/60,
// 				tStart/1000/60%60,
// 				tStart/1000%60,
// 				tStart%1000,
// 				tEnd/1000/60/60,
// 				tEnd/1000/60%60,
// 				tEnd/1000%60,
// 				tEnd%1000))
// 			bOut.WriteString(text + "\r\n")
// 			tStart = -1
// 			tEnd = 0
// 			text = ""
// 		}
// 	}
// 	if bOut.Len() == 0 {
// 		return nil
// 	}
// 	return ioutil.WriteFile(out, []byte("WEBVTT\r\n\r\n"+bOut.String()), 0644)
// }

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
		if values.Get("order") != "name" {
			sort.Slice(flist, func(i int, j int) bool {
				return flist[i].ModTime().After(flist[j].ModTime())
			})
		}
		// _, showdur := c.Params.Get("dur")
		var fileext, filesrc, filethumb, filename, filebase string
		// var fileSmi, fileVtt string
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
				// fileSmi = filepath.Join(dst, filebase+".smi")
				// fileVtt = filepath.Join(dst, "."+filename+".vtt")

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
				if isExist(strings.TrimSuffix(filesrc, fileext) + ".webp") {
					resize(strings.TrimSuffix(filesrc, fileext)+".webp", filethumb)
				}
				// if runtime.GOARCH == "amd64" {
				// 	if !isExist(filethumb) || isExist(strings.TrimSuffix(filesrc, fileext)+".webp") {
				// 		thumblocker.Add(1)
				// 		go func(in, out, ext string) {
				// 			defer thumblocker.Done()
				// 			if isExist(strings.TrimSuffix(in, ext) + ".jpg") {
				// 				cmd := exec.Command("mogrify", "-resize", "168x", "-quality", "80%", "-write", out, strings.Trim(in, ext)+".jpg")
				// 				cmd.Run()
				// 				os.Remove(strings.TrimSuffix(in, ext) + ".jpg")
				// 			} else if isExist(strings.TrimSuffix(in, ext) + ".webp") {
				// 				cmd := exec.Command("mogrify", "-resize", "168x", "-quality", "80%", "-write", out, strings.Trim(in, ext)+".webp")
				// 				cmd.Run()
				// 				os.Remove(strings.TrimSuffix(in, ext) + ".webp")
				// 			} else {
				// 				cmd := exec.Command("ffmpeg", "-i", in, "-ss", "00:00:03", "-s", "168:94", "-vframes", "1", out)
				// 				cmd.Run()
				// 			}
				// 		}(filesrc, filethumb, fileext)
				// 	}
				playitem.Thumbnail[0] = &psrc{Src: srcdir + "." + filename + ".webp"}
				// 	// playitem, _ = sjson.Set(playitem, "thumbnail.0.src", srcdir+"."+filename+".webp")
				// }
				// 字幕
				// var idx = 0
				// for _, v := range subTypes {
				// 	subraw := filepath.Join(dst, filebase+v+".vtt")
				// 	subdst := filepath.Join(dst, "."+filename+v+".vtt")
				// 	subsrc := srcdir + "." + filename + v + ".vtt"
				// 	if isExist(subraw) {
				// 		os.Rename(subraw, subdst)
				// 	}
				// 	if isExist(subdst) {
				// 		playitem.TextTracks = append(playitem.TextTracks, &psrc{
				// 			Src:     subsrc,
				// 			Label:   v[1:],
				// 			SrcLang: v[1:],
				// 			Default: func(idx int) string {
				// 				if idx == 0 {
				// 					return "true"
				// 				}
				// 				return ""
				// 			}(idx),
				// 		})
				// 		// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.src", idx), subsrc)
				// 		// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.label", idx), v[1:])
				// 		// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.srclang", idx), v[1:])
				// 		// if idx == 0 {
				// 		// 	playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.default", idx), "true")
				// 		// }
				// 		idx++
				// 	}
				// }
				// if idx == 0 {
				// 	if isExist(fileSmi) && !isExist(fileVtt) {
				// 		smi2Vtt(fileSmi, fileVtt)
				// 	}
				// 	if isExist(fileVtt) {
				// 		playitem.TextTracks = append(playitem.TextTracks, &psrc{
				// 			Src:     srcdir + "." + filename + ".vtt",
				// 			Label:   "中文",
				// 			SrcLang: "zh",
				// 			Default: "false",
				// 		})
				// 		// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.src", idx), srcdir+"."+filename+".vtt")
				// 		// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.label", idx), "中文")
				// 		// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.srclang", idx), "zh")
				// 		// playitem, _ = sjson.Set(playitem, fmt.Sprintf("textTracks.%d.default", idx), "false")
				// 	}
				// }
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

func resize(filein, fileout string) {
	if filepath.Ext(filein) != ".webp" {
		return
	}
	fin, err := os.Open(filein)
	if err != nil {
		println(err.Error())
		return
	}
	defer fin.Close()
	// decode jpeg into image.Image
	img, err := webp.Decode(fin)
	if err != nil {
		println(err.Error())
		return
	}
	width := uint(169)
	height := uint(96)
	scaleX, scaleY := calcFactors(width, height, float64(img.Bounds().Dx()), float64(img.Bounds().Dy()))
	// Trivial case: return input image
	if int(width) == img.Bounds().Dx() && int(height) == img.Bounds().Dy() {
		return
	}

	// Input image has no pixels
	if img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		return
	}

	temp := newYCC(image.Rect(0, 0, img.Bounds().Dy(), int(width)), img.(*image.YCbCr).SubsampleRatio)
	result := newYCC(image.Rect(0, 0, int(width), int(height)), image.YCbCrSubsampleRatio444)

	coeffs, offset, filterLength := createWeights8(temp.Bounds().Dy(), 2, 1.0, scaleX, linear)
	in := imageYCbCrToYCC(img.(*image.YCbCr))
	resizeYCbCr(in, temp, scaleX, coeffs, offset, filterLength)

	coeffs, offset, filterLength = createWeights8(result.Bounds().Dy(), 2, 1.0, scaleY, linear)
	resizeYCbCr(temp, result, scaleY, coeffs, offset, filterLength)

	fout, err := os.Create(fileout)
	if err != nil {
		println(err.Error())
		return
	}
	defer fout.Close()
	// write new image to file
	jpeg.Encode(fout, result.YCbCr(), nil)
}

func resizeYCbCr(in *ycc, out *ycc, scale float64, coeffs []int16, offset []int, filterLength int) {
	newBounds := out.Bounds()
	maxX := in.Bounds().Dx() - 1

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		row := in.Pix[x*in.Stride:]
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			var p [3]int32
			var sum int32
			start := offset[y]
			ci := y * filterLength
			for i := 0; i < filterLength; i++ {
				coeff := coeffs[ci+i]
				if coeff != 0 {
					xi := start + i
					switch {
					case uint(xi) < uint(maxX):
						xi *= 3
					case xi >= maxX:
						xi = 3 * maxX
					default:
						xi = 0
					}
					p[0] += int32(coeff) * int32(row[xi+0])
					p[1] += int32(coeff) * int32(row[xi+1])
					p[2] += int32(coeff) * int32(row[xi+2])
					sum += int32(coeff)
				}
			}

			xo := (y-newBounds.Min.Y)*out.Stride + (x-newBounds.Min.X)*3
			out.Pix[xo+0] = clampUint8(p[0] / sum)
			out.Pix[xo+1] = clampUint8(p[1] / sum)
			out.Pix[xo+2] = clampUint8(p[2] / sum)
		}
	}
}

func linear(in float64) float64 {
	in = math.Abs(in)
	if in <= 1 {
		return 1 - in
	}
	return 0
}

// Calculates scaling factors using old and new image dimensions.
func calcFactors(width, height uint, oldWidth, oldHeight float64) (scaleX, scaleY float64) {
	if width == 0 {
		if height == 0 {
			scaleX = 1.0
			scaleY = 1.0
		} else {
			scaleY = oldHeight / float64(height)
			scaleX = scaleY
		}
	} else {
		scaleX = oldWidth / float64(width)
		if height == 0 {
			scaleY = scaleX
		} else {
			scaleY = oldHeight / float64(height)
		}
	}
	return
}

// range [-256,256]
func createWeights8(dy, filterLength int, blur, scale float64, kernel func(float64) float64) ([]int16, []int, int) {
	filterLength = filterLength * int(math.Max(math.Ceil(blur*scale), 1))
	filterFactor := math.Min(1./(blur*scale), 1)

	coeffs := make([]int16, dy*filterLength)
	start := make([]int, dy)
	for y := 0; y < dy; y++ {
		interpX := scale*(float64(y)+0.5) - 0.5
		start[y] = int(interpX) - filterLength/2 + 1
		interpX -= float64(start[y])
		for i := 0; i < filterLength; i++ {
			in := (interpX - float64(i)) * filterFactor
			coeffs[y*filterLength+i] = int16(kernel(in) * 256)
		}
	}

	return coeffs, start, filterLength
}
func clampUint8(in int32) uint8 {
	// casting a negative int to an uint will result in an overflown
	// large uint. this behavior will be exploited here and in other functions
	// to achieve a higher performance.
	if uint32(in) < 256 {
		return uint8(in)
	}
	if in > 255 {
		return 255
	}
	return 0
}

/*
Copyright (c) 2014, Charlie Vieth <charlie.vieth@gmail.com>

Permission to use, copy, modify, and/or distribute this software for any purpose
with or without fee is hereby granted, provided that the above copyright notice
and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF
THIS SOFTWARE.
*/

// ycc is an in memory YCbCr image.  The Y, Cb and Cr samples are held in a
// single slice to increase resizing performance.
type ycc struct {
	// Pix holds the image's pixels, in Y, Cb, Cr order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*3].
	Pix []uint8
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
	// SubsampleRatio is the subsample ratio of the original YCbCr image.
	SubsampleRatio image.YCbCrSubsampleRatio
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *ycc) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*3
}

func (p *ycc) Bounds() image.Rectangle {
	return p.Rect
}

func (p *ycc) ColorModel() color.Model {
	return color.YCbCrModel
}

func (p *ycc) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(p.Rect)) {
		return color.YCbCr{}
	}
	i := p.PixOffset(x, y)
	return color.YCbCr{
		p.Pix[i+0],
		p.Pix[i+1],
		p.Pix[i+2],
	}
}

func (p *ycc) Opaque() bool {
	return true
}

// SubImage returns an image representing the portion of the image p visible
// through r. The returned value shares pixels with the original image.
func (p *ycc) SubImage(r image.Rectangle) image.Image {
	r = r.Intersect(p.Rect)
	if r.Empty() {
		return &ycc{SubsampleRatio: p.SubsampleRatio}
	}
	i := p.PixOffset(r.Min.X, r.Min.Y)
	return &ycc{
		Pix:            p.Pix[i:],
		Stride:         p.Stride,
		Rect:           r,
		SubsampleRatio: p.SubsampleRatio,
	}
}

// newYCC returns a new ycc with the given bounds and subsample ratio.
func newYCC(r image.Rectangle, s image.YCbCrSubsampleRatio) *ycc {
	w, h := r.Dx(), r.Dy()
	buf := make([]uint8, 3*w*h)
	return &ycc{Pix: buf, Stride: 3 * w, Rect: r, SubsampleRatio: s}
}

// Copy of image.YCbCrSubsampleRatio constants - this allows us to support
// older versions of Go where these constants are not defined (i.e. Go 1.4)
const (
	ycbcrSubsampleRatio444 image.YCbCrSubsampleRatio = iota
	ycbcrSubsampleRatio422
	ycbcrSubsampleRatio420
	ycbcrSubsampleRatio440
	ycbcrSubsampleRatio411
	ycbcrSubsampleRatio410
)

// YCbCr converts ycc to a YCbCr image with the same subsample ratio
// as the YCbCr image that ycc was generated from.
func (p *ycc) YCbCr() *image.YCbCr {
	ycbcr := image.NewYCbCr(p.Rect, p.SubsampleRatio)
	switch ycbcr.SubsampleRatio {
	case ycbcrSubsampleRatio422:
		return p.ycbcr422(ycbcr)
	case ycbcrSubsampleRatio420:
		return p.ycbcr420(ycbcr)
	case ycbcrSubsampleRatio440:
		return p.ycbcr440(ycbcr)
	case ycbcrSubsampleRatio444:
		return p.ycbcr444(ycbcr)
	case ycbcrSubsampleRatio411:
		return p.ycbcr411(ycbcr)
	case ycbcrSubsampleRatio410:
		return p.ycbcr410(ycbcr)
	}
	return ycbcr
}

// imageYCbCrToYCC converts a YCbCr image to a ycc image for resizing.
func imageYCbCrToYCC(in *image.YCbCr) *ycc {
	w, h := in.Rect.Dx(), in.Rect.Dy()
	p := ycc{
		Pix:            make([]uint8, 3*w*h),
		Stride:         3 * w,
		Rect:           image.Rect(0, 0, w, h),
		SubsampleRatio: in.SubsampleRatio,
	}
	switch in.SubsampleRatio {
	case ycbcrSubsampleRatio422:
		return convertToYCC422(in, &p)
	case ycbcrSubsampleRatio420:
		return convertToYCC420(in, &p)
	case ycbcrSubsampleRatio440:
		return convertToYCC440(in, &p)
	case ycbcrSubsampleRatio444:
		return convertToYCC444(in, &p)
	case ycbcrSubsampleRatio411:
		return convertToYCC411(in, &p)
	case ycbcrSubsampleRatio410:
		return convertToYCC410(in, &p)
	}
	return &p
}

func (p *ycc) ycbcr422(ycbcr *image.YCbCr) *image.YCbCr {
	var off int
	Pix := p.Pix
	Y := ycbcr.Y
	Cb := ycbcr.Cb
	Cr := ycbcr.Cr
	for y := 0; y < ycbcr.Rect.Max.Y-ycbcr.Rect.Min.Y; y++ {
		yy := y * ycbcr.YStride
		cy := y * ycbcr.CStride
		for x := 0; x < ycbcr.Rect.Max.X-ycbcr.Rect.Min.X; x++ {
			ci := cy + x/2
			Y[yy+x] = Pix[off+0]
			Cb[ci] = Pix[off+1]
			Cr[ci] = Pix[off+2]
			off += 3
		}
	}
	return ycbcr
}

func (p *ycc) ycbcr420(ycbcr *image.YCbCr) *image.YCbCr {
	var off int
	Pix := p.Pix
	Y := ycbcr.Y
	Cb := ycbcr.Cb
	Cr := ycbcr.Cr
	for y := 0; y < ycbcr.Rect.Max.Y-ycbcr.Rect.Min.Y; y++ {
		yy := y * ycbcr.YStride
		cy := (y / 2) * ycbcr.CStride
		for x := 0; x < ycbcr.Rect.Max.X-ycbcr.Rect.Min.X; x++ {
			ci := cy + x/2
			Y[yy+x] = Pix[off+0]
			Cb[ci] = Pix[off+1]
			Cr[ci] = Pix[off+2]
			off += 3
		}
	}
	return ycbcr
}

func (p *ycc) ycbcr440(ycbcr *image.YCbCr) *image.YCbCr {
	var off int
	Pix := p.Pix
	Y := ycbcr.Y
	Cb := ycbcr.Cb
	Cr := ycbcr.Cr
	for y := 0; y < ycbcr.Rect.Max.Y-ycbcr.Rect.Min.Y; y++ {
		yy := y * ycbcr.YStride
		cy := (y / 2) * ycbcr.CStride
		for x := 0; x < ycbcr.Rect.Max.X-ycbcr.Rect.Min.X; x++ {
			ci := cy + x
			Y[yy+x] = Pix[off+0]
			Cb[ci] = Pix[off+1]
			Cr[ci] = Pix[off+2]
			off += 3
		}
	}
	return ycbcr
}

func (p *ycc) ycbcr444(ycbcr *image.YCbCr) *image.YCbCr {
	var off int
	Pix := p.Pix
	Y := ycbcr.Y
	Cb := ycbcr.Cb
	Cr := ycbcr.Cr
	for y := 0; y < ycbcr.Rect.Max.Y-ycbcr.Rect.Min.Y; y++ {
		yy := y * ycbcr.YStride
		cy := y * ycbcr.CStride
		for x := 0; x < ycbcr.Rect.Max.X-ycbcr.Rect.Min.X; x++ {
			ci := cy + x
			Y[yy+x] = Pix[off+0]
			Cb[ci] = Pix[off+1]
			Cr[ci] = Pix[off+2]
			off += 3
		}
	}
	return ycbcr
}

func (p *ycc) ycbcr411(ycbcr *image.YCbCr) *image.YCbCr {
	var off int
	Pix := p.Pix
	Y := ycbcr.Y
	Cb := ycbcr.Cb
	Cr := ycbcr.Cr
	for y := 0; y < ycbcr.Rect.Max.Y-ycbcr.Rect.Min.Y; y++ {
		yy := y * ycbcr.YStride
		cy := y * ycbcr.CStride
		for x := 0; x < ycbcr.Rect.Max.X-ycbcr.Rect.Min.X; x++ {
			ci := cy + x/4
			Y[yy+x] = Pix[off+0]
			Cb[ci] = Pix[off+1]
			Cr[ci] = Pix[off+2]
			off += 3
		}
	}
	return ycbcr
}

func (p *ycc) ycbcr410(ycbcr *image.YCbCr) *image.YCbCr {
	var off int
	Pix := p.Pix
	Y := ycbcr.Y
	Cb := ycbcr.Cb
	Cr := ycbcr.Cr
	for y := 0; y < ycbcr.Rect.Max.Y-ycbcr.Rect.Min.Y; y++ {
		yy := y * ycbcr.YStride
		cy := (y / 2) * ycbcr.CStride
		for x := 0; x < ycbcr.Rect.Max.X-ycbcr.Rect.Min.X; x++ {
			ci := cy + x/4
			Y[yy+x] = Pix[off+0]
			Cb[ci] = Pix[off+1]
			Cr[ci] = Pix[off+2]
			off += 3
		}
	}
	return ycbcr
}

func convertToYCC422(in *image.YCbCr, p *ycc) *ycc {
	var off int
	Pix := p.Pix
	Y := in.Y
	Cb := in.Cb
	Cr := in.Cr
	for y := 0; y < in.Rect.Max.Y-in.Rect.Min.Y; y++ {
		yy := y * in.YStride
		cy := y * in.CStride
		for x := 0; x < in.Rect.Max.X-in.Rect.Min.X; x++ {
			ci := cy + x/2
			Pix[off+0] = Y[yy+x]
			Pix[off+1] = Cb[ci]
			Pix[off+2] = Cr[ci]
			off += 3
		}
	}
	return p
}

func convertToYCC420(in *image.YCbCr, p *ycc) *ycc {
	var off int
	Pix := p.Pix
	Y := in.Y
	Cb := in.Cb
	Cr := in.Cr
	for y := 0; y < in.Rect.Max.Y-in.Rect.Min.Y; y++ {
		yy := y * in.YStride
		cy := (y / 2) * in.CStride
		for x := 0; x < in.Rect.Max.X-in.Rect.Min.X; x++ {
			ci := cy + x/2
			Pix[off+0] = Y[yy+x]
			Pix[off+1] = Cb[ci]
			Pix[off+2] = Cr[ci]
			off += 3
		}
	}
	return p
}

func convertToYCC440(in *image.YCbCr, p *ycc) *ycc {
	var off int
	Pix := p.Pix
	Y := in.Y
	Cb := in.Cb
	Cr := in.Cr
	for y := 0; y < in.Rect.Max.Y-in.Rect.Min.Y; y++ {
		yy := y * in.YStride
		cy := (y / 2) * in.CStride
		for x := 0; x < in.Rect.Max.X-in.Rect.Min.X; x++ {
			ci := cy + x
			Pix[off+0] = Y[yy+x]
			Pix[off+1] = Cb[ci]
			Pix[off+2] = Cr[ci]
			off += 3
		}
	}
	return p
}

func convertToYCC444(in *image.YCbCr, p *ycc) *ycc {
	var off int
	Pix := p.Pix
	Y := in.Y
	Cb := in.Cb
	Cr := in.Cr
	for y := 0; y < in.Rect.Max.Y-in.Rect.Min.Y; y++ {
		yy := y * in.YStride
		cy := y * in.CStride
		for x := 0; x < in.Rect.Max.X-in.Rect.Min.X; x++ {
			ci := cy + x
			Pix[off+0] = Y[yy+x]
			Pix[off+1] = Cb[ci]
			Pix[off+2] = Cr[ci]
			off += 3
		}
	}
	return p
}

func convertToYCC411(in *image.YCbCr, p *ycc) *ycc {
	var off int
	Pix := p.Pix
	Y := in.Y
	Cb := in.Cb
	Cr := in.Cr
	for y := 0; y < in.Rect.Max.Y-in.Rect.Min.Y; y++ {
		yy := y * in.YStride
		cy := y * in.CStride
		for x := 0; x < in.Rect.Max.X-in.Rect.Min.X; x++ {
			ci := cy + x/4
			Pix[off+0] = Y[yy+x]
			Pix[off+1] = Cb[ci]
			Pix[off+2] = Cr[ci]
			off += 3
		}
	}
	return p
}

func convertToYCC410(in *image.YCbCr, p *ycc) *ycc {
	var off int
	Pix := p.Pix
	Y := in.Y
	Cb := in.Cb
	Cr := in.Cr
	for y := 0; y < in.Rect.Max.Y-in.Rect.Min.Y; y++ {
		yy := y * in.YStride
		cy := (y / 2) * in.CStride
		for x := 0; x < in.Rect.Max.X-in.Rect.Min.X; x++ {
			ci := cy + x/4
			Pix[off+0] = Y[yy+x]
			Pix[off+1] = Cb[ci]
			Pix[off+2] = Cr[ci]
			off += 3
		}
	}
	return p
}

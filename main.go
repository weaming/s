package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	fp "path/filepath"
	"strconv"
	"strings"
)

const DEFAULT_PW = "admin"

var size int
var NEED_AUTH bool
var ADMIN, PASSWORD string

func main() {
	var LISTEN = flag.String("l", ":8000", "Listen [host]:port, default bind to 0.0.0.0")

	flag.BoolVar(&NEED_AUTH, "a", false, "Whether need authorization.")
	flag.StringVar(&ADMIN, "u", "admin", "Basic authorization username")
	flag.StringVar(&PASSWORD, "p", DEFAULT_PW, "Basic authorization password")
	flag.IntVar(&size, "n", 20, "The maximum number of files in each page.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] ROOT\nThe ROOT is the directory to be serve.\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// check the directory path
	ROOT, _ := fp.Abs(flag.Arg(0))
	fi, err := os.Stat(ROOT)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !fi.IsDir() {
		fmt.Fprintln(os.Stderr, "The path should be a directory!!")
		os.Exit(1)
	}

	green(fmt.Sprintf("To be listed direcotry: [%v]", ROOT))

	// basic authentication
	if NEED_AUTH {
		if PASSWORD == DEFAULT_PW {
			red("Warning: set yourself password by option -p")
		}
		green(fmt.Sprintf("Your basic auth name and password: [%v:%v]", ADMIN, PASSWORD))
	} else {
		red("Warning: please set your HTTP basic authentication")
	}

	Redirect("/", "/index")
	ServeFile("/favicon.ico", fp.Join(ROOT, "./favicon.ico"))

	http.Handle("/index/", MyAlbum{root: ROOT})
	ServeDir("/s/", ROOT)

	fmt.Printf("Open http://127.0.0.1:%v to enjoy!\n", strings.Split(*LISTEN, ":")[1])
	for _, ip := range GetIntranetIP() {
		fmt.Printf("Your intranet IP: %v ==> http://%v:%v\n", ip, ip, strings.Split(*LISTEN, ":")[1])
	}
	log.Fatal(http.ListenAndServe(*LISTEN, nil))
}

type MyAlbum struct {
	root string
	dir  *Dir
}

func (album MyAlbum) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logit(r)
	if NEED_AUTH {
		mybasicAuth(album.handlerFunc, ADMIN, PASSWORD)(w, r)
	} else {
		album.handlerFunc(w, r)
	}
}

func (album MyAlbum) handlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "Happy coding!")
	pathName := r.URL.Path

	page, err := getQueryInt(r, "page")
	if err != nil {
		//target, _ := AddQuery(pathName, "page", "1")
		//http.Redirect(w, r, target, http.StatusFound)
		//return
		page = 1
	}

	obj := NewDir(fp.Join(album.root, pathName[6:]))
	if obj == nil {
		w.Write([]byte("Invalid URL"))
		return
	} else {
		album.dir = obj
	}

	pagination, htmlImages, returnPage := Img2Html(pathName, album.dir, page)
	if returnPage != page {
		fmt.Println(returnPage)
		target, _ := AddQuery(pathName, "page", strconv.Itoa(returnPage))
		http.Redirect(w, r, target, http.StatusFound)
	}

	w.Write([]byte(fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
	<meta charset="UTF-8">
	<title>Static Files Index</title>
	<style>
	.right{float: right;}
	.card{
		background-color: #fff;
		box-shadow: 0 2px 5px 0 rgba(0, 0, 0, .16), 0 2px 10px 0 rgba(0, 0, 0, .12);
		margin: 0 auto 1rem auto;
		padding: 1rem;
		max-width: 900px;
		border-radius: 3px;
	}
	a:link{color: #1a0dab;}
	a:visited{color: #609;}
	.directory:hover {
		background-color: #eee;
	}

	div.pagination {
		min-height: 20px;
	}

	div.pagination a{
		display: inline-block;
		border: 1px solid #aaa;
		padding: 5px 10px;
		margin: 5px 10px;
		border-radius: 4px;
		color: black;
		text-decoration: none;
	}
	div.pagination a:hover{
		box-shadow: 0 2px 5px 0 rgba(0, 0, 0, .16), 0 2px 10px 0 rgba(0, 0, 0, .12);
	}

	p.file{
		margin: 5px;
		width: 100%%;
	}
	p.file:hover{
		background-color: #34A853;
	}

	</style>
	</head>
	<body>
	<div class="card directories">
	<h3> Directories: %v <a href="/index" class="right">Home</a> </h3>
	<div>%v</div>
	</div>
	<div class="card files">
	<h3>Files: %v Size: %v</h3>
	<div class="pagiContainer">%v</div>
	<div class="container"> %v </div>
	</div>
	</body>
	</html>`,
		len(album.dir.Dirs),
		strings.Join(Dir2Html(pathName, album.dir), "\n"),
		len(album.dir.Images),
		some_files_size_str(album.dir.AbsImages),
		pagination,
		strings.Join(htmlImages, "\n"),
	)))
	//w.Write([]byte("hah"))
	//return
}

func Img2Html(pathName string, dir *Dir, page int) (string, []string, int) {
	var (
		pagination string
		htmlImages []string
	)

	_images, previous, next, page := Page(dir.Images, page, size)
	_abs_images, previous, next, page := Page(dir.AbsImages, page, size)

	// add pagination
	var htmlPrevious, htmlNext string
	if previous {
		newUrl, _ := AddQuery(pathName, "page", strconv.Itoa(page-1))
		htmlPrevious = fmt.Sprintf(`<a class="previous" href="%v">←Previous</a>`, newUrl)
	}
	if next {
		newUrl, _ := AddQuery(pathName, "page", strconv.Itoa(page+1))
		htmlNext = fmt.Sprintf(`<a class="next" href="%v">Next→</a>`, newUrl)
	}
	if previous || next {
		pagination = fmt.Sprintf(`<div class="pagination card">%v%v</div>`, htmlPrevious, htmlNext)
		//pagination = htmlPrevious + htmlNext
	}

	for index, file := range _images {
		u, _ := url.Parse(pathName[6:])
		u.Path = path.Join("/s/", u.Path, file)

		htmlImages = append(htmlImages, fmt.Sprintf(`<p class="file"><a class="file" href="%v" title="%v">%v</a><span class="size right">%v</span></p>`,
			"/s"+path.Join(pathName[6:], file),
			fmt.Sprintf("%v [%v]", file, file_size_str(_abs_images[index])),
			file,
			file_size_str(_abs_images[index])))
	}
	return pagination, htmlImages, page
}

func Dir2Html(pathName string, dir *Dir) []string {
	rv := []string{}
	for index, file := range dir.Dirs {
		if hasPhoto(dir.AbsDirs[index]) {
			sub_dir := NewDir(dir.AbsDirs[index])

			rv = append(rv, fmt.Sprintf(
				`<div class="directory"><a class="link" href="%v">%v</a><span class="count right">[%v]</span><span class="right">%v</span></div>`,
				"/index"+fp.Join(pathName[6:], file)+"/",
				file+"/",
				len(sub_dir.Images),
				dir_images_size_str(dir.AbsDirs[index])))
		}
	}
	return rv
}

func Page(items []string, page, size int) ([]string, bool, bool, int) {
	if len(items) == 0 {
		return []string{}, false, false, 1
	}

	end := size * page
	start := end - size
	next := end < len(items)

	if len(items) <= start {
		_page := len(items) / size
		if _page*size < len(items) {
			_page++
		}
		return Page(items, _page, size)
	}
	if !next {
		end = len(items)
	}
	return items[start:end], page > 1, next, page
}

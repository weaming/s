package main

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"

    "github.com/weaming/itree/filetree"
)

const rates = 1024

type WWW struct {
	root string
	node *filetree.FileNode
}

func (p WWW) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logit(r)
	if NEED_AUTH {
		mybasicAuth(p.handlerFunc, ADMIN, PASSWORD)(w, r)
	} else {
		p.handlerFunc(w, r)
	}
}

func (p WWW) handlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "Happy coding!")
	pathName := r.URL.Path

	page, err := getQueryInt(r, "page")
	if err != nil {
		//target, _ := AddQuery(pathName, "page", "1")
		//http.Redirect(w, r, target, http.StatusFound)
		//return
		page = 1
	}

	filePath := filepath.Join(p.root, pathName[6:])
	node := filetree.NewFileNode(filePath, filePath, nil, false)
	if node == nil {
		w.Write([]byte("Invalid URL"))
		return
	} else {
		p.node = node
	}

	pagination, htmlFiles, realPage := HtmlOfFiles(pathName, p.node, page)
	if realPage != page {
		fmt.Println(realPage)
		target, _ := AddQuery(pathName, "page", strconv.Itoa(realPage))
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

	body {
		background: linear-gradient(to bottom right, #352482, #EA8F78);
		min-height: 100vh;
		margin: 0;
		padding-top: 1rem;
		font-family 'Avenir', Helvetica, Arial, sans-serif;
	}
	h3 { margin: 0.6rem 1rem; }

	.card{
		background-color: #fff;
		box-shadow: 0 2px 5px 0 rgba(0, 0, 0, .16), 0 2px 10px 0 rgba(0, 0, 0, .12);
		margin: 0 auto 1rem auto;
		padding: 2rem;
		max-width: 61.8vw;
		border-radius: 5px;
		font-size: 18px;
		background: #F6EEF7;
	}
	a:link{color: #1a0dab;}
	a:visited{color: #609;}

	.directory, .file {
		padding: 5px 20px;
		border-radius: 3px;
	}
	.directory:hover, .file:hover{
		background: #f5f5f5;
	}

	<!--pagination-->
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
		len(p.node.Dirs),
		strings.Join(HtmlOfDirs(pathName, p.node), "\n"),
		len(p.node.Files),
		totalSizeStrOfNodes(p.node.Files),
		pagination,
		strings.Join(htmlFiles, "\n"),
	)))
	//w.Write([]byte("hah"))
	//return
}

func HtmlOfFiles(pathName string, node *filetree.FileNode, page int) (string, []string, int) {
	var (
		pagination string
		htmlFiles  []string
	)

	currentFiles, previous, next, realPage := Paging(node.Files, page, size)

	// add pagination
	var htmlPrevious, htmlNext string
	if previous {
		newUrl, _ := AddQuery(pathName, "page", strconv.Itoa(realPage-1))
		htmlPrevious = fmt.Sprintf(`<a class="previous" href="%v">←Previous</a>`, newUrl)
	}
	if next {
		newUrl, _ := AddQuery(pathName, "page", strconv.Itoa(realPage+1))
		htmlNext = fmt.Sprintf(`<a class="next" href="%v">Next→</a>`, newUrl)
	}
	if previous || next {
		pagination = fmt.Sprintf(`<div class="pagination card">%v%v</div>`, htmlPrevious, htmlNext)
		//pagination = htmlPrevious + htmlNext
	}

	// body
	for _, fileNode := range currentFiles {
		u, _ := url.Parse(pathName[6:])
		u.Path = path.Join("/s/", u.Path)

		htmlFiles = append(
			htmlFiles,
			fmt.Sprintf(`<div class="file">
				<a class="file-link" href="%v" title="%v">%v</a>
				<span class="size right">%v</span>
				</div>`,
				"/s"+path.Join(pathName[6:], fileNode.Name),
				fmt.Sprintf("%v [%v]", fileNode.Name, fileNode.TotalSize),
				fileNode.Name,
				filetree.HumanSize(fileNode.TotalSize, rates)))
	}
	return pagination, htmlFiles, realPage
}

func HtmlOfDirs(pathName string, node *filetree.FileNode) []string {
	var rv []string

	for _, dirNode := range node.Dirs {
		if hasFile(dirNode) {
			rv = append(rv, fmt.Sprintf(
				`<div class="directory"><a class="link" href="%v">%v</a><span class="count right">%v</span></div>`,
				"/index"+filepath.Join(pathName[6:], dirNode.Name)+"/",
				dirNode.Name+"/",
				fmt.Sprintf("%v [%v]",
					filetree.HumanSize(dirNode.TotalSize, rates),
					len(dirNode.Children))))
		}
	}
	return rv
}

func Paging(items []*filetree.FileNode, page, size int) ([]*filetree.FileNode, bool, bool, int) {
	if len(items) == 0 {
		return []*filetree.FileNode{}, false, false, 1
	}

	end := size * page
	start := end - size
	next := end < len(items)

	if len(items) <= start {
		lastPage := len(items) / size
		if lastPage*size < len(items) {
			lastPage++
		}
		return Paging(items, lastPage, size)
	}

	if !next {
		end = len(items)
	}

	return items[start:end], page > 1, next, page
}

package main

import (
	"net/http"

	"github.com/weaming/go-git-http"
)

func serveGit(path string, urlPrefix string) {
	// Get git handler to serve a directory of repos
	git := githttp.New(path, urlPrefix)

	// Attach handler to http server
	http.Handle(urlPrefix, git)
}

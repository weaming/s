package main

import (
	"net/http"

	githttp "github.com/weaming/go-git-http"
)

func ServeGit(path string, urlPrefix string) {
	// Get git handler to serve a directory of repos
	git := githttp.New(path, urlPrefix)

	// Attach handler to http server
	http.Handle(urlPrefix, git)
}

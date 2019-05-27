package main

import (
	"net/http"
)

func ServeFileWebsocket(path string, urlPrefix string) {
	handler := NewWatcherMux(path, urlPrefix)
	http.Handle(urlPrefix, handler)
}

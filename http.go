package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/NYTimes/gziphandler"
)

func redirect_handler(to string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logit(r)
		http.Redirect(w, r, to, 302)
	}
}

func Redirect(from, to string) {
	http.Handle(from, redirect_handler(to))
}

func ServeDir(prefix, path string) {
	handler := http.StripPrefix(prefix, http.FileServer(http.Dir(path)))
	_handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Server", "https://github.com/weaming/s")
		handler.ServeHTTP(w, r)
	})
	http.Handle(prefix, gziphandler.GzipHandler(_handler))
}

func ServeFile(path, fp string) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		logit(r)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Server", "https://github.com/weaming/s")
		http.ServeFile(w, r, fp)
	})
}

func logit(r *http.Request) {
	log.Printf(`%v "%v %v %v"`, r.RemoteAddr, r.Method, r.RequestURI, r.Proto)
}

func getQuery(r *http.Request, name string) (v string) {
	v = r.URL.Query().Get("page")
	return
}

func getQueryInt(r *http.Request, name string) (v int, err error) {
	v_str := getQuery(r, name)
	v, err = strconv.Atoi(v_str)
	return
}

// BasicAuth wraps a handler requiring HTTP basic auth for it using the given
// username and password and the specified realm, which shouldn't contain quotes.
//
// Most web browser display a dialog with something like:
//
//    The website says: "<realm>"
//
// Which is really stupid so you may want to set the realm to a message rather than
// an actual realm.
func BasicAuth(handler http.HandlerFunc, realm string, check func(string, string) bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || !check(user, pass) {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}
		handler(w, r)
	}
}

func mybasicAuth(handler http.HandlerFunc, username, password string) http.HandlerFunc {
	return BasicAuth(handler, "hello", func(user, pass string) bool {
		if user == username && pass == password {
			//green(fmt.Sprintf("Auth success: name: [%v]; password: [%v]", user, pass))
			return true
		} else {
			red(fmt.Sprintf("Auth fail: name: [%v]; password: [%v]", user, pass))
			return false
		}
	})
}

func GetIntranetIP() (rv []string) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				rv = append(rv, ipnet.IP.String())
			}
		}
	}
	return
}

func HTTPGetQuery(req *http.Request, name, dft string) string {
	if values, ok := req.URL.Query()[name]; ok {
		return values[len(values)-1]
	} else {
		return dft
	}
}

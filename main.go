package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	fp "path/filepath"
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

	http.Handle("/index/", WWW{root: ROOT})
	ServeDir("/s/", ROOT)

	fmt.Printf("Open http://127.0.0.1:%v to enjoy!\n", strings.Split(*LISTEN, ":")[1])
	for _, ip := range GetIntranetIP() {
		fmt.Printf("Your intranet IP: %v ==> http://%v:%v\n", ip, ip, strings.Split(*LISTEN, ":")[1])
	}
	log.Fatal(http.ListenAndServe(*LISTEN, nil))
}

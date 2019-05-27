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
var GIT bool
var ADMIN, PASSWORD string
var PURE_STATIC bool

func main() {
	var LISTEN = flag.String("l", ":8000", "Listen [host]:port, default bind to 0.0.0.0")

	flag.BoolVar(&NEED_AUTH, "a", false, "Whether need authorization.")
	flag.BoolVar(&GIT, "git", false, "Whether serve as git protocol smart http")
	flag.BoolVar(&PURE_STATIC, "pure", false, "serve static on /")
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

	fmt.Println(green("To be listed direcotry: [%v]", ROOT))

	// basic authentication
	if NEED_AUTH {
		if PASSWORD == DEFAULT_PW {
			fmt.Println(red("Warning: set yourself password by option -p"))
		}
		fmt.Println(green("Your basic auth name and password: [%v:%v]", ADMIN, PASSWORD))
	} else {
		fmt.Println(red("Warning: please set your HTTP basic authentication"))
	}

	if PURE_STATIC {
		ServeDir("/", ROOT)
	} else {
		if GIT {
			urlPrefix := "/"
			fmt.Println(red("Serve git smart http on path: %v", urlPrefix))
			ServeGit(ROOT, urlPrefix)
		} else {
			Redirect("/", "/index/")

			urlPrefix := "/git/"
			fmt.Println(green("Serve git smart http on path: %v", urlPrefix))
			ServeGit(ROOT, urlPrefix)
		}

		ServeFile("/favicon.ico", fp.Join(ROOT, "./favicon.ico"))
		http.Handle("/index/", WWW{root: ROOT})
		ServeDir("/s/", ROOT)
		ServeFileWebsocket(ROOT, "/ws/")
	}

	fmt.Printf("Open http://127.0.0.1:%v to enjoy!\n", strings.Split(*LISTEN, ":")[1])
	for _, ip := range GetIntranetIP() {
		fmt.Printf("Your intranet IP: %v ==> http://%v:%v\n", ip, ip, strings.Split(*LISTEN, ":")[1])
	}
	log.Fatal(http.ListenAndServe(*LISTEN, nil))
}

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

const defaultPassword = "admin"

type Config struct {
	pageSize    int
	needAuth    bool
	git         bool
	admin       string
	password    string
	pureStatic  bool
}

var config Config

func main() {
	var LISTEN = flag.String("l", ":8000", "Listen [host]:port, default bind to 0.0.0.0")

	flag.BoolVar(&config.needAuth, "a", false, "Whether need authorization.")
	flag.BoolVar(&config.git, "git", false, "Whether serve as git protocol smart http")
	flag.BoolVar(&config.pureStatic, "pure", false, "serve static on /")
	flag.StringVar(&config.admin, "u", "admin", "Basic authorization username")
	flag.StringVar(&config.password, "p", defaultPassword, "Basic authorization password")
	flag.IntVar(&config.pageSize, "n", 20, "The maximum number of files in each page.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] ROOT\nThe ROOT is the directory to be serve.\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// check the directory path
	ROOT, err := fp.Abs(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid path:", err)
		os.Exit(1)
	}
	fi, err := os.Stat(ROOT)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !fi.IsDir() {
		fmt.Fprintln(os.Stderr, "The path should be a directory!!")
		os.Exit(1)
	}

	fmt.Println(green("To be listed directory: [%v]", ROOT))

	// basic authentication
	if config.needAuth {
		if config.password == defaultPassword {
			fmt.Println(red("Warning: set yourself password by option -p"))
		}
		fmt.Println(green("Your basic auth name and password: [%v:%v]", config.admin, config.password))
	} else {
		fmt.Println(red("Warning: please set your HTTP basic authentication"))
	}

	if config.pureStatic {
		ServeDir("/", ROOT)
	} else {
		if config.git {
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

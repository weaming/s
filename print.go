package main

import "fmt"

var NC = "\033[0m"

func red(text string) {
	RED := "\033[0;31m"
	fmt.Printf("%v%v%v\n", RED, text, NC)
}

func green(text string) {
	GREEN := "\033[0;32m"
	fmt.Printf("%v%v%v\n", GREEN, text, NC)
}

func brown(text string) {
	BROWN := "\033[0;33m"
	fmt.Printf("%v%v%v\n", BROWN, text, NC)
}

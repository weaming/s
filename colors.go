package main

import "fmt"

const (
	NC    = "\033[0m"
	RED   = "\033[0;31m"
	GREEN = "\033[0;32m"
	BROWN = "\033[0;33m"
)

func color(color string, format string, a ...interface{}) string {
	text := fmt.Sprintf(format, a...)
	return fmt.Sprintf("%v%v%v", color, text, NC)
}

func green(format string, a ...interface{}) string {
	return color(GREEN, format, a...)
}

func red(format string, a ...interface{}) string {
	return color(RED, format, a...)
}

func brown(format string, a ...interface{}) string {
	return color(BROWN, format, a...)
}

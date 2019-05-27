package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func PanicErr(e error) {
	if e != nil {
		panic(e)
	}
}

func MarshalMust(v interface{}) []byte {
	b, err := json.Marshal(v)
	PanicErr(err)
	return b
}

func ExistFile(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func MustNil(e interface{}) {
	if e != nil {
		panic(e)
	}
}

func Sha256(s string) string {
	return strings.ToUpper(fmt.Sprintf("%x", sha256.Sum256([]byte(s))))
}

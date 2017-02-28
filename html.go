package main

import (
	"net/url"
)

func UrlEncoded(str string) string {
	u, err := url.Parse(str)
	if err != nil {
		return ""
	}
	return u.String()
}

func AddQuery(pathName, key, value string) (string, error) {
	u, err := url.Parse(pathName)
	if err != nil {
		return pathName, err
	}
	q := u.Query()
	q.Set("page", value)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

package utils

import (
	"net/url"
	"strings"
)

func ParseQueryString(qry string) map[string] string {
	b, _ := url.PathUnescape(qry)
	s := strings.Split(b,"&")
	var m = make(map[string] string)
	for _, v := range s {
		q := strings.Split(v,"=")
		m[q[0]] = q[1]
	}
	return m
}
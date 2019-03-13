package pkg

import "strings"

func UrlJoin(base, ele string) string {
	if !strings.HasSuffix(base, "/") {
		base = base + "/"
	}
	if len(ele) > 0 && strings.HasPrefix(ele, "/") {
		base = base[1:]
	}
	return base + ele
}

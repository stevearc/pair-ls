package util

import (
	"fmt"
	"net/url"
	"unicode"

	"github.com/sourcegraph/go-lsp"
)

func IsWindowsDriveURI(uri string) bool {
	if len(uri) < 4 {
		return false
	}
	return uri[0] == '/' && unicode.IsLetter(rune(uri[1])) && uri[2] == ':'
}

func FromURI(uri lsp.DocumentURI) (string, error) {
	u, err := url.ParseRequestURI(string(uri))
	if err != nil {
		return "", err
	}
	if u.Scheme != "file" {
		return "", fmt.Errorf("only file URIs are supported, got %v", u.Scheme)
	}
	if IsWindowsDriveURI(u.Path) {
		u.Path = u.Path[1:]
	}
	return u.Path, nil
}

func ContainsStr(container []string, needle string) bool {
	for _, v := range container {
		if v == needle {
			return true
		}
	}
	return false
}

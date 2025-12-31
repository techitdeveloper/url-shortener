package utils

import (
	"net/url"
	"strings"
)

func IsValidURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}

	if !strings.HasPrefix(urlStr, "https://") && !strings.HasPrefix(urlStr, "http://") {
		urlStr = "http://" + urlStr
	}

	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return false
	}

	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

func NormalizeURL(urlStr string) string {
	if !strings.HasPrefix(urlStr, "https://") && !strings.HasPrefix(urlStr, "http://") {
		return "http://" + urlStr
	}
	return urlStr
}

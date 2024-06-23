package main

import (
	"strings"
	"unicode/utf8"
)

// Clip は s を max で切り取る
func Clip(s string, max int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	if utf8.RuneCountInString(s) > max {
		s = string([]rune(s)[0:max]) + "..."
	}

	return s
}

// CleanupTitle は title から siteName を除去する
func CleanupTitle(title, siteName string) string {
	if title == "" || title == siteName {
		return siteName
	}
	if siteName != "" && strings.Contains(title, siteName) {
		title = strings.TrimSuffix(title, siteName)
		title = strings.TrimSpace(title)
		title = strings.TrimRight(title, "-|:・")
		title = strings.TrimSpace(title)
		return title
	}

	return title
}

// https://git.deanishe.net/deanishe/go-favicon/src/tag/v0.1.0/icon.go#L57-L65
// used for sorting icons
// higher number = higher priority
var formatRank = map[string]int{
	"image/svg":                10,
	"image/svg+xml":            10,
	"image/png":                9,
	"image/x-icon":             8, // .ico
	"image/vnd.microsoft.icon": 8, // .ico
	"image/jpeg":               7,
}

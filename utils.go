package main

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// cmp.Orを使えば良さそう
// func ChooseOr(ss ...string) string {
// 	for _, v := range ss {
// 		if v != "" {
// 			return v
// 		}
// 	}
// 	return ""
// }

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
	if title != "" && siteName != "" && strings.Contains(title, siteName) {
		siteName = regexp.QuoteMeta(siteName)
		re := regexp.MustCompile(fmt.Sprintf(`^(\S+?)\s+?[\-\|:・]\s+?%s$`, siteName))
		s := re.FindStringSubmatch(title)
		if len(s) > 0 {
			return s[1]
		}
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

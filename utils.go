package main

import (
	"fmt"
	"regexp"
	"strings"
)

func ChooseOr(ss ...string) string {
	for _, v := range ss {
		if v != "" {
			return v
		}
	}
	return ""
}

// Clip は s を max で切り取る
func Clip(s string, max int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	if len(s) > max {
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

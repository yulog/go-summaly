package main

import "strings"

func ChooseOr(ss ...string) string {
	for _, v := range ss {
		if v != "" {
			return v
		}
	}
	return ""
}

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

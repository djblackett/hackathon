package utils

import (
	"regexp"
	"strings"
)

var invalid = regexp.MustCompile(`[^a-z0-9\-_]+`)

func Sanitize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = invalid.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-_")
	if len(s) > 60 {
		s = s[:60]
	}

	banned := []string{"document", "file", "note", "notes", "txt"}
	for _, b := range banned {
		s = strings.TrimSuffix(s, "-"+b)
		s = strings.TrimSuffix(s, "_"+b)
	}
	return s
}

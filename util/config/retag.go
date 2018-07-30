package config

import (
	"strings"
	"unicode"
)

const (
	delimiter = '_'
)

func toSnakeCase(src string) string {
	src = strings.TrimSpace(src)
	buffer := make([]rune, 0, len(src)+3)

	var prev rune
	var curr rune
	for _, next := range src {
		if isDelimiter(curr) {
			if !isDelimiter(prev) {
				buffer = append(buffer, delimiter)
			}
		} else if unicode.IsUpper(curr) {
			if unicode.IsLower(prev) || (unicode.IsUpper(prev) && unicode.IsLower(next)) {
				buffer = append(buffer, delimiter)
			}
			buffer = append(buffer, unicode.ToLower(curr))
		} else if curr != 0 {
			buffer = append(buffer, curr)
		}
		prev = curr
		curr = next
	}

	if len(src) > 0 {
		if unicode.IsUpper(curr) && unicode.IsLower(prev) && prev != 0 {
			buffer = append(buffer, delimiter)
		}
		buffer = append(buffer, unicode.ToLower(curr))
	}

	return string(buffer)
}

func isDelimiter(ch rune) bool {
	return ch == '-' || ch == '_' || unicode.IsSpace(ch)
}

package util

import (
	"strings"
)

func ExtractMethod(fullMethod string) string {
	parts := strings.Split(fullMethod, "/")
	return parts[len(parts)-1]
}

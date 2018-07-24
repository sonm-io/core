package xgrpc

import (
	"strings"
)

type methodInfo struct {
	Service string
	Method  string
}

func (m *methodInfo) IntoTuple() (string, string) {
	return m.Service, m.Method
}

func MethodInfo(fullMethod string) *methodInfo {
	parts := strings.SplitN(fullMethod, "/", 3)
	if len(parts) != 3 {
		return nil
	}

	m := &methodInfo{
		Service: parts[1],
		Method:  parts[2],
	}

	return m
}

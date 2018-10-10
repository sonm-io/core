package xgrpc

import (
	"strings"
)

type MethodInfo struct {
	Service string
	Method  string
}

func (m *MethodInfo) IntoTuple() (string, string) {
	return m.Service, m.Method
}

func ParseMethodInfo(fullMethod string) *MethodInfo {
	parts := strings.SplitN(fullMethod, "/", 3)
	if len(parts) != 3 {
		return nil
	}

	m := &MethodInfo{
		Service: parts[1],
		Method:  parts[2],
	}

	return m
}

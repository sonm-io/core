package logging

import (
	"google.golang.org/grpc/grpclog"
)

type NullLogger struct{}

// NewNullGRPCLogger returns new grpclog.Logger that log nothing.
//
// useful for debugging to suppress lot of errors
// from gRPC transport when external service is unavailable
//
// Note: logger must be set into the init() function, e.g:
//		func init() {
//     		grpclog.SetLogger(logging.NewNullGRPCLogger())
// 		}
func NewNullGRPCLogger() grpclog.Logger {
	return &NullLogger{}
}

func (n *NullLogger) Fatal(args ...interface{})                 {}
func (n *NullLogger) Fatalf(format string, args ...interface{}) {}
func (n *NullLogger) Fatalln(args ...interface{})               {}
func (n *NullLogger) Print(args ...interface{})                 {}
func (n *NullLogger) Printf(format string, args ...interface{}) {}
func (n *NullLogger) Println(args ...interface{})               {}

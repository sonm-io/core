// Code generated by protoc-gen-go. DO NOT EDIT.
// source: tc.proto

package sonm

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type QOSSetAliasRequest struct {
	LinkName  string `protobuf:"bytes,1,opt,name=linkName" json:"linkName,omitempty"`
	LinkAlias string `protobuf:"bytes,2,opt,name=linkAlias" json:"linkAlias,omitempty"`
}

func (m *QOSSetAliasRequest) Reset()                    { *m = QOSSetAliasRequest{} }
func (m *QOSSetAliasRequest) String() string            { return proto.CompactTextString(m) }
func (*QOSSetAliasRequest) ProtoMessage()               {}
func (*QOSSetAliasRequest) Descriptor() ([]byte, []int) { return fileDescriptor17, []int{0} }

func (m *QOSSetAliasRequest) GetLinkName() string {
	if m != nil {
		return m.LinkName
	}
	return ""
}

func (m *QOSSetAliasRequest) GetLinkAlias() string {
	if m != nil {
		return m.LinkAlias
	}
	return ""
}

type QOSSetAliasResponse struct {
}

func (m *QOSSetAliasResponse) Reset()                    { *m = QOSSetAliasResponse{} }
func (m *QOSSetAliasResponse) String() string            { return proto.CompactTextString(m) }
func (*QOSSetAliasResponse) ProtoMessage()               {}
func (*QOSSetAliasResponse) Descriptor() ([]byte, []int) { return fileDescriptor17, []int{1} }

type QOSAddHTBShapingRequest struct {
	LinkName         string `protobuf:"bytes,1,opt,name=linkName" json:"linkName,omitempty"`
	LinkAlias        string `protobuf:"bytes,2,opt,name=linkAlias" json:"linkAlias,omitempty"`
	RateLimitEgress  uint64 `protobuf:"varint,3,opt,name=rateLimitEgress" json:"rateLimitEgress,omitempty"`
	RateLimitIngress uint64 `protobuf:"varint,4,opt,name=rateLimitIngress" json:"rateLimitIngress,omitempty"`
}

func (m *QOSAddHTBShapingRequest) Reset()                    { *m = QOSAddHTBShapingRequest{} }
func (m *QOSAddHTBShapingRequest) String() string            { return proto.CompactTextString(m) }
func (*QOSAddHTBShapingRequest) ProtoMessage()               {}
func (*QOSAddHTBShapingRequest) Descriptor() ([]byte, []int) { return fileDescriptor17, []int{2} }

func (m *QOSAddHTBShapingRequest) GetLinkName() string {
	if m != nil {
		return m.LinkName
	}
	return ""
}

func (m *QOSAddHTBShapingRequest) GetLinkAlias() string {
	if m != nil {
		return m.LinkAlias
	}
	return ""
}

func (m *QOSAddHTBShapingRequest) GetRateLimitEgress() uint64 {
	if m != nil {
		return m.RateLimitEgress
	}
	return 0
}

func (m *QOSAddHTBShapingRequest) GetRateLimitIngress() uint64 {
	if m != nil {
		return m.RateLimitIngress
	}
	return 0
}

type QOSAddHTBShapingResponse struct {
}

func (m *QOSAddHTBShapingResponse) Reset()                    { *m = QOSAddHTBShapingResponse{} }
func (m *QOSAddHTBShapingResponse) String() string            { return proto.CompactTextString(m) }
func (*QOSAddHTBShapingResponse) ProtoMessage()               {}
func (*QOSAddHTBShapingResponse) Descriptor() ([]byte, []int) { return fileDescriptor17, []int{3} }

type QOSRemoveHTBShapingRequest struct {
	LinkName string `protobuf:"bytes,1,opt,name=linkName" json:"linkName,omitempty"`
}

func (m *QOSRemoveHTBShapingRequest) Reset()                    { *m = QOSRemoveHTBShapingRequest{} }
func (m *QOSRemoveHTBShapingRequest) String() string            { return proto.CompactTextString(m) }
func (*QOSRemoveHTBShapingRequest) ProtoMessage()               {}
func (*QOSRemoveHTBShapingRequest) Descriptor() ([]byte, []int) { return fileDescriptor17, []int{4} }

func (m *QOSRemoveHTBShapingRequest) GetLinkName() string {
	if m != nil {
		return m.LinkName
	}
	return ""
}

type QOSRemoveHTBShapingResponse struct {
}

func (m *QOSRemoveHTBShapingResponse) Reset()                    { *m = QOSRemoveHTBShapingResponse{} }
func (m *QOSRemoveHTBShapingResponse) String() string            { return proto.CompactTextString(m) }
func (*QOSRemoveHTBShapingResponse) ProtoMessage()               {}
func (*QOSRemoveHTBShapingResponse) Descriptor() ([]byte, []int) { return fileDescriptor17, []int{5} }

type QOSFlushRequest struct {
}

func (m *QOSFlushRequest) Reset()                    { *m = QOSFlushRequest{} }
func (m *QOSFlushRequest) String() string            { return proto.CompactTextString(m) }
func (*QOSFlushRequest) ProtoMessage()               {}
func (*QOSFlushRequest) Descriptor() ([]byte, []int) { return fileDescriptor17, []int{6} }

type QOSFlushResponse struct {
}

func (m *QOSFlushResponse) Reset()                    { *m = QOSFlushResponse{} }
func (m *QOSFlushResponse) String() string            { return proto.CompactTextString(m) }
func (*QOSFlushResponse) ProtoMessage()               {}
func (*QOSFlushResponse) Descriptor() ([]byte, []int) { return fileDescriptor17, []int{7} }

func init() {
	proto.RegisterType((*QOSSetAliasRequest)(nil), "sonm.QOSSetAliasRequest")
	proto.RegisterType((*QOSSetAliasResponse)(nil), "sonm.QOSSetAliasResponse")
	proto.RegisterType((*QOSAddHTBShapingRequest)(nil), "sonm.QOSAddHTBShapingRequest")
	proto.RegisterType((*QOSAddHTBShapingResponse)(nil), "sonm.QOSAddHTBShapingResponse")
	proto.RegisterType((*QOSRemoveHTBShapingRequest)(nil), "sonm.QOSRemoveHTBShapingRequest")
	proto.RegisterType((*QOSRemoveHTBShapingResponse)(nil), "sonm.QOSRemoveHTBShapingResponse")
	proto.RegisterType((*QOSFlushRequest)(nil), "sonm.QOSFlushRequest")
	proto.RegisterType((*QOSFlushResponse)(nil), "sonm.QOSFlushResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for QOS service

type QOSClient interface {
	SetAlias(ctx context.Context, in *QOSSetAliasRequest, opts ...grpc.CallOption) (*QOSSetAliasResponse, error)
	AddHTBShaping(ctx context.Context, in *QOSAddHTBShapingRequest, opts ...grpc.CallOption) (*QOSAddHTBShapingResponse, error)
	RemoveHTBShaping(ctx context.Context, in *QOSRemoveHTBShapingRequest, opts ...grpc.CallOption) (*QOSRemoveHTBShapingResponse, error)
	// Flush completely resets
	Flush(ctx context.Context, in *QOSFlushRequest, opts ...grpc.CallOption) (*QOSFlushResponse, error)
}

type qOSClient struct {
	cc *grpc.ClientConn
}

func NewQOSClient(cc *grpc.ClientConn) QOSClient {
	return &qOSClient{cc}
}

func (c *qOSClient) SetAlias(ctx context.Context, in *QOSSetAliasRequest, opts ...grpc.CallOption) (*QOSSetAliasResponse, error) {
	out := new(QOSSetAliasResponse)
	err := grpc.Invoke(ctx, "/sonm.QOS/SetAlias", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *qOSClient) AddHTBShaping(ctx context.Context, in *QOSAddHTBShapingRequest, opts ...grpc.CallOption) (*QOSAddHTBShapingResponse, error) {
	out := new(QOSAddHTBShapingResponse)
	err := grpc.Invoke(ctx, "/sonm.QOS/AddHTBShaping", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *qOSClient) RemoveHTBShaping(ctx context.Context, in *QOSRemoveHTBShapingRequest, opts ...grpc.CallOption) (*QOSRemoveHTBShapingResponse, error) {
	out := new(QOSRemoveHTBShapingResponse)
	err := grpc.Invoke(ctx, "/sonm.QOS/RemoveHTBShaping", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *qOSClient) Flush(ctx context.Context, in *QOSFlushRequest, opts ...grpc.CallOption) (*QOSFlushResponse, error) {
	out := new(QOSFlushResponse)
	err := grpc.Invoke(ctx, "/sonm.QOS/Flush", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for QOS service

type QOSServer interface {
	SetAlias(context.Context, *QOSSetAliasRequest) (*QOSSetAliasResponse, error)
	AddHTBShaping(context.Context, *QOSAddHTBShapingRequest) (*QOSAddHTBShapingResponse, error)
	RemoveHTBShaping(context.Context, *QOSRemoveHTBShapingRequest) (*QOSRemoveHTBShapingResponse, error)
	// Flush completely resets
	Flush(context.Context, *QOSFlushRequest) (*QOSFlushResponse, error)
}

func RegisterQOSServer(s *grpc.Server, srv QOSServer) {
	s.RegisterService(&_QOS_serviceDesc, srv)
}

func _QOS_SetAlias_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QOSSetAliasRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QOSServer).SetAlias(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.QOS/SetAlias",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QOSServer).SetAlias(ctx, req.(*QOSSetAliasRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QOS_AddHTBShaping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QOSAddHTBShapingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QOSServer).AddHTBShaping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.QOS/AddHTBShaping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QOSServer).AddHTBShaping(ctx, req.(*QOSAddHTBShapingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QOS_RemoveHTBShaping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QOSRemoveHTBShapingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QOSServer).RemoveHTBShaping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.QOS/RemoveHTBShaping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QOSServer).RemoveHTBShaping(ctx, req.(*QOSRemoveHTBShapingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QOS_Flush_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QOSFlushRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QOSServer).Flush(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.QOS/Flush",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QOSServer).Flush(ctx, req.(*QOSFlushRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _QOS_serviceDesc = grpc.ServiceDesc{
	ServiceName: "sonm.QOS",
	HandlerType: (*QOSServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetAlias",
			Handler:    _QOS_SetAlias_Handler,
		},
		{
			MethodName: "AddHTBShaping",
			Handler:    _QOS_AddHTBShaping_Handler,
		},
		{
			MethodName: "RemoveHTBShaping",
			Handler:    _QOS_RemoveHTBShaping_Handler,
		},
		{
			MethodName: "Flush",
			Handler:    _QOS_Flush_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tc.proto",
}

func init() { proto.RegisterFile("tc.proto", fileDescriptor17) }

var fileDescriptor17 = []byte{
	// 346 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x93, 0x4f, 0x4f, 0xea, 0x40,
	0x14, 0xc5, 0x5f, 0x81, 0xf7, 0x02, 0x37, 0x79, 0x01, 0xaf, 0x41, 0xeb, 0x28, 0x06, 0x27, 0x2e,
	0x88, 0x89, 0x25, 0xd1, 0x0d, 0x89, 0x2b, 0x48, 0x34, 0x9a, 0x18, 0x6a, 0x3b, 0x6e, 0x74, 0x57,
	0x60, 0x02, 0x13, 0x69, 0x07, 0x3b, 0x83, 0xdf, 0xc9, 0x95, 0x5f, 0xd1, 0xb4, 0x94, 0x41, 0xfe,
	0x99, 0x18, 0x77, 0xd3, 0x73, 0x7f, 0xf7, 0x4c, 0xef, 0xb9, 0x19, 0x28, 0xea, 0xbe, 0x33, 0x89,
	0xa5, 0x96, 0x58, 0x50, 0x32, 0x0a, 0x69, 0x17, 0xd0, 0x73, 0x19, 0xe3, 0xba, 0x3d, 0x16, 0x81,
	0xf2, 0xf9, 0xeb, 0x94, 0x2b, 0x8d, 0x04, 0x8a, 0x63, 0x11, 0xbd, 0x74, 0x83, 0x90, 0xdb, 0x56,
	0xdd, 0x6a, 0x94, 0x7c, 0xf3, 0x8d, 0x47, 0x50, 0x4a, 0xce, 0x29, 0x6f, 0xe7, 0xd2, 0xe2, 0x42,
	0xa0, 0x55, 0xd8, 0x5d, 0xf2, 0x53, 0x13, 0x19, 0x29, 0x4e, 0xdf, 0x2d, 0xd8, 0xf7, 0x5c, 0xd6,
	0x1e, 0x0c, 0x6e, 0x1f, 0x3b, 0x6c, 0x14, 0x4c, 0x44, 0x34, 0xfc, 0xf5, 0x65, 0xd8, 0x80, 0x72,
	0x1c, 0x68, 0x7e, 0x2f, 0x42, 0xa1, 0xaf, 0x87, 0x31, 0x57, 0xca, 0xce, 0xd7, 0xad, 0x46, 0xc1,
	0x5f, 0x95, 0xf1, 0x0c, 0x2a, 0x46, 0xba, 0x8b, 0x66, 0x68, 0x21, 0x45, 0xd7, 0x74, 0x4a, 0xc0,
	0x5e, 0xff, 0xd5, 0x6c, 0x8e, 0x16, 0x10, 0xcf, 0x65, 0x3e, 0x0f, 0xe5, 0x1b, 0xff, 0xd1, 0x24,
	0xb4, 0x06, 0x87, 0x1b, 0x3b, 0x33, 0xe3, 0x1d, 0x28, 0x7b, 0x2e, 0xbb, 0x19, 0x4f, 0xd5, 0x28,
	0x73, 0xa3, 0x08, 0x95, 0x85, 0x34, 0xc3, 0x2e, 0x3e, 0x72, 0x90, 0xf7, 0x5c, 0x86, 0x6d, 0x28,
	0xce, 0x33, 0x46, 0xdb, 0x49, 0x36, 0xe9, 0xac, 0xaf, 0x91, 0x1c, 0x6c, 0xa8, 0x64, 0xf7, 0xfd,
	0xc1, 0x07, 0xf8, 0xbf, 0x34, 0x23, 0xd6, 0x0c, 0xbd, 0x69, 0x4d, 0xe4, 0x78, 0x5b, 0xd9, 0x38,
	0x3e, 0x41, 0x65, 0x75, 0x3e, 0xac, 0x9b, 0xae, 0x2d, 0xa1, 0x91, 0x93, 0x6f, 0x08, 0x63, 0xdd,
	0x82, 0xbf, 0x69, 0x10, 0x58, 0x35, 0xf4, 0xd7, 0xac, 0xc8, 0xde, 0xaa, 0x3c, 0xef, 0xec, 0x9c,
	0x3e, 0xd3, 0xa1, 0xd0, 0xa3, 0x69, 0xcf, 0xe9, 0xcb, 0xb0, 0x99, 0x50, 0xe7, 0x42, 0x36, 0xfb,
	0x32, 0xe6, 0xcd, 0xf4, 0x1d, 0x5c, 0x25, 0x52, 0xef, 0x5f, 0x7a, 0xbe, 0xfc, 0x0c, 0x00, 0x00,
	0xff, 0xff, 0x93, 0x6a, 0x86, 0x28, 0x1f, 0x03, 0x00, 0x00,
}

// Code generated by protoc-gen-go. DO NOT EDIT.
// source: hub.proto

/*
Package sonm is a generated protocol buffer package.

It is generated from these files:
	hub.proto
	insonmnia.proto
	miner.proto

It has these top-level messages:
	ListRequest
	ListReply
	HubInfoRequest
	HubStartTaskRequest
	HubStartTaskReply
	HubStatusMapRequest
	PingRequest
	PingReply
	InfoReply
	StopTaskRequest
	StopTaskReply
	TaskStatusRequest
	TaskStatusReply
	StatusMapReply
	ContainerResources
	ContainerRestartPolicy
	MinerInfoRequest
	MinerHandshakeRequest
	Limits
	MinerHandshakeReply
	MinerStartRequest
	MinerStartReply
	MinerStatusMapRequest
*/
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

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type ListRequest struct {
}

func (m *ListRequest) Reset()                    { *m = ListRequest{} }
func (m *ListRequest) String() string            { return proto.CompactTextString(m) }
func (*ListRequest) ProtoMessage()               {}
func (*ListRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type ListReply struct {
	Info map[string]*ListReply_ListValue `protobuf:"bytes,1,rep,name=info" json:"info,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *ListReply) Reset()                    { *m = ListReply{} }
func (m *ListReply) String() string            { return proto.CompactTextString(m) }
func (*ListReply) ProtoMessage()               {}
func (*ListReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ListReply) GetInfo() map[string]*ListReply_ListValue {
	if m != nil {
		return m.Info
	}
	return nil
}

type ListReply_ListValue struct {
	Values []string `protobuf:"bytes,1,rep,name=values" json:"values,omitempty"`
}

func (m *ListReply_ListValue) Reset()                    { *m = ListReply_ListValue{} }
func (m *ListReply_ListValue) String() string            { return proto.CompactTextString(m) }
func (*ListReply_ListValue) ProtoMessage()               {}
func (*ListReply_ListValue) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1, 0} }

func (m *ListReply_ListValue) GetValues() []string {
	if m != nil {
		return m.Values
	}
	return nil
}

type HubInfoRequest struct {
	Miner string `protobuf:"bytes,1,opt,name=miner" json:"miner,omitempty"`
}

func (m *HubInfoRequest) Reset()                    { *m = HubInfoRequest{} }
func (m *HubInfoRequest) String() string            { return proto.CompactTextString(m) }
func (*HubInfoRequest) ProtoMessage()               {}
func (*HubInfoRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *HubInfoRequest) GetMiner() string {
	if m != nil {
		return m.Miner
	}
	return ""
}

type HubStartTaskRequest struct {
	Miner    string `protobuf:"bytes,1,opt,name=miner" json:"miner,omitempty"`
	Registry string `protobuf:"bytes,2,opt,name=registry" json:"registry,omitempty"`
	Image    string `protobuf:"bytes,3,opt,name=image" json:"image,omitempty"`
	Auth     string `protobuf:"bytes,4,opt,name=auth" json:"auth,omitempty"`
}

func (m *HubStartTaskRequest) Reset()                    { *m = HubStartTaskRequest{} }
func (m *HubStartTaskRequest) String() string            { return proto.CompactTextString(m) }
func (*HubStartTaskRequest) ProtoMessage()               {}
func (*HubStartTaskRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *HubStartTaskRequest) GetMiner() string {
	if m != nil {
		return m.Miner
	}
	return ""
}

func (m *HubStartTaskRequest) GetRegistry() string {
	if m != nil {
		return m.Registry
	}
	return ""
}

func (m *HubStartTaskRequest) GetImage() string {
	if m != nil {
		return m.Image
	}
	return ""
}

func (m *HubStartTaskRequest) GetAuth() string {
	if m != nil {
		return m.Auth
	}
	return ""
}

type HubStartTaskReply struct {
	Id       string   `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Endpoint []string `protobuf:"bytes,2,rep,name=endpoint" json:"endpoint,omitempty"`
}

func (m *HubStartTaskReply) Reset()                    { *m = HubStartTaskReply{} }
func (m *HubStartTaskReply) String() string            { return proto.CompactTextString(m) }
func (*HubStartTaskReply) ProtoMessage()               {}
func (*HubStartTaskReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *HubStartTaskReply) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *HubStartTaskReply) GetEndpoint() []string {
	if m != nil {
		return m.Endpoint
	}
	return nil
}

type HubStatusMapRequest struct {
	Miner string `protobuf:"bytes,1,opt,name=miner" json:"miner,omitempty"`
}

func (m *HubStatusMapRequest) Reset()                    { *m = HubStatusMapRequest{} }
func (m *HubStatusMapRequest) String() string            { return proto.CompactTextString(m) }
func (*HubStatusMapRequest) ProtoMessage()               {}
func (*HubStatusMapRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *HubStatusMapRequest) GetMiner() string {
	if m != nil {
		return m.Miner
	}
	return ""
}

func init() {
	proto.RegisterType((*ListRequest)(nil), "sonm.ListRequest")
	proto.RegisterType((*ListReply)(nil), "sonm.ListReply")
	proto.RegisterType((*ListReply_ListValue)(nil), "sonm.ListReply.ListValue")
	proto.RegisterType((*HubInfoRequest)(nil), "sonm.HubInfoRequest")
	proto.RegisterType((*HubStartTaskRequest)(nil), "sonm.HubStartTaskRequest")
	proto.RegisterType((*HubStartTaskReply)(nil), "sonm.HubStartTaskReply")
	proto.RegisterType((*HubStatusMapRequest)(nil), "sonm.HubStatusMapRequest")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Hub service

type HubClient interface {
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingReply, error)
	List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListReply, error)
	Info(ctx context.Context, in *HubInfoRequest, opts ...grpc.CallOption) (*InfoReply, error)
	StartTask(ctx context.Context, in *HubStartTaskRequest, opts ...grpc.CallOption) (*HubStartTaskReply, error)
	StopTask(ctx context.Context, in *StopTaskRequest, opts ...grpc.CallOption) (*StopTaskReply, error)
	TaskStatus(ctx context.Context, in *TaskStatusRequest, opts ...grpc.CallOption) (*TaskStatusReply, error)
	MinerStatus(ctx context.Context, in *HubStatusMapRequest, opts ...grpc.CallOption) (*StatusMapReply, error)
}

type hubClient struct {
	cc *grpc.ClientConn
}

func NewHubClient(cc *grpc.ClientConn) HubClient {
	return &hubClient{cc}
}

func (c *hubClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingReply, error) {
	out := new(PingReply)
	err := grpc.Invoke(ctx, "/sonm.Hub/Ping", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hubClient) List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListReply, error) {
	out := new(ListReply)
	err := grpc.Invoke(ctx, "/sonm.Hub/List", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hubClient) Info(ctx context.Context, in *HubInfoRequest, opts ...grpc.CallOption) (*InfoReply, error) {
	out := new(InfoReply)
	err := grpc.Invoke(ctx, "/sonm.Hub/Info", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hubClient) StartTask(ctx context.Context, in *HubStartTaskRequest, opts ...grpc.CallOption) (*HubStartTaskReply, error) {
	out := new(HubStartTaskReply)
	err := grpc.Invoke(ctx, "/sonm.Hub/StartTask", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hubClient) StopTask(ctx context.Context, in *StopTaskRequest, opts ...grpc.CallOption) (*StopTaskReply, error) {
	out := new(StopTaskReply)
	err := grpc.Invoke(ctx, "/sonm.Hub/StopTask", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hubClient) TaskStatus(ctx context.Context, in *TaskStatusRequest, opts ...grpc.CallOption) (*TaskStatusReply, error) {
	out := new(TaskStatusReply)
	err := grpc.Invoke(ctx, "/sonm.Hub/TaskStatus", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hubClient) MinerStatus(ctx context.Context, in *HubStatusMapRequest, opts ...grpc.CallOption) (*StatusMapReply, error) {
	out := new(StatusMapReply)
	err := grpc.Invoke(ctx, "/sonm.Hub/MinerStatus", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Hub service

type HubServer interface {
	Ping(context.Context, *PingRequest) (*PingReply, error)
	List(context.Context, *ListRequest) (*ListReply, error)
	Info(context.Context, *HubInfoRequest) (*InfoReply, error)
	StartTask(context.Context, *HubStartTaskRequest) (*HubStartTaskReply, error)
	StopTask(context.Context, *StopTaskRequest) (*StopTaskReply, error)
	TaskStatus(context.Context, *TaskStatusRequest) (*TaskStatusReply, error)
	MinerStatus(context.Context, *HubStatusMapRequest) (*StatusMapReply, error)
}

func RegisterHubServer(s *grpc.Server, srv HubServer) {
	s.RegisterService(&_Hub_serviceDesc, srv)
}

func _Hub_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Hub/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hub_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Hub/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServer).List(ctx, req.(*ListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hub_Info_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HubInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServer).Info(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Hub/Info",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServer).Info(ctx, req.(*HubInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hub_StartTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HubStartTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServer).StartTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Hub/StartTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServer).StartTask(ctx, req.(*HubStartTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hub_StopTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServer).StopTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Hub/StopTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServer).StopTask(ctx, req.(*StopTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hub_TaskStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TaskStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServer).TaskStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Hub/TaskStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServer).TaskStatus(ctx, req.(*TaskStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hub_MinerStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HubStatusMapRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServer).MinerStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Hub/MinerStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServer).MinerStatus(ctx, req.(*HubStatusMapRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Hub_serviceDesc = grpc.ServiceDesc{
	ServiceName: "sonm.Hub",
	HandlerType: (*HubServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _Hub_Ping_Handler,
		},
		{
			MethodName: "List",
			Handler:    _Hub_List_Handler,
		},
		{
			MethodName: "Info",
			Handler:    _Hub_Info_Handler,
		},
		{
			MethodName: "StartTask",
			Handler:    _Hub_StartTask_Handler,
		},
		{
			MethodName: "StopTask",
			Handler:    _Hub_StopTask_Handler,
		},
		{
			MethodName: "TaskStatus",
			Handler:    _Hub_TaskStatus_Handler,
		},
		{
			MethodName: "MinerStatus",
			Handler:    _Hub_MinerStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "hub.proto",
}

func init() { proto.RegisterFile("hub.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 435 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x53, 0x5d, 0x8b, 0xda, 0x40,
	0x14, 0xdd, 0x7c, 0xec, 0xb2, 0xb9, 0xd2, 0xdd, 0x3a, 0x6a, 0x1b, 0xf3, 0x24, 0x53, 0x28, 0x42,
	0xdb, 0x08, 0xf6, 0x45, 0xfa, 0xd0, 0x8f, 0x87, 0x82, 0x85, 0x0a, 0x25, 0x96, 0xbe, 0x27, 0x38,
	0xea, 0xa0, 0x4e, 0x62, 0x32, 0x53, 0xc8, 0xcf, 0xe8, 0x6f, 0xe9, 0x1f, 0x2c, 0x33, 0x93, 0x99,
	0x1a, 0x57, 0x7c, 0xf2, 0x9e, 0x73, 0xcf, 0xb9, 0xf7, 0xe6, 0xc4, 0x40, 0xb0, 0x15, 0x59, 0x5c,
	0x94, 0x39, 0xcf, 0x91, 0x5f, 0xe5, 0xec, 0x10, 0x3d, 0x52, 0x26, 0x7f, 0x19, 0x4d, 0x35, 0x8d,
	0x9f, 0x41, 0xe7, 0x3b, 0xad, 0x78, 0x42, 0x8e, 0x82, 0x54, 0x1c, 0xff, 0x75, 0x20, 0xd0, 0xb8,
	0xd8, 0xd7, 0xe8, 0x1d, 0xf8, 0x94, 0xad, 0xf3, 0xd0, 0x19, 0x79, 0xe3, 0xce, 0x74, 0x18, 0x4b,
	0x6b, 0x6c, 0xdb, 0xf1, 0x37, 0xb6, 0xce, 0xbf, 0x32, 0x5e, 0xd6, 0x89, 0x92, 0x45, 0xaf, 0xb4,
	0xf7, 0x57, 0xba, 0x17, 0x04, 0xbd, 0x80, 0xbb, 0xdf, 0xb2, 0xa8, 0x94, 0x3b, 0x48, 0x1a, 0x14,
	0x25, 0x10, 0x58, 0x1f, 0x7a, 0x0e, 0xde, 0x8e, 0xd4, 0xa1, 0x33, 0x72, 0xc6, 0x41, 0x22, 0x4b,
	0x34, 0x81, 0x5b, 0x25, 0x0c, 0xdd, 0x91, 0x73, 0x69, 0xa7, 0x5d, 0x90, 0x68, 0xdd, 0x07, 0x77,
	0xe6, 0xe0, 0xd7, 0xf0, 0x30, 0x17, 0x99, 0x1c, 0xdb, 0x3c, 0x07, 0xea, 0xc3, 0xed, 0x81, 0x32,
	0x52, 0x36, 0xa3, 0x35, 0xc0, 0x47, 0xe8, 0xcd, 0x45, 0xb6, 0xe4, 0x69, 0xc9, 0x7f, 0xa6, 0xd5,
	0xee, 0xaa, 0x18, 0x45, 0x70, 0x5f, 0x92, 0x0d, 0xad, 0x78, 0x59, 0xab, 0x63, 0x82, 0xc4, 0x62,
	0xe9, 0xa0, 0x87, 0x74, 0x43, 0x42, 0x4f, 0x3b, 0x14, 0x40, 0x08, 0xfc, 0x54, 0xf0, 0x6d, 0xe8,
	0x2b, 0x52, 0xd5, 0xf8, 0x13, 0x74, 0xdb, 0x2b, 0x65, 0xae, 0x0f, 0xe0, 0xd2, 0x55, 0xb3, 0xcd,
	0xa5, 0x2b, 0xb9, 0x8a, 0xb0, 0x55, 0x91, 0x53, 0xc6, 0x43, 0x57, 0xa5, 0x65, 0x31, 0x7e, 0x63,
	0x6e, 0xe6, 0xa2, 0x5a, 0xa4, 0xc5, 0xd5, 0x9b, 0xa7, 0x7f, 0x3c, 0xf0, 0xe6, 0x22, 0x43, 0x6f,
	0xc1, 0xff, 0x41, 0xd9, 0x06, 0x75, 0x75, 0x7c, 0xb2, 0x6e, 0x8c, 0xd1, 0xe3, 0x29, 0x55, 0xec,
	0x6b, 0x7c, 0x23, 0xd5, 0x32, 0x56, 0xa3, 0x3e, 0xf9, 0x3f, 0x18, 0xb5, 0xcd, 0x1f, 0xdf, 0xa0,
	0x09, 0xf8, 0x32, 0x69, 0xd4, 0xd7, 0xad, 0x76, 0xf0, 0xc6, 0xa0, 0x29, 0x6d, 0xf8, 0x02, 0x81,
	0x7d, 0x7e, 0x34, 0xb4, 0xae, 0xf3, 0xd7, 0x10, 0xbd, 0xbc, 0xd4, 0xd2, 0x23, 0x66, 0x70, 0xbf,
	0xe4, 0x79, 0xa1, 0x26, 0x0c, 0xb4, 0xcc, 0x60, 0xe3, 0xee, 0x9d, 0xd3, 0xda, 0xf9, 0x11, 0x40,
	0x42, 0x9d, 0x1f, 0x6a, 0x56, 0xfc, 0x67, 0x8c, 0x7b, 0xf0, 0xb4, 0xa1, 0xfd, 0x9f, 0xa1, 0xb3,
	0x90, 0xd1, 0x36, 0x03, 0x5a, 0xe7, 0xb7, 0xde, 0x48, 0xd4, 0x37, 0x07, 0x58, 0x5e, 0x4d, 0xc8,
	0xee, 0xd4, 0x87, 0xf6, 0xfe, 0x5f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x2c, 0x07, 0x56, 0x5e, 0x8c,
	0x03, 0x00, 0x00,
}

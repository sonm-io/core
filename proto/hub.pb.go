// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proto/hub.proto

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

type MinerListRequest struct {
}

func (m *MinerListRequest) Reset()                    { *m = MinerListRequest{} }
func (m *MinerListRequest) String() string            { return proto.CompactTextString(m) }
func (*MinerListRequest) ProtoMessage()               {}
func (*MinerListRequest) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

type MinerListReply struct {
	Info map[string]*MinerListReply_ListValue `protobuf:"bytes,1,rep,name=info" json:"info,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *MinerListReply) Reset()                    { *m = MinerListReply{} }
func (m *MinerListReply) String() string            { return proto.CompactTextString(m) }
func (*MinerListReply) ProtoMessage()               {}
func (*MinerListReply) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{1} }

func (m *MinerListReply) GetInfo() map[string]*MinerListReply_ListValue {
	if m != nil {
		return m.Info
	}
	return nil
}

type MinerListReply_ListValue struct {
	Values []string `protobuf:"bytes,1,rep,name=values" json:"values,omitempty"`
}

func (m *MinerListReply_ListValue) Reset()                    { *m = MinerListReply_ListValue{} }
func (m *MinerListReply_ListValue) String() string            { return proto.CompactTextString(m) }
func (*MinerListReply_ListValue) ProtoMessage()               {}
func (*MinerListReply_ListValue) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{1, 0} }

func (m *MinerListReply_ListValue) GetValues() []string {
	if m != nil {
		return m.Values
	}
	return nil
}

type MinerInfoRequest struct {
	Miner string `protobuf:"bytes,1,opt,name=miner" json:"miner,omitempty"`
}

func (m *MinerInfoRequest) Reset()                    { *m = MinerInfoRequest{} }
func (m *MinerInfoRequest) String() string            { return proto.CompactTextString(m) }
func (*MinerInfoRequest) ProtoMessage()               {}
func (*MinerInfoRequest) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{2} }

func (m *MinerInfoRequest) GetMiner() string {
	if m != nil {
		return m.Miner
	}
	return ""
}

type HubStartTaskRequest struct {
	Miner         string `protobuf:"bytes,1,opt,name=miner" json:"miner,omitempty"`
	Registry      string `protobuf:"bytes,2,opt,name=registry" json:"registry,omitempty"`
	Image         string `protobuf:"bytes,3,opt,name=image" json:"image,omitempty"`
	Auth          string `protobuf:"bytes,4,opt,name=auth" json:"auth,omitempty"`
	PublicKeyData string `protobuf:"bytes,5,opt,name=PublicKeyData,json=publicKeyData" json:"PublicKeyData,omitempty"`
}

func (m *HubStartTaskRequest) Reset()                    { *m = HubStartTaskRequest{} }
func (m *HubStartTaskRequest) String() string            { return proto.CompactTextString(m) }
func (*HubStartTaskRequest) ProtoMessage()               {}
func (*HubStartTaskRequest) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{3} }

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

func (m *HubStartTaskRequest) GetPublicKeyData() string {
	if m != nil {
		return m.PublicKeyData
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
func (*HubStartTaskReply) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{4} }

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

type HubTaskListMapRequest struct {
	Miner string `protobuf:"bytes,1,opt,name=miner" json:"miner,omitempty"`
}

func (m *HubTaskListMapRequest) Reset()                    { *m = HubTaskListMapRequest{} }
func (m *HubTaskListMapRequest) String() string            { return proto.CompactTextString(m) }
func (*HubTaskListMapRequest) ProtoMessage()               {}
func (*HubTaskListMapRequest) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{5} }

func (m *HubTaskListMapRequest) GetMiner() string {
	if m != nil {
		return m.Miner
	}
	return ""
}

type HubStatusRequest struct {
}

func (m *HubStatusRequest) Reset()                    { *m = HubStatusRequest{} }
func (m *HubStatusRequest) String() string            { return proto.CompactTextString(m) }
func (*HubStatusRequest) ProtoMessage()               {}
func (*HubStatusRequest) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{6} }

type HubStatusReply struct {
	MinerCount uint64 `protobuf:"varint,1,opt,name=minerCount" json:"minerCount,omitempty"`
	Uptime     uint64 `protobuf:"varint,2,opt,name=uptime" json:"uptime,omitempty"`
}

func (m *HubStatusReply) Reset()                    { *m = HubStatusReply{} }
func (m *HubStatusReply) String() string            { return proto.CompactTextString(m) }
func (*HubStatusReply) ProtoMessage()               {}
func (*HubStatusReply) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{7} }

func (m *HubStatusReply) GetMinerCount() uint64 {
	if m != nil {
		return m.MinerCount
	}
	return 0
}

func (m *HubStatusReply) GetUptime() uint64 {
	if m != nil {
		return m.Uptime
	}
	return 0
}

func init() {
	proto.RegisterType((*MinerListRequest)(nil), "sonm.MinerListRequest")
	proto.RegisterType((*MinerListReply)(nil), "sonm.MinerListReply")
	proto.RegisterType((*MinerListReply_ListValue)(nil), "sonm.MinerListReply.ListValue")
	proto.RegisterType((*MinerInfoRequest)(nil), "sonm.MinerInfoRequest")
	proto.RegisterType((*HubStartTaskRequest)(nil), "sonm.HubStartTaskRequest")
	proto.RegisterType((*HubStartTaskReply)(nil), "sonm.HubStartTaskReply")
	proto.RegisterType((*HubTaskListMapRequest)(nil), "sonm.HubTaskListMapRequest")
	proto.RegisterType((*HubStatusRequest)(nil), "sonm.HubStatusRequest")
	proto.RegisterType((*HubStatusReply)(nil), "sonm.HubStatusReply")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Hub service

type HubClient interface {
	// Ping checks Hub network availability
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingReply, error)
	// Status shows Hub status
	Status(ctx context.Context, in *HubStatusRequest, opts ...grpc.CallOption) (*HubStatusReply, error)
	// MinerList returns Miners connected to Hub
	MinerList(ctx context.Context, in *MinerListRequest, opts ...grpc.CallOption) (*MinerListReply, error)
	// MinerInfo returns detailed info about conected Miner
	MinerInfo(ctx context.Context, in *MinerInfoRequest, opts ...grpc.CallOption) (*MinerStatusReply, error)
	// TaskXXX forward task operations to given Miner
	TaskStart(ctx context.Context, in *HubStartTaskRequest, opts ...grpc.CallOption) (*HubStartTaskReply, error)
	TaskStop(ctx context.Context, in *TaskStopRequest, opts ...grpc.CallOption) (*TaskStopReply, error)
	TaskStatus(ctx context.Context, in *TaskDetailsRequest, opts ...grpc.CallOption) (*TaskDetailsReply, error)
	TaskList(ctx context.Context, in *HubTaskListMapRequest, opts ...grpc.CallOption) (*TaskDetailsMapReply, error)
	TaskLogs(ctx context.Context, in *TaskLogsRequest, opts ...grpc.CallOption) (Hub_TaskLogsClient, error)
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

func (c *hubClient) Status(ctx context.Context, in *HubStatusRequest, opts ...grpc.CallOption) (*HubStatusReply, error) {
	out := new(HubStatusReply)
	err := grpc.Invoke(ctx, "/sonm.Hub/Status", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hubClient) MinerList(ctx context.Context, in *MinerListRequest, opts ...grpc.CallOption) (*MinerListReply, error) {
	out := new(MinerListReply)
	err := grpc.Invoke(ctx, "/sonm.Hub/MinerList", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hubClient) MinerInfo(ctx context.Context, in *MinerInfoRequest, opts ...grpc.CallOption) (*MinerStatusReply, error) {
	out := new(MinerStatusReply)
	err := grpc.Invoke(ctx, "/sonm.Hub/MinerInfo", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hubClient) TaskStart(ctx context.Context, in *HubStartTaskRequest, opts ...grpc.CallOption) (*HubStartTaskReply, error) {
	out := new(HubStartTaskReply)
	err := grpc.Invoke(ctx, "/sonm.Hub/TaskStart", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hubClient) TaskStop(ctx context.Context, in *TaskStopRequest, opts ...grpc.CallOption) (*TaskStopReply, error) {
	out := new(TaskStopReply)
	err := grpc.Invoke(ctx, "/sonm.Hub/TaskStop", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hubClient) TaskStatus(ctx context.Context, in *TaskDetailsRequest, opts ...grpc.CallOption) (*TaskDetailsReply, error) {
	out := new(TaskDetailsReply)
	err := grpc.Invoke(ctx, "/sonm.Hub/TaskStatus", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hubClient) TaskList(ctx context.Context, in *HubTaskListMapRequest, opts ...grpc.CallOption) (*TaskDetailsMapReply, error) {
	out := new(TaskDetailsMapReply)
	err := grpc.Invoke(ctx, "/sonm.Hub/TaskList", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hubClient) TaskLogs(ctx context.Context, in *TaskLogsRequest, opts ...grpc.CallOption) (Hub_TaskLogsClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Hub_serviceDesc.Streams[0], c.cc, "/sonm.Hub/TaskLogs", opts...)
	if err != nil {
		return nil, err
	}
	x := &hubTaskLogsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Hub_TaskLogsClient interface {
	Recv() (*TaskLogsChunk, error)
	grpc.ClientStream
}

type hubTaskLogsClient struct {
	grpc.ClientStream
}

func (x *hubTaskLogsClient) Recv() (*TaskLogsChunk, error) {
	m := new(TaskLogsChunk)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Hub service

type HubServer interface {
	// Ping checks Hub network availability
	Ping(context.Context, *PingRequest) (*PingReply, error)
	// Status shows Hub status
	Status(context.Context, *HubStatusRequest) (*HubStatusReply, error)
	// MinerList returns Miners connected to Hub
	MinerList(context.Context, *MinerListRequest) (*MinerListReply, error)
	// MinerInfo returns detailed info about conected Miner
	MinerInfo(context.Context, *MinerInfoRequest) (*MinerStatusReply, error)
	// TaskXXX forward task operations to given Miner
	TaskStart(context.Context, *HubStartTaskRequest) (*HubStartTaskReply, error)
	TaskStop(context.Context, *TaskStopRequest) (*TaskStopReply, error)
	TaskStatus(context.Context, *TaskDetailsRequest) (*TaskDetailsReply, error)
	TaskList(context.Context, *HubTaskListMapRequest) (*TaskDetailsMapReply, error)
	TaskLogs(*TaskLogsRequest, Hub_TaskLogsServer) error
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

func _Hub_Status_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HubStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServer).Status(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Hub/Status",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServer).Status(ctx, req.(*HubStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hub_MinerList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MinerListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServer).MinerList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Hub/MinerList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServer).MinerList(ctx, req.(*MinerListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hub_MinerInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MinerInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServer).MinerInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Hub/MinerInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServer).MinerInfo(ctx, req.(*MinerInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hub_TaskStart_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HubStartTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServer).TaskStart(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Hub/TaskStart",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServer).TaskStart(ctx, req.(*HubStartTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hub_TaskStop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TaskStopRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServer).TaskStop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Hub/TaskStop",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServer).TaskStop(ctx, req.(*TaskStopRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hub_TaskStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TaskDetailsRequest)
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
		return srv.(HubServer).TaskStatus(ctx, req.(*TaskDetailsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hub_TaskList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HubTaskListMapRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HubServer).TaskList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Hub/TaskList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HubServer).TaskList(ctx, req.(*HubTaskListMapRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hub_TaskLogs_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(TaskLogsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(HubServer).TaskLogs(m, &hubTaskLogsServer{stream})
}

type Hub_TaskLogsServer interface {
	Send(*TaskLogsChunk) error
	grpc.ServerStream
}

type hubTaskLogsServer struct {
	grpc.ServerStream
}

func (x *hubTaskLogsServer) Send(m *TaskLogsChunk) error {
	return x.ServerStream.SendMsg(m)
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
			MethodName: "Status",
			Handler:    _Hub_Status_Handler,
		},
		{
			MethodName: "MinerList",
			Handler:    _Hub_MinerList_Handler,
		},
		{
			MethodName: "MinerInfo",
			Handler:    _Hub_MinerInfo_Handler,
		},
		{
			MethodName: "TaskStart",
			Handler:    _Hub_TaskStart_Handler,
		},
		{
			MethodName: "TaskStop",
			Handler:    _Hub_TaskStop_Handler,
		},
		{
			MethodName: "TaskStatus",
			Handler:    _Hub_TaskStatus_Handler,
		},
		{
			MethodName: "TaskList",
			Handler:    _Hub_TaskList_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "TaskLogs",
			Handler:       _Hub_TaskLogs_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/hub.proto",
}

func init() { proto.RegisterFile("proto/hub.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 548 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x7c, 0x54, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0x8e, 0x13, 0x27, 0xaa, 0x07, 0x35, 0x4d, 0xb7, 0x4d, 0x71, 0x8d, 0x54, 0x45, 0x86, 0x43,
	0x0e, 0x10, 0x50, 0xe0, 0x50, 0x15, 0x21, 0x40, 0x0d, 0x52, 0x10, 0x54, 0xaa, 0x0c, 0x82, 0xb3,
	0x4d, 0xb6, 0xc9, 0x2a, 0xce, 0xda, 0xd8, 0xbb, 0x48, 0x7e, 0x12, 0xde, 0x81, 0xc7, 0xe0, 0xc9,
	0xd0, 0xec, 0xae, 0x8d, 0x9d, 0xb4, 0xbd, 0xed, 0x7c, 0x33, 0xdf, 0xcc, 0x37, 0x3f, 0x5a, 0x38,
	0x48, 0xb3, 0x44, 0x24, 0xcf, 0x57, 0x32, 0x9a, 0xa8, 0x17, 0xb1, 0xf3, 0x84, 0x6f, 0xbc, 0xa1,
	0x86, 0x19, 0x47, 0x8b, 0xb3, 0x50, 0x3b, 0x7d, 0x02, 0x83, 0x2b, 0xc6, 0x69, 0xf6, 0x99, 0xe5,
	0x22, 0xa0, 0x3f, 0x25, 0xcd, 0x85, 0xff, 0xd7, 0x82, 0x7e, 0x0d, 0x4c, 0xe3, 0x82, 0x4c, 0xc1,
	0x66, 0xfc, 0x26, 0x71, 0xad, 0x51, 0x67, 0xfc, 0x60, 0x7a, 0x36, 0xc1, 0x24, 0x93, 0x66, 0xcc,
	0xe4, 0x23, 0xbf, 0x49, 0x3e, 0x70, 0x91, 0x15, 0x81, 0x8a, 0xf5, 0x1e, 0x83, 0x83, 0xce, 0x6f,
	0x61, 0x2c, 0x29, 0x39, 0x81, 0xde, 0x2f, 0x7c, 0xe4, 0x2a, 0x85, 0x13, 0x18, 0xcb, 0xfb, 0x0e,
	0x4e, 0xc5, 0x23, 0x03, 0xe8, 0xac, 0x69, 0xe1, 0x5a, 0x23, 0x6b, 0xec, 0x04, 0xf8, 0x24, 0xaf,
	0xa0, 0xab, 0x02, 0xdd, 0xf6, 0xc8, 0xba, 0xb3, 0x70, 0x55, 0x25, 0xd0, 0xc1, 0x17, 0xed, 0x73,
	0xcb, 0x1f, 0x9b, 0xc6, 0x30, 0xbb, 0x69, 0x8c, 0x1c, 0x43, 0x77, 0x83, 0x98, 0xa9, 0xa0, 0x0d,
	0xff, 0xb7, 0x05, 0x47, 0x73, 0x19, 0x7d, 0x11, 0x61, 0x26, 0xbe, 0x86, 0xf9, 0xfa, 0xde, 0x68,
	0xe2, 0xc1, 0x5e, 0x46, 0x97, 0x2c, 0x17, 0x59, 0xa1, 0x44, 0x39, 0x41, 0x65, 0x23, 0x83, 0x6d,
	0xc2, 0x25, 0x75, 0x3b, 0x9a, 0xa1, 0x0c, 0x42, 0xc0, 0x0e, 0xa5, 0x58, 0xb9, 0xb6, 0x02, 0xd5,
	0x9b, 0x3c, 0x81, 0xfd, 0x6b, 0x19, 0xc5, 0xec, 0xc7, 0x27, 0x5a, 0xcc, 0x42, 0x11, 0xba, 0x5d,
	0xe5, 0xdc, 0x4f, 0xeb, 0xa0, 0xff, 0x16, 0x0e, 0x9b, 0xc2, 0x70, 0x15, 0x7d, 0x68, 0xb3, 0x85,
	0xd1, 0xd4, 0x66, 0x0b, 0x14, 0x44, 0xf9, 0x22, 0x4d, 0x18, 0x17, 0x6e, 0x5b, 0xcd, 0xb6, 0xb2,
	0xfd, 0x67, 0x30, 0x9c, 0xcb, 0x08, 0xb9, 0x38, 0xa3, 0xab, 0x30, 0xbd, 0x7f, 0x12, 0x04, 0x06,
	0xba, 0x9e, 0x90, 0x79, 0x79, 0x0c, 0x73, 0xe8, 0xd7, 0x30, 0x14, 0x70, 0x06, 0xa0, 0xc2, 0x2f,
	0x13, 0xc9, 0x85, 0x4a, 0x60, 0x07, 0x35, 0x04, 0x57, 0x2d, 0x53, 0xc1, 0x36, 0x7a, 0x69, 0x76,
	0x60, 0xac, 0xe9, 0x1f, 0x1b, 0x3a, 0x73, 0x19, 0x91, 0xa7, 0x60, 0x5f, 0x33, 0xbe, 0x24, 0x87,
	0x7a, 0x99, 0xf8, 0x36, 0xc5, 0xbc, 0x83, 0x3a, 0x94, 0xc6, 0x85, 0xdf, 0x22, 0xe7, 0xd0, 0xd3,
	0xc5, 0xc9, 0x89, 0x76, 0x6e, 0x2b, 0xf4, 0x8e, 0x77, 0x70, 0xcd, 0x7c, 0x0d, 0x4e, 0x75, 0x28,
	0x25, 0x79, 0xfb, 0xd6, 0x4b, 0x72, 0xf3, 0xa2, 0xfc, 0x16, 0x79, 0x63, 0xc8, 0x78, 0x3e, 0x0d,
	0x72, 0xed, 0x9e, 0xbc, 0x3a, 0xde, 0xac, 0xfd, 0x1e, 0x1c, 0x9c, 0xba, 0x5a, 0x1d, 0x39, 0xad,
	0x0b, 0x6c, 0xdc, 0x98, 0xf7, 0xf0, 0x36, 0x57, 0xd9, 0xf8, 0x9e, 0x4e, 0x91, 0xa4, 0x64, 0xa8,
	0xc3, 0x4a, 0xbb, 0x64, 0x1f, 0x6d, 0xc3, 0x9a, 0xf9, 0x0e, 0xc0, 0x14, 0xc7, 0xb1, 0xb9, 0xff,
	0x83, 0x66, 0x54, 0x84, 0x2c, 0xce, 0xb7, 0xe4, 0x37, 0x3c, 0x3a, 0xc3, 0x4c, 0xd7, 0x56, 0x93,
	0x7b, 0x54, 0x49, 0xdc, 0xbd, 0x23, 0xef, 0x74, 0x27, 0x85, 0x72, 0xea, 0x2c, 0x17, 0x26, 0x4b,
	0xb2, 0xcc, 0xeb, 0x1d, 0xa0, 0x7d, 0x4b, 0x07, 0x08, 0x5f, 0xae, 0x24, 0x5f, 0xfb, 0xad, 0x17,
	0x56, 0xd4, 0x53, 0xdf, 0xd3, 0xcb, 0x7f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x12, 0x5c, 0xc2, 0x55,
	0xce, 0x04, 0x00, 0x00,
}

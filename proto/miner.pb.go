// Code generated by protoc-gen-go. DO NOT EDIT.
// source: miner.proto

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

type MinerHandshakeRequest struct {
	Hub   string      `protobuf:"bytes,1,opt,name=hub" json:"hub,omitempty"`
	Tasks []*TaskInfo `protobuf:"bytes,2,rep,name=tasks" json:"tasks,omitempty"`
}

func (m *MinerHandshakeRequest) Reset()                    { *m = MinerHandshakeRequest{} }
func (m *MinerHandshakeRequest) String() string            { return proto.CompactTextString(m) }
func (*MinerHandshakeRequest) ProtoMessage()               {}
func (*MinerHandshakeRequest) Descriptor() ([]byte, []int) { return fileDescriptor8, []int{0} }

func (m *MinerHandshakeRequest) GetHub() string {
	if m != nil {
		return m.Hub
	}
	return ""
}

func (m *MinerHandshakeRequest) GetTasks() []*TaskInfo {
	if m != nil {
		return m.Tasks
	}
	return nil
}

type MinerHandshakeReply struct {
	Miner        string        `protobuf:"bytes,1,opt,name=miner" json:"miner,omitempty"`
	Capabilities *Capabilities `protobuf:"bytes,2,opt,name=capabilities" json:"capabilities,omitempty"`
	NatType      NATType       `protobuf:"varint,3,opt,name=natType,enum=sonm.NATType" json:"natType,omitempty"`
}

func (m *MinerHandshakeReply) Reset()                    { *m = MinerHandshakeReply{} }
func (m *MinerHandshakeReply) String() string            { return proto.CompactTextString(m) }
func (*MinerHandshakeReply) ProtoMessage()               {}
func (*MinerHandshakeReply) Descriptor() ([]byte, []int) { return fileDescriptor8, []int{1} }

func (m *MinerHandshakeReply) GetMiner() string {
	if m != nil {
		return m.Miner
	}
	return ""
}

func (m *MinerHandshakeReply) GetCapabilities() *Capabilities {
	if m != nil {
		return m.Capabilities
	}
	return nil
}

func (m *MinerHandshakeReply) GetNatType() NATType {
	if m != nil {
		return m.NatType
	}
	return NATType_NONE
}

type MinerStartRequest struct {
	Id            string                    `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Registry      string                    `protobuf:"bytes,2,opt,name=registry" json:"registry,omitempty"`
	Image         string                    `protobuf:"bytes,3,opt,name=image" json:"image,omitempty"`
	Auth          string                    `protobuf:"bytes,4,opt,name=auth" json:"auth,omitempty"`
	RestartPolicy *ContainerRestartPolicy   `protobuf:"bytes,5,opt,name=restartPolicy" json:"restartPolicy,omitempty"`
	Resources     *TaskResourceRequirements `protobuf:"bytes,6,opt,name=resources" json:"resources,omitempty"`
	PublicKeyData string                    `protobuf:"bytes,7,opt,name=PublicKeyData" json:"PublicKeyData,omitempty"`
	CommitOnStop  bool                      `protobuf:"varint,8,opt,name=commitOnStop" json:"commitOnStop,omitempty"`
	Env           map[string]string         `protobuf:"bytes,9,rep,name=env" json:"env,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// OrderId describes an unique order identifier.
	// It is here for proper resource allocation and limitation.
	OrderId string `protobuf:"bytes,10,opt,name=orderId" json:"orderId,omitempty"`
}

func (m *MinerStartRequest) Reset()                    { *m = MinerStartRequest{} }
func (m *MinerStartRequest) String() string            { return proto.CompactTextString(m) }
func (*MinerStartRequest) ProtoMessage()               {}
func (*MinerStartRequest) Descriptor() ([]byte, []int) { return fileDescriptor8, []int{2} }

func (m *MinerStartRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *MinerStartRequest) GetRegistry() string {
	if m != nil {
		return m.Registry
	}
	return ""
}

func (m *MinerStartRequest) GetImage() string {
	if m != nil {
		return m.Image
	}
	return ""
}

func (m *MinerStartRequest) GetAuth() string {
	if m != nil {
		return m.Auth
	}
	return ""
}

func (m *MinerStartRequest) GetRestartPolicy() *ContainerRestartPolicy {
	if m != nil {
		return m.RestartPolicy
	}
	return nil
}

func (m *MinerStartRequest) GetResources() *TaskResourceRequirements {
	if m != nil {
		return m.Resources
	}
	return nil
}

func (m *MinerStartRequest) GetPublicKeyData() string {
	if m != nil {
		return m.PublicKeyData
	}
	return ""
}

func (m *MinerStartRequest) GetCommitOnStop() bool {
	if m != nil {
		return m.CommitOnStop
	}
	return false
}

func (m *MinerStartRequest) GetEnv() map[string]string {
	if m != nil {
		return m.Env
	}
	return nil
}

func (m *MinerStartRequest) GetOrderId() string {
	if m != nil {
		return m.OrderId
	}
	return ""
}

type SocketAddr struct {
	Addr string `protobuf:"bytes,1,opt,name=addr" json:"addr,omitempty"`
	//
	// Actually an `uint16` here. Protobuf is so clear and handy.
	Port uint32 `protobuf:"varint,2,opt,name=port" json:"port,omitempty"`
}

func (m *SocketAddr) Reset()                    { *m = SocketAddr{} }
func (m *SocketAddr) String() string            { return proto.CompactTextString(m) }
func (*SocketAddr) ProtoMessage()               {}
func (*SocketAddr) Descriptor() ([]byte, []int) { return fileDescriptor8, []int{3} }

func (m *SocketAddr) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

func (m *SocketAddr) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	return 0
}

type MinerStartReply struct {
	Container string `protobuf:"bytes,1,opt,name=container" json:"container,omitempty"`
	// PortMap represent port mapping between container network and host ones.
	PortMap map[string]*Endpoints `protobuf:"bytes,2,rep,name=portMap" json:"portMap,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *MinerStartReply) Reset()                    { *m = MinerStartReply{} }
func (m *MinerStartReply) String() string            { return proto.CompactTextString(m) }
func (*MinerStartReply) ProtoMessage()               {}
func (*MinerStartReply) Descriptor() ([]byte, []int) { return fileDescriptor8, []int{4} }

func (m *MinerStartReply) GetContainer() string {
	if m != nil {
		return m.Container
	}
	return ""
}

func (m *MinerStartReply) GetPortMap() map[string]*Endpoints {
	if m != nil {
		return m.PortMap
	}
	return nil
}

type TaskInfo struct {
	Request *MinerStartRequest `protobuf:"bytes,1,opt,name=request" json:"request,omitempty"`
	Reply   *MinerStartReply   `protobuf:"bytes,2,opt,name=reply" json:"reply,omitempty"`
}

func (m *TaskInfo) Reset()                    { *m = TaskInfo{} }
func (m *TaskInfo) String() string            { return proto.CompactTextString(m) }
func (*TaskInfo) ProtoMessage()               {}
func (*TaskInfo) Descriptor() ([]byte, []int) { return fileDescriptor8, []int{5} }

func (m *TaskInfo) GetRequest() *MinerStartRequest {
	if m != nil {
		return m.Request
	}
	return nil
}

func (m *TaskInfo) GetReply() *MinerStartReply {
	if m != nil {
		return m.Reply
	}
	return nil
}

type Endpoints struct {
	Endpoints []*SocketAddr `protobuf:"bytes,1,rep,name=endpoints" json:"endpoints,omitempty"`
}

func (m *Endpoints) Reset()                    { *m = Endpoints{} }
func (m *Endpoints) String() string            { return proto.CompactTextString(m) }
func (*Endpoints) ProtoMessage()               {}
func (*Endpoints) Descriptor() ([]byte, []int) { return fileDescriptor8, []int{6} }

func (m *Endpoints) GetEndpoints() []*SocketAddr {
	if m != nil {
		return m.Endpoints
	}
	return nil
}

type MinerStatusMapRequest struct {
}

func (m *MinerStatusMapRequest) Reset()                    { *m = MinerStatusMapRequest{} }
func (m *MinerStatusMapRequest) String() string            { return proto.CompactTextString(m) }
func (*MinerStatusMapRequest) ProtoMessage()               {}
func (*MinerStatusMapRequest) Descriptor() ([]byte, []int) { return fileDescriptor8, []int{7} }

type SaveRequest struct {
	ImageID string `protobuf:"bytes,1,opt,name=imageID" json:"imageID,omitempty"`
}

func (m *SaveRequest) Reset()                    { *m = SaveRequest{} }
func (m *SaveRequest) String() string            { return proto.CompactTextString(m) }
func (*SaveRequest) ProtoMessage()               {}
func (*SaveRequest) Descriptor() ([]byte, []int) { return fileDescriptor8, []int{8} }

func (m *SaveRequest) GetImageID() string {
	if m != nil {
		return m.ImageID
	}
	return ""
}

func init() {
	proto.RegisterType((*MinerHandshakeRequest)(nil), "sonm.MinerHandshakeRequest")
	proto.RegisterType((*MinerHandshakeReply)(nil), "sonm.MinerHandshakeReply")
	proto.RegisterType((*MinerStartRequest)(nil), "sonm.MinerStartRequest")
	proto.RegisterType((*SocketAddr)(nil), "sonm.SocketAddr")
	proto.RegisterType((*MinerStartReply)(nil), "sonm.MinerStartReply")
	proto.RegisterType((*TaskInfo)(nil), "sonm.TaskInfo")
	proto.RegisterType((*Endpoints)(nil), "sonm.Endpoints")
	proto.RegisterType((*MinerStatusMapRequest)(nil), "sonm.MinerStatusMapRequest")
	proto.RegisterType((*SaveRequest)(nil), "sonm.SaveRequest")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Miner service

type MinerClient interface {
	Ping(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*PingReply, error)
	Info(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*InfoReply, error)
	Handshake(ctx context.Context, in *MinerHandshakeRequest, opts ...grpc.CallOption) (*MinerHandshakeReply, error)
	Save(ctx context.Context, in *SaveRequest, opts ...grpc.CallOption) (Miner_SaveClient, error)
	Load(ctx context.Context, opts ...grpc.CallOption) (Miner_LoadClient, error)
	Start(ctx context.Context, in *MinerStartRequest, opts ...grpc.CallOption) (*MinerStartReply, error)
	Stop(ctx context.Context, in *ID, opts ...grpc.CallOption) (*Empty, error)
	TasksStatus(ctx context.Context, opts ...grpc.CallOption) (Miner_TasksStatusClient, error)
	TaskDetails(ctx context.Context, in *ID, opts ...grpc.CallOption) (*TaskStatusReply, error)
	TaskLogs(ctx context.Context, in *TaskLogsRequest, opts ...grpc.CallOption) (Miner_TaskLogsClient, error)
	DiscoverHub(ctx context.Context, in *DiscoverHubRequest, opts ...grpc.CallOption) (*Empty, error)
}

type minerClient struct {
	cc *grpc.ClientConn
}

func NewMinerClient(cc *grpc.ClientConn) MinerClient {
	return &minerClient{cc}
}

func (c *minerClient) Ping(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*PingReply, error) {
	out := new(PingReply)
	err := grpc.Invoke(ctx, "/sonm.Miner/Ping", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *minerClient) Info(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*InfoReply, error) {
	out := new(InfoReply)
	err := grpc.Invoke(ctx, "/sonm.Miner/Info", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *minerClient) Handshake(ctx context.Context, in *MinerHandshakeRequest, opts ...grpc.CallOption) (*MinerHandshakeReply, error) {
	out := new(MinerHandshakeReply)
	err := grpc.Invoke(ctx, "/sonm.Miner/Handshake", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *minerClient) Save(ctx context.Context, in *SaveRequest, opts ...grpc.CallOption) (Miner_SaveClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Miner_serviceDesc.Streams[0], c.cc, "/sonm.Miner/Save", opts...)
	if err != nil {
		return nil, err
	}
	x := &minerSaveClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Miner_SaveClient interface {
	Recv() (*Chunk, error)
	grpc.ClientStream
}

type minerSaveClient struct {
	grpc.ClientStream
}

func (x *minerSaveClient) Recv() (*Chunk, error) {
	m := new(Chunk)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *minerClient) Load(ctx context.Context, opts ...grpc.CallOption) (Miner_LoadClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Miner_serviceDesc.Streams[1], c.cc, "/sonm.Miner/Load", opts...)
	if err != nil {
		return nil, err
	}
	x := &minerLoadClient{stream}
	return x, nil
}

type Miner_LoadClient interface {
	Send(*Chunk) error
	Recv() (*Progress, error)
	grpc.ClientStream
}

type minerLoadClient struct {
	grpc.ClientStream
}

func (x *minerLoadClient) Send(m *Chunk) error {
	return x.ClientStream.SendMsg(m)
}

func (x *minerLoadClient) Recv() (*Progress, error) {
	m := new(Progress)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *minerClient) Start(ctx context.Context, in *MinerStartRequest, opts ...grpc.CallOption) (*MinerStartReply, error) {
	out := new(MinerStartReply)
	err := grpc.Invoke(ctx, "/sonm.Miner/Start", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *minerClient) Stop(ctx context.Context, in *ID, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := grpc.Invoke(ctx, "/sonm.Miner/Stop", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *minerClient) TasksStatus(ctx context.Context, opts ...grpc.CallOption) (Miner_TasksStatusClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Miner_serviceDesc.Streams[2], c.cc, "/sonm.Miner/TasksStatus", opts...)
	if err != nil {
		return nil, err
	}
	x := &minerTasksStatusClient{stream}
	return x, nil
}

type Miner_TasksStatusClient interface {
	Send(*MinerStatusMapRequest) error
	Recv() (*StatusMapReply, error)
	grpc.ClientStream
}

type minerTasksStatusClient struct {
	grpc.ClientStream
}

func (x *minerTasksStatusClient) Send(m *MinerStatusMapRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *minerTasksStatusClient) Recv() (*StatusMapReply, error) {
	m := new(StatusMapReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *minerClient) TaskDetails(ctx context.Context, in *ID, opts ...grpc.CallOption) (*TaskStatusReply, error) {
	out := new(TaskStatusReply)
	err := grpc.Invoke(ctx, "/sonm.Miner/TaskDetails", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *minerClient) TaskLogs(ctx context.Context, in *TaskLogsRequest, opts ...grpc.CallOption) (Miner_TaskLogsClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Miner_serviceDesc.Streams[3], c.cc, "/sonm.Miner/TaskLogs", opts...)
	if err != nil {
		return nil, err
	}
	x := &minerTaskLogsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Miner_TaskLogsClient interface {
	Recv() (*TaskLogsChunk, error)
	grpc.ClientStream
}

type minerTaskLogsClient struct {
	grpc.ClientStream
}

func (x *minerTaskLogsClient) Recv() (*TaskLogsChunk, error) {
	m := new(TaskLogsChunk)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *minerClient) DiscoverHub(ctx context.Context, in *DiscoverHubRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := grpc.Invoke(ctx, "/sonm.Miner/DiscoverHub", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Miner service

type MinerServer interface {
	Ping(context.Context, *Empty) (*PingReply, error)
	Info(context.Context, *Empty) (*InfoReply, error)
	Handshake(context.Context, *MinerHandshakeRequest) (*MinerHandshakeReply, error)
	Save(*SaveRequest, Miner_SaveServer) error
	Load(Miner_LoadServer) error
	Start(context.Context, *MinerStartRequest) (*MinerStartReply, error)
	Stop(context.Context, *ID) (*Empty, error)
	TasksStatus(Miner_TasksStatusServer) error
	TaskDetails(context.Context, *ID) (*TaskStatusReply, error)
	TaskLogs(*TaskLogsRequest, Miner_TaskLogsServer) error
	DiscoverHub(context.Context, *DiscoverHubRequest) (*Empty, error)
}

func RegisterMinerServer(s *grpc.Server, srv MinerServer) {
	s.RegisterService(&_Miner_serviceDesc, srv)
}

func _Miner_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MinerServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Miner/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MinerServer).Ping(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Miner_Info_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MinerServer).Info(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Miner/Info",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MinerServer).Info(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Miner_Handshake_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MinerHandshakeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MinerServer).Handshake(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Miner/Handshake",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MinerServer).Handshake(ctx, req.(*MinerHandshakeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Miner_Save_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SaveRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(MinerServer).Save(m, &minerSaveServer{stream})
}

type Miner_SaveServer interface {
	Send(*Chunk) error
	grpc.ServerStream
}

type minerSaveServer struct {
	grpc.ServerStream
}

func (x *minerSaveServer) Send(m *Chunk) error {
	return x.ServerStream.SendMsg(m)
}

func _Miner_Load_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(MinerServer).Load(&minerLoadServer{stream})
}

type Miner_LoadServer interface {
	Send(*Progress) error
	Recv() (*Chunk, error)
	grpc.ServerStream
}

type minerLoadServer struct {
	grpc.ServerStream
}

func (x *minerLoadServer) Send(m *Progress) error {
	return x.ServerStream.SendMsg(m)
}

func (x *minerLoadServer) Recv() (*Chunk, error) {
	m := new(Chunk)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Miner_Start_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MinerStartRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MinerServer).Start(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Miner/Start",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MinerServer).Start(ctx, req.(*MinerStartRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Miner_Stop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MinerServer).Stop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Miner/Stop",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MinerServer).Stop(ctx, req.(*ID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Miner_TasksStatus_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(MinerServer).TasksStatus(&minerTasksStatusServer{stream})
}

type Miner_TasksStatusServer interface {
	Send(*StatusMapReply) error
	Recv() (*MinerStatusMapRequest, error)
	grpc.ServerStream
}

type minerTasksStatusServer struct {
	grpc.ServerStream
}

func (x *minerTasksStatusServer) Send(m *StatusMapReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *minerTasksStatusServer) Recv() (*MinerStatusMapRequest, error) {
	m := new(MinerStatusMapRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Miner_TaskDetails_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MinerServer).TaskDetails(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Miner/TaskDetails",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MinerServer).TaskDetails(ctx, req.(*ID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Miner_TaskLogs_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(TaskLogsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(MinerServer).TaskLogs(m, &minerTaskLogsServer{stream})
}

type Miner_TaskLogsServer interface {
	Send(*TaskLogsChunk) error
	grpc.ServerStream
}

type minerTaskLogsServer struct {
	grpc.ServerStream
}

func (x *minerTaskLogsServer) Send(m *TaskLogsChunk) error {
	return x.ServerStream.SendMsg(m)
}

func _Miner_DiscoverHub_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DiscoverHubRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MinerServer).DiscoverHub(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Miner/DiscoverHub",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MinerServer).DiscoverHub(ctx, req.(*DiscoverHubRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Miner_serviceDesc = grpc.ServiceDesc{
	ServiceName: "sonm.Miner",
	HandlerType: (*MinerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _Miner_Ping_Handler,
		},
		{
			MethodName: "Info",
			Handler:    _Miner_Info_Handler,
		},
		{
			MethodName: "Handshake",
			Handler:    _Miner_Handshake_Handler,
		},
		{
			MethodName: "Start",
			Handler:    _Miner_Start_Handler,
		},
		{
			MethodName: "Stop",
			Handler:    _Miner_Stop_Handler,
		},
		{
			MethodName: "TaskDetails",
			Handler:    _Miner_TaskDetails_Handler,
		},
		{
			MethodName: "DiscoverHub",
			Handler:    _Miner_DiscoverHub_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Save",
			Handler:       _Miner_Save_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Load",
			Handler:       _Miner_Load_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "TasksStatus",
			Handler:       _Miner_TasksStatus_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "TaskLogs",
			Handler:       _Miner_TaskLogs_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "miner.proto",
}

func init() { proto.RegisterFile("miner.proto", fileDescriptor8) }

var fileDescriptor8 = []byte{
	// 830 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x55, 0xdb, 0x8e, 0xdb, 0x44,
	0x18, 0x8e, 0x73, 0x68, 0xe2, 0xdf, 0x7b, 0x68, 0xa7, 0x5d, 0xd5, 0x98, 0x0a, 0x22, 0xab, 0xd0,
	0x00, 0x52, 0xb4, 0x04, 0xb4, 0x82, 0xd2, 0x9b, 0xd2, 0x2c, 0xea, 0xaa, 0x5d, 0x1a, 0x4d, 0xf6,
	0x05, 0x26, 0xf1, 0x90, 0x1d, 0x92, 0xcc, 0x98, 0x99, 0x71, 0x24, 0xbf, 0x03, 0x4f, 0x04, 0x17,
	0xbc, 0x1a, 0x9a, 0x83, 0xd7, 0xce, 0x36, 0xbd, 0xca, 0xfc, 0xe7, 0xef, 0xfb, 0xfe, 0xf1, 0x04,
	0xa2, 0x2d, 0xe3, 0x54, 0x8e, 0x73, 0x29, 0xb4, 0x40, 0x5d, 0x25, 0xf8, 0x36, 0x39, 0x65, 0xdc,
	0xfc, 0x72, 0x46, 0x9c, 0x3b, 0x41, 0x4b, 0x92, 0x93, 0x05, 0xdb, 0x30, 0xcd, 0xa8, 0xf2, 0xbe,
	0x90, 0x13, 0xed, 0x8e, 0xe9, 0x07, 0x38, 0xbb, 0x36, 0x4d, 0xde, 0x12, 0x9e, 0xa9, 0x5b, 0xb2,
	0xa6, 0x98, 0xfe, 0x55, 0x50, 0xa5, 0xd1, 0x43, 0xe8, 0xdc, 0x16, 0x8b, 0x38, 0x18, 0x06, 0xa3,
	0x10, 0x9b, 0x23, 0x7a, 0x0e, 0x3d, 0x4d, 0xd4, 0x5a, 0xc5, 0xed, 0x61, 0x67, 0x14, 0x4d, 0x4e,
	0xc6, 0x66, 0xd0, 0xf8, 0x86, 0xa8, 0xf5, 0x15, 0xff, 0x43, 0x60, 0x17, 0x4c, 0xff, 0x0e, 0xe0,
	0xf1, 0xfd, 0x8e, 0xf9, 0xa6, 0x44, 0x4f, 0xa0, 0x67, 0xd1, 0xfa, 0x8e, 0xce, 0x40, 0x17, 0x70,
	0xd4, 0xc4, 0x17, 0xb7, 0x87, 0xc1, 0x28, 0x9a, 0x20, 0xd7, 0xfa, 0x4d, 0x23, 0x82, 0xf7, 0xf2,
	0xd0, 0x0b, 0xe8, 0x73, 0xa2, 0x6f, 0xca, 0x9c, 0xc6, 0x9d, 0x61, 0x30, 0x3a, 0x99, 0x1c, 0xbb,
	0x92, 0xdf, 0x5f, 0xdf, 0x18, 0x27, 0xae, 0xa2, 0xe9, 0xbf, 0x1d, 0x78, 0x64, 0xe1, 0xcc, 0x35,
	0x91, 0xba, 0x22, 0x77, 0x02, 0x6d, 0x96, 0x79, 0x24, 0x6d, 0x96, 0xa1, 0x04, 0x06, 0x92, 0xae,
	0x98, 0xd2, 0xb2, 0xb4, 0x10, 0x42, 0x7c, 0x67, 0x1b, 0xe0, 0x6c, 0x4b, 0x56, 0x6e, 0x50, 0x88,
	0x9d, 0x81, 0x10, 0x74, 0x49, 0xa1, 0x6f, 0xe3, 0xae, 0x75, 0xda, 0x33, 0xfa, 0x15, 0x8e, 0x25,
	0x55, 0x66, 0xce, 0x4c, 0x6c, 0xd8, 0xb2, 0x8c, 0x7b, 0x96, 0xcd, 0x33, 0xcf, 0x46, 0x70, 0x4d,
	0x0c, 0x12, 0xdc, 0xcc, 0xc1, 0xfb, 0x25, 0xe8, 0x15, 0x84, 0x92, 0x2a, 0x51, 0xc8, 0x25, 0x55,
	0xf1, 0x03, 0x5b, 0xff, 0x45, 0x2d, 0x34, 0xf6, 0x21, 0xc3, 0x83, 0x49, 0xba, 0xa5, 0x5c, 0x2b,
	0x5c, 0x17, 0xa0, 0xe7, 0x70, 0x3c, 0x2b, 0x16, 0x1b, 0xb6, 0x7c, 0x47, 0xcb, 0x29, 0xd1, 0x24,
	0xee, 0x5b, 0x78, 0xfb, 0x4e, 0x94, 0xc2, 0xd1, 0x52, 0x6c, 0xb7, 0x4c, 0x7f, 0xe0, 0x73, 0x2d,
	0xf2, 0x78, 0x30, 0x0c, 0x46, 0x03, 0xbc, 0xe7, 0x43, 0x13, 0xe8, 0x50, 0xbe, 0x8b, 0x43, 0xbb,
	0xea, 0xa1, 0x43, 0xf0, 0x91, 0x8e, 0xe3, 0x4b, 0xbe, 0xbb, 0xe4, 0x5a, 0x96, 0xd8, 0x24, 0xa3,
	0x18, 0xfa, 0x42, 0x66, 0x54, 0x5e, 0x65, 0x31, 0xd8, 0xb9, 0x95, 0x99, 0x5c, 0xc0, 0xa0, 0x4a,
	0x35, 0x17, 0x6b, 0x4d, 0xcb, 0xea, 0x62, 0xad, 0xa9, 0x55, 0x78, 0x47, 0x36, 0x05, 0xf5, 0xd2,
	0x3b, 0xe3, 0x65, 0xfb, 0xa7, 0x20, 0xfd, 0x11, 0x60, 0x2e, 0x96, 0x6b, 0xaa, 0x5f, 0x67, 0x99,
	0xb4, 0x9a, 0x67, 0x59, 0x75, 0x83, 0xec, 0xd9, 0xf8, 0x72, 0x21, 0xb5, 0x2d, 0x3d, 0xc6, 0xf6,
	0x9c, 0xfe, 0x13, 0xc0, 0x69, 0x13, 0xab, 0xb9, 0x7e, 0xcf, 0x20, 0x5c, 0x56, 0x0b, 0xf0, 0x0d,
	0x6a, 0x07, 0x7a, 0x05, 0x7d, 0x53, 0x79, 0x4d, 0x72, 0x7f, 0xb9, 0xd3, 0x8f, 0x19, 0xe7, 0x9b,
	0x72, 0x3c, 0x73, 0x49, 0x8e, 0x73, 0x55, 0x92, 0xbc, 0x83, 0xa3, 0x66, 0xe0, 0x00, 0xc3, 0xaf,
	0x9a, 0x0c, 0xa3, 0xc9, 0xa9, 0xeb, 0x7e, 0xc9, 0xb3, 0x5c, 0x30, 0xb3, 0xc2, 0x06, 0xe5, 0x3f,
	0x61, 0x50, 0x7d, 0x52, 0xe8, 0x7b, 0xe8, 0x4b, 0xa7, 0xb4, 0x6d, 0x16, 0x4d, 0x9e, 0x7e, 0x62,
	0x11, 0xb8, 0xca, 0x43, 0xdf, 0x41, 0x4f, 0x1a, 0xa8, 0x7e, 0xd2, 0xd9, 0x41, 0x1e, 0xd8, 0xe5,
	0xa4, 0xbf, 0x40, 0x78, 0x87, 0x01, 0x8d, 0x21, 0xa4, 0x95, 0x11, 0x07, 0x56, 0x85, 0x87, 0xae,
	0xba, 0x5e, 0x01, 0xae, 0x53, 0xd2, 0xa7, 0xfe, 0xe5, 0x98, 0x6b, 0xa2, 0x0b, 0x75, 0x4d, 0x72,
	0x8f, 0x25, 0x7d, 0x01, 0xd1, 0x9c, 0xec, 0xee, 0x1e, 0x92, 0x18, 0xfa, 0xf6, 0x93, 0xb9, 0x9a,
	0x7a, 0x45, 0x2a, 0x73, 0xf2, 0x5f, 0x17, 0x7a, 0xb6, 0x05, 0xfa, 0x1a, 0xba, 0x33, 0xc6, 0x57,
	0x28, 0xf2, 0xc2, 0x6c, 0x73, 0x5d, 0x26, 0x5e, 0x25, 0x13, 0xb0, 0xa8, 0xd3, 0x96, 0xc9, 0xb3,
	0xc2, 0x1c, 0xca, 0xb3, 0x8f, 0x90, 0xcf, 0xbb, 0x84, 0xf0, 0xee, 0xf9, 0x41, 0x9f, 0x37, 0x34,
	0xb8, 0xff, 0xcc, 0x25, 0x9f, 0x1d, 0x0e, 0xba, 0x36, 0xdf, 0x42, 0xd7, 0x30, 0x41, 0x8f, 0xbc,
	0x0e, 0x35, 0xab, 0xc4, 0x23, 0x78, 0x73, 0x5b, 0xf0, 0x75, 0xda, 0x3a, 0x0f, 0xd0, 0x37, 0xd0,
	0x7d, 0x2f, 0x48, 0x86, 0x9a, 0x81, 0xc4, 0xbf, 0x91, 0x33, 0x29, 0x56, 0x92, 0x2a, 0x95, 0xb6,
	0x46, 0xc1, 0x79, 0x80, 0x7e, 0x86, 0x9e, 0xdd, 0x05, 0xfa, 0xd4, 0x3a, 0x93, 0xc3, 0x6b, 0x4b,
	0x5b, 0xe8, 0x4b, 0xe8, 0xda, 0xcf, 0x73, 0xe0, 0x39, 0x4f, 0x93, 0xa6, 0x14, 0x69, 0x0b, 0xfd,
	0x06, 0x91, 0xb9, 0x3e, 0xca, 0x6d, 0x65, 0x8f, 0xfb, 0xfd, 0x45, 0x25, 0x4f, 0x3c, 0xad, 0xda,
	0x6f, 0x87, 0x58, 0x8c, 0xe7, 0xae, 0xcf, 0x94, 0x6a, 0xc2, 0x36, 0xaa, 0x31, 0xef, 0xac, 0x7e,
	0x8d, 0x5c, 0x61, 0x05, 0xed, 0xa5, 0xbb, 0xb8, 0xef, 0xc5, 0x4a, 0xa1, 0x46, 0x92, 0xb1, 0xab,
	0x81, 0x8f, 0xf7, 0xdd, 0xb5, 0x78, 0x17, 0x10, 0x4d, 0x99, 0x5a, 0x8a, 0x1d, 0x95, 0x6f, 0x8b,
	0x05, 0x8a, 0x5d, 0x5e, 0xc3, 0x75, 0x4f, 0x76, 0xcf, 0x76, 0xf1, 0xc0, 0xfe, 0x89, 0xfd, 0xf0,
	0x7f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x13, 0x7c, 0x8d, 0xe1, 0x09, 0x07, 0x00, 0x00,
}

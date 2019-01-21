// Code generated by protoc-gen-go. DO NOT EDIT.
// source: relay.proto

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

type PeerType int32

const (
	PeerType_SERVER   PeerType = 0
	PeerType_CLIENT   PeerType = 1
	PeerType_DISCOVER PeerType = 2
)

var PeerType_name = map[int32]string{
	0: "SERVER",
	1: "CLIENT",
	2: "DISCOVER",
}
var PeerType_value = map[string]int32{
	"SERVER":   0,
	"CLIENT":   1,
	"DISCOVER": 2,
}

func (x PeerType) String() string {
	return proto.EnumName(PeerType_name, int32(x))
}
func (PeerType) EnumDescriptor() ([]byte, []int) { return fileDescriptor13, []int{0} }

type HandshakeRequest struct {
	// PeerType describes a peer's source.
	PeerType PeerType `protobuf:"varint,1,opt,name=peerType,enum=sonm.PeerType" json:"peerType,omitempty"`
	// Addr represents a common Ethereum address both peers are connecting
	// around.
	// In case of servers it's their own id. Must be signed. In case of
	// clients - it's the target server id.
	//
	// In case of discovery requests this field has special meaning.
	// Both client and server must discover the same relay server to be able to
	// meet each other. At this stage there is no parameter verification.
	// It is done in the Handshake method.
	Addr []byte `protobuf:"bytes,2,opt,name=addr,proto3" json:"addr,omitempty"`
	// Signature for ETH address.
	// Should be empty for clients.
	Sign []byte `protobuf:"bytes,3,opt,name=sign,proto3" json:"sign,omitempty"`
	// Optional connection id.
	// It is used when a client wants to connect to a specific server avoiding
	// random select.
	// Should be empty for servers.
	UUID string `protobuf:"bytes,4,opt,name=UUID" json:"UUID,omitempty"`
	// Protocol describes the network protocol the peer wants to publish or to
	// resolve.
	Protocol string `protobuf:"bytes,5,opt,name=protocol" json:"protocol,omitempty"`
}

func (m *HandshakeRequest) Reset()                    { *m = HandshakeRequest{} }
func (m *HandshakeRequest) String() string            { return proto.CompactTextString(m) }
func (*HandshakeRequest) ProtoMessage()               {}
func (*HandshakeRequest) Descriptor() ([]byte, []int) { return fileDescriptor13, []int{0} }

func (m *HandshakeRequest) GetPeerType() PeerType {
	if m != nil {
		return m.PeerType
	}
	return PeerType_SERVER
}

func (m *HandshakeRequest) GetAddr() []byte {
	if m != nil {
		return m.Addr
	}
	return nil
}

func (m *HandshakeRequest) GetSign() []byte {
	if m != nil {
		return m.Sign
	}
	return nil
}

func (m *HandshakeRequest) GetUUID() string {
	if m != nil {
		return m.UUID
	}
	return ""
}

func (m *HandshakeRequest) GetProtocol() string {
	if m != nil {
		return m.Protocol
	}
	return ""
}

type DiscoverResponse struct {
	// Addr represents network address in form "host:port".
	Addr string `protobuf:"bytes,1,opt,name=addr" json:"addr,omitempty"`
}

func (m *DiscoverResponse) Reset()                    { *m = DiscoverResponse{} }
func (m *DiscoverResponse) String() string            { return proto.CompactTextString(m) }
func (*DiscoverResponse) ProtoMessage()               {}
func (*DiscoverResponse) Descriptor() ([]byte, []int) { return fileDescriptor13, []int{1} }

func (m *DiscoverResponse) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

type HandshakeResponse struct {
	// Error describes an error number.
	// Zero value means that there is no error.
	Error int32 `protobuf:"varint,1,opt,name=error" json:"error,omitempty"`
	// Description describes an error above.
	Description string `protobuf:"bytes,2,opt,name=description" json:"description,omitempty"`
}

func (m *HandshakeResponse) Reset()                    { *m = HandshakeResponse{} }
func (m *HandshakeResponse) String() string            { return proto.CompactTextString(m) }
func (*HandshakeResponse) ProtoMessage()               {}
func (*HandshakeResponse) Descriptor() ([]byte, []int) { return fileDescriptor13, []int{2} }

func (m *HandshakeResponse) GetError() int32 {
	if m != nil {
		return m.Error
	}
	return 0
}

func (m *HandshakeResponse) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

type RelayClusterReply struct {
	Members []string `protobuf:"bytes,1,rep,name=members" json:"members,omitempty"`
}

func (m *RelayClusterReply) Reset()                    { *m = RelayClusterReply{} }
func (m *RelayClusterReply) String() string            { return proto.CompactTextString(m) }
func (*RelayClusterReply) ProtoMessage()               {}
func (*RelayClusterReply) Descriptor() ([]byte, []int) { return fileDescriptor13, []int{3} }

func (m *RelayClusterReply) GetMembers() []string {
	if m != nil {
		return m.Members
	}
	return nil
}

type RelayMetrics struct {
	ConnCurrent uint64                 `protobuf:"varint,1,opt,name=connCurrent" json:"connCurrent,omitempty"`
	Net         map[string]*NetMetrics `protobuf:"bytes,2,rep,name=net" json:"net,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Uptime      uint64                 `protobuf:"varint,3,opt,name=uptime" json:"uptime,omitempty"`
}

func (m *RelayMetrics) Reset()                    { *m = RelayMetrics{} }
func (m *RelayMetrics) String() string            { return proto.CompactTextString(m) }
func (*RelayMetrics) ProtoMessage()               {}
func (*RelayMetrics) Descriptor() ([]byte, []int) { return fileDescriptor13, []int{4} }

func (m *RelayMetrics) GetConnCurrent() uint64 {
	if m != nil {
		return m.ConnCurrent
	}
	return 0
}

func (m *RelayMetrics) GetNet() map[string]*NetMetrics {
	if m != nil {
		return m.Net
	}
	return nil
}

func (m *RelayMetrics) GetUptime() uint64 {
	if m != nil {
		return m.Uptime
	}
	return 0
}

type NetMetrics struct {
	TxBytes uint64 `protobuf:"varint,1,opt,name=txBytes" json:"txBytes,omitempty"`
	RxBytes uint64 `protobuf:"varint,2,opt,name=rxBytes" json:"rxBytes,omitempty"`
}

func (m *NetMetrics) Reset()                    { *m = NetMetrics{} }
func (m *NetMetrics) String() string            { return proto.CompactTextString(m) }
func (*NetMetrics) ProtoMessage()               {}
func (*NetMetrics) Descriptor() ([]byte, []int) { return fileDescriptor13, []int{5} }

func (m *NetMetrics) GetTxBytes() uint64 {
	if m != nil {
		return m.TxBytes
	}
	return 0
}

func (m *NetMetrics) GetRxBytes() uint64 {
	if m != nil {
		return m.RxBytes
	}
	return 0
}

// RelayInfo is a response returned from Info handle.
type RelayInfo struct {
	State map[string]*RelayMeeting `protobuf:"bytes,1,rep,name=state" json:"state,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *RelayInfo) Reset()                    { *m = RelayInfo{} }
func (m *RelayInfo) String() string            { return proto.CompactTextString(m) }
func (*RelayInfo) ProtoMessage()               {}
func (*RelayInfo) Descriptor() ([]byte, []int) { return fileDescriptor13, []int{6} }

func (m *RelayInfo) GetState() map[string]*RelayMeeting {
	if m != nil {
		return m.State
	}
	return nil
}

// RelayMeeting represents relay point.
type RelayMeeting struct {
	Servers map[string]*Addr `protobuf:"bytes,2,rep,name=servers" json:"servers,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *RelayMeeting) Reset()                    { *m = RelayMeeting{} }
func (m *RelayMeeting) String() string            { return proto.CompactTextString(m) }
func (*RelayMeeting) ProtoMessage()               {}
func (*RelayMeeting) Descriptor() ([]byte, []int) { return fileDescriptor13, []int{7} }

func (m *RelayMeeting) GetServers() map[string]*Addr {
	if m != nil {
		return m.Servers
	}
	return nil
}

func init() {
	proto.RegisterType((*HandshakeRequest)(nil), "sonm.HandshakeRequest")
	proto.RegisterType((*DiscoverResponse)(nil), "sonm.DiscoverResponse")
	proto.RegisterType((*HandshakeResponse)(nil), "sonm.HandshakeResponse")
	proto.RegisterType((*RelayClusterReply)(nil), "sonm.RelayClusterReply")
	proto.RegisterType((*RelayMetrics)(nil), "sonm.RelayMetrics")
	proto.RegisterType((*NetMetrics)(nil), "sonm.NetMetrics")
	proto.RegisterType((*RelayInfo)(nil), "sonm.RelayInfo")
	proto.RegisterType((*RelayMeeting)(nil), "sonm.RelayMeeting")
	proto.RegisterEnum("sonm.PeerType", PeerType_name, PeerType_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Relay service

type RelayClient interface {
	Cluster(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*RelayClusterReply, error)
	Metrics(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*RelayMetrics, error)
	Info(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*RelayInfo, error)
}

type relayClient struct {
	cc *grpc.ClientConn
}

func NewRelayClient(cc *grpc.ClientConn) RelayClient {
	return &relayClient{cc}
}

func (c *relayClient) Cluster(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*RelayClusterReply, error) {
	out := new(RelayClusterReply)
	err := grpc.Invoke(ctx, "/sonm.Relay/Cluster", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *relayClient) Metrics(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*RelayMetrics, error) {
	out := new(RelayMetrics)
	err := grpc.Invoke(ctx, "/sonm.Relay/Metrics", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *relayClient) Info(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*RelayInfo, error) {
	out := new(RelayInfo)
	err := grpc.Invoke(ctx, "/sonm.Relay/Info", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Relay service

type RelayServer interface {
	Cluster(context.Context, *Empty) (*RelayClusterReply, error)
	Metrics(context.Context, *Empty) (*RelayMetrics, error)
	Info(context.Context, *Empty) (*RelayInfo, error)
}

func RegisterRelayServer(s *grpc.Server, srv RelayServer) {
	s.RegisterService(&_Relay_serviceDesc, srv)
}

func _Relay_Cluster_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RelayServer).Cluster(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Relay/Cluster",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RelayServer).Cluster(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Relay_Metrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RelayServer).Metrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Relay/Metrics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RelayServer).Metrics(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Relay_Info_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RelayServer).Info(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Relay/Info",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RelayServer).Info(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _Relay_serviceDesc = grpc.ServiceDesc{
	ServiceName: "sonm.Relay",
	HandlerType: (*RelayServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Cluster",
			Handler:    _Relay_Cluster_Handler,
		},
		{
			MethodName: "Metrics",
			Handler:    _Relay_Metrics_Handler,
		},
		{
			MethodName: "Info",
			Handler:    _Relay_Info_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "relay.proto",
}

func init() { proto.RegisterFile("relay.proto", fileDescriptor13) }

var fileDescriptor13 = []byte{
	// 582 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x53, 0x4f, 0x6f, 0xd3, 0x4e,
	0x10, 0xcd, 0xe6, 0x4f, 0x93, 0x8c, 0xa3, 0xd6, 0x5d, 0xfd, 0xf4, 0xc3, 0x32, 0x07, 0x2c, 0x1f,
	0x2a, 0xab, 0xa2, 0x56, 0x09, 0x17, 0xe0, 0x04, 0xa4, 0x41, 0x8d, 0x28, 0x05, 0x6d, 0x5a, 0xee,
	0x6e, 0x3c, 0x14, 0xab, 0xc9, 0xda, 0xec, 0x6e, 0x2a, 0xfc, 0x19, 0xb8, 0xc0, 0x85, 0xef, 0xc3,
	0x37, 0x43, 0xbb, 0x6b, 0x37, 0x0e, 0xed, 0x29, 0x33, 0xf3, 0x5e, 0x66, 0xde, 0x9b, 0x59, 0x83,
	0x23, 0x70, 0x99, 0x94, 0x71, 0x21, 0x72, 0x95, 0xd3, 0xae, 0xcc, 0xf9, 0xca, 0xdf, 0xcb, 0xb8,
	0xfe, 0xe5, 0x59, 0x62, 0xcb, 0xfe, 0x90, 0xa3, 0xb2, 0x61, 0xf8, 0x9b, 0x80, 0x7b, 0x9a, 0xf0,
	0x54, 0x7e, 0x4d, 0x6e, 0x90, 0xe1, 0xb7, 0x35, 0x4a, 0x45, 0x0f, 0x61, 0x50, 0x20, 0x8a, 0x8b,
	0xb2, 0x40, 0x8f, 0x04, 0x24, 0xda, 0x1d, 0xef, 0xc6, 0xba, 0x43, 0xfc, 0xa9, 0xaa, 0xb2, 0x3b,
	0x9c, 0x52, 0xe8, 0x26, 0x69, 0x2a, 0xbc, 0x76, 0x40, 0xa2, 0x11, 0x33, 0xb1, 0xae, 0xc9, 0xec,
	0x9a, 0x7b, 0x1d, 0x5b, 0xd3, 0xb1, 0xae, 0x5d, 0x5e, 0xce, 0x4e, 0xbc, 0x6e, 0x40, 0xa2, 0x21,
	0x33, 0x31, 0xf5, 0x61, 0x60, 0x54, 0x2c, 0xf2, 0xa5, 0xd7, 0x33, 0xf5, 0xbb, 0x3c, 0x3c, 0x00,
	0xf7, 0x24, 0x93, 0x8b, 0xfc, 0x16, 0x05, 0x43, 0x59, 0xe4, 0x5c, 0x6e, 0x66, 0x11, 0xdb, 0x43,
	0xc7, 0xe1, 0x7b, 0xd8, 0x6f, 0xe8, 0xaf, 0x88, 0xff, 0x41, 0x0f, 0x85, 0xc8, 0x2d, 0xb3, 0xc7,
	0x6c, 0x42, 0x03, 0x70, 0x52, 0x94, 0x0b, 0x91, 0x15, 0x2a, 0xcb, 0xb9, 0x51, 0x3c, 0x64, 0xcd,
	0x52, 0x78, 0x04, 0xfb, 0x4c, 0xaf, 0x6f, 0xb2, 0x5c, 0x4b, 0xa5, 0x07, 0x17, 0xcb, 0x92, 0x7a,
	0xd0, 0x5f, 0xe1, 0xea, 0x0a, 0x85, 0xf4, 0x48, 0xd0, 0x89, 0x86, 0xac, 0x4e, 0xc3, 0x3f, 0x04,
	0x46, 0x86, 0xff, 0x01, 0x95, 0xc8, 0x16, 0x52, 0x4f, 0x58, 0xe4, 0x9c, 0x4f, 0xd6, 0x42, 0x20,
	0x57, 0x66, 0x7a, 0x97, 0x35, 0x4b, 0xf4, 0x08, 0x3a, 0x1c, 0x95, 0xd7, 0x0e, 0x3a, 0x91, 0x33,
	0x7e, 0x6c, 0xb7, 0xda, 0x6c, 0x11, 0x9f, 0xa3, 0x9a, 0x72, 0x25, 0x4a, 0xa6, 0x79, 0xf4, 0x7f,
	0xd8, 0x59, 0x17, 0x2a, 0x5b, 0xa1, 0xd9, 0x65, 0x97, 0x55, 0x99, 0x7f, 0x0a, 0x83, 0x9a, 0x48,
	0x5d, 0xe8, 0xdc, 0x60, 0x59, 0x2d, 0x45, 0x87, 0xf4, 0x00, 0x7a, 0xb7, 0xc9, 0x72, 0x8d, 0xc6,
	0xa2, 0x33, 0x76, 0xed, 0x98, 0x73, 0x54, 0xd5, 0x10, 0x66, 0xe1, 0x57, 0xed, 0x17, 0x24, 0x7c,
	0x0d, 0xb0, 0x01, 0xb4, 0x57, 0xf5, 0xfd, 0x6d, 0xa9, 0x50, 0x56, 0xe2, 0xeb, 0x54, 0x23, 0xa2,
	0x42, 0xda, 0x16, 0xa9, 0xd2, 0xf0, 0x07, 0x81, 0xa1, 0xb1, 0x30, 0xe3, 0x5f, 0x72, 0x7a, 0x0c,
	0x3d, 0xa9, 0x12, 0x85, 0x66, 0x57, 0xce, 0xd8, 0x6f, 0x58, 0xd4, 0x78, 0x3c, 0xd7, 0xa0, 0x75,
	0x68, 0x89, 0xfe, 0x19, 0xc0, 0xa6, 0xf8, 0x80, 0x9b, 0x68, 0xdb, 0x0d, 0xdd, 0x5a, 0x1a, 0xaa,
	0x8c, 0x5f, 0x37, 0xfd, 0xfc, 0xda, 0xdc, 0xc4, 0x60, 0xf4, 0x25, 0xf4, 0x25, 0x8a, 0x5b, 0x7d,
	0x3e, 0xbb, 0xf5, 0x27, 0xf7, 0x1b, 0xc4, 0x73, 0xcb, 0xb0, 0xba, 0x6a, 0xbe, 0xff, 0x0e, 0x46,
	0x4d, 0xe0, 0x01, 0x6d, 0xc1, 0xb6, 0x36, 0xb0, 0xad, 0xdf, 0xa4, 0xa9, 0x68, 0x68, 0x3a, 0x3c,
	0x86, 0x41, 0xfd, 0xe5, 0x50, 0x80, 0x9d, 0xf9, 0x94, 0x7d, 0x9e, 0x32, 0xb7, 0xa5, 0xe3, 0xc9,
	0xd9, 0x6c, 0x7a, 0x7e, 0xe1, 0x12, 0x3a, 0x82, 0xc1, 0xc9, 0x6c, 0x3e, 0xf9, 0xa8, 0x91, 0xf6,
	0xf8, 0x27, 0x81, 0x9e, 0x11, 0x48, 0x9f, 0x41, 0xbf, 0x7a, 0x8d, 0xd4, 0xb1, 0xdd, 0xa7, 0xab,
	0x42, 0x95, 0xfe, 0xa3, 0x86, 0x8b, 0xe6, 0x73, 0x0d, 0x5b, 0xf4, 0x29, 0xf4, 0xeb, 0x7b, 0x6e,
	0xfd, 0x85, 0xde, 0x7f, 0x6e, 0x61, 0x8b, 0x1e, 0x40, 0xd7, 0x1c, 0x6e, 0x8b, 0xba, 0xf7, 0xcf,
	0xd9, 0xc2, 0xd6, 0xd5, 0x8e, 0xf9, 0x34, 0x9f, 0xff, 0x0d, 0x00, 0x00, 0xff, 0xff, 0x0c, 0xd5,
	0x39, 0x15, 0x61, 0x04, 0x00, 0x00,
}

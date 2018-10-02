// Code generated by protoc-gen-go. DO NOT EDIT.
// source: rendezvous.proto

package sonm

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// grpccmd imports
import (
	"io"

	"github.com/sonm-io/core/util/xcode"
	"github.com/spf13/cobra"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// ConnectRequest describres a connection request to a remote target, possibly
// located under the NAT.
type ConnectRequest struct {
	// ID describes an unique ID of a target. Mainly it's an ETH address.
	ID []byte `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	// Protocol describes network protocol the peer wants to resolve.
	Protocol string `protobuf:"bytes,2,opt,name=protocol" json:"protocol,omitempty"`
	// PrivateAddrs describes source private addresses.
	PrivateAddrs []*Addr `protobuf:"bytes,3,rep,name=privateAddrs" json:"privateAddrs,omitempty"`
}

func (m *ConnectRequest) Reset()                    { *m = ConnectRequest{} }
func (m *ConnectRequest) String() string            { return proto.CompactTextString(m) }
func (*ConnectRequest) ProtoMessage()               {}
func (*ConnectRequest) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{0} }

func (m *ConnectRequest) GetID() []byte {
	if m != nil {
		return m.ID
	}
	return nil
}

func (m *ConnectRequest) GetProtocol() string {
	if m != nil {
		return m.Protocol
	}
	return ""
}

func (m *ConnectRequest) GetPrivateAddrs() []*Addr {
	if m != nil {
		return m.PrivateAddrs
	}
	return nil
}

type PublishRequest struct {
	// Protocol describes network protocol the peer wants to publish.
	Protocol string `protobuf:"bytes,1,opt,name=protocol" json:"protocol,omitempty"`
	// PrivateAddrs describes source private addresses.
	PrivateAddrs []*Addr `protobuf:"bytes,2,rep,name=privateAddrs" json:"privateAddrs,omitempty"`
}

func (m *PublishRequest) Reset()                    { *m = PublishRequest{} }
func (m *PublishRequest) String() string            { return proto.CompactTextString(m) }
func (*PublishRequest) ProtoMessage()               {}
func (*PublishRequest) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{1} }

func (m *PublishRequest) GetProtocol() string {
	if m != nil {
		return m.Protocol
	}
	return ""
}

func (m *PublishRequest) GetPrivateAddrs() []*Addr {
	if m != nil {
		return m.PrivateAddrs
	}
	return nil
}

// RendezvousReply describes a rendezvous point reply.
type RendezvousReply struct {
	// PublicAddr is a public network address of a target.
	PublicAddr *Addr `protobuf:"bytes,1,opt,name=publicAddr" json:"publicAddr,omitempty"`
	// PrivateAddrs describes private network addresses of a target.
	//
	// These addresses should be used to perform an initial connection
	// attempt for cases where both peers are located under the same NAT.
	PrivateAddrs []*Addr `protobuf:"bytes,2,rep,name=privateAddrs" json:"privateAddrs,omitempty"`
}

func (m *RendezvousReply) Reset()                    { *m = RendezvousReply{} }
func (m *RendezvousReply) String() string            { return proto.CompactTextString(m) }
func (*RendezvousReply) ProtoMessage()               {}
func (*RendezvousReply) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{2} }

func (m *RendezvousReply) GetPublicAddr() *Addr {
	if m != nil {
		return m.PublicAddr
	}
	return nil
}

func (m *RendezvousReply) GetPrivateAddrs() []*Addr {
	if m != nil {
		return m.PrivateAddrs
	}
	return nil
}

// RendezvousState is a response returned from Info handle.
type RendezvousState struct {
	State map[string]*RendezvousMeeting `protobuf:"bytes,1,rep,name=state" json:"state,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *RendezvousState) Reset()                    { *m = RendezvousState{} }
func (m *RendezvousState) String() string            { return proto.CompactTextString(m) }
func (*RendezvousState) ProtoMessage()               {}
func (*RendezvousState) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{3} }

func (m *RendezvousState) GetState() map[string]*RendezvousMeeting {
	if m != nil {
		return m.State
	}
	return nil
}

// RendezvousMeeting represents rendezvous point.
type RendezvousMeeting struct {
	Clients map[string]*RendezvousReply `protobuf:"bytes,1,rep,name=clients" json:"clients,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Servers map[string]*RendezvousReply `protobuf:"bytes,2,rep,name=servers" json:"servers,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *RendezvousMeeting) Reset()                    { *m = RendezvousMeeting{} }
func (m *RendezvousMeeting) String() string            { return proto.CompactTextString(m) }
func (*RendezvousMeeting) ProtoMessage()               {}
func (*RendezvousMeeting) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{4} }

func (m *RendezvousMeeting) GetClients() map[string]*RendezvousReply {
	if m != nil {
		return m.Clients
	}
	return nil
}

func (m *RendezvousMeeting) GetServers() map[string]*RendezvousReply {
	if m != nil {
		return m.Servers
	}
	return nil
}

type ResolveMetaReply struct {
	IDs []string `protobuf:"bytes,1,rep,name=IDs" json:"IDs,omitempty"`
}

func (m *ResolveMetaReply) Reset()                    { *m = ResolveMetaReply{} }
func (m *ResolveMetaReply) String() string            { return proto.CompactTextString(m) }
func (*ResolveMetaReply) ProtoMessage()               {}
func (*ResolveMetaReply) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{5} }

func (m *ResolveMetaReply) GetIDs() []string {
	if m != nil {
		return m.IDs
	}
	return nil
}

func init() {
	proto.RegisterType((*ConnectRequest)(nil), "sonm.ConnectRequest")
	proto.RegisterType((*PublishRequest)(nil), "sonm.PublishRequest")
	proto.RegisterType((*RendezvousReply)(nil), "sonm.RendezvousReply")
	proto.RegisterType((*RendezvousState)(nil), "sonm.RendezvousState")
	proto.RegisterType((*RendezvousMeeting)(nil), "sonm.RendezvousMeeting")
	proto.RegisterType((*ResolveMetaReply)(nil), "sonm.ResolveMetaReply")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Rendezvous service

type RendezvousClient interface {
	// Resolve resolves the remote peer addresses.
	Resolve(ctx context.Context, in *ConnectRequest, opts ...grpc.CallOption) (*RendezvousReply, error)
	// ResolveAll resolves remote servers using the specified peer ID,
	// returning the list of unique id's of these servets.
	//
	// Such UUIDs can be used for establishing aimed connection with all
	// servers under the same ID without randomization games.
	ResolveAll(ctx context.Context, in *ID, opts ...grpc.CallOption) (*ResolveMetaReply, error)
	// Publish allows to publish the caller server's endpoints.
	//
	// While published the server can be located using the ID extracted from
	// the transport credentials.
	Publish(ctx context.Context, in *PublishRequest, opts ...grpc.CallOption) (*RendezvousReply, error)
	// Info returns server's internal state.
	Info(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*RendezvousState, error)
}

type rendezvousClient struct {
	cc *grpc.ClientConn
}

func NewRendezvousClient(cc *grpc.ClientConn) RendezvousClient {
	return &rendezvousClient{cc}
}

func (c *rendezvousClient) Resolve(ctx context.Context, in *ConnectRequest, opts ...grpc.CallOption) (*RendezvousReply, error) {
	out := new(RendezvousReply)
	err := grpc.Invoke(ctx, "/sonm.Rendezvous/Resolve", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rendezvousClient) ResolveAll(ctx context.Context, in *ID, opts ...grpc.CallOption) (*ResolveMetaReply, error) {
	out := new(ResolveMetaReply)
	err := grpc.Invoke(ctx, "/sonm.Rendezvous/ResolveAll", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rendezvousClient) Publish(ctx context.Context, in *PublishRequest, opts ...grpc.CallOption) (*RendezvousReply, error) {
	out := new(RendezvousReply)
	err := grpc.Invoke(ctx, "/sonm.Rendezvous/Publish", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rendezvousClient) Info(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*RendezvousState, error) {
	out := new(RendezvousState)
	err := grpc.Invoke(ctx, "/sonm.Rendezvous/Info", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Rendezvous service

type RendezvousServer interface {
	// Resolve resolves the remote peer addresses.
	Resolve(context.Context, *ConnectRequest) (*RendezvousReply, error)
	// ResolveAll resolves remote servers using the specified peer ID,
	// returning the list of unique id's of these servets.
	//
	// Such UUIDs can be used for establishing aimed connection with all
	// servers under the same ID without randomization games.
	ResolveAll(context.Context, *ID) (*ResolveMetaReply, error)
	// Publish allows to publish the caller server's endpoints.
	//
	// While published the server can be located using the ID extracted from
	// the transport credentials.
	Publish(context.Context, *PublishRequest) (*RendezvousReply, error)
	// Info returns server's internal state.
	Info(context.Context, *Empty) (*RendezvousState, error)
}

func RegisterRendezvousServer(s *grpc.Server, srv RendezvousServer) {
	s.RegisterService(&_Rendezvous_serviceDesc, srv)
}

func _Rendezvous_Resolve_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConnectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RendezvousServer).Resolve(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Rendezvous/Resolve",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RendezvousServer).Resolve(ctx, req.(*ConnectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Rendezvous_ResolveAll_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RendezvousServer).ResolveAll(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Rendezvous/ResolveAll",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RendezvousServer).ResolveAll(ctx, req.(*ID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Rendezvous_Publish_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PublishRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RendezvousServer).Publish(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Rendezvous/Publish",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RendezvousServer).Publish(ctx, req.(*PublishRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Rendezvous_Info_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RendezvousServer).Info(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/sonm.Rendezvous/Info",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RendezvousServer).Info(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _Rendezvous_serviceDesc = grpc.ServiceDesc{
	ServiceName: "sonm.Rendezvous",
	HandlerType: (*RendezvousServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Resolve",
			Handler:    _Rendezvous_Resolve_Handler,
		},
		{
			MethodName: "ResolveAll",
			Handler:    _Rendezvous_ResolveAll_Handler,
		},
		{
			MethodName: "Publish",
			Handler:    _Rendezvous_Publish_Handler,
		},
		{
			MethodName: "Info",
			Handler:    _Rendezvous_Info_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rendezvous.proto",
}

// Begin grpccmd
var _ = xcode.RunE

// Rendezvous
var _RendezvousCmd = &cobra.Command{
	Use:   "rendezvous [method]",
	Short: "Subcommand for the Rendezvous service.",
}

var _Rendezvous_ResolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Make the Resolve method call, input-type: sonm.ConnectRequest output-type: sonm.RendezvousReply",
	RunE: xcode.RunE(
		"Resolve",
		"sonm.ConnectRequest",
		func(c io.Closer) interface{} {
			cc := c.(*grpc.ClientConn)
			return NewRendezvousClient(cc)
		},
	),
}

var _Rendezvous_ResolveCmd_gen = &cobra.Command{
	Use:   "resolve-gen",
	Short: "Generate JSON for method call of Resolve (input-type: sonm.ConnectRequest)",
	RunE:  xcode.TypeToJson("sonm.ConnectRequest"),
}

var _Rendezvous_ResolveAllCmd = &cobra.Command{
	Use:   "resolveAll",
	Short: "Make the ResolveAll method call, input-type: sonm.ID output-type: sonm.ResolveMetaReply",
	RunE: xcode.RunE(
		"ResolveAll",
		"sonm.ID",
		func(c io.Closer) interface{} {
			cc := c.(*grpc.ClientConn)
			return NewRendezvousClient(cc)
		},
	),
}

var _Rendezvous_ResolveAllCmd_gen = &cobra.Command{
	Use:   "resolveAll-gen",
	Short: "Generate JSON for method call of ResolveAll (input-type: sonm.ID)",
	RunE:  xcode.TypeToJson("sonm.ID"),
}

var _Rendezvous_PublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Make the Publish method call, input-type: sonm.PublishRequest output-type: sonm.RendezvousReply",
	RunE: xcode.RunE(
		"Publish",
		"sonm.PublishRequest",
		func(c io.Closer) interface{} {
			cc := c.(*grpc.ClientConn)
			return NewRendezvousClient(cc)
		},
	),
}

var _Rendezvous_PublishCmd_gen = &cobra.Command{
	Use:   "publish-gen",
	Short: "Generate JSON for method call of Publish (input-type: sonm.PublishRequest)",
	RunE:  xcode.TypeToJson("sonm.PublishRequest"),
}

var _Rendezvous_InfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Make the Info method call, input-type: sonm.Empty output-type: sonm.RendezvousState",
	RunE: xcode.RunE(
		"Info",
		"sonm.Empty",
		func(c io.Closer) interface{} {
			cc := c.(*grpc.ClientConn)
			return NewRendezvousClient(cc)
		},
	),
}

var _Rendezvous_InfoCmd_gen = &cobra.Command{
	Use:   "info-gen",
	Short: "Generate JSON for method call of Info (input-type: sonm.Empty)",
	RunE:  xcode.TypeToJson("sonm.Empty"),
}

// Register commands with the root command and service command
func init() {
	xcode.RegisterServiceCmd(_RendezvousCmd)
	_RendezvousCmd.AddCommand(
		_Rendezvous_ResolveCmd,
		_Rendezvous_ResolveCmd_gen,
		_Rendezvous_ResolveAllCmd,
		_Rendezvous_ResolveAllCmd_gen,
		_Rendezvous_PublishCmd,
		_Rendezvous_PublishCmd_gen,
		_Rendezvous_InfoCmd,
		_Rendezvous_InfoCmd_gen,
	)
}

// End grpccmd

func init() { proto.RegisterFile("rendezvous.proto", fileDescriptor12) }

var fileDescriptor12 = []byte{
	// 456 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x53, 0xcd, 0x6a, 0xdb, 0x40,
	0x10, 0xf6, 0xca, 0x49, 0x13, 0x8f, 0x8d, 0xe3, 0x2e, 0xfd, 0x31, 0x3a, 0x19, 0x91, 0x83, 0xe9,
	0x8f, 0x28, 0x2e, 0x94, 0xd0, 0x43, 0x21, 0xc4, 0x39, 0xe8, 0x10, 0x68, 0x36, 0xd7, 0x5e, 0x14,
	0x79, 0xda, 0x8a, 0xae, 0x77, 0x55, 0xed, 0x4a, 0xa0, 0x3e, 0x4b, 0x5f, 0xa4, 0x6f, 0xd3, 0x47,
	0x29, 0xbb, 0x2b, 0x39, 0x92, 0x13, 0x63, 0x0a, 0xb9, 0xd8, 0xa3, 0xd9, 0xef, 0x67, 0x66, 0x76,
	0x07, 0x26, 0x39, 0x8a, 0x15, 0xfe, 0x2a, 0x65, 0xa1, 0xc2, 0x2c, 0x97, 0x5a, 0xd2, 0x03, 0x25,
	0xc5, 0xda, 0x3f, 0x49, 0x85, 0xf9, 0x17, 0x69, 0xec, 0xd2, 0xfe, 0x40, 0xa0, 0x76, 0x61, 0xc0,
	0x61, 0x7c, 0x21, 0x85, 0xc0, 0x44, 0x33, 0xfc, 0x59, 0xa0, 0xd2, 0x74, 0x0c, 0x5e, 0xb4, 0x9c,
	0x92, 0x19, 0x99, 0x8f, 0x98, 0x17, 0x2d, 0xa9, 0x0f, 0xc7, 0x16, 0x9a, 0x48, 0x3e, 0xf5, 0x66,
	0x64, 0x3e, 0x60, 0x9b, 0x6f, 0x1a, 0xc2, 0x28, 0xcb, 0xd3, 0x32, 0xd6, 0x78, 0xbe, 0x5a, 0xe5,
	0x6a, 0xda, 0x9f, 0xf5, 0xe7, 0xc3, 0x05, 0x84, 0xc6, 0x2e, 0x34, 0x29, 0xd6, 0x39, 0x0f, 0xbe,
	0xc0, 0xf8, 0x73, 0x71, 0xcb, 0x53, 0xf5, 0xbd, 0x71, 0x6b, 0xab, 0x93, 0x3d, 0xea, 0xde, 0x1e,
	0xf5, 0x35, 0x9c, 0xb0, 0xcd, 0x04, 0x18, 0x66, 0xbc, 0xa2, 0xaf, 0x00, 0x32, 0x63, 0x98, 0x18,
	0x84, 0x35, 0xe8, 0x0a, 0xb4, 0x4e, 0xff, 0xdb, 0xee, 0x37, 0x69, 0xfb, 0xdd, 0xe8, 0x58, 0x23,
	0xfd, 0x00, 0x87, 0xca, 0x04, 0x53, 0x62, 0xc9, 0x33, 0x47, 0xde, 0x42, 0x85, 0xf6, 0xf7, 0x52,
	0xe8, 0xbc, 0x62, 0x0e, 0xee, 0x5f, 0x03, 0xdc, 0x25, 0xe9, 0x04, 0xfa, 0x3f, 0xb0, 0xaa, 0xe7,
	0x61, 0x42, 0xfa, 0x16, 0x0e, 0xcb, 0x98, 0x17, 0x68, 0x6f, 0x60, 0xb8, 0x78, 0xb9, 0xad, 0x7b,
	0x85, 0xa8, 0x53, 0xf1, 0x8d, 0x39, 0xd4, 0x47, 0xef, 0x8c, 0x04, 0x7f, 0x3c, 0x78, 0x7a, 0x0f,
	0x40, 0x3f, 0xc1, 0x51, 0xc2, 0x53, 0x14, 0x5a, 0xd5, 0x25, 0x9e, 0xee, 0x90, 0x0a, 0x2f, 0x1c,
	0xcc, 0x95, 0xd9, 0x90, 0x0c, 0x5f, 0x61, 0x5e, 0xe2, 0x66, 0x3e, 0x3b, 0xf9, 0x37, 0x0e, 0x56,
	0xf3, 0x6b, 0x92, 0x7f, 0x0d, 0xa3, 0xb6, 0xf0, 0x03, 0xad, 0xbe, 0xee, 0xb6, 0xfa, 0x7c, 0x5b,
	0xdf, 0x5e, 0x6c, 0xab, 0x51, 0x23, 0xd9, 0xf6, 0x7a, 0x04, 0xc9, 0xe0, 0x14, 0x26, 0x0c, 0x95,
	0xe4, 0x25, 0x5e, 0xa1, 0x8e, 0xdd, 0x53, 0x9a, 0x40, 0x3f, 0x5a, 0xba, 0xa9, 0x0d, 0x98, 0x09,
	0x17, 0x7f, 0x09, 0xc0, 0x9d, 0x08, 0x3d, 0x83, 0xa3, 0x9a, 0x44, 0x9f, 0x39, 0x87, 0xee, 0x66,
	0xf9, 0x0f, 0xfb, 0x06, 0x3d, 0xfa, 0xce, 0xe8, 0x58, 0xe6, 0x39, 0xe7, 0xf4, 0xd8, 0xc1, 0xa2,
	0xa5, 0xff, 0xa2, 0x21, 0x74, 0x4b, 0x09, 0x7a, 0xc6, 0xab, 0x5e, 0xa4, 0xc6, 0xab, 0xbb, 0x57,
	0xbb, 0xbd, 0xde, 0xc0, 0x41, 0x24, 0xbe, 0x4a, 0x3a, 0x74, 0x80, 0xcb, 0x75, 0xa6, 0xab, 0xfb,
	0x68, 0xfb, 0x18, 0x83, 0xde, 0xed, 0x13, 0xbb, 0x8c, 0xef, 0xff, 0x05, 0x00, 0x00, 0xff, 0xff,
	0xeb, 0x0f, 0xce, 0x01, 0x5b, 0x04, 0x00, 0x00,
}

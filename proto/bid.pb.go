// Code generated by protoc-gen-go. DO NOT EDIT.
// source: bid.proto

package sonm

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type BidNetwork struct {
	Overlay  bool `protobuf:"varint,1,opt,name=overlay" json:"overlay,omitempty"`
	Outbound bool `protobuf:"varint,2,opt,name=outbound" json:"outbound,omitempty"`
	Incoming bool `protobuf:"varint,3,opt,name=incoming" json:"incoming,omitempty"`
}

func (m *BidNetwork) Reset()                    { *m = BidNetwork{} }
func (m *BidNetwork) String() string            { return proto.CompactTextString(m) }
func (*BidNetwork) ProtoMessage()               {}
func (*BidNetwork) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{0} }

func (m *BidNetwork) GetOverlay() bool {
	if m != nil {
		return m.Overlay
	}
	return false
}

func (m *BidNetwork) GetOutbound() bool {
	if m != nil {
		return m.Outbound
	}
	return false
}

func (m *BidNetwork) GetIncoming() bool {
	if m != nil {
		return m.Incoming
	}
	return false
}

type BidResources struct {
	Network    *BidNetwork       `protobuf:"bytes,1,opt,name=network" json:"network,omitempty"`
	Benchmarks map[string]uint64 `protobuf:"bytes,2,rep,name=benchmarks" json:"benchmarks,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"varint,2,opt,name=value"`
}

func (m *BidResources) Reset()                    { *m = BidResources{} }
func (m *BidResources) String() string            { return proto.CompactTextString(m) }
func (*BidResources) ProtoMessage()               {}
func (*BidResources) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{1} }

func (m *BidResources) GetNetwork() *BidNetwork {
	if m != nil {
		return m.Network
	}
	return nil
}

func (m *BidResources) GetBenchmarks() map[string]uint64 {
	if m != nil {
		return m.Benchmarks
	}
	return nil
}

type BidOrder struct {
	ID        string              `protobuf:"bytes,1,opt,name=ID" json:"ID,omitempty"`
	Duration  *Duration           `protobuf:"bytes,2,opt,name=duration" json:"duration,omitempty"`
	Price     *Price              `protobuf:"bytes,3,opt,name=price" json:"price,omitempty"`
	Blacklist *EthAddress         `protobuf:"bytes,4,opt,name=blacklist" json:"blacklist,omitempty"`
	Identity  MarketIdentityLevel `protobuf:"varint,5,opt,name=identity,enum=sonm.MarketIdentityLevel" json:"identity,omitempty"`
	Tag       string              `protobuf:"bytes,6,opt,name=tag" json:"tag,omitempty"`
	Resources *BidResources       `protobuf:"bytes,7,opt,name=resources" json:"resources,omitempty"`
}

func (m *BidOrder) Reset()                    { *m = BidOrder{} }
func (m *BidOrder) String() string            { return proto.CompactTextString(m) }
func (*BidOrder) ProtoMessage()               {}
func (*BidOrder) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{2} }

func (m *BidOrder) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *BidOrder) GetDuration() *Duration {
	if m != nil {
		return m.Duration
	}
	return nil
}

func (m *BidOrder) GetPrice() *Price {
	if m != nil {
		return m.Price
	}
	return nil
}

func (m *BidOrder) GetBlacklist() *EthAddress {
	if m != nil {
		return m.Blacklist
	}
	return nil
}

func (m *BidOrder) GetIdentity() MarketIdentityLevel {
	if m != nil {
		return m.Identity
	}
	return MarketIdentityLevel_MARKET_ANONIMOUS
}

func (m *BidOrder) GetTag() string {
	if m != nil {
		return m.Tag
	}
	return ""
}

func (m *BidOrder) GetResources() *BidResources {
	if m != nil {
		return m.Resources
	}
	return nil
}

func init() {
	proto.RegisterType((*BidNetwork)(nil), "sonm.BidNetwork")
	proto.RegisterType((*BidResources)(nil), "sonm.BidResources")
	proto.RegisterType((*BidOrder)(nil), "sonm.BidOrder")
}

func init() { proto.RegisterFile("bid.proto", fileDescriptor2) }

var fileDescriptor2 = []byte{
	// 386 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x92, 0xcb, 0xaa, 0xdb, 0x30,
	0x10, 0x86, 0xb1, 0x73, 0xb3, 0xc7, 0x25, 0x49, 0x45, 0x17, 0xaa, 0x57, 0xa9, 0x57, 0x21, 0x0b,
	0x53, 0x5c, 0x0a, 0xa5, 0xd0, 0x45, 0x4d, 0xb2, 0x08, 0xf4, 0x86, 0x1e, 0xa0, 0x20, 0x5b, 0x22,
	0x11, 0x76, 0xa4, 0x20, 0xcb, 0x29, 0x79, 0x8f, 0xbe, 0xce, 0x79, 0xb7, 0x83, 0xa5, 0xd8, 0x09,
	0xe7, 0xac, 0xec, 0x99, 0xff, 0x1b, 0xcd, 0xcc, 0x2f, 0x41, 0x58, 0x08, 0x96, 0x9e, 0xb5, 0x32,
	0x0a, 0x8d, 0x1b, 0x25, 0x4f, 0xf1, 0x42, 0xc8, 0xee, 0x2b, 0x05, 0x75, 0xe9, 0xf8, 0xed, 0x89,
	0xea, 0x8a, 0x9b, 0x73, 0x4d, 0x4b, 0xee, 0x52, 0xc9, 0x5f, 0x80, 0x5c, 0xb0, 0x5f, 0xdc, 0xfc,
	0x53, 0xba, 0x42, 0x18, 0x66, 0xea, 0xc2, 0x75, 0x4d, 0xaf, 0xd8, 0x5b, 0x79, 0xeb, 0x80, 0xf4,
	0x21, 0x8a, 0x21, 0x50, 0xad, 0x29, 0x54, 0x2b, 0x19, 0xf6, 0xad, 0x34, 0xc4, 0x9d, 0x26, 0x64,
	0xa9, 0x4e, 0x42, 0x1e, 0xf0, 0xc8, 0x69, 0x7d, 0x9c, 0x3c, 0x79, 0xf0, 0x26, 0x17, 0x8c, 0xf0,
	0x46, 0xb5, 0xba, 0xe4, 0x0d, 0xda, 0xc0, 0x4c, 0xba, 0x6e, 0xb6, 0x45, 0x94, 0x2d, 0xd3, 0x6e,
	0xc8, 0xf4, 0x3e, 0x05, 0xe9, 0x01, 0x94, 0x03, 0x14, 0x5c, 0x96, 0xc7, 0x6e, 0xec, 0x06, 0xfb,
	0xab, 0xd1, 0x3a, 0xca, 0x92, 0x01, 0x1f, 0xce, 0x4c, 0xf3, 0x01, 0xda, 0x49, 0xa3, 0xaf, 0xe4,
	0xa1, 0x2a, 0xfe, 0x06, 0x8b, 0x17, 0x32, 0x5a, 0xc2, 0xa8, 0xe2, 0x6e, 0xc3, 0x90, 0x74, 0xbf,
	0xe8, 0x1d, 0x4c, 0x2e, 0xb4, 0x6e, 0xb9, 0x5d, 0x6d, 0x4c, 0x5c, 0xf0, 0xd5, 0xff, 0xe2, 0x25,
	0xff, 0x7d, 0x08, 0x72, 0xc1, 0x7e, 0x6b, 0xc6, 0x35, 0x9a, 0x83, 0xbf, 0xdf, 0xde, 0xea, 0xfc,
	0xfd, 0x16, 0x6d, 0x20, 0x60, 0xad, 0xa6, 0x46, 0x28, 0x69, 0x2b, 0xa3, 0x6c, 0xee, 0xa6, 0xdb,
	0xde, 0xb2, 0x64, 0xd0, 0xd1, 0x07, 0x98, 0x9c, 0xb5, 0x28, 0xb9, 0x75, 0x28, 0xca, 0x22, 0x07,
	0xfe, 0xe9, 0x52, 0xc4, 0x29, 0x28, 0x85, 0xb0, 0xa8, 0x69, 0x59, 0xd5, 0xa2, 0x31, 0x78, 0xfc,
	0x68, 0xce, 0xce, 0x1c, 0xbf, 0x33, 0xa6, 0x79, 0xd3, 0x90, 0x3b, 0x82, 0x3e, 0x43, 0x20, 0x18,
	0x97, 0x46, 0x98, 0x2b, 0x9e, 0xac, 0xbc, 0xf5, 0x3c, 0x7b, 0xef, 0xf0, 0x9f, 0xf6, 0x9a, 0xf7,
	0x37, 0xed, 0x07, 0xbf, 0xf0, 0x9a, 0x0c, 0x68, 0xb7, 0xbe, 0xa1, 0x07, 0x3c, 0x75, 0xeb, 0x1b,
	0x7a, 0x40, 0x1f, 0x21, 0xd4, 0xbd, 0x99, 0x78, 0x66, 0x1b, 0xa3, 0xd7, 0x36, 0x93, 0x3b, 0x54,
	0x4c, 0xed, 0xeb, 0xf9, 0xf4, 0x1c, 0x00, 0x00, 0xff, 0xff, 0xcf, 0x43, 0x59, 0x54, 0x74, 0x02,
	0x00, 0x00,
}

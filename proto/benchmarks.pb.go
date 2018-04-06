// Code generated by protoc-gen-go. DO NOT EDIT.
// source: benchmarks.proto

/*
Package sonm is a generated protocol buffer package.

It is generated from these files:
	benchmarks.proto
	bid.proto
	bigint.proto
	capabilities.proto
	container.proto
	deal.proto
	hub.proto
	insonmnia.proto
	marketplace.proto
	miner.proto
	net.proto
	node.proto
	relay.proto
	rendezvous.proto
	timestamp.proto
	volume.proto

It has these top-level messages:
	Benchmark
	Geo
	Resources
	Slot
	Order
	BigInt
	Capabilities
	CPUDevice
	CPU
	RAMDevice
	RAM
	GPUDevice
	GPU
	NetworkDevice
	Network
	StorageDevice
	Storage
	NetworkSpec
	Container
	Deal
	StartTaskRequest
	HubJoinNetworkRequest
	StartTaskReply
	HubStatusReply
	DealRequest
	ApproveDealRequest
	AskPlansReply
	TaskListReply
	DevicesReply
	CreateAskPlanRequest
	PullTaskRequest
	DealInfoReply
	Empty
	ID
	TaskID
	PingReply
	CPUUsage
	MemoryUsage
	NetworkUsage
	ResourceUsage
	InfoReply
	TaskStatusReply
	AvailableResources
	StatusMapReply
	ContainerRestartPolicy
	TaskLogsRequest
	TaskLogsChunk
	DiscoverHubRequest
	TaskResourceRequirements
	Chunk
	Progress
	GetOrdersRequest
	GetOrdersReply
	GetProcessingReply
	TouchOrdersRequest
	MarketOrder
	MarketDeal
	MinerStartRequest
	MinerStartReply
	TaskInfo
	Endpoints
	SaveRequest
	Addr
	SocketAddr
	JoinNetworkRequest
	TaskListRequest
	DealListRequest
	DealListReply
	DealStatusReply
	Worker
	WorkerListReply
	HandshakeRequest
	DiscoverResponse
	HandshakeResponse
	RelayClusterReply
	RelayMetrics
	NetMetrics
	ConnectRequest
	PublishRequest
	RendezvousReply
	RendezvousState
	RendezvousMeeting
	ResolveMetaReply
	Timestamp
	Volume
*/
package sonm

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

//  BenchmarkType describes hardware group for which this benchmark is applicable
type DeviceType int32

const (
	DeviceType_DEV_UNKNOWN DeviceType = 0
	DeviceType_DEV_CPU     DeviceType = 1
	DeviceType_DEV_GPU     DeviceType = 2
	DeviceType_DEV_RAM     DeviceType = 3
	DeviceType_DEV_STORAGE DeviceType = 4
	DeviceType_DEV_NETWORK DeviceType = 5
)

var DeviceType_name = map[int32]string{
	0: "DEV_UNKNOWN",
	1: "DEV_CPU",
	2: "DEV_GPU",
	3: "DEV_RAM",
	4: "DEV_STORAGE",
	5: "DEV_NETWORK",
}
var DeviceType_value = map[string]int32{
	"DEV_UNKNOWN": 0,
	"DEV_CPU":     1,
	"DEV_GPU":     2,
	"DEV_RAM":     3,
	"DEV_STORAGE": 4,
	"DEV_NETWORK": 5,
}

func (x DeviceType) String() string {
	return proto.EnumName(DeviceType_name, int32(x))
}
func (DeviceType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// Benchmark describes a way to measure hardware performance
type Benchmark struct {
	ID          uint64     `protobuf:"varint,1,opt,name=ID" json:"ID,omitempty"`
	Code        string     `protobuf:"bytes,2,opt,name=code" json:"code,omitempty"`
	Type        DeviceType `protobuf:"varint,3,opt,name=type,enum=sonm.DeviceType" json:"type,omitempty"`
	Description string     `protobuf:"bytes,4,opt,name=description" json:"description,omitempty"`
	Image       string     `protobuf:"bytes,5,opt,name=image" json:"image,omitempty"`
	Result      uint64     `protobuf:"varint,6,opt,name=result" json:"result,omitempty"`
}

func (m *Benchmark) Reset()                    { *m = Benchmark{} }
func (m *Benchmark) String() string            { return proto.CompactTextString(m) }
func (*Benchmark) ProtoMessage()               {}
func (*Benchmark) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Benchmark) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *Benchmark) GetCode() string {
	if m != nil {
		return m.Code
	}
	return ""
}

func (m *Benchmark) GetType() DeviceType {
	if m != nil {
		return m.Type
	}
	return DeviceType_DEV_UNKNOWN
}

func (m *Benchmark) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *Benchmark) GetImage() string {
	if m != nil {
		return m.Image
	}
	return ""
}

func (m *Benchmark) GetResult() uint64 {
	if m != nil {
		return m.Result
	}
	return 0
}

func init() {
	proto.RegisterType((*Benchmark)(nil), "sonm.Benchmark")
	proto.RegisterEnum("sonm.DeviceType", DeviceType_name, DeviceType_value)
}

func init() { proto.RegisterFile("benchmarks.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 246 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x90, 0xc1, 0x4a, 0xc3, 0x40,
	0x10, 0x86, 0xdd, 0x74, 0x13, 0xe9, 0x04, 0xea, 0x32, 0x88, 0xec, 0x31, 0x88, 0x87, 0xe0, 0x21,
	0x07, 0x7d, 0x82, 0x6a, 0x42, 0x29, 0xc5, 0xa4, 0xac, 0x89, 0x3d, 0x4a, 0x9b, 0x8e, 0x1a, 0x34,
	0xd9, 0x90, 0x44, 0xa1, 0x0f, 0xe4, 0x7b, 0x4a, 0x56, 0x17, 0x7b, 0xdb, 0xef, 0xff, 0xd8, 0x7f,
	0x86, 0x01, 0xb1, 0xa3, 0xa6, 0x7c, 0xab, 0xb7, 0xdd, 0x7b, 0x1f, 0xb5, 0x9d, 0x1e, 0x34, 0xf2,
	0x5e, 0x37, 0xf5, 0xe5, 0x37, 0x83, 0xe9, 0x9d, 0x55, 0x38, 0x03, 0x67, 0x19, 0x4b, 0x16, 0xb0,
	0x90, 0x2b, 0x67, 0x19, 0x23, 0x02, 0x2f, 0xf5, 0x9e, 0xa4, 0x13, 0xb0, 0x70, 0xaa, 0xcc, 0x1b,
	0xaf, 0x80, 0x0f, 0x87, 0x96, 0xe4, 0x24, 0x60, 0xe1, 0xec, 0x46, 0x44, 0x63, 0x4d, 0x14, 0xd3,
	0x57, 0x55, 0x52, 0x7e, 0x68, 0x49, 0x19, 0x8b, 0x01, 0xf8, 0x7b, 0xea, 0xcb, 0xae, 0x6a, 0x87,
	0x4a, 0x37, 0x92, 0x9b, 0x82, 0xe3, 0x08, 0xcf, 0xc1, 0xad, 0xea, 0xed, 0x2b, 0x49, 0xd7, 0xb8,
	0x5f, 0xc0, 0x0b, 0xf0, 0x3a, 0xea, 0x3f, 0x3f, 0x06, 0xe9, 0x99, 0x2d, 0xfe, 0xe8, 0xfa, 0x05,
	0xe0, 0x7f, 0x06, 0x9e, 0x81, 0x1f, 0x27, 0x4f, 0xcf, 0x45, 0xba, 0x4a, 0xb3, 0x4d, 0x2a, 0x4e,
	0xd0, 0x87, 0xd3, 0x31, 0xb8, 0x5f, 0x17, 0x82, 0x59, 0x58, 0xac, 0x0b, 0xe1, 0x58, 0x50, 0xf3,
	0x07, 0x31, 0xb1, 0xff, 0x1e, 0xf3, 0x4c, 0xcd, 0x17, 0x89, 0xe0, 0x36, 0x48, 0x93, 0x7c, 0x93,
	0xa9, 0x95, 0x70, 0x77, 0x9e, 0x39, 0xce, 0xed, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x20, 0xfb,
	0x5a, 0x65, 0x30, 0x01, 0x00, 0x00,
}

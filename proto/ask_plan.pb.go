// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ask_plan.proto

/*
Package sonm is a generated protocol buffer package.

It is generated from these files:
	ask_plan.proto
	benchmarks.proto
	bigint.proto
	capabilities.proto
	container.proto
	dwh.proto
	insonmnia.proto
	marketplace.proto
	net.proto
	node.proto
	relay.proto
	rendezvous.proto
	timestamp.proto
	volume.proto
	worker.proto

It has these top-level messages:
	AskPlanCPU
	AskPlanGPU
	AskPlanRAM
	AskPlanStorage
	AskPlanNetwork
	AskPlanResources
	AskPlan
	Benchmark
	BigInt
	CPUDevice
	CPU
	RAMDevice
	RAM
	GPUDevice
	GPU
	Network
	StorageDevice
	Storage
	NetworkSpec
	Container
	SortingOption
	DealsRequest
	DWHDealsReply
	DWHDeal
	DealConditionsRequest
	DealConditionsReply
	OrdersRequest
	MatchingOrdersRequest
	DWHOrdersReply
	DWHOrder
	DealCondition
	DWHWorker
	ProfilesRequest
	ProfilesReply
	Profile
	BlacklistRequest
	BlacklistReply
	ValidatorsRequest
	ValidatorsReply
	Validator
	DealChangeRequestsReply
	DealChangeRequest
	DealPayment
	WorkersRequest
	WorkersReply
	Certificate
	MaxMinUint64
	MaxMinBig
	MaxMinTimestamp
	CmpUint64
	BlacklistQuery
	Empty
	ID
	EthID
	TaskID
	Count
	CPUUsage
	MemoryUsage
	NetworkUsage
	ResourceUsage
	ContainerRestartPolicy
	TaskLogsRequest
	TaskLogsChunk
	TaskResourceRequirements
	Chunk
	Progress
	Duration
	EthAddress
	DataSize
	DataSizeRate
	Price
	GetOrdersReply
	Benchmarks
	Deal
	Order
	BidNetwork
	BidResources
	BidOrder
	Addr
	SocketAddr
	Endpoints
	JoinNetworkRequest
	TaskListRequest
	DealFinishRequest
	DealsReply
	OpenDealRequest
	WorkerRemoveRequest
	WorkerListReply
	BalanceReply
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
	StartTaskRequest
	WorkerJoinNetworkRequest
	StartTaskReply
	StatusReply
	AskPlansReply
	TaskListReply
	DevicesReply
	PullTaskRequest
	DealInfoReply
	TaskStatusReply
	StatusMapReply
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

type AskPlan_Status int32

const (
	AskPlan_ACTIVE           AskPlan_Status = 0
	AskPlan_PENDING_DELETION AskPlan_Status = 1
)

var AskPlan_Status_name = map[int32]string{
	0: "ACTIVE",
	1: "PENDING_DELETION",
}
var AskPlan_Status_value = map[string]int32{
	"ACTIVE":           0,
	"PENDING_DELETION": 1,
}

func (x AskPlan_Status) String() string {
	return proto.EnumName(AskPlan_Status_name, int32(x))
}
func (AskPlan_Status) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{6, 0} }

type AskPlanCPU struct {
	CorePercents uint64 `protobuf:"varint,1,opt,name=core_percents,json=corePercents" json:"core_percents,omitempty"`
}

func (m *AskPlanCPU) Reset()                    { *m = AskPlanCPU{} }
func (m *AskPlanCPU) String() string            { return proto.CompactTextString(m) }
func (*AskPlanCPU) ProtoMessage()               {}
func (*AskPlanCPU) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *AskPlanCPU) GetCorePercents() uint64 {
	if m != nil {
		return m.CorePercents
	}
	return 0
}

type AskPlanGPU struct {
	Indexes []uint64 `protobuf:"varint,1,rep,packed,name=indexes" json:"indexes,omitempty"`
	Hashes  []string `protobuf:"bytes,2,rep,name=hashes" json:"hashes,omitempty"`
}

func (m *AskPlanGPU) Reset()                    { *m = AskPlanGPU{} }
func (m *AskPlanGPU) String() string            { return proto.CompactTextString(m) }
func (*AskPlanGPU) ProtoMessage()               {}
func (*AskPlanGPU) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *AskPlanGPU) GetIndexes() []uint64 {
	if m != nil {
		return m.Indexes
	}
	return nil
}

func (m *AskPlanGPU) GetHashes() []string {
	if m != nil {
		return m.Hashes
	}
	return nil
}

type AskPlanRAM struct {
	Size *DataSize `protobuf:"bytes,1,opt,name=size" json:"size,omitempty"`
}

func (m *AskPlanRAM) Reset()                    { *m = AskPlanRAM{} }
func (m *AskPlanRAM) String() string            { return proto.CompactTextString(m) }
func (*AskPlanRAM) ProtoMessage()               {}
func (*AskPlanRAM) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *AskPlanRAM) GetSize() *DataSize {
	if m != nil {
		return m.Size
	}
	return nil
}

type AskPlanStorage struct {
	Size *DataSize `protobuf:"bytes,1,opt,name=size" json:"size,omitempty"`
}

func (m *AskPlanStorage) Reset()                    { *m = AskPlanStorage{} }
func (m *AskPlanStorage) String() string            { return proto.CompactTextString(m) }
func (*AskPlanStorage) ProtoMessage()               {}
func (*AskPlanStorage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *AskPlanStorage) GetSize() *DataSize {
	if m != nil {
		return m.Size
	}
	return nil
}

type AskPlanNetwork struct {
	ThroughputIn  *DataSizeRate `protobuf:"bytes,1,opt,name=throughputIn" json:"throughputIn,omitempty"`
	ThroughputOut *DataSizeRate `protobuf:"bytes,2,opt,name=throughputOut" json:"throughputOut,omitempty"`
	Overlay       bool          `protobuf:"varint,3,opt,name=overlay" json:"overlay,omitempty"`
	Outbound      bool          `protobuf:"varint,4,opt,name=outbound" json:"outbound,omitempty"`
	Incoming      bool          `protobuf:"varint,5,opt,name=incoming" json:"incoming,omitempty"`
}

func (m *AskPlanNetwork) Reset()                    { *m = AskPlanNetwork{} }
func (m *AskPlanNetwork) String() string            { return proto.CompactTextString(m) }
func (*AskPlanNetwork) ProtoMessage()               {}
func (*AskPlanNetwork) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *AskPlanNetwork) GetThroughputIn() *DataSizeRate {
	if m != nil {
		return m.ThroughputIn
	}
	return nil
}

func (m *AskPlanNetwork) GetThroughputOut() *DataSizeRate {
	if m != nil {
		return m.ThroughputOut
	}
	return nil
}

func (m *AskPlanNetwork) GetOverlay() bool {
	if m != nil {
		return m.Overlay
	}
	return false
}

func (m *AskPlanNetwork) GetOutbound() bool {
	if m != nil {
		return m.Outbound
	}
	return false
}

func (m *AskPlanNetwork) GetIncoming() bool {
	if m != nil {
		return m.Incoming
	}
	return false
}

type AskPlanResources struct {
	CPU     *AskPlanCPU     `protobuf:"bytes,1,opt,name=CPU" json:"CPU,omitempty"`
	RAM     *AskPlanRAM     `protobuf:"bytes,2,opt,name=RAM" json:"RAM,omitempty"`
	Storage *AskPlanStorage `protobuf:"bytes,3,opt,name=storage" json:"storage,omitempty"`
	GPU     *AskPlanGPU     `protobuf:"bytes,4,opt,name=GPU" json:"GPU,omitempty"`
	Network *AskPlanNetwork `protobuf:"bytes,5,opt,name=network" json:"network,omitempty"`
}

func (m *AskPlanResources) Reset()                    { *m = AskPlanResources{} }
func (m *AskPlanResources) String() string            { return proto.CompactTextString(m) }
func (*AskPlanResources) ProtoMessage()               {}
func (*AskPlanResources) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *AskPlanResources) GetCPU() *AskPlanCPU {
	if m != nil {
		return m.CPU
	}
	return nil
}

func (m *AskPlanResources) GetRAM() *AskPlanRAM {
	if m != nil {
		return m.RAM
	}
	return nil
}

func (m *AskPlanResources) GetStorage() *AskPlanStorage {
	if m != nil {
		return m.Storage
	}
	return nil
}

func (m *AskPlanResources) GetGPU() *AskPlanGPU {
	if m != nil {
		return m.GPU
	}
	return nil
}

func (m *AskPlanResources) GetNetwork() *AskPlanNetwork {
	if m != nil {
		return m.Network
	}
	return nil
}

type AskPlan struct {
	ID           string            `protobuf:"bytes,1,opt,name=ID" json:"ID,omitempty"`
	OrderID      *BigInt           `protobuf:"bytes,2,opt,name=orderID" json:"orderID,omitempty"`
	DealID       *BigInt           `protobuf:"bytes,3,opt,name=dealID" json:"dealID,omitempty"`
	Duration     *Duration         `protobuf:"bytes,4,opt,name=duration" json:"duration,omitempty"`
	Price        *Price            `protobuf:"bytes,5,opt,name=price" json:"price,omitempty"`
	Blacklist    *EthAddress       `protobuf:"bytes,6,opt,name=blacklist" json:"blacklist,omitempty"`
	Counterparty *EthAddress       `protobuf:"bytes,7,opt,name=counterparty" json:"counterparty,omitempty"`
	Identity     IdentityLevel     `protobuf:"varint,8,opt,name=identity,enum=sonm.IdentityLevel" json:"identity,omitempty"`
	Tag          []byte            `protobuf:"bytes,9,opt,name=tag,proto3" json:"tag,omitempty"`
	Resources    *AskPlanResources `protobuf:"bytes,10,opt,name=resources" json:"resources,omitempty"`
	Status       AskPlan_Status    `protobuf:"varint,11,opt,name=status,enum=sonm.AskPlan_Status" json:"status,omitempty"`
}

func (m *AskPlan) Reset()                    { *m = AskPlan{} }
func (m *AskPlan) String() string            { return proto.CompactTextString(m) }
func (*AskPlan) ProtoMessage()               {}
func (*AskPlan) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *AskPlan) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *AskPlan) GetOrderID() *BigInt {
	if m != nil {
		return m.OrderID
	}
	return nil
}

func (m *AskPlan) GetDealID() *BigInt {
	if m != nil {
		return m.DealID
	}
	return nil
}

func (m *AskPlan) GetDuration() *Duration {
	if m != nil {
		return m.Duration
	}
	return nil
}

func (m *AskPlan) GetPrice() *Price {
	if m != nil {
		return m.Price
	}
	return nil
}

func (m *AskPlan) GetBlacklist() *EthAddress {
	if m != nil {
		return m.Blacklist
	}
	return nil
}

func (m *AskPlan) GetCounterparty() *EthAddress {
	if m != nil {
		return m.Counterparty
	}
	return nil
}

func (m *AskPlan) GetIdentity() IdentityLevel {
	if m != nil {
		return m.Identity
	}
	return IdentityLevel_ANONYMOUS
}

func (m *AskPlan) GetTag() []byte {
	if m != nil {
		return m.Tag
	}
	return nil
}

func (m *AskPlan) GetResources() *AskPlanResources {
	if m != nil {
		return m.Resources
	}
	return nil
}

func (m *AskPlan) GetStatus() AskPlan_Status {
	if m != nil {
		return m.Status
	}
	return AskPlan_ACTIVE
}

func init() {
	proto.RegisterType((*AskPlanCPU)(nil), "sonm.AskPlanCPU")
	proto.RegisterType((*AskPlanGPU)(nil), "sonm.AskPlanGPU")
	proto.RegisterType((*AskPlanRAM)(nil), "sonm.AskPlanRAM")
	proto.RegisterType((*AskPlanStorage)(nil), "sonm.AskPlanStorage")
	proto.RegisterType((*AskPlanNetwork)(nil), "sonm.AskPlanNetwork")
	proto.RegisterType((*AskPlanResources)(nil), "sonm.AskPlanResources")
	proto.RegisterType((*AskPlan)(nil), "sonm.AskPlan")
	proto.RegisterEnum("sonm.AskPlan_Status", AskPlan_Status_name, AskPlan_Status_value)
}

func init() { proto.RegisterFile("ask_plan.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 632 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x54, 0xdd, 0x6a, 0xdb, 0x30,
	0x14, 0x9e, 0x93, 0x34, 0x3f, 0x27, 0x69, 0x96, 0x69, 0xa5, 0x98, 0x5e, 0x65, 0xde, 0x18, 0xa1,
	0x8c, 0x6c, 0xeb, 0xca, 0xd8, 0xd5, 0x20, 0xab, 0x43, 0x30, 0xb4, 0xa9, 0x51, 0x9b, 0xdd, 0x16,
	0xc5, 0x16, 0x89, 0x88, 0x2b, 0x19, 0x49, 0xee, 0xd6, 0x3e, 0xe7, 0xae, 0xf7, 0x0a, 0x7b, 0x85,
	0x21, 0x5b, 0x4e, 0xe6, 0x42, 0x61, 0x77, 0x3e, 0xdf, 0xcf, 0xf9, 0xd3, 0xc1, 0xd0, 0x27, 0x6a,
	0x73, 0x93, 0x26, 0x84, 0x8f, 0x53, 0x29, 0xb4, 0x40, 0x0d, 0x25, 0xf8, 0xed, 0x51, 0x6f, 0xc9,
	0x56, 0x8c, 0xeb, 0x02, 0x3b, 0x7a, 0xce, 0xb8, 0x41, 0x39, 0x23, 0x16, 0x78, 0x71, 0x4b, 0xe4,
	0x86, 0xea, 0x34, 0x21, 0x11, 0x2d, 0x20, 0xef, 0x23, 0xc0, 0x44, 0x6d, 0xc2, 0x84, 0xf0, 0xb3,
	0x70, 0x81, 0x5e, 0xc3, 0x7e, 0x24, 0x24, 0xbd, 0x49, 0xa9, 0x8c, 0x28, 0xd7, 0xca, 0x75, 0x86,
	0xce, 0xa8, 0x81, 0x7b, 0x06, 0x0c, 0x2d, 0xe6, 0x7d, 0xdd, 0x5a, 0x66, 0xe1, 0x02, 0xb9, 0xd0,
	0x62, 0x3c, 0xa6, 0x3f, 0xa9, 0x11, 0xd7, 0x47, 0x0d, 0x5c, 0x86, 0xe8, 0x10, 0x9a, 0x6b, 0xa2,
	0xd6, 0x54, 0xb9, 0xb5, 0x61, 0x7d, 0xd4, 0xc1, 0x36, 0xf2, 0x3e, 0x6c, 0xfd, 0x78, 0x72, 0x81,
	0x3c, 0x68, 0x28, 0xf6, 0x40, 0xf3, 0x4a, 0xdd, 0x93, 0xfe, 0xd8, 0x74, 0x3c, 0xf6, 0x89, 0x26,
	0x57, 0xec, 0x81, 0xe2, 0x9c, 0xf3, 0x4e, 0xa1, 0x6f, 0x1d, 0x57, 0x5a, 0x48, 0xb2, 0xa2, 0xff,
	0xe5, 0xfa, 0xe5, 0x6c, 0x6d, 0x73, 0xaa, 0x7f, 0x08, 0xb9, 0x41, 0x9f, 0xa1, 0xa7, 0xd7, 0x52,
	0x64, 0xab, 0x75, 0x9a, 0xe9, 0x80, 0x5b, 0x3b, 0x7a, 0x64, 0x27, 0x9a, 0xe2, 0x8a, 0x0e, 0x7d,
	0x81, 0xfd, 0x5d, 0x7c, 0x99, 0x69, 0xb7, 0xf6, 0xa4, 0xb1, 0x2a, 0x34, 0xeb, 0x11, 0x77, 0x54,
	0x26, 0xe4, 0xde, 0xad, 0x0f, 0x9d, 0x51, 0x1b, 0x97, 0x21, 0x3a, 0x82, 0xb6, 0xc8, 0xf4, 0x52,
	0x64, 0x3c, 0x76, 0x1b, 0x39, 0xb5, 0x8d, 0x0d, 0xc7, 0x78, 0x24, 0x6e, 0x19, 0x5f, 0xb9, 0x7b,
	0x05, 0x57, 0xc6, 0xde, 0x6f, 0x07, 0x06, 0xe5, 0xfe, 0xa8, 0x12, 0x99, 0x8c, 0xa8, 0x42, 0x1e,
	0xd4, 0xcf, 0xc2, 0x85, 0x9d, 0x67, 0x50, 0xb4, 0xb5, 0x7b, 0x57, 0x6c, 0x48, 0xa3, 0xc1, 0x93,
	0x0b, 0xdb, 0x7a, 0x55, 0x83, 0x27, 0x17, 0xd8, 0x90, 0x68, 0x0c, 0x2d, 0x55, 0xac, 0x38, 0x6f,
	0xb7, 0x7b, 0x72, 0x50, 0xd1, 0xd9, 0xf5, 0xe3, 0x52, 0x64, 0x72, 0xce, 0xc2, 0x45, 0xde, 0xff,
	0xe3, 0x9c, 0x33, 0x53, 0xd7, 0x5c, 0xc8, 0x18, 0x5a, 0xbc, 0xd8, 0x7f, 0x3e, 0xcb, 0xe3, 0x9c,
	0xf6, 0x6d, 0x70, 0x29, 0xf2, 0xfe, 0xd4, 0xa1, 0x65, 0x39, 0xd4, 0x87, 0x5a, 0xe0, 0xe7, 0x63,
	0x75, 0x70, 0x2d, 0xf0, 0xd1, 0x5b, 0x68, 0x09, 0x19, 0x53, 0x19, 0xf8, 0x76, 0x8e, 0x5e, 0x91,
	0xeb, 0x1b, 0x5b, 0x05, 0x5c, 0xe3, 0x92, 0x44, 0x6f, 0xa0, 0x19, 0x53, 0x92, 0x04, 0xbe, 0x1d,
	0xa3, 0x2a, 0xb3, 0x1c, 0x3a, 0x86, 0x76, 0x9c, 0x49, 0xa2, 0x99, 0xe0, 0x76, 0x84, 0xf2, 0x92,
	0x2c, 0x8a, 0xb7, 0x3c, 0x7a, 0x05, 0x7b, 0xa9, 0x64, 0x11, 0xb5, 0x33, 0x74, 0x0b, 0x61, 0x68,
	0x20, 0x5c, 0x30, 0x68, 0x0c, 0x9d, 0x65, 0x42, 0xa2, 0x4d, 0xc2, 0x94, 0x76, 0x9b, 0xff, 0xae,
	0x64, 0xaa, 0xd7, 0x93, 0x38, 0x96, 0x54, 0x29, 0xbc, 0x93, 0xa0, 0x53, 0xe8, 0x45, 0x22, 0xe3,
	0x9a, 0xca, 0x94, 0x48, 0x7d, 0xef, 0xb6, 0x9e, 0xb0, 0x54, 0x54, 0xe8, 0x3d, 0xb4, 0x59, 0x4c,
	0xb9, 0x66, 0xfa, 0xde, 0x6d, 0x0f, 0x9d, 0x51, 0xff, 0xe4, 0x65, 0xe1, 0x08, 0x2c, 0x7a, 0x4e,
	0xef, 0x68, 0x82, 0xb7, 0x22, 0x34, 0x80, 0xba, 0x26, 0x2b, 0xb7, 0x33, 0x74, 0x46, 0x3d, 0x6c,
	0x3e, 0xd1, 0x29, 0x74, 0x64, 0x79, 0x3a, 0x2e, 0xe4, 0x55, 0x0f, 0xab, 0xf7, 0x50, 0xb2, 0x78,
	0x27, 0x44, 0xef, 0xa0, 0xa9, 0x34, 0xd1, 0x99, 0x72, 0xbb, 0x79, 0xd9, 0xea, 0x33, 0x8e, 0xaf,
	0x72, 0x0e, 0x5b, 0x8d, 0x77, 0x0c, 0xcd, 0x02, 0x41, 0x00, 0xcd, 0xc9, 0xd9, 0x75, 0xf0, 0x7d,
	0x3a, 0x78, 0x86, 0x0e, 0x60, 0x10, 0x4e, 0xe7, 0x7e, 0x30, 0x9f, 0xdd, 0xf8, 0xd3, 0xf3, 0xe9,
	0x75, 0x70, 0x39, 0x1f, 0x38, 0xcb, 0x66, 0xfe, 0x2f, 0xfa, 0xf4, 0x37, 0x00, 0x00, 0xff, 0xff,
	0xbd, 0xd9, 0x4e, 0xf0, 0xd5, 0x04, 0x00, 0x00,
}

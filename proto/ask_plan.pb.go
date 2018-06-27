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
	NetFlags
	Network
	StorageDevice
	Storage
	Registry
	ContainerRestartPolicy
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
	BlacklistsContainingUserReply
	ValidatorsRequest
	ValidatorsReply
	DWHValidator
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
	TaskLogsRequest
	TaskLogsChunk
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
	QuickBuyRequest
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
	TaskSpec
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
	ResourcePool
	SchedulerData
	SalesmanData
	DebugStateReply
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
	NetFlags      *NetFlags     `protobuf:"bytes,3,opt,name=netFlags" json:"netFlags,omitempty"`
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

func (m *AskPlanNetwork) GetNetFlags() *NetFlags {
	if m != nil {
		return m.NetFlags
	}
	return nil
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
	return IdentityLevel_UNKNOWN
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
	// 619 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x94, 0xd1, 0x6e, 0xd3, 0x30,
	0x14, 0x86, 0x49, 0xdb, 0xa5, 0xeb, 0x59, 0x57, 0x8a, 0x99, 0xa6, 0x68, 0x57, 0x25, 0x20, 0x54,
	0x4d, 0x28, 0xc0, 0x98, 0x10, 0x57, 0x48, 0x65, 0x29, 0x55, 0xa4, 0xad, 0x8b, 0xbc, 0x95, 0xdb,
	0xca, 0x4d, 0xac, 0xd6, 0x6a, 0xe6, 0x44, 0xb6, 0x03, 0x6c, 0x0f, 0xc5, 0x23, 0xf1, 0x0a, 0xbc,
	0x02, 0x72, 0xe2, 0xb4, 0x0b, 0xda, 0x24, 0xee, 0xe2, 0xff, 0xff, 0x7e, 0xfb, 0x9c, 0x63, 0x2b,
	0xd0, 0x23, 0x72, 0x3d, 0xcf, 0x12, 0xc2, 0xbd, 0x4c, 0xa4, 0x2a, 0x45, 0x2d, 0x99, 0xf2, 0x9b,
	0xa3, 0xee, 0x82, 0x2d, 0x19, 0x57, 0xa5, 0x76, 0x84, 0x22, 0x92, 0x91, 0x05, 0x4b, 0x98, 0x62,
	0x54, 0x1a, 0xed, 0x29, 0xe3, 0x9a, 0xe4, 0x8c, 0x18, 0xe1, 0xd9, 0x0d, 0x11, 0x6b, 0xaa, 0xb2,
	0x84, 0x44, 0xb4, 0x94, 0xdc, 0xf7, 0x00, 0x23, 0xb9, 0x0e, 0x13, 0xc2, 0xcf, 0xc2, 0x19, 0x7a,
	0x09, 0xfb, 0x51, 0x2a, 0xe8, 0x3c, 0xa3, 0x22, 0xa2, 0x5c, 0x49, 0xc7, 0x1a, 0x58, 0xc3, 0x16,
	0xee, 0x6a, 0x31, 0x34, 0x9a, 0xfb, 0x79, 0x13, 0x99, 0x84, 0x33, 0xe4, 0x40, 0x9b, 0xf1, 0x98,
	0xfe, 0xa4, 0x1a, 0x6e, 0x0e, 0x5b, 0xb8, 0x5a, 0xa2, 0x43, 0xb0, 0x57, 0x44, 0xae, 0xa8, 0x74,
	0x1a, 0x83, 0xe6, 0xb0, 0x83, 0xcd, 0xca, 0x7d, 0xb7, 0xc9, 0xe3, 0xd1, 0x05, 0x72, 0xa1, 0x25,
	0xd9, 0x1d, 0x2d, 0x4e, 0xda, 0x3b, 0xe9, 0x79, 0xba, 0x62, 0xcf, 0x27, 0x8a, 0x5c, 0xb1, 0x3b,
	0x8a, 0x0b, 0xcf, 0x3d, 0x85, 0x9e, 0x49, 0x5c, 0xa9, 0x54, 0x90, 0x25, 0xfd, 0xaf, 0xd4, 0x2f,
	0x6b, 0x13, 0x9b, 0x52, 0xf5, 0x23, 0x15, 0x6b, 0xf4, 0x11, 0xba, 0x6a, 0x25, 0xd2, 0x7c, 0xb9,
	0xca, 0x72, 0x15, 0x70, 0x13, 0x47, 0xff, 0xc4, 0x89, 0xa2, 0xb8, 0xc6, 0xa1, 0x4f, 0xb0, 0xbf,
	0x5d, 0x5f, 0xe6, 0xca, 0x69, 0x3c, 0x1a, 0xac, 0x83, 0xe8, 0x18, 0x76, 0x39, 0x55, 0x5f, 0x13,
	0xb2, 0x94, 0x4e, 0xf3, 0x7e, 0xb1, 0x53, 0xa3, 0xe2, 0x8d, 0xef, 0xfe, 0xb6, 0xa0, 0x5f, 0x4d,
	0x86, 0xca, 0x34, 0x17, 0x11, 0x95, 0xc8, 0x85, 0xe6, 0x59, 0x38, 0x33, 0x95, 0xf6, 0xcb, 0xec,
	0xf6, 0xc6, 0xb0, 0x36, 0x35, 0x83, 0x47, 0x17, 0xa6, 0xa8, 0x3a, 0x83, 0x47, 0x17, 0x58, 0x9b,
	0xc8, 0x83, 0xb6, 0x2c, 0x87, 0x67, 0xea, 0x38, 0xa8, 0x71, 0x66, 0xb0, 0xb8, 0x82, 0xf4, 0x9e,
	0x93, 0x70, 0xe6, 0xb4, 0x1e, 0xd8, 0x73, 0xa2, 0xcf, 0xd5, 0x77, 0xef, 0x41, 0x9b, 0x97, 0x93,
	0x75, 0x76, 0x1e, 0xd8, 0xd3, 0x4c, 0x1d, 0x57, 0x90, 0xfb, 0xa7, 0x09, 0x6d, 0xe3, 0xa1, 0x1e,
	0x34, 0x02, 0xbf, 0x68, 0xab, 0x83, 0x1b, 0x81, 0x8f, 0x5e, 0x43, 0x3b, 0x15, 0x31, 0x15, 0x81,
	0x6f, 0xfa, 0xe8, 0x96, 0x7b, 0x7d, 0x61, 0xcb, 0x80, 0x2b, 0x5c, 0x99, 0xe8, 0x15, 0xd8, 0x31,
	0x25, 0x49, 0xe0, 0x9b, 0x36, 0xea, 0x98, 0xf1, 0xf4, 0xd8, 0xe3, 0x5c, 0x10, 0xc5, 0x52, 0x6e,
	0x5a, 0xa8, 0xde, 0x88, 0x51, 0xf1, 0xc6, 0x47, 0x2f, 0x60, 0x27, 0x13, 0x2c, 0xa2, 0xa6, 0x87,
	0xbd, 0x12, 0x0c, 0xb5, 0x84, 0x4b, 0x07, 0x79, 0xd0, 0x59, 0x24, 0x24, 0x5a, 0x27, 0x4c, 0x2a,
	0xc7, 0xbe, 0x3f, 0x92, 0xb1, 0x5a, 0x8d, 0xe2, 0x58, 0x50, 0x29, 0xf1, 0x16, 0x41, 0xa7, 0xd0,
	0x8d, 0xd2, 0x9c, 0x2b, 0x2a, 0x32, 0x22, 0xd4, 0xad, 0xd3, 0x7e, 0x24, 0x52, 0xa3, 0xd0, 0x5b,
	0xd8, 0x65, 0x31, 0xe5, 0x8a, 0xa9, 0x5b, 0x67, 0x77, 0x60, 0x0d, 0x7b, 0x27, 0xcf, 0xcb, 0x44,
	0x60, 0xd4, 0x73, 0xfa, 0x9d, 0x26, 0x78, 0x03, 0xa1, 0x3e, 0x34, 0x15, 0x59, 0x3a, 0x9d, 0x81,
	0x35, 0xec, 0x62, 0xfd, 0x89, 0x4e, 0xa1, 0x23, 0xaa, 0xa7, 0xe3, 0x40, 0x71, 0xea, 0x61, 0xfd,
	0x3d, 0x54, 0x2e, 0xde, 0x82, 0xe8, 0x0d, 0xd8, 0x52, 0x11, 0x95, 0x4b, 0x67, 0xaf, 0x38, 0xb6,
	0x7e, 0x8d, 0xde, 0x55, 0xe1, 0x61, 0xc3, 0xb8, 0xc7, 0x60, 0x97, 0x0a, 0x02, 0xb0, 0x47, 0x67,
	0xd7, 0xc1, 0xb7, 0x71, 0xff, 0x09, 0x3a, 0x80, 0x7e, 0x38, 0x9e, 0xfa, 0xc1, 0x74, 0x32, 0xf7,
	0xc7, 0xe7, 0xe3, 0xeb, 0xe0, 0x72, 0xda, 0xb7, 0x16, 0x76, 0xf1, 0x97, 0xf9, 0xf0, 0x37, 0x00,
	0x00, 0xff, 0xff, 0x5e, 0x0d, 0x70, 0xb8, 0xc3, 0x04, 0x00, 0x00,
}

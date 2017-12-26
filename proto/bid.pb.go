// Code generated by protoc-gen-go. DO NOT EDIT.
// source: bid.proto

/*
Package sonm is a generated protocol buffer package.

It is generated from these files:
	bid.proto
	capabilities.proto
	deal.proto
	hub.proto
	insonmnia.proto
	locator.proto
	marketplace.proto
	miner.proto
	node.proto
	price.proto

It has these top-level messages:
	Geo
	Resources
	Slot
	Order
	Capabilities
	CPUDevice
	RAMDevice
	GPUDevice
	Deal
	ListReply
	HubStartTaskRequest
	HubStartTaskReply
	HubStatusReply
	DealRequest
	GetDevicePropertiesReply
	SetDevicePropertiesRequest
	SlotsReply
	GetAllSlotsReply
	AddSlotRequest
	RemoveSlotRequest
	GetRegisteredWorkersReply
	TaskListReply
	CPUDeviceInfo
	GPUDeviceInfo
	DevicesReply
	InsertSlotRequest
	PullTaskRequest
	DealInfoReply
	CompletedTask
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
	Timestamp
	Chunk
	Progress
	AnnounceRequest
	ResolveRequest
	ResolveReply
	GetOrdersRequest
	GetOrdersReply
	GetProcessingReply
	MinerHandshakeRequest
	MinerHandshakeReply
	MinerStartRequest
	SocketAddr
	MinerStartReply
	TaskInfo
	Route
	MinerStatusMapRequest
	SaveRequest
	TaskListRequest
	DealListRequest
	DealListReply
	BigInt
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

type OrderType int32

const (
	OrderType_ANY OrderType = 0
	OrderType_BID OrderType = 1
	OrderType_ASK OrderType = 2
)

var OrderType_name = map[int32]string{
	0: "ANY",
	1: "BID",
	2: "ASK",
}
var OrderType_value = map[string]int32{
	"ANY": 0,
	"BID": 1,
	"ASK": 2,
}

func (x OrderType) String() string {
	return proto.EnumName(OrderType_name, int32(x))
}
func (OrderType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// Geo represent GeoIP results for node
type Geo struct {
	Country string  `protobuf:"bytes,1,opt,name=country" json:"country,omitempty"`
	City    string  `protobuf:"bytes,2,opt,name=city" json:"city,omitempty"`
	Lat     float32 `protobuf:"fixed32,3,opt,name=lat" json:"lat,omitempty"`
	Lon     float32 `protobuf:"fixed32,4,opt,name=lon" json:"lon,omitempty"`
}

func (m *Geo) Reset()                    { *m = Geo{} }
func (m *Geo) String() string            { return proto.CompactTextString(m) }
func (*Geo) ProtoMessage()               {}
func (*Geo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Geo) GetCountry() string {
	if m != nil {
		return m.Country
	}
	return ""
}

func (m *Geo) GetCity() string {
	if m != nil {
		return m.City
	}
	return ""
}

func (m *Geo) GetLat() float32 {
	if m != nil {
		return m.Lat
	}
	return 0
}

func (m *Geo) GetLon() float32 {
	if m != nil {
		return m.Lon
	}
	return 0
}

type Resources struct {
	// CPU core count
	CpuCores uint64 `protobuf:"varint,1,opt,name=cpuCores" json:"cpuCores,omitempty"`
	// RAM, in bytes
	RamBytes uint64 `protobuf:"varint,2,opt,name=ramBytes" json:"ramBytes,omitempty"`
	// GPU devices count
	GpuCount GPUCount `protobuf:"varint,3,opt,name=gpuCount,enum=sonm.GPUCount" json:"gpuCount,omitempty"`
	// todo: discuss
	// storage volume, in Megabytes
	Storage uint64 `protobuf:"varint,4,opt,name=storage" json:"storage,omitempty"`
	// Inbound network traffic (the higher value), in bytes
	NetTrafficIn uint64 `protobuf:"varint,5,opt,name=netTrafficIn" json:"netTrafficIn,omitempty"`
	// Outbound network traffic (the higher value), in bytes
	NetTrafficOut uint64 `protobuf:"varint,6,opt,name=netTrafficOut" json:"netTrafficOut,omitempty"`
	// Allowed network connections
	NetworkType NetworkType `protobuf:"varint,7,opt,name=networkType,enum=sonm.NetworkType" json:"networkType,omitempty"`
	// Other properties/benchmarks. The higher means better.
	Properties map[string]float64 `protobuf:"bytes,8,rep,name=properties" json:"properties,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"fixed64,2,opt,name=value"`
}

func (m *Resources) Reset()                    { *m = Resources{} }
func (m *Resources) String() string            { return proto.CompactTextString(m) }
func (*Resources) ProtoMessage()               {}
func (*Resources) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Resources) GetCpuCores() uint64 {
	if m != nil {
		return m.CpuCores
	}
	return 0
}

func (m *Resources) GetRamBytes() uint64 {
	if m != nil {
		return m.RamBytes
	}
	return 0
}

func (m *Resources) GetGpuCount() GPUCount {
	if m != nil {
		return m.GpuCount
	}
	return GPUCount_NO_GPU
}

func (m *Resources) GetStorage() uint64 {
	if m != nil {
		return m.Storage
	}
	return 0
}

func (m *Resources) GetNetTrafficIn() uint64 {
	if m != nil {
		return m.NetTrafficIn
	}
	return 0
}

func (m *Resources) GetNetTrafficOut() uint64 {
	if m != nil {
		return m.NetTrafficOut
	}
	return 0
}

func (m *Resources) GetNetworkType() NetworkType {
	if m != nil {
		return m.NetworkType
	}
	return NetworkType_NO_NETWORK
}

func (m *Resources) GetProperties() map[string]float64 {
	if m != nil {
		return m.Properties
	}
	return nil
}

type Slot struct {
	// Buyer’s rating. Got from Buyer’s profile for BID orders rating_supplier.
	BuyerRating int64 `protobuf:"varint,1,opt,name=buyerRating" json:"buyerRating,omitempty"`
	// Supplier’s rating. Got from Supplier’s profile for ASK orders.
	SupplierRating int64 `protobuf:"varint,2,opt,name=supplierRating" json:"supplierRating,omitempty"`
	// Geo represent Worker's position
	Geo *Geo `protobuf:"bytes,3,opt,name=geo" json:"geo,omitempty"`
	// Hardware resources requirements
	Resources *Resources `protobuf:"bytes,4,opt,name=resources" json:"resources,omitempty"`
	// Duration is resource rent duration in seconds
	Duration uint64 `protobuf:"varint,5,opt,name=duration" json:"duration,omitempty"`
}

func (m *Slot) Reset()                    { *m = Slot{} }
func (m *Slot) String() string            { return proto.CompactTextString(m) }
func (*Slot) ProtoMessage()               {}
func (*Slot) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Slot) GetBuyerRating() int64 {
	if m != nil {
		return m.BuyerRating
	}
	return 0
}

func (m *Slot) GetSupplierRating() int64 {
	if m != nil {
		return m.SupplierRating
	}
	return 0
}

func (m *Slot) GetGeo() *Geo {
	if m != nil {
		return m.Geo
	}
	return nil
}

func (m *Slot) GetResources() *Resources {
	if m != nil {
		return m.Resources
	}
	return nil
}

func (m *Slot) GetDuration() uint64 {
	if m != nil {
		return m.Duration
	}
	return 0
}

type Order struct {
	// Order ID, UUIDv4
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	// Buyer's EtherumID
	ByuerID string `protobuf:"bytes,2,opt,name=byuerID" json:"byuerID,omitempty"`
	// Supplier's is EtherumID
	SupplierID string `protobuf:"bytes,3,opt,name=supplierID" json:"supplierID,omitempty"`
	// Order price
	Price string `protobuf:"bytes,4,opt,name=price" json:"price,omitempty"`
	// Order type (Bid or Ask)
	OrderType OrderType `protobuf:"varint,5,opt,name=orderType,enum=sonm.OrderType" json:"orderType,omitempty"`
	// Slot describe resource requiements
	Slot *Slot `protobuf:"bytes,6,opt,name=slot" json:"slot,omitempty"`
}

func (m *Order) Reset()                    { *m = Order{} }
func (m *Order) String() string            { return proto.CompactTextString(m) }
func (*Order) ProtoMessage()               {}
func (*Order) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Order) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Order) GetByuerID() string {
	if m != nil {
		return m.ByuerID
	}
	return ""
}

func (m *Order) GetSupplierID() string {
	if m != nil {
		return m.SupplierID
	}
	return ""
}

func (m *Order) GetPrice() string {
	if m != nil {
		return m.Price
	}
	return ""
}

func (m *Order) GetOrderType() OrderType {
	if m != nil {
		return m.OrderType
	}
	return OrderType_ANY
}

func (m *Order) GetSlot() *Slot {
	if m != nil {
		return m.Slot
	}
	return nil
}

func init() {
	proto.RegisterType((*Geo)(nil), "sonm.Geo")
	proto.RegisterType((*Resources)(nil), "sonm.Resources")
	proto.RegisterType((*Slot)(nil), "sonm.Slot")
	proto.RegisterType((*Order)(nil), "sonm.Order")
	proto.RegisterEnum("sonm.OrderType", OrderType_name, OrderType_value)
}

func init() { proto.RegisterFile("bid.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 514 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x53, 0x5d, 0x6b, 0x13, 0x41,
	0x14, 0x75, 0x3f, 0xd2, 0x64, 0xef, 0x6a, 0x12, 0x2f, 0x3e, 0x2c, 0x11, 0x6a, 0x08, 0x52, 0x42,
	0xc1, 0x3c, 0xa4, 0x2f, 0x22, 0x88, 0x58, 0x23, 0x25, 0x08, 0x4d, 0x99, 0x56, 0xc4, 0xc7, 0xcd,
	0x66, 0x1a, 0x86, 0xa6, 0x33, 0xcb, 0xec, 0x8c, 0xb2, 0xff, 0xc0, 0x5f, 0x24, 0xfe, 0x3c, 0x99,
	0xbb, 0x1f, 0x49, 0xf3, 0xb4, 0xf7, 0x9c, 0x7b, 0x2e, 0xf7, 0xe3, 0xcc, 0x42, 0xb4, 0x16, 0x9b,
	0x59, 0xae, 0x95, 0x51, 0x18, 0x16, 0x4a, 0x3e, 0x8e, 0x06, 0x42, 0xba, 0xaf, 0x14, 0x69, 0x45,
	0x4f, 0x7e, 0x40, 0x70, 0xc5, 0x15, 0x26, 0xd0, 0xcd, 0x94, 0x95, 0x46, 0x97, 0x89, 0x37, 0xf6,
	0xa6, 0x11, 0x6b, 0x20, 0x22, 0x84, 0x99, 0x30, 0x65, 0xe2, 0x13, 0x4d, 0x31, 0x0e, 0x21, 0xd8,
	0xa5, 0x26, 0x09, 0xc6, 0xde, 0xd4, 0x67, 0x2e, 0x24, 0x46, 0xc9, 0x24, 0xac, 0x19, 0x25, 0x27,
	0x7f, 0x02, 0x88, 0x18, 0x2f, 0x94, 0xd5, 0x19, 0x2f, 0x70, 0x04, 0xbd, 0x2c, 0xb7, 0x5f, 0x94,
	0xe6, 0x05, 0x35, 0x08, 0x59, 0x8b, 0x5d, 0x4e, 0xa7, 0x8f, 0x97, 0xa5, 0xe1, 0x05, 0x75, 0x09,
	0x59, 0x8b, 0xf1, 0x1c, 0x7a, 0x5b, 0xa7, 0xb3, 0xb2, 0x6a, 0xd7, 0x9f, 0xf7, 0x67, 0x6e, 0x81,
	0xd9, 0xd5, 0xcd, 0x77, 0x62, 0x59, 0x9b, 0x77, 0x3b, 0x14, 0x46, 0xe9, 0x74, 0xcb, 0x69, 0x8e,
	0x90, 0x35, 0x10, 0x27, 0xf0, 0x5c, 0x72, 0x73, 0xa7, 0xd3, 0xfb, 0x7b, 0x91, 0x2d, 0x65, 0xd2,
	0xa1, 0xf4, 0x13, 0x0e, 0xdf, 0xc2, 0x8b, 0x3d, 0x5e, 0x59, 0x93, 0x9c, 0x90, 0xe8, 0x29, 0x89,
	0x17, 0x10, 0x4b, 0x6e, 0x7e, 0x2b, 0xfd, 0x70, 0x57, 0xe6, 0x3c, 0xe9, 0xd2, 0x48, 0x2f, 0xab,
	0x91, 0xae, 0xf7, 0x09, 0x76, 0xa8, 0xc2, 0x4f, 0x00, 0xb9, 0x56, 0x39, 0xd7, 0x46, 0xf0, 0x22,
	0xe9, 0x8d, 0x83, 0x69, 0x3c, 0x7f, 0x53, 0xd5, 0xb4, 0x17, 0x9a, 0xdd, 0xb4, 0x8a, 0xaf, 0xee,
	0xee, 0xec, 0xa0, 0x64, 0xf4, 0x11, 0x06, 0x47, 0x69, 0x77, 0xf0, 0x07, 0xde, 0x98, 0xe5, 0x42,
	0x7c, 0x05, 0x9d, 0x5f, 0xe9, 0xce, 0x72, 0xba, 0xa1, 0xc7, 0x2a, 0xf0, 0xc1, 0x7f, 0xef, 0x4d,
	0xfe, 0x79, 0x10, 0xde, 0xee, 0x94, 0xc1, 0x31, 0xc4, 0x6b, 0x5b, 0x72, 0xcd, 0x52, 0x23, 0xe4,
	0x96, 0x8a, 0x03, 0x76, 0x48, 0xe1, 0x19, 0xf4, 0x0b, 0x9b, 0xe7, 0x3b, 0xd1, 0x8a, 0x7c, 0x12,
	0x1d, 0xb1, 0xf8, 0x1a, 0x82, 0x2d, 0x57, 0x64, 0x49, 0x3c, 0x8f, 0x6a, 0x4b, 0xb8, 0x62, 0x8e,
	0xc5, 0x77, 0x10, 0xe9, 0x66, 0x2f, 0xb2, 0x22, 0x9e, 0x0f, 0x8e, 0xd6, 0x65, 0x7b, 0x85, 0xf3,
	0x7f, 0x63, 0x75, 0x6a, 0x84, 0x6a, 0x9c, 0x69, 0xf1, 0xe4, 0xaf, 0x07, 0x9d, 0x95, 0xde, 0x70,
	0x8d, 0x7d, 0xf0, 0xc5, 0xa6, 0xde, 0xd7, 0x17, 0x1b, 0xe7, 0xf6, 0xba, 0xb4, 0x5c, 0x2f, 0x17,
	0xf5, 0xd3, 0x6c, 0x20, 0x9e, 0x02, 0x34, 0xd3, 0x2e, 0x17, 0x34, 0x62, 0xc4, 0x0e, 0x18, 0x77,
	0xa8, 0x5c, 0x8b, 0xac, 0x7a, 0x25, 0x11, 0xab, 0x80, 0x1b, 0x5a, 0xb9, 0x46, 0xe4, 0x6b, 0x87,
	0x7c, 0xad, 0x87, 0x5e, 0x35, 0x34, 0xdb, 0x2b, 0xf0, 0x14, 0xc2, 0x62, 0xa7, 0xaa, 0x57, 0x12,
	0xcf, 0xa1, 0x52, 0xba, 0x23, 0x33, 0xe2, 0xcf, 0xcf, 0x20, 0x6a, 0xeb, 0xb0, 0x0b, 0xc1, 0xe7,
	0xeb, 0x9f, 0xc3, 0x67, 0x2e, 0xb8, 0x5c, 0x2e, 0x86, 0x1e, 0x31, 0xb7, 0xdf, 0x86, 0xfe, 0xfa,
	0x84, 0x7e, 0xc3, 0x8b, 0xff, 0x01, 0x00, 0x00, 0xff, 0xff, 0x31, 0x5a, 0x63, 0xb4, 0xaa, 0x03,
	0x00, 0x00,
}

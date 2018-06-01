// Code generated by protoc-gen-go. DO NOT EDIT.
// source: capabilities.proto

package sonm

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type GPUVendorType int32

const (
	GPUVendorType_GPU_UNKNOWN GPUVendorType = 0
	GPUVendorType_NVIDIA      GPUVendorType = 1
	GPUVendorType_RADEON      GPUVendorType = 2
	GPUVendorType_FAKE        GPUVendorType = 99
)

var GPUVendorType_name = map[int32]string{
	0:  "GPU_UNKNOWN",
	1:  "NVIDIA",
	2:  "RADEON",
	99: "FAKE",
}
var GPUVendorType_value = map[string]int32{
	"GPU_UNKNOWN": 0,
	"NVIDIA":      1,
	"RADEON":      2,
	"FAKE":        99,
}

func (x GPUVendorType) String() string {
	return proto.EnumName(GPUVendorType_name, int32(x))
}
func (GPUVendorType) EnumDescriptor() ([]byte, []int) { return fileDescriptor3, []int{0} }

type CPUDevice struct {
	// ModelName describes full model name.
	// For example "Intel(R) Core(TM) i5-5257U CPU @ 2.70GHz".
	ModelName string `protobuf:"bytes,1,opt,name=modelName" json:"modelName,omitempty"`
	// Cores describes number of cores on a CPU device.
	Cores uint32 `protobuf:"varint,2,opt,name=cores" json:"cores,omitempty"`
	// Sockets describes number of CPU sockets on a host system.
	Sockets uint32 `protobuf:"varint,3,opt,name=sockets" json:"sockets,omitempty"`
}

func (m *CPUDevice) Reset()                    { *m = CPUDevice{} }
func (m *CPUDevice) String() string            { return proto.CompactTextString(m) }
func (*CPUDevice) ProtoMessage()               {}
func (*CPUDevice) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{0} }

func (m *CPUDevice) GetModelName() string {
	if m != nil {
		return m.ModelName
	}
	return ""
}

func (m *CPUDevice) GetCores() uint32 {
	if m != nil {
		return m.Cores
	}
	return 0
}

func (m *CPUDevice) GetSockets() uint32 {
	if m != nil {
		return m.Sockets
	}
	return 0
}

type CPU struct {
	Device     *CPUDevice            `protobuf:"bytes,1,opt,name=device" json:"device,omitempty"`
	Benchmarks map[uint64]*Benchmark `protobuf:"bytes,2,rep,name=benchmarks" json:"benchmarks,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *CPU) Reset()                    { *m = CPU{} }
func (m *CPU) String() string            { return proto.CompactTextString(m) }
func (*CPU) ProtoMessage()               {}
func (*CPU) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{1} }

func (m *CPU) GetDevice() *CPUDevice {
	if m != nil {
		return m.Device
	}
	return nil
}

func (m *CPU) GetBenchmarks() map[uint64]*Benchmark {
	if m != nil {
		return m.Benchmarks
	}
	return nil
}

type RAMDevice struct {
	Total     uint64 `protobuf:"varint,1,opt,name=total" json:"total,omitempty"`
	Available uint64 `protobuf:"varint,2,opt,name=available" json:"available,omitempty"`
}

func (m *RAMDevice) Reset()                    { *m = RAMDevice{} }
func (m *RAMDevice) String() string            { return proto.CompactTextString(m) }
func (*RAMDevice) ProtoMessage()               {}
func (*RAMDevice) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{2} }

func (m *RAMDevice) GetTotal() uint64 {
	if m != nil {
		return m.Total
	}
	return 0
}

func (m *RAMDevice) GetAvailable() uint64 {
	if m != nil {
		return m.Available
	}
	return 0
}

type RAM struct {
	Device     *RAMDevice            `protobuf:"bytes,1,opt,name=device" json:"device,omitempty"`
	Benchmarks map[uint64]*Benchmark `protobuf:"bytes,2,rep,name=benchmarks" json:"benchmarks,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *RAM) Reset()                    { *m = RAM{} }
func (m *RAM) String() string            { return proto.CompactTextString(m) }
func (*RAM) ProtoMessage()               {}
func (*RAM) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{3} }

func (m *RAM) GetDevice() *RAMDevice {
	if m != nil {
		return m.Device
	}
	return nil
}

func (m *RAM) GetBenchmarks() map[uint64]*Benchmark {
	if m != nil {
		return m.Benchmarks
	}
	return nil
}

type GPUDevice struct {
	// ID returns unique device ID on workers machine,
	// typically PCI bus ID
	ID string `protobuf:"bytes,1,opt,name=ID" json:"ID,omitempty"`
	// VendorID returns an unique device vendor identifier
	VendorID uint64 `protobuf:"varint,2,opt,name=vendorID" json:"vendorID,omitempty"`
	// VendorName returns GPU vendor name.
	VendorName string `protobuf:"bytes,3,opt,name=vendorName" json:"vendorName,omitempty"`
	// DeviceID returns device ID (e.g.: NVidia)
	DeviceID uint64 `protobuf:"varint,5,opt,name=deviceID" json:"deviceID,omitempty"`
	// DeviceName returns device name, (e.g.: 1080Ti)
	DeviceName string `protobuf:"bytes,6,opt,name=deviceName" json:"deviceName,omitempty"`
	// MajorNumber returns device's major number
	MajorNumber uint64 `protobuf:"varint,7,opt,name=majorNumber" json:"majorNumber,omitempty"`
	// MinorNumber returns device's minor number
	MinorNumber uint64 `protobuf:"varint,8,opt,name=minorNumber" json:"minorNumber,omitempty"`
	// Memory is amount of vmem for device, in bytes
	Memory uint64 `protobuf:"varint,9,opt,name=Memory" json:"Memory,omitempty"`
	// Hash string built from device parameters
	Hash string `protobuf:"bytes,10,opt,name=hash" json:"hash,omitempty"`
}

func (m *GPUDevice) Reset()                    { *m = GPUDevice{} }
func (m *GPUDevice) String() string            { return proto.CompactTextString(m) }
func (*GPUDevice) ProtoMessage()               {}
func (*GPUDevice) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{4} }

func (m *GPUDevice) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *GPUDevice) GetVendorID() uint64 {
	if m != nil {
		return m.VendorID
	}
	return 0
}

func (m *GPUDevice) GetVendorName() string {
	if m != nil {
		return m.VendorName
	}
	return ""
}

func (m *GPUDevice) GetDeviceID() uint64 {
	if m != nil {
		return m.DeviceID
	}
	return 0
}

func (m *GPUDevice) GetDeviceName() string {
	if m != nil {
		return m.DeviceName
	}
	return ""
}

func (m *GPUDevice) GetMajorNumber() uint64 {
	if m != nil {
		return m.MajorNumber
	}
	return 0
}

func (m *GPUDevice) GetMinorNumber() uint64 {
	if m != nil {
		return m.MinorNumber
	}
	return 0
}

func (m *GPUDevice) GetMemory() uint64 {
	if m != nil {
		return m.Memory
	}
	return 0
}

func (m *GPUDevice) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

type GPU struct {
	Device     *GPUDevice            `protobuf:"bytes,1,opt,name=device" json:"device,omitempty"`
	Benchmarks map[uint64]*Benchmark `protobuf:"bytes,2,rep,name=benchmarks" json:"benchmarks,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *GPU) Reset()                    { *m = GPU{} }
func (m *GPU) String() string            { return proto.CompactTextString(m) }
func (*GPU) ProtoMessage()               {}
func (*GPU) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{5} }

func (m *GPU) GetDevice() *GPUDevice {
	if m != nil {
		return m.Device
	}
	return nil
}

func (m *GPU) GetBenchmarks() map[uint64]*Benchmark {
	if m != nil {
		return m.Benchmarks
	}
	return nil
}

type Network struct {
	In            uint64                `protobuf:"varint,1,opt,name=in" json:"in,omitempty"`
	Out           uint64                `protobuf:"varint,2,opt,name=out" json:"out,omitempty"`
	Overlay       bool                  `protobuf:"varint,3,opt,name=overlay" json:"overlay,omitempty"`
	Incoming      bool                  `protobuf:"varint,4,opt,name=incoming" json:"incoming,omitempty"`
	Outbound      bool                  `protobuf:"varint,5,opt,name=outbound" json:"outbound,omitempty"`
	BenchmarksIn  map[uint64]*Benchmark `protobuf:"bytes,6,rep,name=benchmarksIn" json:"benchmarksIn,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	BenchmarksOut map[uint64]*Benchmark `protobuf:"bytes,7,rep,name=benchmarksOut" json:"benchmarksOut,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *Network) Reset()                    { *m = Network{} }
func (m *Network) String() string            { return proto.CompactTextString(m) }
func (*Network) ProtoMessage()               {}
func (*Network) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{6} }

func (m *Network) GetIn() uint64 {
	if m != nil {
		return m.In
	}
	return 0
}

func (m *Network) GetOut() uint64 {
	if m != nil {
		return m.Out
	}
	return 0
}

func (m *Network) GetOverlay() bool {
	if m != nil {
		return m.Overlay
	}
	return false
}

func (m *Network) GetIncoming() bool {
	if m != nil {
		return m.Incoming
	}
	return false
}

func (m *Network) GetOutbound() bool {
	if m != nil {
		return m.Outbound
	}
	return false
}

func (m *Network) GetBenchmarksIn() map[uint64]*Benchmark {
	if m != nil {
		return m.BenchmarksIn
	}
	return nil
}

func (m *Network) GetBenchmarksOut() map[uint64]*Benchmark {
	if m != nil {
		return m.BenchmarksOut
	}
	return nil
}

type StorageDevice struct {
	BytesAvailable uint64 `protobuf:"varint,1,opt,name=bytesAvailable" json:"bytesAvailable,omitempty"`
}

func (m *StorageDevice) Reset()                    { *m = StorageDevice{} }
func (m *StorageDevice) String() string            { return proto.CompactTextString(m) }
func (*StorageDevice) ProtoMessage()               {}
func (*StorageDevice) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{7} }

func (m *StorageDevice) GetBytesAvailable() uint64 {
	if m != nil {
		return m.BytesAvailable
	}
	return 0
}

type Storage struct {
	Device     *StorageDevice        `protobuf:"bytes,1,opt,name=device" json:"device,omitempty"`
	Benchmarks map[uint64]*Benchmark `protobuf:"bytes,2,rep,name=benchmarks" json:"benchmarks,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *Storage) Reset()                    { *m = Storage{} }
func (m *Storage) String() string            { return proto.CompactTextString(m) }
func (*Storage) ProtoMessage()               {}
func (*Storage) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{8} }

func (m *Storage) GetDevice() *StorageDevice {
	if m != nil {
		return m.Device
	}
	return nil
}

func (m *Storage) GetBenchmarks() map[uint64]*Benchmark {
	if m != nil {
		return m.Benchmarks
	}
	return nil
}

func init() {
	proto.RegisterType((*CPUDevice)(nil), "sonm.CPUDevice")
	proto.RegisterType((*CPU)(nil), "sonm.CPU")
	proto.RegisterType((*RAMDevice)(nil), "sonm.RAMDevice")
	proto.RegisterType((*RAM)(nil), "sonm.RAM")
	proto.RegisterType((*GPUDevice)(nil), "sonm.GPUDevice")
	proto.RegisterType((*GPU)(nil), "sonm.GPU")
	proto.RegisterType((*Network)(nil), "sonm.Network")
	proto.RegisterType((*StorageDevice)(nil), "sonm.StorageDevice")
	proto.RegisterType((*Storage)(nil), "sonm.Storage")
	proto.RegisterEnum("sonm.GPUVendorType", GPUVendorType_name, GPUVendorType_value)
}

func init() { proto.RegisterFile("capabilities.proto", fileDescriptor3) }

var fileDescriptor3 = []byte{
	// 651 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x55, 0xcb, 0x6e, 0xd3, 0x4a,
	0x18, 0x3e, 0x76, 0xee, 0x7f, 0x4e, 0xda, 0x9c, 0x39, 0x15, 0x32, 0x11, 0x97, 0x28, 0x12, 0x50,
	0x81, 0x94, 0x45, 0x59, 0x70, 0x91, 0x10, 0x72, 0xeb, 0x36, 0xb2, 0xaa, 0x38, 0x61, 0x20, 0x45,
	0xac, 0xd0, 0xd8, 0x19, 0xb5, 0x26, 0xb6, 0xa7, 0xb2, 0xc7, 0x41, 0x59, 0xf3, 0x40, 0xac, 0x79,
	0x00, 0xde, 0x0b, 0x79, 0x66, 0x3a, 0x4e, 0x5a, 0x2a, 0x81, 0x2a, 0x75, 0xf7, 0x5f, 0xe6, 0xfb,
	0xfe, 0xdb, 0x17, 0x07, 0x50, 0x40, 0xce, 0x89, 0x1f, 0x46, 0x21, 0x0f, 0x69, 0x36, 0x3c, 0x4f,
	0x19, 0x67, 0xa8, 0x9a, 0xb1, 0x24, 0xee, 0x75, 0x7d, 0x9a, 0x04, 0x67, 0x31, 0x49, 0x17, 0x2a,
	0x3e, 0xf8, 0x04, 0xad, 0x83, 0xe9, 0xcc, 0xa1, 0xcb, 0x30, 0xa0, 0xe8, 0x1e, 0xb4, 0x62, 0x36,
	0xa7, 0x91, 0x47, 0x62, 0x6a, 0x19, 0x7d, 0x63, 0xb7, 0x85, 0xcb, 0x00, 0xda, 0x81, 0x5a, 0xc0,
	0x52, 0x9a, 0x59, 0x66, 0xdf, 0xd8, 0xed, 0x60, 0xe9, 0x20, 0x0b, 0x1a, 0x19, 0x0b, 0x16, 0x94,
	0x67, 0x56, 0x45, 0xc4, 0x2f, 0xdc, 0xc1, 0x0f, 0x03, 0x2a, 0x07, 0xd3, 0x19, 0x7a, 0x02, 0xf5,
	0xb9, 0xe0, 0x17, 0x94, 0xed, 0xbd, 0xed, 0x61, 0xd1, 0xcb, 0x50, 0x97, 0xc5, 0x2a, 0x8d, 0x5e,
	0x01, 0x94, 0xfd, 0x59, 0x66, 0xbf, 0xb2, 0xdb, 0xde, 0xbb, 0xab, 0x1f, 0x0f, 0xf7, 0x75, 0xee,
	0x30, 0xe1, 0xe9, 0x0a, 0xaf, 0x3d, 0xee, 0x79, 0xb0, 0x7d, 0x29, 0x8d, 0xba, 0x50, 0x59, 0xd0,
	0x95, 0xa8, 0x59, 0xc5, 0x85, 0x89, 0x1e, 0x41, 0x6d, 0x49, 0xa2, 0x9c, 0x8a, 0x01, 0x74, 0x1f,
	0x1a, 0x87, 0x65, 0xf6, 0xb5, 0xf9, 0xd2, 0x18, 0xbc, 0x85, 0x16, 0xb6, 0xc7, 0x6a, 0x2d, 0x3b,
	0x50, 0xe3, 0x8c, 0x93, 0x48, 0x71, 0x49, 0xa7, 0x58, 0x16, 0x59, 0x92, 0x30, 0x22, 0x7e, 0x24,
	0x19, 0xab, 0xb8, 0x0c, 0x88, 0xe1, 0xb1, 0x3d, 0xbe, 0x6e, 0x78, 0x4d, 0xfe, 0x27, 0xc3, 0x63,
	0x7b, 0x7c, 0xab, 0xc3, 0x7f, 0x33, 0xa1, 0x35, 0xd2, 0xa2, 0xd8, 0x02, 0xd3, 0x75, 0x94, 0x1a,
	0x4c, 0xd7, 0x41, 0x3d, 0x68, 0x2e, 0x69, 0x32, 0x67, 0xa9, 0xeb, 0xa8, 0xb1, 0xb5, 0x8f, 0x1e,
	0x00, 0x48, 0x5b, 0x28, 0xa8, 0x22, 0x30, 0x6b, 0x91, 0x02, 0x2b, 0xc7, 0x75, 0x1d, 0xab, 0x26,
	0xb1, 0x17, 0x7e, 0x81, 0x95, 0xb6, 0xc0, 0xd6, 0x25, 0xb6, 0x8c, 0xa0, 0x3e, 0xb4, 0x63, 0xf2,
	0x85, 0xa5, 0x5e, 0x1e, 0xfb, 0x34, 0xb5, 0x1a, 0x02, 0xbe, 0x1e, 0x12, 0x2f, 0xc2, 0x44, 0xbf,
	0x68, 0xaa, 0x17, 0x65, 0x08, 0xdd, 0x81, 0xfa, 0x98, 0xc6, 0x2c, 0x5d, 0x59, 0x2d, 0x91, 0x54,
	0x1e, 0x42, 0x50, 0x3d, 0x23, 0xd9, 0x99, 0x05, 0xa2, 0xaa, 0xb0, 0xc5, 0x05, 0x47, 0xd7, 0xcb,
	0x77, 0xf4, 0x37, 0xf2, 0x1d, 0xdd, 0xb2, 0x7c, 0xbf, 0x57, 0xa0, 0xe1, 0x51, 0xfe, 0x95, 0xa5,
	0x8b, 0xe2, 0x7e, 0x61, 0xa2, 0x78, 0xcc, 0x30, 0x29, 0x88, 0x59, 0xce, 0xd5, 0xe9, 0x0a, 0xb3,
	0xf8, 0x09, 0xb3, 0x25, 0x4d, 0x23, 0xb2, 0x12, 0x27, 0x6b, 0xe2, 0x0b, 0xb7, 0xb8, 0x57, 0x98,
	0x04, 0x2c, 0x0e, 0x93, 0x53, 0xab, 0x2a, 0x52, 0xda, 0x2f, 0x72, 0x2c, 0xe7, 0x3e, 0xcb, 0x93,
	0xb9, 0xb8, 0x65, 0x13, 0x6b, 0x1f, 0x1d, 0xc0, 0xbf, 0xe5, 0x74, 0x6e, 0x62, 0xd5, 0xc5, 0x32,
	0x1e, 0xca, 0x8e, 0x55, 0x63, 0x6b, 0x0b, 0x71, 0x13, 0xb9, 0x92, 0x0d, 0x10, 0x3a, 0x82, 0x4e,
	0xe9, 0x4f, 0x72, 0x6e, 0x35, 0x04, 0x4b, 0xff, 0x3a, 0x96, 0x49, 0xce, 0x25, 0xcd, 0x26, 0xac,
	0x37, 0x85, 0xff, 0xae, 0x94, 0xba, 0xd1, 0x7a, 0x7b, 0xef, 0x00, 0x5d, 0x2d, 0x7b, 0xb3, 0x8b,
	0xbd, 0x80, 0xce, 0x7b, 0xce, 0x52, 0x72, 0x4a, 0xd5, 0xcf, 0xee, 0x31, 0x6c, 0xf9, 0x2b, 0x4e,
	0x33, 0x5b, 0x7f, 0x63, 0x24, 0xf1, 0xa5, 0xe8, 0xe0, 0xa7, 0x01, 0x0d, 0x85, 0x44, 0xcf, 0x2e,
	0x49, 0xf5, 0x7f, 0x59, 0x70, 0x83, 0x58, 0xcb, 0xf5, 0xcd, 0x6f, 0xe4, 0x7a, 0x7f, 0x03, 0x70,
	0x9b, 0x92, 0x7d, 0xba, 0x0f, 0x9d, 0xd1, 0x74, 0x76, 0x22, 0xbe, 0x15, 0x1f, 0x56, 0xe7, 0x14,
	0x6d, 0x43, 0x7b, 0x34, 0x9d, 0x7d, 0x9e, 0x79, 0xc7, 0xde, 0xe4, 0xa3, 0xd7, 0xfd, 0x07, 0x01,
	0xd4, 0xbd, 0x13, 0xd7, 0x71, 0xed, 0xae, 0x51, 0xd8, 0xd8, 0x76, 0x0e, 0x27, 0x5e, 0xd7, 0x44,
	0x4d, 0xa8, 0x1e, 0xd9, 0xc7, 0x87, 0xdd, 0xc0, 0xaf, 0x8b, 0xff, 0xb4, 0xe7, 0xbf, 0x02, 0x00,
	0x00, 0xff, 0xff, 0x91, 0x87, 0x95, 0x3c, 0x01, 0x07, 0x00, 0x00,
}

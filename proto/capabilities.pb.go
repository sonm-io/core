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
	GPUVendorType_REMOTE      GPUVendorType = 100
)

var GPUVendorType_name = map[int32]string{
	0:   "GPU_UNKNOWN",
	1:   "NVIDIA",
	2:   "RADEON",
	99:  "FAKE",
	100: "REMOTE",
}
var GPUVendorType_value = map[string]int32{
	"GPU_UNKNOWN": 0,
	"NVIDIA":      1,
	"RADEON":      2,
	"FAKE":        99,
	"REMOTE":      100,
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
	// Total amount of RAM on machine
	Total uint64 `protobuf:"varint,1,opt,name=total" json:"total,omitempty"`
	// Available amount of RAM for task scheduling
	Available uint64 `protobuf:"varint,2,opt,name=available" json:"available,omitempty"`
	// Used amount of RAM on machine
	Used uint64 `protobuf:"varint,3,opt,name=used" json:"used,omitempty"`
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

func (m *RAMDevice) GetUsed() uint64 {
	if m != nil {
		return m.Used
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
	// DeviceFiles is something like "/dev/nvidia0",
	// "/dev/nvidiactl", "/dev/dri/card0" and so on.
	// This devices should be attached to a container.
	DeviceFiles []string `protobuf:"bytes,11,rep,name=deviceFiles" json:"deviceFiles,omitempty"`
	// DriverVolumes maps volume name into
	// "hostPath:containerPath" pair.
	// Applicable to nvidia drivers.
	DriverVolumes map[string]string `protobuf:"bytes,12,rep,name=driverVolumes" json:"driverVolumes,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
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

func (m *GPUDevice) GetDeviceFiles() []string {
	if m != nil {
		return m.DeviceFiles
	}
	return nil
}

func (m *GPUDevice) GetDriverVolumes() map[string]string {
	if m != nil {
		return m.DriverVolumes
	}
	return nil
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

type NetFlags struct {
	Flags uint64 `protobuf:"varint,1,opt,name=flags" json:"flags,omitempty"`
}

func (m *NetFlags) Reset()                    { *m = NetFlags{} }
func (m *NetFlags) String() string            { return proto.CompactTextString(m) }
func (*NetFlags) ProtoMessage()               {}
func (*NetFlags) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{6} }

func (m *NetFlags) GetFlags() uint64 {
	if m != nil {
		return m.Flags
	}
	return 0
}

type Network struct {
	In            uint64                `protobuf:"varint,1,opt,name=in" json:"in,omitempty"`
	Out           uint64                `protobuf:"varint,2,opt,name=out" json:"out,omitempty"`
	NetFlags      *NetFlags             `protobuf:"bytes,3,opt,name=netFlags" json:"netFlags,omitempty"`
	BenchmarksIn  map[uint64]*Benchmark `protobuf:"bytes,4,rep,name=benchmarksIn" json:"benchmarksIn,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	BenchmarksOut map[uint64]*Benchmark `protobuf:"bytes,5,rep,name=benchmarksOut" json:"benchmarksOut,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *Network) Reset()                    { *m = Network{} }
func (m *Network) String() string            { return proto.CompactTextString(m) }
func (*Network) ProtoMessage()               {}
func (*Network) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{7} }

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

func (m *Network) GetNetFlags() *NetFlags {
	if m != nil {
		return m.NetFlags
	}
	return nil
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
func (*StorageDevice) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{8} }

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
func (*Storage) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{9} }

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
	proto.RegisterType((*NetFlags)(nil), "sonm.NetFlags")
	proto.RegisterType((*Network)(nil), "sonm.Network")
	proto.RegisterType((*StorageDevice)(nil), "sonm.StorageDevice")
	proto.RegisterType((*Storage)(nil), "sonm.Storage")
	proto.RegisterEnum("sonm.GPUVendorType", GPUVendorType_name, GPUVendorType_value)
}

func init() { proto.RegisterFile("capabilities.proto", fileDescriptor3) }

var fileDescriptor3 = []byte{
	// 729 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x55, 0xdb, 0x6e, 0xd3, 0x4a,
	0x14, 0x3d, 0x89, 0x73, 0xf3, 0x4e, 0xd3, 0xe6, 0xcc, 0xa9, 0x8e, 0x7c, 0xa2, 0x03, 0x44, 0x91,
	0x80, 0xa8, 0x48, 0x79, 0x28, 0x0f, 0x5c, 0x24, 0x24, 0x4c, 0x9d, 0x04, 0xab, 0x8a, 0x13, 0xa6,
	0x4d, 0x11, 0x4f, 0xc8, 0x49, 0x86, 0xd6, 0xc4, 0xf6, 0x54, 0xbe, 0x04, 0xe5, 0x13, 0xf8, 0x1c,
	0x1e, 0xf9, 0x00, 0xfe, 0x0b, 0xcd, 0x25, 0x63, 0xa7, 0xa5, 0x12, 0xa8, 0x52, 0xdf, 0xf6, 0x6d,
	0xed, 0xed, 0x59, 0x7b, 0xcd, 0x18, 0xd0, 0xdc, 0xbd, 0x74, 0x67, 0x9e, 0xef, 0x25, 0x1e, 0x89,
	0x7b, 0x97, 0x11, 0x4d, 0x28, 0x2a, 0xc5, 0x34, 0x0c, 0x5a, 0xcd, 0x19, 0x09, 0xe7, 0x17, 0x81,
	0x1b, 0x2d, 0x65, 0xbc, 0xf3, 0x01, 0xf4, 0xa3, 0xc9, 0xd4, 0x22, 0x2b, 0x6f, 0x4e, 0xd0, 0xff,
	0xa0, 0x07, 0x74, 0x41, 0x7c, 0xc7, 0x0d, 0x88, 0x51, 0x68, 0x17, 0xba, 0x3a, 0xce, 0x02, 0x68,
	0x1f, 0xca, 0x73, 0x1a, 0x91, 0xd8, 0x28, 0xb6, 0x0b, 0xdd, 0x06, 0x16, 0x0e, 0x32, 0xa0, 0x1a,
	0xd3, 0xf9, 0x92, 0x24, 0xb1, 0xa1, 0xf1, 0xf8, 0xc6, 0xed, 0x7c, 0x2f, 0x80, 0x76, 0x34, 0x99,
	0xa2, 0xc7, 0x50, 0x59, 0xf0, 0xfe, 0xbc, 0x65, 0xfd, 0x70, 0xaf, 0xc7, 0xbe, 0xa5, 0xa7, 0xc6,
	0x62, 0x99, 0x46, 0x2f, 0x00, 0xb2, 0xef, 0x33, 0x8a, 0x6d, 0xad, 0x5b, 0x3f, 0xfc, 0x4f, 0x15,
	0xf7, 0xde, 0xa8, 0x5c, 0x3f, 0x4c, 0xa2, 0x35, 0xce, 0x15, 0xb7, 0x1c, 0xd8, 0xbb, 0x92, 0x46,
	0x4d, 0xd0, 0x96, 0x64, 0xcd, 0x67, 0x96, 0x30, 0x33, 0xd1, 0x43, 0x28, 0xaf, 0x5c, 0x3f, 0x25,
	0xfc, 0x00, 0xea, 0x3b, 0x14, 0x0e, 0x8b, 0xec, 0xcb, 0xe2, 0xf3, 0x42, 0xe7, 0x04, 0x74, 0x6c,
	0x8e, 0x24, 0x2d, 0xfb, 0x50, 0x4e, 0x68, 0xe2, 0xfa, 0xb2, 0x97, 0x70, 0x18, 0x59, 0xee, 0xca,
	0xf5, 0x7c, 0x77, 0xe6, 0x8b, 0x8e, 0x25, 0x9c, 0x05, 0x10, 0x82, 0x52, 0x1a, 0x93, 0x05, 0xe7,
	0xa4, 0x84, 0xb9, 0xcd, 0x09, 0xc1, 0xe6, 0xe8, 0x26, 0x42, 0xd4, 0xc0, 0xdf, 0x21, 0x04, 0x9b,
	0xa3, 0x3b, 0x25, 0xe4, 0x9b, 0x06, 0xfa, 0x50, 0x09, 0x65, 0x17, 0x8a, 0xb6, 0x25, 0x15, 0x52,
	0xb4, 0x2d, 0xd4, 0x82, 0xda, 0x8a, 0x84, 0x0b, 0x1a, 0xd9, 0x96, 0xa4, 0x42, 0xf9, 0xe8, 0x3e,
	0x80, 0xb0, 0xb9, 0xaa, 0x34, 0x8e, 0xc9, 0x45, 0x18, 0x56, 0x1c, 0xd7, 0xb6, 0x8c, 0xb2, 0xc0,
	0x6e, 0x7c, 0x86, 0x15, 0x36, 0xc7, 0x56, 0x04, 0x36, 0x8b, 0xa0, 0x36, 0xd4, 0x03, 0xf7, 0x33,
	0x8d, 0x9c, 0x34, 0x98, 0x91, 0xc8, 0xa8, 0x72, 0x78, 0x3e, 0xc4, 0x2b, 0xbc, 0x50, 0x55, 0xd4,
	0x64, 0x45, 0x16, 0x42, 0xff, 0x42, 0x65, 0x44, 0x02, 0x1a, 0xad, 0x0d, 0x9d, 0x27, 0xa5, 0xc7,
	0x36, 0x78, 0xe1, 0xc6, 0x17, 0x06, 0xf0, 0xa9, 0xdc, 0x66, 0xdd, 0xc4, 0xf4, 0x81, 0xe7, 0x93,
	0xd8, 0xa8, 0xb7, 0xb5, 0xae, 0x8e, 0xf3, 0x21, 0xf4, 0x16, 0x1a, 0x8b, 0xc8, 0x5b, 0x91, 0xe8,
	0x8c, 0xfa, 0x69, 0x40, 0x62, 0x63, 0x87, 0x6f, 0xad, 0x23, 0xa8, 0x55, 0x0c, 0xf6, 0xac, 0x7c,
	0x91, 0x58, 0xdf, 0x36, 0xb0, 0xf5, 0x1a, 0xd0, 0xf5, 0xa2, 0xfc, 0x12, 0x75, 0xb1, 0xc4, 0xfd,
	0xfc, 0x12, 0xf5, 0xfc, 0xce, 0x98, 0xde, 0x86, 0x37, 0x5f, 0xc0, 0xe1, 0x9f, 0x5c, 0xc0, 0xe1,
	0x1d, 0x5f, 0xc0, 0x36, 0xd4, 0x1c, 0x92, 0x0c, 0x7c, 0xf7, 0x3c, 0x66, 0x27, 0xfc, 0xc4, 0x8c,
	0xcd, 0xfd, 0xe3, 0x4e, 0xe7, 0xab, 0x06, 0x55, 0x87, 0x24, 0x5f, 0x68, 0xb4, 0x64, 0x7a, 0xf4,
	0x42, 0x99, 0x2e, 0x7a, 0x21, 0x1b, 0x4d, 0xd3, 0x44, 0x4a, 0x91, 0x99, 0xe8, 0x00, 0x6a, 0xa1,
	0xec, 0xc7, 0x35, 0x58, 0x3f, 0xdc, 0x15, 0xd3, 0x37, 0x53, 0xb0, 0xca, 0xa3, 0x23, 0xd8, 0xc9,
	0x4e, 0x66, 0x87, 0x46, 0x89, 0x13, 0xf1, 0x40, 0xd5, 0xb3, 0x91, 0x39, 0x32, 0xec, 0x50, 0xd0,
	0xb1, 0x05, 0x42, 0x03, 0x68, 0x64, 0xfe, 0x38, 0x4d, 0x8c, 0x32, 0xef, 0xd2, 0xbe, 0xa9, 0xcb,
	0x38, 0x4d, 0xa4, 0x0c, 0xb6, 0x60, 0xad, 0x09, 0xfc, 0x7d, 0x6d, 0xd4, 0xad, 0xa8, 0x6d, 0xbd,
	0x03, 0x74, 0x7d, 0xec, 0xed, 0xb6, 0xf5, 0x0c, 0x1a, 0x27, 0x09, 0x8d, 0xdc, 0x73, 0x22, 0x1f,
	0x88, 0x47, 0xb0, 0x3b, 0x5b, 0x27, 0x24, 0x36, 0xd5, 0x0b, 0x29, 0x1a, 0x5f, 0x89, 0x76, 0x7e,
	0x14, 0xa0, 0x2a, 0x91, 0xe8, 0xc9, 0x15, 0x99, 0xfe, 0x23, 0x06, 0x6e, 0x35, 0x56, 0x52, 0x7d,
	0xf5, 0x0b, 0xa9, 0xde, 0xdb, 0x02, 0xdc, 0xa5, 0x5c, 0x0f, 0x1c, 0x68, 0x0c, 0x27, 0xd3, 0x33,
	0xfe, 0xaa, 0x9d, 0xae, 0x2f, 0x09, 0xda, 0x83, 0xfa, 0x70, 0x32, 0xfd, 0x38, 0x75, 0x8e, 0x9d,
	0xf1, 0x7b, 0xa7, 0xf9, 0x17, 0x02, 0xa8, 0x38, 0x67, 0xb6, 0x65, 0x9b, 0xcd, 0x02, 0xb3, 0xb1,
	0x69, 0xf5, 0xc7, 0x4e, 0xb3, 0x88, 0x6a, 0x50, 0x1a, 0x98, 0xc7, 0xfd, 0xe6, 0x9c, 0x47, 0xfb,
	0xa3, 0xf1, 0x69, 0xbf, 0xb9, 0x98, 0x55, 0xf8, 0xdf, 0xf9, 0xe9, 0xcf, 0x00, 0x00, 0x00, 0xff,
	0xff, 0x15, 0x42, 0xcb, 0xdb, 0xcb, 0x07, 0x00, 0x00,
}

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
)

var GPUVendorType_name = map[int32]string{
	0: "GPU_UNKNOWN",
	1: "NVIDIA",
	2: "RADEON",
}
var GPUVendorType_value = map[string]int32{
	"GPU_UNKNOWN": 0,
	"NVIDIA":      1,
	"RADEON":      2,
}

func (x GPUVendorType) String() string {
	return proto.EnumName(GPUVendorType_name, int32(x))
}
func (GPUVendorType) EnumDescriptor() ([]byte, []int) { return fileDescriptor4, []int{0} }

// Deprecated: TODO: no longer used.
type Capabilities struct {
	Cpu []*CPUDevice `protobuf:"bytes,1,rep,name=cpu" json:"cpu,omitempty"`
	Mem *RAMDevice   `protobuf:"bytes,2,opt,name=mem" json:"mem,omitempty"`
	Gpu []*GPUDevice `protobuf:"bytes,3,rep,name=gpu" json:"gpu,omitempty"`
}

func (m *Capabilities) Reset()                    { *m = Capabilities{} }
func (m *Capabilities) String() string            { return proto.CompactTextString(m) }
func (*Capabilities) ProtoMessage()               {}
func (*Capabilities) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{0} }

func (m *Capabilities) GetCpu() []*CPUDevice {
	if m != nil {
		return m.Cpu
	}
	return nil
}

func (m *Capabilities) GetMem() *RAMDevice {
	if m != nil {
		return m.Mem
	}
	return nil
}

func (m *Capabilities) GetGpu() []*GPUDevice {
	if m != nil {
		return m.Gpu
	}
	return nil
}

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
func (*CPUDevice) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{1} }

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
func (*CPU) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{2} }

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
func (*RAMDevice) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{3} }

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
func (*RAM) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{4} }

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
}

func (m *GPUDevice) Reset()                    { *m = GPUDevice{} }
func (m *GPUDevice) String() string            { return proto.CompactTextString(m) }
func (*GPUDevice) ProtoMessage()               {}
func (*GPUDevice) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{5} }

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

type GPU struct {
	Device     *GPUDevice            `protobuf:"bytes,1,opt,name=device" json:"device,omitempty"`
	Benchmarks map[uint64]*Benchmark `protobuf:"bytes,2,rep,name=benchmarks" json:"benchmarks,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *GPU) Reset()                    { *m = GPU{} }
func (m *GPU) String() string            { return proto.CompactTextString(m) }
func (*GPU) ProtoMessage()               {}
func (*GPU) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{6} }

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

type NetworkDevice struct {
	BandwidthIn  uint64 `protobuf:"varint,1,opt,name=bandwidthIn" json:"bandwidthIn,omitempty"`
	BandwidthOut uint64 `protobuf:"varint,2,opt,name=bandwidthOut" json:"bandwidthOut,omitempty"`
}

func (m *NetworkDevice) Reset()                    { *m = NetworkDevice{} }
func (m *NetworkDevice) String() string            { return proto.CompactTextString(m) }
func (*NetworkDevice) ProtoMessage()               {}
func (*NetworkDevice) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{7} }

func (m *NetworkDevice) GetBandwidthIn() uint64 {
	if m != nil {
		return m.BandwidthIn
	}
	return 0
}

func (m *NetworkDevice) GetBandwidthOut() uint64 {
	if m != nil {
		return m.BandwidthOut
	}
	return 0
}

type Network struct {
	Device     *NetworkDevice        `protobuf:"bytes,1,opt,name=device" json:"device,omitempty"`
	Benchmarks map[uint64]*Benchmark `protobuf:"bytes,2,rep,name=benchmarks" json:"benchmarks,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *Network) Reset()                    { *m = Network{} }
func (m *Network) String() string            { return proto.CompactTextString(m) }
func (*Network) ProtoMessage()               {}
func (*Network) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{8} }

func (m *Network) GetDevice() *NetworkDevice {
	if m != nil {
		return m.Device
	}
	return nil
}

func (m *Network) GetBenchmarks() map[uint64]*Benchmark {
	if m != nil {
		return m.Benchmarks
	}
	return nil
}

type StorageDevice struct {
	BytesAvailable uint64 `protobuf:"varint,1,opt,name=bytesAvailable" json:"bytesAvailable,omitempty"`
}

func (m *StorageDevice) Reset()                    { *m = StorageDevice{} }
func (m *StorageDevice) String() string            { return proto.CompactTextString(m) }
func (*StorageDevice) ProtoMessage()               {}
func (*StorageDevice) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{9} }

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
func (*Storage) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{10} }

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
	proto.RegisterType((*Capabilities)(nil), "sonm.Capabilities")
	proto.RegisterType((*CPUDevice)(nil), "sonm.CPUDevice")
	proto.RegisterType((*CPU)(nil), "sonm.CPU")
	proto.RegisterType((*RAMDevice)(nil), "sonm.RAMDevice")
	proto.RegisterType((*RAM)(nil), "sonm.RAM")
	proto.RegisterType((*GPUDevice)(nil), "sonm.GPUDevice")
	proto.RegisterType((*GPU)(nil), "sonm.GPU")
	proto.RegisterType((*NetworkDevice)(nil), "sonm.NetworkDevice")
	proto.RegisterType((*Network)(nil), "sonm.Network")
	proto.RegisterType((*StorageDevice)(nil), "sonm.StorageDevice")
	proto.RegisterType((*Storage)(nil), "sonm.Storage")
	proto.RegisterEnum("sonm.GPUVendorType", GPUVendorType_name, GPUVendorType_value)
}

func init() { proto.RegisterFile("capabilities.proto", fileDescriptor4) }

var fileDescriptor4 = []byte{
	// 596 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x55, 0x5d, 0x8b, 0xd3, 0x40,
	0x14, 0x35, 0xc9, 0x6e, 0x77, 0x73, 0xbb, 0xdd, 0x96, 0x51, 0x24, 0x16, 0x95, 0x1a, 0x50, 0x17,
	0x85, 0x3e, 0xe8, 0x83, 0xab, 0x20, 0x12, 0x9b, 0x25, 0x04, 0x69, 0x5a, 0xa2, 0x59, 0xf1, 0x49,
	0x26, 0xe9, 0xd0, 0x8d, 0x4d, 0x32, 0x21, 0x1f, 0x2d, 0xfd, 0x69, 0xfe, 0x00, 0x7f, 0x8f, 0x7f,
	0x41, 0x32, 0x99, 0x26, 0x69, 0xd9, 0xc2, 0xfa, 0xd2, 0xb7, 0xb9, 0xe7, 0x9e, 0x73, 0xa6, 0xf7,
	0xcc, 0x0d, 0x05, 0xe4, 0xe1, 0x18, 0xbb, 0x7e, 0xe0, 0x67, 0x3e, 0x49, 0x87, 0x71, 0x42, 0x33,
	0x8a, 0x8e, 0x52, 0x1a, 0x85, 0xfd, 0x9e, 0x4b, 0x22, 0xef, 0x26, 0xc4, 0xc9, 0x82, 0xe3, 0xea,
	0x0a, 0xce, 0x46, 0x0d, 0x36, 0x7a, 0x06, 0x92, 0x17, 0xe7, 0x8a, 0x30, 0x90, 0x2e, 0xda, 0x6f,
	0xba, 0xc3, 0x42, 0x35, 0x1c, 0x4d, 0x1d, 0x9d, 0x2c, 0x7d, 0x8f, 0xd8, 0x45, 0xaf, 0xa0, 0x84,
	0x24, 0x54, 0xc4, 0x81, 0x50, 0x53, 0x6c, 0x6d, 0xbc, 0xa1, 0x84, 0x24, 0x2c, 0x28, 0xf3, 0x38,
	0x57, 0xa4, 0xa6, 0x8b, 0x51, 0xbb, 0xcc, 0xe3, 0x5c, 0xfd, 0x01, 0x72, 0xe5, 0x8b, 0x1e, 0x83,
	0x1c, 0xd2, 0x19, 0x09, 0x2c, 0x1c, 0x12, 0x45, 0x18, 0x08, 0x17, 0xb2, 0x5d, 0x03, 0xe8, 0x01,
	0x1c, 0x7b, 0x34, 0x21, 0x29, 0xbb, 0xb2, 0x63, 0x97, 0x05, 0x52, 0xe0, 0x24, 0xa5, 0xde, 0x82,
	0x64, 0xa9, 0x22, 0x31, 0x7c, 0x53, 0xaa, 0xbf, 0x05, 0x90, 0x46, 0x53, 0x07, 0xbd, 0x84, 0xd6,
	0x8c, 0xf9, 0x33, 0xcb, 0x5b, 0xc6, 0xe1, 0x6d, 0xf4, 0x1e, 0xa0, 0x0e, 0x46, 0x11, 0xd9, 0xaf,
	0x7e, 0x54, 0x91, 0x87, 0x9f, 0xab, 0xde, 0x55, 0x94, 0x25, 0x6b, 0xbb, 0x41, 0xee, 0x5b, 0xd0,
	0xdd, 0x69, 0xa3, 0x1e, 0x48, 0x0b, 0xb2, 0x66, 0x77, 0x1e, 0xd9, 0xc5, 0x11, 0x3d, 0x87, 0xe3,
	0x25, 0x0e, 0x72, 0xb2, 0x9d, 0x59, 0xa5, 0xb3, 0xcb, 0xee, 0x07, 0xf1, 0x52, 0x50, 0x3f, 0x81,
	0x5c, 0x65, 0x59, 0x0c, 0x9e, 0xd1, 0x0c, 0x07, 0xdc, 0xab, 0x2c, 0x8a, 0xb0, 0xf0, 0x12, 0xfb,
	0x01, 0x76, 0x83, 0xd2, 0xf1, 0xc8, 0xae, 0x01, 0x36, 0xbc, 0xad, 0x8d, 0xf7, 0x0d, 0x5f, 0x3f,
	0xd4, 0x1d, 0x86, 0xb7, 0xb5, 0xf1, 0x41, 0x87, 0xff, 0x2b, 0x80, 0x5c, 0xad, 0x09, 0x3a, 0x07,
	0xd1, 0xd4, 0xf9, 0x36, 0x88, 0xa6, 0x8e, 0xfa, 0x70, 0xba, 0x24, 0xd1, 0x8c, 0x26, 0xa6, 0xce,
	0xc7, 0xae, 0x6a, 0xf4, 0x14, 0xa0, 0x3c, 0xb3, 0x0d, 0x92, 0x98, 0xa6, 0x81, 0x14, 0xda, 0x72,
	0x5c, 0x53, 0x57, 0x8e, 0x4b, 0xed, 0xa6, 0x2e, 0xb4, 0xe5, 0x99, 0x69, 0x5b, 0xa5, 0xb6, 0x46,
	0xd0, 0x00, 0xda, 0x21, 0xfe, 0x45, 0x13, 0x2b, 0x0f, 0x5d, 0x92, 0x28, 0x27, 0x4c, 0xde, 0x84,
	0x18, 0xc3, 0x8f, 0x2a, 0xc6, 0x29, 0x67, 0xd4, 0x10, 0x7a, 0x08, 0xad, 0x31, 0x09, 0x69, 0xb2,
	0x56, 0x64, 0xd6, 0xe4, 0x15, 0x7b, 0x2d, 0x63, 0xff, 0xaa, 0x1a, 0xff, 0xb3, 0xaa, 0xc6, 0x81,
	0x57, 0xd5, 0x81, 0x8e, 0x45, 0xb2, 0x15, 0x4d, 0x16, 0xfc, 0xc1, 0x06, 0xd0, 0x76, 0x71, 0x34,
	0x5b, 0xf9, 0xb3, 0xec, 0xc6, 0x8c, 0xb8, 0x6b, 0x13, 0x42, 0x2a, 0x9c, 0x55, 0xe5, 0x24, 0xcf,
	0xf8, 0x33, 0x6e, 0x61, 0xea, 0x1f, 0x01, 0x4e, 0xb8, 0x2f, 0x7a, 0xbd, 0x13, 0xcb, 0xfd, 0xf2,
	0xe7, 0x6c, 0x5d, 0x5b, 0x45, 0xf3, 0xf1, 0x96, 0x68, 0x9e, 0x6c, 0x09, 0x0e, 0x1a, 0xcf, 0x3b,
	0xe8, 0x7c, 0xcd, 0x68, 0x82, 0xe7, 0x84, 0xc7, 0xf3, 0x02, 0xce, 0xdd, 0x75, 0x46, 0x52, 0xad,
	0xfa, 0x78, 0x4b, 0xe3, 0x1d, 0x94, 0x05, 0xc0, 0x95, 0xfb, 0x02, 0xd8, 0x32, 0xbe, 0x4b, 0x00,
	0x5c, 0x70, 0xc8, 0x00, 0x5e, 0x5d, 0x42, 0xc7, 0x98, 0x3a, 0xd7, 0xec, 0x23, 0xfc, 0xb6, 0x8e,
	0x09, 0xea, 0x42, 0xdb, 0x98, 0x3a, 0x3f, 0x1d, 0xeb, 0x8b, 0x35, 0xf9, 0x6e, 0xf5, 0xee, 0x21,
	0x80, 0x96, 0x75, 0x6d, 0xea, 0xa6, 0xd6, 0x13, 0x8a, 0xb3, 0xad, 0xe9, 0x57, 0x13, 0xab, 0x27,
	0xba, 0x2d, 0xf6, 0xdf, 0xf4, 0xf6, 0x5f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x13, 0xf9, 0x24, 0xad,
	0xc9, 0x06, 0x00, 0x00,
}

// Code generated by protoc-gen-go. DO NOT EDIT.
// source: volume.proto

package sonm

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Volume describes volume settings.
// One may notice - why the Hell should we describe an entire message with
// a single field? Map of maps - that's why.
type Volume struct {
	// Driver describes a volume driver.
	Driver string `protobuf:"bytes,1,opt,name=driver" json:"driver,omitempty"`
	// Settings describes a place for your volume settings.
	Settings map[string]string `protobuf:"bytes,2,rep,name=settings" json:"settings,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *Volume) Reset()                    { *m = Volume{} }
func (m *Volume) String() string            { return proto.CompactTextString(m) }
func (*Volume) ProtoMessage()               {}
func (*Volume) Descriptor() ([]byte, []int) { return fileDescriptor13, []int{0} }

func (m *Volume) GetDriver() string {
	if m != nil {
		return m.Driver
	}
	return ""
}

func (m *Volume) GetSettings() map[string]string {
	if m != nil {
		return m.Settings
	}
	return nil
}

func init() {
	proto.RegisterType((*Volume)(nil), "sonm.Volume")
}

func init() { proto.RegisterFile("volume.proto", fileDescriptor13) }

var fileDescriptor13 = []byte{
	// 149 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x29, 0xcb, 0xcf, 0x29,
	0xcd, 0x4d, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x29, 0xce, 0xcf, 0xcb, 0x55, 0x9a,
	0xca, 0xc8, 0xc5, 0x16, 0x06, 0x16, 0x16, 0x12, 0xe3, 0x62, 0x4b, 0x29, 0xca, 0x2c, 0x4b, 0x2d,
	0x92, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x0c, 0x82, 0xf2, 0x84, 0xcc, 0xb8, 0x38, 0x8a, 0x53, 0x4b,
	0x4a, 0x32, 0xf3, 0xd2, 0x8b, 0x25, 0x98, 0x14, 0x98, 0x35, 0xb8, 0x8d, 0xa4, 0xf4, 0x40, 0x7a,
	0xf5, 0x20, 0xfa, 0xf4, 0x82, 0xa1, 0x92, 0xae, 0x79, 0x25, 0x45, 0x95, 0x41, 0x70, 0xb5, 0x52,
	0xd6, 0x5c, 0xbc, 0x28, 0x52, 0x42, 0x02, 0x5c, 0xcc, 0xd9, 0xa9, 0x95, 0x50, 0xd3, 0x41, 0x4c,
	0x21, 0x11, 0x2e, 0xd6, 0xb2, 0xc4, 0x9c, 0xd2, 0x54, 0x09, 0x26, 0xb0, 0x18, 0x84, 0x63, 0xc5,
	0x64, 0xc1, 0x98, 0xc4, 0x06, 0x76, 0xa4, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0xd0, 0x3b, 0x90,
	0x7c, 0xb4, 0x00, 0x00, 0x00,
}

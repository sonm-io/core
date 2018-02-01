// Code generated by protoc-gen-go. DO NOT EDIT.
// source: nat.proto

package sonm

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type NATType int32

const (
	NATType_NONE                   NATType = 0
	NATType_BLOCKED                NATType = 1
	NATType_FULL                   NATType = 2
	NATType_SYMMETRIC              NATType = 3
	NATType_RESTRICTED             NATType = 4
	NATType_PORT_RESTRICTED        NATType = 5
	NATType_SYMMETRIC_UDP_FIREWALL NATType = 6
	NATType_UNKNOWN                NATType = 7
)

var NATType_name = map[int32]string{
	0: "NONE",
	1: "BLOCKED",
	2: "FULL",
	3: "SYMMETRIC",
	4: "RESTRICTED",
	5: "PORT_RESTRICTED",
	6: "SYMMETRIC_UDP_FIREWALL",
	7: "UNKNOWN",
}
var NATType_value = map[string]int32{
	"NONE":                   0,
	"BLOCKED":                1,
	"FULL":                   2,
	"SYMMETRIC":              3,
	"RESTRICTED":             4,
	"PORT_RESTRICTED":        5,
	"SYMMETRIC_UDP_FIREWALL": 6,
	"UNKNOWN":                7,
}

func (x NATType) String() string {
	return proto.EnumName(NATType_name, int32(x))
}
func (NATType) EnumDescriptor() ([]byte, []int) { return fileDescriptor10, []int{0} }

func init() {
	proto.RegisterEnum("sonm.NATType", NATType_name, NATType_value)
}

func init() { proto.RegisterFile("nat.proto", fileDescriptor10) }

var fileDescriptor10 = []byte{
	// 162 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xcc, 0x4b, 0x2c, 0xd1,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x29, 0xce, 0xcf, 0xcb, 0xd5, 0x6a, 0x67, 0xe4, 0x62,
	0xf7, 0x73, 0x0c, 0x09, 0xa9, 0x2c, 0x48, 0x15, 0xe2, 0xe0, 0x62, 0xf1, 0xf3, 0xf7, 0x73, 0x15,
	0x60, 0x10, 0xe2, 0xe6, 0x62, 0x77, 0xf2, 0xf1, 0x77, 0xf6, 0x76, 0x75, 0x11, 0x60, 0x04, 0x09,
	0xbb, 0x85, 0xfa, 0xf8, 0x08, 0x30, 0x09, 0xf1, 0x72, 0x71, 0x06, 0x47, 0xfa, 0xfa, 0xba, 0x86,
	0x04, 0x79, 0x3a, 0x0b, 0x30, 0x0b, 0xf1, 0x71, 0x71, 0x05, 0xb9, 0x06, 0x83, 0x38, 0x21, 0xae,
	0x2e, 0x02, 0x2c, 0x42, 0xc2, 0x5c, 0xfc, 0x01, 0xfe, 0x41, 0x21, 0xf1, 0x48, 0x82, 0xac, 0x42,
	0x52, 0x5c, 0x62, 0x70, 0x3d, 0xf1, 0xa1, 0x2e, 0x01, 0xf1, 0x6e, 0x9e, 0x41, 0xae, 0xe1, 0x8e,
	0x3e, 0x3e, 0x02, 0x6c, 0x20, 0x6b, 0x42, 0xfd, 0xbc, 0xfd, 0xfc, 0xc3, 0xfd, 0x04, 0xd8, 0x93,
	0xd8, 0xc0, 0xce, 0x32, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0xb6, 0x73, 0x98, 0x37, 0xa3, 0x00,
	0x00, 0x00,
}

// Code generated by protoc-gen-go. DO NOT EDIT.
// source: deal.proto

package sonm

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Deprecated: migrate to new marketplace API.
type DealStatus int32

const (
	DealStatus_ANY_STATUS DealStatus = 0
	DealStatus_PENDING    DealStatus = 1
	DealStatus_ACCEPTED   DealStatus = 2
	DealStatus_CLOSED     DealStatus = 3
)

var DealStatus_name = map[int32]string{
	0: "ANY_STATUS",
	1: "PENDING",
	2: "ACCEPTED",
	3: "CLOSED",
}
var DealStatus_value = map[string]int32{
	"ANY_STATUS": 0,
	"PENDING":    1,
	"ACCEPTED":   2,
	"CLOSED":     3,
}

func (x DealStatus) String() string {
	return proto.EnumName(DealStatus_name, int32(x))
}
func (DealStatus) EnumDescriptor() ([]byte, []int) { return fileDescriptor6, []int{0} }

// Deprecated: migrate to new marketplace API.
type Deal struct {
	BuyerID           string     `protobuf:"bytes,1,opt,name=BuyerID" json:"BuyerID,omitempty"`
	SupplierID        string     `protobuf:"bytes,2,opt,name=SupplierID" json:"SupplierID,omitempty"`
	Status            DealStatus `protobuf:"varint,3,opt,name=status,enum=sonm.DealStatus" json:"status,omitempty"`
	Price             *BigInt    `protobuf:"bytes,4,opt,name=price" json:"price,omitempty"`
	StartTime         *Timestamp `protobuf:"bytes,5,opt,name=startTime" json:"startTime,omitempty"`
	EndTime           *Timestamp `protobuf:"bytes,6,opt,name=endTime" json:"endTime,omitempty"`
	SpecificationHash string     `protobuf:"bytes,7,opt,name=SpecificationHash" json:"SpecificationHash,omitempty"`
	WorkTime          uint64     `protobuf:"varint,8,opt,name=workTime" json:"workTime,omitempty"`
	Id                string     `protobuf:"bytes,9,opt,name=id" json:"id,omitempty"`
}

func (m *Deal) Reset()                    { *m = Deal{} }
func (m *Deal) String() string            { return proto.CompactTextString(m) }
func (*Deal) ProtoMessage()               {}
func (*Deal) Descriptor() ([]byte, []int) { return fileDescriptor6, []int{0} }

func (m *Deal) GetBuyerID() string {
	if m != nil {
		return m.BuyerID
	}
	return ""
}

func (m *Deal) GetSupplierID() string {
	if m != nil {
		return m.SupplierID
	}
	return ""
}

func (m *Deal) GetStatus() DealStatus {
	if m != nil {
		return m.Status
	}
	return DealStatus_ANY_STATUS
}

func (m *Deal) GetPrice() *BigInt {
	if m != nil {
		return m.Price
	}
	return nil
}

func (m *Deal) GetStartTime() *Timestamp {
	if m != nil {
		return m.StartTime
	}
	return nil
}

func (m *Deal) GetEndTime() *Timestamp {
	if m != nil {
		return m.EndTime
	}
	return nil
}

func (m *Deal) GetSpecificationHash() string {
	if m != nil {
		return m.SpecificationHash
	}
	return ""
}

func (m *Deal) GetWorkTime() uint64 {
	if m != nil {
		return m.WorkTime
	}
	return 0
}

func (m *Deal) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func init() {
	proto.RegisterType((*Deal)(nil), "sonm.Deal")
	proto.RegisterEnum("sonm.DealStatus", DealStatus_name, DealStatus_value)
}

func init() { proto.RegisterFile("deal.proto", fileDescriptor6) }

var fileDescriptor6 = []byte{
	// 315 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x91, 0xcd, 0x4e, 0xf2, 0x40,
	0x14, 0x86, 0xbf, 0x96, 0xd2, 0xc2, 0x81, 0x40, 0xbf, 0xb3, 0x9a, 0xb0, 0x30, 0x0d, 0xab, 0x6a,
	0x94, 0x05, 0x5e, 0x01, 0xb4, 0x8d, 0x92, 0x18, 0x24, 0x6d, 0x5d, 0xb8, 0x32, 0x03, 0x1d, 0xf1,
	0x44, 0xfa, 0x93, 0xce, 0x10, 0xe3, 0xed, 0x79, 0x65, 0x86, 0xa9, 0x88, 0x89, 0x2e, 0xcf, 0xfb,
	0x3c, 0x73, 0x66, 0xde, 0x0c, 0x40, 0x26, 0xf8, 0x6e, 0x52, 0xd5, 0xa5, 0x2a, 0xd1, 0x92, 0x65,
	0x91, 0x8f, 0xfa, 0x6b, 0xda, 0x52, 0xa1, 0x9a, 0x6c, 0x34, 0x54, 0x94, 0x0b, 0xa9, 0x78, 0x5e,
	0x35, 0xc1, 0xf8, 0xc3, 0x04, 0x2b, 0x14, 0x7c, 0x87, 0x0c, 0x9c, 0xf9, 0xfe, 0x5d, 0xd4, 0x8b,
	0x90, 0x19, 0x9e, 0xe1, 0x77, 0xe3, 0xe3, 0x88, 0x67, 0x00, 0xc9, 0xbe, 0xaa, 0x76, 0xa4, 0xa1,
	0xa9, 0xe1, 0x8f, 0x04, 0x7d, 0xb0, 0xa5, 0xe2, 0x6a, 0x2f, 0x59, 0xcb, 0x33, 0xfc, 0xc1, 0xd4,
	0x9d, 0x1c, 0x2e, 0x9e, 0x1c, 0xb6, 0x26, 0x3a, 0x8f, 0xbf, 0x38, 0x8e, 0xa1, 0x5d, 0xd5, 0xb4,
	0x11, 0xcc, 0xf2, 0x0c, 0xbf, 0x37, 0xed, 0x37, 0xe2, 0x9c, 0xb6, 0x8b, 0x42, 0xc5, 0x0d, 0xc2,
	0x2b, 0xe8, 0x4a, 0xc5, 0x6b, 0x95, 0x52, 0x2e, 0x58, 0x5b, 0x7b, 0xc3, 0xc6, 0x4b, 0x8f, 0x4f,
	0x8f, 0x4f, 0x06, 0x9e, 0x83, 0x23, 0x8a, 0x4c, 0xcb, 0xf6, 0xdf, 0xf2, 0x91, 0xe3, 0x25, 0xfc,
	0x4f, 0x2a, 0xb1, 0xa1, 0x67, 0xda, 0x70, 0x45, 0x65, 0x71, 0xcb, 0xe5, 0x0b, 0x73, 0x74, 0x9d,
	0xdf, 0x00, 0x47, 0xd0, 0x79, 0x2b, 0xeb, 0x57, 0xbd, 0xb9, 0xe3, 0x19, 0xbe, 0x15, 0x7f, 0xcf,
	0x38, 0x00, 0x93, 0x32, 0xd6, 0xd5, 0x47, 0x4d, 0xca, 0x2e, 0x02, 0x80, 0x53, 0x5b, 0x1c, 0x00,
	0xcc, 0x96, 0x8f, 0x4f, 0x49, 0x3a, 0x4b, 0x1f, 0x12, 0xf7, 0x1f, 0xf6, 0xc0, 0x59, 0x45, 0xcb,
	0x70, 0xb1, 0xbc, 0x71, 0x0d, 0xec, 0x43, 0x67, 0x16, 0x04, 0xd1, 0x2a, 0x8d, 0x42, 0xd7, 0x44,
	0x00, 0x3b, 0xb8, 0xbb, 0x4f, 0xa2, 0xd0, 0x6d, 0xad, 0x6d, 0xfd, 0x21, 0xd7, 0x9f, 0x01, 0x00,
	0x00, 0xff, 0xff, 0x5b, 0x3d, 0x5b, 0x19, 0xc3, 0x01, 0x00, 0x00,
}

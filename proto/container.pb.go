// Code generated by protoc-gen-go. DO NOT EDIT.
// source: container.proto

package sonm

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Container struct {
	// Image describes a Docker image name. Required.
	Image string `protobuf:"bytes,1,opt,name=image" json:"image,omitempty"`
	// Registry describes Docker registry.
	Registry string `protobuf:"bytes,2,opt,name=registry" json:"registry,omitempty"`
	// Auth describes authentication info used for registry.
	Auth string `protobuf:"bytes,3,opt,name=auth" json:"auth,omitempty"`
	// SSH public key used to attach to the container.
	PublicKeyData string `protobuf:"bytes,4,opt,name=publicKeyData" json:"publicKeyData,omitempty"`
	// CommitOnStop points whether a container should commit when stopped.
	// Committed containers can be fetched later while there is an active
	// deal.
	CommitOnStop bool `protobuf:"varint,5,opt,name=commitOnStop" json:"commitOnStop,omitempty"`
	// Env describes environment variables forwarded into the container.
	Env map[string]string `protobuf:"bytes,7,rep,name=env" json:"env,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// Volumes describes network volumes that are used to be mounted inside
	// the container.
	// Mapping from the volume type (cifs, nfs, etc.) to its settings.
	Volumes map[string]*Volume `protobuf:"bytes,8,rep,name=volumes" json:"volumes,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// Mounts describes mount points from the volume name to the container.
	// TODO: Dragons nearby - beware of injection attacks.
	Mounts []string `protobuf:"bytes,9,rep,name=mounts" json:"mounts,omitempty"`
	Expose []string `protobuf:"bytes,10,rep,name=expose" json:"expose,omitempty"`
}

func (m *Container) Reset()                    { *m = Container{} }
func (m *Container) String() string            { return proto.CompactTextString(m) }
func (*Container) ProtoMessage()               {}
func (*Container) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{0} }

func (m *Container) GetImage() string {
	if m != nil {
		return m.Image
	}
	return ""
}

func (m *Container) GetRegistry() string {
	if m != nil {
		return m.Registry
	}
	return ""
}

func (m *Container) GetAuth() string {
	if m != nil {
		return m.Auth
	}
	return ""
}

func (m *Container) GetPublicKeyData() string {
	if m != nil {
		return m.PublicKeyData
	}
	return ""
}

func (m *Container) GetCommitOnStop() bool {
	if m != nil {
		return m.CommitOnStop
	}
	return false
}

func (m *Container) GetEnv() map[string]string {
	if m != nil {
		return m.Env
	}
	return nil
}

func (m *Container) GetVolumes() map[string]*Volume {
	if m != nil {
		return m.Volumes
	}
	return nil
}

func (m *Container) GetMounts() []string {
	if m != nil {
		return m.Mounts
	}
	return nil
}

func (m *Container) GetExpose() []string {
	if m != nil {
		return m.Expose
	}
	return nil
}

func init() {
	proto.RegisterType((*Container)(nil), "sonm.Container")
}

func init() { proto.RegisterFile("container.proto", fileDescriptor3) }

var fileDescriptor3 = []byte{
	// 295 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x51, 0x41, 0x6b, 0xb4, 0x30,
	0x14, 0xc4, 0x8d, 0xbb, 0xab, 0x6f, 0xfd, 0xf8, 0xca, 0xa3, 0x94, 0x20, 0x3d, 0x88, 0xf4, 0x20,
	0x3d, 0x78, 0xd8, 0xc2, 0x52, 0x7a, 0x6d, 0x17, 0x0a, 0x3d, 0x14, 0x2c, 0xf4, 0x9e, 0x95, 0xb0,
	0x95, 0x9a, 0x44, 0x34, 0x4a, 0xfd, 0x59, 0xfd, 0x87, 0xc5, 0x44, 0x17, 0xb7, 0xf4, 0xf6, 0x66,
	0xde, 0x4c, 0x32, 0x99, 0xc0, 0xff, 0x5c, 0x49, 0xcd, 0x0a, 0xc9, 0xeb, 0xb4, 0xaa, 0x95, 0x56,
	0xe8, 0x36, 0x4a, 0x8a, 0x30, 0xe8, 0x54, 0xd9, 0x0a, 0x6e, 0xb9, 0xf8, 0x9b, 0x80, 0xff, 0x38,
	0xe9, 0xf0, 0x12, 0x96, 0x85, 0x60, 0x47, 0x4e, 0x9d, 0xc8, 0x49, 0xfc, 0xcc, 0x02, 0x0c, 0xc1,
	0xab, 0xf9, 0xb1, 0x68, 0x74, 0xdd, 0xd3, 0x85, 0x59, 0x9c, 0x30, 0x22, 0xb8, 0xac, 0xd5, 0x1f,
	0x94, 0x18, 0xde, 0xcc, 0x78, 0x03, 0xff, 0xaa, 0xf6, 0x50, 0x16, 0xf9, 0x0b, 0xef, 0x9f, 0x98,
	0x66, 0xd4, 0x35, 0xcb, 0x73, 0x12, 0x63, 0x08, 0x72, 0x25, 0x44, 0xa1, 0x5f, 0xe5, 0x9b, 0x56,
	0x15, 0x5d, 0x46, 0x4e, 0xe2, 0x65, 0x67, 0x1c, 0xde, 0x02, 0xe1, 0xb2, 0xa3, 0xeb, 0x88, 0x24,
	0x9b, 0x2d, 0x4d, 0x87, 0xfc, 0xe9, 0x29, 0x6d, 0xba, 0x97, 0xdd, 0x5e, 0xea, 0xba, 0xcf, 0x06,
	0x11, 0xee, 0x60, 0x6d, 0x5f, 0xd6, 0x50, 0xcf, 0xe8, 0xaf, 0x7f, 0xeb, 0xdf, 0xed, 0xda, 0x7a,
	0x26, 0x31, 0x5e, 0xc1, 0x4a, 0xa8, 0x56, 0xea, 0x86, 0xfa, 0x11, 0x49, 0xfc, 0x6c, 0x44, 0x03,
	0xcf, 0xbf, 0x2a, 0xd5, 0x70, 0x0a, 0x96, 0xb7, 0x28, 0xdc, 0x81, 0x37, 0x5d, 0x8c, 0x17, 0x40,
	0x3e, 0x79, 0x3f, 0xb6, 0x35, 0x8c, 0x43, 0x83, 0x1d, 0x2b, 0x5b, 0x3e, 0x16, 0x65, 0xc1, 0xc3,
	0xe2, 0xde, 0x09, 0x9f, 0x21, 0x98, 0x07, 0xf8, 0xc3, 0x1b, 0xcf, 0xbd, 0x9b, 0x6d, 0x60, 0xf3,
	0x5b, 0xd3, 0xec, 0xa4, 0xc3, 0xca, 0x7c, 0xdd, 0xdd, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xd3,
	0xd5, 0x14, 0x07, 0xe1, 0x01, 0x00, 0x00,
}

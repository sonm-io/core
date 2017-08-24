// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proto/insonmnia.proto

package sonm

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type TaskStatusReply_Status int32

const (
	TaskStatusReply_UNKNOWN  TaskStatusReply_Status = 0
	TaskStatusReply_SPOOLING TaskStatusReply_Status = 1
	TaskStatusReply_SPAWNING TaskStatusReply_Status = 2
	TaskStatusReply_RUNNING  TaskStatusReply_Status = 3
	TaskStatusReply_FINISHED TaskStatusReply_Status = 4
	TaskStatusReply_BROKEN   TaskStatusReply_Status = 5
)

var TaskStatusReply_Status_name = map[int32]string{
	0: "UNKNOWN",
	1: "SPOOLING",
	2: "SPAWNING",
	3: "RUNNING",
	4: "FINISHED",
	5: "BROKEN",
}
var TaskStatusReply_Status_value = map[string]int32{
	"UNKNOWN":  0,
	"SPOOLING": 1,
	"SPAWNING": 2,
	"RUNNING":  3,
	"FINISHED": 4,
	"BROKEN":   5,
}

func (x TaskStatusReply_Status) String() string {
	return proto.EnumName(TaskStatusReply_Status_name, int32(x))
}
func (TaskStatusReply_Status) EnumDescriptor() ([]byte, []int) { return fileDescriptor2, []int{10, 0} }

type TaskLogsRequest_Type int32

const (
	TaskLogsRequest_STDOUT TaskLogsRequest_Type = 0
	TaskLogsRequest_STDERR TaskLogsRequest_Type = 1
	TaskLogsRequest_BOTH   TaskLogsRequest_Type = 2
)

var TaskLogsRequest_Type_name = map[int32]string{
	0: "STDOUT",
	1: "STDERR",
	2: "BOTH",
}
var TaskLogsRequest_Type_value = map[string]int32{
	"STDOUT": 0,
	"STDERR": 1,
	"BOTH":   2,
}

func (x TaskLogsRequest_Type) String() string {
	return proto.EnumName(TaskLogsRequest_Type_name, int32(x))
}
func (TaskLogsRequest_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptor2, []int{14, 0} }

type PingRequest struct {
}

func (m *PingRequest) Reset()                    { *m = PingRequest{} }
func (m *PingRequest) String() string            { return proto.CompactTextString(m) }
func (*PingRequest) ProtoMessage()               {}
func (*PingRequest) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{0} }

type PingReply struct {
	Status string `protobuf:"bytes,1,opt,name=status" json:"status,omitempty"`
}

func (m *PingReply) Reset()                    { *m = PingReply{} }
func (m *PingReply) String() string            { return proto.CompactTextString(m) }
func (*PingReply) ProtoMessage()               {}
func (*PingReply) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{1} }

func (m *PingReply) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

type CPUUsage struct {
	Total uint64 `protobuf:"varint,1,opt,name=total" json:"total,omitempty"`
}

func (m *CPUUsage) Reset()                    { *m = CPUUsage{} }
func (m *CPUUsage) String() string            { return proto.CompactTextString(m) }
func (*CPUUsage) ProtoMessage()               {}
func (*CPUUsage) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{2} }

func (m *CPUUsage) GetTotal() uint64 {
	if m != nil {
		return m.Total
	}
	return 0
}

type MemoryUsage struct {
	MaxUsage uint64 `protobuf:"varint,1,opt,name=maxUsage" json:"maxUsage,omitempty"`
}

func (m *MemoryUsage) Reset()                    { *m = MemoryUsage{} }
func (m *MemoryUsage) String() string            { return proto.CompactTextString(m) }
func (*MemoryUsage) ProtoMessage()               {}
func (*MemoryUsage) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{3} }

func (m *MemoryUsage) GetMaxUsage() uint64 {
	if m != nil {
		return m.MaxUsage
	}
	return 0
}

type NetworkUsage struct {
	TxBytes   uint64 `protobuf:"varint,1,opt,name=txBytes" json:"txBytes,omitempty"`
	RxBytes   uint64 `protobuf:"varint,2,opt,name=rxBytes" json:"rxBytes,omitempty"`
	TxPackets uint64 `protobuf:"varint,3,opt,name=txPackets" json:"txPackets,omitempty"`
	RxPackets uint64 `protobuf:"varint,4,opt,name=rxPackets" json:"rxPackets,omitempty"`
	TxErrors  uint64 `protobuf:"varint,5,opt,name=txErrors" json:"txErrors,omitempty"`
	RxErrors  uint64 `protobuf:"varint,6,opt,name=rxErrors" json:"rxErrors,omitempty"`
	TxDropped uint64 `protobuf:"varint,7,opt,name=txDropped" json:"txDropped,omitempty"`
	RxDropped uint64 `protobuf:"varint,8,opt,name=rxDropped" json:"rxDropped,omitempty"`
}

func (m *NetworkUsage) Reset()                    { *m = NetworkUsage{} }
func (m *NetworkUsage) String() string            { return proto.CompactTextString(m) }
func (*NetworkUsage) ProtoMessage()               {}
func (*NetworkUsage) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{4} }

func (m *NetworkUsage) GetTxBytes() uint64 {
	if m != nil {
		return m.TxBytes
	}
	return 0
}

func (m *NetworkUsage) GetRxBytes() uint64 {
	if m != nil {
		return m.RxBytes
	}
	return 0
}

func (m *NetworkUsage) GetTxPackets() uint64 {
	if m != nil {
		return m.TxPackets
	}
	return 0
}

func (m *NetworkUsage) GetRxPackets() uint64 {
	if m != nil {
		return m.RxPackets
	}
	return 0
}

func (m *NetworkUsage) GetTxErrors() uint64 {
	if m != nil {
		return m.TxErrors
	}
	return 0
}

func (m *NetworkUsage) GetRxErrors() uint64 {
	if m != nil {
		return m.RxErrors
	}
	return 0
}

func (m *NetworkUsage) GetTxDropped() uint64 {
	if m != nil {
		return m.TxDropped
	}
	return 0
}

func (m *NetworkUsage) GetRxDropped() uint64 {
	if m != nil {
		return m.RxDropped
	}
	return 0
}

type ResourceUsage struct {
	Cpu     *CPUUsage                `protobuf:"bytes,1,opt,name=cpu" json:"cpu,omitempty"`
	Memory  *MemoryUsage             `protobuf:"bytes,2,opt,name=memory" json:"memory,omitempty"`
	Network map[string]*NetworkUsage `protobuf:"bytes,3,rep,name=network" json:"network,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *ResourceUsage) Reset()                    { *m = ResourceUsage{} }
func (m *ResourceUsage) String() string            { return proto.CompactTextString(m) }
func (*ResourceUsage) ProtoMessage()               {}
func (*ResourceUsage) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{5} }

func (m *ResourceUsage) GetCpu() *CPUUsage {
	if m != nil {
		return m.Cpu
	}
	return nil
}

func (m *ResourceUsage) GetMemory() *MemoryUsage {
	if m != nil {
		return m.Memory
	}
	return nil
}

func (m *ResourceUsage) GetNetwork() map[string]*NetworkUsage {
	if m != nil {
		return m.Network
	}
	return nil
}

type MinerStatusReply struct {
	Usage        map[string]*ResourceUsage `protobuf:"bytes,1,rep,name=usage" json:"usage,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Name         string                    `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Capabilities *Capabilities             `protobuf:"bytes,3,opt,name=capabilities" json:"capabilities,omitempty"`
}

func (m *MinerStatusReply) Reset()                    { *m = MinerStatusReply{} }
func (m *MinerStatusReply) String() string            { return proto.CompactTextString(m) }
func (*MinerStatusReply) ProtoMessage()               {}
func (*MinerStatusReply) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{6} }

func (m *MinerStatusReply) GetUsage() map[string]*ResourceUsage {
	if m != nil {
		return m.Usage
	}
	return nil
}

func (m *MinerStatusReply) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *MinerStatusReply) GetCapabilities() *Capabilities {
	if m != nil {
		return m.Capabilities
	}
	return nil
}

type StopTaskRequest struct {
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *StopTaskRequest) Reset()                    { *m = StopTaskRequest{} }
func (m *StopTaskRequest) String() string            { return proto.CompactTextString(m) }
func (*StopTaskRequest) ProtoMessage()               {}
func (*StopTaskRequest) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{7} }

func (m *StopTaskRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type StopTaskReply struct {
}

func (m *StopTaskReply) Reset()                    { *m = StopTaskReply{} }
func (m *StopTaskReply) String() string            { return proto.CompactTextString(m) }
func (*StopTaskReply) ProtoMessage()               {}
func (*StopTaskReply) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{8} }

type TaskStatusRequest struct {
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *TaskStatusRequest) Reset()                    { *m = TaskStatusRequest{} }
func (m *TaskStatusRequest) String() string            { return proto.CompactTextString(m) }
func (*TaskStatusRequest) ProtoMessage()               {}
func (*TaskStatusRequest) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{9} }

func (m *TaskStatusRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type TaskStatusReply struct {
	Status    TaskStatusReply_Status `protobuf:"varint,1,opt,name=status,enum=sonm.TaskStatusReply_Status" json:"status,omitempty"`
	ImageName string                 `protobuf:"bytes,2,opt,name=imageName" json:"imageName,omitempty"`
	Ports     string                 `protobuf:"bytes,3,opt,name=ports" json:"ports,omitempty"`
	Uptime    uint64                 `protobuf:"varint,4,opt,name=uptime" json:"uptime,omitempty"`
	Usage     *ResourceUsage         `protobuf:"bytes,5,opt,name=usage" json:"usage,omitempty"`
}

func (m *TaskStatusReply) Reset()                    { *m = TaskStatusReply{} }
func (m *TaskStatusReply) String() string            { return proto.CompactTextString(m) }
func (*TaskStatusReply) ProtoMessage()               {}
func (*TaskStatusReply) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{10} }

func (m *TaskStatusReply) GetStatus() TaskStatusReply_Status {
	if m != nil {
		return m.Status
	}
	return TaskStatusReply_UNKNOWN
}

func (m *TaskStatusReply) GetImageName() string {
	if m != nil {
		return m.ImageName
	}
	return ""
}

func (m *TaskStatusReply) GetPorts() string {
	if m != nil {
		return m.Ports
	}
	return ""
}

func (m *TaskStatusReply) GetUptime() uint64 {
	if m != nil {
		return m.Uptime
	}
	return 0
}

func (m *TaskStatusReply) GetUsage() *ResourceUsage {
	if m != nil {
		return m.Usage
	}
	return nil
}

type StatusMapReply struct {
	Statuses map[string]*TaskStatusReply `protobuf:"bytes,1,rep,name=statuses" json:"statuses,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *StatusMapReply) Reset()                    { *m = StatusMapReply{} }
func (m *StatusMapReply) String() string            { return proto.CompactTextString(m) }
func (*StatusMapReply) ProtoMessage()               {}
func (*StatusMapReply) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{11} }

func (m *StatusMapReply) GetStatuses() map[string]*TaskStatusReply {
	if m != nil {
		return m.Statuses
	}
	return nil
}

type ContainerResources struct {
	Memory   int64 `protobuf:"varint,1,opt,name=memory" json:"memory,omitempty"`
	NanoCPUs int64 `protobuf:"varint,2,opt,name=nanoCPUs" json:"nanoCPUs,omitempty"`
}

func (m *ContainerResources) Reset()                    { *m = ContainerResources{} }
func (m *ContainerResources) String() string            { return proto.CompactTextString(m) }
func (*ContainerResources) ProtoMessage()               {}
func (*ContainerResources) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{12} }

func (m *ContainerResources) GetMemory() int64 {
	if m != nil {
		return m.Memory
	}
	return 0
}

func (m *ContainerResources) GetNanoCPUs() int64 {
	if m != nil {
		return m.NanoCPUs
	}
	return 0
}

type ContainerRestartPolicy struct {
	Name              string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	MaximumRetryCount uint32 `protobuf:"varint,2,opt,name=maximumRetryCount" json:"maximumRetryCount,omitempty"`
}

func (m *ContainerRestartPolicy) Reset()                    { *m = ContainerRestartPolicy{} }
func (m *ContainerRestartPolicy) String() string            { return proto.CompactTextString(m) }
func (*ContainerRestartPolicy) ProtoMessage()               {}
func (*ContainerRestartPolicy) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{13} }

func (m *ContainerRestartPolicy) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ContainerRestartPolicy) GetMaximumRetryCount() uint32 {
	if m != nil {
		return m.MaximumRetryCount
	}
	return 0
}

type TaskLogsRequest struct {
	Type          TaskLogsRequest_Type `protobuf:"varint,1,opt,name=type,enum=sonm.TaskLogsRequest_Type" json:"type,omitempty"`
	Id            string               `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
	Since         string               `protobuf:"bytes,3,opt,name=since" json:"since,omitempty"`
	AddTimestamps bool                 `protobuf:"varint,4,opt,name=addTimestamps" json:"addTimestamps,omitempty"`
	Follow        bool                 `protobuf:"varint,5,opt,name=Follow,json=follow" json:"Follow,omitempty"`
	Tail          string               `protobuf:"bytes,6,opt,name=Tail,json=tail" json:"Tail,omitempty"`
	Details       bool                 `protobuf:"varint,7,opt,name=Details,json=details" json:"Details,omitempty"`
}

func (m *TaskLogsRequest) Reset()                    { *m = TaskLogsRequest{} }
func (m *TaskLogsRequest) String() string            { return proto.CompactTextString(m) }
func (*TaskLogsRequest) ProtoMessage()               {}
func (*TaskLogsRequest) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{14} }

func (m *TaskLogsRequest) GetType() TaskLogsRequest_Type {
	if m != nil {
		return m.Type
	}
	return TaskLogsRequest_STDOUT
}

func (m *TaskLogsRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *TaskLogsRequest) GetSince() string {
	if m != nil {
		return m.Since
	}
	return ""
}

func (m *TaskLogsRequest) GetAddTimestamps() bool {
	if m != nil {
		return m.AddTimestamps
	}
	return false
}

func (m *TaskLogsRequest) GetFollow() bool {
	if m != nil {
		return m.Follow
	}
	return false
}

func (m *TaskLogsRequest) GetTail() string {
	if m != nil {
		return m.Tail
	}
	return ""
}

func (m *TaskLogsRequest) GetDetails() bool {
	if m != nil {
		return m.Details
	}
	return false
}

type TaskLogsChunk struct {
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *TaskLogsChunk) Reset()                    { *m = TaskLogsChunk{} }
func (m *TaskLogsChunk) String() string            { return proto.CompactTextString(m) }
func (*TaskLogsChunk) ProtoMessage()               {}
func (*TaskLogsChunk) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{15} }

func (m *TaskLogsChunk) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*PingRequest)(nil), "sonm.PingRequest")
	proto.RegisterType((*PingReply)(nil), "sonm.PingReply")
	proto.RegisterType((*CPUUsage)(nil), "sonm.CPUUsage")
	proto.RegisterType((*MemoryUsage)(nil), "sonm.MemoryUsage")
	proto.RegisterType((*NetworkUsage)(nil), "sonm.NetworkUsage")
	proto.RegisterType((*ResourceUsage)(nil), "sonm.ResourceUsage")
	proto.RegisterType((*MinerStatusReply)(nil), "sonm.MinerStatusReply")
	proto.RegisterType((*StopTaskRequest)(nil), "sonm.StopTaskRequest")
	proto.RegisterType((*StopTaskReply)(nil), "sonm.StopTaskReply")
	proto.RegisterType((*TaskStatusRequest)(nil), "sonm.TaskStatusRequest")
	proto.RegisterType((*TaskStatusReply)(nil), "sonm.TaskStatusReply")
	proto.RegisterType((*StatusMapReply)(nil), "sonm.StatusMapReply")
	proto.RegisterType((*ContainerResources)(nil), "sonm.ContainerResources")
	proto.RegisterType((*ContainerRestartPolicy)(nil), "sonm.ContainerRestartPolicy")
	proto.RegisterType((*TaskLogsRequest)(nil), "sonm.TaskLogsRequest")
	proto.RegisterType((*TaskLogsChunk)(nil), "sonm.TaskLogsChunk")
	proto.RegisterEnum("sonm.TaskStatusReply_Status", TaskStatusReply_Status_name, TaskStatusReply_Status_value)
	proto.RegisterEnum("sonm.TaskLogsRequest_Type", TaskLogsRequest_Type_name, TaskLogsRequest_Type_value)
}

func init() { proto.RegisterFile("proto/insonmnia.proto", fileDescriptor2) }

var fileDescriptor2 = []byte{
	// 890 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x7c, 0x55, 0x5d, 0x8f, 0xdb, 0x44,
	0x14, 0xad, 0x13, 0x27, 0x71, 0x6e, 0x36, 0xbb, 0xde, 0x81, 0x56, 0x51, 0xd4, 0x87, 0xd4, 0xcb,
	0xc3, 0xae, 0x40, 0x41, 0x5a, 0x10, 0xa0, 0x3e, 0x20, 0xb1, 0xd9, 0x94, 0x5d, 0xb5, 0xeb, 0x44,
	0x93, 0x44, 0x45, 0xbc, 0x4d, 0x93, 0x61, 0xb1, 0x62, 0x7b, 0x8c, 0x3d, 0xa6, 0xf1, 0x3f, 0xe1,
	0x07, 0xf0, 0xdb, 0x10, 0xcf, 0xbc, 0x23, 0xa1, 0x99, 0x3b, 0x76, 0x9c, 0x12, 0xf5, 0x6d, 0xce,
	0x9c, 0xe3, 0xb9, 0x1f, 0x73, 0xee, 0x18, 0x9e, 0x26, 0xa9, 0x90, 0xe2, 0xcb, 0x20, 0xce, 0x44,
	0x1c, 0xc5, 0x01, 0x1b, 0x6b, 0x4c, 0x6c, 0x05, 0x87, 0x03, 0x24, 0xd7, 0x2c, 0x61, 0xef, 0x82,
	0x30, 0x90, 0x01, 0xcf, 0x90, 0xf7, 0xfa, 0xd0, 0x9b, 0x07, 0xf1, 0x23, 0xe5, 0xbf, 0xe5, 0x3c,
	0x93, 0xde, 0x05, 0x74, 0x11, 0x26, 0x61, 0x41, 0x9e, 0x41, 0x3b, 0x93, 0x4c, 0xe6, 0xd9, 0xc0,
	0x1a, 0x59, 0x97, 0x5d, 0x6a, 0x90, 0x37, 0x02, 0x67, 0x32, 0x5f, 0xad, 0x32, 0xf6, 0xc8, 0xc9,
	0xa7, 0xd0, 0x92, 0x42, 0xb2, 0x50, 0x4b, 0x6c, 0x8a, 0xc0, 0xbb, 0x82, 0xde, 0x03, 0x8f, 0x44,
	0x5a, 0xa0, 0x68, 0x08, 0x4e, 0xc4, 0x76, 0x7a, 0x6d, 0x74, 0x15, 0xf6, 0xfe, 0xb1, 0xe0, 0xc4,
	0xe7, 0xf2, 0xbd, 0x48, 0xb7, 0x28, 0x1e, 0x40, 0x47, 0xee, 0x6e, 0x0a, 0xc9, 0x33, 0xa3, 0x2d,
	0xa1, 0x62, 0x52, 0xc3, 0x34, 0x90, 0x31, 0x90, 0x3c, 0x87, 0xae, 0xdc, 0xcd, 0xd9, 0x7a, 0xcb,
	0x65, 0x36, 0x68, 0x6a, 0x6e, 0xbf, 0xa1, 0xd8, 0xb4, 0x62, 0x6d, 0x64, 0xab, 0x0d, 0x95, 0x9c,
	0xdc, 0x4d, 0xd3, 0x54, 0xa4, 0xd9, 0xa0, 0x85, 0xc9, 0x95, 0x58, 0x71, 0x69, 0xc9, 0xb5, 0x91,
	0x2b, 0x31, 0xc6, 0xbc, 0x4d, 0x45, 0x92, 0xf0, 0xcd, 0xa0, 0x53, 0xc6, 0x34, 0x1b, 0x18, 0xb3,
	0x64, 0x9d, 0x32, 0xa6, 0xd9, 0xf0, 0xfe, 0xb6, 0xa0, 0x4f, 0x79, 0x26, 0xf2, 0x74, 0xcd, 0xb1,
	0xea, 0x11, 0x34, 0xd7, 0x49, 0xae, 0x2b, 0xee, 0x5d, 0x9f, 0x8e, 0xd5, 0xad, 0x8d, 0xcb, 0x26,
	0x53, 0x45, 0x91, 0x2b, 0x68, 0x47, 0xba, 0xa7, 0xba, 0xf8, 0xde, 0xf5, 0x39, 0x8a, 0x6a, 0x7d,
	0xa6, 0x46, 0x40, 0x5e, 0x42, 0x27, 0xc6, 0x96, 0x0e, 0x9a, 0xa3, 0xe6, 0x65, 0xef, 0x7a, 0x84,
	0xda, 0x83, 0x90, 0x63, 0xd3, 0xf5, 0x69, 0x2c, 0xd3, 0x82, 0x96, 0x1f, 0x0c, 0xfd, 0xea, 0x3a,
	0x34, 0x41, 0x5c, 0x68, 0x6e, 0x79, 0x61, 0x1c, 0xa0, 0x96, 0xe4, 0x12, 0x5a, 0xbf, 0xb3, 0x30,
	0xe7, 0x26, 0x0f, 0x82, 0x67, 0xd7, 0xef, 0x90, 0xa2, 0xe0, 0x65, 0xe3, 0x3b, 0xcb, 0xfb, 0xcb,
	0x02, 0xf7, 0x21, 0x88, 0x79, 0xba, 0xd0, 0xe6, 0x41, 0x67, 0x7d, 0x0b, 0xad, 0xdc, 0xb8, 0x41,
	0xa5, 0xf7, 0xc2, 0x94, 0xf2, 0x81, 0x6c, 0xac, 0x0f, 0xc3, 0xfc, 0x50, 0x4f, 0x08, 0xd8, 0x31,
	0x8b, 0x30, 0x74, 0x97, 0xea, 0x35, 0xf9, 0x06, 0x4e, 0xea, 0xc6, 0xd6, 0xf7, 0x5f, 0xa5, 0x35,
	0xa9, 0x31, 0xf4, 0x40, 0x37, 0x7c, 0x00, 0xd8, 0x07, 0x38, 0x52, 0xe7, 0xd5, 0x61, 0x9d, 0x9f,
	0x1c, 0xe9, 0x61, 0xbd, 0xd0, 0x17, 0x70, 0xb6, 0x90, 0x22, 0x59, 0xb2, 0x6c, 0x6b, 0xa6, 0x89,
	0x9c, 0x42, 0x23, 0xd8, 0x98, 0x23, 0x1b, 0xc1, 0xc6, 0x3b, 0x83, 0xfe, 0x5e, 0x92, 0x84, 0x85,
	0x77, 0x01, 0xe7, 0x0a, 0x94, 0x35, 0x1f, 0xff, 0xea, 0x8f, 0x06, 0x9c, 0xd5, 0x55, 0xaa, 0x81,
	0x5f, 0x1f, 0x8c, 0xe6, 0xe9, 0xf5, 0x73, 0x4c, 0xee, 0x03, 0xd9, 0xd8, 0xac, 0x8d, 0x56, 0x99,
	0x32, 0x88, 0xd8, 0x23, 0xf7, 0xf7, 0x2d, 0xdc, 0x6f, 0xa8, 0x51, 0x4e, 0x44, 0x6a, 0x06, 0xa8,
	0x4b, 0x11, 0xa8, 0x47, 0x20, 0x4f, 0x64, 0x10, 0x71, 0x33, 0x39, 0x06, 0xa9, 0xee, 0xe0, 0x15,
	0xb6, 0x3e, 0xd2, 0x1d, 0xad, 0xf0, 0x7e, 0x82, 0x36, 0x26, 0x42, 0x7a, 0xd0, 0x59, 0xf9, 0xaf,
	0xfd, 0xd9, 0x5b, 0xdf, 0x7d, 0x42, 0x4e, 0xc0, 0x59, 0xcc, 0x67, 0xb3, 0x37, 0xf7, 0xfe, 0x8f,
	0xae, 0x85, 0xe8, 0x87, 0xb7, 0xbe, 0x42, 0x0d, 0x25, 0xa4, 0x2b, 0x5f, 0x83, 0xa6, 0xa2, 0x5e,
	0xdd, 0xfb, 0xf7, 0x8b, 0xbb, 0xe9, 0xad, 0x6b, 0x13, 0x80, 0xf6, 0x0d, 0x9d, 0xbd, 0x9e, 0xfa,
	0x6e, 0xcb, 0xfb, 0xd3, 0x82, 0x53, 0x3c, 0xfa, 0x81, 0x25, 0xd8, 0x99, 0xef, 0xc1, 0xc1, 0x6a,
	0xf5, 0xfb, 0xa1, 0xdc, 0xe5, 0x61, 0x6a, 0x87, 0x3a, 0x03, 0x79, 0x86, 0xf6, 0xaa, 0xbe, 0x19,
	0x52, 0x75, 0x47, 0x35, 0xea, 0x88, 0x31, 0x3e, 0x3f, 0x34, 0xc6, 0xd3, 0xa3, 0xbd, 0xaf, 0x5b,
	0xe3, 0x0e, 0xc8, 0x44, 0xc4, 0x92, 0x29, 0x7f, 0x97, 0x1d, 0xd2, 0x9d, 0x35, 0x03, 0xad, 0xce,
	0x6e, 0x56, 0xd3, 0x3b, 0x04, 0x27, 0x66, 0xb1, 0x98, 0xcc, 0x57, 0xf8, 0xce, 0x35, 0x69, 0x85,
	0xbd, 0x9f, 0xe1, 0x59, 0xfd, 0x24, 0xc9, 0x52, 0x39, 0x17, 0x61, 0xb0, 0x2e, 0xaa, 0xc9, 0xb0,
	0x6a, 0x93, 0xf1, 0x05, 0x9c, 0x47, 0x6c, 0x17, 0x44, 0x79, 0x44, 0xb9, 0x4c, 0x8b, 0x89, 0xc8,
	0x63, 0xa9, 0x8f, 0xec, 0xd3, 0xff, 0x13, 0xde, 0xbf, 0x16, 0xfa, 0xec, 0x8d, 0x78, 0xac, 0xbc,
	0x38, 0x06, 0x5b, 0x16, 0x09, 0x37, 0x2e, 0x1b, 0xee, 0x2b, 0xad, 0x89, 0xc6, 0xcb, 0x22, 0xe1,
	0x54, 0xeb, 0x8c, 0x77, 0x1b, 0xa5, 0x77, 0x95, 0xa7, 0xb2, 0x20, 0x5e, 0xf3, 0xd2, 0x53, 0x1a,
	0x90, 0xcf, 0xa0, 0xcf, 0x36, 0x9b, 0x65, 0x10, 0xa9, 0x0a, 0xa2, 0x04, 0x1f, 0x65, 0x87, 0x1e,
	0x6e, 0xaa, 0xfe, 0xbc, 0x12, 0x61, 0x28, 0xde, 0x6b, 0x8b, 0x39, 0xb4, 0xfd, 0x8b, 0x46, 0xaa,
	0xd2, 0x25, 0x0b, 0x42, 0xfd, 0x20, 0x77, 0xa9, 0x2d, 0x59, 0x10, 0xaa, 0x5f, 0xc3, 0x2d, 0x57,
	0xab, 0x4c, 0x3f, 0xc5, 0x0e, 0xed, 0x6c, 0x10, 0x7a, 0x97, 0x60, 0xab, 0xfc, 0x94, 0x6d, 0x16,
	0xcb, 0xdb, 0xd9, 0x6a, 0xe9, 0x3e, 0x31, 0xeb, 0x29, 0xa5, 0xae, 0x45, 0x1c, 0xb0, 0x6f, 0x66,
	0xcb, 0x3b, 0xb7, 0xe1, 0x5d, 0x40, 0xbf, 0xac, 0x6c, 0xf2, 0x6b, 0x1e, 0x6f, 0x55, 0xa0, 0x0d,
	0x93, 0x4c, 0x17, 0x7f, 0x42, 0xf5, 0xfa, 0x5d, 0x5b, 0xff, 0x36, 0xbf, 0xfa, 0x2f, 0x00, 0x00,
	0xff, 0xff, 0x6a, 0x57, 0xa1, 0x25, 0x6f, 0x07, 0x00, 0x00,
}

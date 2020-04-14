// Code generated by protoc-gen-go. DO NOT EDIT.
// source: goserver.proto

package goserver

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type ContextType int32

const (
	ContextType_UNKNOWN  ContextType = 0
	ContextType_REGULAR  ContextType = 1
	ContextType_MEDIUM   ContextType = 4
	ContextType_LONG     ContextType = 2
	ContextType_INFINITE ContextType = 3
	ContextType_NO_TRACE ContextType = 5
)

var ContextType_name = map[int32]string{
	0: "UNKNOWN",
	1: "REGULAR",
	4: "MEDIUM",
	2: "LONG",
	3: "INFINITE",
	5: "NO_TRACE",
}

var ContextType_value = map[string]int32{
	"UNKNOWN":  0,
	"REGULAR":  1,
	"MEDIUM":   4,
	"LONG":     2,
	"INFINITE": 3,
	"NO_TRACE": 5,
}

func (x ContextType) String() string {
	return proto.EnumName(ContextType_name, int32(x))
}

func (ContextType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_13e7ca56064c33f1, []int{0}
}

type Empty struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()    {}
func (*Empty) Descriptor() ([]byte, []int) {
	return fileDescriptor_13e7ca56064c33f1, []int{0}
}

func (m *Empty) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Empty.Unmarshal(m, b)
}
func (m *Empty) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Empty.Marshal(b, m, deterministic)
}
func (m *Empty) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Empty.Merge(m, src)
}
func (m *Empty) XXX_Size() int {
	return xxx_messageInfo_Empty.Size(m)
}
func (m *Empty) XXX_DiscardUnknown() {
	xxx_messageInfo_Empty.DiscardUnknown(m)
}

var xxx_messageInfo_Empty proto.InternalMessageInfo

type Alive struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Alive) Reset()         { *m = Alive{} }
func (m *Alive) String() string { return proto.CompactTextString(m) }
func (*Alive) ProtoMessage()    {}
func (*Alive) Descriptor() ([]byte, []int) {
	return fileDescriptor_13e7ca56064c33f1, []int{1}
}

func (m *Alive) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Alive.Unmarshal(m, b)
}
func (m *Alive) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Alive.Marshal(b, m, deterministic)
}
func (m *Alive) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Alive.Merge(m, src)
}
func (m *Alive) XXX_Size() int {
	return xxx_messageInfo_Alive.Size(m)
}
func (m *Alive) XXX_DiscardUnknown() {
	xxx_messageInfo_Alive.DiscardUnknown(m)
}

var xxx_messageInfo_Alive proto.InternalMessageInfo

func (m *Alive) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type MoteRequest struct {
	Master               bool     `protobuf:"varint,1,opt,name=master,proto3" json:"master,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MoteRequest) Reset()         { *m = MoteRequest{} }
func (m *MoteRequest) String() string { return proto.CompactTextString(m) }
func (*MoteRequest) ProtoMessage()    {}
func (*MoteRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_13e7ca56064c33f1, []int{2}
}

func (m *MoteRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MoteRequest.Unmarshal(m, b)
}
func (m *MoteRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MoteRequest.Marshal(b, m, deterministic)
}
func (m *MoteRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MoteRequest.Merge(m, src)
}
func (m *MoteRequest) XXX_Size() int {
	return xxx_messageInfo_MoteRequest.Size(m)
}
func (m *MoteRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_MoteRequest.DiscardUnknown(m)
}

var xxx_messageInfo_MoteRequest proto.InternalMessageInfo

func (m *MoteRequest) GetMaster() bool {
	if m != nil {
		return m.Master
	}
	return false
}

type EmbeddedTest struct {
	Blah                 *Alive   `protobuf:"bytes,1,opt,name=blah,proto3" json:"blah,omitempty"`
	Test                 string   `protobuf:"bytes,2,opt,name=test,proto3" json:"test,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *EmbeddedTest) Reset()         { *m = EmbeddedTest{} }
func (m *EmbeddedTest) String() string { return proto.CompactTextString(m) }
func (*EmbeddedTest) ProtoMessage()    {}
func (*EmbeddedTest) Descriptor() ([]byte, []int) {
	return fileDescriptor_13e7ca56064c33f1, []int{3}
}

func (m *EmbeddedTest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EmbeddedTest.Unmarshal(m, b)
}
func (m *EmbeddedTest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EmbeddedTest.Marshal(b, m, deterministic)
}
func (m *EmbeddedTest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EmbeddedTest.Merge(m, src)
}
func (m *EmbeddedTest) XXX_Size() int {
	return xxx_messageInfo_EmbeddedTest.Size(m)
}
func (m *EmbeddedTest) XXX_DiscardUnknown() {
	xxx_messageInfo_EmbeddedTest.DiscardUnknown(m)
}

var xxx_messageInfo_EmbeddedTest proto.InternalMessageInfo

func (m *EmbeddedTest) GetBlah() *Alive {
	if m != nil {
		return m.Blah
	}
	return nil
}

func (m *EmbeddedTest) GetTest() string {
	if m != nil {
		return m.Test
	}
	return ""
}

type State struct {
	Key                  string   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	TimeValue            int64    `protobuf:"varint,2,opt,name=time_value,json=timeValue,proto3" json:"time_value,omitempty"`
	Value                int64    `protobuf:"varint,3,opt,name=value,proto3" json:"value,omitempty"`
	Text                 string   `protobuf:"bytes,4,opt,name=text,proto3" json:"text,omitempty"`
	Fraction             float64  `protobuf:"fixed64,5,opt,name=fraction,proto3" json:"fraction,omitempty"`
	TimeDuration         int64    `protobuf:"varint,6,opt,name=time_duration,json=timeDuration,proto3" json:"time_duration,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *State) Reset()         { *m = State{} }
func (m *State) String() string { return proto.CompactTextString(m) }
func (*State) ProtoMessage()    {}
func (*State) Descriptor() ([]byte, []int) {
	return fileDescriptor_13e7ca56064c33f1, []int{4}
}

func (m *State) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_State.Unmarshal(m, b)
}
func (m *State) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_State.Marshal(b, m, deterministic)
}
func (m *State) XXX_Merge(src proto.Message) {
	xxx_messageInfo_State.Merge(m, src)
}
func (m *State) XXX_Size() int {
	return xxx_messageInfo_State.Size(m)
}
func (m *State) XXX_DiscardUnknown() {
	xxx_messageInfo_State.DiscardUnknown(m)
}

var xxx_messageInfo_State proto.InternalMessageInfo

func (m *State) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *State) GetTimeValue() int64 {
	if m != nil {
		return m.TimeValue
	}
	return 0
}

func (m *State) GetValue() int64 {
	if m != nil {
		return m.Value
	}
	return 0
}

func (m *State) GetText() string {
	if m != nil {
		return m.Text
	}
	return ""
}

func (m *State) GetFraction() float64 {
	if m != nil {
		return m.Fraction
	}
	return 0
}

func (m *State) GetTimeDuration() int64 {
	if m != nil {
		return m.TimeDuration
	}
	return 0
}

type ServerState struct {
	States               []*State `protobuf:"bytes,1,rep,name=states,proto3" json:"states,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ServerState) Reset()         { *m = ServerState{} }
func (m *ServerState) String() string { return proto.CompactTextString(m) }
func (*ServerState) ProtoMessage()    {}
func (*ServerState) Descriptor() ([]byte, []int) {
	return fileDescriptor_13e7ca56064c33f1, []int{5}
}

func (m *ServerState) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServerState.Unmarshal(m, b)
}
func (m *ServerState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServerState.Marshal(b, m, deterministic)
}
func (m *ServerState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServerState.Merge(m, src)
}
func (m *ServerState) XXX_Size() int {
	return xxx_messageInfo_ServerState.Size(m)
}
func (m *ServerState) XXX_DiscardUnknown() {
	xxx_messageInfo_ServerState.DiscardUnknown(m)
}

var xxx_messageInfo_ServerState proto.InternalMessageInfo

func (m *ServerState) GetStates() []*State {
	if m != nil {
		return m.States
	}
	return nil
}

type TaskPeriod struct {
	Key                  string   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Period               int64    `protobuf:"varint,2,opt,name=period,proto3" json:"period,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TaskPeriod) Reset()         { *m = TaskPeriod{} }
func (m *TaskPeriod) String() string { return proto.CompactTextString(m) }
func (*TaskPeriod) ProtoMessage()    {}
func (*TaskPeriod) Descriptor() ([]byte, []int) {
	return fileDescriptor_13e7ca56064c33f1, []int{6}
}

func (m *TaskPeriod) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TaskPeriod.Unmarshal(m, b)
}
func (m *TaskPeriod) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TaskPeriod.Marshal(b, m, deterministic)
}
func (m *TaskPeriod) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TaskPeriod.Merge(m, src)
}
func (m *TaskPeriod) XXX_Size() int {
	return xxx_messageInfo_TaskPeriod.Size(m)
}
func (m *TaskPeriod) XXX_DiscardUnknown() {
	xxx_messageInfo_TaskPeriod.DiscardUnknown(m)
}

var xxx_messageInfo_TaskPeriod proto.InternalMessageInfo

func (m *TaskPeriod) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *TaskPeriod) GetPeriod() int64 {
	if m != nil {
		return m.Period
	}
	return 0
}

type ServerConfig struct {
	Periods              []*TaskPeriod `protobuf:"bytes,1,rep,name=periods,proto3" json:"periods,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *ServerConfig) Reset()         { *m = ServerConfig{} }
func (m *ServerConfig) String() string { return proto.CompactTextString(m) }
func (*ServerConfig) ProtoMessage()    {}
func (*ServerConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_13e7ca56064c33f1, []int{7}
}

func (m *ServerConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServerConfig.Unmarshal(m, b)
}
func (m *ServerConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServerConfig.Marshal(b, m, deterministic)
}
func (m *ServerConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServerConfig.Merge(m, src)
}
func (m *ServerConfig) XXX_Size() int {
	return xxx_messageInfo_ServerConfig.Size(m)
}
func (m *ServerConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_ServerConfig.DiscardUnknown(m)
}

var xxx_messageInfo_ServerConfig proto.InternalMessageInfo

func (m *ServerConfig) GetPeriods() []*TaskPeriod {
	if m != nil {
		return m.Periods
	}
	return nil
}

type ShutdownRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ShutdownRequest) Reset()         { *m = ShutdownRequest{} }
func (m *ShutdownRequest) String() string { return proto.CompactTextString(m) }
func (*ShutdownRequest) ProtoMessage()    {}
func (*ShutdownRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_13e7ca56064c33f1, []int{8}
}

func (m *ShutdownRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ShutdownRequest.Unmarshal(m, b)
}
func (m *ShutdownRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ShutdownRequest.Marshal(b, m, deterministic)
}
func (m *ShutdownRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ShutdownRequest.Merge(m, src)
}
func (m *ShutdownRequest) XXX_Size() int {
	return xxx_messageInfo_ShutdownRequest.Size(m)
}
func (m *ShutdownRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ShutdownRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ShutdownRequest proto.InternalMessageInfo

type ShutdownResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ShutdownResponse) Reset()         { *m = ShutdownResponse{} }
func (m *ShutdownResponse) String() string { return proto.CompactTextString(m) }
func (*ShutdownResponse) ProtoMessage()    {}
func (*ShutdownResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_13e7ca56064c33f1, []int{9}
}

func (m *ShutdownResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ShutdownResponse.Unmarshal(m, b)
}
func (m *ShutdownResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ShutdownResponse.Marshal(b, m, deterministic)
}
func (m *ShutdownResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ShutdownResponse.Merge(m, src)
}
func (m *ShutdownResponse) XXX_Size() int {
	return xxx_messageInfo_ShutdownResponse.Size(m)
}
func (m *ShutdownResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ShutdownResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ShutdownResponse proto.InternalMessageInfo

type ReregisterRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReregisterRequest) Reset()         { *m = ReregisterRequest{} }
func (m *ReregisterRequest) String() string { return proto.CompactTextString(m) }
func (*ReregisterRequest) ProtoMessage()    {}
func (*ReregisterRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_13e7ca56064c33f1, []int{10}
}

func (m *ReregisterRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReregisterRequest.Unmarshal(m, b)
}
func (m *ReregisterRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReregisterRequest.Marshal(b, m, deterministic)
}
func (m *ReregisterRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReregisterRequest.Merge(m, src)
}
func (m *ReregisterRequest) XXX_Size() int {
	return xxx_messageInfo_ReregisterRequest.Size(m)
}
func (m *ReregisterRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ReregisterRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ReregisterRequest proto.InternalMessageInfo

type ReregisterResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReregisterResponse) Reset()         { *m = ReregisterResponse{} }
func (m *ReregisterResponse) String() string { return proto.CompactTextString(m) }
func (*ReregisterResponse) ProtoMessage()    {}
func (*ReregisterResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_13e7ca56064c33f1, []int{11}
}

func (m *ReregisterResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReregisterResponse.Unmarshal(m, b)
}
func (m *ReregisterResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReregisterResponse.Marshal(b, m, deterministic)
}
func (m *ReregisterResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReregisterResponse.Merge(m, src)
}
func (m *ReregisterResponse) XXX_Size() int {
	return xxx_messageInfo_ReregisterResponse.Size(m)
}
func (m *ReregisterResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ReregisterResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ReregisterResponse proto.InternalMessageInfo

func init() {
	proto.RegisterEnum("goserver.ContextType", ContextType_name, ContextType_value)
	proto.RegisterType((*Empty)(nil), "goserver.Empty")
	proto.RegisterType((*Alive)(nil), "goserver.Alive")
	proto.RegisterType((*MoteRequest)(nil), "goserver.MoteRequest")
	proto.RegisterType((*EmbeddedTest)(nil), "goserver.EmbeddedTest")
	proto.RegisterType((*State)(nil), "goserver.State")
	proto.RegisterType((*ServerState)(nil), "goserver.ServerState")
	proto.RegisterType((*TaskPeriod)(nil), "goserver.TaskPeriod")
	proto.RegisterType((*ServerConfig)(nil), "goserver.ServerConfig")
	proto.RegisterType((*ShutdownRequest)(nil), "goserver.ShutdownRequest")
	proto.RegisterType((*ShutdownResponse)(nil), "goserver.ShutdownResponse")
	proto.RegisterType((*ReregisterRequest)(nil), "goserver.ReregisterRequest")
	proto.RegisterType((*ReregisterResponse)(nil), "goserver.ReregisterResponse")
}

func init() { proto.RegisterFile("goserver.proto", fileDescriptor_13e7ca56064c33f1) }

var fileDescriptor_13e7ca56064c33f1 = []byte{
	// 536 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x53, 0xdd, 0x6e, 0xda, 0x4c,
	0x10, 0xc5, 0xc1, 0x36, 0xce, 0xc0, 0xf7, 0xe1, 0x4c, 0x69, 0x44, 0x9d, 0x56, 0x42, 0x1b, 0x55,
	0x8d, 0x2a, 0x15, 0xb5, 0xa9, 0x94, 0xcb, 0x4a, 0x88, 0xb8, 0xc8, 0x6a, 0x30, 0xd5, 0x02, 0xad,
	0x7a, 0x15, 0x99, 0x78, 0x43, 0xac, 0x80, 0x4d, 0xed, 0x85, 0x86, 0xe7, 0xe9, 0x33, 0xf4, 0xfd,
	0xaa, 0x5d, 0xdb, 0xb1, 0x43, 0xb9, 0x9b, 0x39, 0xe7, 0xec, 0x9e, 0x9d, 0x9f, 0x85, 0xff, 0xe7,
	0x51, 0xc2, 0xe2, 0x0d, 0x8b, 0xbb, 0xab, 0x38, 0xe2, 0x11, 0x1a, 0x79, 0x4e, 0x6a, 0xa0, 0xd9,
	0xcb, 0x15, 0xdf, 0x92, 0x13, 0xd0, 0x7a, 0x8b, 0x60, 0xc3, 0x10, 0x41, 0x0d, 0xbd, 0x25, 0x6b,
	0x2b, 0x1d, 0xe5, 0xec, 0x90, 0xca, 0x98, 0xbc, 0x86, 0xfa, 0x30, 0xe2, 0x8c, 0xb2, 0x9f, 0x6b,
	0x96, 0x70, 0x3c, 0x06, 0x7d, 0xe9, 0x25, 0x9c, 0xc5, 0x52, 0x64, 0xd0, 0x2c, 0x23, 0x03, 0x68,
	0xd8, 0xcb, 0x19, 0xf3, 0x7d, 0xe6, 0x4f, 0x84, 0xee, 0x14, 0xd4, 0xd9, 0xc2, 0xbb, 0x93, 0xaa,
	0xfa, 0x79, 0xb3, 0xfb, 0xf8, 0x0a, 0xe9, 0x44, 0x25, 0x29, 0xfc, 0x38, 0x4b, 0x78, 0xfb, 0x20,
	0xf5, 0x13, 0x31, 0xf9, 0xad, 0x80, 0x36, 0xe6, 0x1e, 0x67, 0x68, 0x42, 0xf5, 0x9e, 0x6d, 0xb3,
	0xc7, 0x88, 0x10, 0x5f, 0x01, 0xf0, 0x60, 0xc9, 0xae, 0x37, 0xde, 0x62, 0xcd, 0xe4, 0xa9, 0x2a,
	0x3d, 0x14, 0xc8, 0x37, 0x01, 0x60, 0x0b, 0xb4, 0x94, 0xa9, 0x4a, 0x26, 0x4d, 0x52, 0x93, 0x07,
	0xde, 0x56, 0x73, 0x93, 0x07, 0x8e, 0x16, 0x18, 0xb7, 0xb1, 0x77, 0xc3, 0x83, 0x28, 0x6c, 0x6b,
	0x1d, 0xe5, 0x4c, 0xa1, 0x8f, 0x39, 0x9e, 0xc2, 0x7f, 0xd2, 0xc4, 0x5f, 0xc7, 0x9e, 0x14, 0xe8,
	0xf2, 0xb6, 0x86, 0x00, 0x2f, 0x33, 0x8c, 0x5c, 0x40, 0x7d, 0x2c, 0xeb, 0x49, 0x9f, 0xfa, 0x06,
	0xf4, 0x44, 0x04, 0x49, 0x5b, 0xe9, 0x54, 0x9f, 0xd6, 0x2b, 0x05, 0x34, 0xa3, 0xc9, 0x05, 0xc0,
	0xc4, 0x4b, 0xee, 0xbf, 0xb2, 0x38, 0x88, 0xfc, 0x3d, 0x15, 0x1e, 0x83, 0xbe, 0x92, 0x5c, 0x56,
	0x5d, 0x96, 0x91, 0x4f, 0xd0, 0x48, 0xfd, 0xfa, 0x51, 0x78, 0x1b, 0xcc, 0xb1, 0x0b, 0xb5, 0x94,
	0xc9, 0x1d, 0x5b, 0x85, 0x63, 0x61, 0x40, 0x73, 0x11, 0x39, 0x82, 0xe6, 0xf8, 0x6e, 0xcd, 0xfd,
	0xe8, 0x57, 0x98, 0x4d, 0x92, 0x20, 0x98, 0x05, 0x94, 0xac, 0xa2, 0x30, 0x61, 0xe4, 0x19, 0x1c,
	0x51, 0x16, 0xb3, 0x79, 0x20, 0x66, 0x9a, 0x0b, 0x5b, 0x80, 0x65, 0x30, 0x95, 0xbe, 0xfd, 0x01,
	0xf5, 0x7e, 0x14, 0x8a, 0x6e, 0x4e, 0xb6, 0x2b, 0x86, 0x75, 0xa8, 0x4d, 0xdd, 0x2f, 0xee, 0xe8,
	0xbb, 0x6b, 0x56, 0x44, 0x42, 0xed, 0xc1, 0xf4, 0xaa, 0x47, 0x4d, 0x05, 0x01, 0xf4, 0xa1, 0x7d,
	0xe9, 0x4c, 0x87, 0xa6, 0x8a, 0x06, 0xa8, 0x57, 0x23, 0x77, 0x60, 0x1e, 0x60, 0x03, 0x0c, 0xc7,
	0xfd, 0xec, 0xb8, 0xce, 0xc4, 0x36, 0xab, 0x22, 0x73, 0x47, 0xd7, 0x13, 0xda, 0xeb, 0xdb, 0xa6,
	0x76, 0xfe, 0xe7, 0x00, 0x9a, 0x79, 0x35, 0xa2, 0xea, 0xe0, 0x86, 0xe1, 0x3b, 0xa8, 0x39, 0x49,
	0xba, 0xa5, 0xbb, 0xcb, 0x64, 0xed, 0x02, 0xa4, 0x82, 0xef, 0x41, 0x15, 0x5b, 0x8b, 0xcf, 0x0b,
	0xaa, 0xb4, 0xc5, 0xe5, 0x13, 0xe9, 0x17, 0xa8, 0xe0, 0x87, 0x7c, 0xed, 0x76, 0x39, 0xab, 0x74,
	0x47, 0x69, 0xe6, 0xa4, 0x82, 0x7d, 0x30, 0xf2, 0x0e, 0xe2, 0x8b, 0x92, 0xe8, 0x69, 0xa3, 0x2d,
	0x6b, 0x1f, 0x95, 0x35, 0xbc, 0x82, 0x0e, 0x40, 0xd1, 0x5d, 0x3c, 0x29, 0xb4, 0xff, 0x0c, 0xc2,
	0x7a, 0xb9, 0x9f, 0xcc, 0xaf, 0x9a, 0xe9, 0xf2, 0x87, 0x7f, 0xfc, 0x1b, 0x00, 0x00, 0xff, 0xff,
	0xe7, 0xc4, 0x2e, 0x69, 0xf3, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// GoserverServiceClient is the client API for GoserverService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type GoserverServiceClient interface {
	IsAlive(ctx context.Context, in *Alive, opts ...grpc.CallOption) (*Alive, error)
	Mote(ctx context.Context, in *MoteRequest, opts ...grpc.CallOption) (*Empty, error)
	State(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ServerState, error)
	Shutdown(ctx context.Context, in *ShutdownRequest, opts ...grpc.CallOption) (*ShutdownResponse, error)
	Reregister(ctx context.Context, in *ReregisterRequest, opts ...grpc.CallOption) (*ReregisterResponse, error)
}

type goserverServiceClient struct {
	cc *grpc.ClientConn
}

func NewGoserverServiceClient(cc *grpc.ClientConn) GoserverServiceClient {
	return &goserverServiceClient{cc}
}

func (c *goserverServiceClient) IsAlive(ctx context.Context, in *Alive, opts ...grpc.CallOption) (*Alive, error) {
	out := new(Alive)
	err := c.cc.Invoke(ctx, "/goserver.goserverService/IsAlive", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *goserverServiceClient) Mote(ctx context.Context, in *MoteRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/goserver.goserverService/Mote", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *goserverServiceClient) State(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ServerState, error) {
	out := new(ServerState)
	err := c.cc.Invoke(ctx, "/goserver.goserverService/State", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *goserverServiceClient) Shutdown(ctx context.Context, in *ShutdownRequest, opts ...grpc.CallOption) (*ShutdownResponse, error) {
	out := new(ShutdownResponse)
	err := c.cc.Invoke(ctx, "/goserver.goserverService/Shutdown", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *goserverServiceClient) Reregister(ctx context.Context, in *ReregisterRequest, opts ...grpc.CallOption) (*ReregisterResponse, error) {
	out := new(ReregisterResponse)
	err := c.cc.Invoke(ctx, "/goserver.goserverService/Reregister", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GoserverServiceServer is the server API for GoserverService service.
type GoserverServiceServer interface {
	IsAlive(context.Context, *Alive) (*Alive, error)
	Mote(context.Context, *MoteRequest) (*Empty, error)
	State(context.Context, *Empty) (*ServerState, error)
	Shutdown(context.Context, *ShutdownRequest) (*ShutdownResponse, error)
	Reregister(context.Context, *ReregisterRequest) (*ReregisterResponse, error)
}

func RegisterGoserverServiceServer(s *grpc.Server, srv GoserverServiceServer) {
	s.RegisterService(&_GoserverService_serviceDesc, srv)
}

func _GoserverService_IsAlive_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Alive)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoserverServiceServer).IsAlive(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/goserver.goserverService/IsAlive",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoserverServiceServer).IsAlive(ctx, req.(*Alive))
	}
	return interceptor(ctx, in, info, handler)
}

func _GoserverService_Mote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoserverServiceServer).Mote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/goserver.goserverService/Mote",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoserverServiceServer).Mote(ctx, req.(*MoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GoserverService_State_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoserverServiceServer).State(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/goserver.goserverService/State",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoserverServiceServer).State(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _GoserverService_Shutdown_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ShutdownRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoserverServiceServer).Shutdown(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/goserver.goserverService/Shutdown",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoserverServiceServer).Shutdown(ctx, req.(*ShutdownRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GoserverService_Reregister_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReregisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoserverServiceServer).Reregister(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/goserver.goserverService/Reregister",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoserverServiceServer).Reregister(ctx, req.(*ReregisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _GoserverService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "goserver.goserverService",
	HandlerType: (*GoserverServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "IsAlive",
			Handler:    _GoserverService_IsAlive_Handler,
		},
		{
			MethodName: "Mote",
			Handler:    _GoserverService_Mote_Handler,
		},
		{
			MethodName: "State",
			Handler:    _GoserverService_State_Handler,
		},
		{
			MethodName: "Shutdown",
			Handler:    _GoserverService_Shutdown_Handler,
		},
		{
			MethodName: "Reregister",
			Handler:    _GoserverService_Reregister_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "goserver.proto",
}

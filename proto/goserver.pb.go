// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.6.1
// source: goserver.proto

package goserver

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type ContextType int32

const (
	ContextType_UNKNOWN  ContextType = 0
	ContextType_REGULAR  ContextType = 1
	ContextType_MEDIUM   ContextType = 4
	ContextType_LONG     ContextType = 2
	ContextType_INFINITE ContextType = 3
	ContextType_NO_TRACE ContextType = 5
)

// Enum value maps for ContextType.
var (
	ContextType_name = map[int32]string{
		0: "UNKNOWN",
		1: "REGULAR",
		4: "MEDIUM",
		2: "LONG",
		3: "INFINITE",
		5: "NO_TRACE",
	}
	ContextType_value = map[string]int32{
		"UNKNOWN":  0,
		"REGULAR":  1,
		"MEDIUM":   4,
		"LONG":     2,
		"INFINITE": 3,
		"NO_TRACE": 5,
	}
)

func (x ContextType) Enum() *ContextType {
	p := new(ContextType)
	*p = x
	return p
}

func (x ContextType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ContextType) Descriptor() protoreflect.EnumDescriptor {
	return file_goserver_proto_enumTypes[0].Descriptor()
}

func (ContextType) Type() protoreflect.EnumType {
	return &file_goserver_proto_enumTypes[0]
}

func (x ContextType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ContextType.Descriptor instead.
func (ContextType) EnumDescriptor() ([]byte, []int) {
	return file_goserver_proto_rawDescGZIP(), []int{0}
}

type LeadState int32

const (
	LeadState_ELECTING LeadState = 0
	LeadState_FOLLOWER LeadState = 1
	LeadState_LEADER   LeadState = 2
)

// Enum value maps for LeadState.
var (
	LeadState_name = map[int32]string{
		0: "ELECTING",
		1: "FOLLOWER",
		2: "LEADER",
	}
	LeadState_value = map[string]int32{
		"ELECTING": 0,
		"FOLLOWER": 1,
		"LEADER":   2,
	}
)

func (x LeadState) Enum() *LeadState {
	p := new(LeadState)
	*p = x
	return p
}

func (x LeadState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (LeadState) Descriptor() protoreflect.EnumDescriptor {
	return file_goserver_proto_enumTypes[1].Descriptor()
}

func (LeadState) Type() protoreflect.EnumType {
	return &file_goserver_proto_enumTypes[1]
}

func (x LeadState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use LeadState.Descriptor instead.
func (LeadState) EnumDescriptor() ([]byte, []int) {
	return file_goserver_proto_rawDescGZIP(), []int{1}
}

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	if protoimpl.UnsafeEnabled {
		mi := &file_goserver_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_goserver_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_goserver_proto_rawDescGZIP(), []int{0}
}

type Alive struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *Alive) Reset() {
	*x = Alive{}
	if protoimpl.UnsafeEnabled {
		mi := &file_goserver_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Alive) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Alive) ProtoMessage() {}

func (x *Alive) ProtoReflect() protoreflect.Message {
	mi := &file_goserver_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Alive.ProtoReflect.Descriptor instead.
func (*Alive) Descriptor() ([]byte, []int) {
	return file_goserver_proto_rawDescGZIP(), []int{1}
}

func (x *Alive) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type MoteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Master bool `protobuf:"varint,1,opt,name=master,proto3" json:"master,omitempty"`
}

func (x *MoteRequest) Reset() {
	*x = MoteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_goserver_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MoteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MoteRequest) ProtoMessage() {}

func (x *MoteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_goserver_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MoteRequest.ProtoReflect.Descriptor instead.
func (*MoteRequest) Descriptor() ([]byte, []int) {
	return file_goserver_proto_rawDescGZIP(), []int{2}
}

func (x *MoteRequest) GetMaster() bool {
	if x != nil {
		return x.Master
	}
	return false
}

type EmbeddedTest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Blah *Alive `protobuf:"bytes,1,opt,name=blah,proto3" json:"blah,omitempty"`
	Test string `protobuf:"bytes,2,opt,name=test,proto3" json:"test,omitempty"`
}

func (x *EmbeddedTest) Reset() {
	*x = EmbeddedTest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_goserver_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EmbeddedTest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EmbeddedTest) ProtoMessage() {}

func (x *EmbeddedTest) ProtoReflect() protoreflect.Message {
	mi := &file_goserver_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EmbeddedTest.ProtoReflect.Descriptor instead.
func (*EmbeddedTest) Descriptor() ([]byte, []int) {
	return file_goserver_proto_rawDescGZIP(), []int{3}
}

func (x *EmbeddedTest) GetBlah() *Alive {
	if x != nil {
		return x.Blah
	}
	return nil
}

func (x *EmbeddedTest) GetTest() string {
	if x != nil {
		return x.Test
	}
	return ""
}

type State struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key          string  `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	TimeValue    int64   `protobuf:"varint,2,opt,name=time_value,json=timeValue,proto3" json:"time_value,omitempty"`
	Value        int64   `protobuf:"varint,3,opt,name=value,proto3" json:"value,omitempty"`
	Text         string  `protobuf:"bytes,4,opt,name=text,proto3" json:"text,omitempty"`
	Fraction     float64 `protobuf:"fixed64,5,opt,name=fraction,proto3" json:"fraction,omitempty"`
	TimeDuration int64   `protobuf:"varint,6,opt,name=time_duration,json=timeDuration,proto3" json:"time_duration,omitempty"`
}

func (x *State) Reset() {
	*x = State{}
	if protoimpl.UnsafeEnabled {
		mi := &file_goserver_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *State) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*State) ProtoMessage() {}

func (x *State) ProtoReflect() protoreflect.Message {
	mi := &file_goserver_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use State.ProtoReflect.Descriptor instead.
func (*State) Descriptor() ([]byte, []int) {
	return file_goserver_proto_rawDescGZIP(), []int{4}
}

func (x *State) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *State) GetTimeValue() int64 {
	if x != nil {
		return x.TimeValue
	}
	return 0
}

func (x *State) GetValue() int64 {
	if x != nil {
		return x.Value
	}
	return 0
}

func (x *State) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *State) GetFraction() float64 {
	if x != nil {
		return x.Fraction
	}
	return 0
}

func (x *State) GetTimeDuration() int64 {
	if x != nil {
		return x.TimeDuration
	}
	return 0
}

type ServerState struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	States []*State `protobuf:"bytes,1,rep,name=states,proto3" json:"states,omitempty"`
}

func (x *ServerState) Reset() {
	*x = ServerState{}
	if protoimpl.UnsafeEnabled {
		mi := &file_goserver_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServerState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerState) ProtoMessage() {}

func (x *ServerState) ProtoReflect() protoreflect.Message {
	mi := &file_goserver_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerState.ProtoReflect.Descriptor instead.
func (*ServerState) Descriptor() ([]byte, []int) {
	return file_goserver_proto_rawDescGZIP(), []int{5}
}

func (x *ServerState) GetStates() []*State {
	if x != nil {
		return x.States
	}
	return nil
}

type TaskPeriod struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key    string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Period int64  `protobuf:"varint,2,opt,name=period,proto3" json:"period,omitempty"`
}

func (x *TaskPeriod) Reset() {
	*x = TaskPeriod{}
	if protoimpl.UnsafeEnabled {
		mi := &file_goserver_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskPeriod) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskPeriod) ProtoMessage() {}

func (x *TaskPeriod) ProtoReflect() protoreflect.Message {
	mi := &file_goserver_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskPeriod.ProtoReflect.Descriptor instead.
func (*TaskPeriod) Descriptor() ([]byte, []int) {
	return file_goserver_proto_rawDescGZIP(), []int{6}
}

func (x *TaskPeriod) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *TaskPeriod) GetPeriod() int64 {
	if x != nil {
		return x.Period
	}
	return 0
}

type ServerConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Periods []*TaskPeriod `protobuf:"bytes,1,rep,name=periods,proto3" json:"periods,omitempty"`
}

func (x *ServerConfig) Reset() {
	*x = ServerConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_goserver_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfig) ProtoMessage() {}

func (x *ServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_goserver_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerConfig.ProtoReflect.Descriptor instead.
func (*ServerConfig) Descriptor() ([]byte, []int) {
	return file_goserver_proto_rawDescGZIP(), []int{7}
}

func (x *ServerConfig) GetPeriods() []*TaskPeriod {
	if x != nil {
		return x.Periods
	}
	return nil
}

type ShutdownRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ShutdownRequest) Reset() {
	*x = ShutdownRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_goserver_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShutdownRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShutdownRequest) ProtoMessage() {}

func (x *ShutdownRequest) ProtoReflect() protoreflect.Message {
	mi := &file_goserver_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShutdownRequest.ProtoReflect.Descriptor instead.
func (*ShutdownRequest) Descriptor() ([]byte, []int) {
	return file_goserver_proto_rawDescGZIP(), []int{8}
}

type ShutdownResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ShutdownResponse) Reset() {
	*x = ShutdownResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_goserver_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShutdownResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShutdownResponse) ProtoMessage() {}

func (x *ShutdownResponse) ProtoReflect() protoreflect.Message {
	mi := &file_goserver_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShutdownResponse.ProtoReflect.Descriptor instead.
func (*ShutdownResponse) Descriptor() ([]byte, []int) {
	return file_goserver_proto_rawDescGZIP(), []int{9}
}

type ReregisterRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ReregisterRequest) Reset() {
	*x = ReregisterRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_goserver_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReregisterRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReregisterRequest) ProtoMessage() {}

func (x *ReregisterRequest) ProtoReflect() protoreflect.Message {
	mi := &file_goserver_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReregisterRequest.ProtoReflect.Descriptor instead.
func (*ReregisterRequest) Descriptor() ([]byte, []int) {
	return file_goserver_proto_rawDescGZIP(), []int{10}
}

type ReregisterResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ReregisterResponse) Reset() {
	*x = ReregisterResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_goserver_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReregisterResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReregisterResponse) ProtoMessage() {}

func (x *ReregisterResponse) ProtoReflect() protoreflect.Message {
	mi := &file_goserver_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReregisterResponse.ProtoReflect.Descriptor instead.
func (*ReregisterResponse) Descriptor() ([]byte, []int) {
	return file_goserver_proto_rawDescGZIP(), []int{11}
}

type ChooseLeadRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Server string `protobuf:"bytes,1,opt,name=server,proto3" json:"server,omitempty"`
}

func (x *ChooseLeadRequest) Reset() {
	*x = ChooseLeadRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_goserver_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChooseLeadRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChooseLeadRequest) ProtoMessage() {}

func (x *ChooseLeadRequest) ProtoReflect() protoreflect.Message {
	mi := &file_goserver_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChooseLeadRequest.ProtoReflect.Descriptor instead.
func (*ChooseLeadRequest) Descriptor() ([]byte, []int) {
	return file_goserver_proto_rawDescGZIP(), []int{12}
}

func (x *ChooseLeadRequest) GetServer() string {
	if x != nil {
		return x.Server
	}
	return ""
}

type ChooseLeadResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Chosen string `protobuf:"bytes,2,opt,name=chosen,proto3" json:"chosen,omitempty"`
}

func (x *ChooseLeadResponse) Reset() {
	*x = ChooseLeadResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_goserver_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChooseLeadResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChooseLeadResponse) ProtoMessage() {}

func (x *ChooseLeadResponse) ProtoReflect() protoreflect.Message {
	mi := &file_goserver_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChooseLeadResponse.ProtoReflect.Descriptor instead.
func (*ChooseLeadResponse) Descriptor() ([]byte, []int) {
	return file_goserver_proto_rawDescGZIP(), []int{13}
}

func (x *ChooseLeadResponse) GetChosen() string {
	if x != nil {
		return x.Chosen
	}
	return ""
}

var File_goserver_proto protoreflect.FileDescriptor

var file_goserver_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x67, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x08, 0x67, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x22, 0x1b, 0x0a, 0x05, 0x41, 0x6c, 0x69, 0x76, 0x65, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x22, 0x25, 0x0a, 0x0b, 0x4d, 0x6f, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x16, 0x0a, 0x06, 0x6d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x06, 0x6d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x22, 0x47, 0x0a, 0x0c, 0x45, 0x6d, 0x62, 0x65, 0x64,
	0x64, 0x65, 0x64, 0x54, 0x65, 0x73, 0x74, 0x12, 0x23, 0x0a, 0x04, 0x62, 0x6c, 0x61, 0x68, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x67, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x2e, 0x41, 0x6c, 0x69, 0x76, 0x65, 0x52, 0x04, 0x62, 0x6c, 0x61, 0x68, 0x12, 0x12, 0x0a, 0x04,
	0x74, 0x65, 0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x73, 0x74,
	0x22, 0xa3, 0x01, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x1d, 0x0a, 0x0a,
	0x74, 0x69, 0x6d, 0x65, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x74, 0x65, 0x78, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x72, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x01, 0x52, 0x08, 0x66, 0x72, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x23, 0x0a, 0x0d, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0c, 0x74, 0x69, 0x6d, 0x65, 0x44, 0x75,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x36, 0x0a, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x27, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x65, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x67, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x65, 0x73, 0x22, 0x36,
	0x0a, 0x0a, 0x54, 0x61, 0x73, 0x6b, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x16,
	0x0a, 0x06, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06,
	0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x22, 0x3e, 0x0a, 0x0c, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x2e, 0x0a, 0x07, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x52, 0x07, 0x70,
	0x65, 0x72, 0x69, 0x6f, 0x64, 0x73, 0x22, 0x11, 0x0a, 0x0f, 0x53, 0x68, 0x75, 0x74, 0x64, 0x6f,
	0x77, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x12, 0x0a, 0x10, 0x53, 0x68, 0x75,
	0x74, 0x64, 0x6f, 0x77, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x13, 0x0a,
	0x11, 0x52, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x22, 0x14, 0x0a, 0x12, 0x52, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2b, 0x0a, 0x11, 0x43, 0x68, 0x6f, 0x6f,
	0x73, 0x65, 0x4c, 0x65, 0x61, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a,
	0x06, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x22, 0x2c, 0x0a, 0x12, 0x43, 0x68, 0x6f, 0x6f, 0x73, 0x65, 0x4c,
	0x65, 0x61, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x63,
	0x68, 0x6f, 0x73, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x63, 0x68, 0x6f,
	0x73, 0x65, 0x6e, 0x2a, 0x59, 0x0a, 0x0b, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12,
	0x0b, 0x0a, 0x07, 0x52, 0x45, 0x47, 0x55, 0x4c, 0x41, 0x52, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06,
	0x4d, 0x45, 0x44, 0x49, 0x55, 0x4d, 0x10, 0x04, 0x12, 0x08, 0x0a, 0x04, 0x4c, 0x4f, 0x4e, 0x47,
	0x10, 0x02, 0x12, 0x0c, 0x0a, 0x08, 0x49, 0x4e, 0x46, 0x49, 0x4e, 0x49, 0x54, 0x45, 0x10, 0x03,
	0x12, 0x0c, 0x0a, 0x08, 0x4e, 0x4f, 0x5f, 0x54, 0x52, 0x41, 0x43, 0x45, 0x10, 0x05, 0x2a, 0x33,
	0x0a, 0x09, 0x4c, 0x65, 0x61, 0x64, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x0c, 0x0a, 0x08, 0x45,
	0x4c, 0x45, 0x43, 0x54, 0x49, 0x4e, 0x47, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x46, 0x4f, 0x4c,
	0x4c, 0x4f, 0x57, 0x45, 0x52, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x4c, 0x45, 0x41, 0x44, 0x45,
	0x52, 0x10, 0x02, 0x32, 0xce, 0x02, 0x0a, 0x0f, 0x67, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x2d, 0x0a, 0x07, 0x49, 0x73, 0x41, 0x6c, 0x69,
	0x76, 0x65, 0x12, 0x0f, 0x2e, 0x67, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x41, 0x6c,
	0x69, 0x76, 0x65, 0x1a, 0x0f, 0x2e, 0x67, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x41,
	0x6c, 0x69, 0x76, 0x65, 0x22, 0x00, 0x12, 0x31, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12,
	0x0f, 0x2e, 0x67, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x1a, 0x15, 0x2e, 0x67, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x65, 0x22, 0x00, 0x12, 0x43, 0x0a, 0x08, 0x53, 0x68, 0x75,
	0x74, 0x64, 0x6f, 0x77, 0x6e, 0x12, 0x19, 0x2e, 0x67, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x2e, 0x53, 0x68, 0x75, 0x74, 0x64, 0x6f, 0x77, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x1a, 0x2e, 0x67, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x53, 0x68, 0x75, 0x74,
	0x64, 0x6f, 0x77, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x49,
	0x0a, 0x0a, 0x52, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x12, 0x1b, 0x2e, 0x67,
	0x6f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74,
	0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x67, 0x6f, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x49, 0x0a, 0x0a, 0x43, 0x68, 0x6f,
	0x6f, 0x73, 0x65, 0x4c, 0x65, 0x61, 0x64, 0x12, 0x1b, 0x2e, 0x67, 0x6f, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x2e, 0x43, 0x68, 0x6f, 0x6f, 0x73, 0x65, 0x4c, 0x65, 0x61, 0x64, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x67, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e,
	0x43, 0x68, 0x6f, 0x6f, 0x73, 0x65, 0x4c, 0x65, 0x61, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x3b, 0x67, 0x6f, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_goserver_proto_rawDescOnce sync.Once
	file_goserver_proto_rawDescData = file_goserver_proto_rawDesc
)

func file_goserver_proto_rawDescGZIP() []byte {
	file_goserver_proto_rawDescOnce.Do(func() {
		file_goserver_proto_rawDescData = protoimpl.X.CompressGZIP(file_goserver_proto_rawDescData)
	})
	return file_goserver_proto_rawDescData
}

var file_goserver_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_goserver_proto_msgTypes = make([]protoimpl.MessageInfo, 14)
var file_goserver_proto_goTypes = []interface{}{
	(ContextType)(0),           // 0: goserver.ContextType
	(LeadState)(0),             // 1: goserver.LeadState
	(*Empty)(nil),              // 2: goserver.Empty
	(*Alive)(nil),              // 3: goserver.Alive
	(*MoteRequest)(nil),        // 4: goserver.MoteRequest
	(*EmbeddedTest)(nil),       // 5: goserver.EmbeddedTest
	(*State)(nil),              // 6: goserver.State
	(*ServerState)(nil),        // 7: goserver.ServerState
	(*TaskPeriod)(nil),         // 8: goserver.TaskPeriod
	(*ServerConfig)(nil),       // 9: goserver.ServerConfig
	(*ShutdownRequest)(nil),    // 10: goserver.ShutdownRequest
	(*ShutdownResponse)(nil),   // 11: goserver.ShutdownResponse
	(*ReregisterRequest)(nil),  // 12: goserver.ReregisterRequest
	(*ReregisterResponse)(nil), // 13: goserver.ReregisterResponse
	(*ChooseLeadRequest)(nil),  // 14: goserver.ChooseLeadRequest
	(*ChooseLeadResponse)(nil), // 15: goserver.ChooseLeadResponse
}
var file_goserver_proto_depIdxs = []int32{
	3,  // 0: goserver.EmbeddedTest.blah:type_name -> goserver.Alive
	6,  // 1: goserver.ServerState.states:type_name -> goserver.State
	8,  // 2: goserver.ServerConfig.periods:type_name -> goserver.TaskPeriod
	3,  // 3: goserver.goserverService.IsAlive:input_type -> goserver.Alive
	2,  // 4: goserver.goserverService.State:input_type -> goserver.Empty
	10, // 5: goserver.goserverService.Shutdown:input_type -> goserver.ShutdownRequest
	12, // 6: goserver.goserverService.Reregister:input_type -> goserver.ReregisterRequest
	14, // 7: goserver.goserverService.ChooseLead:input_type -> goserver.ChooseLeadRequest
	3,  // 8: goserver.goserverService.IsAlive:output_type -> goserver.Alive
	7,  // 9: goserver.goserverService.State:output_type -> goserver.ServerState
	11, // 10: goserver.goserverService.Shutdown:output_type -> goserver.ShutdownResponse
	13, // 11: goserver.goserverService.Reregister:output_type -> goserver.ReregisterResponse
	15, // 12: goserver.goserverService.ChooseLead:output_type -> goserver.ChooseLeadResponse
	8,  // [8:13] is the sub-list for method output_type
	3,  // [3:8] is the sub-list for method input_type
	3,  // [3:3] is the sub-list for extension type_name
	3,  // [3:3] is the sub-list for extension extendee
	0,  // [0:3] is the sub-list for field type_name
}

func init() { file_goserver_proto_init() }
func file_goserver_proto_init() {
	if File_goserver_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_goserver_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Empty); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_goserver_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Alive); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_goserver_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MoteRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_goserver_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EmbeddedTest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_goserver_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*State); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_goserver_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServerState); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_goserver_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskPeriod); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_goserver_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServerConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_goserver_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShutdownRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_goserver_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShutdownResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_goserver_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReregisterRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_goserver_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReregisterResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_goserver_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChooseLeadRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_goserver_proto_msgTypes[13].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChooseLeadResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_goserver_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   14,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_goserver_proto_goTypes,
		DependencyIndexes: file_goserver_proto_depIdxs,
		EnumInfos:         file_goserver_proto_enumTypes,
		MessageInfos:      file_goserver_proto_msgTypes,
	}.Build()
	File_goserver_proto = out.File
	file_goserver_proto_rawDesc = nil
	file_goserver_proto_goTypes = nil
	file_goserver_proto_depIdxs = nil
}

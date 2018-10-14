// Code generated by protoc-gen-go. DO NOT EDIT.
// source: server.proto

/*
Package goserver is a generated protocol buffer package.

It is generated from these files:
	server.proto

It has these top-level messages:
	Empty
	Alive
	MoteRequest
	EmbeddedTest
	State
	ServerState
	TaskPeriod
	ServerConfig
*/
package goserver

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
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
func (ContextType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Empty struct {
}

func (m *Empty) Reset()                    { *m = Empty{} }
func (m *Empty) String() string            { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()               {}
func (*Empty) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Alive struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
}

func (m *Alive) Reset()                    { *m = Alive{} }
func (m *Alive) String() string            { return proto.CompactTextString(m) }
func (*Alive) ProtoMessage()               {}
func (*Alive) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Alive) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type MoteRequest struct {
	Master bool `protobuf:"varint,1,opt,name=master" json:"master,omitempty"`
}

func (m *MoteRequest) Reset()                    { *m = MoteRequest{} }
func (m *MoteRequest) String() string            { return proto.CompactTextString(m) }
func (*MoteRequest) ProtoMessage()               {}
func (*MoteRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *MoteRequest) GetMaster() bool {
	if m != nil {
		return m.Master
	}
	return false
}

type EmbeddedTest struct {
	Blah *Alive `protobuf:"bytes,1,opt,name=blah" json:"blah,omitempty"`
	Test string `protobuf:"bytes,2,opt,name=test" json:"test,omitempty"`
}

func (m *EmbeddedTest) Reset()                    { *m = EmbeddedTest{} }
func (m *EmbeddedTest) String() string            { return proto.CompactTextString(m) }
func (*EmbeddedTest) ProtoMessage()               {}
func (*EmbeddedTest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

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
	Key       string  `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	TimeValue int64   `protobuf:"varint,2,opt,name=time_value,json=timeValue" json:"time_value,omitempty"`
	Value     int64   `protobuf:"varint,3,opt,name=value" json:"value,omitempty"`
	Text      string  `protobuf:"bytes,4,opt,name=text" json:"text,omitempty"`
	Fraction  float64 `protobuf:"fixed64,5,opt,name=fraction" json:"fraction,omitempty"`
}

func (m *State) Reset()                    { *m = State{} }
func (m *State) String() string            { return proto.CompactTextString(m) }
func (*State) ProtoMessage()               {}
func (*State) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

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

type ServerState struct {
	States []*State `protobuf:"bytes,1,rep,name=states" json:"states,omitempty"`
}

func (m *ServerState) Reset()                    { *m = ServerState{} }
func (m *ServerState) String() string            { return proto.CompactTextString(m) }
func (*ServerState) ProtoMessage()               {}
func (*ServerState) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *ServerState) GetStates() []*State {
	if m != nil {
		return m.States
	}
	return nil
}

type TaskPeriod struct {
	Key    string `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	Period int64  `protobuf:"varint,2,opt,name=period" json:"period,omitempty"`
}

func (m *TaskPeriod) Reset()                    { *m = TaskPeriod{} }
func (m *TaskPeriod) String() string            { return proto.CompactTextString(m) }
func (*TaskPeriod) ProtoMessage()               {}
func (*TaskPeriod) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

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
	Periods []*TaskPeriod `protobuf:"bytes,1,rep,name=periods" json:"periods,omitempty"`
}

func (m *ServerConfig) Reset()                    { *m = ServerConfig{} }
func (m *ServerConfig) String() string            { return proto.CompactTextString(m) }
func (*ServerConfig) ProtoMessage()               {}
func (*ServerConfig) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *ServerConfig) GetPeriods() []*TaskPeriod {
	if m != nil {
		return m.Periods
	}
	return nil
}

func init() {
	proto.RegisterType((*Empty)(nil), "goserver.Empty")
	proto.RegisterType((*Alive)(nil), "goserver.Alive")
	proto.RegisterType((*MoteRequest)(nil), "goserver.MoteRequest")
	proto.RegisterType((*EmbeddedTest)(nil), "goserver.EmbeddedTest")
	proto.RegisterType((*State)(nil), "goserver.State")
	proto.RegisterType((*ServerState)(nil), "goserver.ServerState")
	proto.RegisterType((*TaskPeriod)(nil), "goserver.TaskPeriod")
	proto.RegisterType((*ServerConfig)(nil), "goserver.ServerConfig")
	proto.RegisterEnum("goserver.ContextType", ContextType_name, ContextType_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for GoserverService service

type GoserverServiceClient interface {
	IsAlive(ctx context.Context, in *Alive, opts ...grpc.CallOption) (*Alive, error)
	Mote(ctx context.Context, in *MoteRequest, opts ...grpc.CallOption) (*Empty, error)
	State(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ServerState, error)
}

type goserverServiceClient struct {
	cc *grpc.ClientConn
}

func NewGoserverServiceClient(cc *grpc.ClientConn) GoserverServiceClient {
	return &goserverServiceClient{cc}
}

func (c *goserverServiceClient) IsAlive(ctx context.Context, in *Alive, opts ...grpc.CallOption) (*Alive, error) {
	out := new(Alive)
	err := grpc.Invoke(ctx, "/goserver.goserverService/IsAlive", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *goserverServiceClient) Mote(ctx context.Context, in *MoteRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := grpc.Invoke(ctx, "/goserver.goserverService/Mote", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *goserverServiceClient) State(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ServerState, error) {
	out := new(ServerState)
	err := grpc.Invoke(ctx, "/goserver.goserverService/State", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for GoserverService service

type GoserverServiceServer interface {
	IsAlive(context.Context, *Alive) (*Alive, error)
	Mote(context.Context, *MoteRequest) (*Empty, error)
	State(context.Context, *Empty) (*ServerState, error)
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
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server.proto",
}

func init() { proto.RegisterFile("server.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 443 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x92, 0x6f, 0x6b, 0xd3, 0x50,
	0x14, 0xc6, 0x9b, 0xe5, 0x4f, 0xb3, 0x93, 0xc2, 0xc2, 0x61, 0x8e, 0x50, 0x11, 0xca, 0x15, 0x71,
	0x08, 0x16, 0x9d, 0xb0, 0x97, 0x42, 0xa9, 0xb1, 0x04, 0xd7, 0x54, 0xee, 0x52, 0xc5, 0x57, 0x23,
	0x5d, 0xcf, 0x66, 0x58, 0xd3, 0xd4, 0xe4, 0xae, 0xb4, 0x6f, 0xfc, 0x36, 0x7e, 0x4f, 0xb9, 0xf7,
	0x26, 0xb6, 0x96, 0xbd, 0x3b, 0xcf, 0x79, 0x4e, 0xf2, 0x7b, 0x2e, 0xe7, 0x40, 0xa7, 0xa2, 0x72,
	0x4d, 0x65, 0x7f, 0x55, 0x16, 0xa2, 0x40, 0xf7, 0xbe, 0xd0, 0x9a, 0xb5, 0xc1, 0x0e, 0xf3, 0x95,
	0xd8, 0xb2, 0xe7, 0x60, 0x0f, 0x16, 0xd9, 0x9a, 0x10, 0xc1, 0x5a, 0xa6, 0x39, 0x05, 0x46, 0xcf,
	0x38, 0x3f, 0xe6, 0xaa, 0x66, 0xaf, 0xc0, 0x1b, 0x17, 0x82, 0x38, 0xfd, 0x7a, 0xa4, 0x4a, 0xe0,
	0x19, 0x38, 0x79, 0x5a, 0x09, 0x2a, 0xd5, 0x90, 0xcb, 0x6b, 0xc5, 0x46, 0xd0, 0x09, 0xf3, 0x19,
	0xcd, 0xe7, 0x34, 0x4f, 0xe4, 0xdc, 0x4b, 0xb0, 0x66, 0x8b, 0xf4, 0xa7, 0x9a, 0xf2, 0x2e, 0x4e,
	0xfa, 0x0d, 0xb5, 0xaf, 0x48, 0x5c, 0x99, 0x92, 0x27, 0xa8, 0x12, 0xc1, 0x91, 0xe6, 0xc9, 0x9a,
	0xfd, 0x06, 0xfb, 0x5a, 0xa4, 0x82, 0xd0, 0x07, 0xf3, 0x81, 0xb6, 0x75, 0x16, 0x59, 0xe2, 0x0b,
	0x00, 0x91, 0xe5, 0x74, 0xb3, 0x4e, 0x17, 0x8f, 0xa4, 0x3e, 0x32, 0xf9, 0xb1, 0xec, 0x7c, 0x93,
	0x0d, 0x3c, 0x05, 0x5b, 0x3b, 0xa6, 0x72, 0xb4, 0xd0, 0x8c, 0x8d, 0x08, 0xac, 0x86, 0xb1, 0x11,
	0xd8, 0x05, 0xf7, 0xae, 0x4c, 0x6f, 0x45, 0x56, 0x2c, 0x03, 0xbb, 0x67, 0x9c, 0x1b, 0xfc, 0x9f,
	0x66, 0x97, 0xe0, 0x5d, 0xab, 0xa4, 0x3a, 0xc5, 0x6b, 0x70, 0x2a, 0x59, 0x54, 0x81, 0xd1, 0x33,
	0xff, 0x7f, 0x89, 0x1a, 0xe0, 0xb5, 0xcd, 0x2e, 0x01, 0x92, 0xb4, 0x7a, 0xf8, 0x4a, 0x65, 0x56,
	0xcc, 0x9f, 0x08, 0x7f, 0x06, 0xce, 0x4a, 0x79, 0x75, 0xf0, 0x5a, 0xb1, 0x8f, 0xd0, 0xd1, 0xbc,
	0x61, 0xb1, 0xbc, 0xcb, 0xee, 0xb1, 0x0f, 0x6d, 0xed, 0x34, 0xc4, 0xd3, 0x1d, 0x71, 0x07, 0xe0,
	0xcd, 0xd0, 0x9b, 0x1f, 0xe0, 0x0d, 0x8b, 0xa5, 0x7c, 0x56, 0xb2, 0x5d, 0x11, 0x7a, 0xd0, 0x9e,
	0xc6, 0x5f, 0xe2, 0xc9, 0xf7, 0xd8, 0x6f, 0x49, 0xc1, 0xc3, 0xd1, 0xf4, 0x6a, 0xc0, 0x7d, 0x03,
	0x01, 0x9c, 0x71, 0xf8, 0x29, 0x9a, 0x8e, 0x7d, 0x0b, 0x5d, 0xb0, 0xae, 0x26, 0xf1, 0xc8, 0x3f,
	0xc2, 0x0e, 0xb8, 0x51, 0xfc, 0x39, 0x8a, 0xa3, 0x24, 0xf4, 0x4d, 0xa9, 0xe2, 0xc9, 0x4d, 0xc2,
	0x07, 0xc3, 0xd0, 0xb7, 0x2f, 0xfe, 0x18, 0x70, 0xd2, 0xb0, 0x65, 0xc6, 0xec, 0x96, 0xf0, 0x2d,
	0xb4, 0xa3, 0x4a, 0x5f, 0xcb, 0xe1, 0x52, 0xbb, 0x87, 0x0d, 0xd6, 0xc2, 0x77, 0x60, 0xc9, 0xeb,
	0xc1, 0x67, 0x3b, 0x6b, 0xef, 0x9a, 0xf6, 0xbf, 0xd0, 0xa7, 0xd8, 0xc2, 0xf7, 0xcd, 0xfe, 0x0f,
	0xbd, 0xee, 0xde, 0x3f, 0xf6, 0x36, 0xc4, 0x5a, 0x33, 0x47, 0x5d, 0xf6, 0x87, 0xbf, 0x01, 0x00,
	0x00, 0xff, 0xff, 0xa3, 0x95, 0xfb, 0x6c, 0xe9, 0x02, 0x00, 0x00,
}

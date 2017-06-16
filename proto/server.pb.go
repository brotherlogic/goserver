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

type Empty struct {
}

func (m *Empty) Reset()                    { *m = Empty{} }
func (m *Empty) String() string            { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()               {}
func (*Empty) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Alive struct {
}

func (m *Alive) Reset()                    { *m = Alive{} }
func (m *Alive) String() string            { return proto.CompactTextString(m) }
func (*Alive) ProtoMessage()               {}
func (*Alive) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

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

func init() {
	proto.RegisterType((*Empty)(nil), "goserver.Empty")
	proto.RegisterType((*Alive)(nil), "goserver.Alive")
	proto.RegisterType((*MoteRequest)(nil), "goserver.MoteRequest")
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

// Server API for GoserverService service

type GoserverServiceServer interface {
	IsAlive(context.Context, *Alive) (*Alive, error)
	Mote(context.Context, *MoteRequest) (*Empty, error)
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
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server.proto",
}

func init() { proto.RegisterFile("server.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 150 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x29, 0x4e, 0x2d, 0x2a,
	0x4b, 0x2d, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x48, 0xcf, 0x87, 0xf0, 0x95, 0xd8,
	0xb9, 0x58, 0x5d, 0x73, 0x0b, 0x4a, 0x2a, 0x41, 0x0c, 0xc7, 0x9c, 0xcc, 0xb2, 0x54, 0x25, 0x55,
	0x2e, 0x6e, 0xdf, 0xfc, 0x92, 0xd4, 0xa0, 0xd4, 0xc2, 0xd2, 0xd4, 0xe2, 0x12, 0x21, 0x31, 0x2e,
	0xb6, 0xdc, 0xc4, 0xe2, 0x92, 0xd4, 0x22, 0x09, 0x46, 0x05, 0x46, 0x0d, 0x8e, 0x20, 0x28, 0xcf,
	0xa8, 0x88, 0x8b, 0x1f, 0x66, 0x48, 0x70, 0x6a, 0x51, 0x59, 0x66, 0x72, 0xaa, 0x90, 0x2e, 0x17,
	0xbb, 0x67, 0x31, 0xd8, 0x10, 0x21, 0x7e, 0x3d, 0x98, 0xa4, 0x1e, 0x58, 0x40, 0x0a, 0x5d, 0x40,
	0x89, 0x41, 0xc8, 0x80, 0x8b, 0x05, 0x64, 0x91, 0x90, 0x28, 0x42, 0x0a, 0xc9, 0x62, 0x64, 0x1d,
	0x10, 0x17, 0x32, 0x24, 0xb1, 0x81, 0x5d, 0x6f, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0x81, 0x9a,
	0x91, 0x26, 0xcd, 0x00, 0x00, 0x00,
}

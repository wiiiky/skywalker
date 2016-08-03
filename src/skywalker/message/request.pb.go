// Code generated by protoc-gen-go.
// source: src/skywalker/message/request.proto
// DO NOT EDIT!

/*
Package message is a generated protocol buffer package.

It is generated from these files:
	src/skywalker/message/request.proto
	src/skywalker/message/response.proto

It has these top-level messages:
	StatusRequest
	Request
	StatusResponse
	Response
*/
package message

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type RequestType int32

const (
	RequestType_STATUS RequestType = 1
)

var RequestType_name = map[int32]string{
	1: "STATUS",
}
var RequestType_value = map[string]int32{
	"STATUS": 1,
}

func (x RequestType) Enum() *RequestType {
	p := new(RequestType)
	*p = x
	return p
}
func (x RequestType) String() string {
	return proto.EnumName(RequestType_name, int32(x))
}
func (x *RequestType) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(RequestType_value, data, "RequestType")
	if err != nil {
		return err
	}
	*x = RequestType(value)
	return nil
}
func (RequestType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type StatusRequest struct {
	Name             []string `protobuf:"bytes,1,rep,name=name" json:"name,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *StatusRequest) Reset()                    { *m = StatusRequest{} }
func (m *StatusRequest) String() string            { return proto.CompactTextString(m) }
func (*StatusRequest) ProtoMessage()               {}
func (*StatusRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *StatusRequest) GetName() []string {
	if m != nil {
		return m.Name
	}
	return nil
}

type Request struct {
	Type             *RequestType   `protobuf:"varint,1,req,name=type,enum=message.RequestType" json:"type,omitempty"`
	Status           *StatusRequest `protobuf:"bytes,2,opt,name=status" json:"status,omitempty"`
	XXX_unrecognized []byte         `json:"-"`
}

func (m *Request) Reset()                    { *m = Request{} }
func (m *Request) String() string            { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()               {}
func (*Request) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Request) GetType() RequestType {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return RequestType_STATUS
}

func (m *Request) GetStatus() *StatusRequest {
	if m != nil {
		return m.Status
	}
	return nil
}

func init() {
	proto.RegisterType((*StatusRequest)(nil), "message.StatusRequest")
	proto.RegisterType((*Request)(nil), "message.Request")
	proto.RegisterEnum("message.RequestType", RequestType_name, RequestType_value)
}

func init() { proto.RegisterFile("src/skywalker/message/request.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 156 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0x52, 0x2e, 0x2e, 0x4a, 0xd6,
	0x2f, 0xce, 0xae, 0x2c, 0x4f, 0xcc, 0xc9, 0x4e, 0x2d, 0xd2, 0xcf, 0x4d, 0x2d, 0x2e, 0x4e, 0x4c,
	0x4f, 0xd5, 0x2f, 0x4a, 0x2d, 0x2c, 0x4d, 0x2d, 0x2e, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17,
	0x62, 0x87, 0x0a, 0x2b, 0xc9, 0x72, 0xf1, 0x06, 0x97, 0x24, 0x96, 0x94, 0x16, 0x07, 0x41, 0xe4,
	0x85, 0x78, 0xb8, 0x58, 0xf2, 0x12, 0x73, 0x53, 0x25, 0x18, 0x15, 0x98, 0x35, 0x38, 0x95, 0x42,
	0xb9, 0xd8, 0x61, 0x12, 0x4a, 0x5c, 0x2c, 0x25, 0x95, 0x05, 0x20, 0x09, 0x26, 0x0d, 0x3e, 0x23,
	0x11, 0x3d, 0xa8, 0x09, 0x7a, 0x50, 0xf9, 0x10, 0xa0, 0x9c, 0x90, 0x1a, 0x17, 0x5b, 0x31, 0xd8,
	0x34, 0x09, 0x26, 0x05, 0x46, 0x0d, 0x6e, 0x23, 0x31, 0xb8, 0x2a, 0x14, 0x4b, 0xb4, 0x24, 0xb9,
	0xb8, 0x91, 0xb5, 0x71, 0x71, 0xb1, 0x05, 0x87, 0x38, 0x86, 0x84, 0x06, 0x0b, 0x30, 0x02, 0x02,
	0x00, 0x00, 0xff, 0xff, 0x28, 0xf5, 0x1f, 0xd1, 0xbf, 0x00, 0x00, 0x00,
}

// Copyright 2020 Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.22.0
// 	protoc        v3.11.4
// source: proto/bwreserver/v1/bwreserver.proto

package bwreserver

import (
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

type Protocol int32

const (
	// Unspecified protocol
	Protocol_PROTOCOL_UNSPECIFIED Protocol = 0
	// UDP
	Protocol_PROTOCOL_UDP Protocol = 1
	// QUIC
	Protocol_PROTOCOL_QUIC Protocol = 2
)

// Enum value maps for Protocol.
var (
	Protocol_name = map[int32]string{
		0: "PROTOCOL_UNSPECIFIED",
		1: "PROTOCOL_UDP",
		2: "PROTOCOL_QUIC",
	}
	Protocol_value = map[string]int32{
		"PROTOCOL_UNSPECIFIED": 0,
		"PROTOCOL_UDP":         1,
		"PROTOCOL_QUIC":        2,
	}
)

func (x Protocol) Enum() *Protocol {
	p := new(Protocol)
	*p = x
	return p
}

func (x Protocol) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Protocol) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_bwreserver_v1_bwreserver_proto_enumTypes[0].Descriptor()
}

func (Protocol) Type() protoreflect.EnumType {
	return &file_proto_bwreserver_v1_bwreserver_proto_enumTypes[0]
}

func (x Protocol) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Protocol.Descriptor instead.
func (Protocol) EnumDescriptor() ([]byte, []int) {
	return file_proto_bwreserver_v1_bwreserver_proto_rawDescGZIP(), []int{0}
}

type OpenSessionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The source ISD-AS.
	SrcIsdAs uint64 `protobuf:"varint,1,opt,name=src_isd_as,json=srcIsdAs,proto3" json:"src_isd_as,omitempty"`
	// The source host that requests to open a data session.
	SrcHost []byte `protobuf:"bytes,2,opt,name=src_host,json=srcHost,proto3" json:"src_host,omitempty"`
	// The protocol to be used.
	Protocol Protocol `protobuf:"varint,3,opt,name=protocol,proto3,enum=anapaya.proto.bwreserver.v1.Protocol" json:"protocol,omitempty"`
}

func (x *OpenSessionRequest) Reset() {
	*x = OpenSessionRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_bwreserver_v1_bwreserver_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OpenSessionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OpenSessionRequest) ProtoMessage() {}

func (x *OpenSessionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_bwreserver_v1_bwreserver_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OpenSessionRequest.ProtoReflect.Descriptor instead.
func (*OpenSessionRequest) Descriptor() ([]byte, []int) {
	return file_proto_bwreserver_v1_bwreserver_proto_rawDescGZIP(), []int{0}
}

func (x *OpenSessionRequest) GetSrcIsdAs() uint64 {
	if x != nil {
		return x.SrcIsdAs
	}
	return 0
}

func (x *OpenSessionRequest) GetSrcHost() []byte {
	if x != nil {
		return x.SrcHost
	}
	return nil
}

func (x *OpenSessionRequest) GetProtocol() Protocol {
	if x != nil {
		return x.Protocol
	}
	return Protocol_PROTOCOL_UNSPECIFIED
}

type OpenSessionResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The session ID for the data session.
	SessionId uint64 `protobuf:"varint,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	// The destination port.
	Port uint32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	// The protocol to be used.
	Protocol Protocol `protobuf:"varint,3,opt,name=protocol,proto3,enum=anapaya.proto.bwreserver.v1.Protocol" json:"protocol,omitempty"`
}

func (x *OpenSessionResponse) Reset() {
	*x = OpenSessionResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_bwreserver_v1_bwreserver_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OpenSessionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OpenSessionResponse) ProtoMessage() {}

func (x *OpenSessionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_bwreserver_v1_bwreserver_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OpenSessionResponse.ProtoReflect.Descriptor instead.
func (*OpenSessionResponse) Descriptor() ([]byte, []int) {
	return file_proto_bwreserver_v1_bwreserver_proto_rawDescGZIP(), []int{1}
}

func (x *OpenSessionResponse) GetSessionId() uint64 {
	if x != nil {
		return x.SessionId
	}
	return 0
}

func (x *OpenSessionResponse) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *OpenSessionResponse) GetProtocol() Protocol {
	if x != nil {
		return x.Protocol
	}
	return Protocol_PROTOCOL_UNSPECIFIED
}

type StatsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The session ID for the data session.
	SessionId uint64 `protobuf:"varint,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
}

func (x *StatsRequest) Reset() {
	*x = StatsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_bwreserver_v1_bwreserver_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatsRequest) ProtoMessage() {}

func (x *StatsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_bwreserver_v1_bwreserver_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatsRequest.ProtoReflect.Descriptor instead.
func (*StatsRequest) Descriptor() ([]byte, []int) {
	return file_proto_bwreserver_v1_bwreserver_proto_rawDescGZIP(), []int{2}
}

func (x *StatsRequest) GetSessionId() uint64 {
	if x != nil {
		return x.SessionId
	}
	return 0
}

type StatsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The start timestamp of the session.
	Start *timestamp.Timestamp `protobuf:"bytes,1,opt,name=start,proto3" json:"start,omitempty"`
	// The current timestamp.
	Current *timestamp.Timestamp `protobuf:"bytes,2,opt,name=current,proto3" json:"current,omitempty"`
	// The number of bytes received from start to current.
	BytesReceived uint64 `protobuf:"varint,3,opt,name=bytes_received,json=bytesReceived,proto3" json:"bytes_received,omitempty"`
}

func (x *StatsResponse) Reset() {
	*x = StatsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_bwreserver_v1_bwreserver_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatsResponse) ProtoMessage() {}

func (x *StatsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_bwreserver_v1_bwreserver_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatsResponse.ProtoReflect.Descriptor instead.
func (*StatsResponse) Descriptor() ([]byte, []int) {
	return file_proto_bwreserver_v1_bwreserver_proto_rawDescGZIP(), []int{3}
}

func (x *StatsResponse) GetStart() *timestamp.Timestamp {
	if x != nil {
		return x.Start
	}
	return nil
}

func (x *StatsResponse) GetCurrent() *timestamp.Timestamp {
	if x != nil {
		return x.Current
	}
	return nil
}

func (x *StatsResponse) GetBytesReceived() uint64 {
	if x != nil {
		return x.BytesReceived
	}
	return 0
}

var File_proto_bwreserver_v1_bwreserver_proto protoreflect.FileDescriptor

var file_proto_bwreserver_v1_bwreserver_proto_rawDesc = []byte{
	0x0a, 0x24, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x62, 0x77, 0x72, 0x65, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x62, 0x77, 0x72, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1b, 0x61, 0x6e, 0x61, 0x70, 0x61, 0x79, 0x61, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x62, 0x77, 0x72, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x90, 0x01, 0x0a, 0x12, 0x4f, 0x70, 0x65, 0x6e, 0x53, 0x65, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x0a, 0x73,
	0x72, 0x63, 0x5f, 0x69, 0x73, 0x64, 0x5f, 0x61, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x08, 0x73, 0x72, 0x63, 0x49, 0x73, 0x64, 0x41, 0x73, 0x12, 0x19, 0x0a, 0x08, 0x73, 0x72, 0x63,
	0x5f, 0x68, 0x6f, 0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x73, 0x72, 0x63,
	0x48, 0x6f, 0x73, 0x74, 0x12, 0x41, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x25, 0x2e, 0x61, 0x6e, 0x61, 0x70, 0x61, 0x79, 0x61,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x62, 0x77, 0x72, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x52, 0x08, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x22, 0x8b, 0x01, 0x0a, 0x13, 0x4f, 0x70, 0x65, 0x6e,
	0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x09, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x12,
	0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x70, 0x6f,
	0x72, 0x74, 0x12, 0x41, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x25, 0x2e, 0x61, 0x6e, 0x61, 0x70, 0x61, 0x79, 0x61, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x62, 0x77, 0x72, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e,
	0x76, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x52, 0x08, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x22, 0x2d, 0x0a, 0x0c, 0x53, 0x74, 0x61, 0x74, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x73, 0x65, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x49, 0x64, 0x22, 0x9e, 0x01, 0x0a, 0x0d, 0x53, 0x74, 0x61, 0x74, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x30, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x52, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x34, 0x0a, 0x07, 0x63, 0x75, 0x72, 0x72,
	0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x07, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x12, 0x25,
	0x0a, 0x0e, 0x62, 0x79, 0x74, 0x65, 0x73, 0x5f, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0d, 0x62, 0x79, 0x74, 0x65, 0x73, 0x52, 0x65, 0x63,
	0x65, 0x69, 0x76, 0x65, 0x64, 0x2a, 0x49, 0x0a, 0x08, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f,
	0x6c, 0x12, 0x18, 0x0a, 0x14, 0x50, 0x52, 0x4f, 0x54, 0x4f, 0x43, 0x4f, 0x4c, 0x5f, 0x55, 0x4e,
	0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x10, 0x0a, 0x0c, 0x50,
	0x52, 0x4f, 0x54, 0x4f, 0x43, 0x4f, 0x4c, 0x5f, 0x55, 0x44, 0x50, 0x10, 0x01, 0x12, 0x11, 0x0a,
	0x0d, 0x50, 0x52, 0x4f, 0x54, 0x4f, 0x43, 0x4f, 0x4c, 0x5f, 0x51, 0x55, 0x49, 0x43, 0x10, 0x02,
	0x32, 0xe6, 0x01, 0x0a, 0x0e, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x72, 0x0a, 0x0b, 0x4f, 0x70, 0x65, 0x6e, 0x53, 0x65, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x12, 0x2f, 0x2e, 0x61, 0x6e, 0x61, 0x70, 0x61, 0x79, 0x61, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x62, 0x77, 0x72, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x4f, 0x70, 0x65, 0x6e, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x30, 0x2e, 0x61, 0x6e, 0x61, 0x70, 0x61, 0x79, 0x61, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x62, 0x77, 0x72, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76,
	0x31, 0x2e, 0x4f, 0x70, 0x65, 0x6e, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x60, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74, 0x73,
	0x12, 0x29, 0x2e, 0x61, 0x6e, 0x61, 0x70, 0x61, 0x79, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x62, 0x77, 0x72, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x53,
	0x74, 0x61, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2a, 0x2e, 0x61, 0x6e,
	0x61, 0x70, 0x61, 0x79, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x62, 0x77, 0x72, 0x65,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x3d, 0x5a, 0x3b, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x63, 0x69, 0x6f, 0x6e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x73, 0x63, 0x69, 0x6f, 0x6e, 0x2f, 0x61, 0x6e, 0x61, 0x70, 0x61, 0x79, 0x61,
	0x2f, 0x67, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x62, 0x77,
	0x72, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_bwreserver_v1_bwreserver_proto_rawDescOnce sync.Once
	file_proto_bwreserver_v1_bwreserver_proto_rawDescData = file_proto_bwreserver_v1_bwreserver_proto_rawDesc
)

func file_proto_bwreserver_v1_bwreserver_proto_rawDescGZIP() []byte {
	file_proto_bwreserver_v1_bwreserver_proto_rawDescOnce.Do(func() {
		file_proto_bwreserver_v1_bwreserver_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_bwreserver_v1_bwreserver_proto_rawDescData)
	})
	return file_proto_bwreserver_v1_bwreserver_proto_rawDescData
}

var file_proto_bwreserver_v1_bwreserver_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_bwreserver_v1_bwreserver_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_proto_bwreserver_v1_bwreserver_proto_goTypes = []interface{}{
	(Protocol)(0),               // 0: anapaya.proto.bwreserver.v1.Protocol
	(*OpenSessionRequest)(nil),  // 1: anapaya.proto.bwreserver.v1.OpenSessionRequest
	(*OpenSessionResponse)(nil), // 2: anapaya.proto.bwreserver.v1.OpenSessionResponse
	(*StatsRequest)(nil),        // 3: anapaya.proto.bwreserver.v1.StatsRequest
	(*StatsResponse)(nil),       // 4: anapaya.proto.bwreserver.v1.StatsResponse
	(*timestamp.Timestamp)(nil), // 5: google.protobuf.Timestamp
}
var file_proto_bwreserver_v1_bwreserver_proto_depIdxs = []int32{
	0, // 0: anapaya.proto.bwreserver.v1.OpenSessionRequest.protocol:type_name -> anapaya.proto.bwreserver.v1.Protocol
	0, // 1: anapaya.proto.bwreserver.v1.OpenSessionResponse.protocol:type_name -> anapaya.proto.bwreserver.v1.Protocol
	5, // 2: anapaya.proto.bwreserver.v1.StatsResponse.start:type_name -> google.protobuf.Timestamp
	5, // 3: anapaya.proto.bwreserver.v1.StatsResponse.current:type_name -> google.protobuf.Timestamp
	1, // 4: anapaya.proto.bwreserver.v1.SessionService.OpenSession:input_type -> anapaya.proto.bwreserver.v1.OpenSessionRequest
	3, // 5: anapaya.proto.bwreserver.v1.SessionService.Stats:input_type -> anapaya.proto.bwreserver.v1.StatsRequest
	2, // 6: anapaya.proto.bwreserver.v1.SessionService.OpenSession:output_type -> anapaya.proto.bwreserver.v1.OpenSessionResponse
	4, // 7: anapaya.proto.bwreserver.v1.SessionService.Stats:output_type -> anapaya.proto.bwreserver.v1.StatsResponse
	6, // [6:8] is the sub-list for method output_type
	4, // [4:6] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_proto_bwreserver_v1_bwreserver_proto_init() }
func file_proto_bwreserver_v1_bwreserver_proto_init() {
	if File_proto_bwreserver_v1_bwreserver_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_bwreserver_v1_bwreserver_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OpenSessionRequest); i {
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
		file_proto_bwreserver_v1_bwreserver_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OpenSessionResponse); i {
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
		file_proto_bwreserver_v1_bwreserver_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatsRequest); i {
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
		file_proto_bwreserver_v1_bwreserver_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatsResponse); i {
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
			RawDescriptor: file_proto_bwreserver_v1_bwreserver_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_bwreserver_v1_bwreserver_proto_goTypes,
		DependencyIndexes: file_proto_bwreserver_v1_bwreserver_proto_depIdxs,
		EnumInfos:         file_proto_bwreserver_v1_bwreserver_proto_enumTypes,
		MessageInfos:      file_proto_bwreserver_v1_bwreserver_proto_msgTypes,
	}.Build()
	File_proto_bwreserver_v1_bwreserver_proto = out.File
	file_proto_bwreserver_v1_bwreserver_proto_rawDesc = nil
	file_proto_bwreserver_v1_bwreserver_proto_goTypes = nil
	file_proto_bwreserver_v1_bwreserver_proto_depIdxs = nil
}

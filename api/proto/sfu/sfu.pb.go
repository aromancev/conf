// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.13.0
// source: sfu.proto

package sfu

import (
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

type SignalRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Types that are assignable to Payload:
	//	*SignalRequest_Join
	//	*SignalRequest_Offer
	//	*SignalRequest_Answer
	//	*SignalRequest_Trickle
	Payload isSignalRequest_Payload `protobuf_oneof:"payload"`
}

func (x *SignalRequest) Reset() {
	*x = SignalRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sfu_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignalRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignalRequest) ProtoMessage() {}

func (x *SignalRequest) ProtoReflect() protoreflect.Message {
	mi := &file_sfu_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignalRequest.ProtoReflect.Descriptor instead.
func (*SignalRequest) Descriptor() ([]byte, []int) {
	return file_sfu_proto_rawDescGZIP(), []int{0}
}

func (x *SignalRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (m *SignalRequest) GetPayload() isSignalRequest_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (x *SignalRequest) GetJoin() []byte {
	if x, ok := x.GetPayload().(*SignalRequest_Join); ok {
		return x.Join
	}
	return nil
}

func (x *SignalRequest) GetOffer() []byte {
	if x, ok := x.GetPayload().(*SignalRequest_Offer); ok {
		return x.Offer
	}
	return nil
}

func (x *SignalRequest) GetAnswer() []byte {
	if x, ok := x.GetPayload().(*SignalRequest_Answer); ok {
		return x.Answer
	}
	return nil
}

func (x *SignalRequest) GetTrickle() []byte {
	if x, ok := x.GetPayload().(*SignalRequest_Trickle); ok {
		return x.Trickle
	}
	return nil
}

type isSignalRequest_Payload interface {
	isSignalRequest_Payload()
}

type SignalRequest_Join struct {
	Join []byte `protobuf:"bytes,2,opt,name=join,proto3,oneof"`
}

type SignalRequest_Offer struct {
	Offer []byte `protobuf:"bytes,3,opt,name=offer,proto3,oneof"`
}

type SignalRequest_Answer struct {
	Answer []byte `protobuf:"bytes,4,opt,name=answer,proto3,oneof"`
}

type SignalRequest_Trickle struct {
	Trickle []byte `protobuf:"bytes,5,opt,name=trickle,proto3,oneof"`
}

func (*SignalRequest_Join) isSignalRequest_Payload() {}

func (*SignalRequest_Offer) isSignalRequest_Payload() {}

func (*SignalRequest_Answer) isSignalRequest_Payload() {}

func (*SignalRequest_Trickle) isSignalRequest_Payload() {}

type SignalReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Types that are assignable to Payload:
	//	*SignalReply_Offer
	//	*SignalReply_Trickle
	//	*SignalReply_Join
	//	*SignalReply_Answer
	//	*SignalReply_Description
	//	*SignalReply_Error
	Payload isSignalReply_Payload `protobuf_oneof:"payload"`
}

func (x *SignalReply) Reset() {
	*x = SignalReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sfu_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignalReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignalReply) ProtoMessage() {}

func (x *SignalReply) ProtoReflect() protoreflect.Message {
	mi := &file_sfu_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignalReply.ProtoReflect.Descriptor instead.
func (*SignalReply) Descriptor() ([]byte, []int) {
	return file_sfu_proto_rawDescGZIP(), []int{1}
}

func (x *SignalReply) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (m *SignalReply) GetPayload() isSignalReply_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (x *SignalReply) GetOffer() []byte {
	if x, ok := x.GetPayload().(*SignalReply_Offer); ok {
		return x.Offer
	}
	return nil
}

func (x *SignalReply) GetTrickle() []byte {
	if x, ok := x.GetPayload().(*SignalReply_Trickle); ok {
		return x.Trickle
	}
	return nil
}

func (x *SignalReply) GetJoin() []byte {
	if x, ok := x.GetPayload().(*SignalReply_Join); ok {
		return x.Join
	}
	return nil
}

func (x *SignalReply) GetAnswer() []byte {
	if x, ok := x.GetPayload().(*SignalReply_Answer); ok {
		return x.Answer
	}
	return nil
}

func (x *SignalReply) GetDescription() []byte {
	if x, ok := x.GetPayload().(*SignalReply_Description); ok {
		return x.Description
	}
	return nil
}

func (x *SignalReply) GetError() string {
	if x, ok := x.GetPayload().(*SignalReply_Error); ok {
		return x.Error
	}
	return ""
}

type isSignalReply_Payload interface {
	isSignalReply_Payload()
}

type SignalReply_Offer struct {
	Offer []byte `protobuf:"bytes,2,opt,name=offer,proto3,oneof"`
}

type SignalReply_Trickle struct {
	Trickle []byte `protobuf:"bytes,3,opt,name=trickle,proto3,oneof"`
}

type SignalReply_Join struct {
	Join []byte `protobuf:"bytes,4,opt,name=join,proto3,oneof"`
}

type SignalReply_Answer struct {
	Answer []byte `protobuf:"bytes,5,opt,name=answer,proto3,oneof"`
}

type SignalReply_Description struct {
	Description []byte `protobuf:"bytes,6,opt,name=description,proto3,oneof"`
}

type SignalReply_Error struct {
	Error string `protobuf:"bytes,7,opt,name=error,proto3,oneof"`
}

func (*SignalReply_Offer) isSignalReply_Payload() {}

func (*SignalReply_Trickle) isSignalReply_Payload() {}

func (*SignalReply_Join) isSignalReply_Payload() {}

func (*SignalReply_Answer) isSignalReply_Payload() {}

func (*SignalReply_Description) isSignalReply_Payload() {}

func (*SignalReply_Error) isSignalReply_Payload() {}

var File_sfu_proto protoreflect.FileDescriptor

var file_sfu_proto_rawDesc = []byte{
	0x0a, 0x09, 0x73, 0x66, 0x75, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8e, 0x01, 0x0a, 0x0d,
	0x53, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x14, 0x0a,
	0x04, 0x6a, 0x6f, 0x69, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x04, 0x6a,
	0x6f, 0x69, 0x6e, 0x12, 0x16, 0x0a, 0x05, 0x6f, 0x66, 0x66, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0c, 0x48, 0x00, 0x52, 0x05, 0x6f, 0x66, 0x66, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x06, 0x61,
	0x6e, 0x73, 0x77, 0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x06, 0x61,
	0x6e, 0x73, 0x77, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x07, 0x74, 0x72, 0x69, 0x63, 0x6b, 0x6c, 0x65,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x07, 0x74, 0x72, 0x69, 0x63, 0x6b, 0x6c,
	0x65, 0x42, 0x09, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x22, 0xc8, 0x01, 0x0a,
	0x0b, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x05,
	0x6f, 0x66, 0x66, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x05, 0x6f,
	0x66, 0x66, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x07, 0x74, 0x72, 0x69, 0x63, 0x6b, 0x6c, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x07, 0x74, 0x72, 0x69, 0x63, 0x6b, 0x6c, 0x65,
	0x12, 0x14, 0x0a, 0x04, 0x6a, 0x6f, 0x69, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00,
	0x52, 0x04, 0x6a, 0x6f, 0x69, 0x6e, 0x12, 0x18, 0x0a, 0x06, 0x61, 0x6e, 0x73, 0x77, 0x65, 0x72,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x06, 0x61, 0x6e, 0x73, 0x77, 0x65, 0x72,
	0x12, 0x22, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x42, 0x09, 0x0a, 0x07,
	0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x32, 0x33, 0x0a, 0x03, 0x53, 0x46, 0x55, 0x12, 0x2c,
	0x0a, 0x06, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x12, 0x0e, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x61,
	0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0c, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x61,
	0x6c, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_sfu_proto_rawDescOnce sync.Once
	file_sfu_proto_rawDescData = file_sfu_proto_rawDesc
)

func file_sfu_proto_rawDescGZIP() []byte {
	file_sfu_proto_rawDescOnce.Do(func() {
		file_sfu_proto_rawDescData = protoimpl.X.CompressGZIP(file_sfu_proto_rawDescData)
	})
	return file_sfu_proto_rawDescData
}

var file_sfu_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_sfu_proto_goTypes = []interface{}{
	(*SignalRequest)(nil), // 0: SignalRequest
	(*SignalReply)(nil),   // 1: SignalReply
}
var file_sfu_proto_depIdxs = []int32{
	0, // 0: SFU.Signal:input_type -> SignalRequest
	1, // 1: SFU.Signal:output_type -> SignalReply
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_sfu_proto_init() }
func file_sfu_proto_init() {
	if File_sfu_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_sfu_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignalRequest); i {
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
		file_sfu_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignalReply); i {
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
	file_sfu_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*SignalRequest_Join)(nil),
		(*SignalRequest_Offer)(nil),
		(*SignalRequest_Answer)(nil),
		(*SignalRequest_Trickle)(nil),
	}
	file_sfu_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*SignalReply_Offer)(nil),
		(*SignalReply_Trickle)(nil),
		(*SignalReply_Join)(nil),
		(*SignalReply_Answer)(nil),
		(*SignalReply_Description)(nil),
		(*SignalReply_Error)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_sfu_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_sfu_proto_goTypes,
		DependencyIndexes: file_sfu_proto_depIdxs,
		MessageInfos:      file_sfu_proto_msgTypes,
	}.Build()
	File_sfu_proto = out.File
	file_sfu_proto_rawDesc = nil
	file_sfu_proto_goTypes = nil
	file_sfu_proto_depIdxs = nil
}
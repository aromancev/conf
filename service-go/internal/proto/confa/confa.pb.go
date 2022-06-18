// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.2
// source: confa.proto

package confa

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

type UpdateProfile struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserId []byte                `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Avatar *UpdateProfile_Source `protobuf:"bytes,2,opt,name=avatar,proto3" json:"avatar,omitempty"`
}

func (x *UpdateProfile) Reset() {
	*x = UpdateProfile{}
	if protoimpl.UnsafeEnabled {
		mi := &file_confa_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateProfile) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateProfile) ProtoMessage() {}

func (x *UpdateProfile) ProtoReflect() protoreflect.Message {
	mi := &file_confa_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateProfile.ProtoReflect.Descriptor instead.
func (*UpdateProfile) Descriptor() ([]byte, []int) {
	return file_confa_proto_rawDescGZIP(), []int{0}
}

func (x *UpdateProfile) GetUserId() []byte {
	if x != nil {
		return x.UserId
	}
	return nil
}

func (x *UpdateProfile) GetAvatar() *UpdateProfile_Source {
	if x != nil {
		return x.Avatar
	}
	return nil
}

type StartRecording struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TalkId []byte `protobuf:"bytes,1,opt,name=talk_id,json=talkId,proto3" json:"talk_id,omitempty"`
	RoomId []byte `protobuf:"bytes,2,opt,name=room_id,json=roomId,proto3" json:"room_id,omitempty"`
}

func (x *StartRecording) Reset() {
	*x = StartRecording{}
	if protoimpl.UnsafeEnabled {
		mi := &file_confa_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StartRecording) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartRecording) ProtoMessage() {}

func (x *StartRecording) ProtoReflect() protoreflect.Message {
	mi := &file_confa_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartRecording.ProtoReflect.Descriptor instead.
func (*StartRecording) Descriptor() ([]byte, []int) {
	return file_confa_proto_rawDescGZIP(), []int{1}
}

func (x *StartRecording) GetTalkId() []byte {
	if x != nil {
		return x.TalkId
	}
	return nil
}

func (x *StartRecording) GetRoomId() []byte {
	if x != nil {
		return x.RoomId
	}
	return nil
}

type StopRecording struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TalkId []byte `protobuf:"bytes,1,opt,name=talk_id,json=talkId,proto3" json:"talk_id,omitempty"`
	RoomId []byte `protobuf:"bytes,2,opt,name=room_id,json=roomId,proto3" json:"room_id,omitempty"`
}

func (x *StopRecording) Reset() {
	*x = StopRecording{}
	if protoimpl.UnsafeEnabled {
		mi := &file_confa_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StopRecording) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StopRecording) ProtoMessage() {}

func (x *StopRecording) ProtoReflect() protoreflect.Message {
	mi := &file_confa_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StopRecording.ProtoReflect.Descriptor instead.
func (*StopRecording) Descriptor() ([]byte, []int) {
	return file_confa_proto_rawDescGZIP(), []int{2}
}

func (x *StopRecording) GetTalkId() []byte {
	if x != nil {
		return x.TalkId
	}
	return nil
}

func (x *StopRecording) GetRoomId() []byte {
	if x != nil {
		return x.RoomId
	}
	return nil
}

type UpdateProfile_Source struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Storage   *UpdateProfile_Source_Storage   `protobuf:"bytes,1,opt,name=storage,proto3" json:"storage,omitempty"`
	PublicUrl *UpdateProfile_Source_PublicURL `protobuf:"bytes,2,opt,name=public_url,json=publicUrl,proto3" json:"public_url,omitempty"`
}

func (x *UpdateProfile_Source) Reset() {
	*x = UpdateProfile_Source{}
	if protoimpl.UnsafeEnabled {
		mi := &file_confa_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateProfile_Source) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateProfile_Source) ProtoMessage() {}

func (x *UpdateProfile_Source) ProtoReflect() protoreflect.Message {
	mi := &file_confa_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateProfile_Source.ProtoReflect.Descriptor instead.
func (*UpdateProfile_Source) Descriptor() ([]byte, []int) {
	return file_confa_proto_rawDescGZIP(), []int{0, 0}
}

func (x *UpdateProfile_Source) GetStorage() *UpdateProfile_Source_Storage {
	if x != nil {
		return x.Storage
	}
	return nil
}

func (x *UpdateProfile_Source) GetPublicUrl() *UpdateProfile_Source_PublicURL {
	if x != nil {
		return x.PublicUrl
	}
	return nil
}

type UpdateProfile_Source_Storage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Bucket string `protobuf:"bytes,1,opt,name=bucket,proto3" json:"bucket,omitempty"`
	Path   string `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
}

func (x *UpdateProfile_Source_Storage) Reset() {
	*x = UpdateProfile_Source_Storage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_confa_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateProfile_Source_Storage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateProfile_Source_Storage) ProtoMessage() {}

func (x *UpdateProfile_Source_Storage) ProtoReflect() protoreflect.Message {
	mi := &file_confa_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateProfile_Source_Storage.ProtoReflect.Descriptor instead.
func (*UpdateProfile_Source_Storage) Descriptor() ([]byte, []int) {
	return file_confa_proto_rawDescGZIP(), []int{0, 0, 0}
}

func (x *UpdateProfile_Source_Storage) GetBucket() string {
	if x != nil {
		return x.Bucket
	}
	return ""
}

func (x *UpdateProfile_Source_Storage) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

type UpdateProfile_Source_PublicURL struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Url string `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
}

func (x *UpdateProfile_Source_PublicURL) Reset() {
	*x = UpdateProfile_Source_PublicURL{}
	if protoimpl.UnsafeEnabled {
		mi := &file_confa_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateProfile_Source_PublicURL) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateProfile_Source_PublicURL) ProtoMessage() {}

func (x *UpdateProfile_Source_PublicURL) ProtoReflect() protoreflect.Message {
	mi := &file_confa_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateProfile_Source_PublicURL.ProtoReflect.Descriptor instead.
func (*UpdateProfile_Source_PublicURL) Descriptor() ([]byte, []int) {
	return file_confa_proto_rawDescGZIP(), []int{0, 0, 1}
}

func (x *UpdateProfile_Source_PublicURL) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

var File_confa_proto protoreflect.FileDescriptor

var file_confa_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x63, 0x6f, 0x6e, 0x66, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb1, 0x02,
	0x0a, 0x0d, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x12,
	0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x2d, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74,
	0x61, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52,
	0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x1a, 0xd7, 0x01, 0x0a, 0x06, 0x53, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x12, 0x37, 0x0a, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x66,
	0x69, 0x6c, 0x65, 0x2e, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x61,
	0x67, 0x65, 0x52, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x12, 0x3e, 0x0a, 0x0a, 0x70,
	0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1f, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x2e,
	0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x55, 0x52, 0x4c,
	0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x55, 0x72, 0x6c, 0x1a, 0x35, 0x0a, 0x07, 0x53,
	0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61,
	0x74, 0x68, 0x1a, 0x1d, 0x0a, 0x09, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x55, 0x52, 0x4c, 0x12,
	0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72,
	0x6c, 0x22, 0x42, 0x0a, 0x0e, 0x53, 0x74, 0x61, 0x72, 0x74, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64,
	0x69, 0x6e, 0x67, 0x12, 0x17, 0x0a, 0x07, 0x74, 0x61, 0x6c, 0x6b, 0x5f, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x74, 0x61, 0x6c, 0x6b, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07,
	0x72, 0x6f, 0x6f, 0x6d, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x72,
	0x6f, 0x6f, 0x6d, 0x49, 0x64, 0x22, 0x41, 0x0a, 0x0d, 0x53, 0x74, 0x6f, 0x70, 0x52, 0x65, 0x63,
	0x6f, 0x72, 0x64, 0x69, 0x6e, 0x67, 0x12, 0x17, 0x0a, 0x07, 0x74, 0x61, 0x6c, 0x6b, 0x5f, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x74, 0x61, 0x6c, 0x6b, 0x49, 0x64, 0x12,
	0x17, 0x0a, 0x07, 0x72, 0x6f, 0x6f, 0x6d, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x06, 0x72, 0x6f, 0x6f, 0x6d, 0x49, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_confa_proto_rawDescOnce sync.Once
	file_confa_proto_rawDescData = file_confa_proto_rawDesc
)

func file_confa_proto_rawDescGZIP() []byte {
	file_confa_proto_rawDescOnce.Do(func() {
		file_confa_proto_rawDescData = protoimpl.X.CompressGZIP(file_confa_proto_rawDescData)
	})
	return file_confa_proto_rawDescData
}

var file_confa_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_confa_proto_goTypes = []interface{}{
	(*UpdateProfile)(nil),                  // 0: UpdateProfile
	(*StartRecording)(nil),                 // 1: StartRecording
	(*StopRecording)(nil),                  // 2: StopRecording
	(*UpdateProfile_Source)(nil),           // 3: UpdateProfile.Source
	(*UpdateProfile_Source_Storage)(nil),   // 4: UpdateProfile.Source.Storage
	(*UpdateProfile_Source_PublicURL)(nil), // 5: UpdateProfile.Source.PublicURL
}
var file_confa_proto_depIdxs = []int32{
	3, // 0: UpdateProfile.avatar:type_name -> UpdateProfile.Source
	4, // 1: UpdateProfile.Source.storage:type_name -> UpdateProfile.Source.Storage
	5, // 2: UpdateProfile.Source.public_url:type_name -> UpdateProfile.Source.PublicURL
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_confa_proto_init() }
func file_confa_proto_init() {
	if File_confa_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_confa_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateProfile); i {
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
		file_confa_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StartRecording); i {
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
		file_confa_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StopRecording); i {
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
		file_confa_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateProfile_Source); i {
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
		file_confa_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateProfile_Source_Storage); i {
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
		file_confa_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateProfile_Source_PublicURL); i {
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
			RawDescriptor: file_confa_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_confa_proto_goTypes,
		DependencyIndexes: file_confa_proto_depIdxs,
		MessageInfos:      file_confa_proto_msgTypes,
	}.Build()
	File_confa_proto = out.File
	file_confa_proto_rawDesc = nil
	file_confa_proto_goTypes = nil
	file_confa_proto_depIdxs = nil
}

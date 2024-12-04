// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v3.19.6
// source: rso-comms/game.proto

package ___

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

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	mi := &file_rso_comms_game_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_rso_comms_game_proto_msgTypes[0]
	if x != nil {
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
	return file_rso_comms_game_proto_rawDescGZIP(), []int{0}
}

type GameID struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GameID) Reset() {
	*x = GameID{}
	mi := &file_rso_comms_game_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GameID) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GameID) ProtoMessage() {}

func (x *GameID) ProtoReflect() protoreflect.Message {
	mi := &file_rso_comms_game_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GameID.ProtoReflect.Descriptor instead.
func (*GameID) Descriptor() ([]byte, []int) {
	return file_rso_comms_game_proto_rawDescGZIP(), []int{1}
}

func (x *GameID) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GameIDList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ids []string `protobuf:"bytes,1,rep,name=ids,proto3" json:"ids,omitempty"`
}

func (x *GameIDList) Reset() {
	*x = GameIDList{}
	mi := &file_rso_comms_game_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GameIDList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GameIDList) ProtoMessage() {}

func (x *GameIDList) ProtoReflect() protoreflect.Message {
	mi := &file_rso_comms_game_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GameIDList.ProtoReflect.Descriptor instead.
func (*GameIDList) Descriptor() ([]byte, []int) {
	return file_rso_comms_game_proto_rawDescGZIP(), []int{2}
}

func (x *GameIDList) GetIds() []string {
	if x != nil {
		return x.Ids
	}
	return nil
}

var File_rso_comms_game_proto protoreflect.FileDescriptor

var file_rso_comms_game_proto_rawDesc = []byte{
	0x0a, 0x14, 0x72, 0x73, 0x6f, 0x2d, 0x63, 0x6f, 0x6d, 0x6d, 0x73, 0x2f, 0x67, 0x61, 0x6d, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22,
	0x18, 0x0a, 0x06, 0x47, 0x61, 0x6d, 0x65, 0x49, 0x44, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x1e, 0x0a, 0x0a, 0x47, 0x61, 0x6d,
	0x65, 0x49, 0x44, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x69, 0x64, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x03, 0x69, 0x64, 0x73, 0x32, 0x59, 0x0a, 0x0b, 0x47, 0x61, 0x6d,
	0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x1f, 0x0a, 0x0a, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x47, 0x61, 0x6d, 0x65, 0x12, 0x06, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x07,
	0x2e, 0x47, 0x61, 0x6d, 0x65, 0x49, 0x44, 0x22, 0x00, 0x12, 0x29, 0x0a, 0x10, 0x4c, 0x69, 0x73,
	0x74, 0x52, 0x75, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x47, 0x61, 0x6d, 0x65, 0x73, 0x12, 0x06, 0x2e,
	0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x0b, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x49, 0x44, 0x4c, 0x69,
	0x73, 0x74, 0x22, 0x00, 0x42, 0x05, 0x5a, 0x03, 0x2e, 0x2e, 0x2f, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_rso_comms_game_proto_rawDescOnce sync.Once
	file_rso_comms_game_proto_rawDescData = file_rso_comms_game_proto_rawDesc
)

func file_rso_comms_game_proto_rawDescGZIP() []byte {
	file_rso_comms_game_proto_rawDescOnce.Do(func() {
		file_rso_comms_game_proto_rawDescData = protoimpl.X.CompressGZIP(file_rso_comms_game_proto_rawDescData)
	})
	return file_rso_comms_game_proto_rawDescData
}

var file_rso_comms_game_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_rso_comms_game_proto_goTypes = []any{
	(*Empty)(nil),      // 0: Empty
	(*GameID)(nil),     // 1: GameID
	(*GameIDList)(nil), // 2: GameIDList
}
var file_rso_comms_game_proto_depIdxs = []int32{
	0, // 0: GameService.CreateGame:input_type -> Empty
	0, // 1: GameService.ListRunningGames:input_type -> Empty
	1, // 2: GameService.CreateGame:output_type -> GameID
	2, // 3: GameService.ListRunningGames:output_type -> GameIDList
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_rso_comms_game_proto_init() }
func file_rso_comms_game_proto_init() {
	if File_rso_comms_game_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_rso_comms_game_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_rso_comms_game_proto_goTypes,
		DependencyIndexes: file_rso_comms_game_proto_depIdxs,
		MessageInfos:      file_rso_comms_game_proto_msgTypes,
	}.Build()
	File_rso_comms_game_proto = out.File
	file_rso_comms_game_proto_rawDesc = nil
	file_rso_comms_game_proto_goTypes = nil
	file_rso_comms_game_proto_depIdxs = nil
}

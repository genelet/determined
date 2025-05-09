// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v3.21.12
// source: proto/det.proto

package det

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

type Struct struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClassName string            `protobuf:"bytes,1,opt,name=ClassName,proto3" json:"ClassName,omitempty"`
	Fields    map[string]*Value `protobuf:"bytes,2,rep,name=fields,proto3" json:"fields,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Struct) Reset() {
	*x = Struct{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_det_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Struct) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Struct) ProtoMessage() {}

func (x *Struct) ProtoReflect() protoreflect.Message {
	mi := &file_proto_det_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Struct.ProtoReflect.Descriptor instead.
func (*Struct) Descriptor() ([]byte, []int) {
	return file_proto_det_proto_rawDescGZIP(), []int{0}
}

func (x *Struct) GetClassName() string {
	if x != nil {
		return x.ClassName
	}
	return ""
}

func (x *Struct) GetFields() map[string]*Value {
	if x != nil {
		return x.Fields
	}
	return nil
}

type Value struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The kind of value.
	//
	// Types that are assignable to Kind:
	//
	//	*Value_SingleStruct
	//	*Value_ListStruct
	//	*Value_MapStruct
	Kind isValue_Kind `protobuf_oneof:"kind"`
}

func (x *Value) Reset() {
	*x = Value{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_det_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Value) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Value) ProtoMessage() {}

func (x *Value) ProtoReflect() protoreflect.Message {
	mi := &file_proto_det_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Value.ProtoReflect.Descriptor instead.
func (*Value) Descriptor() ([]byte, []int) {
	return file_proto_det_proto_rawDescGZIP(), []int{1}
}

func (m *Value) GetKind() isValue_Kind {
	if m != nil {
		return m.Kind
	}
	return nil
}

func (x *Value) GetSingleStruct() *Struct {
	if x, ok := x.GetKind().(*Value_SingleStruct); ok {
		return x.SingleStruct
	}
	return nil
}

func (x *Value) GetListStruct() *ListStruct {
	if x, ok := x.GetKind().(*Value_ListStruct); ok {
		return x.ListStruct
	}
	return nil
}

func (x *Value) GetMapStruct() *MapStruct {
	if x, ok := x.GetKind().(*Value_MapStruct); ok {
		return x.MapStruct
	}
	return nil
}

type isValue_Kind interface {
	isValue_Kind()
}

type Value_SingleStruct struct {
	SingleStruct *Struct `protobuf:"bytes,1,opt,name=single_struct,json=singleStruct,proto3,oneof"`
}

type Value_ListStruct struct {
	ListStruct *ListStruct `protobuf:"bytes,2,opt,name=list_struct,json=listStruct,proto3,oneof"`
}

type Value_MapStruct struct {
	MapStruct *MapStruct `protobuf:"bytes,3,opt,name=map_struct,json=mapStruct,proto3,oneof"`
}

func (*Value_SingleStruct) isValue_Kind() {}

func (*Value_ListStruct) isValue_Kind() {}

func (*Value_MapStruct) isValue_Kind() {}

type ListStruct struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ListFields []*Struct `protobuf:"bytes,1,rep,name=list_fields,json=listFields,proto3" json:"list_fields,omitempty"`
}

func (x *ListStruct) Reset() {
	*x = ListStruct{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_det_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListStruct) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListStruct) ProtoMessage() {}

func (x *ListStruct) ProtoReflect() protoreflect.Message {
	mi := &file_proto_det_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListStruct.ProtoReflect.Descriptor instead.
func (*ListStruct) Descriptor() ([]byte, []int) {
	return file_proto_det_proto_rawDescGZIP(), []int{2}
}

func (x *ListStruct) GetListFields() []*Struct {
	if x != nil {
		return x.ListFields
	}
	return nil
}

type MapStruct struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MapFields map[string]*Struct `protobuf:"bytes,1,rep,name=map_fields,json=mapFields,proto3" json:"map_fields,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *MapStruct) Reset() {
	*x = MapStruct{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_det_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MapStruct) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MapStruct) ProtoMessage() {}

func (x *MapStruct) ProtoReflect() protoreflect.Message {
	mi := &file_proto_det_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MapStruct.ProtoReflect.Descriptor instead.
func (*MapStruct) Descriptor() ([]byte, []int) {
	return file_proto_det_proto_rawDescGZIP(), []int{3}
}

func (x *MapStruct) GetMapFields() map[string]*Struct {
	if x != nil {
		return x.MapFields
	}
	return nil
}

var File_proto_det_proto protoreflect.FileDescriptor

var file_proto_det_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x64, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x03, 0x64, 0x65, 0x74, 0x22, 0x9e, 0x01, 0x0a, 0x06, 0x53, 0x74, 0x72, 0x75, 0x63,
	0x74, 0x12, 0x1c, 0x0a, 0x09, 0x43, 0x6c, 0x61, 0x73, 0x73, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x43, 0x6c, 0x61, 0x73, 0x73, 0x4e, 0x61, 0x6d, 0x65, 0x12,
	0x2f, 0x0a, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x17, 0x2e, 0x64, 0x65, 0x74, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x46, 0x69, 0x65,
	0x6c, 0x64, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73,
	0x1a, 0x45, 0x0a, 0x0b, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x20, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0a, 0x2e, 0x64, 0x65, 0x74, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xa8, 0x01, 0x0a, 0x05, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x12, 0x32, 0x0a, 0x0d, 0x73, 0x69, 0x6e, 0x67, 0x6c, 0x65, 0x5f, 0x73, 0x74, 0x72, 0x75,
	0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x64, 0x65, 0x74, 0x2e, 0x53,
	0x74, 0x72, 0x75, 0x63, 0x74, 0x48, 0x00, 0x52, 0x0c, 0x73, 0x69, 0x6e, 0x67, 0x6c, 0x65, 0x53,
	0x74, 0x72, 0x75, 0x63, 0x74, 0x12, 0x32, 0x0a, 0x0b, 0x6c, 0x69, 0x73, 0x74, 0x5f, 0x73, 0x74,
	0x72, 0x75, 0x63, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x64, 0x65, 0x74,
	0x2e, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x48, 0x00, 0x52, 0x0a, 0x6c,
	0x69, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x12, 0x2f, 0x0a, 0x0a, 0x6d, 0x61, 0x70,
	0x5f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e,
	0x64, 0x65, 0x74, 0x2e, 0x4d, 0x61, 0x70, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x48, 0x00, 0x52,
	0x09, 0x6d, 0x61, 0x70, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x42, 0x06, 0x0a, 0x04, 0x6b, 0x69,
	0x6e, 0x64, 0x22, 0x3a, 0x0a, 0x0a, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74,
	0x12, 0x2c, 0x0a, 0x0b, 0x6c, 0x69, 0x73, 0x74, 0x5f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x64, 0x65, 0x74, 0x2e, 0x53, 0x74, 0x72, 0x75,
	0x63, 0x74, 0x52, 0x0a, 0x6c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x22, 0x94,
	0x01, 0x0a, 0x09, 0x4d, 0x61, 0x70, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x12, 0x3c, 0x0a, 0x0a,
	0x6d, 0x61, 0x70, 0x5f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x1d, 0x2e, 0x64, 0x65, 0x74, 0x2e, 0x4d, 0x61, 0x70, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74,
	0x2e, 0x4d, 0x61, 0x70, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x09, 0x6d, 0x61, 0x70, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x1a, 0x49, 0x0a, 0x0e, 0x4d, 0x61,
	0x70, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x21,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b, 0x2e,
	0x64, 0x65, 0x74, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x07, 0x5a, 0x05, 0x2e, 0x2f, 0x64, 0x65, 0x74, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_det_proto_rawDescOnce sync.Once
	file_proto_det_proto_rawDescData = file_proto_det_proto_rawDesc
)

func file_proto_det_proto_rawDescGZIP() []byte {
	file_proto_det_proto_rawDescOnce.Do(func() {
		file_proto_det_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_det_proto_rawDescData)
	})
	return file_proto_det_proto_rawDescData
}

var file_proto_det_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_proto_det_proto_goTypes = []interface{}{
	(*Struct)(nil),     // 0: det.Struct
	(*Value)(nil),      // 1: det.Value
	(*ListStruct)(nil), // 2: det.ListStruct
	(*MapStruct)(nil),  // 3: det.MapStruct
	nil,                // 4: det.Struct.FieldsEntry
	nil,                // 5: det.MapStruct.MapFieldsEntry
}
var file_proto_det_proto_depIdxs = []int32{
	4, // 0: det.Struct.fields:type_name -> det.Struct.FieldsEntry
	0, // 1: det.Value.single_struct:type_name -> det.Struct
	2, // 2: det.Value.list_struct:type_name -> det.ListStruct
	3, // 3: det.Value.map_struct:type_name -> det.MapStruct
	0, // 4: det.ListStruct.list_fields:type_name -> det.Struct
	5, // 5: det.MapStruct.map_fields:type_name -> det.MapStruct.MapFieldsEntry
	1, // 6: det.Struct.FieldsEntry.value:type_name -> det.Value
	0, // 7: det.MapStruct.MapFieldsEntry.value:type_name -> det.Struct
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_proto_det_proto_init() }
func file_proto_det_proto_init() {
	if File_proto_det_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_det_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Struct); i {
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
		file_proto_det_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Value); i {
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
		file_proto_det_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListStruct); i {
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
		file_proto_det_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MapStruct); i {
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
	file_proto_det_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*Value_SingleStruct)(nil),
		(*Value_ListStruct)(nil),
		(*Value_MapStruct)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_det_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_det_proto_goTypes,
		DependencyIndexes: file_proto_det_proto_depIdxs,
		MessageInfos:      file_proto_det_proto_msgTypes,
	}.Build()
	File_proto_det_proto = out.File
	file_proto_det_proto_rawDesc = nil
	file_proto_det_proto_goTypes = nil
	file_proto_det_proto_depIdxs = nil
}

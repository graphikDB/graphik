// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api/graphik.proto

package apipb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	any "github.com/golang/protobuf/ptypes/any"
	_struct "github.com/golang/protobuf/ptypes/struct"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Op int32

const (
	Op_CREATE_NODE Op = 0
	Op_PATCH_NODE  Op = 1
	Op_DELETE_NODE Op = 2
	Op_CREATE_EDGE Op = 3
	Op_PATCH_EDGE  Op = 4
	Op_DELETE_EDGE Op = 5
)

var Op_name = map[int32]string{
	0: "CREATE_NODE",
	1: "PATCH_NODE",
	2: "DELETE_NODE",
	3: "CREATE_EDGE",
	4: "PATCH_EDGE",
	5: "DELETE_EDGE",
}

var Op_value = map[string]int32{
	"CREATE_NODE": 0,
	"PATCH_NODE":  1,
	"DELETE_NODE": 2,
	"CREATE_EDGE": 3,
	"PATCH_EDGE":  4,
	"DELETE_EDGE": 5,
}

func (x Op) String() string {
	return proto.EnumName(Op_name, int32(x))
}

func (Op) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_063490d3009de3e6, []int{0}
}

type Keyword int32

const (
	Keyword_ANY     Keyword = 0
	Keyword_DEFAULT Keyword = 1
)

var Keyword_name = map[int32]string{
	0: "ANY",
	1: "DEFAULT",
}

var Keyword_value = map[string]int32{
	"ANY":     0,
	"DEFAULT": 1,
}

func (x Keyword) String() string {
	return proto.EnumName(Keyword_name, int32(x))
}

func (Keyword) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_063490d3009de3e6, []int{1}
}

type Command struct {
	Op                   Op                   `protobuf:"varint,1,opt,name=op,proto3,enum=api.Op" json:"op,omitempty"`
	Val                  *any.Any             `protobuf:"bytes,2,opt,name=val,proto3" json:"val,omitempty"`
	Timestamp            *timestamp.Timestamp `protobuf:"bytes,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *Command) Reset()         { *m = Command{} }
func (m *Command) String() string { return proto.CompactTextString(m) }
func (*Command) ProtoMessage()    {}
func (*Command) Descriptor() ([]byte, []int) {
	return fileDescriptor_063490d3009de3e6, []int{0}
}

func (m *Command) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Command.Unmarshal(m, b)
}
func (m *Command) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Command.Marshal(b, m, deterministic)
}
func (m *Command) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Command.Merge(m, src)
}
func (m *Command) XXX_Size() int {
	return xxx_messageInfo_Command.Size(m)
}
func (m *Command) XXX_DiscardUnknown() {
	xxx_messageInfo_Command.DiscardUnknown(m)
}

var xxx_messageInfo_Command proto.InternalMessageInfo

func (m *Command) GetOp() Op {
	if m != nil {
		return m.Op
	}
	return Op_CREATE_NODE
}

func (m *Command) GetVal() *any.Any {
	if m != nil {
		return m.Val
	}
	return nil
}

func (m *Command) GetTimestamp() *timestamp.Timestamp {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

type Counter struct {
	Count                int64    `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Counter) Reset()         { *m = Counter{} }
func (m *Counter) String() string { return proto.CompactTextString(m) }
func (*Counter) ProtoMessage()    {}
func (*Counter) Descriptor() ([]byte, []int) {
	return fileDescriptor_063490d3009de3e6, []int{1}
}

func (m *Counter) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Counter.Unmarshal(m, b)
}
func (m *Counter) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Counter.Marshal(b, m, deterministic)
}
func (m *Counter) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Counter.Merge(m, src)
}
func (m *Counter) XXX_Size() int {
	return xxx_messageInfo_Counter.Size(m)
}
func (m *Counter) XXX_DiscardUnknown() {
	xxx_messageInfo_Counter.DiscardUnknown(m)
}

var xxx_messageInfo_Counter proto.InternalMessageInfo

func (m *Counter) GetCount() int64 {
	if m != nil {
		return m.Count
	}
	return 0
}

type Path struct {
	Type                 string   `protobuf:"bytes,1,opt,name=Type,proto3" json:"Type,omitempty"`
	ID                   string   `protobuf:"bytes,2,opt,name=ID,proto3" json:"ID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Path) Reset()         { *m = Path{} }
func (m *Path) String() string { return proto.CompactTextString(m) }
func (*Path) ProtoMessage()    {}
func (*Path) Descriptor() ([]byte, []int) {
	return fileDescriptor_063490d3009de3e6, []int{2}
}

func (m *Path) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Path.Unmarshal(m, b)
}
func (m *Path) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Path.Marshal(b, m, deterministic)
}
func (m *Path) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Path.Merge(m, src)
}
func (m *Path) XXX_Size() int {
	return xxx_messageInfo_Path.Size(m)
}
func (m *Path) XXX_DiscardUnknown() {
	xxx_messageInfo_Path.DiscardUnknown(m)
}

var xxx_messageInfo_Path proto.InternalMessageInfo

func (m *Path) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *Path) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

type Node struct {
	Path                 *Path                `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	Attributes           *_struct.Struct      `protobuf:"bytes,2,opt,name=attributes,proto3" json:"attributes,omitempty"`
	CreatedAt            *timestamp.Timestamp `protobuf:"bytes,3,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt            *timestamp.Timestamp `protobuf:"bytes,4,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *Node) Reset()         { *m = Node{} }
func (m *Node) String() string { return proto.CompactTextString(m) }
func (*Node) ProtoMessage()    {}
func (*Node) Descriptor() ([]byte, []int) {
	return fileDescriptor_063490d3009de3e6, []int{3}
}

func (m *Node) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Node.Unmarshal(m, b)
}
func (m *Node) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Node.Marshal(b, m, deterministic)
}
func (m *Node) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Node.Merge(m, src)
}
func (m *Node) XXX_Size() int {
	return xxx_messageInfo_Node.Size(m)
}
func (m *Node) XXX_DiscardUnknown() {
	xxx_messageInfo_Node.DiscardUnknown(m)
}

var xxx_messageInfo_Node proto.InternalMessageInfo

func (m *Node) GetPath() *Path {
	if m != nil {
		return m.Path
	}
	return nil
}

func (m *Node) GetAttributes() *_struct.Struct {
	if m != nil {
		return m.Attributes
	}
	return nil
}

func (m *Node) GetCreatedAt() *timestamp.Timestamp {
	if m != nil {
		return m.CreatedAt
	}
	return nil
}

func (m *Node) GetUpdatedAt() *timestamp.Timestamp {
	if m != nil {
		return m.UpdatedAt
	}
	return nil
}

type Message struct {
	Channel              string          `protobuf:"bytes,1,opt,name=channel,proto3" json:"channel,omitempty"`
	Type                 string          `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	Attributes           *_struct.Struct `protobuf:"bytes,3,opt,name=attributes,proto3" json:"attributes,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *Message) Reset()         { *m = Message{} }
func (m *Message) String() string { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()    {}
func (*Message) Descriptor() ([]byte, []int) {
	return fileDescriptor_063490d3009de3e6, []int{4}
}

func (m *Message) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Message.Unmarshal(m, b)
}
func (m *Message) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Message.Marshal(b, m, deterministic)
}
func (m *Message) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Message.Merge(m, src)
}
func (m *Message) XXX_Size() int {
	return xxx_messageInfo_Message.Size(m)
}
func (m *Message) XXX_DiscardUnknown() {
	xxx_messageInfo_Message.DiscardUnknown(m)
}

var xxx_messageInfo_Message proto.InternalMessageInfo

func (m *Message) GetChannel() string {
	if m != nil {
		return m.Channel
	}
	return ""
}

func (m *Message) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *Message) GetAttributes() *_struct.Struct {
	if m != nil {
		return m.Attributes
	}
	return nil
}

type Filter struct {
	Type                 string   `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Expressions          []string `protobuf:"bytes,2,rep,name=expressions,proto3" json:"expressions,omitempty"`
	Limit                int32    `protobuf:"varint,3,opt,name=limit,proto3" json:"limit,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Filter) Reset()         { *m = Filter{} }
func (m *Filter) String() string { return proto.CompactTextString(m) }
func (*Filter) ProtoMessage()    {}
func (*Filter) Descriptor() ([]byte, []int) {
	return fileDescriptor_063490d3009de3e6, []int{5}
}

func (m *Filter) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Filter.Unmarshal(m, b)
}
func (m *Filter) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Filter.Marshal(b, m, deterministic)
}
func (m *Filter) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Filter.Merge(m, src)
}
func (m *Filter) XXX_Size() int {
	return xxx_messageInfo_Filter.Size(m)
}
func (m *Filter) XXX_DiscardUnknown() {
	xxx_messageInfo_Filter.DiscardUnknown(m)
}

var xxx_messageInfo_Filter proto.InternalMessageInfo

func (m *Filter) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *Filter) GetExpressions() []string {
	if m != nil {
		return m.Expressions
	}
	return nil
}

func (m *Filter) GetLimit() int32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

type Export struct {
	Nodes                []*Node  `protobuf:"bytes,1,rep,name=nodes,proto3" json:"nodes,omitempty"`
	Edges                []*Edge  `protobuf:"bytes,2,rep,name=edges,proto3" json:"edges,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Export) Reset()         { *m = Export{} }
func (m *Export) String() string { return proto.CompactTextString(m) }
func (*Export) ProtoMessage()    {}
func (*Export) Descriptor() ([]byte, []int) {
	return fileDescriptor_063490d3009de3e6, []int{6}
}

func (m *Export) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Export.Unmarshal(m, b)
}
func (m *Export) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Export.Marshal(b, m, deterministic)
}
func (m *Export) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Export.Merge(m, src)
}
func (m *Export) XXX_Size() int {
	return xxx_messageInfo_Export.Size(m)
}
func (m *Export) XXX_DiscardUnknown() {
	xxx_messageInfo_Export.DiscardUnknown(m)
}

var xxx_messageInfo_Export proto.InternalMessageInfo

func (m *Export) GetNodes() []*Node {
	if m != nil {
		return m.Nodes
	}
	return nil
}

func (m *Export) GetEdges() []*Edge {
	if m != nil {
		return m.Edges
	}
	return nil
}

type Patch struct {
	Path                 *Path           `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	Patch                *_struct.Struct `protobuf:"bytes,2,opt,name=patch,proto3" json:"patch,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *Patch) Reset()         { *m = Patch{} }
func (m *Patch) String() string { return proto.CompactTextString(m) }
func (*Patch) ProtoMessage()    {}
func (*Patch) Descriptor() ([]byte, []int) {
	return fileDescriptor_063490d3009de3e6, []int{7}
}

func (m *Patch) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Patch.Unmarshal(m, b)
}
func (m *Patch) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Patch.Marshal(b, m, deterministic)
}
func (m *Patch) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Patch.Merge(m, src)
}
func (m *Patch) XXX_Size() int {
	return xxx_messageInfo_Patch.Size(m)
}
func (m *Patch) XXX_DiscardUnknown() {
	xxx_messageInfo_Patch.DiscardUnknown(m)
}

var xxx_messageInfo_Patch proto.InternalMessageInfo

func (m *Patch) GetPath() *Path {
	if m != nil {
		return m.Path
	}
	return nil
}

func (m *Patch) GetPatch() *_struct.Struct {
	if m != nil {
		return m.Patch
	}
	return nil
}

type Edge struct {
	Path                 *Path                `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	Mutual               bool                 `protobuf:"varint,2,opt,name=mutual,proto3" json:"mutual,omitempty"`
	Attributes           *_struct.Struct      `protobuf:"bytes,3,opt,name=attributes,proto3" json:"attributes,omitempty"`
	From                 *Path                `protobuf:"bytes,4,opt,name=from,proto3" json:"from,omitempty"`
	To                   *Path                `protobuf:"bytes,5,opt,name=to,proto3" json:"to,omitempty"`
	CreatedAt            *timestamp.Timestamp `protobuf:"bytes,6,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt            *timestamp.Timestamp `protobuf:"bytes,7,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *Edge) Reset()         { *m = Edge{} }
func (m *Edge) String() string { return proto.CompactTextString(m) }
func (*Edge) ProtoMessage()    {}
func (*Edge) Descriptor() ([]byte, []int) {
	return fileDescriptor_063490d3009de3e6, []int{8}
}

func (m *Edge) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Edge.Unmarshal(m, b)
}
func (m *Edge) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Edge.Marshal(b, m, deterministic)
}
func (m *Edge) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Edge.Merge(m, src)
}
func (m *Edge) XXX_Size() int {
	return xxx_messageInfo_Edge.Size(m)
}
func (m *Edge) XXX_DiscardUnknown() {
	xxx_messageInfo_Edge.DiscardUnknown(m)
}

var xxx_messageInfo_Edge proto.InternalMessageInfo

func (m *Edge) GetPath() *Path {
	if m != nil {
		return m.Path
	}
	return nil
}

func (m *Edge) GetMutual() bool {
	if m != nil {
		return m.Mutual
	}
	return false
}

func (m *Edge) GetAttributes() *_struct.Struct {
	if m != nil {
		return m.Attributes
	}
	return nil
}

func (m *Edge) GetFrom() *Path {
	if m != nil {
		return m.From
	}
	return nil
}

func (m *Edge) GetTo() *Path {
	if m != nil {
		return m.To
	}
	return nil
}

func (m *Edge) GetCreatedAt() *timestamp.Timestamp {
	if m != nil {
		return m.CreatedAt
	}
	return nil
}

func (m *Edge) GetUpdatedAt() *timestamp.Timestamp {
	if m != nil {
		return m.UpdatedAt
	}
	return nil
}

type EdgeFilters struct {
	EdgesFrom            []*Filter `protobuf:"bytes,1,rep,name=edges_from,json=edgesFrom,proto3" json:"edges_from,omitempty"`
	EdgesTo              []*Filter `protobuf:"bytes,2,rep,name=edges_to,json=edgesTo,proto3" json:"edges_to,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *EdgeFilters) Reset()         { *m = EdgeFilters{} }
func (m *EdgeFilters) String() string { return proto.CompactTextString(m) }
func (*EdgeFilters) ProtoMessage()    {}
func (*EdgeFilters) Descriptor() ([]byte, []int) {
	return fileDescriptor_063490d3009de3e6, []int{9}
}

func (m *EdgeFilters) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EdgeFilters.Unmarshal(m, b)
}
func (m *EdgeFilters) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EdgeFilters.Marshal(b, m, deterministic)
}
func (m *EdgeFilters) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EdgeFilters.Merge(m, src)
}
func (m *EdgeFilters) XXX_Size() int {
	return xxx_messageInfo_EdgeFilters.Size(m)
}
func (m *EdgeFilters) XXX_DiscardUnknown() {
	xxx_messageInfo_EdgeFilters.DiscardUnknown(m)
}

var xxx_messageInfo_EdgeFilters proto.InternalMessageInfo

func (m *EdgeFilters) GetEdgesFrom() []*Filter {
	if m != nil {
		return m.EdgesFrom
	}
	return nil
}

func (m *EdgeFilters) GetEdgesTo() []*Filter {
	if m != nil {
		return m.EdgesTo
	}
	return nil
}

func init() {
	proto.RegisterEnum("api.Op", Op_name, Op_value)
	proto.RegisterEnum("api.Keyword", Keyword_name, Keyword_value)
	proto.RegisterType((*Command)(nil), "api.Command")
	proto.RegisterType((*Counter)(nil), "api.Counter")
	proto.RegisterType((*Path)(nil), "api.Path")
	proto.RegisterType((*Node)(nil), "api.Node")
	proto.RegisterType((*Message)(nil), "api.Message")
	proto.RegisterType((*Filter)(nil), "api.Filter")
	proto.RegisterType((*Export)(nil), "api.Export")
	proto.RegisterType((*Patch)(nil), "api.Patch")
	proto.RegisterType((*Edge)(nil), "api.Edge")
	proto.RegisterType((*EdgeFilters)(nil), "api.EdgeFilters")
}

func init() { proto.RegisterFile("api/graphik.proto", fileDescriptor_063490d3009de3e6) }

var fileDescriptor_063490d3009de3e6 = []byte{
	// 633 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x93, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0xc7, 0xeb, 0xaf, 0xb8, 0x19, 0x4b, 0x25, 0xac, 0x2a, 0x9a, 0x56, 0xa0, 0x44, 0x3e, 0x54,
	0x55, 0x24, 0x52, 0xa9, 0x1c, 0x80, 0xa3, 0x69, 0x5c, 0x28, 0x94, 0xb6, 0x5a, 0xdc, 0x03, 0x5c,
	0xa2, 0x4d, 0xbc, 0x4d, 0x2c, 0x62, 0xef, 0xca, 0x5e, 0x43, 0x73, 0xe7, 0x59, 0x78, 0x22, 0x1e,
	0x08, 0xed, 0xae, 0xdd, 0x9a, 0xf4, 0x90, 0xd2, 0x9b, 0x67, 0xe6, 0x37, 0xb3, 0x33, 0xff, 0x19,
	0xc3, 0x53, 0xc2, 0x93, 0xc3, 0x59, 0x4e, 0xf8, 0x3c, 0xf9, 0x3e, 0xe4, 0x39, 0x13, 0x0c, 0x59,
	0x84, 0x27, 0x7b, 0xcf, 0x67, 0x8c, 0xcd, 0x16, 0xf4, 0x50, 0xb9, 0x26, 0xe5, 0xf5, 0x61, 0x21,
	0xf2, 0x72, 0x2a, 0x34, 0xb2, 0xd7, 0x5b, 0x8d, 0x8a, 0x24, 0xa5, 0x85, 0x20, 0x29, 0xaf, 0x80,
	0xdd, 0x55, 0x80, 0x64, 0x4b, 0x1d, 0xf2, 0x7f, 0x19, 0xe0, 0x1e, 0xb3, 0x34, 0x25, 0x59, 0x8c,
	0x76, 0xc0, 0x64, 0xbc, 0x6b, 0xf4, 0x8d, 0x83, 0xad, 0x23, 0x77, 0x48, 0x78, 0x32, 0xbc, 0xe0,
	0xd8, 0x64, 0x1c, 0xed, 0x83, 0xf5, 0x83, 0x2c, 0xba, 0x66, 0xdf, 0x38, 0xf0, 0x8e, 0xb6, 0x87,
	0xba, 0xda, 0xb0, 0xae, 0x36, 0x0c, 0xb2, 0x25, 0x96, 0x00, 0x7a, 0x03, 0xed, 0xdb, 0xa7, 0xbb,
	0x96, 0xa2, 0xf7, 0xee, 0xd1, 0x51, 0x4d, 0xe0, 0x3b, 0xd8, 0xef, 0xc9, 0x2e, 0xca, 0x4c, 0xd0,
	0x1c, 0x6d, 0x83, 0x33, 0x95, 0x9f, 0xaa, 0x11, 0x0b, 0x6b, 0xc3, 0x1f, 0x80, 0x7d, 0x49, 0xc4,
	0x1c, 0x21, 0xb0, 0xa3, 0x25, 0xa7, 0x2a, 0xd8, 0xc6, 0xea, 0x1b, 0x6d, 0x81, 0x79, 0x3a, 0x52,
	0xdd, 0xb5, 0xb1, 0x79, 0x3a, 0xf2, 0xff, 0x18, 0x60, 0x9f, 0xb3, 0x98, 0xa2, 0x17, 0x60, 0x73,
	0x22, 0xe6, 0x0a, 0xf6, 0x8e, 0xda, 0x6a, 0x24, 0x59, 0x05, 0x2b, 0x37, 0x7a, 0x0d, 0x40, 0x84,
	0xc8, 0x93, 0x49, 0x29, 0x68, 0x51, 0x4d, 0xb7, 0x73, 0xaf, 0xdf, 0x2f, 0x4a, 0x6a, 0xdc, 0x40,
	0xd1, 0x5b, 0x80, 0x69, 0x4e, 0x89, 0xa0, 0xf1, 0x98, 0x88, 0x87, 0x0c, 0x5a, 0xd1, 0x81, 0x90,
	0xa9, 0x25, 0x8f, 0xeb, 0x54, 0x7b, 0x7d, 0x6a, 0x45, 0x07, 0xc2, 0xe7, 0xe0, 0x7e, 0xa6, 0x45,
	0x41, 0x66, 0x14, 0x75, 0xc1, 0x9d, 0xce, 0x49, 0x96, 0xd1, 0x45, 0x25, 0x44, 0x6d, 0x4a, 0x7d,
	0x84, 0xd4, 0x47, 0xab, 0xa1, 0xbe, 0x57, 0xe6, 0xb4, 0x1e, 0x3c, 0xa7, 0x1f, 0x41, 0xeb, 0x24,
	0x59, 0xc8, 0xa5, 0xd4, 0x65, 0x8d, 0x46, 0xd9, 0x3e, 0x78, 0xf4, 0x86, 0xe7, 0xb4, 0x28, 0x12,
	0x96, 0x49, 0xfd, 0xac, 0x83, 0x36, 0x6e, 0xba, 0xe4, 0x2a, 0x17, 0x49, 0x9a, 0x68, 0x89, 0x1c,
	0xac, 0x0d, 0xff, 0x23, 0xb4, 0xc2, 0x1b, 0xce, 0x72, 0x81, 0x7a, 0xe0, 0x64, 0x2c, 0xa6, 0x45,
	0xd7, 0xe8, 0x5b, 0xb7, 0x0b, 0x92, 0x9b, 0xc3, 0xda, 0x2f, 0x01, 0x1a, 0xcf, 0xa8, 0x2e, 0x5e,
	0x03, 0x61, 0x3c, 0xa3, 0x58, 0xfb, 0xfd, 0x2b, 0x70, 0x2e, 0x89, 0x98, 0xce, 0xd7, 0xad, 0xfa,
	0x25, 0x38, 0x5c, 0x72, 0xeb, 0xb6, 0xac, 0x29, 0xff, 0xb7, 0x09, 0xb6, 0x7c, 0x66, 0x5d, 0xd9,
	0x67, 0xd0, 0x4a, 0x4b, 0x51, 0x56, 0xff, 0xc6, 0x26, 0xae, 0xac, 0x47, 0x2b, 0x2e, 0xdf, 0xbb,
	0xce, 0x59, 0x5a, 0x1d, 0x46, 0xf3, 0x3d, 0xe9, 0x46, 0xbb, 0x60, 0x0a, 0xd6, 0x75, 0x56, 0x83,
	0xa6, 0x60, 0x2b, 0x37, 0xd9, 0x7a, 0xfc, 0x4d, 0xba, 0xff, 0x73, 0x93, 0x04, 0x3c, 0xa9, 0x93,
	0xbe, 0x92, 0x02, 0x0d, 0x00, 0xd4, 0x5e, 0xc6, 0x6a, 0x08, 0xbd, 0x55, 0x4f, 0xf5, 0xa9, 0x09,
	0xdc, 0x56, 0xe1, 0x13, 0x39, 0xcb, 0x3e, 0x6c, 0x6a, 0x56, 0xb0, 0x6a, 0xbd, 0xff, 0x90, 0xae,
	0x0a, 0x46, 0x6c, 0x30, 0x07, 0xf3, 0x82, 0xa3, 0x27, 0xe0, 0x1d, 0xe3, 0x30, 0x88, 0xc2, 0xf1,
	0xf9, 0xc5, 0x28, 0xec, 0x6c, 0xa0, 0x2d, 0x80, 0xcb, 0x20, 0x3a, 0xfe, 0xa0, 0x6d, 0x43, 0x02,
	0xa3, 0xf0, 0x2c, 0xac, 0x01, 0xb3, 0x91, 0x11, 0x8e, 0xde, 0x87, 0x1d, 0xeb, 0x2e, 0x43, 0xd9,
	0x76, 0x23, 0x43, 0x39, 0x9c, 0x41, 0x0f, 0xdc, 0x4f, 0x74, 0xf9, 0x93, 0xe5, 0x31, 0x72, 0xc1,
	0x0a, 0xce, 0xbf, 0x76, 0x36, 0x90, 0x07, 0xee, 0x28, 0x3c, 0x09, 0xae, 0xce, 0xa2, 0x8e, 0xf1,
	0xce, 0xfd, 0xe6, 0x10, 0x9e, 0xf0, 0xc9, 0xa4, 0xa5, 0x54, 0x79, 0xf5, 0x37, 0x00, 0x00, 0xff,
	0xff, 0xa6, 0x0d, 0xfe, 0xb8, 0xb0, 0x05, 0x00, 0x00,
}

package apipb

import (
	"github.com/autom8ter/graphik/sortable"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	Any = "*"
)

func NewStruct(data map[string]interface{}) *structpb.Struct {
	x, _ := structpb.NewStruct(data)
	return x
}

type Mapper interface {
	AsMap() map[string]interface{}
}

func (m *Path) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"gid":   m.GetGid(),
		"gtype": m.GetGtype(),
	}
}

func (m *Path) FromMap(data map[string]interface{}) {
	if val, ok := data["gid"]; ok {
		m.Gid = val.(string)
	}
	if val, ok := data["gtype"]; ok {
		m.Gtype = val.(string)
	}
}


func (m *Metadata) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"created_at": m.GetCreatedAt(),
		"updated_at": m.GetUpdatedAt(),
		"created_by": m.GetCreatedBy(),
		"updated_by": m.GetUpdatedBy(),
	}
}

func (m *Metadata) FromMap(data map[string]interface{}) {
	if val, ok := data["created_at"]; ok {
		m.CreatedAt = val.(int64)
	}
	if val, ok := data["updated_at"]; ok {
		m.UpdatedAt = val.(int64)
	}
	if val, ok := data["created_by"]; ok {
		m.CreatedBy = val.(string)
	}
	if val, ok := data["updated_by"]; ok {
		m.UpdatedBy = val.(string)
	}
}

func (n *Node) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"path":       n.GetPath().AsMap(),
		"attributes": n.GetAttributes().AsMap(),
		"metadata":   n.GetMetadata().AsMap(),
	}
}

func (m *Node) FromMap(data map[string]interface{}) {
	if val, ok := data["metadata"]; ok {
		if m.Metadata == nil {
			m.Metadata = &Metadata{}
		}
		m.Metadata.FromMap(val.(map[string]interface{}))
	}
	if val, ok := data["path"]; ok {
		if m.Path == nil {
			m.Path = &Path{}
		}
		m.Path.FromMap(val.(map[string]interface{}))
	}
	if val, ok := data["attributes"]; ok {
		m.Attributes = NewStruct(val.(map[string]interface{}))
	}
}

func (n *Edge) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"path":       n.GetPath().AsMap(),
		"attributes": n.GetAttributes().AsMap(),
		"from":       n.GetFrom().AsMap(),
		"to":         n.GetTo().AsMap(),
		"metadata":   n.GetMetadata().AsMap(),
	}
}


func (m *Edge) FromMap(data map[string]interface{}) {
	if val, ok := data["metadata"]; ok {
		if m.Metadata == nil {
			m.Metadata = &Metadata{}
		}
		m.Metadata.FromMap(val.(map[string]interface{}))
	}
	if val, ok := data["path"]; ok {
		if m.Path == nil {
			m.Path = &Path{}
		}
		m.Path.FromMap(val.(map[string]interface{}))
	}
	if val, ok := data["from"]; ok {
		if m.From == nil {
			m.From = &Path{}
		}
		m.From.FromMap(val.(map[string]interface{}))
	}
	if val, ok := data["to"]; ok {
		if m.To == nil {
			m.To = &Path{}
		}
		m.To.FromMap(val.(map[string]interface{}))
	}
	if val, ok := data["attributes"]; ok {
		m.Attributes = NewStruct(val.(map[string]interface{}))
	}
}

func (n *Message) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"channel":   n.GetChannel(),
		"sender":    n.GetSender().AsMap(),
		"data":      n.Data.AsMap(),
		"timestamp": n.GetTimestamp(),
	}
}

func (m *Message) FromMap(data map[string]interface{}) {
	if val, ok := data["sender"]; ok {
		if m.Sender == nil {
			m.Sender = &Path{}
		}
		m.Sender.FromMap(val.(map[string]interface{}))
	}
	if val, ok := data["timestamp"]; ok {
		m.Timestamp = val.(int64)
	}
	if val, ok := data["data"]; ok {
		m.Data = NewStruct(val.(map[string]interface{}))
	}
	if val, ok := data["channel"]; ok {
		m.Channel = val.(string)
	}
}

func (n *Nodes) Sort() {
	s := sortable.Sortable{
		LenFunc: func() int {
			return len(n.GetNodes())
		},
		LessFunc: func(i, j int) bool {
			return n.GetNodes()[i].GetMetadata().GetUpdatedAt() < n.GetNodes()[j].GetMetadata().GetUpdatedAt()
		},
		SwapFunc: func(i, j int) {
			n.GetNodes()[i], n.GetNodes()[j] = n.GetNodes()[j], n.GetNodes()[i]
		},
	}
	s.Sort()
}

func (e *Edges) Sort() {
	s := sortable.Sortable{
		LenFunc: func() int {
			return len(e.GetEdges())
		},
		LessFunc: func(i, j int) bool {
			return e.GetEdges()[i].GetMetadata().GetUpdatedAt() < e.GetEdges()[j].GetMetadata().GetUpdatedAt()
		},
		SwapFunc: func(i, j int) {
			e.GetEdges()[i], e.GetEdges()[j] = e.GetEdges()[j], e.GetEdges()[i]
		},
	}
	s.Sort()
}

func (e *EdgeDetails) Sort() {
	s := sortable.Sortable{
		LenFunc: func() int {
			return len(e.GetEdges())
		},
		LessFunc: func(i, j int) bool {
			return e.GetEdges()[i].GetMetadata().GetUpdatedAt() < e.GetEdges()[j].GetMetadata().GetUpdatedAt()
		},
		SwapFunc: func(i, j int) {
			e.GetEdges()[i], e.GetEdges()[j] = e.GetEdges()[j], e.GetEdges()[i]
		},
	}
	s.Sort()
}

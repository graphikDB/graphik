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

func (m *Metadata) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"created_at": m.GetCreatedAt(),
		"updated_at": m.GetUpdatedAt(),
		"created_by": m.GetCreatedBy(),
		"updated_by": m.GetUpdatedBy(),
	}
}

func (n *Node) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"path":       n.GetPath().AsMap(),
		"attributes": n.GetAttributes(),
		"metadata":   n.GetMetadata().AsMap(),
	}
}

func (n *Edge) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"path":       n.GetPath().AsMap(),
		"attributes": n.GetAttributes(),
		"from":       n.GetFrom().AsMap(),
		"to":         n.GetTo().AsMap(),
		"metadata":   n.GetMetadata().AsMap(),
	}
}

func (n *Message) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"channel":   n.GetChannel(),
		"sender":    n.GetSender().AsMap(),
		"data":      n.Data,
		"timestamp": n.GetTimestamp(),
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

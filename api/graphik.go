package apipb

import (
	"github.com/autom8ter/graphik/sortable"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
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

type FromMapper interface {
	FromMap(data map[string]interface{})
}

func (m *Path) AsMap() map[string]interface{} {
	if m == nil {
		return map[string]interface{}{}
	}
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
	if m == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"created_at": m.GetCreatedAt(),
		"updated_at": m.GetUpdatedAt(),
		"updated_by": m.GetUpdatedBy(),
		"sequence":   m.GetSequence(),
		"version":    m.GetVersion(),
		"hash":       m.GetHash(),
	}
}

func (m *Metadata) FromMap(data map[string]interface{}) {
	if val, ok := data["created_at"]; ok {
		m.CreatedAt = timestamppb.New(val.(time.Time))
	}
	if val, ok := data["updated_at"]; ok {
		m.UpdatedAt = timestamppb.New(val.(time.Time))
	}
	if val, ok := data["sequence"]; ok {
		m.Sequence = val.(uint64)
	}
	if val, ok := data["updated_by"]; ok {
		if val, ok := val.(map[string]interface{}); ok {
			m.UpdatedBy.FromMap(val)
		}
	}
}

func (n *Node) AsMap() map[string]interface{} {
	if n == nil {
		return map[string]interface{}{}
	}
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
		if val, ok := val.(map[string]interface{}); ok {
			m.Metadata.FromMap(val)
		}
		if val, ok := val.(*Metadata); ok {
			m.Metadata = val
		}
	}
	if val, ok := data["path"]; ok {
		if m.Path == nil {
			m.Path = &Path{}
		}
		if val, ok := val.(map[string]interface{}); ok {
			m.Path.FromMap(val)
		}
		if val, ok := val.(*Path); ok {
			m.Path = val
		}
	}
	if val, ok := data["attributes"]; ok {
		if val, ok := val.(map[string]interface{}); ok {
			m.Attributes = NewStruct(val)
		}
		if val, ok := val.(*structpb.Struct); ok {
			m.Attributes = val
		}
	}
}

func (n *Edge) AsMap() map[string]interface{} {
	if n == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"path":       n.GetPath().AsMap(),
		"attributes": n.GetAttributes().AsMap(),
		"directed":   n.GetDirected(),
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
		if val, ok := val.(map[string]interface{}); ok {
			m.Path.FromMap(val)
		}
		if val, ok := val.(*Path); ok {
			m.Path = val
		}
	}
	if val, ok := data["from"]; ok {
		if m.From == nil {
			m.From = &Path{}
		}
		if val, ok := val.(map[string]interface{}); ok {
			m.From.FromMap(val)
		}
		if val, ok := val.(*Path); ok {
			m.From = val
		}
	}
	if val, ok := data["to"]; ok {
		if m.To == nil {
			m.To = &Path{}
		}
		if val, ok := val.(map[string]interface{}); ok {
			m.To.FromMap(val)
		}
		if val, ok := val.(*Path); ok {
			m.To = val
		}
	}
	if val, ok := data["attributes"]; ok {
		if val, ok := val.(map[string]interface{}); ok {
			m.Attributes = NewStruct(val)
		}
		if val, ok := val.(*structpb.Struct); ok {
			m.Attributes = val
		}
	}
}

func (n *Message) AsMap() map[string]interface{} {
	if n == nil {
		return map[string]interface{}{}
	}
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
		if val, ok := val.(map[string]interface{}); ok {
			m.Sender.FromMap(val)
		}
		if val, ok := val.(*Path); ok {
			m.Sender = val
		}
	}
	if val, ok := data["timestamp"]; ok {
		m.Timestamp = timestamppb.New(val.(time.Time))
	}
	if val, ok := data["data"]; ok {
		m.Data = NewStruct(val.(map[string]interface{}))
	}
	if val, ok := data["channel"]; ok {
		m.Channel = val.(string)
	}
}

func (c *Change) AsMap() map[string]interface{} {
	if c == nil {
		return map[string]interface{}{}
	}
	var edgeChanges []interface{}
	var nodeChanges []interface{}
	for _, change := range c.GetEdgeChanges() {
		edgeChanges = append(edgeChanges, map[string]interface{}{
			"before": change.GetBefore().AsMap(),
			"after":  change.GetAfter().AsMap(),
		})
	}
	for _, change := range c.GetNodeChanges() {
		nodeChanges = append(nodeChanges, map[string]interface{}{
			"before": change.GetBefore().AsMap(),
			"after":  change.GetAfter().AsMap(),
		})
	}
	return map[string]interface{}{
		"method":       c.Method,
		"timestamp":    c.Timestamp,
		"identity":     c.Identity,
		"edge_changes": edgeChanges,
		"node_changes": nodeChanges,
	}
}

func (n *Filter) AsMap() map[string]interface{} {
	if n == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"gtype":       n.GetGtype(),
		"expressions": n.GetExpressions(),
		"limit":       n.GetLimit(),
	}
}

func (n *EdgeFilter) AsMap() map[string]interface{} {
	if n == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"gtype":       n.GetGtype(),
		"expressions": n.GetExpressions(),
		"limit":       n.GetLimit(),
		"node_path":   n.GetNodePath(),
	}
}

func (n *ChannelFilter) AsMap() map[string]interface{} {
	if n == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"channel":     n.GetChannel(),
		"expressions": n.GetExpressions(),
	}
}

func (n *ExpressionFilter) AsMap() map[string]interface{} {
	if n == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"expressions": n.GetExpressions(),
	}
}

func (p *Patch) AsMap() map[string]interface{} {
	if p == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"path":       p.GetPath(),
		"attributes": p.GetAttributes(),
	}
}

func (n *PatchFilter) AsMap() map[string]interface{} {
	if n == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"patch":  n.GetPatch().AsMap(),
		"filter": n.GetFilter().AsMap(),
	}
}

func (m *MeFilter) AsMap() map[string]interface{} {
	if m == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"from_edges": m.GetEdgesFrom().AsMap(),
		"to_edges":   m.GetEdgesTo().AsMap(),
	}
}

func (e *EdgeConstructor) AsMap() map[string]interface{} {
	if e == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"path":       e.GetPath().AsMap(),
		"attributes": e.GetAttributes().AsMap(),
		"directed":   e.GetDirected(),
		"from":       e.GetFrom().AsMap(),
		"to":         e.GetTo().AsMap(),
	}
}

func (e *NodeConstructor) AsMap() map[string]interface{} {
	if e == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"path":       e.GetPath().AsMap(),
		"attributes": e.GetAttributes().AsMap(),
	}
}

func (p *Paths) Sort() {
	s := sortable.Sortable{
		LenFunc: func() int {
			return len(p.GetPaths())
		},
		LessFunc: func(i, j int) bool {
			return p.GetPaths()[i].String() < p.GetPaths()[j].String()
		},
		SwapFunc: func(i, j int) {
			p.GetPaths()[i], p.GetPaths()[j] = p.GetPaths()[j], p.GetPaths()[i]
		},
	}
	s.Sort()
}

func (n *Nodes) Sort() {
	s := sortable.Sortable{
		LenFunc: func() int {
			return len(n.GetNodes())
		},
		LessFunc: func(i, j int) bool {
			return n.GetNodes()[i].GetMetadata().GetUpdatedAt().AsTime().Nanosecond() < n.GetNodes()[j].GetMetadata().GetUpdatedAt().AsTime().Nanosecond()
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
			return e.GetEdges()[i].GetMetadata().GetUpdatedAt().AsTime().Nanosecond() < e.GetEdges()[j].GetMetadata().GetUpdatedAt().AsTime().Nanosecond()
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
			return e.GetEdges()[i].GetMetadata().GetUpdatedAt().AsTime().Nanosecond() < e.GetEdges()[j].GetMetadata().GetUpdatedAt().AsTime().Nanosecond()
		},
		SwapFunc: func(i, j int) {
			e.GetEdges()[i], e.GetEdges()[j] = e.GetEdges()[j], e.GetEdges()[i]
		},
	}
	s.Sort()
}

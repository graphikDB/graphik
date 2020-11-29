package apipb

import (
	"fmt"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sort"
	"strings"
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

func (n *Doc) AsMap() map[string]interface{} {
	if n == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"path":       n.GetPath().AsMap(),
		"attributes": n.GetAttributes().AsMap(),
		"metadata":   n.GetMetadata().AsMap(),
	}
}

func (m *Doc) FromMap(data map[string]interface{}) {
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

func (n *Connection) AsMap() map[string]interface{} {
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

func (m *Connection) FromMap(data map[string]interface{}) {
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
		"data":      n.GetData().AsMap(),
		"timestamp": n.GetTimestamp(),
	}
}

func (c *Change) AsMap() map[string]interface{} {
	if c == nil {
		return map[string]interface{}{}
	}
	var connectionChanges []interface{}
	var docChanges []interface{}
	for _, change := range c.GetConnectionChanges() {
		connectionChanges = append(connectionChanges, map[string]interface{}{
			"before": change.GetBefore().AsMap(),
			"after":  change.GetAfter().AsMap(),
		})
	}
	for _, change := range c.GetDocChanges() {
		docChanges = append(docChanges, map[string]interface{}{
			"before": change.GetBefore().AsMap(),
			"after":  change.GetAfter().AsMap(),
		})
	}
	return map[string]interface{}{
		"method":             c.GetMethod(),
		"timestamp":          c.GetTimestamp(),
		"identity":           c.GetIdentity(),
		"connection_changes": connectionChanges,
		"doc_changes":        docChanges,
	}
}

func (n *Filter) AsMap() map[string]interface{} {
	if n == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"gtype":      n.GetGtype(),
		"expression": n.GetExpression(),
		"limit":      n.GetLimit(),
		"sort":       n.GetSort(),
		"reverse":    n.GetReverse(),
		"seek":       n.GetSeek(),
	}
}

func (n *ConnectionFilter) AsMap() map[string]interface{} {
	if n == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"gtype":      n.GetGtype(),
		"expression": n.GetExpression(),
		"doc_path":   n.GetDocPath(),
		"limit":      n.GetLimit(),
		"sort":       n.GetSort(),
		"reverse":    n.GetReverse(),
		"seek":       n.GetSeek(),
	}
}

func (n *SubGraphFilter) AsMap() map[string]interface{} {
	if n == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"connection_filter": n.GetConnectionFilter().AsMap(),
		"doc_filter":        n.GetDocFilter().AsMap(),
	}
}

func (n *ChannelFilter) AsMap() map[string]interface{} {
	if n == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"channel":    n.GetChannel(),
		"expression": n.GetExpression(),
	}
}

func (n *ExpressionFilter) AsMap() map[string]interface{} {
	if n == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"expression": n.GetExpression(),
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
		"patch":  n.GetAttributes().AsMap(),
		"filter": n.GetFilter().AsMap(),
	}
}

func (m *MeFilter) AsMap() map[string]interface{} {
	if m == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"from_connections": m.GetConnectionsFrom().AsMap(),
		"to_connections":   m.GetConnectionsTo().AsMap(),
	}
}

func (e *ConnectionConstructor) AsMap() map[string]interface{} {
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

func (e *DocConstructor) AsMap() map[string]interface{} {
	if e == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"path":       e.GetPath().AsMap(),
		"attributes": e.GetAttributes().AsMap(),
	}
}

func (o *OutboundMessage) AsMap() map[string]interface{} {
	if o == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"channel": o.GetChannel(),
		"data":    o.GetData().AsMap(),
	}
}

func (e *Connections) AsMap() map[string]interface{} {
	var connections []interface{}
	for _, connection := range e.GetConnections() {
		connections = append(connections, connection.AsMap())
	}
	return map[string]interface{}{
		"connections": connections,
	}
}

func (e *Docs) AsMap() map[string]interface{} {
	var docs []interface{}
	for _, connection := range e.GetDocs() {
		docs = append(docs, connection.AsMap())
	}
	return map[string]interface{}{
		"docs": docs,
	}
}

func (g *Graph) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"connections": g.GetConnections().AsMap(),
		"docs":        g.GetDocs().AsMap(),
	}
}

func (p *Paths) Sort(field string) {
	if p == nil {
		return
	}
	switch field {
	case "path.gid":
		sort.Slice(p.GetPaths(), func(i, j int) bool {
			return p.GetPaths()[i].GetGid() < p.GetPaths()[j].GetGid()
		})
	default:
		sort.Slice(p.GetPaths(), func(i, j int) bool {
			return p.GetPaths()[i].GetGtype() < p.GetPaths()[j].GetGtype()
		})
	}
}

func (n *Docs) Sort(field string) {
	if n == nil {
		return
	}
	switch {
	case field == "path.gid":
		sort.Slice(n.GetDocs(), func(i, j int) bool {
			return n.GetDocs()[i].GetPath().GetGid() < n.GetDocs()[j].GetPath().GetGid()
		})
	case field == "path.gtype":
		sort.Slice(n.GetDocs(), func(i, j int) bool {
			return n.GetDocs()[i].GetPath().GetGtype() < n.GetDocs()[j].GetPath().GetGtype()
		})
	case field == "metadata.sequence":
		sort.Slice(n.GetDocs(), func(i, j int) bool {
			return n.GetDocs()[i].GetMetadata().GetSequence() < n.GetDocs()[j].GetMetadata().GetSequence()
		})
	case field == "metadata.version":
		sort.Slice(n.GetDocs(), func(i, j int) bool {
			return n.GetDocs()[i].GetMetadata().GetVersion() < n.GetDocs()[j].GetMetadata().GetVersion()
		})
	case field == "metadata.created_at":
		sort.Slice(n.GetDocs(), func(i, j int) bool {
			return n.GetDocs()[i].GetMetadata().GetCreatedAt().AsTime().Nanosecond() < n.GetDocs()[j].GetMetadata().GetCreatedAt().AsTime().Nanosecond()
		})
	case strings.Contains(field, "attributes."):
		split := strings.Split(field, "attributes.")
		if len(split) == 2 {
			key := split[1]
			sort.Slice(n.GetDocs(), func(i, j int) bool {
				fields := n.GetDocs()[i].GetAttributes().GetFields()
				if fields == nil {
					return true
				}
				if fields[key] == nil {
					return true
				}
				switch n.GetDocs()[i].GetAttributes().GetFields()[key].GetKind() {
				case &structpb.Value_NumberValue{}:
					return n.GetDocs()[i].GetAttributes().GetFields()[key].GetNumberValue() < n.GetDocs()[j].GetAttributes().GetFields()[key].GetNumberValue()
				case &structpb.Value_StringValue{}:
					return n.GetDocs()[i].GetAttributes().GetFields()[key].GetStringValue() < n.GetDocs()[j].GetAttributes().GetFields()[key].GetStringValue()
				default:
					return fmt.Sprint(n.GetDocs()[i].GetAttributes().GetFields()[key].AsInterface()) < fmt.Sprint(n.GetDocs()[j].GetAttributes().GetFields()[key].AsInterface())
				}
			})
		}
	default:
		sort.Slice(n.GetDocs(), func(i, j int) bool {
			return n.GetDocs()[i].GetMetadata().GetUpdatedAt().AsTime().Nanosecond() < n.GetDocs()[j].GetMetadata().GetUpdatedAt().AsTime().Nanosecond()
		})
	}
}

func (e *Connections) Sort(field string) {
	if e == nil {
		return
	}
	switch {
	case field == "path.gid":
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetPath().GetGid() < e.GetConnections()[j].GetPath().GetGid()
		})
	case field == "path.gtype":
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetPath().GetGtype() < e.GetConnections()[j].GetPath().GetGtype()
		})
	case field == "metadata.sequence":
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetMetadata().GetSequence() < e.GetConnections()[j].GetMetadata().GetSequence()
		})
	case field == "metadata.version":
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetMetadata().GetVersion() < e.GetConnections()[j].GetMetadata().GetVersion()
		})
	case field == "metadata.updated_at":
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetMetadata().GetUpdatedAt().AsTime().Nanosecond() < e.GetConnections()[j].GetMetadata().GetUpdatedAt().AsTime().Nanosecond()
		})
	case field == "metadata.created_at":
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetMetadata().GetCreatedAt().AsTime().Nanosecond() < e.GetConnections()[j].GetMetadata().GetCreatedAt().AsTime().Nanosecond()
		})
	case strings.Contains(field, "attributes."):
		split := strings.Split(field, "attributes.")
		if len(split) == 2 {
			key := split[1]
			sort.Slice(e.GetConnections(), func(i, j int) bool {
				fields := e.GetConnections()[i].GetAttributes().GetFields()
				if fields == nil {
					return true
				}
				if fields[key] == nil {
					return true
				}
				switch e.GetConnections()[i].GetAttributes().GetFields()[key].GetKind() {
				case &structpb.Value_NumberValue{}:
					return e.GetConnections()[i].GetAttributes().GetFields()[key].GetNumberValue() < e.GetConnections()[j].GetAttributes().GetFields()[key].GetNumberValue()
				case &structpb.Value_StringValue{}:
					return e.GetConnections()[i].GetAttributes().GetFields()[key].GetStringValue() < e.GetConnections()[j].GetAttributes().GetFields()[key].GetStringValue()
				default:
					return fmt.Sprint(e.GetConnections()[i].GetAttributes().GetFields()[key].AsInterface()) < fmt.Sprint(e.GetConnections()[j].GetAttributes().GetFields()[key].AsInterface())
				}
			})
		}
	default:
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetMetadata().GetUpdatedAt().AsTime().Nanosecond() < e.GetConnections()[j].GetMetadata().GetUpdatedAt().AsTime().Nanosecond()
		})
	}
}

func (e *ConnectionDetails) Sort(field string) {
	if e == nil {
		return
	}
	switch {
	case field == "path.gid":
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetPath().GetGid() < e.GetConnections()[j].GetPath().GetGid()
		})
	case field == "path.gtype":
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetPath().GetGtype() < e.GetConnections()[j].GetPath().GetGtype()
		})
	case field == "to.path.gtype":
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetTo().GetPath().GetGtype() < e.GetConnections()[j].GetTo().GetPath().GetGtype()
		})
	case field == "to.path.gid":
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetTo().GetPath().GetGid() < e.GetConnections()[j].GetTo().GetPath().GetGid()
		})
	case field == "from.path.gtype":
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetFrom().GetPath().GetGtype() < e.GetConnections()[j].GetFrom().GetPath().GetGtype()
		})
	case field == "from.path.gid":
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetFrom().GetPath().GetGid() < e.GetConnections()[j].GetFrom().GetPath().GetGid()
		})
	case field == "metadata.sequence":
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetMetadata().GetSequence() < e.GetConnections()[j].GetMetadata().GetSequence()
		})
	case field == "metadata.version":
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetMetadata().GetVersion() < e.GetConnections()[j].GetMetadata().GetVersion()
		})
	case field == "metadata.updated_at":
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetMetadata().GetUpdatedAt().AsTime().Nanosecond() < e.GetConnections()[j].GetMetadata().GetUpdatedAt().AsTime().Nanosecond()
		})
	case field == "metadata.created_at":
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetMetadata().GetCreatedAt().AsTime().Nanosecond() < e.GetConnections()[j].GetMetadata().GetCreatedAt().AsTime().Nanosecond()
		})
	case strings.Contains(field, "attributes."):
		split := strings.Split(field, "attributes.")
		if len(split) == 2 {
			key := split[1]
			sort.Slice(e.GetConnections(), func(i, j int) bool {
				fields := e.GetConnections()[i].GetAttributes().GetFields()
				if fields == nil {
					return true
				}
				if fields[key] == nil {
					return true
				}
				switch e.GetConnections()[i].GetAttributes().GetFields()[key].GetKind() {
				case &structpb.Value_NumberValue{}:
					return e.GetConnections()[i].GetAttributes().GetFields()[key].GetNumberValue() < e.GetConnections()[j].GetAttributes().GetFields()[key].GetNumberValue()
				case &structpb.Value_StringValue{}:
					return e.GetConnections()[i].GetAttributes().GetFields()[key].GetStringValue() < e.GetConnections()[j].GetAttributes().GetFields()[key].GetStringValue()
				default:
					return fmt.Sprint(e.GetConnections()[i].GetAttributes().GetFields()[key].AsInterface()) < fmt.Sprint(e.GetConnections()[j].GetAttributes().GetFields()[key].AsInterface())
				}
			})
		}
	default:
		sort.Slice(e.GetConnections(), func(i, j int) bool {
			return e.GetConnections()[i].GetMetadata().GetUpdatedAt().AsTime().Nanosecond() < e.GetConnections()[j].GetMetadata().GetUpdatedAt().AsTime().Nanosecond()
		})
	}

}

func (n *Docs) Reverse() {
	if n == nil {
		return
	}
	docs := n.GetDocs()
	for i := 0; i < len(docs)/2; i++ {
		j := len(docs) - i - 1
		docs[i], docs[j] = docs[j], docs[i]
	}
	n.Docs = docs
}

func (n *Connections) Reverse() {
	if n == nil {
		return
	}
	values := n.GetConnections()
	for i := 0; i < len(values)/2; i++ {
		j := len(values) - i - 1
		values[i], values[j] = values[j], values[i]
	}
	n.Connections = values
}

func (n *ConnectionDetails) Reverse() {
	if n == nil {
		return
	}
	values := n.GetConnections()
	for i := 0; i < len(values)/2; i++ {
		j := len(values) - i - 1
		values[i], values[j] = values[j], values[i]
	}
	n.Connections = values
}

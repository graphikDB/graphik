package apipb

import (
	"fmt"
	"gonum.org/v1/gonum/floats"
	"google.golang.org/protobuf/types/known/structpb"
	"sort"
	"strings"
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
	if m == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"gid":   m.GetGid(),
		"gtype": m.GetGtype(),
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
		"version":    m.GetVersion(),
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

func (p *Paths) AsMap() map[string]interface{} {
	if p == nil {
		return map[string]interface{}{}
	}
	var paths []interface{}
	for _, path := range p.GetPaths() {
		paths = append(paths, path.AsMap())
	}
	return map[string]interface{}{
		"paths": paths,
	}
}
func (c *Change) AsMap() map[string]interface{} {
	if c == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"method":         c.GetMethod(),
		"timestamp":      c.GetTimestamp(),
		"identity":       c.GetIdentity(),
		"paths_affected": c.PathsAffected.AsMap(),
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

func (p *Edit) AsMap() map[string]interface{} {
	if p == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"path":       p.GetPath(),
		"attributes": p.GetAttributes(),
	}
}

func (n *EditFilter) AsMap() map[string]interface{} {
	if n == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"patch":  n.GetAttributes().AsMap(),
		"filter": n.GetFilter().AsMap(),
	}
}

func (e *PathConstructor) AsMap() map[string]interface{} {
	if e == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"gtype": e.GetGtype(),
		"gid":   e.GetGid(),
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
	case "gid":
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

func (n *DocTraversals) Sort(field string) {
	if n == nil {
		return
	}
	switch {
	case field == "path.gid":
		sort.Slice(n.GetTraversals(), func(i, j int) bool {
			return n.GetTraversals()[i].GetDoc().GetPath().GetGid() < n.GetTraversals()[j].GetDoc().GetPath().GetGid()
		})
	case field == "path.gtype":
		sort.Slice(n.GetTraversals(), func(i, j int) bool {
			return n.GetTraversals()[i].GetDoc().GetPath().GetGtype() < n.GetTraversals()[j].GetDoc().GetPath().GetGtype()
		})
	case field == "metadata.version":
		sort.Slice(n.GetTraversals(), func(i, j int) bool {
			return n.GetTraversals()[i].GetDoc().GetMetadata().GetVersion() < n.GetTraversals()[j].GetDoc().GetMetadata().GetVersion()
		})
	case field == "metadata.created_at":
		sort.Slice(n.GetTraversals(), func(i, j int) bool {
			return n.GetTraversals()[i].GetDoc().GetMetadata().GetCreatedAt().AsTime().UnixNano() < n.GetTraversals()[j].GetDoc().GetMetadata().GetCreatedAt().AsTime().UnixNano()
		})
	case field == "relative_path":
		sort.Slice(n.GetTraversals(), func(i, j int) bool {
			return len(n.GetTraversals()[i].GetRelativePath().GetPaths()) < len(n.GetTraversals()[j].GetRelativePath().GetPaths())
		})
	case strings.Contains(field, "attributes."):
		split := strings.Split(field, "attributes.")
		if len(split) == 2 {
			key := split[1]
			sort.Slice(n.GetTraversals(), func(i, j int) bool {
				fields := n.GetTraversals()[i].GetDoc().GetAttributes().GetFields()
				if fields == nil {
					return true
				}
				if fields[key] == nil {
					return true
				}
				switch n.GetTraversals()[i].GetDoc().GetAttributes().GetFields()[key].GetKind() {
				case &structpb.Value_NumberValue{}:
					return n.GetTraversals()[i].GetDoc().GetAttributes().GetFields()[key].GetNumberValue() < n.GetTraversals()[j].GetDoc().GetAttributes().GetFields()[key].GetNumberValue()
				case &structpb.Value_StringValue{}:
					return n.GetTraversals()[i].GetDoc().GetAttributes().GetFields()[key].GetStringValue() < n.GetTraversals()[j].GetDoc().GetAttributes().GetFields()[key].GetStringValue()
				default:
					return fmt.Sprint(n.GetTraversals()[i].GetDoc().GetAttributes().GetFields()[key].AsInterface()) < fmt.Sprint(n.GetTraversals()[j].GetDoc().GetAttributes().GetFields()[key].AsInterface())
				}
			})
		}
	default:
		sort.Slice(n.GetTraversals(), func(i, j int) bool {
			return n.GetTraversals()[i].GetDoc().GetMetadata().GetUpdatedAt().AsTime().UnixNano() < n.GetTraversals()[j].GetDoc().GetMetadata().GetUpdatedAt().AsTime().UnixNano()
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

func (d *Docs) Range(fn func(d *Doc) bool) {
	for _, d := range d.GetDocs() {
		if !fn(d) {
			return
		}
	}
}

func (d *Connections) Range(fn func(d *Connection) bool) {
	for _, d := range d.GetConnections() {
		if !fn(d) {
			return
		}
	}
}

func (d *Docs) Aggregate(aggregate string, field string) float64 {
	if aggregate == "count" {
		return float64(len(d.GetDocs()))
	}
	var values []float64
	d.Range(func(d *Doc) bool {
		switch {
		case field == "metadata.version":
			values = append(values, float64(d.GetMetadata().GetVersion()))
		case field == "metadata.created_at":
			values = append(values, float64(d.GetMetadata().GetCreatedAt().GetSeconds()))
		case field == "metadata.updated_at":
			values = append(values, float64(d.GetMetadata().GetUpdatedAt().GetSeconds()))
		case strings.Contains(field, "attributes."):
			if fields := d.GetAttributes().GetFields(); fields != nil {
				split := strings.Split(field, "attributes.")
				if len(split) == 2 {
					key := split[1]
					if val := fields[key]; val != nil {
						values = append(values, val.GetNumberValue())
					}
				}
			}
		}
		return true
	})
	if len(values) == 0 {
		return 0
	}
	switch aggregate {
	case "sum":
		return floats.Sum(values)
	case "min":
		return floats.Min(values)
	case "max":
		return floats.Max(values)
	case "prod":
		return floats.Prod(values)
	case "avg":
		return floats.Sum(values) / float64(len(values))
	}
	return 0
}

func (c *Connections) Aggregate(aggregate string, field string) float64 {
	if aggregate == "count" {
		return float64(len(c.GetConnections()))
	}
	var values []float64
	c.Range(func(c *Connection) bool {
		switch {
		case field == "metadata.version":
			values = append(values, float64(c.GetMetadata().GetVersion()))
		case field == "metadata.created_at":
			values = append(values, float64(c.GetMetadata().GetCreatedAt().GetSeconds()))
		case field == "metadata.updated_at":
			values = append(values, float64(c.GetMetadata().GetUpdatedAt().GetSeconds()))
		case strings.Contains(field, "attributes."):
			if fields := c.GetAttributes().GetFields(); fields != nil {
				split := strings.Split(field, "attributes.")
				if len(split) == 2 {
					key := split[1]
					if val := fields[key]; val != nil {
						values = append(values, val.GetNumberValue())
					}
				}
			}
		}
		return true
	})
	if len(values) == 0 {
		return 0
	}
	switch aggregate {
	case "sum":
		return floats.Sum(values)
	case "min":
		return floats.Min(values)
	case "max":
		return floats.Max(values)
	case "count":
		return float64(floats.Count(func(f float64) bool {
			return true
		}, values))
	case "prod":
		return floats.Prod(values)
	case "avg":
		return floats.Sum(values) / float64(len(values))
	}
	return 0
}

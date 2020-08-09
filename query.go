package graphik

import "errors"

func NewNodeQuery() NodeQuery {
	return &nodeQuery{}
}

func NewEdgeQuery() EdgeQuery {
	return &edgeQuery{}
}

type nodeQuery struct {
	typ     string
	key     string
	where   WhereFunc
	handler NodeHandlerFunc
	limit   int
}

func (q *nodeQuery) Mod(fn NodeQueryModFunc) NodeQuery {
	return fn(q)
}

func (q *nodeQuery) Handler() NodeHandlerFunc {
	return q.handler
}

func (q *nodeQuery) Type() string {
	return q.typ
}

func (q *nodeQuery) Key() string {
	return q.key
}

func (q *nodeQuery) Limit() int {
	return q.limit
}

func (q *nodeQuery) Where() WhereFunc {
	return q.where
}

func (q *nodeQuery) Validate() (NodeQuery, error) {
	if q.handler == nil {
		return nil, errors.New("graphy: node query handler required")
	}
	return q, nil
}

func NodeModType(typ string) NodeQueryModFunc {
	return func(query NodeQuery) NodeQuery {
		e := &nodeQuery{
			typ:   typ,
			key:   query.Key(),
			limit: query.Limit(),
		}
		if query.Where() != nil {
			e.where = query.Where()
		}
		if query.Handler() != nil {
			e.handler = query.Handler()
		}
		return e
	}
}

func NodeModKey(key string) NodeQueryModFunc {
	return func(query NodeQuery) NodeQuery {
		e := &nodeQuery{
			typ:   query.Type(),
			key:   key,
			limit: query.Limit(),
		}
		if query.Where() != nil {
			e.where = query.Where()
		}
		if query.Handler() != nil {
			e.handler = query.Handler()
		}
		return e
	}
}

func NodeModLimit(limit int) NodeQueryModFunc {
	return func(query NodeQuery) NodeQuery {
		e := &nodeQuery{
			typ:   query.Type(),
			key:   query.Key(),
			limit: limit,
		}
		if query.Where() != nil {
			e.where = query.Where()
		}
		if query.Handler() != nil {
			e.handler = query.Handler()
		}
		return e
	}
}

func NodeModHandler(handler NodeHandlerFunc) NodeQueryModFunc {
	return func(query NodeQuery) NodeQuery {
		e := &nodeQuery{
			typ:     query.Type(),
			key:     query.Key(),
			limit:   query.Limit(),
			handler: handler,
		}
		if query.Where() != nil {
			e.where = query.Where()
		}
		return e
	}
}
func NodeModWhere(where WhereFunc) NodeQueryModFunc {
	return func(query NodeQuery) NodeQuery {
		e := &nodeQuery{
			typ:   query.Type(),
			key:   query.Key(),
			where: where,
			limit: query.Limit(),
		}
		if query.Handler() != nil {
			e.handler = query.Handler()
		}
		return e
	}
}

type edgeQuery struct {
	fromType     string
	fromKey      string
	relationship string
	toType       string
	toKey        string
	where        WhereFunc
	handler      EdgeHandlerFunc
	limit        int
}

func (q *edgeQuery) Mod(fn EdgeQueryModFunc) EdgeQuery {
	return fn(q)
}

func (q *edgeQuery) FromType() string {
	return q.fromType
}

func (q *edgeQuery) FromKey() string {
	return q.fromKey
}

func (q *edgeQuery) Relationship() string {
	return q.relationship
}

func (q *edgeQuery) Validate() (EdgeQuery, error) {
	if q.handler == nil {
		return nil, errors.New("graphy: edge query handler required")
	}
	return q, nil
}

func (q *edgeQuery) ToType() string {
	return q.toType
}

func (q *edgeQuery) ToKey() string {
	return q.toKey
}

func (q *edgeQuery) Limit() int {
	return q.limit
}

func (q *edgeQuery) Where() WhereFunc {
	return q.where
}

func (q *edgeQuery) Handler() EdgeHandlerFunc {
	return q.handler
}

func EdgeModFromType(fromType string) EdgeQueryModFunc {
	return func(query EdgeQuery) EdgeQuery {
		e := &edgeQuery{
			fromType:     fromType,
			fromKey:      query.FromKey(),
			relationship: query.Relationship(),
			toType:       query.ToType(),
			toKey:        query.ToKey(),
			limit:        query.Limit(),
		}
		if query.Where() != nil {
			e.where = query.Where()
		}
		if query.Handler() != nil {
			e.handler = query.Handler()
		}
		return e
	}
}

func EdgeModFromKey(fromKey string) EdgeQueryModFunc {
	return func(query EdgeQuery) EdgeQuery {
		e := &edgeQuery{
			fromType:     query.FromType(),
			fromKey:      fromKey,
			relationship: query.Relationship(),
			toType:       query.ToType(),
			toKey:        query.ToKey(),
			limit:        query.Limit(),
		}
		if query.Where() != nil {
			e.where = query.Where()
		}
		if query.Handler() != nil {
			e.handler = query.Handler()
		}
		return e
	}
}

func EdgeModRelationship(relationship string) EdgeQueryModFunc {
	return func(query EdgeQuery) EdgeQuery {
		e := &edgeQuery{
			fromType:     query.FromType(),
			fromKey:      query.FromKey(),
			relationship: relationship,
			toType:       query.ToType(),
			toKey:        query.ToKey(),
			where:        query.Where(),
			handler:      query.Handler(),
			limit:        query.Limit(),
		}
		return e
	}
}

func EdgeModHandler(handler EdgeHandlerFunc) EdgeQueryModFunc {
	return func(query EdgeQuery) EdgeQuery {
		e := &edgeQuery{
			fromType:     query.FromType(),
			fromKey:      query.FromKey(),
			relationship: query.Relationship(),
			where:        query.Where(),
			handler:      handler,
			toKey:        query.ToKey(),
			toType:       query.ToType(),
			limit:        query.Limit(),
		}
		if query.Where() != nil {
			e.where = query.Where()
		}
		return e
	}
}

func EdgeModToType(toType string) EdgeQueryModFunc {
	return func(query EdgeQuery) EdgeQuery {
		e := &edgeQuery{
			fromType:     query.FromType(),
			fromKey:      query.FromKey(),
			relationship: query.Relationship(),
			toType:       toType,
			toKey:        query.ToKey(),
			limit:        query.Limit(),
			where:        query.Where(),
			handler:      query.Handler(),
		}
		return e
	}
}

func EdgeModToKey(toKey string) EdgeQueryModFunc {
	return func(query EdgeQuery) EdgeQuery {
		e := &edgeQuery{
			fromType:     query.FromType(),
			fromKey:      query.FromKey(),
			relationship: query.Relationship(),
			toType:       query.ToType(),
			toKey:        toKey,
			limit:        query.Limit(),
			where:        query.Where(),
			handler:      query.Handler(),
		}
		return e
	}
}

func EdgeModLimit(limit int) EdgeQueryModFunc {
	return func(query EdgeQuery) EdgeQuery {
		e := &edgeQuery{
			fromType:     query.FromType(),
			fromKey:      query.FromKey(),
			relationship: query.Relationship(),
			toType:       query.ToType(),
			toKey:        query.ToKey(),
			limit:        limit,
			where:        query.Where(),
			handler:      query.Handler(),
		}
		return e
	}
}

func EdgeModWhere(where WhereFunc) EdgeQueryModFunc {
	return func(query EdgeQuery) EdgeQuery {
		e := &edgeQuery{
			fromType:     query.FromType(),
			fromKey:      query.FromKey(),
			relationship: query.Relationship(),
			toType:       query.ToType(),
			toKey:        query.ToKey(),
			where:        where,
			limit:        query.Limit(),
			handler:      query.Handler(),
		}
		return e
	}
}

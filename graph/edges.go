package graph

type EdgeStore struct {
	edges     map[string]map[string]Values
	edgesTo   map[string][]string
	edgesFrom map[string][]string
}

func newEdgeStore() *EdgeStore {
	return &EdgeStore{
		edges:     map[string]map[string]Values{},
		edgesTo:   map[string][]string{},
		edgesFrom: map[string][]string{},
	}
}

func (n *EdgeStore) Len(edgeType string) int {
	if c, ok := n.edges[edgeType]; ok {
		return len(c)
	}
	return 0
}

func (n *EdgeStore) Types() []string {
	var edgeTypes []string
	for k, _ := range n.edges {
		edgeTypes = append(edgeTypes, k)
	}
	return edgeTypes
}

func (n *EdgeStore) All() ValueSet {
	var edges ValueSet
	n.Range(Any, func(edge Values) bool {
		edges = append(edges, edge)
		return true
	})
	return edges
}

func (n *EdgeStore) Get(path string) (Values, bool) {
	typ, id := SplitPath(path)
	if c, ok := n.edges[typ]; ok {
		node := c[id]
		return c[id], node != nil
	}
	return nil, false
}

func (n *EdgeStore) Set(value Values) Values {
	if _, ok := n.edges[value.GetType()]; !ok {
		n.edges[value.GetType()] = map[string]Values{}
	}

	n.edges[value.GetType()][value.GetID()] = value

	n.edgesFrom[value.GetString(FromKey)] = append(n.edgesFrom[value.GetString(FromKey)], value.PathString())
	n.edgesTo[value.GetString(ToKey)] = append(n.edgesTo[value.GetString(ToKey)], value.PathString())

	if value.GetBool(MutualKey) {
		n.edgesTo[value.GetString(FromKey)] = append(n.edgesTo[value.GetString(FromKey)], value.PathString())
		n.edgesFrom[value.GetString(ToKey)] = append(n.edgesFrom[value.GetString(ToKey)], value.PathString())
	}
	return value
}

func (n *EdgeStore) Range(edgeType string, f func(edge Values) bool) {
	if edgeType == Any {
		for _, c := range n.edges {
			for _, v := range c {
				f(v)
			}
		}
	} else {
		if c, ok := n.edges[edgeType]; ok {
			for _, v := range c {
				f(v)
			}
		}
	}
}

func (n *EdgeStore) Delete(path string) {
	xtype, xid := SplitPath(path)
	edge, ok := n.Get(path)
	if !ok {
		return
	}
	n.edgesFrom[edge.GetString(FromKey)] = removeEdge(path, n.edgesFrom[edge.GetString(FromKey)])
	n.edgesTo[edge.GetString(FromKey)] = removeEdge(path, n.edgesTo[edge.GetString(FromKey)])
	n.edgesFrom[edge.GetString(ToKey)] = removeEdge(path, n.edgesFrom[edge.GetString(ToKey)])
	n.edgesTo[edge.GetString(ToKey)] = removeEdge(path, n.edgesTo[edge.GetString(ToKey)])
	delete(n.edges[xtype], xid)
}

func (n *EdgeStore) Exists(path string) bool {
	_, ok := n.Get(path)
	return ok
}

func (e *EdgeStore) RangeFrom(path string, fn func(e Values) bool) {
	for _, edge := range e.edgesFrom[path] {
		e, ok := e.Get(edge)
		if ok {
			if !fn(e) {
				break
			}
		}
	}
}

func (e *EdgeStore) RangeTo(path string, fn func(e Values) bool) {
	for _, edge := range e.edgesTo[path] {
		e, ok := e.Get(edge)
		if ok {
			if !fn(e) {
				break
			}
		}
	}
}

func (e *EdgeStore) RangeFilterFrom(path string, filter *Filter) ValueSet {
	var edges ValueSet
	e.RangeFrom(path, func(e Values) bool {
		if e.GetType() != filter.Type {
			return true
		}
		pass, _ := BooleanExpression(filter.Expressions, e)
		if pass {
			edges = append(edges, e)
		}
		return len(edges) < int(filter.Limit)
	})
	return edges
}

func (e *EdgeStore) RangeFilterTo(path string, filter *Filter) ValueSet {
	var edges ValueSet
	e.RangeTo(path, func(e Values) bool {
		if e.GetType() != filter.Type {
			return true
		}
		pass, _ := BooleanExpression(filter.Expressions, e)
		if pass {
			edges = append(edges, e)
		}
		return len(edges) < int(filter.Limit)
	})
	return edges
}

func (n *EdgeStore) Filter(edgeType string, filter func(edge Values) bool) ValueSet {
	var filtered ValueSet
	n.Range(edgeType, func(node Values) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	return filtered
}

func (n *EdgeStore) SetAll(edges ValueSet) {
	for _, edge := range edges {
		n.Set(edge)
	}
}

func (n *EdgeStore) DeleteAll(edges ValueSet) {
	for _, edge := range edges {
		n.Delete(edge.PathString())
	}
}

func (n *EdgeStore) Clear(edgeType string) {
	if cache, ok := n.edges[edgeType]; ok {
		for _, v := range cache {
			n.Delete(v.PathString())
		}
	}
}

func (n *EdgeStore) Close() {
	for _, edgeType := range n.Types() {
		n.Clear(edgeType)
	}
}

func (e *EdgeStore) Patch(updatedAt int64, value Values) Values {
	if _, ok := e.edges[value.GetType()]; !ok {
		return nil
	}
	for k, v := range value {
		e.edges[value.GetType()][value.GetID()][k] = v
	}
	e.edges[value.GetType()][value.GetID()][UpdatedAtKey] = updatedAt
	return e.edges[value.GetType()][value.GetID()]
}

func (e *EdgeStore) FilterSearch(filter *Filter) (ValueSet, error) {
	var edges ValueSet
	var err error
	var pass bool
	e.Range(filter.Type, func(edge Values) bool {
		pass, err = BooleanExpression(filter.Expressions, edge)
		if err != nil {
			return false
		}
		if pass {
			edges = append(edges, edge)
		}
		return len(edges) < int(filter.Limit)
	})
	return edges, err
}

func removeEdge(path string, paths []string) []string {
	var newPaths []string
	for _, p := range paths {
		if p != path {
			newPaths = append(newPaths, p)
		}
	}
	return newPaths
}

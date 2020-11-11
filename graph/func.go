package graph

import apipb "github.com/autom8ter/graphik/api"

type EdgeFunc func(e *apipb.Edge) bool

func (e EdgeFunc) ascendFrom(g *Graph, visited map[*apipb.Path]struct{}) EdgeFunc {
	return func(edge *apipb.Edge) bool {
		if _, ok := visited[edge.GetPath()]; ok {
			return true
		}
		visited[edge.GetPath()] = struct{}{}
		connected, ok := g.GetNode(edge.To)
		if !ok {
			return true
		}
		if e(edge) {
			g.RangeFrom(connected.Path, 0, func(edge *apipb.Edge) bool {
				if _, ok := visited[edge.GetPath()]; ok {
					return true
				}
				return e(edge)
			})
		}
		return true
	}
}

func (e EdgeFunc) ascendTo(g *Graph, visited map[*apipb.Path]struct{}) EdgeFunc {
	return func(edge *apipb.Edge) bool {
		if _, ok := visited[edge.GetPath()]; ok {
			return true
		}
		visited[edge.GetPath()] = struct{}{}
		connected, ok := g.GetNode(edge.From)
		if !ok {
			return true
		}
		if e(edge) {
			g.RangeTo(connected.Path, 0, func(edge *apipb.Edge) bool {
				if _, ok := visited[edge.GetPath()]; ok {
					return true
				}
				return e(edge)
			})
		}
		return true
	}
}

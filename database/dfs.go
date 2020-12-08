package database

import (
	"context"
	apipb "github.com/autom8ter/graphik/gen/go"
	"github.com/autom8ter/graphik/generic/stack"
	"github.com/autom8ter/graphik/logger"
	"github.com/google/cel-go/cel"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"strings"
)

// depthFirst implements stateful depth-first graph traversal.
type depthFirst struct {
	g             *Graph
	stack         *stack.Stack
	visited       map[string]struct{}
	traversals    *apipb.Traversals
	filter        *apipb.TFilter
	traversalPath []*apipb.Ref
}

func (g *Graph) newDepthFirst(filter *apipb.TFilter) *depthFirst {
	return &depthFirst{
		g:             g,
		filter:        filter,
		stack:         stack.New(),
		visited:       map[string]struct{}{},
		traversals:    &apipb.Traversals{},
		traversalPath: []*apipb.Ref{},
	}
}

func (d *depthFirst) Walk(ctx context.Context, tx *bbolt.Tx) error {
	//defer func() {
	//	d.docs.Sort(d.filter.Sort)
	//}()
	var (
		docProgram        *cel.Program
		connectionProgram *cel.Program
	)
	if d.filter.GetConnectionExpression() != "" {
		program, err := d.g.vm.Connection().Program(d.filter.GetConnectionExpression())
		if err != nil {
			return err
		}
		connectionProgram = &program
	}
	if d.filter.GetDocExpression() != "" {
		program, err := d.g.vm.Doc().Program(d.filter.GetDocExpression())
		if err != nil {
			return err
		}
		docProgram = &program
	}
	if _, ok := d.visited[d.filter.Root.String()]; !ok {
		doc, err := d.g.getDoc(ctx, tx, d.filter.Root)
		if err != nil {
			return err
		}
		d.stack.Push(doc)
		if docProgram == nil {
			d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
				Doc:           doc,
				TraversalPath: d.traversalPath,
			})
		} else {
			res, err := d.g.vm.Doc().Eval(doc, *docProgram)
			if err != nil {
				return err
			}
			if res {
				d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
					Doc:           doc,
					TraversalPath: d.traversalPath,
				})
			}
		}
		d.visited[d.filter.Root.String()] = struct{}{}
	}
	for d.stack.Len() > 0 && len(d.traversals.GetTraversals()) < int(d.filter.Limit) {
		if err := ctx.Err(); err != nil {
			return nil
		}
		popped := d.stack.Pop().(*apipb.Doc)
		d.traversalPath = append(d.traversalPath, popped.GetRef())
		if d.filter.GetReverse() {
			if err := d.rangeTo(ctx, tx, popped, connectionProgram, docProgram); err != nil {
				return err
			}
		} else {
			if err := d.rangeFrom(ctx, tx, popped, connectionProgram, docProgram); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *depthFirst) rangeFrom(ctx context.Context, tx *bbolt.Tx, popped *apipb.Doc, connectionProgram, docProgram *cel.Program) error {
	if err := d.g.rangeFrom(ctx, tx, popped.GetRef(), func(e *apipb.Connection) bool {
		if connectionProgram != nil {
			res, err := d.g.vm.Connection().Eval(e, *connectionProgram)
			if err != nil {
				if !strings.Contains(err.Error(), "no such key") {
					logger.Error("dfs failure", zap.Error(err))
				}
				return true
			}
			if !res {
				return true
			}
		}
		if _, ok := d.visited[e.To.String()]; !ok {
			to, err := d.g.getDoc(ctx, tx, e.To)
			if err != nil {
				if !strings.Contains(err.Error(), "no such key") {
					logger.Error("dfs failure", zap.Error(err))
				}
				return true
			}
			if docProgram == nil {
				d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
					Doc:           to,
					TraversalPath: d.traversalPath,
				})
			} else {
				res, err := d.g.vm.Doc().Eval(to, *docProgram)
				if err != nil {
					if !strings.Contains(err.Error(), "no such key") {
						logger.Error("dfs failure", zap.Error(err))
					}
					return true
				}
				if res {
					d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
						Doc:           to,
						TraversalPath: d.traversalPath,
					})
				}
			}
			d.visited[to.Ref.String()] = struct{}{}
			d.stack.Push(to)
		}
		return len(d.traversals.GetTraversals()) < int(d.filter.Limit)
	}); err != nil {
		return err
	}
	return nil
}

func (d *depthFirst) rangeTo(ctx context.Context, tx *bbolt.Tx, popped *apipb.Doc, connectionProgram, docProgram *cel.Program) error {
	if err := d.g.rangeTo(ctx, tx, popped.GetRef(), func(e *apipb.Connection) bool {
		if connectionProgram != nil {
			res, err := d.g.vm.Connection().Eval(e, *connectionProgram)
			if err != nil {
				if !strings.Contains(err.Error(), "no such key") {
					logger.Error("dfs failure", zap.Error(err))
				}
				return true
			}
			if !res {
				return true
			}
		}
		if _, ok := d.visited[e.GetFrom().String()]; !ok {
			from, err := d.g.getDoc(ctx, tx, e.GetFrom())
			if err != nil {
				if !strings.Contains(err.Error(), "no such key") {
					logger.Error("dfs failure", zap.Error(err))
				}
				return true
			}
			if docProgram == nil {
				d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
					Doc:           from,
					TraversalPath: d.traversalPath,
				})
			} else {
				res, err := d.g.vm.Doc().Eval(from, *docProgram)
				if err != nil {
					if !strings.Contains(err.Error(), "no such key") {
						logger.Error("dfs failure", zap.Error(err))
					}
					return true
				}
				if res {
					d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
						Doc:           from,
						TraversalPath: d.traversalPath,
					})
				}
			}
			d.visited[from.Ref.String()] = struct{}{}
			d.stack.Push(from)
		}
		return len(d.traversals.GetTraversals()) < int(d.filter.Limit)
	}); err != nil {
		return err
	}
	return nil
}

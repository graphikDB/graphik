package database

import (
	"context"
	apipb "github.com/graphikDB/graphik/gen/grpc/go"
	"github.com/graphikDB/graphik/generic"
	"github.com/graphikDB/graphik/logger"
	"github.com/google/cel-go/cel"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"strings"
	"time"
)

// traversal implements stateful depth-first graph traversal.
type traversal struct {
	g                 *Graph
	stack             *generic.Stack
	queue             *generic.Queue
	visited           map[string]struct{}
	traversals        *apipb.Traversals
	filter            *apipb.TFilter
	traversalPath     []*apipb.Ref
	connectionProgram *cel.Program
	docProgram        *cel.Program
}

func (g *Graph) newTraversal(filter *apipb.TFilter) (*traversal, error) {
	t := &traversal{
		g:             g,
		filter:        filter,
		stack:         generic.NewStack(),
		queue:         generic.NewQueue(),
		visited:       map[string]struct{}{},
		traversals:    &apipb.Traversals{},
		traversalPath: []*apipb.Ref{},
	}
	if filter.GetConnectionExpression() != "" {
		program, err := g.vm.Connection().Program(filter.GetConnectionExpression())
		if err != nil {
			return nil, err
		}
		t.connectionProgram = &program
	}
	if filter.GetDocExpression() != "" {
		program, err := g.vm.Doc().Program(filter.GetDocExpression())
		if err != nil {
			return nil, err
		}
		t.docProgram = &program
	}
	return t, nil
}

func (d *traversal) Walk(ctx context.Context, tx *bbolt.Tx) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	switch d.filter.GetAlgorithm() {
	case apipb.Algorithm_DFS:
		return d.walkDFS(ctx, tx)
	case apipb.Algorithm_BFS:
		return d.walkBFS(ctx, tx)
	}
	return ErrUnsupportedAlgorithm
}

func (d *traversal) walkDFS(ctx context.Context, tx *bbolt.Tx) error {
	doc, err := d.g.getDoc(ctx, tx, d.filter.Root)
	if err != nil {
		return err
	}
	d.stack.Push(doc)
	if d.docProgram == nil {
		d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
			Doc:           doc,
			TraversalPath: d.traversalPath,
			Depth:         uint64(len(d.traversalPath)),
			Hops:          uint64(len(d.visited)),
		})
	} else {
		res, err := d.g.vm.Doc().Eval(doc, *d.docProgram)
		if err != nil {
			return err
		}
		if res {
			d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
				Doc:           doc,
				TraversalPath: d.traversalPath,
				Depth:         uint64(len(d.traversalPath)),
				Hops:          uint64(len(d.visited)),
			})
		}
	}
	d.visited[d.filter.Root.String()] = struct{}{}
	for d.stack.Len() > 0 && len(d.traversals.GetTraversals()) < int(d.filter.Limit) && len(d.visited) <= int(d.filter.MaxHops) {
		if err := ctx.Err(); err != nil {
			return nil
		}
		popped := d.stack.Pop().(*apipb.Doc)
		d.traversalPath = append(d.traversalPath, popped.GetRef())
		if d.filter.GetReverse() {
			if err := d.dfsTo(ctx, tx, popped, d.connectionProgram, d.docProgram); err != nil {
				return err
			}
		} else {
			if err := d.dfsFrom(ctx, tx, popped, d.connectionProgram, d.docProgram); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *traversal) dfsFrom(ctx context.Context, tx *bbolt.Tx, popped *apipb.Doc, connectionProgram, docProgram *cel.Program) error {
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
				if len(d.traversalPath) <= int(d.filter.MaxDepth) && len(d.visited) <= int(d.filter.MaxHops) {
					d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
						Doc:           to,
						TraversalPath: d.traversalPath,
						Depth:         uint64(len(d.traversalPath)),
						Hops:          uint64(len(d.visited)),
					})
				}
			} else {
				res, err := d.g.vm.Doc().Eval(to, *docProgram)
				if err != nil {
					if !strings.Contains(err.Error(), "no such key") {
						logger.Error("dfs failure", zap.Error(err))
					}
					return true
				}
				if res {
					if len(d.traversalPath) <= int(d.filter.MaxDepth) && len(d.visited) <= int(d.filter.MaxHops) {
						d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
							Doc:           to,
							TraversalPath: d.traversalPath,
							Depth:         uint64(len(d.traversalPath)),
							Hops:          uint64(len(d.visited)),
						})
					}
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

func (d *traversal) dfsTo(ctx context.Context, tx *bbolt.Tx, popped *apipb.Doc, connectionProgram, docProgram *cel.Program) error {
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
				if len(d.traversalPath) <= int(d.filter.MaxDepth) && len(d.visited) <= int(d.filter.MaxHops) {
					d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
						Doc:           from,
						TraversalPath: d.traversalPath,
						Depth:         uint64(len(d.traversalPath)),
						Hops:          uint64(len(d.visited)),
					})
				}
			} else {
				res, err := d.g.vm.Doc().Eval(from, *docProgram)
				if err != nil {
					if !strings.Contains(err.Error(), "no such key") {
						logger.Error("dfs failure", zap.Error(err))
					}
					return true
				}
				if res {
					if len(d.traversalPath) <= int(d.filter.MaxDepth) && len(d.visited) <= int(d.filter.MaxHops) {
						d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
							Doc:           from,
							TraversalPath: d.traversalPath,
							Depth:         uint64(len(d.traversalPath)),
							Hops:          uint64(len(d.visited)),
						})
					}
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

func (d *traversal) walkBFS(ctx context.Context, tx *bbolt.Tx) error {
	doc, err := d.g.getDoc(ctx, tx, d.filter.Root)
	if err != nil {
		return err
	}
	d.queue.Enqueue(doc)
	if d.docProgram == nil {
		if len(d.traversalPath) <= int(d.filter.MaxDepth) && len(d.visited) <= int(d.filter.MaxHops) {
			d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
				Doc:           doc,
				TraversalPath: d.traversalPath,
				Depth:         uint64(len(d.traversalPath)),
				Hops:          uint64(len(d.visited)),
			})
		}
	} else {
		res, err := d.g.vm.Doc().Eval(doc, *d.docProgram)
		if err != nil {
			return err
		}
		if res {
			if len(d.traversalPath) <= int(d.filter.MaxDepth) && len(d.visited) <= int(d.filter.MaxHops) {
				d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
					Doc:           doc,
					TraversalPath: d.traversalPath,
					Depth:         uint64(len(d.traversalPath)),
					Hops:          uint64(len(d.visited)),
				})
			}
		}
	}
	d.visited[d.filter.Root.String()] = struct{}{}
	for d.queue.Len() > 0 && len(d.traversals.GetTraversals()) < int(d.filter.Limit) && len(d.visited) <= int(d.filter.MaxHops) {
		if err := ctx.Err(); err != nil {
			return nil
		}
		dequeued := d.queue.Dequeue().(*apipb.Doc)
		d.traversalPath = append(d.traversalPath, dequeued.GetRef())
		if d.filter.GetReverse() {
			if err := d.bfsTo(ctx, tx, dequeued, d.connectionProgram, d.docProgram); err != nil {
				return err
			}
		} else {
			if err := d.bfsFrom(ctx, tx, dequeued, d.connectionProgram, d.docProgram); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *traversal) bfsTo(ctx context.Context, tx *bbolt.Tx, dequeued *apipb.Doc, connectionProgram, docProgram *cel.Program) error {
	if err := d.g.rangeTo(ctx, tx, dequeued.GetRef(), func(e *apipb.Connection) bool {
		if connectionProgram != nil {
			res, err := d.g.vm.Connection().Eval(e, *connectionProgram)
			if err != nil {
				if !strings.Contains(err.Error(), "no such key") {
					logger.Error("bfs failure", zap.Error(err))
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
					logger.Error("bfs failure", zap.Error(err))
				}
				return true
			}
			if docProgram == nil {
				if len(d.traversalPath) <= int(d.filter.MaxDepth) && len(d.visited) <= int(d.filter.MaxHops) {
					d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
						Doc:           from,
						TraversalPath: d.traversalPath,
						Depth:         uint64(len(d.traversalPath)),
						Hops:          uint64(len(d.visited)),
					})
				}
			} else {
				res, err := d.g.vm.Doc().Eval(from, *docProgram)
				if err != nil {
					if !strings.Contains(err.Error(), "no such key") {
						logger.Error("bfs failure", zap.Error(err))
					}
					return true
				}
				if res {
					if len(d.traversalPath) <= int(d.filter.MaxDepth) && len(d.visited) <= int(d.filter.MaxHops) {
						d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
							Doc:           from,
							TraversalPath: d.traversalPath,
							Depth:         uint64(len(d.traversalPath)),
							Hops:          uint64(len(d.visited)),
						})
					}
				}
			}
			d.visited[from.Ref.String()] = struct{}{}
			d.queue.Enqueue(from)
		}
		return len(d.traversals.GetTraversals()) < int(d.filter.Limit)
	}); err != nil {
		return err
	}
	return nil
}

func (d *traversal) bfsFrom(ctx context.Context, tx *bbolt.Tx, dequeue *apipb.Doc, connectionProgram, docProgram *cel.Program) error {
	if err := d.g.rangeFrom(ctx, tx, dequeue.GetRef(), func(e *apipb.Connection) bool {

		if connectionProgram != nil {
			res, err := d.g.vm.Connection().Eval(e, *connectionProgram)
			if err != nil {
				if !strings.Contains(err.Error(), "no such key") {
					logger.Error("bfs failure", zap.Error(err))
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
					logger.Error("bfs failure", zap.Error(err))
				}
				return true
			}
			if docProgram == nil {
				if len(d.traversalPath) <= int(d.filter.MaxDepth) && len(d.visited) <= int(d.filter.MaxHops) {
					d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
						Doc:           to,
						TraversalPath: d.traversalPath,
						Depth:         uint64(len(d.traversalPath)),
						Hops:          uint64(len(d.visited)),
					})
				}
			} else {
				res, err := d.g.vm.Doc().Eval(to, *docProgram)
				if err != nil {
					if !strings.Contains(err.Error(), "no such key") {
						logger.Error("bfs failure", zap.Error(err))
					}
					return true
				}
				if res {
					if len(d.traversalPath) <= int(d.filter.MaxDepth) && len(d.visited) <= int(d.filter.MaxHops) {
						d.traversals.Traversals = append(d.traversals.Traversals, &apipb.Traversal{
							Doc:           to,
							TraversalPath: d.traversalPath,
							Depth:         uint64(len(d.traversalPath)),
							Hops:          uint64(len(d.visited)),
						})
					}
				}
			}
			d.visited[to.Ref.String()] = struct{}{}
			d.queue.Enqueue(to)
		}
		return len(d.traversals.GetTraversals()) < int(d.filter.Limit)
	}); err != nil {
		return err
	}
	return nil
}

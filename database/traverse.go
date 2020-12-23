package database

import (
	"context"
	"github.com/graphikDB/generic"
	apipb "github.com/graphikDB/graphik/gen/grpc/go"
	"github.com/graphikDB/trigger"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"time"
)

// traversal implements stateful depth-first graph traversal.
type traversal struct {
	g                 *Graph
	stack             *generic.Stack
	queue             *generic.Queue
	visited           map[string]struct{}
	traversals        *apipb.Traversals
	filter            *apipb.TraverseFilter
	traversalPath     []*apipb.Ref
	connectionProgram *trigger.Decision
	docProgram        *trigger.Decision
}

func (g *Graph) newTraversal(filter *apipb.TraverseFilter) (*traversal, error) {
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
		decision, err := trigger.NewDecision(filter.GetConnectionExpression())
		if err != nil {
			return nil, err
		}
		t.connectionProgram = decision
	}
	if filter.GetDocExpression() != "" {
		decision, err := trigger.NewDecision(filter.GetDocExpression())
		if err != nil {
			return nil, err
		}
		t.docProgram = decision
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
		if err := d.docProgram.Eval(doc.AsMap()); err == nil {
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

func (d *traversal) dfsFrom(ctx context.Context, tx *bbolt.Tx, popped *apipb.Doc, connectionProgram, docProgram *trigger.Decision) error {
	if err := d.g.rangeFrom(ctx, tx, popped.GetRef(), func(e *apipb.Connection) bool {
		if connectionProgram != nil {
			if err := connectionProgram.Eval(e.AsMap()); err != nil {
				return true
			}
		}
		if _, ok := d.visited[e.GetTo().String()]; !ok {
			to, err := d.g.getDoc(ctx, tx, e.GetTo())
			if err != nil {
				if err == ErrNotFound {
					d.g.logger.Error("dfs getDoc failure(to)", zap.Error(err), zap.String("path", refString(e.GetTo())))
					return true
				}
				d.g.logger.Error("dfs getDoc failure(to)", zap.Error(err))
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
				if err := docProgram.Eval(to.AsMap()); err == nil {
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

func (d *traversal) dfsTo(ctx context.Context, tx *bbolt.Tx, popped *apipb.Doc, connectionProgram, docProgram *trigger.Decision) error {
	if err := d.g.rangeTo(ctx, tx, popped.GetRef(), func(e *apipb.Connection) bool {
		if connectionProgram != nil {
			if err := connectionProgram.Eval(e.AsMap()); err != nil {
				return true
			}
		}
		if _, ok := d.visited[e.GetFrom().String()]; !ok {
			from, err := d.g.getDoc(ctx, tx, e.GetFrom())
			if err != nil {
				if err == ErrNotFound {
					d.g.logger.Error("dfs getDoc failure(from)", zap.Error(err), zap.String("path", refString(e.GetFrom())))
					return true
				}
				d.g.logger.Error("dfs getDoc failure(from)", zap.Error(err))
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
				if err := docProgram.Eval(from.AsMap()); err == nil {
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
		if err := d.docProgram.Eval(doc.AsMap()); err == nil {
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

func (d *traversal) bfsTo(ctx context.Context, tx *bbolt.Tx, dequeued *apipb.Doc, connectionProgram, docProgram *trigger.Decision) error {
	if err := d.g.rangeTo(ctx, tx, dequeued.GetRef(), func(e *apipb.Connection) bool {
		if connectionProgram != nil {
			if err := connectionProgram.Eval(e.AsMap()); err != nil {
				return true
			}
		}
		if _, ok := d.visited[e.GetFrom().String()]; !ok {
			from, err := d.g.getDoc(ctx, tx, e.GetFrom())
			if err != nil {
				if err == ErrNotFound {
					d.g.logger.Error("bfs getDoc failure(from)", zap.Error(err), zap.String("path", refString(e.GetFrom())))
					return true
				}
				d.g.logger.Error("bfs getDoc failure(from)", zap.Error(err))
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
				if err := docProgram.Eval(from.AsMap()); err == nil {
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

func (d *traversal) bfsFrom(ctx context.Context, tx *bbolt.Tx, dequeue *apipb.Doc, connectionProgram, docProgram *trigger.Decision) error {
	if err := d.g.rangeFrom(ctx, tx, dequeue.GetRef(), func(e *apipb.Connection) bool {

		if connectionProgram != nil {
			if err := connectionProgram.Eval(e.AsMap()); err != nil {
				return true
			}
		}
		if _, ok := d.visited[e.To.String()]; !ok {
			to, err := d.g.getDoc(ctx, tx, e.GetTo())
			if err != nil {
				if err == ErrNotFound {
					d.g.logger.Error("bfs getDoc failure(to)", zap.Error(err), zap.String("path", refString(e.GetTo())))
					return true
				}
				d.g.logger.Error("bfs getDoc failure(to)", zap.Error(err))
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
				if err := docProgram.Eval(to.AsMap()); err == nil {
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

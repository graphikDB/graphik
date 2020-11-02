package store

import (
	"encoding/json"
	"fmt"
	"github.com/autom8ter/graphik/generic"
	"github.com/autom8ter/graphik/graph/model"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/machine"
	"github.com/hashicorp/raft"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func (f *Store) Apply(log *raft.Log) interface{} {
	var c model.Command
	if err := json.Unmarshal(log.Data, &c); err != nil {
		return fmt.Errorf("failed to unmarshal command: %s", err.Error())
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	switch c.Op {
	case model.OpCreateNode:
		var val model.NodeConstructor
		if err := f.decode(c.Value, &val); err != nil {
			return errors.Wrap(err, "failed to decode node constructor")
		}
		n := f.nodes.Set(&model.Node{
			Path:       val.Path,
			Attributes: val.Attributes,
			CreatedAt:  c.Timestamp,
			UpdatedAt:  c.Timestamp,
		})
		channel := fmt.Sprintf("nodes.%s", n.Path.Type)
		f.opts.machine.Go(func(routine machine.Routine) {
			if err := routine.Publish(channel, n); err != nil {
				logger.Error("failed to publish message", zap.String("channel", channel))
			}
		})
		return n
	case model.OpPatchNode:
		var val model.Patch
		if err := f.decode(c.Value, &val); err != nil {
			return errors.Wrap(err, "failed to decode node patch")
		}
		if !f.nodes.Exists(val.Path) {
			return errors.Errorf("node %s does not exist", val.Path.String())
		}
		n := f.nodes.Patch(c.Timestamp, &val)
		channel := fmt.Sprintf("nodes.%s", n.Path.Type)
		f.opts.machine.Go(func(routine machine.Routine) {
			if err := routine.Publish(channel, n); err != nil {
				logger.Error("failed to publish message", zap.String("channel", channel))
			}
		})
		return n
	case model.OpDeleteNode:
		var val model.Path
		if err := f.decode(c.Value, &val); err != nil {
			return errors.Wrap(err, "failed to decode foreign key")
		}
		deleted := 0
		if val.ID == generic.Any {
			f.nodes.Range(val.Type, func(node *model.Node) bool {
				if f.nodes.Delete(node.Path) {
					deleted += 1
				}
				return true
			})
		} else {
			if f.nodes.Delete(val) {
				deleted += 1
			}
		}
		return &model.Counter{Count: deleted}
	case model.OpCreateEdge:
		var val model.EdgeConstructor
		if err := f.decode(c.Value, &val); err != nil {
			return errors.Wrap(err, "failed to decode edge constructor")
		}
		from, ok := f.nodes.Get(val.From)
		if !ok {
			return errors.Errorf("from node %s does not exist", (val.From.String()))
		}
		to, ok := f.nodes.Get(val.To)
		if !ok {
			return errors.Errorf("to node %s does not exist", val.To.String())
		}

		e := f.edges.Set(&model.Edge{
			Path:       val.Path,
			Mutual:     val.Mutual,
			Attributes: val.Attributes,
			From:       from.Path,
			To:         to.Path,
			CreatedAt:  c.Timestamp,
			UpdatedAt:  c.Timestamp,
		})
		channel := fmt.Sprintf("edges.%s", e.Path.Type)
		f.opts.machine.Go(func(routine machine.Routine) {
			if err := routine.Publish(channel, e); err != nil {
				logger.Error("failed to publish edge", zap.String("channel", channel))
			}
		})
		return e
	case model.OpPatchEdge:
		var val model.Patch
		if err := f.decode(c.Value, &val); err != nil {
			return errors.Wrap(err, "failed to decode edge patch")
		}
		if !f.edges.Exists(val.Path) {
			return errors.Errorf("edge %s does not exist", val.Path.String())
		}
		e := f.edges.Patch(c.Timestamp, &val)
		channel := fmt.Sprintf("edges.%s", e.Path.Type)
		f.opts.machine.Go(func(routine machine.Routine) {
			if err := routine.Publish(channel, e); err != nil {
				logger.Error("failed to publish edge", zap.String("channel", channel))
			}
		})
		return e

	case model.OpDeleteEdge:
		var val model.Path
		if err := f.decode(c.Value, &val); err != nil {
			return errors.Wrap(err, "failed to decode path")
		}
		if f.edges.Exists(val) {
			f.edges.Delete(val)
			return &model.Counter{Count: 1}
		}
		return &model.Counter{Count: 0}
	default:
		return fmt.Errorf("unsupported command: %v", c.Op)
	}
}

func (f *Store) Snapshot() (raft.FSMSnapshot, error) {
	return f, nil
}

func (f *Store) Restore(closer io.ReadCloser) error {
	export := &model.Export{}
	if err := json.NewDecoder(closer).Decode(export); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nodes.SetAll(export.Nodes...)
	f.edges.SetAll(export.Edges...)
	return nil
}

func (f *Store) Persist(sink raft.SnapshotSink) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if err := json.NewEncoder(sink).Encode(&model.Export{
		Nodes: f.nodes.All(),
		Edges: f.edges.All(),
	}); err != nil {
		return err
	}
	if err := sink.Close(); err != nil {
		sink.Cancel()
		return err
	}
	return nil
}
func (s *Store) Export() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		defer s.mu.RUnlock()
		json.NewEncoder(w).Encode(&model.Export{
			Nodes: s.nodes.All(),
			Edges: s.edges.All(),
		})
	}
}

func (s *Store) Import() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		export := &model.Export{}
		if err := json.NewDecoder(r.Body).Decode(export); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		s.nodes.SetAll(export.Nodes...)
		s.edges.SetAll(export.Edges...)
	}
}

func (f *Store) Release() {}

func (f *Store) decode(input, output interface{}) error {
	return mapstructure.Decode(input, output)
}

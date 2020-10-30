package store

import (
	"encoding/json"
	"fmt"
	"github.com/autom8ter/graphik/command"
	"github.com/autom8ter/graphik/graph/model"
	"github.com/hashicorp/raft"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"io"
)

func (f *Store) Apply(log *raft.Log) interface{} {
	var c command.Command
	if err := json.Unmarshal(log.Data, &c); err != nil {
		return fmt.Errorf("failed to unmarshal command: %s", err.Error())
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	switch c.Op {
	case command.CREATE_NODE:
		var val model.NodeConstructor
		if err := f.decode(c.Val, &val); err != nil {
			return errors.Wrap(err, "failed to decode node constructor")
		}
		return f.nodes.Set(&model.Node{
			Type:       val.Type,
			Attributes: val.Attributes,
			CreatedAt: c.Timestamp,
		})
	case command.PATCH_NODE:
		var val model.Patch
		if err := f.decode(c.Val, &val); err != nil {
			return errors.Wrap(err, "failed to decode node patch")
		}
		if !f.nodes.Exists(model.ForeignKey{
			ID:   val.ID,
			Type: val.Type,
		}) {
			return errors.Errorf("node %s.%s does not exist", val.Type, val.ID)
		}
		return f.nodes.Patch(c.Timestamp, &val)
	case command.DELETE_NODE:
		var val model.ForeignKey
		if err := f.decode(c.Val, &val); err != nil {
			return errors.Wrap(err, "failed to decode foreign key")
		}
		node, ok := f.nodes.Get(val)
		if !ok {
			return &model.Counter{Count: 0}
		}
		f.edges.RangeFrom(node, func(e *model.Edge) bool {
			f.edges.Delete(model.ForeignKey{
				ID:   e.ID,
				Type: e.Type,
			})
			return true
		})
		f.edges.RangeTo(node, func(e *model.Edge) bool {
			f.edges.Delete(model.ForeignKey{
				ID:   e.ID,
				Type: e.Type,
			})
			return true
		})
		f.nodes.Delete(model.ForeignKey{
			ID:   node.ID,
			Type: node.Type,
		})
		return &model.Counter{Count: 1}
	case command.CREATE_EDGE:
		var val model.EdgeConstructor
		if err := f.decode(c.Val, &val); err != nil {
			return errors.Wrap(err, "failed to decode edge constructor")
		}
		from, ok := f.nodes.Get(*val.From)
		if !ok {
			return errors.Errorf("from node %s.%s does not exist", from.Type, from.ID)
		}
		to, ok := f.nodes.Get(*val.To)
		if !ok {
			return errors.Errorf("to node %s.%s does not exist", to.Type, to.ID)
		}
		return f.edges.Set(&model.Edge{
			Type:       val.Type,
			Attributes: val.Attributes,
			From:       from,
			To:         to,
			CreatedAt:  c.Timestamp,
			Mutual: val.Mutual,
		})
	case command.PATCH_EDGE:
		var val model.Patch
		if err := f.decode(c.Val, &val); err != nil {
			return errors.Wrap(err, "failed to decode edge patch")
		}
		if !f.edges.Exists(model.ForeignKey{
			ID:   val.ID,
			Type: val.Type,
		}) {
			return errors.Errorf("edge %s.%s does not exist", val.Type, val.ID)
		}
		return f.edges.Patch(c.Timestamp, &val)

	case command.DELETE_EDGE:
		var val model.ForeignKey
		if err := f.decode(c.Val, &val); err != nil {
			return errors.Wrap(err, "failed to decode foreign key")
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
	if err := json.NewEncoder(sink).Encode(&model.Export{
		Nodes: f.nodes.All(),
		Edges: f.edges.All(),
	}); err != nil {
		return err
	}
	f.mu.RUnlock()
	if err := sink.Close(); err != nil {
		sink.Cancel()
		return err
	}
	return nil
}

func (f *Store) Release() {}

func (f *Store) decode(input, output interface{}) error {
	return mapstructure.Decode(input, output)
}
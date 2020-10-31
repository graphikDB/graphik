package store

import (
	"encoding/json"
	"fmt"
	"github.com/autom8ter/graphik/command"
	"github.com/autom8ter/graphik/generic"
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
			Path:       val.Path,
			Attributes: val.Attributes,
			CreatedAt:  c.Timestamp,
			UpdatedAt:  c.Timestamp,
		})
	case command.PATCH_NODE:
		var val model.Patch
		if err := f.decode(c.Val, &val); err != nil {
			return errors.Wrap(err, "failed to decode node patch")
		}
		if !f.nodes.Exists(val.Path) {
			return errors.Errorf("node %s does not exist", val.Path.String())
		}
		return f.nodes.Patch(c.Timestamp, &val)
	case command.DELETE_NODE:
		var val model.Path
		if err := f.decode(c.Val, &val); err != nil {
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
	case command.CREATE_EDGE:
		var val model.EdgeConstructor
		if err := f.decode(c.Val, &val); err != nil {
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
		return f.edges.Set(&model.Edge{
			Path:       val.Path,
			Mutual:  	val.Mutual,
			Attributes: val.Attributes,
			From:       from.Path,
			To:         to.Path,
			CreatedAt:  c.Timestamp,
			UpdatedAt:  c.Timestamp,
		})
	case command.PATCH_EDGE:
		var val model.Patch
		if err := f.decode(c.Val, &val); err != nil {
			return errors.Wrap(err, "failed to decode edge patch")
		}
		if !f.edges.Exists(val.Path) {
			return errors.Errorf("edge %s does not exist", val.Path.String())
		}
		return f.edges.Patch(c.Timestamp, &val)

	case command.DELETE_EDGE:
		var val model.Path
		if err := f.decode(c.Val, &val); err != nil {
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

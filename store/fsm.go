package store

import (
	"encoding/json"
	"fmt"
	"github.com/autom8ter/dagger"
	"github.com/autom8ter/graphik/command"
	"github.com/autom8ter/graphik/graph/model"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"io"
)

func (f *Store) Apply(log *raft.Log) interface{} {
	var c command.Command
	if err := json.Unmarshal(log.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}
	switch c.Op {
	case command.CREATE_NODE:
		val := c.Val.(*model.NodeConstructor)
		return f.nodes.Set(&model.Node{
			Type:       val.Type,
			Attributes: val.Attributes,
		})
	case command.PATCH_NODE:
		return f.nodes.Patch(c.Val.(*model.Patch))
	case command.DELETE_NODE:
		val := c.Val.(*model.ForeignKey)
		if f.nodes.Exists(val) {
			f.nodes.Delete(val)
			return &model.Counter{Count: 1}
		}
		return &model.Counter{Count: 0}
	case command.CREATE_EDGE:
		val := c.Val.(*model.EdgeConstructor)
		from, ok := f.nodes.Get(val.From)
		if !ok {
			return errors.Errorf("from node %s.%s does not exist", from.Type, from.ID)
		}
		to, ok := f.nodes.Get(val.To)
		if !ok {
			return errors.Errorf("to node %s.%s does not exist", to.Type, to.ID)
		}
		return f.edges.Set(&model.Edge{
			Type:       val.Type,
			Attributes: val.Attributes,
			From:       from,
			To:         to,
		})
	case command.PATCH_EDGE:
		return f.edges.Patch(c.Val.(*model.Patch))

	case command.DELETE_EDGE:
		key := c.Val.(*model.ForeignKey)
		if f.edges.Exists(key) {
			f.edges.Delete(key)
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
	return dagger.ImportJSON(closer)
}

func (f *Store) Persist(sink raft.SnapshotSink) error {
	if err := dagger.ExportJSON(sink); err != nil {
		return err
	}
	if err := sink.Close(); err != nil {
		sink.Cancel()
		return err
	}
	return nil
}

func (f *Store) Release() {}

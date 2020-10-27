package fsm

import (
	"encoding/json"
	"fmt"
	"github.com/autom8ter/dagger"
	"github.com/autom8ter/graphik/command"
	"github.com/autom8ter/graphik/graph/model"
	"github.com/hashicorp/raft"
	"io"
)

func New() *Store {
	return &Store{}
}

type Store struct{}

func (f *Store) Apply(log *raft.Log) interface{} {
	var c command.Command
	if err := json.Unmarshal(log.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}
	switch c.Op {
	case command.SET_NODE:
		input := c.Val.(*model.NodeInput)
		if input.ID != nil {
			node := dagger.NewNode(input.Type, *input.ID, input.Attributes)
			return &model.Node{
				ID:         node.ID(),
				Type:       node.Type(),
				Attributes: node.Raw(),
				Edges:      nil,
			}
		} else {
			node := dagger.NewNode(input.Type, "", input.Attributes)
			return &model.Node{
				ID:         node.ID(),
				Type:       node.Type(),
				Attributes: node.Raw(),
				Edges:      nil,
			}
		}

	case command.SET_EDGE:
		input := c.Val.(*model.EdgeInput)
		from, ok := dagger.GetNode(dagger.ForeignKey(input.From.Type, input.From.ID))
		if !ok {
			return fmt.Errorf("%s.%s does not exist", input.From.Type, input.From.ID)
		}
		to, ok := dagger.GetNode(dagger.ForeignKey(input.To.Type, input.To.ID))
		if !ok {
			return fmt.Errorf("%s.%s does not exist", input.To.Type, input.To.ID)
		}
		edge, err := dagger.NewEdge(input.Node.Type, nullString(input.Node.ID), input.Node.Attributes, from, to)
		if err != nil {
			return err
		}
		return &model.Edge{
			Node: &model.Node{
				ID:         edge.ID(),
				Type:       edge.Type(),
				Attributes: edge.Node().Raw(),
				Edges:      nil,
			},
			From: &model.Node{
				ID:         from.ID(),
				Type:       from.Type(),
				Attributes: from.Raw(),
				Edges:      nil,
			},
			To: &model.Node{
				ID:         to.ID(),
				Type:       to.Type(),
				Attributes: to.Raw(),
				Edges:      nil,
			},
		}
	case command.DELETE_NODE:
		input := c.Val.(*model.ForeignKey)
		dagger.DelNode(dagger.ForeignKey(input.Type, input.ID))
		return nil
	case command.DELETE_EDGE:
		input := c.Val.(*model.ForeignKey)
		dagger.DelEdge(dagger.ForeignKey(input.Type, input.ID))
		return nil
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

func nullString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

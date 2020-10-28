package fsm

import (
	"encoding/json"
	"fmt"
	"github.com/autom8ter/dagger"
	"github.com/autom8ter/dagger/primitive"
	"github.com/autom8ter/graphik/command"
	"github.com/autom8ter/graphik/graph/model"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
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
	case command.CREATE_NODE:
		input := c.Val.(map[string]interface{})
		node := dagger.NewNode(input[primitive.TYPE_KEY].(string), input[primitive.ID_KEY].(string), input)
		return &model.Node{
			Attributes: node.Raw(),
			Edges:      nil,
		}
	case command.SET_NODE:
		input := c.Val.(map[string]interface{})
		key := &dagger.ForeignKey{
			XID:   input[primitive.ID_KEY].(string),
			XType: input[primitive.TYPE_KEY].(string),
		}
		node, ok := dagger.GetNode(key)
		if !ok {
			return errors.Errorf("node %s does not exist", key.Path())
		}
		node.Patch(input)
		return &model.Node{
			Attributes: node.Raw(),
			Edges:      nil,
		}
	case command.CREATE_EDGE:
		input := c.Val.(map[string]interface {})
		nodeInput := input["node"].(map[string]interface {})
		from, ok := dagger.GetNode(&dagger.ForeignKey{
			XID:   input["from"].(map[string]interface {})[primitive.ID_KEY].(string),
			XType: input["from"].(map[string]interface {})[primitive.TYPE_KEY].(string),
		})
		if !ok {
			return fmt.Errorf("FROM node does not exist")
		}
		to, ok := dagger.GetNode(&dagger.ForeignKey{
			XID:   input["to"].(map[string]interface {})[primitive.ID_KEY].(string),
			XType: input["to"].(map[string]interface {})[primitive.TYPE_KEY].(string),
		})
		if !ok {
			return fmt.Errorf("TO node does not exist")
		}
		edge, err := dagger.NewEdge(nodeInput[primitive.TYPE_KEY].(string), nodeInput[primitive.ID_KEY].(string), nodeInput, from, to)
		if err != nil {
			return err
		}
		return &model.Edge{
			Node: &model.Node{
				Attributes: edge.Node().Raw(),
				Edges:      nil,
			},
			From: &model.Node{
				Attributes: from.Raw(),
				Edges:      nil,
			},
			To: &model.Node{
				Attributes: to.Raw(),
				Edges:      nil,
			},
		}
	case command.SET_EDGE:
		input := c.Val.(model.EdgeInput)
		key := &dagger.ForeignKey{
			XID:   input.From.ID,
			XType: input.From.Type,
		}
		from, ok := dagger.GetNode(key)
		if !ok {
			return fmt.Errorf("%s does not exist", key.Path())
		}
		to, ok := dagger.GetNode(&dagger.ForeignKey{
			XID:   input.To.ID,
			XType: input.To.Type,
		})
		if !ok {
			return fmt.Errorf("%s.%s does not exist", input.To.Type, input.To.ID)
		}
		edge, err := dagger.NewEdge(input.Node[primitive.TYPE_KEY].(string), input.Node[primitive.ID_KEY].(string), input.Node, from, to)
		if err != nil {
			return err
		}
		return &model.Edge{
			Node: &model.Node{
				Attributes: edge.Node().Raw(),
				Edges:      nil,
			},
			From: &model.Node{
				Attributes: from.Raw(),
				Edges:      nil,
			},
			To: &model.Node{
				Attributes: to.Raw(),
				Edges:      nil,
			},
		}
	case command.DELETE_NODE:
		input := c.Val.(map[string]interface{})
		dagger.DelNode(&dagger.ForeignKey{
			XID:   input[primitive.ID_KEY].(string),
			XType: input[primitive.TYPE_KEY].(string),
		})
		return nil
	case command.DELETE_EDGE:
		input := c.Val.(map[string]interface{})
		dagger.DelEdge(&dagger.ForeignKey{
			XID:   input[primitive.ID_KEY].(string),
			XType: input[primitive.TYPE_KEY].(string),
		})
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

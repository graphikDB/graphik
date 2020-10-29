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
		node := dagger.NewNode(input)
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
		input := c.Val.(map[string]interface{})
		attributes := input["attributes"].(map[string]interface{})
		from, ok := dagger.GetNode(&dagger.ForeignKey{
			XID:   input["from"].(map[string]interface{})[primitive.ID_KEY].(string),
			XType: input["from"].(map[string]interface{})[primitive.TYPE_KEY].(string),
		})
		if !ok {
			return fmt.Errorf("FROM node does not exist")
		}
		to, ok := dagger.GetNode(&dagger.ForeignKey{
			XID:   input["to"].(map[string]interface{})[primitive.ID_KEY].(string),
			XType: input["to"].(map[string]interface{})[primitive.TYPE_KEY].(string),
		})
		if !ok {
			return fmt.Errorf("TO node does not exist")
		}
		if val, ok := attributes["_mutual"].(bool); ok {
			edge, err := dagger.NewEdge(attributes["_type"].(string), from, to, val)
			if err != nil {
				return err
			}
			edge.Patch(attributes)
			return &model.Edge{
				Attributes: edge.Node().Raw(),
				From: &model.Node{
					Attributes: from.Raw(),
					Edges:      nil,
				},
				To: &model.Node{
					Attributes: to.Raw(),
					Edges:      nil,
				},
			}
		}
		edge, err := dagger.NewEdge(attributes["_type"].(string), from, to, false)
		if err != nil {
			return err
		}
		edge.Patch(attributes)
		return &model.Edge{
			Attributes: edge.Node().Raw(),
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
		input := c.Val.(map[string]interface{})
		attributes := input["attributes"].(map[string]interface{})

		edge, ok := dagger.GetEdge(&dagger.ForeignKey{
			XID:   attributes[primitive.ID_KEY].(string),
			XType: attributes[primitive.TYPE_KEY].(string),
		})
		if !ok {
			return errors.Errorf("edge does not exist")
		}
		edge.Patch(attributes)
		return &model.Edge{
			Attributes: edge.Node().Raw(),
			From: &model.Node{
				Attributes: edge.From().Raw(),
				Edges:      nil,
			},
			To: &model.Node{
				Attributes: edge.To().Raw(),
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

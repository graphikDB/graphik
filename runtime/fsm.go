package runtime

import (
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/machine"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

func (f *Runtime) Apply(log *raft.Log) interface{} {
	var c apipb.Command
	if err := proto.Unmarshal(log.Data, &c); err != nil {
		return fmt.Errorf("failed to unmarshal command: %s", err.Error())
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	switch c.Op {
	case apipb.Op_CREATE_NODE:
		var val = &apipb.Node{}
		if err := ptypes.UnmarshalAny(c.Val, val); err != nil {
			return errors.Wrap(err, "failed to decode node")
		}
		n := f.nodes.Set(&apipb.Node{
			Path:       val.Path,
			Attributes: val.Attributes,
			CreatedAt:  c.Timestamp,
			UpdatedAt:  c.Timestamp,
		})
		channel := fmt.Sprintf("nodes.%s", n.Path.Type)
		f.machine.Go(func(routine machine.Routine) {
			if err := routine.Publish(channel, n); err != nil {
				logger.Error("failed to publish message", zap.String("channel", channel))
			}
		})
		return n
	case apipb.Op_PATCH_NODE:
		var val = &apipb.Patch{}
		if err := ptypes.UnmarshalAny(c.Val, val); err != nil {
			return errors.Wrap(err, "failed to decode node patch")
		}
		if !f.nodes.Exists(val.Path) {
			return errors.Errorf("node %s does not exist", val.Path.String())
		}
		n := f.nodes.Patch(c.Timestamp, val)
		channel := fmt.Sprintf("nodes.%s", n.Path.Type)
		f.machine.Go(func(routine machine.Routine) {
			if err := routine.Publish(channel, n); err != nil {
				logger.Error("failed to publish message", zap.String("channel", channel))
			}
		})
		return n
	case apipb.Op_DELETE_NODE:
		var val = &apipb.Path{}
		if err := ptypes.UnmarshalAny(c.Val, val); err != nil {
			return errors.Wrap(err, "failed to decode node path")
		}
		deleted := 0
		if val.ID == apipb.Keyword_ANY.String() {
			f.nodes.Range(val.Type, func(node *apipb.Node) bool {
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
		return &apipb.Counter{Count: int64(deleted)}
	case apipb.Op_CREATE_EDGE:
		var val = &apipb.Edge{}
		if err := ptypes.UnmarshalAny(c.Val, val); err != nil {
			return errors.Wrap(err, "failed to decode edge")
		}
		_, ok := f.nodes.Get(val.From)
		if !ok {
			return errors.Errorf("from node %s does not exist", (val.From.String()))
		}
		_, ok = f.nodes.Get(val.To)
		if !ok {
			return errors.Errorf("to node %s does not exist", val.To.String())
		}
		val.CreatedAt = c.Timestamp
		val.UpdatedAt = c.Timestamp
		e := f.edges.Set(val)
		channel := fmt.Sprintf("edges.%s", e.Path.Type)
		f.machine.Go(func(routine machine.Routine) {
			if err := routine.Publish(channel, e); err != nil {
				logger.Error("failed to publish edge", zap.String("channel", channel))
			}
		})
		return e
	case apipb.Op_PATCH_EDGE:
		var val = &apipb.Patch{}
		if err := ptypes.UnmarshalAny(c.Val, val); err != nil {
			return errors.Wrap(err, "failed to decode edge patch")
		}
		if !f.edges.Exists(val.Path) {
			return errors.Errorf("edge %s does not exist", val.Path.String())
		}
		e := f.edges.Patch(c.Timestamp, val)
		channel := fmt.Sprintf("edges.%s", e.Path.Type)
		f.machine.Go(func(routine machine.Routine) {
			if err := routine.Publish(channel, e); err != nil {
				logger.Error("failed to publish edge", zap.String("channel", channel))
			}
		})
		return e

	case apipb.Op_DELETE_EDGE:
		var val = &apipb.Path{}
		if err := ptypes.UnmarshalAny(c.Val, val); err != nil {
			return errors.Wrap(err, "failed to decode edge path")
		}
		if f.edges.Exists(val) {
			f.edges.Delete(val)
			return &apipb.Counter{Count: 1}
		}
		return &apipb.Counter{Count: 0}
	default:
		return fmt.Errorf("unsupported command: %v", c.Op)
	}
}

func (f *Runtime) Snapshot() (raft.FSMSnapshot, error) {
	return f, nil
}

func (f *Runtime) Restore(closer io.ReadCloser) error {
	export := &apipb.Export{}
	bits, err := ioutil.ReadAll(closer)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(bits, export); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nodes.SetAll(export.Nodes...)
	f.edges.SetAll(export.Edges...)
	return nil
}

func (f *Runtime) Persist(sink raft.SnapshotSink) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	export := &apipb.Export{
		Nodes: f.nodes.All(),
		Edges: f.edges.All(),
	}
	bits, err := proto.Marshal(export)
	if err != nil {
		return err
	}
	_, err = sink.Write(bits)
	if err != nil {
		return err
	}
	if err := sink.Close(); err != nil {
		sink.Cancel()
		return err
	}
	return nil
}

func (s *Runtime) Export() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		defer s.mu.RUnlock()
		export := &apipb.Export{
			Nodes: s.nodes.All(),
			Edges: s.edges.All(),
		}
		bits, err := proto.Marshal(export)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(bits)
	}
}

func (s *Runtime) Import() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		export := &apipb.Export{}
		bits, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := proto.Unmarshal(bits, export); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for _, n := range export.Nodes {
			any, _ := ptypes.MarshalAny(n)
			cmd := &apipb.Command{
				Op:  apipb.Op_CREATE_NODE,
				Val: any,
				Timestamp: &timestamp.Timestamp{
					Seconds: time.Now().Unix(),
				},
			}
			s.Apply(cmd.Log())
		}
		for _, e := range export.Edges {
			any, _ := ptypes.MarshalAny(e)
			cmd := apipb.Command{
				Op:  apipb.Op_CREATE_EDGE,
				Val: any,
				Timestamp: &timestamp.Timestamp{
					Seconds: time.Now().Unix(),
				},
			}
			s.Apply(cmd.Log())
		}
	}
}

func (f *Runtime) Release() {}

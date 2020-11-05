package runtime

import (
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
)

func (f *Runtime) Apply(log *raft.Log) interface{} {
	var c apipb.Command
	if err := proto.Unmarshal(log.Data, &c); err != nil {
		return fmt.Errorf("failed to unmarshal command: %s", err.Error())
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	switch c.Op {
	case apipb.Op_CREATE_NODES:
		var values = &apipb.Nodes{}
		if err := ptypes.UnmarshalAny(c.Val, values); err != nil {
			return errors.Wrap(err, "failed to decode nodes")
		}
		for _, val := range values.Nodes {
			f.nodes.Set(&apipb.Node{
				Path:       val.Path,
				Attributes: val.Attributes,
				CreatedAt:  c.Timestamp,
				UpdatedAt:  c.Timestamp,
			})
		}
		return values
	case apipb.Op_PATCH_NODES:
		var values = &apipb.Patches{}
		if err := ptypes.UnmarshalAny(c.Val, values); err != nil {
			return errors.Wrap(err, "failed to decode node patches")
		}
		var nodes = &apipb.Nodes{}
		for _, val := range values.Patches {
			if !f.nodes.Exists(val.Path) {
				return errors.Errorf("node %s does not exist", val.Path.String())
			}
			n := f.nodes.Patch(c.Timestamp, val)
			nodes.Nodes = append(nodes.Nodes, n)
		}
		return nodes
	case apipb.Op_DELETE_NODES:
		var values = &apipb.Paths{}
		if err := ptypes.UnmarshalAny(c.Val, values); err != nil {
			return errors.Wrap(err, "failed to decode node paths")
		}
		deleted := 0
		for _, val := range values.Paths {
			if f.nodes.Delete(val) {
				deleted += 1
			}
		}
		return &apipb.Counter{Count: int64(deleted)}
	case apipb.Op_CREATE_EDGES:
		var values = &apipb.Edges{}
		if err := ptypes.UnmarshalAny(c.Val, values); err != nil {
			return errors.Wrap(err, "failed to decode edge")
		}
		for _, val := range values.Edges {
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
			f.edges.Set(val)
		}
		return values
	case apipb.Op_PATCH_EDGES:
		var val = &apipb.Patches{}
		if err := ptypes.UnmarshalAny(c.Val, val); err != nil {
			return errors.Wrap(err, "failed to decode edge patch")
		}
		var edges = &apipb.Edges{}
		for _, val := range val.Patches {
			if !f.edges.Exists(val.Path) {
				return errors.Errorf("edge %s does not exist", val.Path.String())
			}
			edges.Edges = append(edges.Edges, f.edges.Patch(c.Timestamp, val))
		}
		return edges

	case apipb.Op_DELETE_EDGES:
		var values = &apipb.Paths{}
		if err := ptypes.UnmarshalAny(c.Val, values); err != nil {
			return errors.Wrap(err, "failed to decode edge path")
		}
		deleted := 0
		for _, val := range values.Paths {
			if f.edges.Exists(val) {
				f.edges.Delete(val)
				deleted += 1
			}
		}
		return &apipb.Counter{Count: int64(deleted)}
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
	f.nodes.SetAll(export.Nodes)
	f.edges.SetAll(export.Edges)
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
		_, err = s.CreateNodes(export.Nodes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err = s.CreateEdges(export.Edges)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func (f *Runtime) Release() {}

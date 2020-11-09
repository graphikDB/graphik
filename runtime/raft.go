package runtime

import (
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/logger"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
)

func (s *Runtime) JoinNode(nodeID, addr string) error {
	logger.Info("received join request for remote node",
		zap.String("node", nodeID),
		zap.String("address", addr),
	)
	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				logger.Info("node already member of cluster, ignoring join request",
					zap.String("node", nodeID),
					zap.String("address", addr),
				)
				return nil
			}

			future := s.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}
	f := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	logger.Info("node joined successfully",
		zap.String("node", nodeID),
		zap.String("address", addr),
	)
	return nil
}

func (f *Runtime) Apply(log *raft.Log) interface{} {
	var c apipb.Command
	if err := proto.Unmarshal(log.Data, &c); err != nil {
		return fmt.Errorf("failed to unmarshal command: %s", err.Error())
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	switch c.Op {
	case apipb.Op_SET_AUTH:
		var values = &apipb.AuthConfig{}
		if err := ptypes.UnmarshalAny(c.Val, values); err != nil {
			return errors.Wrap(err, "failed to decode jwks sources")
		}
		if err := f.auth.Override(values); err != nil {
			return errors.Wrap(err, "failed to override auth")
		}
		return values
	case apipb.Op_CREATE_NODES:
		var values = &apipb.Nodes{}
		if err := ptypes.UnmarshalAny(c.Val, values); err != nil {
			return errors.Wrap(err, "failed to decode nodes")
		}
		for _, val := range values.Nodes {
			f.graph.Nodes().Set(&apipb.Node{
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
			if !f.graph.Nodes().Exists(val.Path) {
				return errors.Errorf("node %s does not exist", val.Path)
			}
			n := f.graph.Nodes().Patch(c.Timestamp, val)
			nodes.Nodes = append(nodes.Nodes, n)
		}
		return nodes
	case apipb.Op_DELETE_NODES:
		var values = &apipb.Paths{}
		if err := ptypes.UnmarshalAny(c.Val, values); err != nil {
			return errors.Wrap(err, "failed to decode node paths")
		}
		deleted := 0
		for _, val := range values.Values {
			if f.graph.Nodes().Delete(val) {
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
			_, ok := f.graph.Nodes().Get(val.From)
			if !ok {
				return errors.Errorf("from node %s does not exist", (val.From))
			}
			_, ok = f.graph.Nodes().Get(val.To)
			if !ok {
				return errors.Errorf("to node %s does not exist", val.To)
			}
			val.CreatedAt = c.Timestamp
			val.UpdatedAt = c.Timestamp
			f.graph.Edges().Set(val)
		}
		return values
	case apipb.Op_PATCH_EDGES:
		var val = &apipb.Patches{}
		if err := ptypes.UnmarshalAny(c.Val, val); err != nil {
			return errors.Wrap(err, "failed to decode edge patch")
		}
		var edges = &apipb.Edges{}
		for _, val := range val.Patches {
			if !f.graph.Edges().Exists(val.Path) {
				return errors.Errorf("edge %s does not exist", val.Path)
			}
			edges.Edges = append(edges.Edges, f.graph.Edges().Patch(c.Timestamp, val))
		}
		return edges

	case apipb.Op_DELETE_EDGES:
		var values = &apipb.Paths{}
		if err := ptypes.UnmarshalAny(c.Val, values); err != nil {
			return errors.Wrap(err, "failed to decode edge path")
		}
		deleted := 0
		for _, val := range values.Values {
			if f.graph.Edges().Exists(val) {
				f.graph.Edges().Delete(val)
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
	f.graph.Nodes().SetAll(export.Nodes)
	f.graph.Edges().SetAll(export.Edges)
	return nil
}

func (f *Runtime) Persist(sink raft.SnapshotSink) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	export := &apipb.Export{
		Nodes: f.graph.Nodes().All(),
		Edges: f.graph.Edges().All(),
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
			Nodes: s.graph.Nodes().All(),
			Edges: s.graph.Edges().All(),
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

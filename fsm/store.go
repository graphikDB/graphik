package fsm

import (
	"encoding/json"
	"fmt"
	"github.com/autom8ter/dagger"
	"github.com/autom8ter/graphik/command"
	"github.com/autom8ter/graphik/generic"
	"github.com/hashicorp/raft"
	"io"
	"sync"
)

func New() *Store {
	return &Store{
		mu:    sync.RWMutex{},
		nodes: generic.NewNodes(),
		edges: generic.NewEdges(),
		close: sync.Once{},
	}
}

type Store struct {
	mu        sync.RWMutex
	nodes     *generic.Nodes
	edges     *generic.Edges
	close     sync.Once
}

func (f *Store) Apply(log *raft.Log) interface{} {
	var c command.Command
	if err := json.Unmarshal(log.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}
	switch c.Op {
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

package fsm

import (
	"encoding/gob"
	"errors"
	"github.com/hashicorp/raft"
	"io"
)

func init() {
	gob.Register(&raft.Log{})
}

type Snapshot struct {
	PersistFunc func(sink raft.SnapshotSink) error
	ReleaseFunc func()
}

func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
	return s.PersistFunc(sink)
}

func (s *Snapshot) Release() {
	s.ReleaseFunc()
}

type FSM struct {
	ApplyFunc    func(log *raft.Log) interface{}
	SnapshotFunc func() (*Snapshot, error)
	RestoreFunc  func(closer io.ReadCloser) error
}

func (f *FSM) Validate() error {
	if f.ApplyFunc == nil {
		return errors.New("raft: empty apply function")
	}
	if f.SnapshotFunc == nil {
		return errors.New("raft: empty snapshot function")
	}
	if f.RestoreFunc == nil {
		return errors.New("raft: empty restore function")
	}
	return nil
}

func (f *FSM) Apply(log *raft.Log) interface{} {
	return f.ApplyFunc(log)
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return f.SnapshotFunc()
}

func (f *FSM) Restore(closer io.ReadCloser) error {
	return f.RestoreFunc(closer)
}

package raft

import (
	"github.com/hashicorp/raft"
	"io"
)

type FSM struct {
	ApplyFunc    func(log *raft.Log) interface{}
	SnapshotFunc func() (raft.FSMSnapshot, error)
	RestoreFunc  func(closer io.ReadCloser) error
}

func (f FSM) Apply(log *raft.Log) interface{} {
	return f.ApplyFunc(log)
}

func (f FSM) Snapshot() (raft.FSMSnapshot, error) {
	return f.SnapshotFunc()
}

func (f FSM) Restore(closer io.ReadCloser) error {
	return f.RestoreFunc(closer)
}

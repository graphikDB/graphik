package raft

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/hashicorp/raft"
	"io"
)

func init() {
	gob.Register(&Command{})
}

type Command struct {
	Type int
	Data []byte
}

type FSM struct {
	ApplyFunc    func(command *Command) interface{}
	SnapshotFunc func() (raft.FSMSnapshot, error)
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

func (f FSM) Apply(log *raft.Log) interface{} {
	var c Command
	if err := gob.NewDecoder(bytes.NewBuffer(log.Data)).Decode(&c); err != nil {
		return err
	}
	return f.ApplyFunc(&c)
}

func (f FSM) Snapshot() (raft.FSMSnapshot, error) {
	return f.SnapshotFunc()
}

func (f FSM) Restore(closer io.ReadCloser) error {
	return f.RestoreFunc(closer)
}

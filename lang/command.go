package lang

import (
	"bytes"
	"encoding/gob"
	"github.com/hashicorp/raft"
)

type Op int32

const (
	Op_CREATE_NODES Op = 0
	Op_PATCH_NODES  Op = 1
	Op_DELETE_NODES Op = 2
	Op_CREATE_EDGES Op = 3
	Op_PATCH_EDGES  Op = 4
	Op_DELETE_EDGES Op = 5
	Op_SET_AUTH     Op = 6
)

type Command struct {
	Op        Op          `json:"op,omitempty"`
	Val       interface{} `json:"val,omitempty"`
	Timestamp int64       `json:"timestamp,omitempty"`
}

func (c *Command) Log() *raft.Log {
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(c); err != nil {
		panic(err)
	}
	return &raft.Log{
		Data: buf.Bytes(),
	}
}

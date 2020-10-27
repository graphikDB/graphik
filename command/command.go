package command

import (
	"encoding/json"
	"github.com/hashicorp/raft"
)

type Operation int

const (
	SET_NODE    Operation = 1
	DELETE_NODE Operation = 3
	SET_EDGE    Operation = 5
	DELETE_EDGE Operation = 6
)

type Command struct {
	Op  Operation
	Val interface{}
}

func (c *Command) Log() (*raft.Log, error) {
	bits, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return &raft.Log{
		Data: bits,
	}, nil
}

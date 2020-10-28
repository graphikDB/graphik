package command

import (
	"encoding/json"
	"github.com/hashicorp/raft"
)

type Op int

const (
	CREATE_NODE Op = 1
	SET_NODE    Op = 2
	DELETE_NODE Op = 3
	CREATE_EDGE Op = 4
	SET_EDGE    Op = 5
	DELETE_EDGE Op = 6
)

type Command struct {
	Op  Op
	Val interface{}
}

func (c *Command) Log() (raft.Log, error) {
	bits, err := json.Marshal(c)
	if err != nil {
		return raft.Log{}, err
	}
	return raft.Log{
		Data: bits,
	}, nil
}

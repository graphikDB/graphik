package command

import (
	"encoding/json"
	"github.com/hashicorp/raft"
	"time"
)

type Op int

const (
	CREATE_NODE Op = 1
	PATCH_NODE  Op = 2
	DELETE_NODE Op = 3
	CREATE_EDGE Op = 4
	PATCH_EDGE  Op = 5
	DELETE_EDGE Op = 6
)

type Command struct {
	Op  Op
	Val interface{}
	Timestamp time.Time
}

func (c *Command) Log() (raft.Log, error) {
	c.Timestamp = time.Now()
	bits, err := json.Marshal(c)
	if err != nil {
		return raft.Log{}, err
	}
	return raft.Log{
		Data: bits,
	}, nil
}

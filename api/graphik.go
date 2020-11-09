package apipb

import (
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	"sort"
)

func (c *Command) Log() *raft.Log {
	bits, _ := proto.Marshal(c)
	return &raft.Log{
		Data: bits,
	}
}

func (e *Edges) Sort() {
	sorter := Sortable{
		LenFunc: func() int {
			if e == nil {
				return 0
			}
			return len(e.Edges)
		},
		LessFunc: func(i, j int) bool {
			if e == nil {
				return false
			}
			return e.Edges[i].UpdatedAt.Nanos > e.Edges[j].UpdatedAt.Nanos
		},
		SwapFunc: func(i, j int) {
			if e == nil {
				return
			}
			e.Edges[i], e.Edges[j] = e.Edges[j], e.Edges[i]
		},
	}
	sorter.Sort()
}

func (n *Nodes) Sort() {
	sorter := Sortable{
		LenFunc: func() int {
			if n == nil {
				return 0
			}
			return len(n.Nodes)
		},
		LessFunc: func(i, j int) bool {
			if n == nil {
				return false
			}
			return n.Nodes[i].UpdatedAt.Nanos > n.Nodes[j].UpdatedAt.Nanos
		},
		SwapFunc: func(i, j int) {
			if n == nil {
				return
			}
			n.Nodes[i], n.Nodes[j] = n.Nodes[j], n.Nodes[i]
		},
	}
	sorter.Sort()
}

func (c *Config) SetDefaults() {
	if c.Grpc == nil {
		c.Grpc = &GRPCConfig{}
	}
	if c.Grpc.Bind == "" {
		c.Grpc.Bind = ":7820"
	}
	if c.Http == nil {
		c.Http = &HTTPConfig{}
	}
	if c.Http.Bind == "" {
		c.Http.Bind = ":7830"
	}
	if c.Raft == nil {
		c.Raft = &RaftConfig{}
	}
	if c.Raft.Bind == "" {
		c.Raft.Bind = "localhost:7840"
	}
	if c.Raft.StoragePath == "" {
		c.Raft.StoragePath = "/tmp/graphik"
	}
	if c.Raft.NodeId == "" {
		c.Raft.NodeId = "default"
	}
}

type Sortable struct {
	LenFunc  func() int
	LessFunc func(i, j int) bool
	SwapFunc func(i, j int)
}

func (s Sortable) Len() int {
	if s.LenFunc == nil {
		return 0
	}
	return s.LenFunc()
}

func (s Sortable) Less(i, j int) bool {
	if s.LessFunc == nil {
		return false
	}
	return s.LessFunc(i, j)
}

func (s Sortable) Swap(i, j int) {
	if s.SwapFunc == nil {
		return
	}
	s.SwapFunc(i, j)
}

func (s Sortable) Sort() {
	sort.Sort(s)
}

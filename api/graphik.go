package apipb

import (
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
)

func (n *Node) Path() *Path {
	return &Path{
		Type:                 n.Type,
		Id:                   n.Id,
	}
}

func (e *Edge) Path() *Path {
	return &Path{
		Type:                 e.Type,
		Id:                   e.Id,
	}
}

func (c *Command) Log() raft.Log {
	bits, _ := proto.Marshal(c)
	return raft.Log{
		Data: bits,
	}
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

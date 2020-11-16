package apipb

import (
	"github.com/autom8ter/graphik/sortable"
	"google.golang.org/protobuf/types/known/structpb"
)

func NewStruct(data map[string]interface{}) *structpb.Struct {
	x, _ := structpb.NewStruct(data)
	return x
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

	if c.Runtime == nil {
		c.Runtime = &RuntimeConfig{
			Auth:    &AuthConfig{},
			Trigger: &TriggerConfig{},
		}
	}
	if c.Runtime.Trigger == nil {
		c.Runtime.Trigger = &TriggerConfig{}
	}
	if c.Runtime.Auth == nil {
		c.Runtime.Auth = &AuthConfig{}
	}
	if c.Raft.Bind == "" {
		c.Raft.Bind = "localhost:7840"
	}
	if c.Raft.StoragePath == "" {
		c.Raft.StoragePath = "/tmp/graphik"
	}
	if c.Raft.NodeId == "" {
		c.Raft.NodeId = Keyword_DEFAULT.String()
	}
}

func (n *Nodes) Sort() {
	s := sortable.Sortable{
		LenFunc: func() int {
			return len(n.GetNodes())
		},
		LessFunc: func(i, j int) bool {
			return n.GetNodes()[i].UpdatedAt < n.GetNodes()[j].UpdatedAt
		},
		SwapFunc: func(i, j int) {
			n.GetNodes()[i], n.GetNodes()[j] = n.GetNodes()[j], n.GetNodes()[i]
		},
	}
	s.Sort()
}

func (e *Edges) Sort() {
	s := sortable.Sortable{
		LenFunc: func() int {
			return len(e.GetEdges())
		},
		LessFunc: func(i, j int) bool {
			return e.GetEdges()[i].UpdatedAt < e.GetEdges()[j].UpdatedAt
		},
		SwapFunc: func(i, j int) {
			e.GetEdges()[i], e.GetEdges()[j] = e.GetEdges()[j], e.GetEdges()[i]
		},
	}
	s.Sort()
}

func (e *EdgeDetails) Sort() {
	s := sortable.Sortable{
		LenFunc: func() int {
			return len(e.GetEdges())
		},
		LessFunc: func(i, j int) bool {
			return e.GetEdges()[i].UpdatedAt < e.GetEdges()[j].UpdatedAt
		},
		SwapFunc: func(i, j int) {
			e.GetEdges()[i], e.GetEdges()[j] = e.GetEdges()[j], e.GetEdges()[i]
		},
	}
	s.Sort()
}

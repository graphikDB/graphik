package apipb

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

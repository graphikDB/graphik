# Graphik [![GoDoc](https://godoc.org/github.com/autom8ter/graphik?status.svg)](https://godoc.org/github.com/autom8ter/graphik)

![dag](images/dag.png)

    git@github.com:autom8ter/graphik.git
    
    docker pull colemanword/graphik

An identity-aware, permissioned, persistant [labelled property graph](https://en.wikipedia.org/wiki/Graph_database#Labeled-property_graph) database written in Go

- [x] 100% Go
- [x] Containerized
- [x] Native gRPC Support
- [x] Persistant(bbolt LMDB)
- [x] Horizontally Scaleable ([Raft](https://raft.github.io/))
- [x] Loosely typed(mongo-esque)
- [x] [Prometheus Metrics](https://prometheus.io/)
- [x] [Pprof Metrics](https://blog.golang.org/pprof)
- [x] [Context-Based Timeouts](https://blog.golang.org/context)
- [x] Secure JWT based auth with remote [JWKS](https://auth0.com/docs/tokens/json-web-tokens/json-web-key-sets) support
- [x] Auto JWKS refresh
- [x] Bulk Export
- [x] Bulk Import
- [x] Change Stream Subscriptions
- [x] Channel Based PubSub
- [x] [Common Expression Language](https://opensource.google/projects/cel) Query Filtering
- [x] gRPC Based External Trigger Implementation(sidecar)
- [x] gRPC Based External Authorizer Implementation(sidecar)
- [ ] Kubernetes Operator
- [ ] Helm Chart

## Goals
- Zero database administrative requirements(schema-less)
- Model any number of complex relationships
- Extensible via external plugin model
- Built in identity awareness(Oauth2)
- Integration with existing identity providers(ex: Google, Microsoft, etc)
- Designed for microservice architecture
- Flexible filtering/querying via interpreted expression language
- Horizontally Scaleable
- Fault Tolerant
- Pubsub implementation
- Events (change-streams) 

## API Spec

[API Spec](https://github.com/autom8ter/graphik/blob/master/api/graphik.proto)

[Examples](https://github.com/autom8ter/graphik/blob/master/example_test.go)

```proto
// GraphService is the primary Graph service
service GraphService {
  // Ping returns PONG if the server is health
  rpc Ping(google.protobuf.Empty) returns(Pong) {}
  // JoinCluster Joins the raft node to the cluster
  rpc JoinCluster(RaftNode) returns(google.protobuf.Empty) {}
  // GetSchema gets schema about the Graph node & edge types
  rpc GetSchema(google.protobuf.Empty) returns(Schema){}
 // Me returns a NodeDetail of the currently logged in user(the subject of the JWT)
  rpc Me(MeFilter) returns(NodeDetail){}
  // CreateNode creates a node in the graph
  rpc CreateNode(NodeConstructor) returns(Node){}
  // CreateNodes creates a batch of nodes in the graph
  rpc CreateNodes(NodeConstructors) returns(Nodes){}
  // GetNode gets a single node in the graph
  rpc GetNode(Path) returns(Node){}
  // SearchNodes searches the graph for nodes
  rpc SearchNodes(Filter) returns(Nodes){}
  // PatchNode patches a nodes attributes
  rpc PatchNode(Patch) returns(Node){}
  // PatchNodes patches a batch of nodes attributes
  rpc PatchNodes(Patches) returns(Nodes){}
  // DelNode deletes a node from the graph
  rpc DelNode(Path) returns(google.protobuf.Empty){}
  // DelNodes deletes a batch of nodes from the graph
  rpc DelNodes(Paths) returns(google.protobuf.Empty){}
  // CreateEdge creates an edge in the graph
  rpc CreateEdge(EdgeConstructor) returns(Edge){}
  // CreateEdges creates a batch of edges in the graph
  rpc CreateEdges(EdgeConstructors) returns(Edges){}
  // GetEdge gets a single edge in the graph
  rpc GetEdge(Path) returns(Edge){}
  // SearchEdges searches the graph for edges
  rpc SearchEdges(Filter) returns(Edges){}
  // PatchEdge patches an edges attributes
  rpc PatchEdge(Patch) returns(Edge){}
  // PatchEdges patches a batch of edges attributes
  rpc PatchEdges(Patches) returns(Edges){}
  // DelEdge deletes an edge from the graph
  rpc DelEdge(Path) returns(google.protobuf.Empty){}
  // DelEdges deletes a batch of edges from the graph
  rpc DelEdges(Paths) returns(google.protobuf.Empty){}
  // EdgesFrom returns edges that source from the given node path that pass the filter
  rpc EdgesFrom(EdgeFilter) returns(Edges){}
  // EdgesTo returns edges that point to the given node path that pass the filter
  rpc EdgesTo(EdgeFilter) returns(Edges){}
  // ChangeStream subscribes to filtered changes on a stream
  rpc ChangeStream(ChangeFilter) returns (stream StateChange) {}
  // Publish publishes a message to a pubsub channel
  rpc Publish(OutboundMessage) returns(google.protobuf.Empty){}
  // Subscribe subscribes to messages on a pubsub channel
  rpc Subscribe(ChannelFilter) returns(stream Message){}
  // Import imports the Graph into the database
  rpc Import(Graph) returns(Graph){}
  // Export returns the Graph data
  rpc Export(google.protobuf.Empty) returns (Graph){}
  // SubGraph returns a subgraph using the given filter
  rpc SubGraph(SubGraphFilter) returns(Graph){}
  // Shutdown shuts down the database
  rpc Shutdown(google.protobuf.Empty) returns(google.protobuf.Empty){}
}
```
## Key Dependencies

- google.golang.org/grpc
- github.com/hashicorp/raft
- github.com/autom8ter/machine
- github.com/google/cel-go/cel
- go.etcd.io/bbolt
- go.uber.org/zap
- golang.org/x/oauth2

## Use Cases

- relational state-machine for identity-aware applications

## Flags

```text
      --authorizers strings    registered authorizers (env: GRAPHIK_AUTHORIZERS)
      --grpc.bind string       grpc server bind address (default ":7820")
      --http.bind string       http server bind address (default ":7830")
      --http.headers strings   cors allowed headers (env: GRAPHIK_HTTP_HEADERS)
      --http.methods strings   cors allowed methods (env: GRAPHIK_HTTP_METHODS)
      --http.origins strings   cors allowed origins (env: GRAPHIK_HTTP_ORIGINS)
      --jwks strings           authorized jwks uris ex: https://www.googleapis.com/oauth2/v3/certs (env: GRAPHIK_JWKS_URIS)
      --metrics                enable prometheus & pprof metrics
      --raft.bind string       raft protocol bind address (default "localhost:7840")
      --raft.id string         raft node id (env: GRAPHIK_RAFT_ID) (default "leader")
      --raft.join string       join raft at target address
      --storage string         persistant storage path (env: GRAPHIK_STORAGE_PATH) (default "/tmp/graphik")
      --triggers strings       registered triggers (env: GRAPHIK_TRIGGERS)


```

## Graphik Plugins (optional)

![plugins](images/graphdb-plugins.png)


Graphik plugins are custom, single-method, grpc-based sidecars that the Graphik server integrates with. 
This pattern is similar to Envoy external filters & Kubernetes mutating webhooks / admission controller

Plugin API Spec:

```proto
// Triggers are executed before & after all graph state changes
service TriggerService {
  rpc HandleTrigger(Trigger) returns(StateChange){}
}

// Authorizers are executed within a request middleware/interceptor to determine whether the request is permitted
service AuthorizationService {
 rpc Authorize(RequestIntercept) returns(Decision){}
}
```


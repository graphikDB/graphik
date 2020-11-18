# Graphik

    git@github.com:autom8ter/graphik.git
    
    docker pull colemanword/graphik

An identity-aware, permissioned, persistant labelled property graph database written in Go

- [x] 100% Go
- [x] Containerized
- [x] Native gRPC Support
- [x] Persistant (Raft)
- [x] Horizontally Scaleable (Raft)
- [x] Loosely typed(mongo-esque)
- [x] Prometheus Metrics
- [x] Pprof Metrics
- [x] Context-Based Timeouts
- [x] Live runtime config update
- [x] Secure JWT based auth with remote JWKS support
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


[API Spec](https://github.com/autom8ter/graphik/blob/master/api/graphik.proto)

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

## TODO

- [ ] Auto redirect mutations to Raft leader
- [ ] E2E Tests
- [ ] Benchmarks Against Other Graph Databases

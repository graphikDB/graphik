# Graphik

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
- [x] Bulk Export
- [x] Bulk Import
- [x] Change Stream Subscriptions
- [x] Channel Based PubSub
- [x] [Common Expression Language](https://opensource.google/projects/cel) Query Filtering
- [x] [Common Expression Language](https://opensource.google/projects/cel) Based Authorization
- [x] gRPC Based External Trigger Implementation(sidecar)
- [x] gRPC Based External Authorizer Implementation(sidecar)
- [ ] Kubernetes Operator
- [ ] Helm Chart

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

## TODO

- [ ] Auto redirect mutations to Raft leader
- [ ] E2E Tests
- [ ] Benchmarks Against Other Graph Databases

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

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
- [x] Live runtime config update
- [x] Private/Admin api
- [x] Secure JWT based auth with remote JWKS support
- [x] Bulk Export
- [x] Bulk Import
- [x] Change Stream Subscriptions
- [x] Channel Based PubSub
- [x] [Common Expression Language](https://opensource.google/projects/cel) Query Filtering
- [x] [Common Expression Language](https://opensource.google/projects/cel) Based Authorization
- [ ] [Common Expression Language](https://opensource.google/projects/cel) Based Constraints
- [ ] [Common Expression Language](https://opensource.google/projects/cel) Based Triggers
- [ ] Kubernetes Operator
- [ ] Helm Chart

## Key Dependencies

- github.com/hashicorp/raft
- github.com/autom8ter/machine
- github.com/google/cel-go/cel

## Use Cases

- relational state-machine fine-grained authorization

## TODO

- [ ] Auto redirect mutations to Raft leader
- [ ] E2E Tests
- [ ] Benchmarks Against Other Graph Databases

## Flags

```text
      --auth.expressions strings   auth middleware expressions (env: GRAPHIK_AUTH_EXPRESSIONS)
      --auth.jwks strings          authorizaed jwks uris ex: https://www.googleapis.com/oauth2/v3/certs (env: GRAPHIK_JWKS_URIS)
      --grpc.bind string           grpc server bind address (default ":7820")
      --http.bind string           http server bind address (default ":7830")
      --http.headers strings       cors allowed headers (env: GRAPHIK_HTTP_HEADERS)
      --http.methods strings       cors allowed methods (env: GRAPHIK_HTTP_METHODS)
      --http.origins strings       cors allowed origins (env: GRAPHIK_HTTP_ORIGINS)
      --raft.bind string           raft protocol bind address (default "localhost:7840")
      --raft.nodeid string         raft node id (env: GRAPHIK_RAFT_ID)
      --raft.storage.path string   raft storage path (default "/tmp/graphik")

```

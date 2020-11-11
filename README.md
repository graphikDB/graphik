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


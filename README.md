# Graphik

An identity-aware, permissioned, persistant labelled property graph database written in Go

- [x] 100% Go
- [x] Containerized
- [x] Directed edge Support
- [x] UnDirected Edge Support
- [x] Native gRPC Support
- [x] Persistant (Raft)
- [x] Horizontally Scaleable (Raft)
- [ ] Auto redirect mutations to Raft leader
- [x] Loosely typed(mongo-esque)
- [x] Prometheus Metrics
- [x] Pprof Metrics
- [x] Live config update
- [ ] Public/User api
- [x] Private/Admin api
- [ ] Cloud Storage(file upload)
- [ ] E2E Tests
- [ ] Benchmarks Against Other Graph Databases
- [x] Secure JWT based auth with remote JWKS support
- [x] Bulk Export
- [x] Bulk Import
- [ ] Change Stream Subscriptions
- [ ] Channel Based PubSub
- [x] [Common Expression Language](https://opensource.google/projects/cel) Query Filtering
- [x] [Common Expression Language](https://opensource.google/projects/cel) Based Authorization
- [ ] [Common Expression Language](https://opensource.google/projects/cel) Based Constraints
- [ ] Triggers/Plugins
- [ ] Kubernetes Operator
- [ ] Helm Chart

## Key Dependencies

- github.com/hashicorp/raft
- github.com/autom8ter/machine
- github.com/google/cel-go/cel

## Use Cases

- relational state-machine fine-grained authorization


# Graphik

![dag](images/dag.png)

    git clone git@github.com:autom8ter/graphik.git
    
    docker pull colemanword/graphik:v0.0.20

Graphik is an identity-aware, permissioned, persistant [labelled property graph](https://en.wikipedia.org/wiki/Graph_database#Labeled-property_graph) database written in Go

## Helpful Links

- [API Docs](docs/README.md)
- [Protobuf API Spec](https://github.com/autom8ter/graphik/blob/master/api/graphik.proto)
- [Graphql API Spec](https://github.com/autom8ter/graphik/blob/master/api/schema.graphqls)
- [Common Expression Language Code Lab](https://codelabs.developers.google.com/codelabs/cel-go/index.html#0)
- [CEL Standard Functions/Definitions](https://github.com/google/cel-spec/blob/master/doc/langdef.md#standard-definitions)
- [Directed Graph Wiki](https://en.wikipedia.org/wiki/Directed_graph)

## Client SDKs
- [x] [graphik-client-go](https://github.com/autom8ter/graphik-client-go)
- [ ] graphik-client-python
- [ ] graphik-client-doc

## Features
- [x] 100% Go
- [x] Native gRPC & GraphQl Support
- [x] Built in GraphQl Playground
- [x] Native OAuth Support & Single Sign On
- [x] Persistant(bbolt LMDB)
- [x] Channel Based PubSub
- [x] Change Stream Subscriptions
- [x] [Common Expression Language](https://opensource.google/projects/cel) Query Filtering
- [x] [Common Expression Language](https://opensource.google/projects/cel) Request Authorization
- [ ] [Common Expression Language](https://opensource.google/projects/cel) Based Triggers
- [x] Object metadata - Auto track created/updated timestamps & who is making updates to objects
- [x] Loosely-Typed(mongo-esque)
- [x] [Prometheus Metrics](https://prometheus.io/)
- [x] [Pprof Metrics](https://blog.golang.org/pprof)
- [x] Secure JWT based auth with remote [JWKS](https://auth0.com/docs/tokens/json-web-tokens/json-web-key-sets) support
- [x] Auto JWKS refresh
- [x] Bulk Export
- [x] Bulk Import

## Key Dependencies

- google.golang.org/grpc
- github.com/autom8ter/machine
- github.com/google/cel-go/cel
- go.etcd.io/bbolt
- go.uber.org/zap
- golang.org/x/oauth2
- github.com/99designs/gqlgen

## Flags

```text
      --allow-headers strings   cors allow headers (env: GRAPHIK_ALLOW_HEADERS) (default [*])
      --allow-methods strings   cors allow methods (env: GRAPHIK_ALLOW_METHODS) (default [HEAD,GET,POST,PUT,PATCH,DELETE])
      --allow-origins strings   cors allow origins (env: GRAPHIK_ALLOW_ORIGINS) (default [*])
      --authorizers strings     registered authorizers (env: GRAPHIK_AUTHORIZERS)
      --jwks strings            authorized jwks uris ex: https://www.googleapis.com/oauth2/v3/certs (env: GRAPHIK_JWKS_URIS)
      --metrics                 enable prometheus & pprof metrics (emv: GRAPHIK_METRICS = true)
      --storage string          persistant storage path (env: GRAPHIK_STORAGE_PATH) (default "/tmp/graphik")

```
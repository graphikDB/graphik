# Graphik[![GoDoc](https://godoc.org/github.com/graphikDB/graphik?status.svg)](https://godoc.org/github.com/graphikDB/graphik)

https://graphikdb.github.io/graphik/


    git clone git@github.com:graphikDB/graphik.git
    
`    docker pull graphikdb/graphik:v0.2.0`

Graphik is an identity-aware, permissioned, persistant document & graph database written in Go


## Helpful Links

- [GraphQL Documentation Site](https://graphikdb.github.io/graphik/)
- [Protobuf/gRPC API Spec](https://github.com/graphikDB/graphik/blob/master/graphik.proto)
- [Graphql API Spec](https://github.com/graphikDB/graphik/blob/master/schema.graphql)
- [Common Expression Language Code Lab](https://codelabs.developers.google.com/codelabs/cel-go/index.html#0)
- [CEL Standard Functions/Definitions](https://github.com/google/cel-spec/blob/master/doc/langdef.md#standard-definitions)
- [OpenID Connect](https://openid.net/connect/)
- [Directed Graph Wiki](https://en.wikipedia.org/wiki/Directed_graph)

## Features
- [x] 100% Go
- [x] Native gRPC Support
- [x] GraphQL Support
- [x] Native Document & Graph Database
- [x] [Index-free Adjacency](https://dzone.com/articles/the-secret-sauce-of-graph-databases)
- [x] Native OAuth/OIDC Support & Single Sign On
- [x] Embedded SSO protected GraphQl Playground
- [x] Persistant(bbolt LMDB)
- [x] Channel Based PubSub
- [x] [Common Expression Language](https://opensource.google/projects/cel) Query Filtering
- [x] [Common Expression Language](https://opensource.google/projects/cel) Request Authorization
- [x] [Common Expression Language](https://opensource.google/projects/cel) Type Validators
- [ ] [Common Expression Language](https://opensource.google/projects/cel) Based Triggers
- [x] Loosely-Typed(mongo-esque)
- [x] [Prometheus Metrics](https://prometheus.io/)
- [x] [Pprof Metrics](https://blog.golang.org/pprof)
- [x] Safe to Deploy Publicly(with authorizers)
- [x] Read-Optimized

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
      --allow-headers strings             cors allow headers (env: GRAPHIK_ALLOW_HEADERS) (default [*])
      --allow-methods strings             cors allow methods (env: GRAPHIK_ALLOW_METHODS) (default [HEAD,GET,POST,PUT,PATCH,DELETE])
      --allow-origins strings             cors allow origins (env: GRAPHIK_ALLOW_ORIGINS) (default [*])
      --metrics                           enable prometheus & pprof metrics (emv: GRAPHIK_METRICS = true) (default true)
      --open-id string                    open id connect discovery uri ex: https://accounts.google.com/.well-known/openid-configuration (env: GRAPHIK_OPEN_ID)
      --playground-client-id string       playground oauth client id (env: GRAPHIK_PLAYGROUND_CLIENT_ID)
      --playground-client-secret string   playground oauth client secret (env: GRAPHIK_PLAYGROUND_CLIENT_SECRET
      --playground-redirect string        playground oauth redirect (env: GRAPHIK_PLAYGROUND_REDIRECT) (default "http://localhost:7820/playground/callback")
      --root-users strings                cors allow methods (env: GRAPHIK_ROOT_USERS)
      --storage string                    persistant storage path (env: GRAPHIK_STORAGE_PATH) (default "/tmp/graphik")
      --tls-cert string                   path to tls certificate (env: GRAPHIK_TLS_CERT)
      --tls-key string                    path to tls key (env: GRAPHIK_TLS_KEY)
```

## gRPC Client SDKs
- [x] [Go](https://godoc.org/github.com/graphikDB/graphik/graphik-client-go)
- [x] [Python](gen/grpc/python)
- [x] [PHP](gen/grpc/php)
- [x] [Javascript](gen/grpc/js)
- [x] [Java](gen/grpc/java)
- [x] [C#](gen/grpc/csharp)
- [x] [Ruby](gen/grpc/ruby)
    
## Sample GraphQL Queries

### Node Traversal
```graphql
# Write your query or mutation here
query {
  traverse(input: {
    root: {
      gid: "coleman.word@graphikdb.io"
      gtype: "user"
    }
    algorithm: BFS
    limit: 6
		max_depth: 1
		max_hops: 10
  }){
    traversals {
      doc {
        ref {
          gid
          gtype
        }
      }
      traversal_path {
        gid
        gtype
      }
			depth
			hops
    }
  }
}
```


## Deployment

### Docker-Compose

add this docker-compose.yml to ${pwd}:

    version: '3.7'
    services:
      graphik:
        image: graphikdb/graphik:v0.2.0
        env_file:
          - .env
        ports:
          - "7820:7820"
          - "7821:7821"
        volumes:
          - default:/tmp/graphik
        networks:
          default:
            aliases:
              - graphikdb
    networks:
      default:
    
    volumes:
      default:

add a .env file to ${pwd}
    
    GRAPHIK_PLAYGROUND_CLIENT_ID=${client_id}
    GRAPHIK_PLAYGROUND_CLIENT_SECRET=${client_secret}
    GRAPHIK_PLAYGROUND_REDIRECT=http://localhost:7820/playground/callback
    GRAPHIK_OPEN_ID=${open_id_connect_metadata_url}
    #GRAPHIK_ALLOW_HEADERS=${cors_headers}
    #GRAPHIK_ALLOW_METHOD=${cors_methos}
    #GRAPHIK_ALLOW_ORIGINS=${cors_origins}
    #GRAPHIK_ROOT_USERS=${root_users}
    #GRAPHIK_TLS_CERT=${tls_cert_path}
    #GRAPHIK_TLS_KEY=${tls_key_path}
    
then run:

    docker-compose -f docker-compose.yml pull
    docker-compose -f docker-compose.yml up -d
    
to shutdown:
    
    docker-compose -f docker-compose.yml down --remove-orphans
    
 ### Kubernetes
 
 Coming Soon
 
 ### Virtual Machine
 
 Coming Soon

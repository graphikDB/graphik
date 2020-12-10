# Graphik

https://graphikdb.github.io/graphik/

[![GoDoc](https://godoc.org/github.com/graphikDB/graphik?status.svg)](https://godoc.org/github.com/graphikDB/graphik)

    git clone git@github.com:graphikDB/graphik.git
    
`    docker pull graphikdb/graphik:v0.2.3`

Graphik is an identity-aware, permissioned, persistant document & graph database written in Go

  * [Helpful Links](#helpful-links)
  * [Features](#features)
  * [Key Dependencies](#key-dependencies)
  * [Flags](#flags)
  * [gRPC Client SDKs](#grpc-client-sdks)
  * [Implemenation Details](#implemenation-details)
    + [Primitives](#primitives)
    + [Login/Authorization/Authorizers](#login-authorization-authorizers)
      - [Authorizers Examples](#authorizers-examples)
    + [Secondary Indexes](#secondary-indexes)
      - [Secondary Index Examples](#secondary-index-examples)
    + [Type Validators](#type-validators)
      - [Type Validator Examples](#type-validator-examples)
    + [Identity Graph](#identity-graph)
    + [Additional Details](#additional-details)
  * [Sample GraphQL Queries](#sample-graphql-queries)
    + [Node Traversal](#node-traversal)
  * [Deployment](#deployment)
    + [Docker-Compose](#docker-compose)
    + [Kubernetes](#kubernetes)
    + [Linux](#linux)
    + [Mac/Darwin](#mac-darwin)
    + [Windows](#windows)


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
- [x] Identity-Aware PubSub with Channels & Message Filtering(gRPC & graphQL)
- [x] [Common Expression Language](https://opensource.google/projects/cel) Query Filtering
- [x] [Common Expression Language](https://opensource.google/projects/cel) Request Authorization
- [x] [Common Expression Language](https://opensource.google/projects/cel) Type Validators
- [x] Loosely-Typed(mongo-esque)
- [x] [Prometheus Metrics](https://prometheus.io/)
- [x] [Pprof Metrics](https://blog.golang.org/pprof)
- [x] Safe to Deploy Publicly(with authorizers)
- [x] Read-Optimized
- [x] Full Text Search(CEL)
- [x] Regular Expressions(CEL)
- [x] Client to Server streaming(gRPC only)

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
      --root-users strings                a list of email addresses that bypass registered authorizers(env: GRAPHIK_ROOT_USERS)
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

## Implemenation Details

Please see [GraphQL Documentation Site](https://graphikdb.github.io/graphik/) for additional details

### Primitives

- `Ref` == direct pointer to an doc or connection.

```proto
message Ref {
  // gtype is the type of the doc/connection ex: pet
  string gtype =1 [(validator.field) = {regex : "^.{1,225}$"}];
  // gid is the unique id of the doc/connection within the context of it's type
  string gid =2 [(validator.field) = {regex : "^.{1,225}$"}];
}
```

- `Doc` == JSON document in document storage terms AND vertex/node in graph theory

```proto
message Doc {
    // ref is the ref to the doc
    Ref ref =1 [(validator.field) = {msg_exists : true}];
    // k/v pairs
    google.protobuf.Struct attributes =2;
}
```
       
- `Connection` == graph edge/relationship in graph theory. Connections relate Docs to one another.

```proto
message Connection {
  // ref is the ref to the connection
  Ref ref =1 [(validator.field) = {msg_exists : true}];
  // attributes are k/v pairs
  google.protobuf.Struct attributes =2;
  // directed is false if the connection is bi-directional
  bool directed =3;
  // from is the doc ref that is the source of the connection
  Ref from =4 [(validator.field) = {msg_exists : true}];
  // to is the doc ref that is the destination of the connection
  Ref to =5 [(validator.field) = {msg_exists : true}];
}
```

### Login/Authorization/Authorizers
- an access token `Authorization: Bearer ${token}` from the configured open-id connect identity provider is required for all database functionality
- the access token is used to fetch the users info from the oidc userinfo endpoint fetched from the oidc metadata url
- if a user is not present in the database, one will be automatically created under the gtype: `user` with their email address as their `gid`
- once the user is fetched, it is evaluated(along with the request & request method) against any registered authorizers(CEL expression) in the database.
    - if an authorizer evaluates false, the request will be denied
    - authorizers may be used to restrict access to functionality by domain, role, email, etc
    - registered root users(see flags) bypass these authorizers
- authorizers are completely optional but highly recommended

#### Authorizers Examples
Coming Soon

### Secondary Indexes
- secondary indexes are CEL expressions evaluated against a particular type of Doc or Connection
- indexes may be used to speed up queries that iterate over a large number of elements
- secondary indexes are completely optional but recommended

#### Secondary Index Examples
Coming Soon

### Type Validators
- type validators are CEL expressions evaluated against a particular type of Doc or Connection to enforce custom constraints
- type validators are completely optional

#### Type Validator Examples
Coming Soon

### Identity Graph
- any time a document is created, a connection of type `created` from the origin user to the new document is also created
- any time a document is created, a connection of type `created_by` from the new document to the origin user is also created
- any time a document is edited, a connection of type `edited` from the origin user to the new document is also created(if none exists)
- any time a document is edited, a connection of type `edited_by` from the new document to the origin user is also created(if none exists)
- every document a user has ever interacted with may be queried via the Traverse method with the user as the root document of the traversal

### Additional Details
- any time a Doc is deleted, so are all of its connections

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

Regardless of deployment methodology, please set the following environmental variables or include them in a ${pwd}/.env file

```
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
```

### Docker-Compose

add this docker-compose.yml to ${pwd}:

    version: '3.7'
    services:
      graphik:
        image: graphikdb/graphik:v0.2.3
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

    
then run:

    docker-compose -f docker-compose.yml pull
    docker-compose -f docker-compose.yml up -d
    
to shutdown:
    
    docker-compose -f docker-compose.yml down --remove-orphans
    
 ### Kubernetes
 
 Coming Soon
 
 ### Linux
 
    curl -L https://github.com/graphikDB/graphik/releases/download/v0.2.3/graphik_linux_amd64 \
    --output /usr/local/bin/graphik && \
    chmod +x /usr/local/bin/graphik 
    
 ### Mac/Darwin
 
    curl -L https://github.com/graphikDB/graphik/releases/download/v0.2.3/graphik_darwin_amd64 \
    --output /usr/local/bin/graphik && \
    chmod +x /usr/local/bin/graphik

### Windows

    Nope - use docker
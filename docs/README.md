# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [api/graphik.proto](#api/graphik.proto)
    - [Change](#api.Change)
    - [ChannelFilter](#api.ChannelFilter)
    - [Connection](#api.Connection)
    - [ConnectionChange](#api.ConnectionChange)
    - [ConnectionConstructor](#api.ConnectionConstructor)
    - [ConnectionConstructors](#api.ConnectionConstructors)
    - [ConnectionDetail](#api.ConnectionDetail)
    - [ConnectionDetails](#api.ConnectionDetails)
    - [ConnectionFilter](#api.ConnectionFilter)
    - [Connections](#api.Connections)
    - [Doc](#api.Doc)
    - [DocChange](#api.DocChange)
    - [DocConstructor](#api.DocConstructor)
    - [DocConstructors](#api.DocConstructors)
    - [DocDetail](#api.DocDetail)
    - [DocDetailFilter](#api.DocDetailFilter)
    - [DocDetails](#api.DocDetails)
    - [Docs](#api.Docs)
    - [ExpressionFilter](#api.ExpressionFilter)
    - [Filter](#api.Filter)
    - [Graph](#api.Graph)
    - [Index](#api.Index)
    - [MeFilter](#api.MeFilter)
    - [Message](#api.Message)
    - [Metadata](#api.Metadata)
    - [OutboundMessage](#api.OutboundMessage)
    - [Patch](#api.Patch)
    - [PatchFilter](#api.PatchFilter)
    - [Path](#api.Path)
    - [Paths](#api.Paths)
    - [Pong](#api.Pong)
    - [Request](#api.Request)
    - [Schema](#api.Schema)
    - [SubGraphFilter](#api.SubGraphFilter)
  
    - [DatabaseService](#api.DatabaseService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="api/graphik.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## api/graphik.proto



<a name="api.Change"></a>

### Change
Change represents a set of state changes in the graph


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| method | [string](#string) |  | method is the gRPC method invoked |
| identity | [Doc](#api.Doc) |  | identity is the identity invoking the change |
| timestamp | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | timestamp is when the change was made |
| connection_changes | [ConnectionChange](#api.ConnectionChange) | repeated | connection_changes are state changes to connections |
| doc_changes | [DocChange](#api.DocChange) | repeated | doc_changes are state changes to docs |






<a name="api.ChannelFilter"></a>

### ChannelFilter
ChannelFilter is used to filter messages in a pubsub channel


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| channel | [string](#string) |  | channel is the target channel to filter from |
| expression | [string](#string) |  | expression is CEL expression used to filter messages |






<a name="api.Connection"></a>

### Connection
Connection is a graph primitive that represents a relationship between two docs


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [Path](#api.Path) |  | path is the path to the connection |
| attributes | [google.protobuf.Struct](#google.protobuf.Struct) |  | attributes are k/v pairs |
| directed | [bool](#bool) |  | directed is false if the connection is bi-directional |
| from | [Path](#api.Path) |  | from is the doc path that is the source of the connection |
| to | [Path](#api.Path) |  | to is the doc path that is the destination of the connection |
| metadata | [Metadata](#api.Metadata) |  | metadata is general metadata collected about the connection |






<a name="api.ConnectionChange"></a>

### ConnectionChange



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| before | [Connection](#api.Connection) |  |  |
| after | [Connection](#api.Connection) |  |  |






<a name="api.ConnectionConstructor"></a>

### ConnectionConstructor
ConnectionConstructor is used to create an Connection


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [Path](#api.Path) |  | path is the path to the connection. If an id is not provided, a unique id will be generated |
| attributes | [google.protobuf.Struct](#google.protobuf.Struct) |  | attributes are k/v pairs |
| directed | [bool](#bool) |  | directed is false if the connection is bi-directional |
| from | [Path](#api.Path) |  | from is the doc path that is the root of the connection |
| to | [Path](#api.Path) |  | to is the doc path that is the destination of the connection |






<a name="api.ConnectionConstructors"></a>

### ConnectionConstructors
ConnectionConstructors is an array of ConnectionConstructor


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| connections | [ConnectionConstructor](#api.ConnectionConstructor) | repeated |  |






<a name="api.ConnectionDetail"></a>

### ConnectionDetail
ConnectionDetail is an connection with both of it&#39;s connected docs fully loaded


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [Path](#api.Path) |  | path is the path to the connection |
| attributes | [google.protobuf.Struct](#google.protobuf.Struct) |  | attributes are k/v pairs |
| directed | [bool](#bool) |  | directed is false if the connection is bi-directional |
| from | [Doc](#api.Doc) |  | from is the full doc that is the root of the connection |
| to | [Doc](#api.Doc) |  | to is the full doc that is the destination of the connection |
| metadata | [Metadata](#api.Metadata) |  | metadata is general metadata collected about the connection |






<a name="api.ConnectionDetails"></a>

### ConnectionDetails
ConnectionDetails is an array of ConnectionDetail


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| connections | [ConnectionDetail](#api.ConnectionDetail) | repeated |  |
| seek_next | [string](#string) |  |  |






<a name="api.ConnectionFilter"></a>

### ConnectionFilter
ConnectionFilter is used to fetch connections related to a single noted


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| doc_path | [Path](#api.Path) |  | doc_path is the path to the target doc |
| gtype | [string](#string) |  | gtype is the type of connections to return |
| expression | [string](#string) |  | expression is a CEL expression used to filter connections/modes |
| limit | [int32](#int32) |  | limit is the maximum number of items to return |
| sort | [string](#string) |  | custom sorting of the results. |
| seek | [string](#string) |  | seek to a specific key for pagination |
| reverse | [bool](#bool) |  | reverse the results |






<a name="api.Connections"></a>

### Connections
Connections is an array of Connection


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| connections | [Connection](#api.Connection) | repeated |  |
| seek_next | [string](#string) |  |  |






<a name="api.Doc"></a>

### Doc
Doc is a Graph primitive representing a single entity/resource. It is connected to other docs via Connections


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [Path](#api.Path) |  | path is the path to the doc |
| attributes | [google.protobuf.Struct](#google.protobuf.Struct) |  | k/v pairs |
| metadata | [Metadata](#api.Metadata) |  | metadata is general metadata collected about the doc |






<a name="api.DocChange"></a>

### DocChange
DocChange is a single state change to a doc


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| before | [Doc](#api.Doc) |  | before is the doc before state change |
| after | [Doc](#api.Doc) |  | after is the doc after state change |






<a name="api.DocConstructor"></a>

### DocConstructor
DocConstructor is used to create a doc


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [Path](#api.Path) |  | path is the path to the doc. If an id is not provided, a unique id will be generated |
| attributes | [google.protobuf.Struct](#google.protobuf.Struct) |  | arbitrary k/v pairs |






<a name="api.DocConstructors"></a>

### DocConstructors
DocConstructor is used to create a batch of docs


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| docs | [DocConstructor](#api.DocConstructor) | repeated | docs is an array of doc constructors |






<a name="api.DocDetail"></a>

### DocDetail
DocDetail is a doc with its connected connections


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [Path](#api.Path) |  | path is the path to the doc |
| attributes | [google.protobuf.Struct](#google.protobuf.Struct) |  | arbitrary k/v pairs |
| connections_from | [ConnectionDetails](#api.ConnectionDetails) |  | connections_from are connections that source from this doc |
| connections_to | [ConnectionDetails](#api.ConnectionDetails) |  | connections_to are connections that point toward this doc |
| metadata | [Metadata](#api.Metadata) |  | metadata is general metadata collected about the doc |






<a name="api.DocDetailFilter"></a>

### DocDetailFilter
DocDetailFilter is used to fetch doc details


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [Path](#api.Path) |  | path is the path to the doc |
| from_connections | [Filter](#api.Filter) |  |  |
| to_connections | [Filter](#api.Filter) |  |  |






<a name="api.DocDetails"></a>

### DocDetails
DocDetails is an array of DocDetail


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| doc_details | [DocDetail](#api.DocDetail) | repeated |  |
| seek_next | [string](#string) |  |  |






<a name="api.Docs"></a>

### Docs
Docs is an array of docs


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| docs | [Doc](#api.Doc) | repeated | docs is an array of docs |
| seek_next | [string](#string) |  |  |






<a name="api.ExpressionFilter"></a>

### ExpressionFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| expression | [string](#string) |  | expression is a CEL expression used to filter connections/nodes |






<a name="api.Filter"></a>

### Filter
Filter is a generic filter using Common Expression Language


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gtype | [string](#string) |  | gtype is the doc/connection type to be filtered |
| expression | [string](#string) |  | expression is a CEL expression used to filter connections/modes |
| limit | [int32](#int32) |  | limit is the maximum number of items to return |
| sort | [string](#string) |  | custom sorting of the results. |
| seek | [string](#string) |  | seek to a specific key for pagination |
| reverse | [bool](#bool) |  | reverse the results |






<a name="api.Graph"></a>

### Graph
Graph is an array of docs and connections


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| docs | [Docs](#api.Docs) |  | docs are docs present in the graph |
| connections | [Connections](#api.Connections) |  | connections are connections present in the graph |






<a name="api.Index"></a>

### Index



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| gtype | [string](#string) |  | gtype is the doc/connection type to be filtered |
| expression | [string](#string) |  | expression is a CEL expression used to filter connections/modes |
| sequence | [uint64](#uint64) |  |  |






<a name="api.MeFilter"></a>

### MeFilter
MeFilter is used to fetch a DocDetail representing the identity in the inbound JWT token


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| connections_from | [Filter](#api.Filter) |  | connections_from is a filter used to filter connections from the identity making the request |
| connections_to | [Filter](#api.Filter) |  | connections_to is a filter used to filter connections to the identity making the request |






<a name="api.Message"></a>

### Message
Message is received on PubSub subscriptions


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| channel | [string](#string) |  | channel is the channel the message was sent to |
| data | [google.protobuf.Struct](#google.protobuf.Struct) |  | data is the data sent with the message |
| sender | [Path](#api.Path) |  | sender is the identity that sent the message |
| timestamp | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | timestamp is when the message was sent |






<a name="api.Metadata"></a>

### Metadata
Metadata is general metadata collected on docs/connections


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | created_at is the unix timestamp when the doc/connection was created |
| updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | updated_at is the unix timestamp when the doc/connection was last updated |
| created_by | [Path](#api.Path) |  | created_by is the identity that initially created the doc/connection |
| updated_by | [Path](#api.Path) |  | updated_by is the identity that last modified the doc/connection |
| sequence | [uint64](#uint64) |  | sequence is the sequence within the context of the doc/connection type |
| version | [uint64](#uint64) |  | version iterates by 1 every time the doc/connection is modified |






<a name="api.OutboundMessage"></a>

### OutboundMessage
OutboundMessage is a message to be published to a pubsub channel


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| channel | [string](#string) |  | channel is the target channel to send the message to |
| data | [google.protobuf.Struct](#google.protobuf.Struct) |  | data is the data to send with the message |






<a name="api.Patch"></a>

### Patch
Patch patches the attributes of a Doc or Connection


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [Path](#api.Path) |  | path is the path to the target doc/connection to patch |
| attributes | [google.protobuf.Struct](#google.protobuf.Struct) |  | attributes are k/v pairs used to overwrite k/v pairs on a doc/connection |






<a name="api.PatchFilter"></a>

### PatchFilter
PatchFilter is used to patch docs/connections


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| filter | [Filter](#api.Filter) |  | filter is used to filter docs/connections to patch |
| attributes | [google.protobuf.Struct](#google.protobuf.Struct) |  | attributes are k/v pairs used to overwrite k/v pairs on all docs/connections that pass the filter |






<a name="api.Path"></a>

### Path
Path describes a doc/connection type &amp; id


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gtype | [string](#string) |  | gtype is the type of the doc/connection ex: pet |
| gid | [string](#string) |  | gid is the unique id of the doc/connection within the context of it&#39;s type |






<a name="api.Paths"></a>

### Paths
Paths is an array of paths


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| paths | [Path](#api.Path) | repeated |  |






<a name="api.Pong"></a>

### Pong
Pong returns PONG if the server is healthy


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  | message returns PONG if healthy |






<a name="api.Request"></a>

### Request



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| method | [string](#string) |  | method is the rpc method |
| identity | [Doc](#api.Doc) |  | identity is the identity making the request |
| timestamp | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | timestamp is when the intercept was received |
| request | [google.protobuf.Struct](#google.protobuf.Struct) |  | request is the intercepted request |






<a name="api.Schema"></a>

### Schema
Schema returns registered connection &amp; doc types


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| connection_types | [string](#string) | repeated | connection_types are the types of connections in the graph |
| doc_types | [string](#string) | repeated | doc_types are the types of docs in the graph |






<a name="api.SubGraphFilter"></a>

### SubGraphFilter
SubGraphFilter is used to filter docs/connections in the graph


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| doc_filter | [Filter](#api.Filter) |  | doc_filter is a filter used to filter docs in the graph |
| connection_filter | [Filter](#api.Filter) |  | connection_filter is a filter used to filter the connections of each doc returned by the doc_filter |





 

 

 


<a name="api.DatabaseService"></a>

### DatabaseService
DatabaseService is the primary database service

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Ping | [.google.protobuf.Empty](#google.protobuf.Empty) | [Pong](#api.Pong) | Ping returns PONG if the server is health |
| GetSchema | [.google.protobuf.Empty](#google.protobuf.Empty) | [Schema](#api.Schema) | GetSchema gets schema about the Graph doc &amp; connection types |
| Me | [MeFilter](#api.MeFilter) | [DocDetail](#api.DocDetail) | Me returns a DocDetail of the currently logged in identity(the subject of the JWT) |
| CreateDoc | [DocConstructor](#api.DocConstructor) | [Doc](#api.Doc) | CreateDoc creates a doc in the graph |
| CreateDocs | [DocConstructors](#api.DocConstructors) | [Docs](#api.Docs) | CreateDocs creates a batch of docs in the graph |
| GetDoc | [Path](#api.Path) | [Doc](#api.Doc) | GetDoc gets a single doc in the graph |
| SearchDocs | [Filter](#api.Filter) | [Docs](#api.Docs) | SearchDocs searches the graph for docs |
| PatchDoc | [Patch](#api.Patch) | [Doc](#api.Doc) | PatchDoc patches a docs attributes |
| PatchDocs | [PatchFilter](#api.PatchFilter) | [Docs](#api.Docs) | PatchDocs patches a batch of docs attributes that pass the patch filter |
| DelDoc | [Path](#api.Path) | [.google.protobuf.Empty](#google.protobuf.Empty) | DelDoc deletes a doc &amp; all of it&#39;s connected connections |
| DelDocs | [Filter](#api.Filter) | [.google.protobuf.Empty](#google.protobuf.Empty) | DelDocs deletes a batch of docs that pass the filter |
| CreateConnection | [ConnectionConstructor](#api.ConnectionConstructor) | [Connection](#api.Connection) | CreateConnection creates an connection in the graph |
| CreateConnections | [ConnectionConstructors](#api.ConnectionConstructors) | [Connections](#api.Connections) | CreateConnections creates a batch of connections in the graph |
| GetConnection | [Path](#api.Path) | [Connection](#api.Connection) | GetConnection gets a single connection in the graph |
| SearchConnections | [Filter](#api.Filter) | [Connections](#api.Connections) | SearchConnections searches the graph for connections |
| PatchConnection | [Patch](#api.Patch) | [Connection](#api.Connection) | PatchConnection patches an connections attributes |
| PatchConnections | [PatchFilter](#api.PatchFilter) | [Connections](#api.Connections) | PatchConnections patches a batch of connections attributes that pass the patch filter |
| DelConnection | [Path](#api.Path) | [.google.protobuf.Empty](#google.protobuf.Empty) | DelConnection deletes an connection from the graph |
| DelConnections | [Filter](#api.Filter) | [.google.protobuf.Empty](#google.protobuf.Empty) | DelConnections deletes a batch of connections that pass the filter |
| ConnectionsFrom | [ConnectionFilter](#api.ConnectionFilter) | [Connections](#api.Connections) | ConnectionsFrom returns connections that source from the given doc path that pass the filter |
| ConnectionsTo | [ConnectionFilter](#api.ConnectionFilter) | [Connections](#api.Connections) | ConnectionsTo returns connections that point to the given doc path that pass the filter |
| Publish | [OutboundMessage](#api.OutboundMessage) | [.google.protobuf.Empty](#google.protobuf.Empty) | Publish publishes a message to a pubsub channel |
| Subscribe | [ChannelFilter](#api.ChannelFilter) | [Message](#api.Message) stream | Subscribe subscribes to messages on a pubsub channel |
| SubscribeChanges | [ExpressionFilter](#api.ExpressionFilter) | [Change](#api.Change) stream |  |
| Import | [Graph](#api.Graph) | [Graph](#api.Graph) | Import imports the Graph into the database |
| Export | [.google.protobuf.Empty](#google.protobuf.Empty) | [Graph](#api.Graph) | Export returns the Graph data |
| SubGraph | [SubGraphFilter](#api.SubGraphFilter) | [Graph](#api.Graph) | SubGraph returns a subgraph using the given filter |
| Shutdown | [.google.protobuf.Empty](#google.protobuf.Empty) | [.google.protobuf.Empty](#google.protobuf.Empty) | Shutdown shuts down the database |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |


# graphik
--
    import "github.com/autom8ter/graphik"

go:generate godocdown -o DOCS.md

## Usage

#### type Attributer

```go
type Attributer interface {
	// Counter returns the length of the map
	Counter
	// SetAttribute overwrites an k-v in the map
	SetAttribute(key string, val interface{})
	// GetAttribute returns the value if it exists or nil if it doesnt
	GetAttribute(key string) interface{}
	// Range iterates over the values. If false is returned, the range function will break.
	Range(fn func(k string, v interface{}) bool)
	// Attributer can marshal/unmarshal itself
	Encoder
	// returns a human readable string
	fmt.Stringer
}
```

Attributer is a concurrency safe map[string]interface{}

#### func  NewAttributer

```go
func NewAttributer(attr map[string]interface{}) Attributer
```

#### type BasicQuery

```go
type BasicQuery interface {
	Where() WhereFunc
	Limit() int
}
```

BasicQuery is a WHERE and a LIMIT clause

#### type Counter

```go
type Counter interface {
	Count() int
}
```

Counter returns a count of something

#### type Edge

```go
type Edge interface {
	Attributer
	EdgePath
}
```

Edge is the path from one node to another with a relationship and attributes

#### func  NewEdge

```go
func NewEdge(path EdgePath) Edge
```

#### type EdgeHandlerFunc

```go
type EdgeHandlerFunc func(g Graph, e Edge) error
```

EdgeHandlerFunc that executes logic against an edge in a graph

#### type EdgePath

```go
type EdgePath interface {
	Relationship() string
	From() Path
	To() Path
	PathString() string
}
```


#### func  NewEdgePath

```go
func NewEdgePath(from Path, relationship string, to Path) EdgePath
```

#### type EdgeQuery

```go
type EdgeQuery interface {
	// WHERE and a LIMIT clause
	BasicQuery
	// Query modifier for chaining
	Mod(fn EdgeQueryModFunc) EdgeQuery
	// FromType returns the root type of node to search for aka: person
	FromType() string
	// FromKey returns the root key of node to search for aka: coleman_word
	FromKey() string
	// Relationship returns the relationship of the edge to search for aka: coleman_word
	Relationship() string
	// ToType returns the destination type of the edge to search for aka: person
	ToType() string
	// ToKey returns the destination key of the edge to search for aka: tyler_washburne
	ToKey() string
	// When an edge is found, the handler is executed against it
	Handler() EdgeHandlerFunc
	// Validate returns the query and an error if it's invalid
	Validate() (EdgeQuery, error)
}
```

EdgeQuery is used to find edges and execute logic on them

#### func  NewEdgeQuery

```go
func NewEdgeQuery() EdgeQuery
```

#### type EdgeQueryModFunc

```go
type EdgeQueryModFunc func(query EdgeQuery) EdgeQuery
```

EdgeQueryModFunc modifies an edge query

#### func  EdgeModFromKey

```go
func EdgeModFromKey(fromKey string) EdgeQueryModFunc
```

#### func  EdgeModFromType

```go
func EdgeModFromType(fromType string) EdgeQueryModFunc
```

#### func  EdgeModHandler

```go
func EdgeModHandler(handler EdgeHandlerFunc) EdgeQueryModFunc
```

#### func  EdgeModLimit

```go
func EdgeModLimit(limit int) EdgeQueryModFunc
```

#### func  EdgeModRelationship

```go
func EdgeModRelationship(relationship string) EdgeQueryModFunc
```

#### func  EdgeModToKey

```go
func EdgeModToKey(toKey string) EdgeQueryModFunc
```

#### func  EdgeModToType

```go
func EdgeModToType(toType string) EdgeQueryModFunc
```

#### func  EdgeModWhere

```go
func EdgeModWhere(where WhereFunc) EdgeQueryModFunc
```

#### type EdgeTriggerFunc

```go
type EdgeTriggerFunc func(g Graph, edge Edge, deleted bool) error
```

EdgeTriggerFunc is a function runs after edges are added to the graph. If an
error occurs, it will be returned but the edge will still be added to the graph.

#### type Encoder

```go
type Encoder interface {
	Marshal() ([]byte, error)
	Unmarshal(data []byte) error
}
```

Encoder can marshal to bytes and unmarshal itself from bytes

#### type ErrHandler

```go
type ErrHandler func(err error)
```

ErrHandler executes a function against the input error

#### func  DefaultErrHandler

```go
func DefaultErrHandler() ErrHandler
```
DefaultErrHandler simply logs the error

#### type Graph

```go
type Graph interface {
	// AddNode adds a node to the graph
	AddNode(ctx context.Context, n Node) error
	// QueryNodes executes the query against graph nodes
	QueryNodes(ctx context.Context, query NodeQuery) error
	// DelNode deletes a node by path
	DelNode(ctx context.Context, path Path) error
	// GetNode gets a node by path
	GetNode(ctx context.Context, path Path) (Node, error)
	// AddEdge adds an edge to the graph
	AddEdge(ctx context.Context, e Edge) error
	// GetEdge gets an edge from the graph
	GetEdge(ctx context.Context, path EdgePath) (Edge, error)
	// QueryEdges executes the query against graph edges
	QueryEdges(ctx context.Context, query EdgeQuery) error
	// DelEdge deletes the edge by path
	DelEdge(ctx context.Context, e Edge) error

	// Close closes the graph
	Close(ctx context.Context) error
}
```

Graph is a directed acyclic graph (DAG)

#### type GraphOpenerFunc

```go
type GraphOpenerFunc func() (Graph, error)
```

GraphOpenerFunc opens a Graph. Backends should export an opener method.

#### type Graphik

```go
type Graphik interface {
	Graph
	// NodeConstraints adds the node constraints to the graph
	NodeConstraints(constraints ...NodeTriggerFunc)
	// NodeTriggers adds the triggers to the graph
	NodeTriggers(triggers ...NodeTriggerFunc)
	NodeLabelers(labelers ...NodeTriggerFunc)

	// EdgeConstraints adds the edge constraints to the graph
	EdgeConstraints(constraints ...EdgeTriggerFunc)
	// NodeTriggers adds the node triggers to the graph
	EdgeTriggers(triggers ...EdgeTriggerFunc)
	EdgeLabelers(labelers ...EdgeTriggerFunc)

	StartWorkers(ctx context.Context)
	AddWorkers(workers ...Worker)
	StopWorker(ctx context.Context, name string)
	StopWorkers(ctx context.Context)
}
```

Graphik is a directed acyclic graph (DAG) that can run asynchronous workers
against itself

#### func  New

```go
func New(opener GraphOpenerFunc) (Graphik, error)
```
New returns a new Graphik instance with the provided opener. Backends should
export an opener method.

#### type Node

```go
type Node interface {
	Path
	Attributer
}
```

Node is the path to a node + its own custom attributes

#### func  NewNode

```go
func NewNode(path Path) Node
```

#### type NodeHandlerFunc

```go
type NodeHandlerFunc func(g Graph, n Node) error
```

NodeHandlerFunc that executes logic against a node in a graph

#### type NodeQuery

```go
type NodeQuery interface {
	// WHERE and a LIMIT clause
	BasicQuery
	// Query modifier for chaining
	Mod(fn NodeQueryModFunc) NodeQuery
	// Type returns the type of node to search for aka: person
	Type() string
	// Key returns the key of the node to search for aka: coleman_word
	Key() string
	// When a node is found, the handler is executed against it
	Handler() NodeHandlerFunc
	// Validate returns the query and an error if it's invalid
	Validate() (NodeQuery, error)
}
```

NodeQuery is used to find nodes and execute logic on them

#### func  NewNodeQuery

```go
func NewNodeQuery() NodeQuery
```

#### type NodeQueryModFunc

```go
type NodeQueryModFunc func(query NodeQuery) NodeQuery
```

NodeQueryModFunc modifies anode query

#### func  NodeModHandler

```go
func NodeModHandler(handler NodeHandlerFunc) NodeQueryModFunc
```

#### func  NodeModKey

```go
func NodeModKey(key string) NodeQueryModFunc
```

#### func  NodeModLimit

```go
func NodeModLimit(limit int) NodeQueryModFunc
```

#### func  NodeModType

```go
func NodeModType(typ string) NodeQueryModFunc
```

#### func  NodeModWhere

```go
func NodeModWhere(where WhereFunc) NodeQueryModFunc
```

#### type NodeTriggerFunc

```go
type NodeTriggerFunc func(g Graph, node Node, deleted bool) error
```

NodeTriggerFunc is a function runs after nodes are added to the graph. If an
error occurs, it will be returned but the node will still be added to the graph.

#### type Path

```go
type Path interface {
	Type() string
	Key() string
	PathString() string
}
```

Path is a type and a key string aka: type= person key = coleman_word

#### func  NewPath

```go
func NewPath(typ string, key string) Path
```

#### type WhereFunc

```go
type WhereFunc func(g Graph, a Attributer) bool
```

WhereFunc is a WHERE clause that returns true/false based on it's implementation

#### type Worker

```go
type Worker interface {
	Name() string
	Stop(ctx context.Context)
	Start(ctx context.Context, g Graphik)
}
```

Worker is an asynchronous process that executes logic against a graphik
instance. It can be stopped and started.

#### func  NewWorker

```go
func NewWorker(name string, work WorkerFunc, errHandler ErrHandler, every time.Duration) Worker
```

#### type WorkerFunc

```go
type WorkerFunc func(g Graphik) error
```

WorkerFunc executes logic against a graphik instance

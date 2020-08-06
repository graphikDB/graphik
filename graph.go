//go:generate godocdown -o DOCS.md
package graphik

import (
	"fmt"
	"log"
)

// Path is a type and a key string aka: type= person key = coleman_word
type Path interface {
	Type() string
	Key() string
}

// Encoder can marshal to bytes and unmarshal itself from bytes
type Encoder interface {
	Marshal() ([]byte, error)
	Unmarshal(data []byte) error
}

// Counter returns a count of something
type Counter interface {
	Count() int
}

// Attributer is a concurrency safe map[string]interface{}
type Attributer interface {
	// Counter returns the length of the map
	Counter
	// SetAttribute overwrites an k-v in the map
	SetAttribute(key string, val interface{})
	// GetAttribute returns the value if it exists or nil if it doesnt
	GetAttribute(key string) interface{}
	// Attributer can marshal/unmarshal itself
	Encoder
	// returns a human readable string
	fmt.Stringer
}

// Node is the path to a node + its own custom attributes
type Node interface {
	Path
	Attributer
}

// Edge is the path from one node to another with a relationship and attributes
type Edge interface {
	Attributer
	Relationship() string
	From() Path
	To() Path
	Reversed() Edge
}

// Graph is a directed acyclic graph (DAG)
type Graph interface {
	// AddNode adds a node to the graph
	AddNode(n Node) error
	// QueryNodes executes the query against graph nodes
	QueryNodes(query NodeQuery) error
	// DelNode deletes a node by path
	DelNode(path Path) error
	// GetNode gets a node by path
	GetNode(path Path) (Node, error)
	// NodeConstraints adds the node constraints to the graph
	NodeConstraints(constraints ...NodeConstraintFunc)
	// NodeTriggers adds the triggers to the graph
	NodeTriggers(triggers ...NodeTriggerFunc)

	// AddEdge adds an edge to the graph
	AddEdge(e Edge) error
	// GetEdge gets an edge from the graph
	GetEdge(from Path, relationship string, to Path) (Edge, error)
	// QueryEdges executes the query against graph edges
	QueryEdges(query EdgeQuery) error
	// DelEdge deletes the edge by path
	DelEdge(e Edge) error
	// EdgeConstraints adds the edge constraints to the graph
	EdgeConstraints(constraints ...EdgeConstraintFunc)
	// NodeTriggers adds the node triggers to the graph
	EdgeTriggers(triggers ...EdgeTriggerFunc)

	// Close closes the graph
	Close() error
}

// BasicQuery is a WHERE and a LIMIT clause
type BasicQuery interface {
	Where() WhereFunc
	Limit() int
}

// NodeQuery is used to find nodes and execute logic on them
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

// EdgeQuery is used to find edges and execute logic on them
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

// NodeTriggerFunc is a function runs after nodes are added to the graph. If an error occurs, it will be returned
// but the node will still be added to the graph.
type NodeTriggerFunc func(g Graph, node Node, deleted bool) error

// EdgeTriggerFunc is a function runs after edges are added to the graph. If an error occurs, it will be returned
// but the edge will still be added to the graph.
type EdgeTriggerFunc func(g Graph, edge Edge, deleted bool) error

// EdgeConstraintFunc is a function run before edges are added to the graph. If an error occurs, it will be returned
// and the edge will not be added to the graph
type EdgeConstraintFunc func(g Graph, edge Edge, toDelete bool) error

// EdgeHandlerFunc that executes logic against an edge in a graph
type EdgeHandlerFunc func(g Graph, e Edge) error

// NodeHandlerFunc that executes logic against a node in a graph
type NodeHandlerFunc func(g Graph, n Node) error

// NodeConstraintFunc is a function run before nodes are added to the database. If an error occurs, it will be returned
// and the node will not be added to the graph
type NodeConstraintFunc func(g Graph, node Node, toDelete bool) error

// WhereFunc is a WHERE clause that returns true/false based on it's implementation
type WhereFunc func(g Graph, a Attributer) bool

// EdgeQueryModFunc modifies an edge query
type EdgeQueryModFunc func(query EdgeQuery) EdgeQuery

// NodeQueryModFunc modifies anode query
type NodeQueryModFunc func(query NodeQuery) NodeQuery

// GraphOpenerFunc opens a Graph. Backends should export an opener method.
type GraphOpenerFunc func() (Graph, error)

// Graphik is a directed acyclic graph (DAG) that can run asynchronous workers against itself
type Graphik interface {
	Graph
	StartWorkers()
	AddWorkers(workers ...Worker)
	StopWorker(name string)
	StopWorkers()
}

// WorkerFunc executes logic against a graphik instance
type WorkerFunc func(g Graphik) error

// Worker is an asynchronous process that executes logic against a graphik instance. It can be stopped and started.
type Worker interface {
	Name() string
	Stop()
	Start(g Graphik)
}

// ErrHandler executes a function against the input error
type ErrHandler func(err error)

// DefaultErrHandler simply logs the error
func DefaultErrHandler() ErrHandler {
	return func(err error) {
		log.Printf("graphik: %s\n", err.Error())
	}
}

type graphik struct {
	Graph
	workers map[string]Worker
}

func (g *graphik) AddWorkers(workers ...Worker) {
	for _, worker := range workers {
		g.workers[worker.Name()] = worker
	}
}

func (g *graphik) StartWorkers() {
	for _, worker := range g.workers {
		worker.Start(g)
	}
}

func (g *graphik) StopWorker(name string) {
	if worker, ok := g.workers[name]; ok {
		worker.Stop()
	}
}

func (g *graphik) StopWorkers() {
	for _, worker := range g.workers {
		worker.Stop()
	}
}

// New returns a new Graphik instance with the provided opener. Backends should export an opener method.
func New(opener GraphOpenerFunc) (Graphik, error) {
	g, err := opener()
	if err != nil {
		return nil, err
	}
	return &graphik{
		Graph:   g,
		workers: map[string]Worker{},
	}, nil
}

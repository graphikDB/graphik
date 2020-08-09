//go:generate godocdown -o DOCS.md
package graphik

import (
	"context"
	"fmt"
	"log"
)

// Path is a type and a key string aka: type= person key = coleman_word
type Path interface {
	Type() string
	Key() string
	PathString() string
}

type EdgePath interface {
	Relationship() string
	From() Path
	To() Path
	PathString() string
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
	// Range iterates over the values. If false is returned, the range function will break.
	Range(fn func(k string, v interface{}) bool)
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
	EdgePath
}

// Graph is a directed acyclic graph (DAG)
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
	Closer() func() error
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
	Closer() func() error
}

// NodeTriggerFunc is a function runs after nodes are added to the graph. If an error occurs, it will be returned
// but the node will still be added to the graph.
type NodeTriggerFunc func(g Graph, node Node, deleted bool) error

// EdgeTriggerFunc is a function runs after edges are added to the graph. If an error occurs, it will be returned
// but the edge will still be added to the graph.
type EdgeTriggerFunc func(g Graph, edge Edge, deleted bool) error

// EdgeHandlerFunc that executes logic against an edge in a graph
type EdgeHandlerFunc func(g Graph, e Edge) error

// NodeHandlerFunc that executes logic against a node in a graph
type NodeHandlerFunc func(g Graph, n Node) error

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

// WorkerFunc executes logic against a graphik instance
type WorkerFunc func(g Graphik) error

// Worker is an asynchronous process that executes logic against a graphik instance. It can be stopped and started.
type Worker interface {
	Name() string
	Stop(ctx context.Context)
	Start(ctx context.Context, g Graphik)
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
	graph           Graph
	workers         map[string]Worker
	nodeTriggers    []NodeTriggerFunc
	edgeTriggers    []EdgeTriggerFunc
	nodeConstraints []NodeTriggerFunc
	edgeConstraints []EdgeTriggerFunc
	edgeLabelers    []EdgeTriggerFunc
	nodeLabelers    []NodeTriggerFunc
}

func (g *graphik) AddNode(ctx context.Context, n Node) error {
	for _, fn := range g.nodeConstraints {
		if err := fn(g, n, false); err != nil {
			return err
		}
	}
	for _, fn := range g.nodeLabelers {
		if err := fn(g, n, false); err != nil {
			return err
		}
	}
	if err := g.graph.AddNode(ctx, n); err != nil {
		return err
	}
	for _, fn := range g.nodeTriggers {
		if err := fn(g, n, false); err != nil {
			return err
		}
	}
	return nil
}

func (g *graphik) QueryNodes(ctx context.Context, query NodeQuery) error {
	return g.graph.QueryNodes(ctx, query)
}

func (g *graphik) DelNode(ctx context.Context, path Path) error {
	n, err := g.GetNode(ctx, path)
	if err != nil {
		return err
	}
	for _, fn := range g.nodeConstraints {
		if err := fn(g, n, true); err != nil {
			return err
		}
	}
	for _, fn := range g.nodeLabelers {
		if err := fn(g, n, true); err != nil {
			return err
		}
	}
	if err := g.graph.DelNode(ctx, n); err != nil {
		return err
	}
	for _, fn := range g.nodeTriggers {
		if err := fn(g, n, true); err != nil {
			return err
		}
	}
	return nil
}

func (g *graphik) GetNode(ctx context.Context, path Path) (Node, error) {
	return g.graph.GetNode(ctx, path)
}

func (g *graphik) AddEdge(ctx context.Context, e Edge) error {
	for _, fn := range g.edgeConstraints {
		if err := fn(g, e, false); err != nil {
			return err
		}
	}
	for _, fn := range g.edgeLabelers {
		if err := fn(g, e, false); err != nil {
			return err
		}
	}
	if err := g.graph.AddEdge(ctx, e); err != nil {
		return err
	}
	for _, fn := range g.edgeTriggers {
		if err := fn(g, e, false); err != nil {
			return err
		}
	}
	return nil
}

func (g *graphik) GetEdge(ctx context.Context, path EdgePath) (Edge, error) {
	return g.graph.GetEdge(ctx, path)
}

func (g *graphik) QueryEdges(ctx context.Context, query EdgeQuery) error {
	return g.graph.QueryEdges(ctx, query)
}

func (g *graphik) DelEdge(ctx context.Context, e Edge) error {
	for _, fn := range g.edgeConstraints {
		if err := fn(g, e, true); err != nil {
			return err
		}
	}
	for _, fn := range g.edgeLabelers {
		if err := fn(g, e, false); err != nil {
			return err
		}
	}
	if err := g.graph.DelEdge(ctx, e); err != nil {
		return err
	}
	for _, fn := range g.edgeTriggers {
		if err := fn(g, e, true); err != nil {
			return err
		}
	}
	return nil
}

func (g *graphik) Close(ctx context.Context) error {
	return g.graph.Close(ctx)
}

func (g *graphik) NodeConstraints(constraints ...NodeTriggerFunc) {
	g.nodeConstraints = append(g.nodeConstraints, constraints...)
}

func (g *graphik) NodeTriggers(triggers ...NodeTriggerFunc) {
	g.nodeTriggers = append(g.nodeTriggers, triggers...)
}

func (g *graphik) NodeLabelers(labelers ...NodeTriggerFunc) {
	g.nodeLabelers = append(g.nodeLabelers, labelers...)
}

func (g *graphik) EdgeConstraints(constraints ...EdgeTriggerFunc) {
	g.edgeConstraints = append(g.edgeConstraints, constraints...)
}

func (g *graphik) EdgeLabelers(labelers ...EdgeTriggerFunc) {
	g.edgeLabelers = append(g.edgeLabelers, labelers...)
}

func (g *graphik) EdgeTriggers(triggers ...EdgeTriggerFunc) {
	g.edgeTriggers = append(g.edgeTriggers, triggers...)
}

func (g *graphik) AddWorkers(workers ...Worker) {
	for _, worker := range workers {
		g.workers[worker.Name()] = worker
	}
}

func (g *graphik) StartWorkers(ctx context.Context) {
	for _, worker := range g.workers {
		worker.Start(ctx, g)
	}
}

func (g *graphik) StopWorker(ctx context.Context, name string) {
	if worker, ok := g.workers[name]; ok {
		worker.Stop(ctx)
	}
}

func (g *graphik) StopWorkers(ctx context.Context) {
	for _, worker := range g.workers {
		worker.Stop(ctx)
	}
}

// New returns a new Graphik instance with the provided opener. Backends should export an opener method.
func New(opener GraphOpenerFunc) (Graphik, error) {
	g, err := opener()
	if err != nil {
		return nil, err
	}
	return &graphik{
		graph:   g,
		workers: map[string]Worker{},
	}, nil
}

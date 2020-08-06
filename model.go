package graphik

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"sync"
	"time"
)

type attributer struct {
	mu         *sync.RWMutex
	attributes map[string]interface{}
}

func (a *attributer) init() {
	if a.mu == nil {
		a.mu = &sync.RWMutex{}
	}
	if a.attributes == nil {
		a.attributes = map[string]interface{}{}
	}
}

func NewAttributer(attr map[string]interface{}) Attributer {
	return &attributer{
		mu:         &sync.RWMutex{},
		attributes: attr,
	}
}

func (n *attributer) SetAttribute(key string, val interface{}) {
	n.init()
	n.mu.Lock()
	n.attributes[key] = val
	n.mu.Unlock()
}

func (n *attributer) GetAttribute(key string) interface{} {
	n.init()
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.attributes[key]
}

func (n *attributer) Marshal() ([]byte, error) {
	n.init()
	n.mu.RLock()
	defer n.mu.RUnlock()
	bits, err := bson.Marshal(n.attributes)
	if err != nil {
		return nil, err
	}
	return bits, nil
}

func (n *attributer) Unmarshal(data []byte) error {
	n.init()
	n.mu.Lock()
	defer n.mu.Unlock()
	attr := map[string]interface{}{}
	if err := bson.Unmarshal(data, &attr); err != nil {
		return err
	}
	n.attributes = attr
	return nil
}

func (n *attributer) String() string {
	n.init()
	n.mu.RLock()
	defer n.mu.RUnlock()
	bits, _ := json.MarshalIndent(&n.attributes, "", "    ")
	return string(bits)
}

func (n *attributer) Range(fn func(k string, v interface{}) bool) {
	n.init()
	n.mu.RLock()
	defer n.mu.RUnlock()
	for k, v := range n.attributes {
		if !fn(k, v) {
			break
		}
	}
}

func (a *attributer) Count() int {
	a.init()
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.attributes)
}

func (a *attributer) implements() Attributer {
	return a
}

type path struct {
	typ string
	key string
}

func NewPath(typ string, key string) Path {
	return &path{typ: typ, key: key}
}

func (p *path) Type() string {
	return p.typ
}

func (p *path) Key() string {
	return p.key
}

func (p *path) implements() Path {
	return p
}

type node struct {
	Attributer
}

func NewNode(path Path, attr Attributer) Node {
	if attr == nil {
		attr = NewAttributer(map[string]interface{}{
			"type": path.Type(),
			"key":  path.Key(),
		})
	}
	return &node{Attributer: attr}
}

func (n *node) implements() Node {
	return n
}

func (n *node) Type() string {
	return n.GetAttribute("type").(string)
}

func (n *node) Key() string {
	return n.GetAttribute("key").(string)
}

// Edge is a simple graph edge.
type edge struct {
	Attributer
}

func NewEdge(from Path, relationship string, to Path, attr Attributer) Edge {
	if attr == nil {
		attr = NewAttributer(map[string]interface{}{
			"fromType":     from.Type(),
			"fromKey":      from.Key(),
			"relationship": relationship,
			"toType":       to.Type(),
			"toKey":        to.Key(),
		})
	}
	return &edge{
		Attributer: NewAttributer(map[string]interface{}{
			"fromType":     from.Type(),
			"fromKey":      from.Key(),
			"relationship": relationship,
			"toType":       to.Type(),
			"toKey":        to.Key(),
		}),
	}
}

func (e *edge) implements() Edge {
	return e
}

func (e *edge) Relationship() string {
	return e.GetAttribute("relationship").(string)
}

func (e *edge) From() Path {
	return &path{
		typ: e.GetAttribute("fromType").(string),
		key: e.GetAttribute("fromKey").(string),
	}
}

func (e *edge) To() Path {
	return &path{
		typ: e.GetAttribute("toType").(string),
		key: e.GetAttribute("toKey").(string),
	}
}

func (e *edge) Reversed() Edge {
	return NewEdge(e.To(), e.Relationship(), e.From(), e)
}

type worker struct {
	wg         *sync.WaitGroup
	worker     WorkerFunc
	errHandler ErrHandler
	every      time.Duration
	name       string
	done       chan struct{}
}

func NewWorker(name string, work WorkerFunc, errHandler ErrHandler, every time.Duration) Worker {
	if errHandler == nil {
		errHandler = DefaultErrHandler()
	}
	return &worker{
		wg:         &sync.WaitGroup{},
		worker:     work,
		errHandler: errHandler,
		every:      every,
		name:       name,
		done:       make(chan struct{}, 1),
	}
}

func (w *worker) HandleError(err error) {
	w.errHandler(err)
}

func (w *worker) Start(g Graphik) {
	ticker := time.NewTicker(w.every)
	w.wg.Add(1)

	go func() {
		defer ticker.Stop()
		defer w.wg.Done()
		for {
			select {
			case <-ticker.C:
				w.wg.Add(1)
				go func() {
					defer w.wg.Done()
					if err := w.worker(g); err != nil {
						w.errHandler(err)
					}
				}()
			case <-w.done:
				return
			}
		}
	}()
}

func (w *worker) Name() string {
	return w.name
}

func (w *worker) Stop() {
	w.done <- struct{}{}
	w.wg.Wait()
}

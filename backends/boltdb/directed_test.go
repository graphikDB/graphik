package boltdb_test

import (
	"fmt"
	"github.com/autom8ter/graphik"
	"github.com/autom8ter/graphik/backends/boltdb"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func Test(t *testing.T) {
	os.RemoveAll("/tmp/graphik")
	graph, err := graphik.New(boltdb.Open(boltdb.DefaultPath))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer graph.Close()

	for i := 0; i < 10; i++ {
		node := graphik.NewNode(graphik.NewPath("user", fmt.Sprintf("%s-%d", "cword3", time.Now().UnixNano())), nil)
		node.SetAttribute("name", "coleman")
		if err := graph.AddNode(node); err != nil {
			t.Fatal(err.Error())
		}
		t.Log(node.String())
		node2 := graphik.NewNode(graphik.NewPath("user", fmt.Sprintf("%s-%d", "twash2", time.Now().UnixNano())), nil)
		node2.SetAttribute("name", "tyler")
		if err := graph.AddNode(node2); err != nil {
			t.Fatal(err.Error())
		}
		t.Log(node2.String())
		edge := graphik.NewEdge(node, "friend", node2, nil)
		if err := graph.AddEdge(edge); err != nil {
			t.Fatal(err.Error())
		}
		fmt.Println(edge.String())
		edge2 := graphik.NewEdge(node, "groomsman", node2, nil)
		if err := graph.AddEdge(edge2); err != nil {
			t.Fatal(err.Error())
		}
		fmt.Println(edge2.String())
	}
}

func Test2(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.RemoveAll(tmpdir)
	graph, err := graphik.New(boltdb.Open(tmpdir))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer graph.Close()
	node := graphik.NewNode(graphik.NewPath("user", fmt.Sprintf("%s-%d", "cword3", time.Now().UnixNano())), nil)
	node.SetAttribute("name", "coleman")
	if err := graph.AddNode(node); err != nil {
		t.Fatal(err.Error())
	}
	node2 := graphik.NewNode(graphik.NewPath("user", fmt.Sprintf("%s-%d", "twash2", time.Now().UnixNano())), nil)
	node2.SetAttribute("name", "tyler")
	if err := graph.AddNode(node2); err != nil {
		t.Fatal(err.Error())
	}
	for i := 0; i < 100; i++ {
		edge := graphik.NewEdge(node, "friends", node2, nil)
		edge.SetAttribute("testing", true)
		if err := graph.AddEdge(edge); err != nil {
			t.Fatal(err.Error())
		}
	}
	nquery, err := graphik.NewNodeQuery().
		Mod(graphik.NodeModLimit(1)).
		Mod(graphik.NodeModType("user")).
		Mod(graphik.NodeModHandler(func(g graphik.Graph, n graphik.Node) error {
			t.Logf("node = %s", n.String())
			return nil
		})).
		Mod(graphik.NodeModWhere(func(g graphik.Graph, a graphik.Attributer) bool {
			if a.GetAttribute("name") == "coleman" {
				return true
			}
			return false
		})).Validate()
	if err != nil {
		t.Fatal(err.Error())
	}
	now1 := time.Now()
	if err := graph.QueryNodes(nquery); err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("node query time: %s", time.Since(now1).String())

	equery, err := graphik.NewEdgeQuery().
		Mod(graphik.EdgeModLimit(1)).
		Mod(graphik.EdgeModFromType("user")).
		Mod(graphik.EdgeModRelationship("friends")).
		Mod(graphik.EdgeModWhere(func(g graphik.Graph, a graphik.Attributer) bool {
			if a.GetAttribute("testing") == true {
				return true
			}
			return false
		})).
		Mod(graphik.EdgeModHandler(func(g graphik.Graph, e graphik.Edge) error {
			t.Logf("edge= %s", e.String())
			return nil
		})).Validate()
	if err != nil {
		t.Fatal(err.Error())
	}
	now2 := time.Now()
	if err := graph.QueryEdges(equery); err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("edge query time: %s", time.Since(now2).String())
}

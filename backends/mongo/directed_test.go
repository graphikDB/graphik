package mongo_test

import (
	"context"
	"fmt"
	"github.com/autom8ter/graphik"
	"github.com/autom8ter/graphik/backends/mongo"
	"reflect"
	"testing"
	"time"
)

const mongoString = "mongodb://localhost:27017"

func TestMongo(t *testing.T) {
	graph, err := graphik.New(mongo.Open(mongoString))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer graph.Close(context.Background())

	for i := 0; i < 5; i++ {
		node := graphik.NewNode(graphik.NewPath("user", fmt.Sprintf("%s-%d", "cword3", time.Now().UnixNano())))
		node.SetAttribute("name", "coleman")
		if err := graph.AddNode(context.Background(), node); err != nil {
			t.Fatal(err.Error())
		}

		t.Log(node.String())
		node2 := graphik.NewNode(graphik.NewPath("user", fmt.Sprintf("%s-%d", "twash2", time.Now().UnixNano())))
		node2.SetAttribute("name", "tyler")
		if err := graph.AddNode(context.Background(), node2); err != nil {
			t.Fatal(err.Error())
		}
		t.Log(node2.String())
		edge := graphik.NewEdge(graphik.NewEdgePath(node, "friend", node2))
		if err := graph.AddEdge(context.Background(), edge); err != nil {
			t.Fatal(err.Error())
		}
		fmt.Println(edge.String())
		edge2 := graphik.NewEdge(graphik.NewEdgePath(node, "groomsman", node2))
		if err := graph.AddEdge(context.Background(), edge2); err != nil {
			t.Fatal(err.Error())
		}
		fmt.Println(edge2.String())

		n, err := graph.GetNode(context.Background(), node)
		if err != nil {
			t.Fatal(err.Error())
		}
		if !reflect.DeepEqual(n, node) {
			t.Fatalf("unequal nodes\n got: %s expected: %s", node.String(), n.String())
		}
		e, err := graph.GetEdge(context.Background(), edge)
		if err != nil {
			t.Fatal(err.Error())
		}
		if !reflect.DeepEqual(e, edge) {
			t.Fatalf("unequal edges\n got: %s expected: %s", e.String(), edge.String())
		}
	}
}

package graph_test

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/graph"
	"google.golang.org/protobuf/types/known/structpb"
	"os"
	"testing"
)

func Test(t *testing.T) {
	os.MkdirAll("/tmp/testing", 0700)
	defer os.Remove("/tmp/testing")
	g, err := graph.New("/tmp/testing/graph.db")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer g.Close()
	attributes, err := structpb.NewStruct(map[string]interface{}{
		"testing": true,
		"title":   "THIS IS A TEST",
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err := g.SetNode(context.Background(), &apipb.Node{
		Path: &apipb.Path{
			Gtype: "note",
		},
		Attributes: attributes,
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	if n.GetAttributes().GetFields()["testing"].GetBoolValue() != true {
		t.Fatal("expecting testing=true")
	}
}

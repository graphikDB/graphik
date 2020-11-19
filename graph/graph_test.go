package graph_test

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/graph"
	"os"
	"testing"
	"time"
)

func Test(t *testing.T) {
	os.MkdirAll("/tmp/testing", 0700)
	defer os.Remove("/tmp/testing")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	g, err := graph.New("/tmp/testing/graph.db")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer g.Close()
	human, err := g.SetNode(ctx, &apipb.Node{
		Path: &apipb.Path{
			Gtype: "human",
		},
		Attributes: apipb.NewStruct(map[string]interface{}{
			"testing": true,
			"name":    "coleman",
		}),
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	if human.GetAttributes().GetFields()["name"].GetStringValue() != "coleman" {
		t.Fatal("expecting name=coleman")
	}
	note, err := g.SetNode(ctx, &apipb.Node{
		Path: &apipb.Path{
			Gtype: "note",
		},
		Attributes: apipb.NewStruct(map[string]interface{}{
			"testing": true,
			"title":   "THIS IS A TEST",
		}),
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	if note.GetAttributes().GetFields()["testing"].GetBoolValue() != true {
		t.Fatal("expecting testing=true")
	}
	owner, err := g.SetEdge(ctx, &apipb.Edge{
		Path: &apipb.Path{
			Gtype: "owner",
		},
		Attributes: apipb.NewStruct(map[string]interface{}{
			"testing": true,
			"primary": true,
		}),
		Cascade: apipb.Cascade_CASCADE_NONE,
		From:    note.Path,
		To:      human.Path,
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	if owner.GetAttributes().GetFields()["testing"].GetBoolValue() != true {
		t.Fatal("expecting testing=true")
	}
}

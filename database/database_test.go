package database

import (
	"context"
	"fmt"
	"github.com/autom8ter/graphik/flags"
	apipb "github.com/autom8ter/graphik/gen/go/api"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
)

func init() {
	graph, err = NewGraph(context.Background(), &flags.Flags{
		OpenIDConnect:  "",
		StoragePath:    "/tmp/testing/graphik",
		Metrics:        false,
		Authorizers:    nil,
		AllowedHeaders: nil,
		AllowedMethods: nil,
		AllowedOrigins: nil,
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

var (
	graph *Graph
	err   error
	ctx   = context.WithValue(context.Background(), authCtxKey, &apipb.Doc{
		Path: &apipb.Path{
			Gtype: identityType,
			Gid:   4,
		},
		Attributes: apipb.NewStruct(map[string]interface{}{}),
		Metadata: &apipb.Metadata{
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
			CreatedBy: nil,
			UpdatedBy: nil,
			Version:   0,
		},
	})
)

func TestGraph_Ping(t *testing.T) {
	res, err := graph.Ping(ctx, &empty.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Message != "PONG" {
		t.Fatal("expected PONG")
	}
}

var (
	ashe       *apipb.Doc
	pikachu    *apipb.Doc
	charmander *apipb.Doc
	bulbasaur  *apipb.Doc
	squirtle   *apipb.Doc
)

func TestGraph_SetIndex(t *testing.T) {
	_, err := graph.SetIndex(ctx, &apipb.Index{
		Name:       "testing",
		Gtype:      "pokemon",
		Expression: `doc.attributes.type.contains("fire")`,
		Docs:       true,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = graph.SetIndex(ctx, &apipb.Index{
		Name:        "testing2",
		Gtype:       "owner",
		Expression:  `connection.attributes.primary`,
		Connections: true,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGraph_CreateDocs(t *testing.T) {
	res, err := graph.CreateDocs(ctx, &apipb.DocConstructors{
		Docs: []*apipb.DocConstructor{
			{
				Gtype: "trainer",
				Attributes: apipb.NewStruct(map[string]interface{}{
					"name": "ashe",
				}),
			},
			{
				Gtype: "pokemon",
				Attributes: apipb.NewStruct(map[string]interface{}{
					"name": "pikachu",
				}),
			},
			{
				Gtype: "pokemon",
				Attributes: apipb.NewStruct(map[string]interface{}{
					"name": "charmander",
					"type": "fire",
				}),
			},
			{
				Gtype: "pokemon",
				Attributes: apipb.NewStruct(map[string]interface{}{
					"name": "bulbasaur",
				}),
			},
			{
				Gtype: "pokemon",
				Attributes: apipb.NewStruct(map[string]interface{}{
					"name": "squirtle",
				}),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range res.GetDocs() {
		if err := r.Validate(); err != nil {
			t.Fatal(err)
		}
		switch r.GetAttributes().GetFields()["name"].GetStringValue() {
		case "ashe":
			ashe = r
		case "bulbasaur":
			bulbasaur = r
		case "pikachu":
			pikachu = r
		case "charmander":
			charmander = r
		case "squirtle":
			squirtle = r
		}
	}
}

func TestGraph_CreateConnections(t *testing.T) {
	res, err := graph.CreateConnections(ctx, &apipb.ConnectionConstructors{
		Connections: []*apipb.ConnectionConstructor{
			{
				Gtype: "owner",
				Attributes: apipb.NewStruct(map[string]interface{}{
					"primary": true,
				}),
				From:     ashe.GetPath(),
				Directed: true,
				To:       pikachu.GetPath(),
			},
			{
				Gtype: "owner",
				Attributes: apipb.NewStruct(map[string]interface{}{
					"primary": true,
				}),
				From:     ashe.GetPath(),
				Directed: true,
				To:       charmander.GetPath(),
			},
			{
				Gtype: "owner",
				Attributes: apipb.NewStruct(map[string]interface{}{
					"primary": true,
				}),
				From:     ashe.GetPath(),
				Directed: true,
				To:       bulbasaur.GetPath(),
			},
			{
				Gtype: "owner",
				Attributes: apipb.NewStruct(map[string]interface{}{
					"primary": true,
				}),
				From:     ashe.GetPath(),
				Directed: true,
				To:       squirtle.GetPath(),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range res.GetConnections() {
		if err := r.Validate(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGraph_SearchDocs(t *testing.T) {
	res, err := graph.SearchDocs(ctx, &apipb.Filter{
		Gtype:      "pokemon",
		Expression: `doc.attributes.name.contains("charmander")`,
		Limit:      1,
		Index:      "testing",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.GetDocs()) != 1 {
		t.Fatal("expected to find one document")
	}
	t.Log(res.GetDocs()[0].String())
}

func TestGraph_SearchConnections(t *testing.T) {
	res, err := graph.SearchConnections(ctx, &apipb.Filter{
		Gtype: "owner",
		Limit: 5,
		Index: "testing2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.GetConnections()) != 4 {
		t.Fatal("expected to find one connection", len(res.GetConnections()))
	}
}

func TestGraph_PatchDoc(t *testing.T) {
	res, err := graph.PatchDoc(ctx, &apipb.Patch{
		Path: ashe.GetPath(),
		Attributes: apipb.NewStruct(map[string]interface{}{
			"age": 13,
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.GetAttributes().GetFields()["age"].GetNumberValue() != 13 {
		t.Fatal("expected ashe to be 13 years old")
	}
}

func TestGraph_AllDocs(t *testing.T) {
	res, err := graph.AllDocs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.GetDocs()) != 5 {
		t.Fatal("expected 5 docs:", len(res.GetDocs()))
	}
	for _, r := range res.GetDocs() {
		if err := r.Validate(); err != nil {
			t.Fatal(err)
		}
		t.Log(r.String())
	}
}

func TestGraph_AllConnections(t *testing.T) {
	res, err := graph.AllConnections(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.GetConnections()) != 4 {
		t.Fatal("expected 4 connections:", len(res.GetConnections()))
	}
	for _, r := range res.GetConnections() {
		if err := r.Validate(); err != nil {
			t.Fatal(err)
		}
		t.Log(r.String())
	}
}

func TestGraph_GetSchema(t *testing.T) {
	res, err := graph.GetSchema(ctx, &empty.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res.String())
}

func TestGraph_DelDocs(t *testing.T) {
	res, err := graph.AllDocs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range res.GetDocs() {
		_, err := graph.DelDoc(ctx, r.GetPath())
		if err != nil {
			t.Fatal(err)
		}
	}
	res, err = graph.AllDocs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.GetDocs()) != 0 {
		t.Fatal("expected 0 docs")
	}
}

func TestGraph_DelConnections(t *testing.T) {
	res, err := graph.AllConnections(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range res.GetConnections() {
		_, err := graph.DelDoc(ctx, r.GetPath())
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(res.GetConnections()) != 0 {
		t.Fatal("expected 0 connections")
	}
}

func TestGraph_Close(t *testing.T) {
	graph.Close()
}
package database

import (
	"context"
	"fmt"
	apipb "github.com/autom8ter/graphik/gen/go"
	apipb2 "github.com/autom8ter/graphik/gen/go"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"testing"
)

func init() {
	err := os.RemoveAll("/tmp/testing")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	graph, err = NewGraph(context.Background(), &apipb.Flags{
		OpenIdDiscovery: "",
		StoragePath:     "/tmp/testing/graphik",
		Metrics:         false,
		AllowHeaders:    nil,
		AllowMethods:    nil,
		AllowOrigins:    nil,
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
			Gid:   ksuid.New().String(),
		},
		Attributes: apipb2.NewStruct(map[string]interface{}{}),
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
	_, err := graph.SetIndexes(ctx, &apipb.Indexes{Indexes: []*apipb.Index{
		{
			Name:       "testing",
			Gtype:      "pokemon",
			Expression: `doc.attributes.type.contains("fire")`,
			Docs:       true,
		},
		{
			Name:        "testing2",
			Gtype:       "owner",
			Expression:  `connection.attributes.primary`,
			Connections: true,
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGraph_SetAuthorizers(t *testing.T) {
	_, err := graph.SetAuthorizers(ctx, &apipb.Authorizers{Authorizers: []*apipb.Authorizer{
		{
			Name:       "testing",
			Expression: `request.method.contains("test")`,
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGraph_CreateDocs(t *testing.T) {
	res, err := graph.CreateDocs(ctx, &apipb.DocConstructors{
		Docs: []*apipb.DocConstructor{
			{
				Path: &apipb.PathConstructor{Gtype: "trainer"},
				Attributes: apipb2.NewStruct(map[string]interface{}{
					"name": "ashe",
				}),
			},
			{
				Path: &apipb.PathConstructor{Gtype: "pokemon"},
				Attributes: apipb2.NewStruct(map[string]interface{}{
					"name": "pikachu",
				}),
			},
			{
				Path: &apipb.PathConstructor{Gtype: "pokemon"},
				Attributes: apipb2.NewStruct(map[string]interface{}{
					"name": "charmander",
					"type": "fire",
				}),
			},
			{
				Path: &apipb.PathConstructor{Gtype: "pokemon"},
				Attributes: apipb2.NewStruct(map[string]interface{}{
					"name": "bulbasaur",
				}),
			},
			{
				Path: &apipb.PathConstructor{Gtype: "pokemon"},
				Attributes: apipb2.NewStruct(map[string]interface{}{
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
				Path: &apipb.PathConstructor{Gtype: "owner"},
				Attributes: apipb2.NewStruct(map[string]interface{}{
					"primary": true,
				}),
				From:     ashe.GetPath(),
				Directed: true,
				To:       pikachu.GetPath(),
			},
			{
				Path: &apipb.PathConstructor{Gtype: "owner"},
				Attributes: apipb2.NewStruct(map[string]interface{}{
					"primary": true,
				}),
				From:     ashe.GetPath(),
				Directed: true,
				To:       charmander.GetPath(),
			},
			{
				Path: &apipb.PathConstructor{Gtype: "owner"},
				Attributes: apipb2.NewStruct(map[string]interface{}{
					"primary": true,
				}),
				From:     ashe.GetPath(),
				Directed: true,
				To:       bulbasaur.GetPath(),
			},
			{
				Path: &apipb.PathConstructor{Gtype: "owner"},
				Attributes: apipb2.NewStruct(map[string]interface{}{
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

func TestGraph_GetDoc(t *testing.T) {
	res, err := graph.GetDoc(ctx, ashe.GetPath())
	if err != nil {
		t.Fatal(err)
	}
	if err := res.Validate(); err != nil {
		t.Fatal(err)
	}
	if res.String() != ashe.String() {
		t.Fatal("expected ashe")
	}
}

func TestGraph_GetDocDetail(t *testing.T) {
	res, err := graph.GetDocDetail(ctx, &apipb.DocDetailFilter{
		Path: ashe.GetPath(),
		FromConnections: &apipb.Filter{
			Gtype: "owner",
		},
		ToConnections: &apipb.Filter{
			Gtype: "owner",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := res.Validate(); err != nil {
		t.Fatal(err)
	}
	//t.Log(res.String())
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
	//t.Log(res.GetDocs()[0].String())
}

func TestGraph_DepthSearchDocs(t *testing.T) {
	res, err := graph.DepthSearchDocs(ctx, &apipb.DepthFilter{
		Root: ashe.GetPath(),
		//DocExpression:        `doc.path.gtype.contains("pokemon")`,
		ConnectionExpression: "connection.attributes.secondary",
		Limit:                4,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, doc := range res.GetTraversals() {
		t.Log(doc.String())
	}
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

func TestGraph_EditDoc(t *testing.T) {
	res, err := graph.EditDoc(ctx, &apipb.Edit{
		Path: ashe.GetPath(),
		Attributes: apipb2.NewStruct(map[string]interface{}{
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
		//	t.Log(r.String())
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
		//	t.Log(r.String())
	}
}

func TestGraph_GetSchema(t *testing.T) {
	schema, err := graph.GetSchema(ctx, &empty.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.GetAuthorizers().GetAuthorizers()) == 0 {
		t.Fatal("expected at least one authorizer")
	}
	if len(schema.GetIndexes().GetIndexes()) == 0 {
		t.Fatal("expected at least one index")
	}
	t.Log(schema.String())
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

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

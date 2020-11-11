package main

import (
	"context"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"testing"
	"time"
)

func init() {
	godotenv.Load()
	j, err := google.DefaultTokenSource(context.Background(), "https://www.googleapis.com/auth/devstorage.full_control")
	if err != nil {
		log.Print(err)
		return
	}
	token, err := j.Token()
	if err != nil {
		log.Print(err)
		return
	}
	id := token.Extra("id_token")
	ctx = metadata.AppendToOutgoingContext(context.Background(), "Authorization", fmt.Sprintf("Bearer %v", id))
}

var ctx context.Context

func Test(t *testing.T) {
	time.Sleep(3 * time.Second)
	conn, err := grpc.DialContext(ctx, "localhost:7820", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	var (
		cfgClient = apipb.NewConfigServiceClient(conn)
		gClient   = apipb.NewGraphServiceClient(conn)
	)
	pong, err := cfgClient.Ping(ctx, &empty.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	if pong.Message != "PONG" {
		t.Fatal("not PONG")
	}
	me, err := gClient.Me(ctx, &empty.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(me.String())

	note, err := gClient.CreateNode(ctx, &apipb.Node{
		Path: &apipb.Path{
			Gtype: "note",
		},
		Attributes: apipb.NewStruct(map[string]interface{}{
			"title":       "this is a test note",
			"description": "testing creating a node of type note",
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = gClient.CreateEdge(ctx, &apipb.Edge{
		Path: &apipb.Path{
			Gtype: "personal_notes",
		},
		Attributes: apipb.NewStruct(map[string]interface{}{
			"weight": 5,
		}),
		Cascade: apipb.Cascade_CASCADE_TO,
		From:    me.Path,
		To:      note.Path,
	})
	if err != nil {
		t.Fatal(err)
	}
	edges, err := gClient.EdgesFrom(ctx, &apipb.EdgeFilter{
		NodePath: me.Path,
		Gtype:    "personal_notes",
		//Expressions:          []string{
		//	`attributes.weight > 3`,
		//},
		Limit:     1,
		MaxDegree: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(edges.GetEdges()) == 0 {
		t.Fatal("zero edges found")
	}
	for _, e := range edges.GetEdges() {
		t.Log(e.String())
	}
}

func Benchmark(b *testing.B) {
	b.ReportAllocs()
	conn, err := grpc.DialContext(ctx, "localhost:7820", grpc.WithInsecure())
	if err != nil {
		b.Fatal(err)
	}
	var (
		gClient = apipb.NewGraphServiceClient(conn)
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gClient.SearchNodes(ctx, &apipb.TypeFilter{
			Gtype:       apipb.Keyword_ANY.String(),
			Expressions: []string{`attributes.name.contains("cole")`},
			Limit:       1,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

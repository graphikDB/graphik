package main

import (
	"context"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"testing"
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
//
//func Test(t *testing.T) {
//	time.Sleep(3 * time.Second)
//	conn, err := grpc.DialContext(ctx, "localhost:7820", grpc.WithInsecure())
//	if err != nil {
//		t.Fatal(err)
//	}
//	var (
//		cfgClient = apipb.NewConfigServiceClient(conn)
//		gClient   = apipb.NewGraphServiceClient(conn)
//	)
//	pong, err := cfgClient.Ping(ctx, &empty.Empty{})
//	if err != nil {
//		t.Fatal(err)
//	}
//	if pong.Message != "PONG" {
//		t.Fatal("not PONG")
//	}
//	nodes, err := gClient.SearchNodes(ctx, &apipb.TypeFilter{
//		Gtype:       "identity",
//		Expressions: []string{`attributes.email.contains("coleman")`},
//		Limit:       1,
//	})
//	if err != nil {
//		t.Fatal(err)
//	}
//	for _, n := range nodes.GetNodes() {
//		t.Log(n.String())
//	}
//
//	select {
//	case <-ctx.Done():
//		t.Log("done")
//	}
//}

func Benchmark(b *testing.B) {
	b.ReportAllocs()
	conn, err := grpc.DialContext(ctx, "localhost:7820", grpc.WithInsecure())
	if err != nil {
		b.Fatal(err)
	}
	var (
		gClient   = apipb.NewGraphServiceClient(conn)
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node, err := gClient.CreateNode(ctx, &apipb.Node{
			Path:                 &apipb.Path{
				Gtype:                "test-user",
			},
			Attributes:           apipb.NewStruct(map[string]interface{}{
				"name": "coleman",
			}),
		})
		if err != nil {
			b.Fatal(err)
		}
		b.Log(node.Path.String())
	}
}


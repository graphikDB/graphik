package main

import (
	"context"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/lang"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"testing"
	"time"
)

func init() {
	godotenv.Load()
}

func Test(t *testing.T) {
	cfg := &apipb.Config{
		Auth: &apipb.AuthConfig{
			JwksSources: []string{"https://www.googleapis.com/oauth2/v3/certs"},
		},
	}
	cfg.SetDefaults()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	go func() {

		defer cancel()
		run(ctx, cfg)
	}()
	time.Sleep(3 * time.Second)
	j, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/devstorage.full_control")
	if err != nil {
		t.Fatal(err)
	}
	token, err := j.Token()
	if err != nil {
		t.Fatal(err)
	}
	id := token.Extra("id_token")

	t.Logf("token= %s", token)
	ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %v", id))
	conn, err := grpc.DialContext(ctx, "localhost:7820", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	client := apipb.NewPrivateServiceClient(conn)
	pong, err := client.Ping(ctx, &empty.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	if pong.Message != "PONG" {
		t.Fatal("not PONG")
	}
	nodes, err := client.CreateNodes(ctx, &apipb.Nodes{
		Nodes: []*apipb.Node{
			{
				Path: "pet",
				Attributes: lang.ToStruct(map[string]interface{}{
					"name": "charlie",
				}),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(nodes.Nodes[0].String())
	select {
	case <-ctx.Done():
		t.Log("done")
	}
}

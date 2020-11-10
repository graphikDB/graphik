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
	"testing"
	"time"
)

func init() {
	godotenv.Load()
	cfg := &apipb.Config{
		Auth: &apipb.AuthConfig{
			JwksSources: []string{"https://www.googleapis.com/oauth2/v3/certs"},
		},
	}
	cfg.SetDefaults()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	go func() {
		defer cancel()
		run(ctx, cfg)
	}()
}

func Test(t *testing.T) {
	time.Sleep(5 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
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
	nodes, err := gClient.SearchNodes(ctx, &apipb.TypeFilter{
		Type:        "identity",
		Expressions: nil,
		Limit:       1,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range nodes.GetNodes() {
		t.Log(n.String())
	}

	select {
	case <-ctx.Done():
		t.Log("done")
	}
}

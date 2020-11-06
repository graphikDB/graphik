package main

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"os"
	"testing"
	"time"
)

func init() {
	godotenv.Load()
}

func Test(t *testing.T) {
	cfg := &apipb.Config{}
	cfg.SetDefaults()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	go func() {

		defer cancel()
		run(ctx, cfg)
	}()
	time.Sleep(1 * time.Second)
	config := clientcredentials.Config{
		ClientID:       os.Getenv("CLIENT_ID"),
		ClientSecret:   os.Getenv("CLIENT_SECRET"),
		TokenURL:       google.Endpoint.TokenURL,
		Scopes:         nil,
		EndpointParams: nil,
	}
	token, err := config.Token(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token.AccessToken)
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
	select {
	case <-ctx.Done():
		t.Log("done")
	}
}

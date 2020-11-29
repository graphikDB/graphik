package graphik_test

import (
	"context"
	"fmt"
	"github.com/autom8ter/graphik/gen/go/api"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/graphik/sdk/graphik-client-go"
	"github.com/autom8ter/machine"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"log"
	"strings"
	"time"
)

func init() {
	ctx := context.Background()

	// ensure graphik server is started with --auth.jwks=https://www.googleapis.com/oauth2/v3/certs
	tokenSource, err := google.DefaultTokenSource(context.Background(), "https://www.googleapis.com/auth/devstorage.full_control")
	if err != nil {
		log.Print(err)
		return
	}

	client, err = graphik.NewClient(ctx, "localhost:7820", graphik.WithTokenSource(tokenSource))
	if err != nil {
		log.Print(err)
		return
	}
}

var client *graphik.Client

func ExampleNewClient() {
	ctx := context.Background()

	// ensure graphik server is started with --auth.jwks=https://www.googleapis.com/oauth2/v3/certs
	source, err := google.DefaultTokenSource(context.Background(), "https://www.googleapis.com/auth/devstorage.full_control")
	if err != nil {
		log.Print(err)
		return
	}

	cli, err := graphik.NewClient(ctx, "localhost:7820", graphik.WithTokenSource(source))
	if err != nil {
		log.Print(err)
		return
	}
	pong, err := cli.Ping(ctx, &empty.Empty{})
	if err != nil {
		logger.Error("failed to ping server", zap.Error(err))
		return
	}
	fmt.Println(pong.Message)
	// Output: PONG
}

func ExampleClient_Me() {
	me, err := client.Me(context.Background(), &apipb.MeFilter{
		ConnectionsFrom: nil,
		ConnectionsTo:   nil,
	})
	if err != nil {
		log.Print(err)
		return
	}
	issuer := me.GetAttributes().GetFields()["iss"].GetStringValue() // token issuer
	fmt.Println(issuer)
	// Output: https://accounts.google.com
}

func ExampleClient_CreateDoc() {
	charlie, err := client.CreateDoc(context.Background(), &apipb.DocConstructor{
		Path: &apipb.Path{
			Gtype: "dog",
		},
		Attributes: apipb.NewStruct(map[string]interface{}{
			"name": "Charlie",
		}),
	})
	if err != nil {
		log.Print(err)
		return
	}
	name := charlie.Attributes.Fields["name"].GetStringValue()
	fmt.Println(name)
	// Output: Charlie
}

func ExampleClient_SearchDocs() {
	dogs, err := client.SearchDocs(context.Background(), &apipb.Filter{
		Gtype:      "dog",
		Expression: `doc.attributes.name.contains("Charl")`,
		Limit:      1,
		Sort:       "metadata.created_at",
	})
	if err != nil {
		log.Print(err)
		return
	}
	dog := dogs.GetDocs()[0]
	name := dog.Attributes.Fields["name"].GetStringValue()
	fmt.Println(name)
	// Output: Charlie
}

func ExampleClient_CreateConnection() {
	dogs, err := client.SearchDocs(context.Background(), &apipb.Filter{
		Gtype:      "dog",
		Expression: `doc.attributes.name.contains("Charl")`,
		Limit:      1,
	})
	if err != nil {
		log.Print(err)
		return
	}
	charlie := dogs.GetDocs()[0]
	coleman, err := client.CreateDoc(context.Background(), &apipb.DocConstructor{
		Path: &apipb.Path{
			Gtype: "human",
		},
		Attributes: apipb.NewStruct(map[string]interface{}{
			"name": "Coleman",
		}),
	})
	if err != nil {
		log.Print(err)
		return
	}
	ownerConnection, err := client.CreateConnection(context.Background(), &apipb.ConnectionConstructor{
		Path: &apipb.Path{
			Gtype: "owner",
		},
		Attributes: apipb.NewStruct(map[string]interface{}{
			"primary_owner": true,
		}),
		From: charlie.Path,
		To:   coleman.Path,
	})
	if err != nil {
		log.Print(err)
		return
	}
	primary := ownerConnection.Attributes.Fields["primary_owner"].GetBoolValue()
	fmt.Println(primary)
	// Output: true
}

func ExampleClient_SearchConnections() {
	owners, err := client.SearchConnections(context.Background(), &apipb.Filter{
		Gtype:      "owner",
		Expression: `connection.attributes.primary_owner`,
		Limit:      1,
	})
	if err != nil {
		log.Print(err)
		return
	}
	coleman := owners.GetConnections()[0]
	primary := coleman.Attributes.Fields["primary_owner"].GetBoolValue()
	fmt.Println(primary)
	// Output: true
}

func ExampleClient_PatchDoc() {
	dogs, err := client.SearchDocs(context.Background(), &apipb.Filter{
		Gtype:      "dog",
		Expression: `doc.attributes.name.contains("Charl")`,
		Limit:      1,
	})
	if err != nil {
		log.Print(err)
		return
	}
	charlie := dogs.GetDocs()[0]
	charlie, err = client.PatchDoc(context.Background(), &apipb.Patch{
		Path: charlie.Path,
		Attributes: apipb.NewStruct(map[string]interface{}{
			"weight": 25,
		}),
	})
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Println(charlie.GetAttributes().GetFields()["weight"].GetNumberValue())
	// Output: 25
}

func ExampleClient_Publish() {
	res, err := client.Publish(context.Background(), &apipb.OutboundMessage{
		Channel: "testing",
		Data: apipb.NewStruct(map[string]interface{}{
			"text": "hello world",
		}),
	})
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Println(res.String())
	// Output:
}

func ExampleClient_Subscribe() {
	m := machine.New(context.Background())
	m.Go(func(routine machine.Routine) {
		stream, err := client.Subscribe(context.Background(), &apipb.ChannelFilter{
			Channel: "testing",
		})
		if err != nil {
			log.Print("failed to subscribe", err)
			return
		}
		for {
			msg, err := stream.Recv()
			if err != nil {
				log.Print("failed to receive message", err)
				return
			}
			fmt.Println(msg.Data.GetFields()["text"].GetStringValue())
			return
		}
	})
	time.Sleep(1 * time.Second)
	_, err := client.Publish(context.Background(), &apipb.OutboundMessage{
		Channel: "testing",
		Data: apipb.NewStruct(map[string]interface{}{
			"text": "hello world",
		}),
	})
	if err != nil {
		log.Print(err)
		return
	}
	m.Wait()
	// Output: hello world
}

func ExampleClient_GetSchema() {
	schema, err := client.GetSchema(context.Background(), &empty.Empty{})
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Printf("doc types: %s\n", strings.Join(schema.DocTypes, ","))
	fmt.Printf("connection types: %s", strings.Join(schema.ConnectionTypes, ","))
	// Output: doc types: dog,human,identity
	//connection types: owner
}

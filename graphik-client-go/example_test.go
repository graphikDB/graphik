package graphik_test

import (
	"context"
	"fmt"
	"github.com/autom8ter/machine"
	"github.com/golang/protobuf/ptypes/empty"
	apipb2 "github.com/graphikDB/graphik/gen/grpc/go"
	"github.com/graphikDB/graphik/graphik-client-go"
	"github.com/graphikDB/graphik/logger"
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

func ExampleClient_SetAuthorizers() {
	_, err := client.SetAuthorizers(context.Background(), &apipb2.Authorizers{
		Authorizers: []*apipb2.Authorizer{
			{
				Name:       "testing",
				Expression: `this.user.attributes.email.contains("coleman")`,
			},
		},
	})
	if err != nil {
		log.Print(err)
		return
	}
	schema, err := client.GetSchema(context.Background(), &empty.Empty{})
	if err != nil {
		log.Print(err)
		return
	}
	var authorizers []string
	for _, a := range schema.GetAuthorizers().GetAuthorizers() {
		authorizers = append(authorizers, a.Name)
	}
	fmt.Printf("%s", strings.Join(authorizers, ","))
	// Output: testing
}

func ExampleClient_Me() {
	me, err := client.Me(context.Background(), &empty.Empty{})
	if err != nil {
		log.Print(err)
		return
	}
	issuer := me.GetAttributes().GetFields()["sub"].GetStringValue() // token issuer
	fmt.Println(issuer)
	// Output: 107146673535247272789
}

func ExampleClient_HasDoc() {
	ctx := context.Background()
	me, err := client.Me(ctx, &empty.Empty{})
	if err != nil {
		fmt.Print(err)
		return
	}
	has, err := client.HasDoc(ctx, me.GetRef())
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println(has.GetValue())
	// Output: true
}

func ExampleClient_CreateDoc() {
	charlie, err := client.CreateDoc(context.Background(), &apipb2.DocConstructor{
		Ref: &apipb2.RefConstructor{Gtype: "dog"},
		Attributes: apipb2.NewStruct(map[string]interface{}{
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
	dogs, err := client.SearchDocs(context.Background(), &apipb2.Filter{
		Gtype:      "dog",
		Expression: `this.attributes.name.contains("Charl")`,
		Limit:      1,
		Sort:       "ref.gid",
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
	dogs, err := client.SearchDocs(context.Background(), &apipb2.Filter{
		Gtype:      "dog",
		Expression: `this.attributes.name.contains("Charl")`,
		Limit:      1,
	})
	if err != nil {
		log.Print(err)
		return
	}
	charlie := dogs.GetDocs()[0]
	coleman, err := client.CreateDoc(context.Background(), &apipb2.DocConstructor{
		Ref: &apipb2.RefConstructor{Gtype: "human"},
		Attributes: apipb2.NewStruct(map[string]interface{}{
			"name": "Coleman",
		}),
	})
	if err != nil {
		log.Print(err)
		return
	}
	ownerConnection, err := client.CreateConnection(context.Background(), &apipb2.ConnectionConstructor{
		Ref: &apipb2.RefConstructor{Gtype: "owner"},
		Attributes: apipb2.NewStruct(map[string]interface{}{
			"primary_owner": true,
		}),
		From: charlie.Ref,
		To:   coleman.Ref,
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
	owners, err := client.SearchConnections(context.Background(), &apipb2.Filter{
		Gtype:      "owner",
		Expression: `this.attributes.primary_owner`,
		Sort:       "ref.gtype",
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

func ExampleClient_EditDoc() {
	dogs, err := client.SearchDocs(context.Background(), &apipb2.Filter{
		Gtype:      "dog",
		Expression: `this.attributes.name.contains("Charl")`,
		Limit:      1,
	})
	if err != nil {
		log.Print(err)
		return
	}
	charlie := dogs.GetDocs()[0]
	log.Println(charlie.String())
	charlie, err = client.EditDoc(context.Background(), &apipb2.Edit{
		Ref: charlie.Ref,
		Attributes: apipb2.NewStruct(map[string]interface{}{
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
	res, err := client.Publish(context.Background(), &apipb2.OutboundMessage{
		Channel: "testing",
		Data: apipb2.NewStruct(map[string]interface{}{
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
		err := client.Subscribe(context.Background(), &apipb2.ChanFilter{
			Channel:    "testing",
			Expression: `this.data.text.contains("hello")`,
		}, func(msg *apipb2.Message) bool {
			if msg.Data.GetFields()["text"] != nil && msg.Data.GetFields()["text"].GetStringValue() == "hello world" {
				fmt.Println(msg.Data.GetFields()["text"].GetStringValue())
				return false
			}
			return true
		})
		if err != nil {
			log.Print("failed to subscribe", err)
			return
		}
	})
	time.Sleep(1 * time.Second)
	_, err := client.Publish(context.Background(), &apipb2.OutboundMessage{
		Channel: "testing",
		Data: apipb2.NewStruct(map[string]interface{}{
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
	fmt.Printf("connection types: %s\n", strings.Join(schema.ConnectionTypes, ","))
	var authorizers []string
	for _, a := range schema.GetAuthorizers().GetAuthorizers() {
		authorizers = append(authorizers, a.Name)
	}
	fmt.Printf("authorizers: %s", strings.Join(authorizers, ","))
	// Output: doc types: dog,human,user
	//connection types: created,created_by,edited,edited_by,owner
	//authorizers: testing
}

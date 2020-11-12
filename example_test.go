package graphik_test

import (
	"context"
	"fmt"
	"github.com/autom8ter/graphik"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/logger"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"log"
)

func init() {
	ctx := context.Background()

	// ensure graphik server is started with --auth.jwks=https://www.googleapis.com/oauth2/v3/certs
	tokenSource, err := google.DefaultTokenSource(context.Background(), "https://www.googleapis.com/auth/devstorage.full_control")
	if err != nil {
		log.Print(err)
		return
	}

	client, err = graphik.NewClient(ctx, "localhost:7820", tokenSource)
	if err != nil {
		log.Print(err)
		return
	}
}

var client *graphik.Client

func ExampleNewClient() {
	ctx := context.Background()

	// ensure graphik server is started with --auth.jwks=https://www.googleapis.com/oauth2/v3/certs
	j, err := google.DefaultTokenSource(context.Background(), "https://www.googleapis.com/auth/devstorage.full_control")
	if err != nil {
		log.Print(err)
		return
	}

	cli, err := graphik.NewClient(ctx, "localhost:7820", j)
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
	me, err := client.Me(context.Background(), &empty.Empty{})
	if err != nil {
		log.Print(err)
		return
	}
	issuer := me.GetAttributes().GetFields()["iss"].GetStringValue() // token issuer
	fmt.Println(issuer)
	// Output: https://accounts.google.com
}

func ExampleClient_CreateNode() {
	charlie, err := client.CreateNode(context.Background(), &apipb.Node{
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

func ExampleClient_SearchNodes() {
	dogs, err := client.SearchNodes(context.Background(), &apipb.TypeFilter{
		Gtype: "dog",
		Expressions: []string{
			`attributes.name.contains("Charl")`,
		},
		Limit: 1,
	})
	if err != nil {
		log.Print(err)
		return
	}
	dog := dogs.GetNodes()[0]
	name := dog.Attributes.Fields["name"].GetStringValue()
	fmt.Println(name)
	// Output: Charlie
}

func ExampleClient_CreateEdge() {
	dogs, err := client.SearchNodes(context.Background(), &apipb.TypeFilter{
		Gtype: "dog",
		Expressions: []string{
			`attributes.name.contains("Charl")`,
		},
		Limit: 1,
	})
	if err != nil {
		log.Print(err)
		return
	}
	charlie := dogs.GetNodes()[0]
	coleman, err := client.CreateNode(context.Background(), &apipb.Node{
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
	ownerEdge, err := client.CreateEdge(context.Background(), &apipb.Edge{
		Path: &apipb.Path{
			Gtype: "owner",
		},
		Attributes: apipb.NewStruct(map[string]interface{}{
			"primary_owner": true,
		}),
		Cascade: apipb.Cascade_CASCADE_NONE,
		From:    charlie.Path,
		To:      coleman.Path,
	})
	if err != nil {
		log.Print(err)
		return
	}
	primary := ownerEdge.Attributes.Fields["primary_owner"].GetBoolValue()
	fmt.Println(primary)
	// Output: true
}

func ExampleClient_SearchEdges() {
	owners, err := client.SearchEdges(context.Background(), &apipb.TypeFilter{
		Gtype: "owner",
		Expressions: []string{
			`attributes.primary_owner`,
		},
		Limit: 1,
	})
	if err != nil {
		log.Print(err)
		return
	}
	coleman := owners.GetEdges()[0]
	primary := coleman.Attributes.Fields["primary_owner"].GetBoolValue()
	fmt.Println(primary)
	// Output: true
}

func ExampleClient_PatchNode() {
	dogs, err := client.SearchNodes(context.Background(), &apipb.TypeFilter{
		Gtype: "dog",
		Expressions: []string{
			`attributes.name.contains("Charl")`,
		},
		Limit: 1,
	})
	if err != nil {
		log.Print(err)
		return
	}
	charlie := dogs.GetNodes()[0]
	charlie, err = client.PatchNode(context.Background(), &apipb.Patch{
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

func ExampleClient_SetAuth() {
	auth, err := client.SetAuth(context.Background(), &apipb.AuthConfig{
		JwksSources: []string{"https://www.googleapis.com/oauth2/v3/certs"},
		//AuthExpressions:      []string{`user.attributes.email.contains("cole")`},
	})
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Println(auth.GetJwksSources()[0])
	// Output: https://www.googleapis.com/oauth2/v3/certs
}

func ExampleClient_Publish() {

}

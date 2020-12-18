package graphik_test

import (
	"context"
	"fmt"
	"github.com/autom8ter/machine"
	"github.com/golang/protobuf/ptypes/empty"
	apipb2 "github.com/graphikDB/graphik/gen/grpc/go"
	"github.com/graphikDB/graphik/graphik-client-go"
	"golang.org/x/oauth2/google"
	"strings"
	"time"
)

func init() {
	ctx := context.Background()

	// ensure graphik server is started with --auth.jwks=https://www.googleapis.com/oauth2/v3/certs
	tokenSource, err := google.DefaultTokenSource(context.Background(), "https://www.googleapis.com/auth/devstorage.full_control")
	if err != nil {
		fmt.Print(err)
		return
	}

	client, err = graphik.NewClient(ctx, "localhost:7820", graphik.WithTokenSource(tokenSource))
	if err != nil {
		fmt.Print(err)
		return
	}
}

var client *graphik.Client

func ExampleNewClient() {
	ctx := context.Background()

	// ensure graphik server is started with --auth.jwks=https://www.googleapis.com/oauth2/v3/certs
	source, err := google.DefaultTokenSource(context.Background(), "https://www.googleapis.com/auth/devstorage.full_control")
	if err != nil {
		fmt.Print(err)
		return
	}
	cli, err := graphik.NewClient(ctx, "localhost:7820", graphik.WithTokenSource(source))
	if err != nil {
		fmt.Print(err)
		return
	}
	pong, err := cli.Ping(ctx, &empty.Empty{})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(pong.Message)
	// Output: PONG
}

func ExampleClient_Ping() {
	ctx := context.Background()
	msg, err := client.Ping(ctx, &empty.Empty{})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(msg.Message)
	// Output: PONG
}

func ExampleClient_SetAuthorizers() {
	err := client.SetAuthorizers(context.Background(), &apipb2.Authorizers{
		Authorizers: []*apipb2.Authorizer{
			{
				Name:            "testing-request",
				Method:          "/api.DatabaseService/GetSchema",
				Expression:      `this.user.attributes.email.contains("coleman")`,
				TargetRequests:  true,
				TargetResponses: false,
			},
		},
	})
	if err != nil {
		fmt.Print(err)
		return
	}
	schema, err := client.GetSchema(context.Background(), &empty.Empty{})
	if err != nil {
		fmt.Print(err)
		return
	}
	var authorizers []string
	for _, a := range schema.GetAuthorizers().GetAuthorizers() {
		authorizers = append(authorizers, a.Name)
	}
	fmt.Printf("%s", strings.Join(authorizers, ","))
	// Output: testing-request
}

func ExampleClient_SetTypeValidators() {
	err := client.SetTypeValidators(context.Background(), &apipb2.TypeValidators{
		Validators: []*apipb2.TypeValidator{
			{
				Name:              "testing",
				Gtype:             "dog",
				Expression:        `int(this.attributes.weight) > 0`,
				TargetDocs:        true,
				TargetConnections: false,
			},
		},
	})
	if err != nil {
		fmt.Print(err)
		return
	}
	schema, err := client.GetSchema(context.Background(), &empty.Empty{})
	if err != nil {
		fmt.Print(err)
		return
	}
	var validators []string
	for _, a := range schema.GetValidators().GetValidators() {
		validators = append(validators, a.Name)
	}
	fmt.Printf("%s", strings.Join(validators, ","))
	// Output: testing
}

func ExampleClient_SetIndexes() {
	err := client.SetIndexes(context.Background(), &apipb2.Indexes{
		Indexes: []*apipb2.Index{
			{
				Name:        "testing",
				Gtype:       "owner",
				Expression:  `this.attributes.primary_owner`,
				Docs:        false,
				Connections: true,
			},
		},
	})
	if err != nil {
		fmt.Print(err)
		return
	}
	schema, err := client.GetSchema(context.Background(), &empty.Empty{})
	if err != nil {
		fmt.Print(err)
		return
	}
	var indexes []string
	for _, a := range schema.GetIndexes().GetIndexes() {
		indexes = append(indexes, a.Name)
	}
	fmt.Printf("%s", strings.Join(indexes, ","))
	// Output: testing
}

func ExampleClient_Me() {
	me, err := client.Me(context.Background(), &empty.Empty{})
	if err != nil {
		fmt.Print(err)
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
	ctx := context.Background()
	charlie, err := client.CreateDoc(ctx, &apipb2.DocConstructor{
		Ref: &apipb2.RefConstructor{Gtype: "dog"},
		Attributes: apipb2.NewStruct(map[string]interface{}{
			"name":   "Charlie",
			"weight": 18,
		}),
	})
	if err != nil {
		fmt.Print("createDoc: ", err)
		return
	}
	has, err := client.HasDoc(ctx, charlie.GetRef())
	if err != nil {
		fmt.Print("hasDoc: ", err)
		return
	}
	if !has.GetValue() {
		fmt.Print("failed to find charlie")
		return
	}
	exists, err := client.ExistsDoc(ctx, &apipb2.ExistsFilter{
		Gtype:      "dog",
		Expression: "this.attributes.name.contains('Charlie')",
	})
	if err != nil {
		fmt.Print("existsDoc: ", err)
		return
	}
	if !exists.GetValue() {
		fmt.Print("failed to find charlie")
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
		fmt.Print(err)
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
		fmt.Print(err)
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
		fmt.Print(err)
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
		fmt.Print(err)
		return
	}
	has, err := client.HasConnection(context.Background(), ownerConnection.Ref)
	if err != nil {
		fmt.Print(err)
		return
	}
	if !has.GetValue() {
		fmt.Print("failed to find owner connection")
		return
	}
	exists, err := client.ExistsConnection(context.Background(), &apipb2.ExistsFilter{
		Gtype:      "owner",
		Expression: "this.attributes.primary_owner",
	})
	if err != nil {
		fmt.Print(err)
		return
	}
	if !exists.GetValue() {
		fmt.Print("failed to find owner connection")
		return
	}
	primary := ownerConnection.Attributes.Fields["primary_owner"].GetBoolValue()
	fmt.Println(primary)
	// Output: true
}

func ExampleClient_SearchConnections() {
	owners, err := client.SearchConnections(context.Background(), &apipb2.Filter{
		Gtype:      "owner",
		Expression: `this.attributes.primary_owner == true`,
		Sort:       "ref.gtype",
		Limit:      1,
		Index:      "testing",
	})
	if err != nil {
		fmt.Print(err)
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
		fmt.Print(err)
		return
	}
	charlie := dogs.GetDocs()[0]
	charlie, err = client.EditDoc(context.Background(), &apipb2.Edit{
		Ref: charlie.Ref,
		Attributes: apipb2.NewStruct(map[string]interface{}{
			"weight": 25,
		}),
	})
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println(charlie.GetAttributes().GetFields()["weight"].GetNumberValue())
	// Output: 25
}

func ExampleClient_Broadcast() {
	err := client.Broadcast(context.Background(), &apipb2.OutboundMessage{
		Channel: "testing",
		Data: apipb2.NewStruct(map[string]interface{}{
			"text": "hello world",
		}),
	})
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("success!")
	// Output: success!
}

func ExampleClient_Stream() {
	m := machine.New(context.Background())
	m.Go(func(routine machine.Routine) {
		err := client.Stream(context.Background(), &apipb2.StreamFilter{
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
			fmt.Print("failed to subscribe", err)
			return
		}
	})
	time.Sleep(1 * time.Second)
	err := client.Broadcast(context.Background(), &apipb2.OutboundMessage{
		Channel: "testing",
		Data: apipb2.NewStruct(map[string]interface{}{
			"text": "hello world",
		}),
	})
	if err != nil {
		fmt.Print(err)
		return
	}
	m.Wait()
	// Output: hello world
}

func ExampleClient_GetSchema() {
	schema, err := client.GetSchema(context.Background(), &empty.Empty{})
	if err != nil {
		fmt.Print(err)
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
	//authorizers: testing-request
}

func ExampleClient_Traverse() {
	me, err := client.Me(context.Background(), &empty.Empty{})
	if err != nil {
		fmt.Print(err)
		return
	}
	ctx := context.Background()
	bfsTraversals, err := client.Traverse(ctx, &apipb2.TraverseFilter{
		Root:                 me.GetRef(),
		DocExpression:        "this.attributes.name == 'Charlie'",
		ConnectionExpression: "",
		Limit:                1,
		Sort:                 "",
		Reverse:              false,
		Algorithm:            apipb2.Algorithm_BFS,
		MaxDepth:             1,
		MaxHops:              50,
	})
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println(len(bfsTraversals.GetTraversals()))
	dfsTraversals, err := client.Traverse(ctx, &apipb2.TraverseFilter{
		Root:                 me.GetRef(),
		DocExpression:        "",
		ConnectionExpression: "",
		Limit:                3,
		Sort:                 "",
		Reverse:              false,
		Algorithm:            apipb2.Algorithm_DFS,
		MaxDepth:             1,
		MaxHops:              50,
	})
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println(len(dfsTraversals.GetTraversals()))
	//Output: 1
	//3
}

func ExampleClient_DelConnections() {
	ctx := context.Background()
	err := client.DelConnections(ctx, &apipb2.Filter{
		Gtype:      "owner",
		Expression: "this.attributes.primary_owner",
		Limit:      10,
	})
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("Success!")
	//Output: Success!
}

func ExampleClient_DelDocs() {
	ctx := context.Background()
	err := client.DelDocs(ctx, &apipb2.Filter{
		Gtype:      "dog",
		Expression: "this.attributes.name == 'Charlie'",
		Limit:      10,
	})
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("Success!")
	//Output: Success!
}

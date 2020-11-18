package plugins_test

import (
	"context"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/flags"
	"github.com/autom8ter/graphik/plugins"
	"google.golang.org/protobuf/types/known/structpb"
	"time"
)

func ExampleAuthorizerFunc_Serve() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	authorizer := plugins.NewAuthorizer(func(ctx context.Context, r *apipb.RequestIntercept) (*apipb.Decision, error) {
		switch val := r.Request.(type) {
		case *apipb.RequestIntercept_Filter:
			if val.Filter.GetLimit() > 50 {
				return &apipb.Decision{
					Value: false,
				}, nil
			}
		}
		return &apipb.Decision{
			Value: true,
		}, nil
	})
	authorizer.Serve(ctx, &flags.PluginFlags{
		BindGrpc: ":8080",
		BindHTTP: ":8081",
		Metrics:  true,
	})
	fmt.Println("Done")
	// Output: Done
}

func ExampleTriggerFunc_Serve() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	trigger := plugins.NewTrigger(func(ctx context.Context, trigger *apipb.Trigger) (*apipb.StateChange, error) {
		state := trigger.State
		switch r := state.GetMutation().GetObject().(type) {
		case *apipb.Mutation_NodeConstructor:
			r.NodeConstructor.Attributes.Fields["testing"] = structpb.NewBoolValue(true)
		}
		return state, nil
	})
	trigger.Serve(ctx, &flags.PluginFlags{
		BindGrpc: ":8080",
		BindHTTP: ":8081",
		Metrics:  true,
	})
	fmt.Println("Done")
	// Output: Done
}

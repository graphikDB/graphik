package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	generated1 "github.com/autom8ter/graphik/api/admin/generated"
	"github.com/autom8ter/graphik/api/model"
	"github.com/autom8ter/graphik/generic"
	"github.com/autom8ter/machine"
)

func (r *mutationResolver) CreateNode(ctx context.Context, input model.NodeConstructor) (*model.Node, error) {
	if input.Path.ID == "" {
		random := generic.UUID()
		input.Path.ID = random
	}
	if input.Path.Type == "" {
		input.Path.Type = generic.Default
	}
	res, err := r.runtime.Execute(&model.Command{
		Op:    model.OpCreateNode,
		Value: input,
	})
	if err != nil {
		return nil, err
	}
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res.(*model.Node), nil
}

func (r *mutationResolver) PatchNode(ctx context.Context, input *model.Patch) (*model.Node, error) {
	res, err := r.runtime.Execute(&model.Command{
		Op:    model.OpPatchNode,
		Value: input,
	})
	if err != nil {
		return nil, err
	}
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res.(*model.Node), nil
}

func (r *mutationResolver) DelNode(ctx context.Context, input model.Path) (*model.Counter, error) {
	res, err := r.runtime.Execute(&model.Command{
		Op:    model.OpDeleteNode,
		Value: input,
	})
	if err != nil {
		return nil, err
	}
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res.(*model.Counter), nil
}

func (r *mutationResolver) CreateEdge(ctx context.Context, input model.EdgeConstructor) (*model.Edge, error) {
	if input.Path.ID == "" {
		random := generic.UUID()
		input.Path.ID = random
	}
	if input.Path.Type == "" {
		input.Path.Type = generic.Default
	}
	res, err := r.runtime.Execute(&model.Command{
		Op:    model.OpCreateEdge,
		Value: input,
	})
	if err != nil {
		return nil, err
	}
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res.(*model.Edge), nil
}

func (r *mutationResolver) PatchEdge(ctx context.Context, input model.Patch) (*model.Edge, error) {
	res, err := r.runtime.Execute(&model.Command{
		Op:    model.OpPatchEdge,
		Value: input,
	})
	if err != nil {
		return nil, err
	}
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res.(*model.Edge), nil
}

func (r *mutationResolver) DelEdge(ctx context.Context, input model.Path) (*model.Counter, error) {
	res, err := r.runtime.Execute(&model.Command{
		Op:    model.OpDeleteEdge,
		Value: input,
	})
	if err != nil {
		return nil, err
	}
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res.(*model.Counter), nil
}

func (r *mutationResolver) Publish(ctx context.Context, input model.Message) (*model.Counter, error) {
	r.runtime.Machine().Go(func(routine machine.Routine) {
		routine.Publish(input.Channel, input)
	})
	return &model.Counter{Count: 1}, nil
}

func (r *queryResolver) GetNode(ctx context.Context, input model.Path) (*model.Node, error) {
	return r.runtime.Node(ctx, input)
}

func (r *queryResolver) GetNodes(ctx context.Context, input model.Filter) ([]*model.Node, error) {
	return r.runtime.Nodes(ctx, input)
}

func (r *queryResolver) DepthSearch(ctx context.Context, input model.DepthFilter) ([]*model.Node, error) {
	if input.Reverse != nil && *input.Reverse {
		return r.runtime.DepthTo(ctx, input)
	}
	return r.runtime.DepthFrom(ctx, input)
}

func (r *queryResolver) GetEdge(ctx context.Context, input model.Path) (*model.Edge, error) {
	return r.runtime.Edge(ctx, input)
}

func (r *queryResolver) GetEdges(ctx context.Context, input model.Filter) ([]*model.Edge, error) {
	return r.runtime.Edges(ctx, input)
}

func (r *subscriptionResolver) Subscribe(ctx context.Context, channel string) (<-chan *model.Message, error) {
	ch := make(chan *model.Message)
	r.runtime.Machine().Go(func(routine machine.Routine) {
		routine.Subscribe(channel, func(obj interface{}) {
			for {
				select {
				case <-routine.Context().Done():
					return
				case <-ctx.Done():
					return
				default:
					if msg, ok := obj.(*model.Message); ok {
						ch <- msg
					}
				}
			}
		})
	})
	return ch, nil
}

func (r *subscriptionResolver) NodeChange(ctx context.Context, typeArg model.ChangeFilter) (<-chan *model.Node, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	filter := func(obj interface{}) bool {
		node, ok := obj.(*model.Node)
		if !ok {
			return false
		}
		pass, _ := model.Evaluate(typeArg.Expressions, node)
		return pass
	}
	ch := make(chan *model.Node)
	channel := fmt.Sprintf("nodes.%s", typeArg.Type)
	r.runtime.Machine().Go(func(routine machine.Routine) {
		routine.SubscribeFilter(channel, filter, func(msg interface{}) {
			for {
				select {
				case <-routine.Context().Done():
					return
				case <-ctx.Done():
					return
				default:
					ch <- msg.(*model.Node)
				}
			}
		})
	})
	return ch, nil
}

func (r *subscriptionResolver) EdgeChange(ctx context.Context, typeArg model.ChangeFilter) (<-chan *model.Edge, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	filter := func(obj interface{}) bool {
		edge, ok := obj.(*model.Edge)
		if !ok {
			return false
		}
		pass, _ := model.Evaluate(typeArg.Expressions, edge)
		return pass
	}
	ch := make(chan *model.Edge)
	channel := fmt.Sprintf("edges.%s", typeArg.Type)
	r.runtime.Machine().Go(func(routine machine.Routine) {
		routine.SubscribeFilter(channel, filter, func(msg interface{}) {
			for {
				select {
				case <-routine.Context().Done():
					return
				case <-ctx.Done():
					return
				default:
					ch <- msg.(*model.Edge)
				}
			}
		})
	})
	return ch, nil
}

// Mutation returns generated1.MutationResolver implementation.
func (r *Resolver) Mutation() generated1.MutationResolver { return &mutationResolver{r} }

// Query returns generated1.QueryResolver implementation.
func (r *Resolver) Query() generated1.QueryResolver { return &queryResolver{r} }

// Subscription returns generated1.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated1.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *mutationResolver) CreateNodes(ctx context.Context, input model.NodeConstructor) (*model.Node, error) {
	panic(fmt.Errorf("not implemented"))
}

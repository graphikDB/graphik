package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/autom8ter/graphik/api/model"
	"github.com/autom8ter/graphik/api/user/generated"
)

func (r *queryResolver) Me(ctx context.Context, input *model.Empty) (*model.Node, error) {
	panic(fmt.Errorf("not implemented"))
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }

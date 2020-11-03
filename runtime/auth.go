package runtime

import (
	"context"
	"github.com/autom8ter/graphik/model"
	"net/http"
)

const (
	authCtxKey = "x-graphik-auth-ctx"
)


func (a *Store) toContext(r *http.Request, payload map[string]interface{}) *http.Request {
	path := model.Path{
		Type: "subject",
		ID: payload["sub"].(string),
	}
	n, ok := a.nodes.Get(path)
	if !ok {
		n = a.nodes.Set(&model.Node{
			Path:       path,
			Attributes: n.Attributes,
		})
	}
	return r.WithContext(context.WithValue(r.Context(), authCtxKey, n))
}

func (a *Store) getNode(r *http.Request) *model.Node {
	val, ok := r.Context().Value(authCtxKey).(*model.Node)
	if !ok {
		return nil
	}
	return val
}

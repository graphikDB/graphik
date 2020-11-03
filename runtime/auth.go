package runtime

import (
	"context"
	"encoding/json"
	"github.com/autom8ter/graphik/graph/model"
	"net/http"
	"strings"
	"time"
)

const (
	authCtxKey = "x-graphik-auth-ctx"
)

func (a *Store) Middleware() func(handler http.Handler) http.HandlerFunc {
	return func(handler http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			tokenString := strings.Replace(authHeader, "Bearer ", "", -1)
			if tokenString == "" {
				http.Error(w, "empty authorization header", http.StatusUnauthorized)
				return
			}
			payload, err := a.jwks.VerifyJWT(tokenString)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			if exp, ok := payload["exp"].(int64); ok {
				if exp < time.Now().Unix() {
					http.Error(w, "token expired", http.StatusUnauthorized)
					return
				}
			}
			if exp, ok := payload["exp"].(int); ok {
				if int64(exp) < time.Now().Unix() {
					http.Error(w, "token expired", http.StatusUnauthorized)
					return
				}
			}
			handler.ServeHTTP(w, a.toContext(r, payload))
		}
	}
}

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

func (a *Store) PutJWKS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var uris []string
		json.NewDecoder(r.Body).Decode(&uris)
		if err := a.jwks.Override(uris); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func (a *Store) GetJWKS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwks := a.jwks.List()
		json.NewEncoder(w).Encode(&jwks)
	}
}


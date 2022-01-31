package gql

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/apollotracing"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
	"github.com/graphikDB/graphik/gen/gql/go/generated"
	"github.com/graphikDB/graphik/gen/grpc/go"
	"github.com/graphikDB/graphik/gql/session"
	"github.com/graphikDB/graphik/logger"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/metadata"
	"net/http"
	"time"
)

func init() {
	gob.Register(&oauth2.Token{})
}

// This file will not be regenerated automatically.
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	client  apipb.DatabaseServiceClient
	clusterClient  apipb.RaftServiceClient
	logger  *logger.Logger
	session session.SessionManager
	uiPath  string
}

func NewResolver(client apipb.DatabaseServiceClient, clusterClient apipb.RaftServiceClient, logger *logger.Logger, session session.SessionManager) *Resolver {
	return &Resolver{
		client:  client,
		clusterClient: clusterClient,
		logger:  logger,
		session: session,
		uiPath:  "/ui",
	}
}

func (r *Resolver) QueryHandler() http.Handler {
	srv := handler.New(generated.NewExecutableSchema(generated.Config{
		Resolvers:  r,
		Directives: generated.DirectiveRoot{},
		Complexity: generated.ComplexityRoot{},
	}))
	srv.AddTransport(transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, error) {
			auth := initPayload.Authorization()
			ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", auth)
			return ctx, nil
		},
		KeepAlivePingInterval: 10 * time.Second,
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})
	srv.SetQueryCache(lru.New(1000))
	srv.Use(extension.Introspection{})
	srv.Use(&apollotracing.Tracer{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})
	return r.authMiddleware(srv)
}

func (r *Resolver) authMiddleware(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if r.session != nil && req.Header.Get("Authorization") == "" {
			authToken, err := r.session.GetToken(req)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if authToken == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if !authToken.Token.Valid() {
				http.Error(w, "unauthorized - token expired", http.StatusUnauthorized)
				return
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken.Token.AccessToken))
		}

		ctx := req.Context()
		for k, arr := range req.Header {
			if len(arr) > 0 {
				ctx = metadata.AppendToOutgoingContext(ctx, k, arr[0])
			}
		}
		handler.ServeHTTP(w, req.WithContext(ctx))
	}
}

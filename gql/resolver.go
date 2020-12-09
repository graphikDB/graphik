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
	"github.com/graphikDB/graphik/gen/gql/go/generated"
	"github.com/graphikDB/graphik/gen/grpc/go"
	"github.com/graphikDB/graphik/helpers"
	"github.com/graphikDB/graphik/logger"
	"github.com/autom8ter/machine"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/metadata"
	"html/template"
	"net/http"
	"time"
)

func init() {
	gob.Register(&oauth2.Token{})
}

// This file will not be regenerated automatically.
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	client     apipb.DatabaseServiceClient
	cors       *cors.Cors
	machine    *machine.Machine
	store      *sessions.CookieStore
	config     *oauth2.Config
	cookieName string
}

func NewResolver(ctx context.Context, client apipb.DatabaseServiceClient, cors *cors.Cors, config *oauth2.Config) *Resolver {
	r := &Resolver{
		client:     client,
		cors:       cors,
		machine:    machine.New(ctx),
		config:     config,
		cookieName: "graphik",
	}
	if config != nil {
		r.store = sessions.NewCookieStore([]byte(config.ClientSecret))
	}
	return r
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
	return r.cors.Handler(r.authMiddleware(srv))
}

func (r *Resolver) authMiddleware(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		if r.store != nil {
			sess, _ := r.store.Get(req, r.cookieName)
			if sess != nil && req.Header.Get("Authorization") == "" && sess.Values["token"] != nil {
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sess.Values["token"].(*oauth2.Token).AccessToken))
			}
		}
		for k, arr := range req.Header {
			if len(arr) > 0 {
				ctx = metadata.AppendToOutgoingContext(ctx, k, arr[0])
			}
		}
		handler.ServeHTTP(w, req.WithContext(ctx))
	}
}

func (r *Resolver) Playground() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if r.config == nil || r.config.ClientID == "" {
			http.Error(w, "playground disabled", http.StatusNotFound)
			return
		}
		sess, err := r.store.Get(req, r.cookieName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if sess.Values["token"] == nil {
			r.redirectLogin(sess, w, req)
			return
		}
		val, ok := sess.Values["token"]
		if !ok {
			r.redirectLogin(sess, w, req)
			return
		}
		authToken, ok := val.(*oauth2.Token)
		if !ok {
			r.redirectLogin(sess, w, req)
			return
		}
		if authToken == nil {
			r.redirectLogin(sess, w, req)
			return
		}
		rtoken, err := r.refreshToken(authToken)
		if err == nil {
			authToken = rtoken
		} else {
			if time.Unix(sess.Values["exp"].(int64), 0).Before(time.Now()) {
				r.redirectLogin(sess, w, req)
				return
			}
		}
		if !authToken.Valid() {
			r.redirectLogin(sess, w, req)
			return
		}
		w.Header().Add("Content-Type", "text/html")
		var playground = template.Must(template.New("playground").Parse(`<!DOCTYPE html>
<html>

<head>
  <meta charset=utf-8/>
  <meta name="viewport" content="user-scalable=no, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, minimal-ui">
  <title>Graphik Playground</title>
  <link rel="stylesheet" href="//cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
  <link rel="shortcut icon" href="//cdn.jsdelivr.net/npm/graphql-playground-react/build/favicon.png" />
  <script src="//cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
</head>

<body>
  <div id="root">
    <style>
      body {
        background-color: rgb(23, 42, 58);
        font-family: Open Sans, sans-serif;
        height: 90vh;
      }

      #root {
        height: 100%;
        width: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
      }

      .loading {
        font-size: 32px;
        font-weight: 200;
        color: rgba(255, 255, 255, .6);
        margin-left: 20px;
      }

      img {
        width: 78px;
        height: 78px;
      }

      .title {
        font-weight: 400;
      }
    </style>
    <img src='//cdn.jsdelivr.net/npm/graphql-playground-react/build/logo.png' alt=''>
    <div class="loading"> Loading
      <span class="title">Graphik Playground</span>
    </div>
  </div>
  <script>window.addEventListener('load', function (event) {
 		const wsProto = location.protocol == 'https:' ? 'wss:' : 'ws:'
      GraphQLPlayground.init(document.getElementById('root'), {
		endpoint: location.protocol + '//' + location.host,
		subscriptionsEndpoint: wsProto + '//' + location.host,
		shareEnabled: true,
		settings: {
			'request.credentials': 'same-origin',
			'prettier.useTabs': true
		}
      })
    })</script>
</body>

</html>
`))

		playground.Execute(w, map[string]string{})
	}
}

func (r *Resolver) redirectLogin(sess *sessions.Session, w http.ResponseWriter, req *http.Request) {
	state := helpers.Hash([]byte(ksuid.New().String()))
	sess.Values["state"] = state
	redirect := r.config.AuthCodeURL(state)
	if err := sess.Save(req, w); err != nil {
		http.Error(w, "failed to save session", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, req, redirect, http.StatusTemporaryRedirect)
}

func (r *Resolver) PlaygroundCallback(playgroundRedirect string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if r.config == nil || r.config.ClientID == "" {
			http.Error(w, "playground disabled", http.StatusNotFound)
			return
		}
		code := req.URL.Query().Get("code")
		state := req.URL.Query().Get("state")
		if code == "" {
			http.Error(w, "empty authorization code", http.StatusBadRequest)
			return
		}
		if state == "" {
			http.Error(w, "empty authorization state", http.StatusBadRequest)
			return
		}
		sess, err := r.store.Get(req, r.cookieName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		stateVal := sess.Values["state"]
		if stateVal == nil {
			http.Error(w, "failed to get session state", http.StatusForbidden)
			return
		}
		if stateVal.(string) != state {
			http.Error(w, fmt.Sprintf("session state mismatch: %s", stateVal.(string)), http.StatusForbidden)
			return
		}
		token, err := r.config.Exchange(req.Context(), code)
		if err != nil {
			logger.Error("failed to exchange authorization code", zap.Error(err))
			http.Error(w, "failed to exchange authorization code", http.StatusInternalServerError)
			return
		}
		sess.Values["token"] = token
		sess.Values["exp"] = time.Now().Add(1 * time.Hour).Unix()
		if err := sess.Save(req, w); err != nil {
			logger.Error("failed to save session", zap.Error(err))
			http.Error(w, "failed to save session", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, req, playgroundRedirect, http.StatusTemporaryRedirect)
	}
}

func (r *Resolver) refreshToken(token *oauth2.Token) (*oauth2.Token, error) {
	return r.config.TokenSource(oauth2.NoContext, token).Token()
}

package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	authCtxKey = "x-graphik-auth-ctx"
)

func New(jwks string) (*Auth, error) {
	set, err := jwk.Fetch(jwks)
	if err != nil {
		return nil, err
	}
	return &Auth{
		set:     set,
		mu:      sync.RWMutex{},
		jwksUri: jwks,
	}, nil
}

type Auth struct {
	mu      sync.RWMutex
	set     *jwk.Set
	jwksUri string
}

func (a *Auth) VerifyJWT(token string) (map[string]interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	message, err := jws.ParseString(token)
	if err != nil {
		return nil, err
	}
	kid, ok := message.Signatures()[0].ProtectedHeaders().Get("kid")
	if !ok {
		return nil, fmt.Errorf("kid not found")
	}
	algI, ok := message.Signatures()[0].ProtectedHeaders().Get("alg")
	if !ok {
		return nil, fmt.Errorf("alg not found")
	}
	alg, ok := algI.(jwa.SignatureAlgorithm)
	if !ok {
		return nil, fmt.Errorf("alg type cast error")
	}
	keys := a.set.LookupKeyID(kid.(string))
	if len(keys) == 0 {
		return nil, fmt.Errorf("kid %v has zero keys", kid)
	}
	var key interface{}
	if err := keys[0].Raw(&key); err != nil {
		return nil, err
	}

	payload, err := jws.Verify([]byte(token), alg, key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to verify message")
	}
	data := map[string]interface{}{}
	return data, json.Unmarshal(payload, &data)
}

func (a *Auth) RefreshKeys() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	set, err := jwk.Fetch(a.jwksUri)
	if err != nil {
		return err
	}
	a.set = set
	return nil
}

func (a *Auth) Middleware() func(handler http.Handler) http.HandlerFunc {
	return func(handler http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			tokenString := strings.Replace(authHeader, "Bearer ", "", -1)
			if tokenString == "" {
				http.Error(w, "empty authorization header", http.StatusUnauthorized)
				return
			}
			payload, err := a.VerifyJWT(tokenString)
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

func (a *Auth) toContext(r *http.Request, payload map[string]interface{}) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), authCtxKey, payload))
}

func (a *Auth) Claims(r *http.Request) map[string]interface{} {
	val, ok := r.Context().Value(authCtxKey).(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}
	return val
}

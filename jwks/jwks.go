package jwks

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/autom8ter/graphik/logger"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	authCtxKey = "x-graphik-auth-ctx"
)

func New(jwks []string) (*Auth, error) {
	sets := map[string]*jwk.Set{}
	for _, uri := range jwks {
		set, err := jwk.Fetch(uri)
		if err != nil {
			return nil, err
		}
		sets[uri] = set
	}

	return &Auth{
		set: sets,
		mu:  sync.RWMutex{},
	}, nil
}

type Auth struct {
	mu  sync.RWMutex
	set map[string]*jwk.Set
}

func (a *Auth) VerifyJWT(token string) (map[string]interface{}, error) {
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
	a.mu.RLock()
	defer a.mu.RUnlock()
	for uri, set := range a.set {
		keys := set.LookupKeyID(kid.(string))
		if len(keys) == 0 {
			continue
		}
		var key interface{}
		if err := keys[0].Raw(&key); err != nil {
			logger.Error("jwks validation failure", zap.String("uri", uri), zap.Error(err))
			continue
		}
		payload, err := jws.Verify([]byte(token), alg, key)
		if err != nil {
			logger.Error("jwks validation failure", zap.String("uri", uri), zap.Error(err))
			continue
		}
		data := map[string]interface{}{}
		return data, json.Unmarshal(payload, &data)
	}
	return nil, errors.New("zero jwks matches")
}

func (a *Auth) RefreshKeys() error {
	sets := map[string]*jwk.Set{}
	for uri, _ := range a.set {
		set, err := jwk.Fetch(uri)
		if err != nil {
			return err
		}
		a.mu.Lock()
		sets[uri] = set
		a.mu.Unlock()
	}
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

func (a *Auth) PutJWKS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var jwks []string
		json.NewDecoder(r.Body).Decode(&jwks)
		for _, uri := range jwks {
			set, err := jwk.Fetch(uri)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			a.mu.Lock()
			a.set[uri] = set
			a.mu.Unlock()
		}
	}
}

func (a *Auth) GetJWKS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var jwks []string
		for uri, _ := range a.set {
			jwks = append(jwks, uri)
		}
		json.NewEncoder(w).Encode(&jwks)
	}
}


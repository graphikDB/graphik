package auth

import (
	"encoding/json"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/logger"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sync"
)

func New() *Auth {
	return &Auth{
		set: map[string]*Set{},
		mu:  sync.RWMutex{},
	}
}

type Set struct {
	URI    string
	Issuer string
	Set    *jwk.Set
}

type Auth struct {
	mu          sync.RWMutex
	set         map[string]*Set
	expressions []string
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
	for _, set := range a.set {
		keys := set.Set.LookupKeyID(kid.(string))
		if len(keys) == 0 {
			continue
		}
		var key interface{}
		if err := keys[0].Raw(&key); err != nil {
			logger.Error("jwks validation failure", zap.String("uri", set.URI), zap.Error(errors.WithStack(err)))
			continue
		}
		payload, err := jws.Verify([]byte(token), alg, key)
		if err != nil {
			logger.Error("jwks validation failure", zap.String("uri", set.URI), zap.Error(errors.WithStack(err)))
			continue
		}
		data := map[string]interface{}{}
		if err := json.Unmarshal(payload, &data); err != nil {
			return nil, err
		}
		if issuer := data["iss"].(string); issuer != set.Issuer {
			continue
		}
		return data, nil
	}
	return nil, errors.New("zero jwks matches")
}

func (a *Auth) RefreshKeys() error {
	for uri, s := range a.set {
		set, err := jwk.Fetch(uri)
		if err != nil {
			return err
		}
		a.mu.Lock()
		a.set[uri] = &Set{
			URI:    uri,
			Issuer: s.Issuer,
			Set:    set,
		}
		a.mu.Unlock()
	}
	return nil
}

func (a *Auth) Override(auth *apipb.Auth) error {
	for _, source := range auth.JwksSources {
		set, err := jwk.Fetch(source.Uri)
		if err != nil {
			return err
		}
		a.mu.Lock()
		a.set[source.Uri] = &Set{
			URI:    source.Uri,
			Issuer: source.Issuer,
			Set:    set,
		}
		a.mu.Unlock()
	}
	a.mu.Lock()
	a.expressions = auth.AuthExpressions
	a.mu.Unlock()
	return nil
}

func (a *Auth) Raw() *apipb.Auth {
	var jwksSources []*apipb.JWKSSource
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, s := range a.set {
		jwksSources = append(jwksSources, &apipb.JWKSSource{
			Uri:    s.URI,
			Issuer: s.Issuer,
		})
	}
	return &apipb.Auth{
		JwksSources:     jwksSources,
		AuthExpressions: a.expressions,
	}
}

func (a *Auth) Authorize(intercept *apipb.RequestIntercept) (bool, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.expressions) == 0 {
		return true, nil
	}
	return apipb.EvaluateExpressions(a.expressions, intercept)
}

package auth

import (
	"encoding/json"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/express"
	"github.com/autom8ter/graphik/logger"
	"github.com/google/cel-go/cel"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sync"
)

func New(cfg *apipb.AuthConfig) (*Auth, error) {
	setMap := map[string]*jwk.Set{}
	for _, source := range cfg.JwksSources {
		set, err := jwk.Fetch(source)
		if err != nil {
			return nil, err
		}
		setMap[source] = set
	}
	a := &Auth{
		set:         setMap,
		mu:          sync.RWMutex{},
		expressions: cfg.AuthExpressions,
	}
	if len(a.expressions) > 0 && a.expressions[0] != "" {
		programs, err := express.Programs(cfg.AuthExpressions)
		if err != nil {
			return nil, err
		}
		a.programs = programs
	}
	return a, nil
}

type Auth struct {
	mu          sync.RWMutex
	set         map[string]*jwk.Set
	expressions []string
	programs    []cel.Program
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
			logger.Error("jwks validation failure", zap.String("uri", uri), zap.Error(errors.WithStack(err)))
			continue
		}
		payload, err := jws.Verify([]byte(token), alg, key)
		if err != nil {
			logger.Error("jwks validation failure", zap.String("uri", uri), zap.Error(errors.WithStack(err)))
			continue
		}
		data := map[string]interface{}{}
		if err := json.Unmarshal(payload, &data); err != nil {
			return nil, err
		}
		return data, nil
	}
	return nil, errors.New("zero jwks matches")
}

func (a *Auth) Expressions() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.expressions
}

func (a *Auth) Programs() []cel.Program {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.programs
}

func (a *Auth) RefreshKeys() error {
	for uri, _ := range a.set {
		set, err := jwk.Fetch(uri)
		if err != nil {
			return err
		}
		a.mu.Lock()
		a.set[uri] = set
		a.mu.Unlock()
	}
	return nil
}

func (a *Auth) Override(auth *apipb.AuthConfig) error {
	for _, source := range auth.JwksSources {
		set, err := jwk.Fetch(source)
		if err != nil {
			return err
		}
		a.mu.Lock()
		a.set[source] = set
		a.mu.Unlock()
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.expressions = auth.AuthExpressions
	if len(a.expressions) > 0 && a.expressions[0] != "" {
		programs, err := express.Programs(a.expressions)
		if err != nil {
			return err
		}
		a.programs = programs
	}
	return nil
}

func (a *Auth) Raw() *apipb.AuthConfig {
	var jwksSources []string
	a.mu.RLock()
	defer a.mu.RUnlock()
	for source, _ := range a.set {
		jwksSources = append(jwksSources, source)
	}
	return &apipb.AuthConfig{
		JwksSources:     jwksSources,
		AuthExpressions: a.expressions,
	}
}

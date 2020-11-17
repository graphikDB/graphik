package auth

import (
	"encoding/json"
	"fmt"
	"github.com/autom8ter/graphik/logger"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sync"
)

type Config struct {
	mu      sync.RWMutex
	jwksSet map[string]*jwk.Set
}

func New(jwks []string) (*Config, error) {
	setMap := map[string]*jwk.Set{}
	for _, source := range jwks {
		set, err := jwk.Fetch(source)
		if err != nil {
			return nil, err
		}
		setMap[source] = set
	}
	c := &Config{
		mu:      sync.RWMutex{},
		jwksSet: setMap,
	}
	return c, nil
}

func (a *Config) VerifyJWT(token string) (map[string]interface{}, error) {
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
	for uri, set := range a.jwksSet {
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

func (a *Config) RefreshKeys() error {
	for uri, _ := range a.jwksSet {
		set, err := jwk.Fetch(uri)
		if err != nil {
			return err
		}
		a.mu.Lock()
		a.jwksSet[uri] = set
		a.mu.Unlock()
	}
	return nil
}

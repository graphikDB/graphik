package jwks

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

func (a *Auth) Override(uris []string) error {
	sets := map[string]*jwk.Set{}
	for _, uri := range uris {
		set, err := jwk.Fetch(uri)
		if err != nil {
			return err
		}
		sets[uri] = set
	}
	a.mu.Lock()
	a.set = sets
	a.mu.Unlock()
	return nil
}

func (a *Auth) List() []string {
	var list []string
	a.mu.RLock()
	defer a.mu.RUnlock()
	for key, _ := range a.set {
		list = append(list, key)
	}
	return list
}
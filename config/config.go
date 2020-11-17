package config

import (
	"encoding/json"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/graphik/vm"
	"github.com/google/cel-go/cel"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sync"
)

type Config struct {
	mu                 sync.RWMutex
	jwksSet            map[string]*jwk.Set
	authExpressions    []string
	authPrograms       []cel.Program
	triggerExpressions []string
	triggerPrograms    []cel.Program
	source             *apipb.RuntimeConfig
}

func New(cfg *apipb.RuntimeConfig) (*Config, error) {
	if cfg.Trigger == nil {
		cfg.Trigger = &apipb.TriggerConfig{}
	}
	if cfg.Auth == nil {
		cfg.Auth = &apipb.AuthConfig{}
	}
	setMap := map[string]*jwk.Set{}
	for _, source := range cfg.GetAuth().GetJwksSources() {
		set, err := jwk.Fetch(source)
		if err != nil {
			return nil, err
		}
		setMap[source] = set
	}
	c := &Config{
		mu:                 sync.RWMutex{},
		jwksSet:            setMap,
		authExpressions:    cfg.GetAuth().GetAuthExpressions(),
		triggerExpressions: nil,
		source:             cfg,
	}
	if len(c.authExpressions) > 0 && c.authExpressions[0] != "" {
		programs, err := vm.Programs(cfg.GetAuth().GetAuthExpressions())
		if err != nil {
			return nil, err
		}
		c.authPrograms = programs
	}
	if len(c.triggerExpressions) > 0 && c.triggerExpressions[0] != "" {
		programs, err := vm.Programs(cfg.GetTrigger().GetExpressions())
		if err != nil {
			return nil, err
		}
		c.triggerPrograms = programs
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

func (a *Config) AuthExpressions() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.authExpressions
}

func (a *Config) AuthPrograms() []cel.Program {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.authPrograms
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

func (a *Config) Override(config *apipb.RuntimeConfig) error {
	if config.Trigger == nil {
		config.Trigger = &apipb.TriggerConfig{}
	}
	if config.Auth == nil {
		config.Auth = &apipb.AuthConfig{}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, source := range config.GetAuth().JwksSources {
		set, err := jwk.Fetch(source)
		if err != nil {
			return err
		}
		a.jwksSet[source] = set
	}

	a.authExpressions = config.GetAuth().GetAuthExpressions()
	if len(a.authExpressions) > 0 && a.authExpressions[0] != "" {
		programs, err := vm.Programs(a.authExpressions)
		if err != nil {
			return err
		}
		a.authPrograms = programs
	}

	a.triggerExpressions = config.GetTrigger().GetExpressions()
	if len(a.triggerExpressions) > 0 && a.triggerExpressions[0] != "" {
		programs, err := vm.Programs(a.triggerExpressions)
		if err != nil {
			return err
		}
		a.triggerPrograms = programs
	}
	a.source = config
	return nil
}

func (a *Config) Config() *apipb.RuntimeConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.source
}

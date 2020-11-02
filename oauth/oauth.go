package oauth

import (
	"encoding/json"
	"github.com/autom8ter/graphik/config"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"golang.org/x/oauth2"
	"net/http"
	"sync"
)

type Metadata struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
}

func New(config *config.Auth) (*Auth, error) {
	resp, err := http.DefaultClient.Get(config.DiscoveryUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var md Metadata
	if err := json.NewDecoder(resp.Body).Decode(&md); err != nil {
		return nil, err
	}
	set, err := jwk.Fetch(md.JwksURI)
	if err != nil {
		return nil, err
	}
	return &Auth{
		oauth: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  md.AuthorizationEndpoint,
				TokenURL: md.TokenEndpoint,
			},
			RedirectURL: config.RedirectURL,
			Scopes:      config.Scopes,
		},
		set:      set,
		metadata: &md,
		mu:       sync.RWMutex{},
	}, nil
}

type Auth struct {
	mu       sync.RWMutex
	oauth    *oauth2.Config
	set      *jwk.Set
	metadata *Metadata
}

func (a *Auth) VerifyJWT(token string) ([]byte, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return jws.VerifyWithJWKSet([]byte(token), a.set, nil)
}

func (a *Auth) ParseJWT(token string) (*jws.Message, error) {
	return jws.ParseString(token)
}

func (a *Auth) Refresh() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	set, err := jwk.Fetch(a.metadata.JwksURI)
	if err != nil {
		return err
	}
	a.set = set
	return nil
}

package oauth

import (
	"encoding/json"
	"github.com/autom8ter/graphik/config"
	"golang.org/x/oauth2"
	"net/http"
)

type Key struct {
	Use string `json:"use"`
	E   string `json:"e"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	Kid string `json:"kid"`
}

type Keys struct {
	Keys []*Key `json:"keys"`
}

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
	resp, err = http.DefaultClient.Get(md.JwksURI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var keys Keys
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
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
		keys:     &keys,
		metadata: &md,
	}, nil
}

type Auth struct {
	oauth    *oauth2.Config
	keys     *Keys
	metadata *Metadata
}

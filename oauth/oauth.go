package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/autom8ter/graphik/config"
	"github.com/gorilla/sessions"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	sessionName = "x-graphik-session"
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
		store:    sessions.NewCookieStore([]byte(config.SessionSecret)),
	}, nil
}

type Auth struct {
	mu       sync.RWMutex
	oauth    *oauth2.Config
	set      *jwk.Set
	metadata *Metadata
	store    *sessions.CookieStore
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

func (a *Auth) ClientCredentials(ctx context.Context, values url.Values) oauth2.TokenSource {
	c := clientcredentials.Config{
		ClientID:       a.oauth.ClientID,
		ClientSecret:   a.oauth.ClientSecret,
		TokenURL:       a.metadata.TokenEndpoint,
		Scopes:         a.oauth.Scopes,
		EndpointParams: values,
	}
	return c.TokenSource(ctx)
}

func (a *Auth) AuthorizationCode(ctx context.Context, code string) (oauth2.TokenSource, error) {
	token, err := a.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return a.oauth.TokenSource(ctx, token), nil
}

func (a *Auth) Client(ctx context.Context, source oauth2.TokenSource) *http.Client {
	return oauth2.NewClient(ctx, source)
}

func (a *Auth) Session(r *http.Request, w http.ResponseWriter, fn func(sess *sessions.Session) error) error {
	sess, err := a.store.Get(r, sessionName)
	if err != nil {
		return err
	}
	if err := fn(sess); err != nil {
		return err
	}
	return sessions.Save(r, w)
}

func (a *Auth) Middleware() func(handler http.HandlerFunc) http.HandlerFunc {
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			tokenSplit := strings.Split(auth, "Bearer ")
			if len(tokenSplit) == 2 {
				payload, err := a.VerifyJWT(tokenSplit[1])
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
				a.Session(r, w, func(sess *sessions.Session) error {
					sess.Values["token"] = tokenSplit[1]
					sess.Values["payload"] = string(payload)
					return nil
				})
				handler.ServeHTTP(w, r)
			} else {
				if err := a.Session(r, w, func(sess *sessions.Session) error {
					if token := sess.Values["token"]; token == nil {
						return errors.New("session token not found")
					} else {
						_, err := a.VerifyJWT(token.(string))
						if err != nil {
							return err
						}
					}
					return nil
				}); err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
				handler.ServeHTTP(w, r)
			}
		}
	}
}

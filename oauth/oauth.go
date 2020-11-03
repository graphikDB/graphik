package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/autom8ter/graphik/config"
	"github.com/gorilla/sessions"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
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
	return sess.Save(r, w)
}

func (a *Auth) Middleware() func(handler http.HandlerFunc) http.HandlerFunc {
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			tokenString := strings.Replace(authHeader, "Bearer ", "", -1)

			if tokenString != "" {
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
				if err := a.Session(r, w, func(sess *sessions.Session) error {
					for k, v := range payload {
						sess.Values[k] = v
					}
					return nil
				}); err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
				handler.ServeHTTP(w, r)
			} else {
				if err := a.Session(r, w, func(sess *sessions.Session) error {
					if len(sess.Values) == 0 {
						return errors.New("empty session")
					}
					if exp, ok := sess.Values["exp"].(int64); ok {
						if exp < time.Now().Unix() {
							return errors.New("session expired")
						}
					}
					if exp, ok := sess.Values["exp"].(int); ok {
						if int64(exp) < time.Now().Unix() {
							return errors.New("session expired")
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

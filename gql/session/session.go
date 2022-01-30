package session

import (
	"github.com/graphikDB/graphik/gql/session/cookie"
	"github.com/graphikDB/graphik/gql/session/token"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"net/http"
)

type SessionManager interface {
	GetToken(req *http.Request) (*token.Token, error)
	GetState(req *http.Request) (string, error)
	RedirectLogin(w http.ResponseWriter, req *http.Request) error
	Exchange(w http.ResponseWriter, req *http.Request, code string) error
}

type SessionName string

const (
	COOKIES    SessionName = "cookies"
	cookieName             = "graphik-ui"
)

var AllManagers = []SessionName{COOKIES}

func GetSessionManager(authconfig *oauth2.Config, config map[string]string) (SessionManager, error) {
	name := config["name"]
	switch SessionName(name) {
	case COOKIES:
		secret := config["secret"]
		if secret == "" {
			return nil, errors.New("failed to get session manager: empty secret")
		}
		return cookie.New(cookieName, authconfig, config["session_secret"]), nil
	default:
		return nil, errors.Errorf("unsupported session manager: %s. must be one of: %v", name, AllManagers)
	}
}

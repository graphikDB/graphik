package cookie

import (
	"encoding/hex"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/graphikDB/graphik/gql/session/token"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"math/rand"
	"net/http"
)

type cookieManager struct {
	cookieName string
	config     *oauth2.Config
	store      *sessions.CookieStore
}

func New(cookieName string, config *oauth2.Config, secret string) *cookieManager {
	return &cookieManager{cookieName: cookieName, config: config, store: sessions.NewCookieStore([]byte(secret))}
}

func (s *cookieManager) GetToken(req *http.Request) (*token.Token, error) {
	sess, err := s.store.Get(req, s.cookieName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get session")
	}
	if sess.Values["token"] == nil {
		return nil, errors.New("failed to get session token")
	}
	val, ok := sess.Values["token"].(*token.Token)
	if !ok || val == nil {
		return nil, errors.New("failed to get session token")
	}

	toke, err := s.config.TokenSource(req.Context(), val.Token).Token()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get session")
	}
	val.Token = toke
	idtoken, ok := val.Token.Extra("id_token").(string)
	if ok {
		val.IDToken = idtoken
	}
	return val, nil
}

func (s *cookieManager) setToken(w http.ResponseWriter, req *http.Request, toke *oauth2.Token) error {
	sess, err := s.store.Get(req, s.cookieName)
	if err != nil {
		return errors.Wrap(err, "failed to get session")
	}
	t := &token.Token{
		Token:   toke,
		IDToken: "",
	}
	idtoken, ok := toke.Extra("id_token").(string)
	if ok {
		t.IDToken = idtoken
	}
	sess.Values["token"] = t
	return sess.Save(req, w)
}

func (s *cookieManager) GetState(req *http.Request) (string, error) {
	sess, err := s.store.Get(req, s.cookieName)
	if err != nil {
		return "", errors.Wrap(err, "failed to get session")
	}
	if sess.Values["state"] == nil {
		return "", errors.New("empty session state")
	}
	return sess.Values["state"].(string), nil
}

func (s *cookieManager) RedirectLogin(w http.ResponseWriter, req *http.Request) error {
	sess, err := s.store.Get(req, s.cookieName)
	if err != nil {
		return errors.Wrap(err, "failed to get session")
	}
	state := hex.EncodeToString([]byte(fmt.Sprint(rand.Int())))
	sess.Values["state"] = state
	if err := sess.Save(req, w); err != nil {
		return errors.Wrap(err, "failed to save session state")
	}
	redirect := s.config.AuthCodeURL(state)
	http.Redirect(w, req, redirect, http.StatusTemporaryRedirect)
	return nil
}

func (s *cookieManager) Exchange(w http.ResponseWriter, req *http.Request, code string) error {
	toke, err := s.config.Exchange(req.Context(), code)
	if err != nil {
		return errors.Wrap(err, "failed to get token")
	}
	return s.setToken(w, req, toke)
}

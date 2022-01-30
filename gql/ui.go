package gql

import (
	"go.uber.org/zap"
	"net/http"
)

func (p *Resolver) UIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		authToken, err := p.session.GetToken(req)
		if err != nil {
			p.logger.Error("ui: failed to get session - redirecting", zap.Error(err))
			if err := p.session.RedirectLogin(w, req); err != nil {
				p.logger.Error("ui: failed to redirect", zap.Error(err))
			}
			return
		}
		if authToken == nil {
			if err := p.session.RedirectLogin(w, req); err != nil {
				p.logger.Error("ui: failed to redirect", zap.Error(err))
			}
			return
		}
		if !authToken.Token.Valid() {
			if err := p.session.RedirectLogin(w, req); err != nil {
				p.logger.Error("ui: failed to redirect", zap.Error(err))
			}
			return
		}
		w.Header().Add("Content-Type", "text/html")
		playground.Execute(w, map[string]string{})
	}
}

func (p *Resolver) OAuthCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		code := req.URL.Query().Get("code")
		state := req.URL.Query().Get("state")
		if code == "" {
			p.logger.Error("ui: empty authorization code - redirecting")
			if err := p.session.RedirectLogin(w, req); err != nil {
				p.logger.Error("ui: failed to redirect", zap.Error(err))
			}
			return
		}
		if state == "" {
			p.logger.Error("ui: empty authorization state - redirecting")
			if err := p.session.RedirectLogin(w, req); err != nil {
				p.logger.Error("ui: failed to redirect", zap.Error(err))
			}
			return
		}

		stateVal, err := p.session.GetState(req)
		if err != nil {
			p.logger.Error("ui: failed to get session state - redirecting", zap.Error(err))
			if err := p.session.RedirectLogin(w, req); err != nil {
				p.logger.Error("ui: failed to redirect", zap.Error(err))
			}
			return
		}
		if stateVal != state {
			p.logger.Error("ui: session state mismatch - redirecting")
			if err := p.session.RedirectLogin(w, req); err != nil {
				p.logger.Error("ui: failed to redirect", zap.Error(err))
			}
			return
		}

		if err := p.session.Exchange(w, req, code); err != nil {
			p.logger.Error("ui: failed to exchange authorization code - redirecting", zap.Error(err))
			if err := p.session.RedirectLogin(w, req); err != nil {
				p.logger.Error("ui: failed to redirect", zap.Error(err))
			}
			return
		}
		http.Redirect(w, req, p.uiPath, http.StatusTemporaryRedirect)
	}
}

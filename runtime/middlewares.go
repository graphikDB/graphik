package runtime

import (
	"net/http"
	"strings"
	"time"
)

func (a *Store) AuthMiddleware() func(handler http.Handler) http.HandlerFunc {
	return func(handler http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			tokenString := strings.Replace(authHeader, "Bearer ", "", -1)
			if tokenString == "" {
				http.Error(w, "empty authorization header", http.StatusUnauthorized)
				return
			}
			payload, err := a.opts.jwks.VerifyJWT(tokenString)
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
			handler.ServeHTTP(w, a.toContext(r, payload))
		}
	}
}

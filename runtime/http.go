package runtime

import (
	"encoding/json"
	"net/http"
)

func (s *Store) Join() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := map[string]string{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(m) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		remoteAddr, ok := m["addr"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		nodeID, ok := m["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := s.join(nodeID, remoteAddr); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}


func (a *Store) PutJWKS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var uris []string
		json.NewDecoder(r.Body).Decode(&uris)
		if err := a.opts.jwks.Override(uris); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func (a *Store) GetJWKS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwks := a.opts.jwks.List()
		json.NewEncoder(w).Encode(&jwks)
	}
}


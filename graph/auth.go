package graph

import (
	"encoding/json"
	"fmt"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/pkg/errors"
)

func (g *GraphStore) VerifyJWT(token string) (map[string]interface{}, error) {
	message, err := jws.ParseString(token)
	if err != nil {
		return nil, err
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.jwksSet == nil {
		data := map[string]interface{}{}
		if err := json.Unmarshal(message.Payload(), &data); err != nil {
			return nil, err
		}
		return data, nil
	}
	if len(message.Signatures()) == 0 {
		return nil, fmt.Errorf("zero jws signatures")
	}
	kid, ok := message.Signatures()[0].ProtectedHeaders().Get("kid")
	if !ok {
		return nil, fmt.Errorf("jws kid not found")
	}
	algI, ok := message.Signatures()[0].ProtectedHeaders().Get("alg")
	if !ok {
		return nil, fmt.Errorf("jw alg not found")
	}
	alg, ok := algI.(jwa.SignatureAlgorithm)
	if !ok {
		return nil, fmt.Errorf("alg type cast error")
	}
	keys := g.jwksSet.LookupKeyID(kid.(string))
	if len(keys) == 0 {
		return nil, errors.Errorf("failed to lookup kid: %s - zero keys", kid.(string))
	}
	var key interface{}
	if err := keys[0].Raw(&key); err != nil {
		return nil, err
	}
	payload, err := jws.Verify([]byte(token), alg, key)
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, err
	}
	return data, nil
}

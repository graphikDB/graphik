package database

import (
	"fmt"
	apipb "github.com/graphikDB/graphik/gen/grpc/go"
	"strings"
)

type openIDConnect struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	DeviceAuthorizationEndpoint       string   `json:"device_authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
	RevocationEndpoint                string   `json:"revocation_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
}

func refString(ref *apipb.Ref) string {
	if ref.GetGid() == "" {
		return ref.GetGtype()
	}
	return fmt.Sprintf("%s////%s", ref.GetGtype(), ref.GetGid())
}

func fromRefString(ref string) *apipb.Ref {
	split := strings.Split(ref, "////")
	if len(split) == 0 {
		return nil
	}
	if len(split) == 1 {
		return &apipb.Ref{
			Gtype: split[0],
			Gid:   "",
		}
	}
	return &apipb.Ref{
		Gtype: split[0],
		Gid:   split[1],
	}
}

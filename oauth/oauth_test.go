package oauth_test

import (
	"github.com/autom8ter/graphik/config"
	"github.com/autom8ter/graphik/oauth"
	"testing"
)

const token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IlZPc293V0ktZDNyVUdPWWVxeGQ2OTVEZ0MzNCJ9.eyJhdWQiOiIwNzM1ZmNlZi0wMmRiLTQ2ZDQtODQ2NC0xOTQ2NTQwMTYzZjAiLCJleHAiOjE2MDQzNzczMDAsImlhdCI6MTYwNDM3MzcwMCwiaXNzIjoiaHR0cHM6Ly9sb2dpbi5kZXYuZnJvbnRkb29yaG9tZS5jb20iLCJzdWIiOiJlY2VmOWY5YS00MTAwLTQ4MTItYmQzOC04M2E2NTQ3YTYxODUiLCJqdGkiOiJhNWM1ZDZhOS1kMDJiLTZiZWUtYjc2OC1lZGEwODg0NGU2NzkiLCJhdXRoZW50aWNhdGlvblR5cGUiOiJQQVNTV09SRCIsImVtYWlsIjoiY29sZW1hbndvcmRAZ21haWwuY29tIiwiZW1haWxfdmVyaWZpZWQiOmZhbHNlLCJhdF9oYXNoIjoiY0tIOS1XTHJSYmc2NzNpMFhBWTU4USIsImNfaGFzaCI6IlRJTWlodjZHdTBZZ3otejlzOG5wemciLCJhcHBsaWNhdGlvbklkIjoiMDczNWZjZWYtMDJkYi00NmQ0LTg0NjQtMTk0NjU0MDE2M2YwIiwicm9sZXMiOlsicGFydG5lciJdLCJzaWQiOiJiZjA0N2YwMC0xNGY1LTRjNWQtYjVjZS0zNDBkOWYwMmU4OTQiLCJ2ZXIiOiIxLjAuMCIsInRlbmFudCI6ImVlMzEyZTc3LTg0NjMtNDRiZC1hZDdlLTJjZDRlNzVjOWUzZCJ9.bsXE20G5FWKhysbJwt-d2GkncOdnOCPxACQmlS9ojg4vY5OpMBRda738_x6Qx2w8ET_8CsHRAs3W2Wu76cOhkl-fy46NvmhjX4SDNXnXLxhBIyyJw8vWIfv8HfbMuIJMdQnnVtWxQcr3PBvIKrBB26i9OiLgTQfx8c1zFZvMCC6-e5lPxXEAwf7iYIoDwrgQtW8FRAZ8odFYxYiwXNFzMMRniOemGqJsa5v7-ofU5OvxVOXSXJpVGsB2EhPBLqYyaldCV3QzWI9kYJQj93B3uLpJ29t2wPg6vV45yegEkFuI1QT3bvHxT7zAIli0iQLpY3v65Tqzx9a_9u3yI3F2sg"

func Test(t *testing.T) {
	auth, err := oauth.New(&config.Auth{
		ClientID:      "0735fcef-02db-46d4-8464-1946540163f0",
		ClientSecret:  "",
		DiscoveryUrl:  "https://frontdoorhome-dev.fusionauth.io/.well-known/openid-configuration/94510e59-3c18-4e7a-bab0-a1828913b9c3",
		RedirectURL:   "",
		Scopes:        []string{"api://ftdrctl-sdk/nonce"},
		SessionSecret: "testing",
	})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := auth.VerifyJWT(token)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(payload)
}

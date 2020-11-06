package jwks_test

import (
	"github.com/autom8ter/graphik/jwks"
	"testing"
)

const token = ""

func Test(t *testing.T) {
	auth := jwks.New()
	payload, err := auth.VerifyJWT(token)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(payload)
}

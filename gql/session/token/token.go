package token

import (
	"encoding/gob"
	"golang.org/x/oauth2"
)

func init() {
	gob.Register(&Token{})
}

type Token struct {
	Token   *oauth2.Token
	IDToken string
}

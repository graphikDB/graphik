package gql

import "github.com/pkg/errors"

var (
	ErrNoTokenSession = errors.New("playground: no token session")
	ErrNoStateSession = errors.New("playground: no state session")
)

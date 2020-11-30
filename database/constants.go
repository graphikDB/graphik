package database

import "errors"

type ctxKey string

const (
	// Permissions to use on the db file. This is only used if the
	// database file does not exist and needs to be created.
	dbFileMode                  = 0600
	changeChannel               = "changes"
	authCtxKey           ctxKey = "x-graphik-auth-ctx"
	userType             ctxKey = "user"
	methodCtxKey         ctxKey = "x-graphik-full-method"
	importOverrideCtxKey ctxKey = "x-graphik-import-override"
)

var (
	DONE = errors.New("DONE")
	// Bucket names we perform transactions in
	dbConnections      = []byte("connections")
	dbDocs             = []byte("docs")
	dbIndexes          = []byte("indexes")
	dbAuthorizers      = []byte("authorizers")
	dbTypeValidators   = []byte("typeValidators")
	dbIndexDocs        = []byte("indexedDocs")
	dbIndexConnections = []byte("indexedConnections")
	// An error indicating a given key does not exist
	ErrNotFound             = errors.New("not found")
	ErrAlreadyExists        = errors.New("already exists")
	ErrUnsupportedAlgorithm = errors.New("unsupported algorithm")
)

package database

import "errors"

type ctxKey string

const (
	// Permissions to use on the db file. This is only used if the
	// database file does not exist and needs to be created.
	dbFileMode                     = 0600
	changeChannel                  = "state"
	authCtxKey              ctxKey = "x-graphik-auth-ctx"
	userType                ctxKey = "user"
	methodCtxKey            ctxKey = "x-graphik-full-method"
	bypassAuthorizersCtxKey ctxKey = "x-graphik-authorzer-override"
	tokenCtxKey             ctxKey = "x-graphik-token"
	raftSecretMDKey                = "x-graphik-raft-secret"
)

var (
	DONE = errors.New("DONE")
	// Bucket names we perform transactions in
	dbConnections      = []byte("connections")
	dbDocs             = []byte("docs")
	dbIndexes          = []byte("indexes")
	dbAuthorizers      = []byte("authorizers")
	dbTriggers         = []byte("triggers")
	dbConstraints      = []byte("constraints")
	dbIndexDocs        = []byte("indexedDocs")
	dbIndexConnections = []byte("indexedConnections")
	dbMessages         = []byte("messages")
	// An error indicating a given key does not exist
	ErrNotFound             = errors.New("not found")
	ErrAlreadyExists        = errors.New("already exists")
	ErrUnsupportedAlgorithm = errors.New("unsupported algorithm")
	ErrFailedToGetUser      = errors.New("failed to get user")
)

package database

import "errors"

const (
	// Permissions to use on the db file. This is only used if the
	// database file does not exist and needs to be created.
	dbFileMode    = 0600
	changeChannel = "changes"
)

var (
	DONE = errors.New("DONE")
	// Bucket names we perform transactions in
	dbConnections      = []byte("connections")
	dbDocs             = []byte("docs")
	dbIndexes          = []byte("indexes")
	dbIndexDocs        = []byte("docs/index")
	dbIndexConnections = []byte("connections/index")
	// An error indicating a given key does not exist
	ErrNotFound = errors.New("not found")
)

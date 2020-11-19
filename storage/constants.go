package storage

import "errors"

const (
	// Permissions to use on the db file. This is only used if the
	// database file does not exist and needs to be created.
	dbFileMode = 0600
)

var (
	// Bucket names we perform transactions in
	dbLogs  = []byte("logs")
	dbSnaps = []byte("snapshots")
	dbEdges = []byte("edges")
	dbNodes = []byte("nodes")
	// An error indicating a given key does not exist
	ErrNotFound = errors.New("not found")
)

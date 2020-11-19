package storage

import (
	"encoding/binary"
	"github.com/pkg/errors"
)

const (
	// Permissions to use on the db file. This is only used if the
	// database file does not exist and needs to be created.
	DbFileMode = 0600
)

var (
	DONE = errors.New("DONE")
	// Bucket names we perform transactions in
	DbLogs  = []byte("logs")
	DbSnaps = []byte("snapshots")
	DbEdges = []byte("edges")
	DbNodes = []byte("nodes")
	// An error indicating a given key does not exist
	ErrNotFound = errors.New("not found")
)

// Converts bytes to an integer
func BytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// Converts a uint to a byte slice
func Uint64ToBytes(u uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, u)
	return buf
}

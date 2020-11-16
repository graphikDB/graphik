package storage

import (
	"go.etcd.io/bbolt"
)

// SnapshotStore provides access to BoltDB for Raft to store and retrieve
// log entries. It also provides key/value storage, and can be used as
// a SnapshotStore.
type SnapshotStore struct {
	// conn is the underlying handle to the db.
	conn *bbolt.DB

	// The path to the Bolt database file
	path string
}

// NewSnapshotStore takes a file path and returns a connected Raft backend.
func NewSnapshotStore(path string) (*SnapshotStore, error) {
	handle, err := bbolt.Open(path, dbFileMode, nil)
	if err != nil {
		return nil, err
	}
	tx, err := handle.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create all the buckets
	if _, err := tx.CreateBucketIfNotExists(dbSnaps); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &SnapshotStore{
		conn: handle,
		path: path,
	}, nil
}

// Close is used to gracefully close the DB connection.
func (b *SnapshotStore) Close() error {
	return b.conn.Close()
}

// Set is used to set a key/value set outside of the raft log
func (b *SnapshotStore) Set(k, v []byte) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(dbSnaps)
	if err := bucket.Put(k, v); err != nil {
		return err
	}

	return tx.Commit()
}

// Get is used to retrieve a value from the k/v store by key
func (b *SnapshotStore) Get(k []byte) ([]byte, error) {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(dbSnaps)
	val := bucket.Get(k)

	if val == nil {
		return nil, ErrKeyNotFound
	}
	return append([]byte(nil), val...), nil
}

// SetUint64 is like Set, but handles uint64 values
func (b *SnapshotStore) SetUint64(key []byte, val uint64) error {
	return b.Set(key, uint64ToBytes(val))
}

// GetUint64 is like Get, but handles uint64 values
func (b *SnapshotStore) GetUint64(key []byte) (uint64, error) {
	val, err := b.Get(key)
	if err != nil {
		return 0, err
	}
	return bytesToUint64(val), nil
}

// Sync performs an fsync on the database file handle. This is not necessary
// under normal operation unless NoSync is enabled, in which this forces the
// database file to sync against the disk.
func (b *SnapshotStore) Sync() error {
	return b.conn.Sync()
}

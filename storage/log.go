package storage

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	"go.etcd.io/bbolt"
)

// LogStore provides access to BoltDB for Raft to store and retrieve
// log entries a LogStore .
type LogStore struct {
	// conn is the underlying handle to the db.
	conn *bbolt.DB

	// The path to the Bolt database file
	path string
}

// NewLogStore takes a file path and returns a connected Raft backend.
func NewLogStore(path string) (*LogStore, error) {
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
	if _, err := tx.CreateBucketIfNotExists(dbLogs); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &LogStore{
		conn: handle,
		path: path,
	}, nil
}

// Close is used to gracefully close the DB connection.
func (b *LogStore) Close() error {
	return b.conn.Close()
}

// FirstIndex returns the first known index from the Raft log.
func (b *LogStore) FirstIndex() (uint64, error) {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	curs := tx.Bucket(dbLogs).Cursor()
	if first, _ := curs.First(); first == nil {
		return 0, nil
	} else {
		return bytesToUint64(first), nil
	}
}

// LastIndex returns the last known index from the Raft log.
func (b *LogStore) LastIndex() (uint64, error) {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	curs := tx.Bucket(dbLogs).Cursor()
	if last, _ := curs.Last(); last == nil {
		return 0, nil
	} else {
		return bytesToUint64(last), nil
	}
}

// GetLog is used to retrieve a log from BoltDB at a given index.
func (b *LogStore) GetLog(idx uint64, log *raft.Log) error {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(dbLogs)
	val := bucket.Get(uint64ToBytes(idx))

	if val == nil {
		return raft.ErrLogNotFound
	}
	l := &apipb.RaftLog{}
	if err := proto.Unmarshal(val, l); err != nil {
		return err
	}
	log.Extensions = l.Extensions
	log.Data = l.Data
	log.Type = raft.LogType(l.Type)
	log.Term = l.Term
	log.Index = l.Index
	return nil
}

// StoreLog is used to store a single raft log
func (b *LogStore) StoreLog(log *raft.Log) error {
	return b.StoreLogs([]*raft.Log{log})
}

// StoreLogs is used to store a set of raft logs
func (b *LogStore) StoreLogs(logs []*raft.Log) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, log := range logs {
		key := uint64ToBytes(log.Index)
		protoLog := &apipb.RaftLog{
			Index:      log.Index,
			Term:       log.Term,
			Type:       uint32(log.Type),
			Data:       log.Data,
			Extensions: log.Extensions,
		}
		bits, err := proto.Marshal(protoLog)
		if err != nil {
			return err
		}
		bucket := tx.Bucket(dbLogs)
		if err := bucket.Put(key, bits); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// DeleteRange is used to delete logs within a given range inclusively.
func (b *LogStore) DeleteRange(min, max uint64) error {
	minKey := uint64ToBytes(min)

	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	curs := tx.Bucket(dbLogs).Cursor()
	for k, _ := curs.Seek(minKey); k != nil; k, _ = curs.Next() {
		// Handle out-of-range log index
		if bytesToUint64(k) > max {
			break
		}

		// Delete in-range log index
		if err := curs.Delete(); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Sync performs an fsync on the database file handle. This is not necessary
// under normal operation unless NoSync is enabled, in which this forces the
// database file to sync against the disk.
func (b *LogStore) Sync() error {
	return b.conn.Sync()
}

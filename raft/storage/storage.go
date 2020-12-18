package storage

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/graphikDB/graphik/helpers"
	"go.etcd.io/bbolt"

	"github.com/hashicorp/raft"
)

func init() {
	gob.Register(&raft.Log{})
}

var (
	// Bucket names we perform transactions in
	dbLogs = []byte("raftLogs")
	dbConf = []byte("raftConfig")
	// An error indicating a given key does not exist
	ErrKeyNotFound = errors.New("not found")
)

type Storage struct {
	db *bbolt.DB
}

func (s *Storage) implementsLogStore() raft.LogStore {
	return s
}

func (s *Storage) implementsStableStore() raft.StableStore {
	return s
}

func NewStorage(path string) (*Storage, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	if err := db.Update(func(tx *bbolt.Tx) error {
		// Create all the buckets
		if _, err := tx.CreateBucketIfNotExists(dbLogs); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(dbConf); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &Storage{
		db: db,
	}, nil
}

func (b *Storage) Close() error {
	return b.db.Close()
}

func (b *Storage) FirstIndex() (uint64, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	curs := tx.Bucket(dbLogs).Cursor()
	if first, _ := curs.First(); first == nil {
		return 0, nil
	} else {
		return helpers.BytesToUint64(first), nil
	}
}

func (b *Storage) LastIndex() (uint64, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	curs := tx.Bucket(dbLogs).Cursor()
	if last, _ := curs.Last(); last == nil {
		return 0, nil
	} else {
		return helpers.BytesToUint64(last), nil
	}
}

func (b *Storage) GetLog(idx uint64, log *raft.Log) error {
	tx, err := b.db.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(dbLogs)
	val := bucket.Get(helpers.Uint64ToBytes(idx))

	if val == nil {
		return raft.ErrLogNotFound
	}
	return gob.NewDecoder(bytes.NewBuffer(val)).Decode(log)
}

func (b *Storage) StoreLog(log *raft.Log) error {
	return b.StoreLogs([]*raft.Log{log})
}

func (b *Storage) StoreLogs(logs []*raft.Log) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		buf := bytes.NewBuffer(nil)
		for _, log := range logs {
			key := helpers.Uint64ToBytes(log.Index)
			if err := gob.NewEncoder(buf).Encode(log); err != nil {
				return err
			}
			bucket := tx.Bucket(dbLogs)
			if err := bucket.Put(key, buf.Bytes()); err != nil {
				return err
			}
			buf.Reset()
		}
		return nil
	})
}

func (b *Storage) DeleteRange(min, max uint64) error {
	minKey := helpers.Uint64ToBytes(min)
	return b.db.Update(func(tx *bbolt.Tx) error {
		curs := tx.Bucket(dbLogs).Cursor()
		for k, _ := curs.Seek(minKey); k != nil; k, _ = curs.Next() {
			// Handle out-of-range log index
			if helpers.BytesToUint64(k) > max {
				break
			}
			// Delete in-range log index
			if err := curs.Delete(); err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *Storage) Set(k, v []byte) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbConf)
		return bucket.Put(k, v)
	})
}

func (b *Storage) Get(k []byte) ([]byte, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(dbConf)
	val := bucket.Get(k)

	if val == nil {
		return nil, ErrKeyNotFound
	}
	return append([]byte(nil), val...), nil
}

func (b *Storage) SetUint64(key []byte, val uint64) error {
	return b.Set(key, helpers.Uint64ToBytes(val))
}

func (b *Storage) GetUint64(key []byte) (uint64, error) {
	val, err := b.Get(key)
	if err != nil {
		return 0, err
	}

	return helpers.BytesToUint64(val), nil
}

func (b *Storage) Sync() error {
	return b.db.Sync()
}

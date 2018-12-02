package boltdb

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/etcd-io/bbolt"
	"h12.io/expay"
)

type (
	// DB represents one boltdb file supporting multiple buckets
	DB struct {
		db *bolt.DB
	}
	// Bucket represents a boltdb bucket that satisifies expay.DB interface
	Bucket struct {
		name string
		db   *bolt.DB
	}
	iter struct {
		tx     *bolt.Tx
		cursor *bolt.Cursor
		key    []byte
		value  []byte
	}
)

// New creates or opens a boltdb file
func New(filename string) (*DB, error) {
	db, err := bolt.Open(filename, 0666, nil)
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

// Bucket returns a bucket from boltdb
func (db *DB) Bucket(name string) *Bucket {
	return &Bucket{name: name, db: db.db}
}

// Create creates a new value into the bucket
func (b *Bucket) Create(v interface{}) (id string, err error) {
	value, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	err = b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(b.name))
		if err != nil {
			return err
		}
		seq, err := bucket.NextSequence()
		if err != nil {
			return err
		}
		key := itob(seq)
		if err := bucket.Put(key, value); err != nil {
			return err
		}
		id = hex.EncodeToString(key)
		return nil
	})
	return id, err
}

// itob returns an 8-byte big endian representation of i
func itob(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

// Get gets a value from the bucket given the id
func (b *Bucket) Get(id string, v interface{}) error {
	key, err := hex.DecodeString(id)
	if err != nil {
		return err
	}
	var value []byte
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b.name))
		if bucket == nil {
			return expay.ErrNotFound
		}
		value = bucket.Get(key)
		return nil
	})
	if err != nil {
		return err
	}
	if value == nil {
		return expay.ErrNotFound
	}
	return json.Unmarshal(value, v)
}

// Update updates a value given the id
func (b *Bucket) Update(id string, v interface{}) error {
	key, err := hex.DecodeString(id)
	if err != nil {
		return err
	}
	value, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(b.name))
		if err != nil {
			return err
		}
		return bucket.Put(key, value)
	})
}

// Delete deletes an id from the bucket, returns nil if not exists
func (b *Bucket) Delete(id string) error {
	key, err := hex.DecodeString(id)
	if err != nil {
		return err
	}
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(b.name))
		if err != nil {
			return err
		}
		return bucket.Delete(key)
	})
}

// List returns an iterator that can be used to interate every key-value pair in
// the bucket
func (b *Bucket) List() (expay.Iter, error) {
	return newIter(b)
}

// Paginate is the same as List but supports pagination (not implmented yet)
func (b *Bucket) Paginate(lastCursor string, limit int) (expay.Iter, error) {
	return nil, errors.New("not implemented yet")
}

func newIter(b *Bucket) (*iter, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	bucket := tx.Bucket([]byte(b.name))
	if bucket == nil {
		tx.Rollback()
		return nil, expay.ErrNotFound
	}
	cursor := bucket.Cursor()
	key, value := cursor.First()
	return &iter{
		tx:     tx,
		cursor: cursor,
		key:    key,
		value:  value,
	}, nil
}

func (it *iter) Next() bool {
	return it.key != nil
}

func (it *iter) Scan(v interface{}) (id string, err error) {
	id = hex.EncodeToString(it.key)
	err = json.Unmarshal(it.value, v)
	it.key, it.value = it.cursor.Next()
	return
}

func (it *iter) Close() error {
	return it.tx.Rollback()
}

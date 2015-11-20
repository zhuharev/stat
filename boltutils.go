package stat

import (
	"encoding/binary"
	"fmt"
	"github.com/boltdb/bolt"
)

func (s *Store) HasBucket(bucket string) (has bool, e error) {
	e = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b != nil {
			has = true
			return nil
		}
		return nil
	})
	return
}

func (s *Store) CreateBucket(bucket string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(bucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

func (s *Store) Has(bucket string, key string) (has bool, e error) {
	e = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not exists", bucket)
		}
		v := b.Get([]byte(key))
		if len(v) > 0 {
			has = true
		}
		return nil
	})
	return
}

func (s *Store) Get(bucket string, key string) (id int64, e error) {
	e = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not exists", bucket)
		}
		v := b.Get([]byte(key))
		if len(v) > 0 {
			id = int64(btoi(v))
			return nil
		}
		return nil
	})
	return
}

func (s *Store) InsertAutoInc(bucket string, key string) (int64, error) {
	var res int64
	e := s.db.Update(func(tx *bolt.Tx) (e error) {
		// Retrieve the users bucket.
		// This should be created when the DB is first opened.
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			b, e = tx.CreateBucket([]byte(bucket))
			if e != nil {
				return fmt.Errorf("create bucket: %s", e)
			}
		}

		// Generate ID for the user.
		// This returns an error only if the Tx is closed or not writeable.
		// That can't happen in an Update() call so I ignore the error check.
		id, e := b.NextSequence()
		if e != nil {
			return e
		}
		res = int64(id)

		// Persist bytes to users bucket.
		return b.Put([]byte(key), itob(id))
	})
	if e != nil {
		return 0, e
	}
	return res, nil
}

func (s *Store) GetOrInsert(bucket string, key string) (id int64, e error) {
	ok, _ := s.Has(bucket, key)
	if ok {
		return s.Get(bucket, key)
	} else {
		return s.InsertAutoInc(bucket, key)
	}
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func itob32(v int64) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(v))
	return b
}

// itob returns an 8-byte big endian representation of v.
func btoi(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

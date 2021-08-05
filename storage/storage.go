package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

// Storage is a disk key value storage wrapper
type Storage struct {
	sync.Mutex
	path       string
	db         *bolt.DB
	timeout    time.Duration
	bucketName []byte
}

// NewStorage creates new cursor file
func NewStorage(path string, timeout time.Duration) (*Storage, error) {
	r := Storage{path: path}
	r.bucketName = []byte("logfiles")
	r.timeout = timeout

	var err error
	r.db, err = bolt.Open(path, 0600, &bolt.Options{Timeout: r.timeout * time.Second})

	if err != nil {
		return nil, err
	}

	err = r.db.Update(func(tx *bolt.Tx) error {
		_, cerr := tx.CreateBucket(r.bucketName)
		if cerr != nil {
			if cerr.Error() == "bucket already exists" {
				return nil
			}
			return fmt.Errorf("create bucket: %s", cerr)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &r, nil
}

// Keys returns all keys from storage or empty array with error
func (s *Storage) Keys() ([]string, error) {
	s.Lock()
	defer s.Unlock()

	var output []string
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			output = append(output, string(k))
		}
		return nil
	})
	return output, err
}

// Delete deletes row from storage by key
func (s *Storage) Delete(key string) error {
	s.Lock()
	defer s.Unlock()

	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		cerr := b.Delete([]byte(key))
		if cerr != nil {
			return cerr
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// Get returns value by key from storage
func (s *Storage) Get(key string) (string, error) {
	s.Lock()
	defer s.Unlock()
	var output string
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		v := b.Get([]byte(key))
		output = string(v)
		return nil
	})
	if err != nil {
		return "", err
	}
	return output, nil
}

// Set updates value by key in storage
func (s *Storage) Set(key string, value string) error {
	s.Lock()
	defer s.Unlock()
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		cerr := b.Put([]byte(key), []byte(value))
		if cerr != nil {
			return cerr
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Close closes storage
func (s *Storage) Close() error {
	s.Lock()
	defer s.Unlock()
	err := s.db.Close()
	if err != nil {
		return err
	}
	return nil
}

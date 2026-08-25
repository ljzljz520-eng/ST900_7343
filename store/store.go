package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	staffBucket    = []byte("staff_accounts")
	contactBucket  = []byte("contact_methods")
	auditBucket    = []byte("audit_entries")
	settingBucket  = []byte("studio_settings")
	metadataBucket = []byte("metadata")
)

var ErrNotFound = errors.New("record not found")
var ErrConflict = errors.New("record conflict")

type Store struct {
	db   *bolt.DB
	path string
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("database path is required")
	}
	parent := filepath.Dir(path)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: time.Second, NoGrowSync: false})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	result := &Store{db: db, path: path}
	if err := result.initialize(); err != nil {
		db.Close()
		return nil, err
	}
	return result, nil
}

func (s *Store) initialize() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, name := range [][]byte{staffBucket, contactBucket, auditBucket, settingBucket, metadataBucket} {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return fmt.Errorf("create bucket %s: %w", name, err)
			}
		}
		metadata := tx.Bucket(metadataBucket)
		if metadata.Get([]byte("schema_version")) == nil {
			if err := metadata.Put([]byte("schema_version"), []byte("1")); err != nil {
				return fmt.Errorf("write schema version: %w", err)
			}
		}
		return nil
	})
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) Health() error {
	if s == nil || s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.View(func(tx *bolt.Tx) error {
		if tx.Bucket(metadataBucket) == nil {
			return errors.New("metadata bucket is missing")
		}
		return nil
	})
}

func (s *Store) SchemaVersion() (string, error) {
	var version string
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(metadataBucket)
		if bucket == nil {
			return errors.New("metadata bucket is missing")
		}
		value := bucket.Get([]byte("schema_version"))
		if value == nil {
			return errors.New("schema version is missing")
		}
		version = string(value)
		return nil
	})
	return version, err
}

func encodeKey(value string) ([]byte, error) {
	if value == "" {
		return nil, errors.New("record key is required")
	}
	return []byte(value), nil
}

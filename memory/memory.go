package memory

import (
	"errors"
	"sync"
	"time"
)

// Storage interface that is implemented by storage providers
type Storage struct {
	mux        sync.RWMutex
	db         map[string]entry
	gcInterval time.Duration
}

// Common storage errors
var ErrNotExist = errors.New("key does not exist")

type entry struct {
	data   []byte
	expiry int64
}

// New creates a new memory storage
func New(config ...Config) *Storage {
	// Set default config
	cfg := configDefault(config...)

	// Create storage
	store := &Storage{
		db:         make(map[string]entry),
		gcInterval: cfg.GCInterval,
	}

	// Start garbage collector
	go store.gc()

	return store
}

// Get value by key
func (s *Storage) Get(key string) ([]byte, error) {
	s.mux.RLock()
	v, ok := s.db[key]
	s.mux.RUnlock()
	if !ok || v.expiry != 0 && v.expiry <= time.Now().Unix() {
		return nil, ErrNotExist
	}

	return v.data, nil
}

// Set key with value
func (s *Storage) Set(key string, val []byte, exp time.Duration) error {
	// Ain't Nobody Got Time For That
	if len(val) <= 0 {
		return nil
	}
	var expire int64
	if exp != 0 {
		expire = time.Now().Add(exp).Unix()
	}

	s.mux.Lock()
	s.db[key] = entry{val, expire}
	s.mux.Unlock()
	return nil
}

// Delete key by key
func (s *Storage) Delete(key string) error {
	s.mux.Lock()
	delete(s.db, key)
	s.mux.Unlock()
	return nil
}

// Reset all keys
func (s *Storage) Reset() error {
	s.mux.Lock()
	s.db = make(map[string]entry)
	s.mux.Unlock()
	return nil
}

// Close the database
func (s *Storage) Close() error {
	return nil
}

func (s *Storage) gc() {
	for t := range time.NewTicker(s.gcInterval).C {
		now := t.Unix()
		s.mux.Lock()
		for id, v := range s.db {
			if v.expiry != 0 && v.expiry < now {
				delete(s.db, id)
			}
		}
		s.mux.Unlock()
	}
}

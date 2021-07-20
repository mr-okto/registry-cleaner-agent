package status

import (
	"git.mills.io/prologic/bitcask"
	"sync"
)

type Storage struct {
	Path string
	mu   *sync.RWMutex
}

var (
	ErrKeyNotFound    = bitcask.ErrKeyNotFound
	KeyUnusedBlobs    = []byte("unused_blobs")
	KeyIndexedAt      = []byte("indexed_at")
	KeyCleanedAt      = []byte("cleaned_at")
	KeyBlobsTotalSize = []byte("blobs_total_size")
)

func NewStorage(storagePath string) *Storage {
	return &Storage{
		Path: storagePath,
		mu:   &sync.RWMutex{},
	}
}

// GetValue TODO: separate functions to open and close Storage
func (s *Storage) GetValue(key []byte, defaultValue []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	db, err := bitcask.Open(s.Path)
	if err != nil {
		return nil, err
	}
	defer func(db *bitcask.Bitcask) {
		_ = db.Close()
	}(db)

	val, err := db.Get(key)
	if err == bitcask.ErrKeyNotFound {
		if defaultValue == nil {
			return nil, ErrKeyNotFound
		}
		err = db.Put(key, defaultValue)
		return defaultValue, err
	}
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (s *Storage) SetValue(key []byte, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	db, err := bitcask.Open(s.Path)
	if err != nil {
		return err
	}
	defer func(db *bitcask.Bitcask) {
		_ = db.Close()
	}(db)

	err = db.Put(key, value)
	if err != nil {
		return err
	}
	return nil
}

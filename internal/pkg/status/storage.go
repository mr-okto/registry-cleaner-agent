package status

import (
	"errors"
	"git.mills.io/prologic/bitcask"
	"sync"
)

type Storage struct {
	Path string
	mu   *sync.RWMutex
	cask *bitcask.Bitcask
}

var (
	ErrKeyNotFound    = bitcask.ErrKeyNotFound
	ErrStorageClosed  = errors.New("storage is not open")
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

func (s *Storage) Open() error {
	db, err := bitcask.Open(s.Path)
	if err != nil {
		return err
	}
	s.cask = db
	return err
}

func (s *Storage) Close() error {
	return s.cask.Close()
}

func (s *Storage) GetValue(key []byte, defaultValue []byte) ([]byte, error) {
	if s.cask == nil {
		return nil, ErrStorageClosed
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, err := s.cask.Get(key)
	if err == bitcask.ErrKeyNotFound {
		if defaultValue == nil {
			return nil, ErrKeyNotFound
		}
		err = s.cask.Put(key, defaultValue)
		return defaultValue, err
	}
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (s *Storage) SetValue(key []byte, value []byte) error {
	if s.cask == nil {
		return ErrStorageClosed
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.cask.Put(key, value)
	if err != nil {
		return err
	}
	return nil
}

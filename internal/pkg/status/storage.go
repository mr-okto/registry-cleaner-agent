package status

import (
	"git.mills.io/prologic/bitcask"
)

type Storage struct {
	Path string
}

var (
	ErrKeyNotFound = bitcask.ErrKeyNotFound
	KeyUnusedBlobs = []byte("unused_blobs")
	KeyIndexedAt   = []byte("indexed_at")
	KeyCleanedAt   = []byte("cleaned_at")
)

// GetValue TODO: separate functions to open and close storage
func (s *Storage) GetValue(key []byte, defaultValue []byte) ([]byte, error) {
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

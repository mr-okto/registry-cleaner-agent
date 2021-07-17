package status

import (
	"strconv"
	"time"
)

//easyjson:json
type Status struct {
	IsAlive        bool   `json:"alive"`
	UnusedBlobs    int    `json:"unusedBlobs"`
	BlobsCleanedAt string `json:"blobsCleanedAt"`
	BlobsIndexedAt string `json:"blobsIndexedAt"`
}

func New() *Status {
	return &Status{
		IsAlive:        true,
		UnusedBlobs:    0,
		BlobsCleanedAt: time.RFC3339,
		BlobsIndexedAt: time.RFC3339,
	}
}

func (s *Status) Restore(store *Storage) error {
	val, err := store.GetValue(KeyUnusedBlobs, []byte(strconv.Itoa(s.UnusedBlobs)))
	if err != nil {
		return err
	}
	s.UnusedBlobs, err = strconv.Atoi(string(val))
	if err != nil {
		return err
	}
	val, err = store.GetValue(KeyIndexedAt, []byte(s.BlobsIndexedAt))
	if err != nil {
		return err
	}
	s.BlobsIndexedAt = string(val)
	val, err = store.GetValue(KeyCleanedAt, []byte(s.BlobsCleanedAt))
	if err != nil {
		return err
	}
	s.BlobsCleanedAt = string(val)
	return nil
}

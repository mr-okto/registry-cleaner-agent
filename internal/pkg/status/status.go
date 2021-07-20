package status

import (
	"time"
)

//easyjson:json
type Status struct {
	IsAlive        bool   `json:"isAlive"`
	UnusedBlobs    int    `json:"unusedBlobs"`
	BlobsCleanedAt string `json:"blobsCleanedAt"`
	BlobsIndexedAt string `json:"blobsIndexedAt"`
	BlobsTotalSize int    `json:"blobsTotalSize"`
}

func NewStatus() *Status {
	return &Status{
		IsAlive:        true,
		UnusedBlobs:    0,
		BlobsCleanedAt: time.RFC3339,
		BlobsIndexedAt: time.RFC3339,
		BlobsTotalSize: 0,
	}
}

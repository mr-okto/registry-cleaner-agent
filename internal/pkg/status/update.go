package status

import "time"

type Update struct {
	UnusedBlobs    *int
	BlobsCleanedAt *time.Time
	BlobsIndexedAt *time.Time
	BlobsTotalSize *int64
}

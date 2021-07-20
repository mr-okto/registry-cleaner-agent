package fs_analyzer

import (
	"os"
	"path"
	"strings"
)

type Analyzer struct {
	mntRoot string
}

const (
	BlobsPath    = "docker/registry/v2/blobs"
	BlobFilename = "data"
)

func NewFSAnalyzer(registryMntRoot string) *Analyzer {
	return &Analyzer{
		mntRoot: registryMntRoot,
	}
}

func (a *Analyzer) GetBlobSize(digest string) (int64, error) {
	sInd := strings.IndexRune(digest, ':')
	digestType := digest[:sInd] // sha:
	rawDigest := digest[sInd+1:]
	prefix := rawDigest[:2] // two leading digest symbols
	blobPath := path.Join(a.mntRoot, BlobsPath, digestType, prefix, rawDigest, BlobFilename)
	blob, err := os.Stat(blobPath)
	if err != nil {
		return 0, err
	}
	return blob.Size(), nil
}

// GetBlobsSize TODO: batch check (goroutines)
func (a *Analyzer) GetBlobsSize(digests []string) (sizes []int64, total int64, err error) {
	sizes = make([]int64, len(digests))
	total = 0
	for i, digest := range digests {
		size, err := a.GetBlobSize(digest)
		if err != nil {
			return nil, 0, err
		}
		sizes[i] = size
		total += size
	}
	return sizes, total, nil
}

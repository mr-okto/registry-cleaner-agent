package garbage_collector

import (
	"encoding/json"
	"net/http"
	"registry-cleaner-agent/internal/pkg/garbage"
)

func (gc *GarbageCollector) GarbageHandler(w http.ResponseWriter, _ *http.Request) {
	blobs, err := gc.ListGarbageBlobs() // TODO: update status index
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	garbageInfo := garbage.New()
	for _, blobDigest := range blobs {
		garbageInfo.Blobs = append(garbageInfo.Blobs, garbage.GarbageBlob{
			Size:   -1, //TODO: get size
			Digest: blobDigest,
		})
	}
	res, err := json.Marshal(&garbageInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(res)
}

func (gc *GarbageCollector) GarbageDeleteHandler(w http.ResponseWriter, _ *http.Request) {
	err := gc.RemoveGarbageBlobs() // TODO: update status
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

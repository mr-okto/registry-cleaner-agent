package garbage_collector

import (
	"encoding/json"
	"net/http"
	"registry-cleaner-agent/internal/pkg/agent_errors"
	"registry-cleaner-agent/internal/pkg/garbage"
	"registry-cleaner-agent/internal/pkg/status"
	"time"
)

type GCHandler struct {
	Gc            *GarbageCollector
	StatusManager *status.Manager
}

func InitGCHandler(gc *GarbageCollector, stm *status.Manager) (*GCHandler, error) {
	if gc == nil || stm == nil {
		return nil, agent_errors.NilPointerReference
	}
	return &GCHandler{
		Gc:            gc,
		StatusManager: stm,
	}, nil
}

func (gch *GCHandler) GarbageGetHandler(w http.ResponseWriter, _ *http.Request) {
	blobs, err := gch.Gc.ListGarbageBlobs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = gch.StatusManager.SetUnusedBlobs(len(blobs))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = gch.StatusManager.SetBlobsIndexedAt(time.Now())
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

func (gch *GCHandler) GarbageDeleteHandler(w http.ResponseWriter, _ *http.Request) {
	currentTime := time.Now()
	err := gch.Gc.RemoveGarbageBlobs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = gch.StatusManager.SetUnusedBlobs(0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = gch.StatusManager.SetBlobsIndexedAt(currentTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = gch.StatusManager.SetBlobsCleanedAt(currentTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

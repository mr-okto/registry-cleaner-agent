package garbage_collector

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"registry-cleaner-agent/internal/pkg/agent_errors"
	"registry-cleaner-agent/internal/pkg/fs_analyzer"
	"registry-cleaner-agent/internal/pkg/garbage"
	"registry-cleaner-agent/internal/pkg/status"
	"time"
)

type GCHandler struct {
	Gc            *GarbageCollector
	StatusManager *status.Manager
	FSAnalyzer    *fs_analyzer.Analyzer
}

func InitGCHandler(
	gc *GarbageCollector, stm *status.Manager,
	fsa *fs_analyzer.Analyzer) (*GCHandler, error) {
	if gc == nil || stm == nil || fsa == nil {
		return nil, agent_errors.NilPointerReference
	}
	return &GCHandler{
		Gc:            gc,
		StatusManager: stm,
		FSAnalyzer:    fsa,
	}, nil
}

func (gch *GCHandler) Cleanup(ctx context.Context) {
	err := gch.StatusManager.Shutdown()
	if err != nil {
		log.Printf("[GCHandler] StatusManager.Shutdown error: %v", err)
	}
	err = gch.Gc.Shutdown(ctx)
	if err != nil {
		log.Printf("[GCHandler] GC.Shutdown error: %v", err)
	}
}

func (gch *GCHandler) GarbageGetHandler(w http.ResponseWriter, _ *http.Request) {
	blobs, err := gch.Gc.ListGarbageBlobs()
	if err != nil && err == ErrAlreadyRunning {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	blobSizes, totalSize, err := gch.FSAnalyzer.GetBlobsSize(blobs)
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
	err = gch.StatusManager.SetBlobsTotalSize(totalSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	garbageInfo := garbage.New()
	for i, blobDigest := range blobs {
		garbageInfo.Blobs = append(garbageInfo.Blobs, garbage.GarbageBlob{
			Size:   blobSizes[i],
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
	if err != nil && err == ErrAlreadyRunning {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = gch.StatusManager.SetUnusedBlobs(0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = gch.StatusManager.SetBlobsTotalSize(0)
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

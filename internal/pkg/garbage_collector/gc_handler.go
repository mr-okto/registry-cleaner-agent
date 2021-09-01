package garbage_collector

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/robfig/cron"
	"log"
	"net/http"
	"registry-cleaner-agent/internal/pkg/agent_errors"
	"registry-cleaner-agent/internal/pkg/fs_analyzer"
	"registry-cleaner-agent/internal/pkg/garbage"
	"registry-cleaner-agent/internal/pkg/status"
	"sync"
	"time"
)

type GCHandler struct {
	Gc            *GarbageCollector
	StatusManager *status.Manager
	FSAnalyzer    *fs_analyzer.Analyzer
	cron          *cron.Cron
	mu            *sync.RWMutex
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
		mu:            &sync.RWMutex{},
		cron:          cron.New(),
	}, nil
}

func (gch *GCHandler) EnableCron(indexSpec string, removalSpec string) error {
	err := gch.cron.AddFunc(indexSpec, gch.IndexGarbage)
	if err != nil {
		return err
	}
	err = gch.cron.AddFunc(removalSpec, gch.RemoveGarbage)
	if err != nil {
		return err
	}
	gch.cron.Start()
	return nil
}

func (gch *GCHandler) DisableCron() {
	gch.cron.Stop()
	gch.cron = cron.New() // Removes entries
}

func (gch *GCHandler) IndexGarbage() {
	gch.mu.RLock()
	defer gch.mu.RUnlock()
	currentTime := time.Now()
	blobs, err := gch.Gc.ListGarbageBlobs()
	if err != nil {
		return
	}
	_, totalSize, err := gch.FSAnalyzer.GetBlobsSize(blobs)
	if err != nil {
		return
	}
	unusedBlobs := len(blobs)
	statusUpdate := status.Update{
		UnusedBlobs:    &unusedBlobs,
		BlobsTotalSize: &totalSize,
		BlobsIndexedAt: &currentTime,
	}
	_ = gch.StatusManager.UpdateStatus(&statusUpdate)
}

func (gch *GCHandler) RemoveGarbage() {
	gch.mu.RLock()
	defer gch.mu.RUnlock()
	currentTime := time.Now()
	err := gch.Gc.RemoveGarbageBlobs()
	if err != nil {
		return
	}
	unusedBlobs := 0
	totalSize := int64(0)
	statusUpdate := status.Update{
		UnusedBlobs:    &unusedBlobs,
		BlobsTotalSize: &totalSize,
		BlobsIndexedAt: &currentTime,
		BlobsCleanedAt: &currentTime,
	}
	_ = gch.StatusManager.UpdateStatus(&statusUpdate)
}

func (gch *GCHandler) Cleanup(ctx context.Context) {
	gch.mu.Lock()
	gch.DisableCron()
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
	gch.mu.RLock()
	defer gch.mu.RUnlock()
	currentTime := time.Now()
	blobs, err := gch.Gc.TryListGarbageBlobs()
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
	unusedBlobs := len(blobs)
	statusUpdate := status.Update{
		UnusedBlobs:    &unusedBlobs,
		BlobsTotalSize: &totalSize,
		BlobsIndexedAt: &currentTime,
	}
	err = gch.StatusManager.UpdateStatus(&statusUpdate)
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
	gch.mu.Lock()
	defer gch.mu.Unlock()
	currentTime := time.Now()
	err := gch.Gc.TryRemoveGarbageBlobs()
	if errors.Is(err, ErrAlreadyRunning) {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	unusedBlobs := 0
	blobsTotalSize := int64(unusedBlobs)
	statusUpdate := status.Update{
		UnusedBlobs:    &unusedBlobs,
		BlobsTotalSize: &blobsTotalSize,
		BlobsIndexedAt: &currentTime,
		BlobsCleanedAt: &currentTime,
	}
	err = gch.StatusManager.UpdateStatus(&statusUpdate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

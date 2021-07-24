package garbage_collector

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/semaphore"
	"log"
	"os/exec"
	"strings"
)

const (
	RegistryBin         = "/bin/registry"
	GcCommand           = "garbage-collect"
	DeleteUntagged      = "--delete-untagged"
	DryRun              = "--dry-run"
	EligibleForDeletion = "blob eligible for deletion: "
	StatSuffix          = "manifests eligible for deletion"
	TimePrefix          = "time="
	LogPrefix           = "level="
)

type GarbageCollector struct {
	ContainerName      string
	RegistryConfigPath string
	sem                *semaphore.Weighted
}

var (
	ErrAlreadyRunning = errors.New("garbage collector already running")
)

func NewGarbageCollector(containerName string, registryConfigPath string) *GarbageCollector {
	return &GarbageCollector{
		ContainerName:      containerName,
		RegistryConfigPath: registryConfigPath,
		sem:                semaphore.NewWeighted(int64(1)),
	}
}

func (gc *GarbageCollector) TryListGarbageBlobs() ([]string, error) {
	if !gc.sem.TryAcquire(1) {
		return nil, ErrAlreadyRunning
	}
	return gc.listGarbageBlobs()
}

func (gc *GarbageCollector) ListGarbageBlobs() ([]string, error) {
	err := gc.sem.Acquire(context.Background(), 1)
	if err != nil {
		return nil, err
	}
	return gc.listGarbageBlobs()
}

func (gc *GarbageCollector) listGarbageBlobs() ([]string, error) {
	defer gc.sem.Release(1)
	cmd := exec.Command("docker", "exec", gc.ContainerName,
		RegistryBin, GcCommand, DeleteUntagged, DryRun, gc.RegistryConfigPath)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logErr := fmt.Errorf("docker garbage-collect failed: %v; srderr: %s", err, stderr.String())
		log.Printf("[ERROR at GarbageCollector.listGarbageBlobs]: %v", logErr)
		return nil, logErr
	}
	sc := bufio.NewScanner(bytes.NewReader(out.Bytes()))
	var blobs []string
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, EligibleForDeletion) {
			blobs = append(blobs, strings.TrimPrefix(line, EligibleForDeletion))
		} else if strings.HasSuffix(line, StatSuffix) {
			log.Printf("[INFO at GarbageCollector.listGarbageBlobs] garbage collector dry run results: %s\n", line)
		}
	}
	return blobs, nil
}

func (gc *GarbageCollector) TryRemoveGarbageBlobs() error {
	if !gc.sem.TryAcquire(1) {
		return ErrAlreadyRunning
	}
	return gc.removeGarbageBlobs()
}

func (gc *GarbageCollector) RemoveGarbageBlobs() error {
	err := gc.sem.Acquire(context.Background(), 1)
	if err != nil {
		return err
	}
	return gc.removeGarbageBlobs()
}

func (gc *GarbageCollector) removeGarbageBlobs() error {
	cmd := exec.Command("docker", "exec", gc.ContainerName,
		RegistryBin, GcCommand, DeleteUntagged, gc.RegistryConfigPath)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logErr := fmt.Errorf("docker garbage-collect failed: %v; srderr: %s", err, stderr.String())
		log.Printf("[ERROR at GarbageCollector.TryRemoveGarbageBlobs]: %v", logErr)
		return logErr
	}
	go func() {
		cmd = exec.Command("docker", "restart", gc.ContainerName)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			log.Printf("[ERROR at GarbageCollector.TryRemoveGarbageBlobs]: docker restart failed: %v; srderr: %s",
				err, stderr.String())
		}
		gc.sem.Release(1)
	}()
	sc := bufio.NewScanner(bytes.NewReader(out.Bytes()))
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, TimePrefix) {
			log.Println(line[strings.Index(line, LogPrefix):])
		} else if strings.HasSuffix(line, StatSuffix) {
			log.Printf("[INFO at GarbageCollector.TryRemoveGarbageBlobs] garbage collector run results %s\n", line)
		}
	}
	return nil
}

func (gc *GarbageCollector) Shutdown(ctx context.Context) error {
	return gc.sem.Acquire(ctx, 1)
}

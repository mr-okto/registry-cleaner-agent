package garbage_collector

import (
	"bufio"
	"bytes"
	"fmt"
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
}

func NewGarbageCollector(containerName string, registryConfigPath string) *GarbageCollector {
	return &GarbageCollector{
		ContainerName:      containerName,
		RegistryConfigPath: registryConfigPath,
	}
}

func (gc *GarbageCollector) ListGarbageBlobs() ([]string, error) {
	out, err := exec.Command("docker", "exec", gc.ContainerName,
		RegistryBin, GcCommand, DeleteUntagged, DryRun, gc.RegistryConfigPath).Output()
	if err != nil {
		return nil, err
	}
	sc := bufio.NewScanner(bytes.NewReader(out))
	var blobs []string
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, EligibleForDeletion) {
			blobs = append(blobs, strings.TrimPrefix(line, EligibleForDeletion))
		} else if strings.HasSuffix(line, StatSuffix) {
			log.Printf("Garbage collector dry run: %s\n", line)
		}
	}
	return blobs, nil
}

func (gc *GarbageCollector) RemoveGarbageBlobs() error {
	out, err := exec.Command("docker", "exec", gc.ContainerName,
		RegistryBin, GcCommand, DeleteUntagged, gc.RegistryConfigPath).Output()
	if err != nil {
		return fmt.Errorf("docker exec returned non-zero exit code: %s", err.Error())
	}
	err = exec.Command("docker", "restart", gc.ContainerName).Run()
	if err != nil {
		return fmt.Errorf("docker restart returned non-zero exit code: %s", err.Error())
	}
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, TimePrefix) {
			log.Println(line[strings.Index(line, LogPrefix):])
		} else if strings.HasSuffix(line, StatSuffix) {
			log.Printf("Garbage collector run: %s\n", line)
		}
	}
	return nil
}

package garbage_collector

import (
	"bufio"
	"bytes"
	"errors"
	"log"
	"os/exec"

	"strings"
)

const (
	REGISTRY_BIN          = "/bin/registry"
	GC_COMMAND            = "garbage-collect"
	DELETE_UNTAGGED       = "--delete-untagged"
	DRY_RUN               = "--dry-run"
	ELIGIBLE_FOR_DELETION = "blob eligible for deletion: "
	STAT_SUFFIX           = "manifests eligible for deletion"
	TIME_PREFIX           = "time="
	LOG_PREFIX            = "level="
)

var (
	DockerExecFailure = errors.New("docker exec returned non-zero exit code")
)

type GarbageCollector struct {
	ContainerName      string
	RegistryConfigPath string
}

func (gc *GarbageCollector) ListGarbageBlobs() ([]string, error) {
	out, err := exec.Command("docker", "exec", gc.ContainerName,
		REGISTRY_BIN, GC_COMMAND, DELETE_UNTAGGED, DRY_RUN, gc.RegistryConfigPath).Output()
	if err != nil {
		return nil, err
	}
	sc := bufio.NewScanner(bytes.NewReader(out))
	var blobs []string
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, ELIGIBLE_FOR_DELETION) {
			blobs = append(blobs, strings.TrimPrefix(line, ELIGIBLE_FOR_DELETION))
		} else if strings.HasSuffix(line, STAT_SUFFIX) {
			log.Printf("Garbage collector dry run: %s\n", line)
		}
	}
	return blobs, nil
}

// RemoveGarbageBlobs TODO: Pause registry
func (gc *GarbageCollector) RemoveGarbageBlobs() error {
	out, err := exec.Command("docker", "exec", gc.ContainerName,
		REGISTRY_BIN, GC_COMMAND, DELETE_UNTAGGED, gc.RegistryConfigPath).Output()
	if err != nil {
		return err
	}
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, TIME_PREFIX) {
			log.Println(line[strings.Index(line, LOG_PREFIX):])
		} else if strings.HasSuffix(line, STAT_SUFFIX) {
			log.Printf("Garbage collector run: %s\n", line)
		}
	}
	return nil
}

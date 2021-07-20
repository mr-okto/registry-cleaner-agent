package status

import (
	"strconv"
	"time"
)

type Manager struct {
	Storage *Storage
	Status  *Status
}

func InitStatusManager(storagePath string) (*Manager, error) {
	storage := NewStorage(storagePath)
	status := NewStatus()
	m := &Manager{
		Storage: storage,
		Status:  status,
	}
	err := m.restoreStatus()
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Manager) restoreStatus() error {
	val, err := m.Storage.GetValue(KeyUnusedBlobs, []byte(strconv.Itoa(m.Status.UnusedBlobs)))
	if err != nil {
		return err
	}
	m.Status.UnusedBlobs, err = strconv.Atoi(string(val))
	if err != nil {
		return err
	}
	val, err = m.Storage.GetValue(KeyIndexedAt, []byte(m.Status.BlobsIndexedAt))
	if err != nil {
		return err
	}
	m.Status.BlobsIndexedAt = string(val)
	val, err = m.Storage.GetValue(KeyCleanedAt, []byte(m.Status.BlobsCleanedAt))
	if err != nil {
		return err
	}
	m.Status.BlobsCleanedAt = string(val)
	val, err = m.Storage.GetValue(KeyBlobsTotalSize, []byte(strconv.FormatInt(m.Status.BlobsTotalSize, 10)))
	if err != nil {
		return err
	}
	m.Status.BlobsTotalSize, err = strconv.ParseInt(string(val), 10, 64)
	return err
}

// SetIsAlive IsAlive status is not stored persistently as it is useless
func (m *Manager) SetIsAlive(isAlive bool) {
	m.Status.IsAlive = isAlive
}

func (m *Manager) SetUnusedBlobs(unusedBlobs int) error {
	err := m.Storage.SetValue(KeyUnusedBlobs,
		[]byte(strconv.Itoa(unusedBlobs)))
	if err != nil {
		return err
	}
	m.Status.UnusedBlobs = unusedBlobs
	return nil
}

func (m *Manager) SetBlobsCleanedAt(blobsCleanedAt time.Time) error {
	timeStr := blobsCleanedAt.Format(time.RFC3339)
	err := m.Storage.SetValue(KeyCleanedAt, []byte(timeStr))
	if err != nil {
		return err
	}
	m.Status.BlobsCleanedAt = timeStr
	return nil
}

func (m *Manager) SetBlobsIndexedAt(blobsIndexedAt time.Time) error {
	timeStr := blobsIndexedAt.Format(time.RFC3339)
	err := m.Storage.SetValue(KeyIndexedAt, []byte(timeStr))
	if err != nil {
		return err
	}
	m.Status.BlobsIndexedAt = timeStr
	return nil
}

func (m *Manager) SetBlobsTotalSize(blobsTotalSize int64) error {
	err := m.Storage.SetValue(KeyIndexedAt,
		[]byte(strconv.FormatInt(blobsTotalSize, 10)))
	if err != nil {
		return err
	}
	m.Status.BlobsTotalSize = blobsTotalSize
	return nil
}

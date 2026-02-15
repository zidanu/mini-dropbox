package metadata

import (
	"github.com/google/uuid"
	"time"
)

type File struct {
	ID           uuid.UUID
	Path         string
	Hash         string
	Size         int64
	ModTime      time.Time
	IsDir        bool
	Version      int
	RemoteHash   string
	LastSyncTime time.Time
}

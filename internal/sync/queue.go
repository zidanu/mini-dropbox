package sync

import (
	"github.com/fsnotify/fsnotify"
)

type SyncState int

const (
	Synced SyncState = iota
	Pending
	Uploading
	Deleted
)

type Direction int

const (
	Upload Direction = iota
	Download
)

type SyncOp struct {
	EventType fsnotify.Op
	Status    SyncState
	Direction Direction
	Retries   int
	Error     error
}

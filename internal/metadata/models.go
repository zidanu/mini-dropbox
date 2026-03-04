package metadata

import (
	"time"
)

type File struct {
	ID           string
	Path         string
	Hash         string
	Size         int64
	ModTime      time.Time
	IsDir        bool
	Version      int
	RemoteHash   string
	LastSyncTime time.Time
	CreatedAt    time.Time
	Deleted      bool
	Renamed      bool
	Inode        uint64
}

func FileConstructor(path string, hash string, size int64, modTime time.Time, isDir bool, ino uint64) *File {
	return &File{
		Path:    path,
		Hash:    hash,
		Size:    size,
		ModTime: modTime,
		IsDir:   isDir,
		Inode:   ino,
	}
}

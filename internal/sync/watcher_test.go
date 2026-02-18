package sync

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestWatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	queue := make(chan *SyncOp, 10)
	errCh := make(chan error, 7)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := Watch(ctx, tmpDir, queue, errCh); err != nil {
			t.Errorf("Watch failed: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	t.Run("Detects file creation/write/rename/removal", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "hello.txt")
		newFile := filepath.Join(tmpDir, "goodbye.txt")

		if err := os.WriteFile(testFile, []byte("world"), 0644); err != nil {
			t.Fatal(err)
		}

		time.Sleep(100 * time.Millisecond)
		if err := os.Rename(testFile, newFile); err != nil {
			t.Fatal(err)
		}

		time.Sleep(100 * time.Millisecond)
		if err := os.Remove(newFile); err != nil {
			t.Fatal(err)
		}

		events := make(map[fsnotify.Op]bool)
		timeout := time.After(3 * time.Second)
		for len(events) < 4 {
			select {
			case op := <-queue:
				events[op.EventType] = true
				t.Logf("Received event: %v on %s", op.EventType, op.File.Path)
			case <-timeout:
				t.Fatalf("Timed out. Events seen: %v", events)
			}
		}
	})
}

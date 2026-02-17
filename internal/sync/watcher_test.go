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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := Watch(ctx, tmpDir, queue); err != nil {
			t.Errorf("Watch failed: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	t.Run("Detects file creation", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "hello.txt")
		if err := os.WriteFile(testFile, []byte("world"), 0644); err != nil {
			t.Fatal(err)
		}

		select {
		case op := <-queue:
			if !op.EventType.Has(fsnotify.Create) {
				t.Errorf("Expected Create event, got %v", op.EventType)
			}
			if op.File.Path != testFile {
				t.Errorf("Expected path %s, got %s", testFile, op.File.Path)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Timed out waiting for Create event")
		}
	})
}

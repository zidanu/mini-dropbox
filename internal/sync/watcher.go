package sync

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"os"
	"time"
)

func watch(ctx context.Context, filepath string, queue chan<- *SyncOp) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to initialize file watching (func watch): %x", err)
	}

	errs := make(chan error, 4)

	go func() {
		debounceMap := make(map[string]int64)
		for {
			select {
			case <-ctx.Done():
				watcher.Close()
				return
			case event, ok := <-watcher.Events:
				if !ok {
					errs <- fmt.Errorf("error occurred with watcher.Events (func watch): %w", err)
				}

				path := event.Name

				currentTime := time.Now().UnixMilli()

				if _, exists := debounceMap[path]; !exists {
					debounceMap[path] = currentTime
				} else if (currentTime - debounceMap[path]) < 6000 {
					continue
				}

				if event.Has(fsnotify.Create) {
					pathInfo, err := os.Stat(path)
					if err != nil {
						if os.IsNotExist(err) {
							errs <- fmt.Errorf("%s file does not exist", path)
						}
						errs <- fmt.Errorf("error occurred with os.Stat (func watch): %w", err)
					}
					if pathInfo.IsDir() {
						watcher.Add(path)
					}
				}
				if event.Has(fsnotify.Write) {

				}
				if event.Has(fsnotify.Remove) {
				}
				if event.Has(fsnotify.Rename) {
				}

			case err, ok := <-watcher.Errors:
				if ok {
					errs <- fmt.Errorf("error occurred with watcher.Errors (func watch): %w", err)
					return
				}
			}
		}
	}()

	if addErr := watcher.Add(filepath); addErr != nil {
		return fmt.Errorf("error occurred with watcher.Add (func watch): %w", addErr)
	}

	finErr, ok := <-errs
	if !ok {
		return nil
	}
	return finErr
}

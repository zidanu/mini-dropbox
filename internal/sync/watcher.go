package sync

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
)

func watch(filepath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to initialize file watching (func watch): %x", err)
	}
	defer watcher.Close()

	go func() error {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return fmt.Errorf("error occured with watcher.Events (func watch): %w", err)
				}
				if event.Has(fsnotify.Create) {
				}
				if event.Has(fsnotify.Write) {

				}
				if event.Has(fsnotify.Remove) {
				}
				if event.Has(fsnotify.Rename) {
				}
				if event.Has(fsnotify.Chmod) {
				}

			case err, ok := <-watcher.Errors:
				if ok {
					return fmt.Errorf("error occurred with watcher.Errors (func watch): %w", err)
				}
				return nil
			}
		}
	}()

	watcher.Add(filepath)

	return nil
}

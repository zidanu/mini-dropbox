package sync

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/zidanu/mini-dropbox/internal/hash"
	"github.com/zidanu/mini-dropbox/internal/metadata"
	"io/fs"
	"os"
	fp "path/filepath"
	"strings"
)

func Watch(ctx context.Context, filepath string, queue chan<- *SyncOp) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to initialize file watching (func watch): %w", err)
	}

	errs := make(chan error, 5)

	go func() {
		defer watcher.Close()
		defer close(errs)
		watchedDirs := make(map[string]bool)
		fileMap := make(map[string]*metadata.File)

		if addErr := watcher.Add(filepath); addErr != nil {
			errs <- fmt.Errorf("error occurred with watcher.Add (func watch): %w", addErr)
			return
		}
		watchedDirs[filepath] = true
		if addSubErr := watchSubdirectory(filepath, watcher, &watchedDirs); addSubErr != nil {
			errs <- addSubErr
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					errs <- fmt.Errorf("error occurred with watcher.Events (func watch): %w", err)
					continue
				}

				path := event.Name

				var pathInfo os.FileInfo
				if !(event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename)) {
					pathInfo, err = os.Stat(path)
					if err != nil {
						if os.IsNotExist(err) {
							errs <- fmt.Errorf("%s file does not exist", path)
							continue
						}
						errs <- fmt.Errorf("error occurred with os.Stat (func watch): %w", err)
						continue
					}
					hashCode, err := hash.ComputeFileHash(path)
					if err != nil {
						errs <- fmt.Errorf("error occurred with hash.ComputeFileHash (func watch): %w", err)
						continue
					}
					if _, exists := fileMap[path]; !exists {
						fileMap[path] = metadata.FileConstructor(path, hashCode, pathInfo.Size(), pathInfo.ModTime(), pathInfo.IsDir())
					} else {
						fileMap[path].ModTime = pathInfo.ModTime()
						fileMap[path].Hash = hashCode
						fileMap[path].Size = pathInfo.Size()
					}
				}

				var syncOp SyncOp
				if event.Has(fsnotify.Create) {
					_, exists := fileMap[path]
					if !exists {
						continue
					}
					if fileMap[path].IsDir {
						if addErr := watcher.Add(path); addErr != nil {
							errs <- fmt.Errorf("error occurred with watcher.Add (func watch goroutine): %w", addErr)
							continue
						}
						watchedDirs[path] = true
						if addSubErr := watchSubdirectory(path, watcher, &watchedDirs); addSubErr != nil {
							errs <- addSubErr
						}
					}
					fileMap[path].CreatedAt = fileMap[path].ModTime
					syncOp = SyncOpConstructor(fileMap[path], event.Op)
				} else if event.Has(fsnotify.Write) {
					_, exists := fileMap[path]
					if !exists {
						continue
					}
					syncOp = SyncOpConstructor(fileMap[path], event.Op)
				} else if event.Has(fsnotify.Remove) {
					_, exists := fileMap[path]
					if !exists {
						// Come back to this when initial sync is implemented
						continue
					}
					if fileMap[path].IsDir {
						if removeErr := watcher.Remove(path); removeErr != nil {
							errs <- fmt.Errorf("error occurred with watcher.Remove (func watch goroutine): %w", removeErr)
						}
						for dir := range watchedDirs {
							if strings.HasPrefix(dir, path+string(os.PathSeparator)) {
								watcher.Remove(dir)
								delete(watchedDirs, dir)
							}
						}
					}
					fileMap[path].Deleted = true
					syncOp = SyncOpConstructor(fileMap[path], event.Op)
				} else if event.Has(fsnotify.Rename) {
					continue
				}
				queue <- &syncOp

			case err, ok := <-watcher.Errors:
				if ok {
					errs <- fmt.Errorf("error occurred with watcher.Errors (func watch): %w", err)
					return
				}
			}
		}
	}()

	finErr, ok := <-errs
	if !ok {
		return nil
	}
	return finErr
}

func watchSubdirectory(root string, watcher *fsnotify.Watcher, watchedDirs *map[string]bool) error {
	return fp.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error occurred with fp.WalkDir (func watchSubdirectory): %w", err)
		}
		if d.IsDir() && path != root {
			err := watcher.Add(path)
			if err != nil {
				return fmt.Errorf("error occurred with watcher.Add (func watchSubdirectory): %w", err)
			}
			(*watchedDirs)[path] = true
		}
		return nil
	})
}

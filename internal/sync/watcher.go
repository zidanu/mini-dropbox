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
	"syscall"
)

func Watch(ctx context.Context, filepath string, database *metadata.Database, queue chan<- *SyncOp, errs chan<- error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to initialize file watching (func watch): %w", err)
	}

	if addErr := watcher.Add(filepath); addErr != nil {
		watcher.Close()
		return addErr
	}

	go func() {
		defer watcher.Close()
		watchedDirs := make(map[string]bool)
		// fileMap := make(map[string]*metadata.File)

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
				file, getFileErr := database.GetFileByPath(path)

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
					hashCode := ""
					if getFileErr != nil {
						err = nil
						if !pathInfo.IsDir() {
							hashCode, err = hash.ComputeFileHash(path)
						}
						if err != nil {
							errs <- fmt.Errorf("error occurred with hash.ComputeFileHash (func watch): %w", err)
							continue
						}
						inode := pathInfo.Sys().(*syscall.Stat_t).Ino
						newFile := metadata.FileConstructor(path, hashCode, pathInfo.Size(), pathInfo.ModTime(), pathInfo.IsDir(), inode)
						if saveErr := database.SaveFile(newFile); saveErr != nil {
							errs <- fmt.Errorf("failed to save: %w", saveErr)
							continue
						}
						file = newFile
						getFileErr = nil
					} else {
						file.ModTime = pathInfo.ModTime()
						file.Hash = hashCode
						file.Size = pathInfo.Size()
					}
				}

				var syncOp SyncOp
				if event.Has(fsnotify.Create) {
					if file.IsDir {
						if addErr := watcher.Add(path); addErr != nil {
							errs <- fmt.Errorf("error occurred with watcher.Add (func watch goroutine): %w", addErr)
							continue
						}
						watchedDirs[path] = true
						if addSubErr := watchSubdirectory(path, watcher, &watchedDirs); addSubErr != nil {
							errs <- addSubErr
						}
					}
					file.CreatedAt = file.ModTime
					syncOp = SyncOpConstructor(file, event.Op)
				} else if event.Has(fsnotify.Write) {
					syncOp = SyncOpConstructor(file, event.Op)
				} else if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					if getFileErr != nil {
						// Come back to this when initial sync is implemented
						continue
					}
					if file.IsDir {
						if removeErr := watcher.Remove(path); removeErr != nil {
							errs <- fmt.Errorf("error occurred with watcher.Remove (func watch goroutine): %w", removeErr)
						}
						if event.Has(fsnotify.Remove) {
							for dir := range watchedDirs {
								if strings.HasPrefix(dir, path+string(os.PathSeparator)) {
									watcher.Remove(dir)
									delete(watchedDirs, dir)
								}
							}
						}
					}
					if event.Has(fsnotify.Remove) {
						file.Deleted = true
					}
					syncOp = SyncOpConstructor(file, event.Op)
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

	return nil
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

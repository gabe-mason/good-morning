package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gabe-mason/good-morning/version"
)

type Watcher struct {
	watcher    *fsnotify.Watcher
	versionDir string
}

func NewWatcher(versionDir string) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %v", err)
	}

	return &Watcher{
		watcher:    watcher,
		versionDir: versionDir,
	}, nil
}

func (w *Watcher) WatchFile(filePath string) error {
	// Watch the directory containing the file
	dir := filepath.Dir(filePath)
	if err := w.watcher.Add(dir); err != nil {
		return fmt.Errorf("failed to watch directory: %v", err)
	}

	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if event.Name == filePath && (event.Op&fsnotify.Write == fsnotify.Write) {
					// File was modified
					content, err := os.ReadFile(filePath)
					if err != nil {
						fmt.Printf("Error reading file: %v\n", err)
						continue
					}

					// Load version history
					versionFile := filepath.Join(w.versionDir, time.Now().Format("2006-01-02")+".json")
					versionHistory, err := version.LoadFromFile(versionFile)
					if err != nil {
						fmt.Printf("Error loading version history: %v\n", err)
						continue
					}

					// Add user change
					versionHistory.AddChange(string(content), version.ChangeTypeUser, "user")
					if err := versionHistory.SaveToFile(versionFile); err != nil {
						fmt.Printf("Error saving version history: %v\n", err)
					}
				}
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("Watcher error: %v\n", err)
			}
		}
	}()

	return nil
}

func (w *Watcher) Close() error {
	return w.watcher.Close()
}

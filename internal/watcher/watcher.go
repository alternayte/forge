package watcher

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher wraps fsnotify with debouncing and file filtering.
type Watcher struct {
	fsw      *fsnotify.Watcher
	onChange func()
	timer    *time.Timer
	debounce time.Duration
	done     chan struct{}
}

// New creates a new file watcher with the given onChange callback and debounce duration.
func New(onChange func(), debounce time.Duration) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		fsw:      fsw,
		onChange: onChange,
		debounce: debounce,
		done:     make(chan struct{}),
	}

	// Start watch loop in background
	go w.watch()

	return w, nil
}

// Add adds a path to the watch list.
func (w *Watcher) Add(path string) error {
	return w.fsw.Add(path)
}

// Close closes the watcher and stops the watch loop.
func (w *Watcher) Close() error {
	close(w.done)
	return w.fsw.Close()
}

// watch runs the event loop that processes file system events.
func (w *Watcher) watch() {
	for {
		select {
		case <-w.done:
			return

		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}

			// Skip Chmod events unconditionally (macOS Spotlight, antivirus, editors trigger these constantly)
			if event.Has(fsnotify.Chmod) {
				continue
			}

			// Filter by file extension and path
			if !isRelevantFile(event.Name) {
				continue
			}

			// Debounce: reset timer on each relevant event
			if w.timer != nil {
				w.timer.Stop()
			}
			w.timer = time.AfterFunc(w.debounce, w.onChange)

		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		}
	}
}

// isRelevantFile checks if a file should trigger regeneration.
func isRelevantFile(path string) bool {
	// Ignore generated files (avoid infinite loops)
	if strings.Contains(path, "gen/") {
		return false
	}

	// Check file extension
	ext := filepath.Ext(path)
	switch ext {
	case ".go", ".templ", ".sql", ".css":
		return true
	default:
		return false
	}
}

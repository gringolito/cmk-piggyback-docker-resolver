package watcher

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gringolito/cmk-piggyback-docker-resolver/pkg/logx"
)

// Event represents a new directory event detected by the watcher.
// Path is the full path to the directory and Name is the base name.
type Event struct {
	Path string // full path to the new directory
	Name string // base name
}

// Watcher watches a directory for newly created container directories
// and emits Event values for interested consumers.
type Watcher struct {
	w    *fsnotify.Watcher
	out  chan Event
	root string
	log  *logx.Logger
}

// New creates and starts a Watcher for the provided root directory.
// The returned Watcher is already running its internal event loop.
func New(root string, l *logx.Logger) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = w.Add(root)
	if err != nil {
		return nil, err
	}
	watcher := &Watcher{w: w, out: make(chan Event, 32), root: root, log: l}
	go watcher.start()
	return watcher, nil
}

// Events returns a receive-only channel supplying new Event values.
// The channel is buffered and is closed when the underlying watcher is closed.
func (w *Watcher) Events() <-chan Event {
	return w.out
}

// Close stops the watcher and releases associated resources.
// It returns any error from the underlying fsnotify watcher Close call.
func (w *Watcher) Close() error {
	return w.w.Close()
}

func (w *Watcher) start() {
	w.scan()
	w.loop()
}

func (w *Watcher) scan() {
	entries, err := os.ReadDir(w.root)
	if err != nil {
		w.log.Warn("watcher_scan", err, "path", w.root)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		w.create(filepath.Join(w.root, e.Name()), true)
	}
}

func (w *Watcher) loop() {
	for {
		select {
		case ev, ok := <-w.w.Events:
			if !ok {
				// Watcher closed?
				return
			}

			if w.filterEvents(ev) {
				continue
			}

			if ev.Has(fsnotify.Create) {
				w.createEvent(ev)
			}
		case err, ok := <-w.w.Errors:
			if !ok {
				// Watcher closed?
				return
			}

			w.log.Warn("watcher_error", err)
		}
	}
}

func (w *Watcher) filterEvents(ev fsnotify.Event) bool {
	filename := filepath.Base(ev.Name)

	// Do not process hidden filenames
	if strings.HasPrefix(filename, ".") {
		return true
	}

	// Do not process symlinks
	if isSymlink(ev.Name) {
		return true
	}

	return false
}

func (w *Watcher) create(path string, addToWatcher bool) {
	base := filepath.Base(path)
	if !maybeContainerID(base) {
		return
	}

	if addToWatcher {
		err := w.w.Add(path)
		if err != nil {
			w.log.Warn("watcher_new_dir", err, "path", path)
		}
	}

	w.out <- Event{Path: path, Name: base}
}

func (w *Watcher) createEvent(ev fsnotify.Event) {
	var path string
	var watch bool
	if isDir(ev.Name) {
		path = ev.Name
		watch = true
	} else {
		path = filepath.Dir(ev.Name)
		watch = false
	}

	w.create(path, watch)
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}

func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeSymlink == os.ModeSymlink
}

func maybeContainerID(name string) bool {
	if len(name) < 12 {
		return false
	}

	for i := 0; i < len(name); i++ {
		b := name[i]
		if !((b >= '0' && b <= '9') || (b >= 'a' && b <= 'f')) {
			return false
		}
	}

	return true
}

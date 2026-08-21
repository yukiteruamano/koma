package gache

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

// loadLocked loads the cache from disk. Caller must hold g.mutex exclusively.
func (g *Cache[T]) loadLocked() error {
	return g.loadInternal()
}

// load is the thread-safe wrapper (kept for backward compat with tests that call g.load()).
func (g *Cache[T]) load() error {
	// Called only from initLocked which already holds Lock, so just delegate.
	return g.loadInternal()
}

func (g *Cache[T]) loadInternal() error {
	if g.options.Path == "" {
		return nil
	}

	if err := g.options.FileSystem.MkdirAll(filepath.Dir(g.options.Path), 0755); err != nil {
		return err
	}

	// Avoid opening empty file as malformed: check size first.
	if info, err := g.options.FileSystem.Stat(g.options.Path); err == nil {
		if info.Size() == 0 {
			return g.tryExpireLocked()
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		// Stat failed for other reason, try to proceed with open.
	}

	file, err := g.options.FileSystem.OpenFile(g.options.Path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	// Decode; handle EOF (empty file) as empty cache.
	if err := g.options.Decoder.Decode(file, &g.data); err != nil {
		if errors.Is(err, io.EOF) {
			var defaultT T
			g.data = &chronoData[T]{
				Internal: defaultT,
				Time:     nil,
			}
		} else {
			// Malformed file: reset to empty and persist.
			var defaultT T
			g.data = &chronoData[T]{
				Internal: defaultT,
				Time:     nil,
			}
			if saveErr := g.saveLocked(); saveErr != nil {
				return saveErr
			}
		}
	}

	return g.tryExpireLocked()
}

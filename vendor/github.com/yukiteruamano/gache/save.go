package gache

import (
	"os"
	"path/filepath"
	"time"
)

// saveLocked persists the cache atomically. Caller must hold g.mutex exclusively.
// It always updates g.data.Time to now, even for in-memory caches.
func (g *Cache[T]) saveLocked() error {
	now := time.Now()
	g.data.Time = &now

	if g.options.Path == "" {
		return nil
	}

	if err := g.options.FileSystem.MkdirAll(filepath.Dir(g.options.Path), 0755); err != nil {
		return err
	}

	tmpPath := g.options.Path + ".tmp"

	file, err := g.options.FileSystem.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	encErr := g.options.Encoder.Encode(file, g.data)
	closeErr := file.Close()

	if encErr != nil {
		_ = g.options.FileSystem.Remove(tmpPath)
		return encErr
	}
	if closeErr != nil {
		_ = g.options.FileSystem.Remove(tmpPath)
		return closeErr
	}

	// Atomic rename.
	if err := g.options.FileSystem.Rename(tmpPath, g.options.Path); err != nil {
		_ = g.options.FileSystem.Remove(tmpPath)
		return err
	}

	return nil
}

// save is the exported helper that acquires the lock. Kept for backward compat.
func (g *Cache[T]) save() error {
	return g.saveLocked()
}

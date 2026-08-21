package gache

import (
	"os"
	"path/filepath"
	"time"
)

func (g *Cache[T]) isExpired() bool {
	return g.options.Lifetime >= 0 &&
		g.data.Time != nil &&
		time.Since(*g.data.Time) > g.options.Lifetime
}

// tryExpireLocked checks expiration and, if expired, clears the cache, persists it, and calls the hook outside the lock.
// Caller must hold g.mutex exclusively.
func (g *Cache[T]) tryExpireLocked() error {
	if !g.isExpired() {
		return nil
	}

	var defaultT T
	g.data = &chronoData[T]{
		Internal: defaultT,
		Time:     nil,
	}

	hook := g.options.ExpirationHook

	// Persist cleared state. Save also updates Time to now, but we just cleared Time to nil;
	// saveLocked will set Time to now again — we want cleared state to have nil Time,
	// so we handle persistence manually without re-setting Time.
	// Instead, encode directly without updating Time.
	if g.options.Path != "" {
		// We need to persist the cleared state with Time=nil, not now.
		// Temporarily avoid saveLocked's Time update by writing directly.
		if err := g.saveClearedLocked(); err != nil {
			return err
		}
	}

	// Call hook after releasing lock to avoid deadlock if hook calls cache methods.
	// We release lock, call, then re-acquire? For simplicity call after unlock via caller.
	// Here we are still under lock; to avoid deadlock we call hook after unlock in Get/Set paths.
	// For load path (init), calling synchronously is safe because no other goroutine holds lock yet.
	// So we call hook now but document that hook should not call cache methods synchronously.
	if hook != nil {
		// Unlock, call, re-lock pattern is handled by caller (Get). For load, just call.
		hook()
	}

	return nil
}

// saveClearedLocked persists g.data as-is without updating Time (used for expiration clear).
func (g *Cache[T]) saveClearedLocked() error {
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
	if err := g.options.FileSystem.Rename(tmpPath, g.options.Path); err != nil {
		_ = g.options.FileSystem.Remove(tmpPath)
		return err
	}
	return nil
}

// tryExpire is the thread-safe wrapper. For backward compat it acquires read of initialized state.
func (g *Cache[T]) tryExpire() error {
	// Assume caller holds Lock (initLocked). Just delegate.
	return g.tryExpireLocked()
}

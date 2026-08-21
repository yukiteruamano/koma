package gache

// Set sets the value of the cache.
// If initialization or marshalling fails, it will return an error.
// In memory-only mode it will never fail.
// It will restart the cache's lifetime.
func (g *Cache[T]) Set(value T) error {
	if err := g.init(); err != nil {
		return err
	}

	g.mutex.Lock()
	defer g.mutex.Unlock()

	// update value
	g.data.Internal = value

	// saveLocked updates Time and persists atomically
	return g.saveLocked()
}

// Get returns the value of the cache.
// If initialization fails, it will return an error.
// In memory-only mode it will never fail.
func (g *Cache[T]) Get() (cached T, expired bool, err error) {
	if err = g.init(); err != nil {
		return
	}

	g.mutex.RLock()
	if !g.isExpired() {
		cached = g.data.Internal
		g.mutex.RUnlock()
		return cached, false, nil
	}
	g.mutex.RUnlock()

	// Expired: need write lock to clear state and persist.
	g.mutex.Lock()
	// Double-check under write lock.
	if !g.isExpired() {
		cached = g.data.Internal
		g.mutex.Unlock()
		return cached, false, nil
	}

	// Clear cache
	var defaultT T
	hook := g.options.ExpirationHook
	g.data = &chronoData[T]{
		Internal: defaultT,
		Time:     nil,
	}

	// Persist cleared state without updating Time.
	var saveErr error
	if g.options.Path != "" {
		saveErr = g.saveClearedLocked()
	}
	g.mutex.Unlock()

	if hook != nil {
		hook()
	}

	if saveErr != nil {
		err = saveErr
		return
	}

	expired = true
	return
}

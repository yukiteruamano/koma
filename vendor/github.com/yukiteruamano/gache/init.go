package gache

// init ensures the cache is loaded from disk exactly once, thread-safely.
// It uses double-checked locking to avoid races.
func (g *Cache[T]) init() error {
	g.mutex.RLock()
	if g.initialized {
		g.mutex.RUnlock()
		return nil
	}
	g.mutex.RUnlock()

	g.mutex.Lock()
	defer g.mutex.Unlock()

	return g.initLocked()
}

// initLocked assumes g.mutex is held exclusively (Lock).
func (g *Cache[T]) initLocked() error {
	if g.initialized {
		return nil
	}

	if err := g.loadLocked(); err != nil {
		return err
	}

	g.initialized = true
	return nil
}

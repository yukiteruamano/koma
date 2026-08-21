package gache

import (
	"sync"
	"time"
)

// chronoData is a struct that contains the data that is stored in the cache file with time of the last update.
// Used to expire the cache.
type chronoData[T any] struct {
	Internal T
	Time     *time.Time
}

// Cache is a generic thread-safe cache that can be used to cache any type of data.
// It is used to cache data that is expensive to fetch, such as API responses.
// Cached data is stored in a file (JSON by default), and is automatically expired after a certain amount of time
// (if lifetime is specified)
type Cache[T any] struct {
	data    *chronoData[T]
	options *Options
	mutex   *sync.RWMutex

	initialized bool
}

// IsExpired reports whether the cache is expired without mutating it.
func (g *Cache[T]) IsExpired() bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.isExpired()
}

// Clear removes the cached value and timestamp, persists the empty state if file-backed, and returns any error.
func (g *Cache[T]) Clear() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if err := g.initLocked(); err != nil {
		return err
	}

	var defaultT T
	g.data = &chronoData[T]{
		Internal: defaultT,
		Time:     nil,
	}

	return g.saveClearedLocked()
}

// New returns a new Cache[T] instance.
func New[T any](options *Options) *Cache[T] {
	if options == nil {
		options = &Options{}
	}

	if options.FileSystem == nil {
		options.FileSystem = defaultFileSystem{}
	}

	if options.Encoder == nil {
		options.Encoder = defaultJSONEncoderDecoder{}
	}

	if options.Decoder == nil {
		options.Decoder = defaultJSONEncoderDecoder{}
	}

	if options.Lifetime == 0 {
		options.Lifetime = -1
	}

	_ = options.Validate()

	var defaultT T
	return &Cache[T]{
		options: options,
		mutex:   &sync.RWMutex{},
		data: &chronoData[T]{
			Internal: defaultT,
			Time:     nil,
		},
	}
}

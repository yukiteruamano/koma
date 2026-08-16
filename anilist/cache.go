package anilist

import (
	"sync"

	"github.com/metafates/gache"
	"github.com/samber/mo"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/where"
	"path/filepath"
	"time"
)

type cacheData[K comparable, T any] struct {
	Mangas map[K]T `json:"mangas"`
}

type cacher[K comparable, T any] struct {
	mu         sync.Mutex
	internal   *gache.Cache[*cacheData[K, T]]
	keyWrapper func(K) K
}

func (c *cacher[K, T]) Get(key K) mo.Option[T] {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, expired, err := c.internal.Get()
	if err != nil || expired || data == nil {
		return mo.None[T]()
	}

	mangas, ok := data.Mangas[c.keyWrapper(key)]
	if ok {
		return mo.Some(mangas)
	}

	return mo.None[T]()
}

func (c *cacher[K, T]) Set(key K, t T) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, expired, err := c.internal.Get()
	if err != nil {
		return err
	}

	if !expired && data != nil {
		data.Mangas[c.keyWrapper(key)] = t
		return c.internal.Set(data)
	} else {
		internal := &cacheData[K, T]{Mangas: make(map[K]T)}
		internal.Mangas[c.keyWrapper(key)] = t
		return c.internal.Set(internal)
	}
}

// SetMany updates many keys in a single read-modify-write, so the whole cache
// file is only rewritten once instead of once per key.
func (c *cacher[K, T]) SetMany(kvs map[K]T) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, expired, err := c.internal.Get()
	if err != nil {
		return err
	}

	if expired || data == nil {
		data = &cacheData[K, T]{Mangas: make(map[K]T)}
	}

	for k, v := range kvs {
		data.Mangas[c.keyWrapper(k)] = v
	}

	return c.internal.Set(data)
}

func (c *cacher[K, T]) Delete(key K) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, expired, err := c.internal.Get()
	if err != nil {
		return err
	}

	if !expired && data != nil {
		delete(data.Mangas, c.keyWrapper(key))
		return c.internal.Set(data)
	}

	return nil
}

var relationCacher = &cacher[string, int]{
	internal: gache.New[*cacheData[string, int]](
		&gache.Options{
			Path:       where.AnilistBinds(),
			Lifetime:   time.Hour * 24 * 30,
			FileSystem: &filesystem.GacheFs{},
		},
	),
	keyWrapper: normalizedName,
}

var searchCacher = &cacher[string, []int]{
	internal: gache.New[*cacheData[string, []int]](
		&gache.Options{
			Path:       filepath.Join(where.Cache(), "anilist_search_cache.json"),
			Lifetime:   time.Hour * 24 * 10,
			FileSystem: &filesystem.GacheFs{},
		},
	),
	keyWrapper: normalizedName,
}

var idCacher = &cacher[int, *Manga]{
	internal: gache.New[*cacheData[int, *Manga]](
		&gache.Options{
			Path:       filepath.Join(where.Cache(), "anilist_id_cache.json"),
			Lifetime:   time.Hour * 24 * 2,
			FileSystem: &filesystem.GacheFs{},
		},
	),
	keyWrapper: func(id int) int { return id },
}

var failCacher = &cacher[string, bool]{
	internal: gache.New[*cacheData[string, bool]](
		&gache.Options{
			Path:       filepath.Join(where.Cache(), "anilist_fail_cache.json"),
			Lifetime:   time.Minute,
			FileSystem: &filesystem.GacheFs{},
		},
	),
	keyWrapper: normalizedName,
}

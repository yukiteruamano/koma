package anilist

import (
	"path/filepath"
	"testing"

	"github.com/yukiteruamano/gache"
	"github.com/samber/mo"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/where"
)

func newTestCacher() *cacher[int, string] {
	filesystem.SetMemMapFs()

	return &cacher[int, string]{
		internal: gache.New[*cacheData[int, string]](
			&gache.Options{
				Path:       filepath.Join(where.Cache(), "test_cache.json"),
				FileSystem: &filesystem.GacheFs{},
			},
		),
		keyWrapper: func(id int) int { return id },
	}
}

func TestCacherSetManyAndGet(t *testing.T) {
	c := newTestCacher()

	values := map[int]string{1: "one", 2: "two", 3: "three"}
	if err := c.SetMany(values); err != nil {
		t.Fatalf("SetMany failed: %v", err)
	}

	for id, want := range values {
		got := c.Get(id)
		if got.IsAbsent() {
			t.Errorf("expected cache hit for %d", id)
			continue
		}
		if got.MustGet() != want {
			t.Errorf("Get(%d) = %q, want %q", id, got.MustGet(), want)
		}
	}
}

func TestCacherGetMiss(t *testing.T) {
	c := newTestCacher()

	if got := c.Get(42); got.IsPresent() {
		t.Errorf("expected a cache miss, got %v", mo.Some(got.MustGet()))
	}
}

func TestCacherSetThenDelete(t *testing.T) {
	c := newTestCacher()

	if err := c.Set(7, "seven"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if c.Get(7).IsAbsent() {
		t.Fatal("expected a cache hit after Set")
	}

	if err := c.Delete(7); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if c.Get(7).IsPresent() {
		t.Error("expected a cache miss after Delete")
	}
}

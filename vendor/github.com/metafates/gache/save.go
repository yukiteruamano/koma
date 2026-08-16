package gache

import (
	"os"
	"path/filepath"
	"time"
)

func (g *Cache[T]) save() error {
	// do nothing if we use in-memory caching
	if g.options.Path == "" {
		return nil
	}

	fs := g.options.FileSystem

	err := fs.MkdirAll(filepath.Dir(g.options.Path), 0777)
	if err != nil {
		return err
	}

	// write to a temp file and rename so a crash mid-write cannot truncate
	// the real cache file
	tmp := g.options.Path + ".tmp"
	file, err := fs.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	// save the time of the last update
	now := time.Now()
	g.data.Time = &now

	if err := g.options.Encoder.Encode(file, g.data); err != nil {
		_ = file.Close()
		_ = fs.Remove(tmp)
		return err
	}

	if err := file.Close(); err != nil {
		_ = fs.Remove(tmp)
		return err
	}

	return fs.Rename(tmp, g.options.Path)
}

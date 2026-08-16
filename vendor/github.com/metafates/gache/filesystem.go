package gache

import (
	"io"
	"os"
)

type FileSystem interface {
	OpenFile(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error)
	MkdirAll(path string, perm os.FileMode) error
	Remove(name string) error
	Rename(oldname, newname string) error
}

type defaultFileSystem struct {
}

func (defaultFileSystem) OpenFile(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	return os.OpenFile(name, flag, perm)
}

func (defaultFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (defaultFileSystem) Remove(name string) error {
	return os.Remove(name)
}

func (defaultFileSystem) Rename(oldname, newname string) error {
	return os.Rename(oldname, newname)
}

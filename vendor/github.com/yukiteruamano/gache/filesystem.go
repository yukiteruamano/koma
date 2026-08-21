package gache

import (
	"io"
	"os"
)

type FileSystem interface {
	OpenFile(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error)
	MkdirAll(path string, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)
	Rename(oldpath, newpath string) error
	Remove(name string) error
}

type defaultFileSystem struct {
}

func (defaultFileSystem) OpenFile(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	return os.OpenFile(name, flag, perm)
}

func (defaultFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (defaultFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (defaultFileSystem) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (defaultFileSystem) Remove(name string) error {
	return os.Remove(name)
}

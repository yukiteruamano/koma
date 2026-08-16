package filesystem

import (
	"io"
	"os"
)

type GacheFs struct {
}

func (GacheFs) OpenFile(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	return Api().OpenFile(name, flag, perm)
}

func (GacheFs) MkdirAll(path string, perm os.FileMode) error {
	return Api().MkdirAll(path, perm)
}

func (GacheFs) Remove(name string) error {
	return Api().Remove(name)
}

func (GacheFs) Rename(oldname, newname string) error {
	return Api().Rename(oldname, newname)
}

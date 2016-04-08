package fuse

import (
	"bazil.org/fuse/fs"
)

type FS struct {
	backend Backend
}

var _ = fs.FS(&FS{})

func (f *FS) Root() (fs.Node, error) {
	n := &Dir{
		fs:   f,
		path: []string{},
	}
	return n, nil
}

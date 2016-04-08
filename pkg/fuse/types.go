package fuse

import (
	"golang.org/x/net/context"
)

type DirSource interface {
	Dir(path []string) (DirLike, error)
}

type Backend interface {
	DirSource
	View(context.Context, func(Context) error) error
	Update(context.Context, func(Context) error) error
}

type Entry struct {
	Name string
	Dir  bool
}

type DirLike interface {
	GetDir(name string) (DirLike, error)
	CreateDir(name string) (DirLike, error)
	DeleteDir(name string) error
	Cursor() <-chan Entry
	Get(name string) ([]byte, error)
	Put(name string, value []byte) error
	Delete(name string) error
}

type Context interface {
	context.Context
	Dir([]string) (DirLike, error)
}

const (
	DirMarker = "~dir~"
)

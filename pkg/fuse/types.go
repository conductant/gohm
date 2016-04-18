package fuse

import (
	"golang.org/x/net/context"
	"io"
	"os"
)

type Backend interface {
	io.Closer
	DirLike
	View(context.Context, func(Context) error) error
	Update(context.Context, func(Context) error) error
}

type Entry struct {
	Name string
	Dir  bool
}

type Meta struct {
	Perm os.FileMode
	Size uint64
	Uid  uint32
	Gid  uint32
}

type DirLike interface {
	DirMeta() (Meta, error)
	GetDir(name string) (DirLike, error)
	CreateDir(name string) (DirLike, error)
	DeleteDir(name string) error
	Cursor() <-chan Entry
	Create(name string) error
	Meta(name string) (Meta, error)
	Get(name string) (interface{}, error)
	Put(name string, value interface{}) error
	Delete(name string) error
}

type Context interface {
	context.Context
	Dir([]string) (DirLike, error)
}

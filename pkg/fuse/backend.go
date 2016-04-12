package fuse

import (
	"golang.org/x/net/context"
)

type SimpleBackend struct {
	DirSource
}

func (this *SimpleBackend) GetContext(c context.Context) Context {
	return NewContext(c, this.DirSource)
}

func (this *SimpleBackend) View(c context.Context, f func(Context) error) error {
	return f(this.GetContext(c))
}

func (this *SimpleBackend) Update(c context.Context, f func(Context) error) error {
	return f(this.GetContext(c))
}

func (this *SimpleBackend) Close() error {
	return nil
}

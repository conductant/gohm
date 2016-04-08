package fuse

import (
	"fmt"
	"golang.org/x/net/context"
)

type context_t struct {
	context.Context
}

type dirSource_t int

const (
	dirSource_k dirSource_t = 2
)

func NewContext(ctx context.Context, dirSource DirSource) Context {
	return contextPutDirSource(&context_t{ctx}, dirSource)
}

func (this *context_t) Dir(path []string) (DirLike, error) {
	b := contextGetDirSource(this)
	if b == nil {
		return nil, fmt.Errorf("assert-dirSource-failed")
	}
	return b.Dir(path)
}

func contextGetDirSource(ctx *context_t) DirSource {
	if b, ok := ctx.Value(dirSource_k).(DirSource); ok {
		return b
	}
	return nil
}

func contextPutDirSource(ctx *context_t, b DirSource) *context_t {
	return &context_t{context.WithValue(ctx, dirSource_k, b)}
}

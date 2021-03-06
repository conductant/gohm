package fuse

import (
	"fmt"
	"golang.org/x/net/context"
)

type context_t struct {
	context.Context
}

type dirLike_t int

const (
	dirLike_k dirLike_t = 1
)

func NewContext(ctx context.Context, dirLike DirLike) Context {
	return contextPutDirLike(&context_t{ctx}, dirLike)
}

func (this *context_t) Dir(path []string) (DirLike, error) {
	b := contextGetDirLike(this)
	if b == nil {
		return nil, fmt.Errorf("assert-DirLike-failed")
	}
	var d DirLike = b
	for _, p := range path {
		if dir, err := d.GetDir(p); err != nil || dir == nil {
			return nil, err
		} else {
			d = dir
		}
	}
	return d, nil
}

func contextGetDirLike(ctx *context_t) DirLike {
	if b, ok := ctx.Value(dirLike_k).(DirLike); ok {
		return b
	}
	return nil
}

func contextPutDirLike(ctx *context_t, b DirLike) *context_t {
	return &context_t{context.WithValue(ctx, dirLike_k, b)}
}

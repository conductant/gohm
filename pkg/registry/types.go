package registry

import (
	"io"
	"net/url"
)

type Registry interface {
	io.Closer
	Id() url.URL
	Exists(Path) (bool, error)
	Get(Path) ([]byte, error)
	Put(Path, []byte, bool) error // Create or set.
	Delete(Path) error
	List(Path) ([]Path, error)
}

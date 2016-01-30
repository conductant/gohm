package registry

import (
	"golang.org/x/net/context"
	"io"
	net "net/url"
	"strings"
	"sync"
)

type Registry interface {
	io.Closer
	Id() net.URL
	Exists(Path) (bool, error)
	Get(Path) ([]byte, error)
	Put(Path, []byte, bool) error // Create or set.
	Delete(Path) error
	List(Path) ([]Path, error)
}

type Implementation func(ctx context.Context, url net.URL) (Registry, error)

type scheme string

var (
	lock      sync.Mutex
	protocols = map[scheme]Implementation{}
)

func Register(protocol string, impl Implementation) {
	lock.Lock()
	defer lock.Unlock()
	protocols[scheme(protocol)] = impl
}

// Get an instance of the registry.  The url can specify host(s) such as
// zk://host1:2181,host2:2181,host3:2181/other/parts/of/path
// The protocol / scheme portion is used to dispatch to different registry implementations (e.g. zk: for zookeeper
// etcd: for ectd, etc.)
func Dial(ctx context.Context, url string) (Registry, error) {
	u, err := net.Parse(url)
	if err != nil {
		return nil, err
	}
	if impl, has := protocols[scheme(u.Scheme)]; !has {
		return nil, &NotSupportedProtocol{u.Scheme}
	} else {
		return impl(ctx, *u)
	}
}

// Given the fully specified url that includes protocol and host and path,
// traverses symlinks where the value of a node is a pointer url to another registry node
// It's possible that the pointer points to a different registry.
// The returned url includes protocol and host information
func FollowUrl(ctx context.Context, url net.URL) (net.URL, []byte, error) {
	reg, err := Dial(ctx, url.String())
	if err != nil {
		return url, nil, err
	}
	return Follow(ctx, reg, NewPath(url.Path))
}

// Traverses symlinks where the value of a node is a pointer url to another registry node
// It's possible that the pointer points to a different registry.
func Follow(ctx context.Context, reg Registry, path Path) (net.URL, []byte, error) {
	here := reg.Id()
	here.Path = path.String()
	if len(path.String()) == 0 {
		return here, nil, nil
	}
	v, err := reg.Get(path)
	if err != nil {
		return here, nil, err
	}
	s := string(v)
	if strings.Contains(s, "://") {
		url, err := net.Parse(s)
		if err != nil {
			return here, nil, err
		}
		next, err := Dial(ctx, s)
		if err != nil {
			return here, nil, err
		}
		return Follow(ctx, next, NewPath(url.Path))
	} else {
		return here, v, nil
	}
}

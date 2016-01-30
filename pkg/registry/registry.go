package registry

import (
	net "net/url"
	"sync"
)

type ReadOnly interface {
}

type Registry interface {
	ReadOnly

	Get(Path) ([]byte, error)
	Put(Path, []byte) error
	Delete(Path) error
	List(Path) ([]Path, error)
}

type Implementation func(url net.URL, options ...interface{}) (Registry, error)

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
func Dial(url string, options ...interface{}) (Registry, error) {
	u, err := net.Parse(url)
	if err != nil {
		return nil, err
	}
	if impl, has := protocols[scheme(u.Scheme)]; !has {
		return nil, &NotSupportedProtocol{u.Scheme}
	} else {
		return impl(*u, options...)
	}
}

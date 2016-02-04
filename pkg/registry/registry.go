package registry

import (
	"golang.org/x/net/context"
	net "net/url"
	"strings"
)

// Get an instance of the registry.  The url can specify host(s) such as
// zk://host1:2181,host2:2181,host3:2181/other/parts/of/path
// The protocol / scheme portion is used to dispatch to different registry implementations (e.g. zk: for zookeeper
// etcd: for ectd, etc.)
func Dial(ctx context.Context, url string) (Registry, error) {
	u, err := net.Parse(url)
	if err != nil {
		return nil, err
	}
	// This allows the implementation to deal with cases where Host is not set, etc., so the implementations
	// have an opportunity to use its own environment variables, etc. to fix up the url.
	if sanitizer, has := sanitizers[scheme(u.Scheme)]; has {
		clean := sanitizer(*u)
		u = &clean
	}
	if ref := registries.get(u); ref != nil {
		registryReferenced <- ref.registry
		return ref.registry, nil
	}
	if impl, has := protocols[scheme(u.Scheme)]; !has {
		return nil, &NotSupportedProtocol{u.Scheme}
	} else {
		d := &dispose{
			propose: proposeToClose,
			accept:  make(chan bool),
		}
		reg, err := impl(ctx, *u, d)
		if err != nil {
			return nil, err
		} else {
			registryCreated <- &reference{registry: reg, count: 1, dispose: d}
			return reg, nil
		}
	}

}

// Given the fully specified url that includes protocol and host and path,
// traverses symlinks where the value of a node is a pointer url to another registry node
// It's possible that the pointer points to a different registry.
// The returned url includes protocol and host information
func FollowUrl(ctx context.Context, url net.URL) (net.URL, []byte, Version, error) {
	reg, err := Dial(ctx, url.String())
	if reg != nil {
		defer reg.Close()
	}
	if err != nil {
		return url, nil, InvalidVersion, err
	}
	return follow(ctx, reg, NewPath(url.Path))
}

// Traverses symlinks where the value of a node is a pointer url to another registry node
// It's possible that the pointer points to a different registry.
func follow(ctx context.Context, reg Registry, path Path) (net.URL, []byte, Version, error) {
	here := reg.Id()
	here.Path = path.String()
	if len(path.String()) == 0 {
		return here, nil, InvalidVersion, nil
	}
	v, ver, err := reg.Get(path)
	if err != nil {
		return here, nil, InvalidVersion, err
	}
	s := string(v)
	if strings.Contains(s, "://") {
		url, err := net.Parse(s)
		if err != nil {
			return here, nil, InvalidVersion, err
		}
		next, err := Dial(ctx, s)
		if next != nil {
			defer next.Close()
		}
		if err != nil {
			return here, nil, InvalidVersion, err
		}
		return follow(ctx, next, NewPath(url.Path))
	} else {
		return here, v, ver, nil
	}
}

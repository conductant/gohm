package registry

import (
	"golang.org/x/net/context"
	net "net/url"
)

// Registry backend implementations should follow this protocol to implement and register its services.

type Implementation func(ctx context.Context, url net.URL, dispose Dispose) (Registry, error)
type UrlSanitizer func(url net.URL) net.URL

// Protocol between implementation and the regstry.  The implementation is expected to
// send to Propose and then listen for a True on Accept.  If true then the implementation
// can actually close and dispose the resources.  This allows the central registry to
// implement shared connections and things like reference counting.
type Dispose interface {
	Propose() chan<- Registry
	Accept() <-chan bool
}

func Register(protocol string, impl Implementation) {
	lock.Lock()
	defer lock.Unlock()
	protocols[scheme(protocol)] = impl
}

func RegisterSanitizer(protocol string, impl UrlSanitizer) {
	lock.Lock()
	defer lock.Unlock()
	sanitizers[scheme(protocol)] = impl
}

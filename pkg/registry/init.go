package registry

import (
	"fmt"
	"github.com/golang/glog"
	net "net/url"
	"sync"
)

type dispose struct {
	propose chan Registry
	accept  chan bool
}

func (this *dispose) Propose() chan<- Registry {
	return this.propose
}
func (this *dispose) Accept() <-chan bool {
	return this.accept
}

// Reference counter for the given regstry
type reference struct {
	registry Registry
	count    int
	dispose  *dispose
}

type scheme string
type cacheKey string
type cache map[cacheKey]*reference

var (
	lock sync.Mutex

	protocols  = map[scheme]Implementation{}
	sanitizers = map[scheme]UrlSanitizer{}
	registries = cache{}

	proposeToClose     = make(chan Registry)
	registryCreated    = make(chan *reference)
	registryReferenced = make(chan Registry)
)

func synchronized(f func()) {
	lock.Lock()
	defer lock.Unlock()
	f()
}

func init() {
	// Go routine here to keep track of registries by protocol and host string.  For a given
	// scheme/host combination, we are using the same registry, unless it's been closed, which
	// will then cause the registry to report its closing and be removed from cache.
	go func() {
		for {
			select {
			case reg := <-proposeToClose:
				url := reg.Id()
				synchronized(func() {
					ref := registries.get(&url)
					glog.Infoln("Propose to close", ref.count)
					if ref != nil {
						ref.count--
						glog.Infoln("--------------------------------------------------------", ref.count)
						if ref.count == 0 {
							registries.remove(&url)
							ref.dispose.accept <- true
							glog.Infoln("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX", ref.count)
						} else {
							ref.dispose.accept <- false // do not dispose resources. others are using.
						}
					} else {
						panic("shouldn't be here!")
					}
				})
			case ref := <-registryCreated:
				synchronized(func() {
					registries.add(ref)
				})
			case reg := <-registryReferenced:
				url := reg.Id()
				synchronized(func() {
					ref := registries.get(&url)
					if ref != nil {
						ref.count++
						glog.Infoln("++++++++++++++++++++++++++++++++++++++++++++++++++++++++", ref.count)
					} else {
						panic("shouldn't be here")
					}
				})
			}
		}
	}()
}

func (this cache) key(u *net.URL) cacheKey {
	return cacheKey(fmt.Sprintf("%s://%s", u.Scheme, u.Host))
}

func (this cache) add(ref *reference) {
	url := ref.registry.Id()
	key := this.key(&url)
	if r, has := this[key]; !has {
		this[key] = ref
	} else {
		panic(fmt.Errorf("shouldn't be here! %v, same=%v", r, ref == r))
	}
}

func (this cache) get(url *net.URL) *reference {
	key := this.key(url)
	if ref, has := this[key]; has {
		return ref
	}
	return nil
}

func (this cache) remove(url *net.URL) {
	key := this.key(url)
	delete(this, key)
}

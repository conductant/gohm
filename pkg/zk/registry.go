package zk

import (
	. "github.com/conductant/gohm/pkg/registry"
	"golang.org/x/net/context"
	"net/url"
	"strings"
)

func init() {
	Register("zk", NewService)
}

// Optional parameter is timeout, in Duration.
func NewService(ctx context.Context, url url.URL) (Registry, error) {
	// Look for a duration and use that as the timeout
	timeout := ContextGetTimeout(ctx)
	servers := strings.Split(url.Host, ",") // host:port,host:port,...
	return Connect(servers, timeout)
}

func (this *client) Id() url.URL {
	return this.url
}

func (this *client) Exists(key Path) (bool, error) {
	_, err := this.GetNode(key.String())
	switch err {
	case ErrNotExist:
		return false, nil
	case nil:
		return true, nil
	default:
		return false, err
	}
}

func (this *client) Get(key Path) ([]byte, error) {
	n, err := this.GetNode(key.String())
	if err != nil {
		return nil, err
	}
	return n.Value, nil
}

func (this *client) List(key Path) ([]Path, error) {
	n, err := this.GetNode(key.String())
	if err != nil {
		return nil, err
	}
	children, err := n.Children()
	if err != nil {
		return nil, err
	}
	paths := []Path{}
	for _, n := range children {
		paths = append(paths, NewPath(n.Path))
	}
	return paths, nil
}

func (this *client) Delete(key Path) error {
	return this.DeleteNode(key.String())
}

func (this *client) Put(key Path, value []byte, ephemeral bool) error {
	_, err := this.PutNode(key.String(), value, ephemeral)
	return err
}

func (this *client) Trigger(t Trigger) (<-chan interface{}, chan<- int, error) {
	stop := make(chan int)
	events := make(chan interface{}, 512)

	var cStop chan<- bool
	var err error
	switch t := t.(type) {
	case Create:
		cStop, err = this.Watch(t.Path.String(),
			func(e Event) {
				if e.Type == EventNodeCreated {
					events <- e
				}
			})
		if err != nil {
			return nil, nil, err
		}
	case Change:
		cStop, err = this.Watch(t.Path.String(),
			func(e Event) {
				if e.Type == EventNodeDataChanged {
					events <- e
				}
			})
		if err != nil {
			return nil, nil, err
		}
	case Delete:
		cStop, err = this.Watch(t.Path.String(),
			func(e Event) {
				if e.Type == EventNodeDeleted {
					events <- e
				}
			})
		if err != nil {
			return nil, nil, err
		}
	case Members:
		// TODO - Implement the matching criteria using min/max/delta, etc.
		cStop, err = this.WatchChildren(t.Path.String(),
			func(e Event) {
				if e.Type == EventNodeChildrenChanged {
					events <- e
				}
			})
		if err != nil {
			return nil, nil, err
		}
	}
	go func() {
		// Stop the watch
		<-stop
		cStop <- true
	}()
	return events, stop, nil
}

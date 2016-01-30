package zk

import (
	. "github.com/conductant/gohm/pkg/registry"
	"net/url"
	"strings"
	"time"
)

func init() {
	Register("zk", NewService)
}

// Optional parameter is timeout, in Duration.
func NewService(url url.URL, options ...interface{}) (Registry, error) {
	// Look for a duration and use that as the timeout
	timeout := DefaultTimeout
	for _, opt := range options {
		if t, ok := opt.(time.Duration); ok {
			timeout = t
			break
		}
	}
	servers := strings.Split(url.Host, ",") // host:port,host:port,...
	return Connect(servers, timeout)
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

func (this *client) Put(key Path, value []byte) error {
	return this.createOrSet(key, value, false)
}

func (this *client) createOrSet(key Path, value []byte, ephemeral bool) error {
	if ephemeral {
		_, err := this.CreateEphemeralNode(key.String(), value)
		return err
	}
	n, err := this.GetNode(key.String())
	switch {
	case err == ErrNotExist:
		n, err = this.CreateNode(key.String(), value)
		if err != nil {
			return err
		}
	case err != nil:
		return err
	}
	err = n.Set(value)
	if err != nil {
		return err
	}
	return nil
}

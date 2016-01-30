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

func (this *client) Put(key Path, value []byte) error {
	_, err := this.PutNode(key.String(), value, false)
	return err
}

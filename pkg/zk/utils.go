package zk

import (
	"github.com/golang/glog"
	"github.com/samuel/go-zookeeper/zk"
	"os"
	"path/filepath"
	"strings"
)

func Hosts() []string {
	servers := []string{"localhost:2181"}
	list := os.Getenv("ZK_HOSTS")
	if len(list) > 0 {
		servers = strings.Split(list, ",")
	}
	glog.Infoln("zk-hosts:", servers)
	return servers
}

func listParents(path string) []string {
	p := path
	if p[0:1] != "/" {
		p = "/" + path // Must begin with /
	}
	pp := strings.Split(p, "/")
	t := []string{}
	root := ""
	for _, x := range pp[1:] {
		z := root + "/" + x
		root = z
		t = append(t, z)
	}
	return t
}

func (this *client) createParents(path string) error {
	dir := filepath.Dir(path)
	if dir == "." {
		return nil
	}
	for _, p := range listParents(dir) {
		exists, _, err := this.conn.Exists(p)
		if err != nil {
			return err
		}
		if !exists {
			_, err := this.createNode(p, []byte{}, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func run_watch(path string, f func(Event), event_chan <-chan zk.Event, optionalStop ...chan bool) (chan bool, error) {
	if f == nil {
		return nil, nil
	}

	stop := make(chan bool, 1)
	if len(optionalStop) > 0 {
		stop = optionalStop[0]
	}

	go func() {
		// Note ZK only fires once and after that we need to reschedule.
		// With this api this may mean we get a new event channel.
		// Therefore, there's no point looping in here for more than 1 event.
		select {
		case event := <-event_chan:
			f(Event{Event: event})
		case b := <-stop:
			if b {
				glog.Infoln("watch-terminated:", "path=", path)
				return
			}
		}
	}()
	return stop, nil
}

func append_string_slices(a, b []string) []string {
	l := len(a)
	ll := make([]string, l+len(b))
	copy(ll, a)
	for i, n := range b {
		ll[i+l] = n
	}
	return ll
}

func append_node_slices(a, b []*Node) []*Node {
	l := len(a)
	ll := make([]*Node, l+len(b))
	copy(ll, a)
	for i, n := range b {
		ll[i+l] = n
	}
	return ll
}
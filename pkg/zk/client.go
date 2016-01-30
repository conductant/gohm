package zk

import (
	"errors"
	"github.com/golang/glog"
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

type client struct {
	conn    *zk.Conn
	servers []string
	timeout time.Duration
	events  chan Event

	ephemeral        map[string]*Node
	ephemeral_add    chan *Node
	ephemeral_remove chan string

	retry      chan *Node
	retry_stop chan int
	stop       chan int

	running bool

	watch_stops_chan chan chan bool
	watch_stops      map[chan bool]bool

	shutdown chan int
}

func (this *client) on_connect() {
	for _, n := range this.ephemeral {
		this.retry <- n
	}
}

// ephemeral flag here is user requested.
func (this *client) trackEphemeral(zn *Node, ephemeral bool) {
	if ephemeral || (zn.Stats != nil && zn.Stats.EphemeralOwner > 0) {
		glog.Infoln("ephemeral-add:", "path=", zn.Path)
		this.ephemeral_add <- zn
	}
}

func (this *client) untrackEphemeral(path string) {
	this.ephemeral_remove <- path
}

func Connect(servers []string, timeout time.Duration) (*client, error) {
	conn, events, err := zk.Connect(servers, timeout)
	if err != nil {
		return nil, err
	}
	zz := &client{
		conn:             conn,
		servers:          servers,
		timeout:          timeout,
		events:           make(chan Event, 1024),
		stop:             make(chan int),
		ephemeral:        map[string]*Node{},
		ephemeral_add:    make(chan *Node),
		ephemeral_remove: make(chan string),
		retry:            make(chan *Node, 1024),
		retry_stop:       make(chan int),
		watch_stops:      make(map[chan bool]bool),
		watch_stops_chan: make(chan chan bool),
		shutdown:         make(chan int),
	}

	go func() {
		<-zz.shutdown
		zz.doShutdown()
		glog.Infoln("Shutdown complete.")
	}()

	go func() {
		defer glog.Infoln("ZK watcher cache stopped.")
		for {
			watch_stop, open := <-zz.watch_stops_chan
			if !open {
				return
			}
			zz.watch_stops[watch_stop] = true
		}
	}()
	go func() {
		defer glog.Infoln("ZK ephemeral cache stopped.")
		for {
			select {
			case add, open := <-zz.ephemeral_add:
				if !open {
					return
				}
				zz.ephemeral[add.Path] = add
				glog.Infoln("EPHEMERAL-CACHE-ADD: Path=", add.Path, "Value=", string(add.Value))

			case remove, open := <-zz.ephemeral_remove:
				if !open {
					return
				}
				if _, has := zz.ephemeral[remove]; has {
					delete(zz.ephemeral, remove)
					glog.Infoln("EPHEMERAL-CACHE-REMOVE: Path=", remove)
				}
			}
		}
	}()
	go func() {
		defer glog.Infoln("ZK event loop stopped")
		for {
			select {
			case evt := <-events:
				glog.Infoln("ZK-Event-Main:", evt)
				switch evt.State {
				case StateExpired:
					glog.Warningln("ZK state expired --> sent by server on reconnection.")
					// This is actually connected, despite the state name, because the server
					// sends this event on reconnection.
					zz.on_connect()
				case StateHasSession:
					glog.Warningln("ZK state has-session")
					zz.on_connect()
				case StateDisconnected:
					glog.Warningln("ZK state disconnected")
				}
				zz.events <- Event{Event: evt}
			case <-zz.stop:
				return
			}
		}
	}()
	go func() {
		defer glog.Infoln("ZK ephemeral resync loop stopped")
		for {
			select {
			case r := <-zz.retry:
				if r != nil {
					_, err := zz.CreateNode(r.Path, r.Value, true)
					switch err {
					case nil, ErrNodeExists:
						glog.Infoln("emphemeral-resync: Key=", r.Path, "retry ok.")
						zz.events <- Event{Event: zk.Event{Path: r.Path}, Action: "Ephemeral-Retry", Note: "retry ok"}
					default:
						glog.Infoln("emphemeral-resync: Key=", r.Path, "Err=", err, "retrying.")
						select {
						case zz.retry <- r:
							glog.Infoln("emphemeral-resync:", r.Path, "submitted")
							select {
							case zz.events <- Event{
								Event:  zk.Event{Path: r.Path},
								Action: "ephemeral-resync",
								Note:   "retrying"}:
							}
						default:
							glog.Warningln("ephemeral-resync: dropped object", r.Path)
						}
					}
				}
			case <-zz.retry_stop:
				return
			}
		}
	}()

	glog.Infoln("Connected to zk:", servers)
	return zz, nil
}

func (this *client) check() error {
	if this.conn == nil {
		return ErrNotConnected
	}
	return nil
}

func (this *client) Events() <-chan Event {
	return this.events
}

func (this *client) Close() error {
	this.shutdown <- 1
	// wait for a close
	<-this.shutdown
	return nil
}

func (this *client) doShutdown() {
	glog.Infoln("Shutting down...")

	close(this.ephemeral_add)
	close(this.ephemeral_remove)

	close(this.stop)
	close(this.retry_stop)

	for w, _ := range this.watch_stops {
		close(w)
	}
	close(this.watch_stops_chan)

	this.conn.Close()
	this.conn = nil

	close(this.shutdown)
}

func (this *client) Reconnect() error {
	p, err := Connect(this.servers, this.timeout)
	if err != nil {
		return err
	} else {
		this = p
		return nil
	}
}

func (this *client) GetNode(path string) (*Node, error) {
	if err := this.check(); err != nil {
		return nil, err
	}

	exists, _, err := this.conn.Exists(path)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotExist
	}
	value, stats, err := this.conn.Get(path)
	if err != nil {
		return nil, err
	}
	return &Node{Path: path, Value: value, Stats: stats, client: this}, nil
}

func (this *client) WatchOnce(path string, f func(Event)) (chan<- bool, error) {
	if err := this.check(); err != nil {
		return nil, err
	}
	_, _, event_chan, err := this.conn.ExistsW(path)
	if err != nil {
		return nil, err
	}
	return run_watch(path, f, event_chan)
}

func (this *client) WatchOnceChildren(path string, f func(Event)) (chan<- bool, error) {
	if err := this.check(); err != nil {
		return nil, err
	}

	_, _, event_chan, err := this.conn.ChildrenW(path)
	switch {

	case err == ErrNotExist:
		_, _, event_chan0, err0 := this.conn.ExistsW(path)
		if err0 != nil {
			return nil, err0
		}
		// First watch for creation
		// Use a common stop
		stop1 := make(chan bool)
		_, err1 := run_watch(path, func(e Event) {
			if e.Type == zk.EventNodeCreated {
				if _, _, event_chan2, err2 := this.conn.ChildrenW(path); err2 == nil {
					// then watch for children
					run_watch(path, f, event_chan2, stop1)
				}
			}
		}, event_chan0, stop1)
		return stop1, err1

	case err == nil:
		return run_watch(path, f, event_chan)

	default:
		return nil, err
	}
}

// Continuously watch a path, optional callbacks for errors.
func (this *client) Watch(path string, f func(Event) bool, alerts ...func(error)) (chan<- bool, error) {
	if err := this.check(); err != nil {
		return nil, err
	}
	if f == nil {
		return nil, errors.New("error-nil-watcher")
	}

	_, _, event_chan, err := this.conn.ExistsW(path)
	if err != nil {
		go func() {
			for _, a := range alerts {
				a(err)
			}
		}()
		return nil, err
	}
	stop := make(chan bool)
	this.watch_stops_chan <- stop
	go func() {
		for {
			select {
			case event := <-event_chan:

				more := true

				glog.Infoln("watch-state-change", "path=", path, "state=", event.State)
				switch event.State {
				case zk.StateExpired:
					for _, a := range alerts {
						a(ErrSessionExpired)
					}
				case zk.StateDisconnected:
					for _, a := range alerts {
						a(ErrConnectionClosed)
					}
				default:
					more = f(Event{Event: event})
				}
				if more {
					// Retry loop
					for {
						glog.Infoln("watch-retry: Trying to set watch on", path)
						_, _, event_chan, err = this.conn.ExistsW(path)
						if err == nil {
							glog.Infoln("watch-retry: Continue watching", path)
							this.events <- Event{Event: zk.Event{Path: path}, Action: "Watch-Retry", Note: "retry ok"}
							break
						} else {
							glog.Warningln("watch-retry: Error -", path, err)
							for _, a := range alerts {
								a(err)
							}
							// Wait a little
							time.Sleep(1 * time.Second)
							glog.Infoln("watch-retry: Finished waiting. Try again to watch", path)
							this.events <- Event{Event: zk.Event{Path: path}, Action: "watch-retry", Note: "retrying"}
						}
					}
				}

			case <-stop:
				glog.Infoln("watch: Watch terminated:", "path=", path)
				return
			}
		}
	}()
	glog.Infoln("watch: Started watch on", "path=", path)
	return stop, nil
}

// Creates a new node.  If node already exists, error will be returned.
func (this *client) CreateNode(path string, value []byte, ephemeral bool) (*Node, error) {
	if err := this.check(); err != nil {
		return nil, err
	}
	// Make sure all parents exist
	err := this.createParents(path)
	if err != nil {
		return nil, err
	}
	return this.createNode(path, value, ephemeral)
}

// Assumes all parent nodes have been created.
func (this *client) createNode(path string, value []byte, ephemeral bool) (*Node, error) {
	key := path
	flags := int32(0)
	if ephemeral {
		flags = int32(zk.FlagEphemeral)
	}
	acl := zk.WorldACL(zk.PermAll) // TODO - PermAll permission
	p, err := this.conn.Create(key, value, flags, acl)
	if err != nil {
		return nil, err
	}
	zn := &Node{Path: p, Value: value, client: this}
	this.trackEphemeral(zn, ephemeral)
	return this.GetNode(p)
}

// Sets the node value, creates if not exists.
func (this *client) PutNode(key string, value []byte, ephemeral bool) (*Node, error) {
	if ephemeral {
		n, err := this.CreateNode(key, value, true)
		return n, err
	}
	n, err := this.GetNode(key)
	switch err {
	case nil:
		return n, n.Set(value)
	case ErrNotExist:
		n, err = this.CreateNode(key, value, false)
		if err != nil {
			return nil, err
		} else {
			return n, n.Set(value)
		}
	default:
		return nil, err
	}
}

func (this *client) DeleteNode(path string) error {
	if err := this.check(); err != nil {
		return err
	}
	this.untrackEphemeral(path)
	return this.conn.Delete(path, -1)
}

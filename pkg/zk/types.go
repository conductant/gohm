package zk

import (
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

const (
	StateUnknown           = zk.StateUnknown
	StateDisconnected      = zk.StateDisconnected
	StateConnecting        = zk.StateConnecting
	StateAuthFailed        = zk.StateAuthFailed
	StateConnectedReadOnly = zk.StateConnectedReadOnly
	StateSaslAuthenticated = zk.StateSaslAuthenticated
	StateExpired           = zk.StateExpired
	StateConnected         = zk.StateConnected
	StateHasSession        = zk.StateHasSession

	DefaultTimeout = 1 * time.Hour
)

type Node struct {
	Path    string
	Value   []byte
	Members []string
	Stats   *zk.Stat
	Leaf    bool
	client  *client
}

type Event struct {
	zk.Event
	Action string
	Note   string
}

type Service interface {
	Reconnect() error
	Close() error
	Events() <-chan Event
	CreateNode(string, []byte) (*Node, error)
	CreateEphemeralNode(string, []byte) (*Node, error)
	GetNode(string) (*Node, error)
	WatchOnce(string, func(Event)) (chan<- bool, error)
	WatchOnceChildren(string, func(Event)) (chan<- bool, error)
	KeepWatch(string, func(Event) bool, ...func(error)) (chan<- bool, error)
	DeleteNode(string) error
}

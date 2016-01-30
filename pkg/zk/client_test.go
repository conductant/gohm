package zk

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	. "gopkg.in/check.v1"
	"testing"
	"time"
)

func TestClient(t *testing.T) { TestingT(t) }

type ClientTests struct{}

var _ = Suite(&ClientTests{})

func (suite *ClientTests) TearDownSuite(c *C) {
	z, err := Connect(Hosts(), 5*time.Second)
	c.Assert(err, Equals, nil)
	z.DeleteNode("/unit-test") // TODO - this fails before there are children under this node
}

func (suite *ClientTests) TestConnect(c *C) {
	z, err := Connect(Hosts(), 5*time.Second)
	c.Assert(err, Equals, nil)
	c.Log("Got client", z)
	c.Assert(z.conn, Not(Equals), nil)
	z.Close()
	c.Assert(z.conn, Equals, (*zk.Conn)(nil))

	// Reconnect
	err = z.Reconnect()
	c.Assert(err, Equals, nil)
	c.Assert(z.conn, Not(Equals), nil)
}

func (suite *ClientTests) TestBasicOperations(c *C) {
	z, err := Connect(Hosts(), time.Second)
	c.Assert(err, Equals, nil)

	defer z.Close()

	path := "/unit-test/test"

	value := []byte("/unit-test/test")
	v, err := z.GetNode(path)
	c.Assert(err, Not(Equals), ErrNotConnected)

	if err == ErrNotExist {
		v, err = z.CreateNode(path, value, false)
		c.Assert(err, Equals, nil)
		c.Assert(v, Not(Equals), nil)
	}

	// Now create a bunch of children
	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("/unit-test/test/%d", i)
		data := fmt.Sprintf("value-test-%04d", i)

		x, err := z.GetNode(k)
		if err == ErrNotExist {
			x, err := z.CreateNode(k, []byte(data), false)
			c.Assert(err, Equals, nil)
			err = x.Load()
			c.Assert(err, Equals, nil)
			c.Assert(string(x.Value), Equals, data)
		} else {
			// update
			err = x.Set([]byte(data))
			c.Assert(err, Equals, nil)
			c.Assert(string(x.Value), Equals, data)
		}
	}

	// Get children
	children, err := v.SubtreeNodes()
	c.Assert(err, Equals, nil)
	for _, n := range children {
		c.Assert(n.CountChildren(), Equals, int32(0)) // expects leaf nodes
	}

	// Get the full list of children
	paths, err := v.SubtreePaths()
	c.Assert(err, Equals, nil)
	for _, p := range paths {
		_, err := z.GetNode(p)
		c.Assert(err, Equals, nil)
	}

	all_children, err := v.SubtreeNodes()
	c.Assert(err, Equals, nil)
	for _, n := range all_children {
		err := n.Delete()
		c.Assert(err, Equals, nil)
	}
}

func (suite *ClientTests) TestFullPathObjects(c *C) {
	z, err := Connect(Hosts(), time.Second)
	c.Assert(err, Equals, nil)

	defer z.Close()

	top, err := z.GetNode("/unit-test/dir1")
	if err == ErrNotExist {
		top, err = z.CreateNode("/unit-test/dir1", nil, false)
		c.Assert(err, Equals, nil)
	}
	c.Assert(top, Not(Equals), (*Node)(nil))
	all_children, err := top.SubtreeNodes()
	c.Assert(err, Equals, nil)
	for _, n := range all_children {
		c.Log("Deleting", n.Path)
		err := n.Delete()
		c.Assert(err, Equals, nil)
	}

	path := "/unit-test/dir1/dir2/dir3"
	value := []byte(path)
	v, err := z.CreateNode(path, value, false)
	c.Assert(err, Equals, nil)
	c.Assert(v, Not(Equals), nil)

	for i := 0; i < 5; i++ {
		k := fmt.Sprintf("/unit-test/dir1/dir2/dir3/dir4/%04d", i)
		v := fmt.Sprintf("%s", i)
		_, err := z.CreateNode(k, []byte(v), false)
		c.Assert(err, Equals, nil)
	}
	// Get the full list of children
	paths, err := v.SubtreePaths()
	c.Assert(err, Equals, nil)
	for _, p := range paths {
		_, err := z.GetNode(p)
		c.Assert(err, Equals, nil)
		c.Log("> ", p)
	}
}

func (suite *ClientTests) TestAppEnvironments(c *C) {
	z, err := Connect(Hosts(), time.Second)
	c.Assert(err, Equals, nil)

	defer z.Close()

	// Common use case of storing properties under different environments
	expects := map[string]string{
		"/unit-test/environments/integration/database/driver":     "postgres",
		"/unit-test/environments/integration/database/dbname":     "pg_db",
		"/unit-test/environments/integration/database/user":       "pg_user",
		"/unit-test/environments/integration/database/password":   "password",
		"/unit-test/environments/integration/database/port":       "5432",
		"/unit-test/environments/integration/env/S3_API_KEY":      "s3-api-key",
		"/unit-test/environments/integration/env/S3_API_SECRET":   "s3-api-secret",
		"/unit-test/environments/integration/env/SERVICE_API_KEY": "service-api-key",
	}

	for k, v := range expects {
		_, err = z.CreateNode(k, []byte(v), false)
		c.Log(k, "err", err)
		//c.Assert(err, Equals, nil)
	}

	integration, err := z.GetNode("/unit-test/environments/integration")
	c.Assert(err, Equals, nil)

	all, err := integration.FilterSubtreeNodes(func(z *Node) bool {
		return !z.Leaf // filter out parent nodes
	})
	c.Assert(err, Equals, nil)

	for _, n := range all {
		c.Log(n.Basename(), "=", n.ValueString())
	}
	c.Assert(len(all), Equals, len(expects)) // exactly as the map since we filter out the parent node /database

	for _, n := range all {
		err = n.Delete()
		c.Assert(err, Equals, nil)
	}
}

func (suite *ClientTests) TestEphemeral(c *C) {
	z1, err := Connect(Hosts(), time.Second)
	c.Assert(err, Equals, nil)

	p := "/unit-test/e1/e2"
	top1, err := z1.GetNode(p)
	if err == ErrNotExist {
		top1, err = z1.CreateNode(p, nil, false)
		c.Assert(err, Equals, nil)
	}
	err = top1.Load()
	c.Assert(err, Equals, nil)
	c.Log("top1", top1)

	top11, err := z1.CreateNode(p+"/11", nil, true)
	c.Assert(err, Equals, nil)
	c.Log("top1", top11)

	z2, err := Connect(Hosts(), time.Second)
	c.Assert(err, Equals, nil)
	top2, err := z2.GetNode(p + "/11")
	c.Assert(err, Not(Equals), ErrNotExist)
	c.Log("z2 sees", top2)

	z1.Close() // the ephemeral node /11 should go away

	err = top2.Load()
	c.Log("top2", top2)
	c.Assert(err, Equals, ErrNotExist)

	// what about the static one
	top22, err := z2.GetNode(p)
	c.Assert(err, Not(Equals), ErrNotExist)
	c.Log("z2 sees", top22)

	z2.Close()
}

func (suite *ClientTests) TestWatcher(c *C) {
	z1, err := Connect(Hosts(), time.Second)
	c.Assert(err, Equals, nil)

	p := "/unit-test/" + fmt.Sprintf("%d", time.Now().Unix()) + "/e1/e2"
	top1, err := z1.GetNode(p)
	if err == ErrNotExist {
		top1, err = z1.CreateNode(p, nil, false)
		c.Assert(err, Equals, nil)
	}
	err = top1.Load()
	c.Assert(err, Equals, nil)
	c.Log("top1", top1)

	top11, err := z1.CreateNode(p+"/11", nil, true)
	c.Assert(err, Equals, nil)
	c.Log("top1", top11)

	// Watched by another client
	z2, err := Connect(Hosts(), time.Second)
	c.Assert(err, Equals, nil)

	top22, err := z2.GetNode(p + "/11")
	c.Assert(err, Not(Equals), ErrNotExist)
	c.Log("z2 sees", top22)

	stop22, err := top22.WatchOnce(func(e Event) {
		if e.State != zk.StateDisconnected {
			c.Log("Got event :::::", e)
		}
	})
	c.Assert(err, Equals, nil)

	// Now do a few things
	top22.Set([]byte("New value"))

	// Now watch something else
	new_path := "/unit-test/new/path/to/be/created"
	stop23, err := z2.WatchOnce(new_path, func(e Event) {
		if e.State != zk.StateDisconnected {
			c.Log("Got event -----", e)
		}
	})
	c.Assert(err, Equals, nil)

	// Create a new node
	_, err = z1.CreateNode(new_path, nil, true)
	c.Assert(err, Equals, nil)

	c.Log("closing z1")
	z1.Close() // the ephemeral node /11 should go away

	time.Sleep(1 * time.Second)
	c.Log("sending stop")
	stop22 <- true
	stop23 <- true
	c.Log("stop sent")
}

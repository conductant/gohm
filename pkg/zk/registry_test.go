package zk

import (
	"fmt"
	"github.com/conductant/gohm/pkg/registry"
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
	"strings"
	"testing"
	"time"
)

func TestRegistry(t *testing.T) { TestingT(t) }

type TestSuiteRegistry struct{}

var _ = Suite(&TestSuiteRegistry{})

func (suite *TestSuiteRegistry) SetUpSuite(c *C) {
}

func (suite *TestSuiteRegistry) TearDownSuite(c *C) {
}

func (suite *TestSuiteRegistry) TestUsage(c *C) {
	ctx := ContextPutTimeout(context.Background(), 1*time.Minute)
	url := "zk://" + strings.Join(Hosts(), ",")
	zk, err := registry.Dial(ctx, url)
	c.Assert(err, IsNil)
	c.Log(zk)

	p := registry.NewPath("/unit-test/registry/test")
	v := []byte("test")
	err = zk.Put(p, v, false)
	c.Assert(err, IsNil)
	read, err := zk.Get(p)
	c.Assert(read, DeepEquals, v)

	check := map[registry.Path]int{}
	for i := 0; i < 10; i++ {
		cp := p.Sub(fmt.Sprintf("child-%d", i))
		err = zk.Put(cp, []byte{0}, false)
		c.Assert(err, IsNil)
		check[cp] = i
	}

	list, err := zk.List(p)
	c.Assert(err, IsNil)
	c.Log(list)
	c.Assert(len(list), Equals, len(check))
	for _, p := range list {
		_, has := check[p]
		c.Assert(has, Equals, true)
	}

	// delete all children
	for i := 0; i < 10; i++ {
		cp := p.Sub(fmt.Sprintf("child-%d", i))
		err = zk.Delete(cp)
		c.Assert(err, IsNil)
	}
	list, err = zk.List(p)
	c.Assert(err, IsNil)
	c.Assert(len(list), Equals, 0)

	exists, err := zk.Exists(p.Sub("child-0"))
	c.Assert(err, IsNil)
	c.Assert(exists, Equals, false)
}

func (suite *TestSuiteRegistry) TestEphemeral(c *C) {
	ctx := ContextPutTimeout(context.Background(), 1*time.Minute)
	url := "zk://" + strings.Join(Hosts(), ",")
	zk, err := registry.Dial(ctx, url)
	c.Assert(err, IsNil)
	c.Log(zk)

	p := registry.NewPath("/unit-test/registry/ephemeral")
	v := []byte("test")
	err = zk.Put(p, v, true)
	c.Assert(err, IsNil)
	read, err := zk.Get(p)
	c.Assert(read, DeepEquals, v)
	exists, _ := zk.Exists(p)
	c.Assert(exists, Equals, true)
	// disconnect
	err = zk.Close()
	c.Assert(err, IsNil)

	// reconnect
	zk, err = registry.Dial(ctx, url)
	c.Assert(err, IsNil)
	_, err = zk.Get(p)
	c.Assert(err, Equals, ErrNotExist)
	exists, _ = zk.Exists(p)
	c.Assert(exists, Equals, false)
}

func (suite *TestSuiteRegistry) TestFollow(c *C) {
	ctx := ContextPutTimeout(context.Background(), 1*time.Minute)
	url := "zk://" + strings.Join(Hosts(), ",")
	zk, err := registry.Dial(ctx, url)
	c.Assert(err, IsNil)
	c.Log(zk)

	p := registry.NewPath("/unit-test/registry/follow")

	err = zk.Put(p.Sub("1"), []byte(url+p.Sub("2").String()), false)
	c.Assert(err, IsNil)

	err = zk.Put(p.Sub("2"), []byte(url+p.Sub("3").String()), false)
	c.Assert(err, IsNil)

	err = zk.Put(p.Sub("3"), []byte(url+p.Sub("4").String()), false)
	c.Assert(err, IsNil)

	err = zk.Put(p.Sub("4"), []byte("end"), false)
	c.Assert(err, IsNil)

	path, value, err := registry.Follow(ctx, zk, p.Sub("1"))
	c.Assert(err, IsNil)
	c.Assert(value, DeepEquals, []byte("end"))
	c.Assert(path.String(), Equals, url+p.Sub("4").String())
}

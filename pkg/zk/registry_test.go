package zk

import (
	"fmt"
	"github.com/conductant/gohm/pkg/registry"
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
	url := "zk://" + strings.Join(Hosts(), ",")
	zk, err := registry.Dial(url, 1*time.Minute)
	c.Assert(err, IsNil)
	c.Log(zk)

	p := registry.NewPath("/unit-test/registry/test")
	v := []byte("test")
	err = zk.Put(p, v)
	c.Assert(err, IsNil)
	read, err := zk.Get(p)
	c.Assert(read, DeepEquals, v)

	check := map[registry.Path]int{}
	for i := 0; i < 10; i++ {
		cp := p.Sub(fmt.Sprintf("child-%d", i))
		err = zk.Put(cp, []byte{0})
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
}

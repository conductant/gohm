package fuse

import (
	. "gopkg.in/check.v1"
	"testing"
)

func TestBackend(t *testing.T) { TestingT(t) }

type TestSuiteBackend struct {
}

var _ = Suite(&TestSuiteBackend{})

func (suite *TestSuiteBackend) SetUpSuite(c *C) {
}

func (suite *TestSuiteBackend) TearDownSuite(c *C) {
}

func (suite *TestSuiteBackend) TestBackend(c *C) {
	err := Serve("/tmp/unit-test", NewMapBackend(map[string]interface{}{}), nil)
	c.Assert(err, IsNil)
}

package registry

import (
	. "gopkg.in/check.v1"
	"testing"
)

func TestTrigger(t *testing.T) { TestingT(t) }

type TestSuiteTrigger struct {
}

var _ = Suite(&TestSuiteTrigger{})

func (suite *TestSuiteTrigger) SetUpSuite(c *C) {
}

func (suite *TestSuiteTrigger) TearDownSuite(c *C) {
}

func (suite *TestSuiteTrigger) TestUsage(c *C) {
	create := Create{NewPath("/this/is/a/path")}
	c.Assert(create.Base(), Equals, "path")
	c.Assert(create.Dir(), Equals, NewPath("/this/is/a"))
	c.Assert(create.Dir().String(), Equals, "/this/is/a")

	members := (&Members{Path: NewPath("/path/to/parent")}).SetMin(32)
	c.Assert(members.Path, Equals, NewPath("/path/to/parent"))
	c.Assert(*members.Min, Equals, 32)
}

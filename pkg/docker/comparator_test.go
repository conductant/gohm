package docker

import (
	. "gopkg.in/check.v1"
	"testing"
)

func TestComparator(t *testing.T) { TestingT(t) }

type TestSuiteComparator struct {
}

var _ = Suite(&TestSuiteComparator{})

func (suite *TestSuiteComparator) SetUpSuite(c *C) {
}

func (suite *TestSuiteComparator) TearDownSuite(c *C) {
}

func (suite *TestSuiteComparator) TestImageEquals(c *C) {
	for _, cmp := range []func(interface{}, interface{}) bool{
		ImageEquals,
	} {
		c.Assert(cmp("postgres", "postgres"), Equals, true)
		c.Assert(cmp("foo/bar:v1.0-23", "foo/bar:v1.0-23"), Equals, true)
		c.Assert(cmp("foo/bar:v1.0-24.5", "foo/bar:v1.0-24.6"), Equals, false)
		c.Assert(cmp("foo/bar:v1.0-23", "foo/baz:v1.0-24"), Equals, false)
		c.Assert(cmp("foo/bar:v1.0.23", []byte("foo/bar:v1.0.23")), Equals, false)
		c.Assert(cmp([]byte("foo/bar:v1.0.23"), []byte("foo/bar:v1.0.23")), Equals, false)
	}
}

func (suite *TestSuiteComparator) TestImageLessBySemanticVersion(c *C) {
	for _, cmp := range []func(interface{}, interface{}) bool{
		ImageLessBySemanticVersion,
	} {
		c.Assert(cmp("foo/bar:v1.0-23", "foo/bar:v1.0-24"), Equals, true)
		c.Assert(cmp("foo/bar:v1.0-24.5", "foo/bar:v1.0-24.6"), Equals, true)
		c.Assert(cmp("foo/bar:v1.0-24.5", "foo/bar:v2.0-24.6"), Equals, true)
		c.Assert(cmp("foo/bar:v1.0-24.5", "foo/barxxx:v2.0-24.6"), Equals, false)
		c.Assert(cmp("foo/bar:v1.0-23", "foo/baz:v1.0-24"), Equals, false)
		c.Assert(cmp("foo/bar:v1.0.23", "foo/baz:v1.0.24"), Equals, false)
		c.Assert(cmp("foo/bar:v1.0.23", "foo/baz:v1.0.22"), Equals, false)
		c.Assert(cmp("foo/bar:master-9996.2864", "foo/bar:master-10002.2867"), Equals, true)
		c.Assert(cmp("foo/bar:master-9996.2864-10", "foo/bar:master-10002.2867-2"), Equals, true)
		c.Assert(cmp("foo/bar:master-9996.2864-1", "foo/bar:master-9996.2864-10"), Equals, true)
		c.Assert(cmp("foo/bar:master-9996.2864-1", "foo/bazz:master-9996.2864-10"), Equals, false)
	}
}

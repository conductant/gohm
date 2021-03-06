package fuse

import (
	log "github.com/Sirupsen/logrus"
	. "gopkg.in/check.v1"
	"os"
	"testing"
)

func TestBackend(t *testing.T) { TestingT(t) }

type TestSuiteBackend struct {
}

var _ = Suite(&TestSuiteBackend{})

func (suite *TestSuiteBackend) SetUpSuite(c *C) {
	log.SetLevel(log.DebugLevel)
}

func (suite *TestSuiteBackend) TearDownSuite(c *C) {
}

func (suite *TestSuiteBackend) TestBackend(c *C) {
	err := Serve(os.Getenv("HOME")+"/tmp/unit-test", NewMapBackend(map[string]interface{}{}), nil)
	c.Assert(err, IsNil)
}

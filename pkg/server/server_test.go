package server

import (
	. "gopkg.in/check.v1"
	"testing"
)

func TestServer(t *testing.T) { TestingT(t) }

type TestSuiteServer struct {
}

var _ = Suite(&TestSuiteServer{})

func (suite *TestSuiteServer) SetUpSuite(c *C) {
}

func (suite *TestSuiteServer) TearDownSuite(c *C) {
}

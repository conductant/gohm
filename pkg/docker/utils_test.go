package docker

import (
	. "gopkg.in/check.v1"
	"testing"
)

func TestUtils(t *testing.T) { TestingT(t) }

type UtilsTests struct{}

var _ = Suite(&UtilsTests{})

func (suite *UtilsTests) TestEth0Interface(c *C) {
	ips, err := GetEth0Ip()
	c.Assert(err, Equals, nil)
	c.Log("found", ips)
}

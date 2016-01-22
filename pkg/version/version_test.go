package version

import (
	"github.com/golang/glog"
	. "gopkg.in/check.v1"
	"testing"
)

func TestVersion(t *testing.T) { TestingT(t) }

type TestSuiteVersion struct {
}

var _ = Suite(&TestSuiteVersion{})

func (suite *TestSuiteVersion) SetUpSuite(c *C) {
}

func (suite *TestSuiteVersion) TearDownSuite(c *C) {
}

func (suite *TestSuiteVersion) TestShowBuild(c *C) {
	info := BuildInfo()
	c.Log(info.Notice())
	glog.Infoln(info.Notice())

	c.Check(info.GetRepoUrl(), Equals, "git@github.com:conductant&gohm.git")
	c.Check(info.GetBranch(), Equals, "master")
	c.Assert(info.GetCommitHash(), Not(Equals), "")
	c.Assert(info.GetNumber(), Not(Equals), "")
}

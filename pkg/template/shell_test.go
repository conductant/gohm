package template

import (
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
	"io"
	"io/ioutil"
	"testing"
)

func TestShell(t *testing.T) { TestingT(t) }

type TestSuiteShell struct {
}

var _ = Suite(&TestSuiteShell{})

func (suite *TestSuiteShell) SetUpSuite(c *C) {
}

func (suite *TestSuiteShell) TearDownSuite(c *C) {
}

func print(c *C, out io.Reader) {
	bytes, err := ioutil.ReadAll(out)
	c.Log(string(bytes), err)
}

func toString(c *C, out io.Reader) string {
	bytes, err := ioutil.ReadAll(out)
	c.Assert(err, IsNil)
	return string(bytes)
}

func (suite *TestSuiteShell) TestRunShell(c *C) {
	f := ExecuteShell(context.Background())
	shell, ok := f.(func(string) (io.Reader, error))
	c.Assert(ok, Equals, true)

	var stdout io.Reader
	var err error
	_, err = shell("echo '***********'")
	c.Assert(err, IsNil)

	stdout, err = shell("echo foo | sed -e 's/f/g/g'")
	sed := toString(c, stdout)

	_, err = shell("echo '***********'")
	c.Assert(err, IsNil)

	stdout, err = shell("echo $USER")
	home := toString(c, stdout)

	_, err = shell("echo '***********'")
	c.Assert(err, IsNil)

	stdout, err = shell("ls | wc -l")
	ls := toString(c, stdout)

	c.Log("sed=", sed)
	c.Assert(len(sed), Not(Equals), 0)
	c.Log("home=", home)
	c.Assert(len(home), Not(Equals), 0)
	c.Log("ls=", ls)
	c.Assert(len(ls), Not(Equals), 0)
}

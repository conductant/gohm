package template

import (
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
	"testing"
)

func TestTemplate(t *testing.T) { TestingT(t) }

type TestSuiteTemplate struct {
}

var _ = Suite(&TestSuiteTemplate{})

func (suite *TestSuiteTemplate) SetUpSuite(c *C) {
}

func (suite *TestSuiteTemplate) TearDownSuite(c *C) {
}

func func1(ctx context.Context, url string, opts ...string) (string, error) {
	return "", nil
}

func (suite *TestSuiteTemplate) TestBuilder(c *C) {

}

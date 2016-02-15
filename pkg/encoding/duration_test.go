package encoding

import (
	"bytes"
	. "gopkg.in/check.v1"
	"testing"
	"time"
)

func TestDuration(t *testing.T) { TestingT(t) }

type TestSuiteDuration struct {
}

var _ = Suite(&TestSuiteDuration{})

func (suite *TestSuiteDuration) SetUpSuite(c *C) {
}

func (suite *TestSuiteDuration) TearDownSuite(c *C) {
}

func (suite *TestSuiteDuration) TestDuration(c *C) {
	type Ticker struct {
		Interval Duration `json:"interval" yaml:"interval"`
	}

	jsonInput := `
{
   "name" : "ticker",
   "interval" : "20s"
}
`
	p := new(Ticker)
	p2 := new(Ticker)

	err := UnmarshalJSON(bytes.NewBufferString(jsonInput), p)
	c.Assert(err, IsNil)
	c.Assert(p.Interval.Duration, Equals, 20*time.Second)

	err = Unmarshal(ContentTypeJSON, bytes.NewBufferString(jsonInput), p2)
	c.Assert(p2.Interval.Duration, Equals, 20*time.Second)

	yamlInput := `
# Test data
   name : jane  # the name
   interval : 20s
`

	err = UnmarshalYAML(bytes.NewBufferString(yamlInput), p)
	c.Assert(err, IsNil)
	c.Assert(p.Interval.Duration, Equals, 20*time.Second)

	err = Unmarshal(ContentTypeYAML, bytes.NewBufferString(yamlInput), p2)
	c.Assert(err, IsNil)
	c.Assert(p2.Interval.Duration, Equals, 20*time.Second)

	// Marshal
	buff := new(bytes.Buffer)
	err = Marshal(ContentTypeYAML, buff, p)
	c.Assert(err, IsNil)
	c.Log(buff.String())
	c.Assert(buff.String(), DeepEquals, "interval: 20s\n")

	buff = new(bytes.Buffer)
	err = Marshal(ContentTypeJSON, buff, p)
	c.Assert(err, IsNil)
	c.Log(buff.String())
	c.Assert(buff.String(), DeepEquals, "{\"interval\":\"20s\"}")
}

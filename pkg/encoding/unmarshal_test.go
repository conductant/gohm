package encoding

import (
	"bytes"
	. "gopkg.in/check.v1"
	"testing"
)

func TestUnmarshal(t *testing.T) { TestingT(t) }

type TestSuiteUnmarshal struct {
}

var _ = Suite(&TestSuiteUnmarshal{})

func (suite *TestSuiteUnmarshal) SetUpSuite(c *C) {
}

func (suite *TestSuiteUnmarshal) TearDownSuite(c *C) {
}

func (suite *TestSuiteUnmarshal) TestUnmarshal(c *C) {
	type Person struct {
		Name string `json:"name,omitempty" yaml:"name,omitempty" flag:"name,n,The name"`
		Age  int    `json:"age,omitempty" yaml:"age,omitempty" flag:"age,a,The age"`
	}

	jsonInput := `
{
   "name" : "joe",
   "age" : 21,
   "sex" : "M"
}
`
	p := new(Person)
	p2 := new(Person)

	err := UnmarshalJSON(bytes.NewBufferString(jsonInput), p)
	c.Assert(err, IsNil)
	c.Assert(p.Name, Equals, "joe")
	c.Assert(p.Age, Equals, 21)

	err = Unmarshal(ContentTypeJSON, bytes.NewBufferString(jsonInput), p2)
	c.Assert(err, IsNil)
	c.Assert(p2.Name, Equals, "joe")
	c.Assert(p2.Age, Equals, 21)

	yamlInput := `
# Test data
   name : jane  # the name
   age : 22
   sex : F  # not parsed
`

	err = UnmarshalYAML(bytes.NewBufferString(yamlInput), p)
	c.Assert(err, IsNil)
	c.Assert(p.Name, Equals, "jane")
	c.Assert(p.Age, Equals, 22)

	err = Unmarshal(ContentTypeYAML, bytes.NewBufferString(yamlInput), p2)
	c.Assert(err, IsNil)
	c.Assert(p2.Name, Equals, "jane")
	c.Assert(p2.Age, Equals, 22)
}

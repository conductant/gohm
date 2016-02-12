package flag

import (
	"flag"
	. "gopkg.in/check.v1"
	"strings"
	"testing"
	"time"
)

func TestFlag(t *testing.T) { TestingT(t) }

type TestSuiteFlag struct {
}

var _ = Suite(&TestSuiteFlag{})

func (suite *TestSuiteFlag) SetUpSuite(c *C) {
}

func (suite *TestSuiteFlag) TearDownSuite(c *C) {
}

func (suite *TestSuiteFlag) TestFlag(c *C) {
	type Address struct {
		Line1 string `flag:"line1"`
		Line2 string `flag:"line2"`
		Post  int    `flag:"post"`
	}

	type Person struct {
		Name           string          `json:"name,omitempty" yaml:"name,omitempty" flag:"name,The name"`
		Age            int             `json:"age,omitempty" yaml:"age,omitempty" flag:"age,The age"`
		Employee       bool            `flag:"True if employee"`
		Height         int64           `flag:"height,The height"`
		Awake          time.Duration   `flag:"awake,How long person is awake."`
		Int64          int64           `flag:"int64 value"`
		Float64        float64         `flag:"float64 value"`
		Uint           uint            `flag:"uint"`
		Uint64         uint64          `flag:"uint64"`
		Addr           Address         `flag:"the Address"`
		Scopes         []string        `flag:"s,the scopes"`
		DefaultStrings []string        `flag:"d, the defaults"`
		DefaultInts    []int           `flag:"i,the default ints"`
		Int64List      []int64         `flag:"i64,The int64 values"`
		Float64List    []float64       `flag:"f64,The float64 list"`
		BoolList       []bool          `flag:"b,The bool list"`
		Uint64List     []uint64        `flag:"ui64,The uint64 list"`
		DurationList   []time.Duration `flag:"td,The time durations"`
		UintList       []uint          `flag:"ui,The uint list"`
		promote        bool
		NoFlag         bool
	}

	p := &Person{
		Age:            18,
		Employee:       true,
		DefaultStrings: []string{"bar", "baz"},
		DefaultInts:    []int{1, 2, 3},
	}

	fs := flag.NewFlagSet("person", flag.ContinueOnError)
	RegisterFlags("person", p, fs)

	c.Log("person=", p)

	var err error

	err = fs.Parse(strings.Split("--name=david --age=20 --person.employee=false --height=170 --awake=20h", " "))
	c.Assert(err, IsNil)

	c.Log("person=", p)

	c.Assert(p.Name, Equals, "david")
	c.Assert(p.Age, Equals, 20)
	c.Assert(p.Employee, Equals, false)
	c.Assert(p.promote, Equals, false)
	c.Assert(p.Height, Equals, int64(170))
	c.Assert(p.Awake, Equals, time.Hour*20)
	c.Assert(p.Int64, Equals, int64(0))
	c.Assert(p.Float64, Equals, 0.)
	c.Assert(p.Uint, Equals, uint(0))
	c.Assert(p.Uint64, Equals, uint64(0))
	c.Assert(p.Addr.Line1, Equals, "")
	c.Assert(p.Addr.Line2, Equals, "")
	c.Assert(p.Addr.Post, Equals, 0)
	c.Assert(p.Scopes, DeepEquals, []string{})
	c.Assert(p.DefaultStrings, DeepEquals, []string{"bar", "baz"})

	// parse agin
	err = fs.Parse(strings.Split("--person.int64=-64 --person.float64=64.0 --person.uint=32 --person.uint64=64", " "))
	c.Assert(err, IsNil)

	c.Log("person=", p)

	c.Assert(p.Name, Equals, "david")
	c.Assert(p.Age, Equals, 20)
	c.Assert(p.Employee, Equals, false)
	c.Assert(p.promote, Equals, false)
	c.Assert(p.Height, Equals, int64(170))
	c.Assert(p.Awake, Equals, time.Hour*20)
	c.Assert(p.Int64, Equals, int64(-64))
	c.Assert(p.Float64, Equals, 64.)
	c.Assert(p.Uint, Equals, uint(32))
	c.Assert(p.Uint64, Equals, uint64(64))
	c.Assert(p.Addr.Line1, Equals, "")
	c.Assert(p.Addr.Line2, Equals, "")
	c.Assert(p.Addr.Post, Equals, 0)
	c.Assert(p.Scopes, DeepEquals, []string{})
	c.Assert(p.DefaultStrings, DeepEquals, []string{"bar", "baz"})

	err = fs.Parse(strings.Split("--person.addr.line1=line1 --person.addr.line2=line2 --person.addr.post=1234", " "))
	c.Assert(err, IsNil)

	c.Log("person=", p)

	c.Assert(p.Name, Equals, "david")
	c.Assert(p.Age, Equals, 20)
	c.Assert(p.Employee, Equals, false)
	c.Assert(p.promote, Equals, false)
	c.Assert(p.Height, Equals, int64(170))
	c.Assert(p.Awake, Equals, time.Hour*20)
	c.Assert(p.Int64, Equals, int64(-64))
	c.Assert(p.Float64, Equals, 64.)
	c.Assert(p.Uint, Equals, uint(32))
	c.Assert(p.Uint64, Equals, uint64(64))
	c.Assert(p.Addr.Line1, Equals, "line1")
	c.Assert(p.Addr.Line2, Equals, "line2")
	c.Assert(p.Addr.Post, Equals, 1234)
	c.Assert(p.Scopes, DeepEquals, []string{})
	c.Assert(p.DefaultStrings, DeepEquals, []string{"bar", "baz"})

	err = fs.Parse(strings.Split("-s=foo -s=bar", " "))
	c.Assert(err, IsNil)

	c.Log("person=", p)

	c.Assert(p.Name, Equals, "david")
	c.Assert(p.Age, Equals, 20)
	c.Assert(p.Employee, Equals, false)
	c.Assert(p.promote, Equals, false)
	c.Assert(p.Height, Equals, int64(170))
	c.Assert(p.Awake, Equals, time.Hour*20)
	c.Assert(p.Int64, Equals, int64(-64))
	c.Assert(p.Float64, Equals, 64.)
	c.Assert(p.Uint, Equals, uint(32))
	c.Assert(p.Uint64, Equals, uint64(64))
	c.Assert(p.Addr.Line1, Equals, "line1")
	c.Assert(p.Addr.Line2, Equals, "line2")
	c.Assert(p.Addr.Post, Equals, 1234)
	c.Assert(p.Scopes, DeepEquals, []string{"foo", "bar"})
	c.Assert(p.DefaultStrings, DeepEquals, []string{"bar", "baz"})

	// Parse again
	err = fs.Parse(strings.Split("--d=beep --d=bop", " "))
	c.Assert(err, IsNil)
	c.Assert(p.DefaultStrings, DeepEquals, []string{"beep", "bop"})

	// Parse again
	c.Assert(p.DefaultInts, DeepEquals, []int{1, 2, 3})
	err = fs.Parse(strings.Split("--i=3 --i=2 --i=1", " "))
	c.Assert(err, IsNil)
	c.Assert(p.DefaultInts, DeepEquals, []int{3, 2, 1})

	// Parse again
	c.Assert(p.Int64List, DeepEquals, []int64{})
	err = fs.Parse(strings.Split("--i64=3000 --i64=2000 --i64=1000", " "))
	c.Assert(err, IsNil)
	c.Assert(p.Int64List, DeepEquals, []int64{3000, 2000, 1000})

	// Parse again
	c.Assert(p.Float64List, DeepEquals, []float64{})
	err = fs.Parse(strings.Split("--f64=3.1415 --f64=2.0 --f64=100.0000001", " "))
	c.Assert(err, IsNil)
	c.Assert(p.Float64List, DeepEquals, []float64{3.1415, 2.0, 100.0000001})

	// Parse again
	c.Assert(p.BoolList, DeepEquals, []bool{})
	err = fs.Parse(strings.Split("--b=1 --b=false --b=true --b=0", " "))
	c.Assert(err, IsNil)
	c.Assert(p.BoolList, DeepEquals, []bool{true, false, true, false})

	// Parse again
	c.Assert(p.Uint64List, DeepEquals, []uint64{})
	err = fs.Parse(strings.Split("--ui64=1 --ui64=2", " "))
	c.Assert(err, IsNil)
	c.Assert(p.Uint64List, DeepEquals, []uint64{1, 2})

	// Parse again
	c.Assert(p.DurationList, DeepEquals, []time.Duration{})
	err = fs.Parse(strings.Split("--td=1m --td=2h --td=35s", " "))
	c.Assert(err, IsNil)
	c.Assert(p.DurationList, DeepEquals, []time.Duration{1 * time.Minute, 2 * time.Hour, 35 * time.Second})

	// Parse again
	c.Assert(p.UintList, DeepEquals, []uint{})
	err = fs.Parse(strings.Split("--ui=1 --ui=655823", " "))
	c.Assert(err, IsNil)
	c.Assert(p.UintList, DeepEquals, []uint{1, 655823})
}

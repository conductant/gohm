package command

import (
	"bytes"
	"github.com/conductant/gohm/pkg/auth"
	"github.com/conductant/gohm/pkg/encoding"
	"github.com/conductant/gohm/pkg/resource"
	"github.com/conductant/gohm/pkg/server"
	"github.com/conductant/gohm/pkg/testutil"
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCommand(t *testing.T) { TestingT(t) }

type TestSuiteCommand struct {
	port     int
	template string // template content to serve
	scope    string
	stop     chan<- int
	stopped  <-chan error
}

var _ = Suite(&TestSuiteCommand{port: 7986})

func (suite *TestSuiteCommand) SetUpSuite(c *C) {
	suite.stop, suite.stopped = server.NewService().
		ListenPort(suite.port).
		WithAuth(server.Auth{VerifyKeyFunc: testutil.PublicKeyFunc}.Init()).
		Route(server.Endpoint{UrlRoute: "/secure", HttpMethod: server.GET, AuthScope: server.AuthScope("secure")}).
		To(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte(suite.template))
	}).Start()
}

func (suite *TestSuiteCommand) TearDownSuite(c *C) {
}

// Note that we only allow name and age to have json / yaml representations as well as flag settings.
// ConfigUrl is only settable via flag.
type person struct {
	ConfigUrl string `json:"-" yaml:"-" flag:"config_url,The config url"`
	Name      string `json:"name,omitempty" yaml:"name,omitempty" flag:"name,The name"`
	Age       int    `json:"age,omitempty" yaml:"age,omitempty" flag:"age,The age"`
	Employee  bool   `json:"-" yaml:"-" flag:"True if employee"`
	Height    int64  `json:"-" yaml:"-" flag:"height,The height"`
}

// Implements Module
func (p *person) Close() error {
	return nil
}
func (p *person) Help(io.Writer) {
}
func (p *person) Run([]string, io.Writer) error {
	ReparseFlags(p)
	return nil
}

// This test here shows how to implement two stages of configuring an object.
// First a field in the struct gets the source of the config url from the command line flags.
// Second, the config is fetched and unmarshalled to the object.
// Finally, we parse again so that additional flag values are overlaid onto the struct.
func (suite *TestSuiteCommand) TestCommandReparseFlag(c *C) {
	Register("person", func() (Module, ErrorHandling) {
		return new(person), PanicOnError
	})

	// Generate the auth token required by the server.
	token := auth.NewToken(1*time.Hour).Add("secure", 1)
	header := http.Header{}
	token.SetHeader(header, testutil.PrivateKeyFunc)
	ctx := resource.ContextPutHttpHeader(context.Background(), header)

	suite.template = `
name: joe
age: 21
employee: false
not-a-field: hello
`

	p := &person{
		Age:      18,
		Employee: true,
	}

	RunModule("person", p, strings.Split("--config_url=http://localhost:7986/secure --age=35", " "), nil)

	c.Assert(p.ConfigUrl, Equals, "http://localhost:7986/secure")
	c.Assert(p.Employee, Equals, true)

	data, err := resource.Fetch(ctx, p.ConfigUrl)
	c.Assert(err, IsNil)
	c.Log(string(data))
	err = encoding.Unmarshal(encoding.ContentTypeYAML, bytes.NewBuffer(data), p)
	c.Assert(err, IsNil)
	c.Assert(p.Age, Equals, 21)
	c.Assert(p.Name, Equals, "joe")
	c.Assert(p.Employee, Equals, true) // we don't expect the yaml to change the field.

	ReparseFlags(p)
	c.Assert(p.Age, Equals, 35) // This is overwritten by the flag value
	c.Assert(p.Name, Equals, "joe")
	c.Assert(p.Employee, Equals, true)
}

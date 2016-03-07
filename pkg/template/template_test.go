package template

import (
	"encoding/json"
	"fmt"
	"github.com/conductant/gohm/pkg/auth"
	"github.com/conductant/gohm/pkg/resource"
	"github.com/conductant/gohm/pkg/server"
	"github.com/conductant/gohm/pkg/testutil"
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestTemplate(t *testing.T) { TestingT(t) }

type TestSuiteTemplate struct {
	port     int
	template string // template content to serve
	scope    string
	stop     chan<- int
	stopped  <-chan error
}

var _ = Suite(&TestSuiteTemplate{port: 7983})

// Spins up a test server that will serve the template text from an authenticated endpoint.
func (suite *TestSuiteTemplate) SetUpSuite(c *C) {
	suite.stop, suite.stopped = server.NewService().
		ListenPort(suite.port).
		WithAuth(server.Auth{VerifyKeyFunc: testutil.PublicKeyFunc}.Init()).
		Route(server.Endpoint{UrlRoute: "/secure", HttpMethod: server.GET, AuthScope: server.AuthScope("secure")}).
		To(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
			resp.Write([]byte(suite.template))
		}).Start()
}

func (suite *TestSuiteTemplate) TearDownSuite(c *C) {
	suite.stop <- 1
	<-suite.stopped
}

func (suite *TestSuiteTemplate) TestTemplateExecute(c *C) {
	// Generate the auth token required by the server.
	token := auth.NewToken(1*time.Hour).Add("secure", 1)
	header := http.Header{}
	token.SetHeader(header, testutil.PrivateKeyFunc)

	// This is the content to be served by the test server.
	suite.template = "My name is {{.Name}} and I am {{.Age}} years old."

	// The url is a template too.
	url := "http://localhost:{{.port}}/secure"

	data := map[string]interface{}{
		"Name": "test",
		"Age":  20,
		"port": suite.port,
	}
	ctx := ContextPutTemplateData(resource.ContextPutHttpHeader(context.Background(), header), data)

	text, err := Execute(ctx, url)
	c.Assert(err, IsNil)
	c.Log(string(text))
	c.Assert(string(text), Equals, "My name is test and I am 20 years old.")
}

func (suite *TestSuiteTemplate) TestApply2(c *C) {

	body := `
{{define "app"}}
version: 1.2
image: repo
build: 1234
{{end}}

{{define "host"}}
label: appserver
name: myhost
{{end}}

{
   "image" : "repo/myapp:` + "{{my `app.version`}}-{{my `app.build`}}" + `",
   "host" : "` + "{{my `host.name`}}" + `"
}
`

	result, err := Apply2([]byte(body), nil)
	c.Assert(err, IsNil)

	c.Log(string(result))

	obj := make(map[string]string)
	err = json.Unmarshal(result, &obj)
	c.Assert(err, IsNil)
	c.Assert(obj["image"], Equals, "repo/myapp:1.2-1234")
	c.Assert(obj["host"], Equals, "myhost")
}

func (suite *TestSuiteTemplate) TestTemplateExecuteWithVarBlock(c *C) {
	// Generate the auth token required by the server.
	token := auth.NewToken(1*time.Hour).Add("secure", 1)
	header := http.Header{}
	token.SetHeader(header, testutil.PrivateKeyFunc)

	// This is the content to be served by the test server.
	// Using the Execute will add additional functions that can be included in the var blocks.
	suite.template = `
{{define "comments"}}
# The variable define blocks allow the definition of variables that are reused throughout the
# main body of the template.  The variables are referenced as '<blockname>.<fieldname>'.
{{end}}

{{define "app"}} # blockname is 'app'
version: 1.2
image: repo
build: 1234
dir: {{sh "pwd"}} # Invoke shell and use that as the value.
user: "{{env "USER"}}"  # Getting the environment variable and use that as value.
{{end}}

{{define "host"}}
label: appserver
name: myhost
port: {{.port}}  # Here we allow the application to pass in a context that's refereceable.
{{end}}

{
   "image" : "repo/myapp:` + "{{my `app.version`}}-{{my `app.build`}}" + `",
   "host" : "` + "{{my `host.name`}}" + `",{{/* use this for comment in JSON :) */}}
   "dir" : "` + "{{my `app.dir`}}" + `",
   "user" : "` + "{{my `app.user`}}" + `",
   "port" : "` + "{{my `host.port`}}" + `"
}`

	// The url is a template too.
	url := "http://localhost:{{.port}}/secure"

	data := map[string]interface{}{
		"Name": "test",
		"Age":  20,
		"port": suite.port,
	}
	ctx := ContextPutTemplateData(resource.ContextPutHttpHeader(context.Background(), header), data)

	text, err := Execute(ctx, url)
	c.Assert(err, IsNil)
	c.Log(string(text))
	obj := make(map[string]string)
	err = json.Unmarshal(text, &obj)
	c.Assert(err, IsNil)
	c.Assert(obj["image"], Equals, "repo/myapp:1.2-1234")
	c.Assert(obj["user"], Equals, os.Getenv("USER"))
	c.Assert(obj["host"], Equals, "myhost")
	c.Assert(obj["port"], Equals, fmt.Sprintf("%d", suite.port))
	c.Assert(obj["dir"], Equals, os.Getenv("PWD"))
}

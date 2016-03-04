package template

import (
	"encoding/json"
	"github.com/conductant/gohm/pkg/auth"
	"github.com/conductant/gohm/pkg/resource"
	"github.com/conductant/gohm/pkg/server"
	"github.com/conductant/gohm/pkg/testutil"
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
	"net/http"
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

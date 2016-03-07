package zk

import (
	"encoding/json"
	"fmt"
	"github.com/conductant/gohm/pkg/auth"
	"github.com/conductant/gohm/pkg/namespace"
	"github.com/conductant/gohm/pkg/resource"
	"github.com/conductant/gohm/pkg/server"
	"github.com/conductant/gohm/pkg/template"
	"github.com/conductant/gohm/pkg/testutil"
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
	"net/http"
	"os"
	"strings"
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

var _ = Suite(&TestSuiteTemplate{port: 7985})

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

func (suite *TestSuiteTemplate) TestGet(c *C) {
	// Write something
	url := "zk://" + strings.Join(Hosts(), ",")
	zk, err := namespace.Dial(context.Background(), url)
	c.Assert(err, IsNil)
	c.Log(zk)
	defer zk.Close()

	k := fmt.Sprintf("/unit-test/registry/template/%d/test", time.Now().Unix())
	p := namespace.NewPath(k)
	v := []byte("test")
	_, err = zk.Put(p, v, false)
	c.Assert(err, IsNil)

	suite.template = "The value is {{get \"zk://" + k + "\"}}!"

	// Generate the auth token required by the server.
	token := auth.NewToken(1*time.Hour).Add("secure", 1)
	header := http.Header{}
	token.SetHeader(header, testutil.PrivateKeyFunc)

	ctx := resource.ContextPutHttpHeader(context.Background(), header)
	applied, err := template.Execute(ctx, fmt.Sprintf("http://localhost:%d/secure", suite.port))
	c.Assert(err, IsNil)
	c.Log(string(applied))
	c.Assert(string(applied), Equals, "The value is test!")
}

func (suite *TestSuiteTemplate) TestList(c *C) {
	// Write something
	url := "zk://" + strings.Join(Hosts(), ",")
	zk, err := namespace.Dial(context.Background(), url)
	c.Assert(err, IsNil)
	c.Log(zk)
	defer zk.Close()

	k := "/unit-test/registry/template/test"
	p := namespace.NewPath(k)

	err = zk.Delete(p)

	// Write new data
	v := []byte("test")
	_, err = zk.Put(p, v, false)
	c.Assert(err, IsNil)

	// write children
	for i := 0; i < 5; i++ {
		cp := p.Sub(fmt.Sprintf("child-%d", i))
		cv := []byte(fmt.Sprintf("value-%d", i))
		_, err = zk.Put(cp, cv, false)
		c.Assert(err, IsNil)
	}

	suite.template = "{{range list \"zk://" + k + "\"}}\n{{.}}{{end}}"

	// Generate the auth token required by the server.
	token := auth.NewToken(1*time.Hour).Add("secure", 1)
	header := http.Header{}
	token.SetHeader(header, testutil.PrivateKeyFunc)

	ctx := resource.ContextPutHttpHeader(context.Background(), header)
	applied, err := template.Execute(ctx, fmt.Sprintf("http://localhost:%d/secure", suite.port))
	c.Assert(err, IsNil)
	c.Log(string(applied))

	l := strings.Split(string(applied), "\n")[1:]
	c.Assert(l, DeepEquals, []string{
		url + "/unit-test/registry/template/test/child-4",
		url + "/unit-test/registry/template/test/child-3",
		url + "/unit-test/registry/template/test/child-2",
		url + "/unit-test/registry/template/test/child-1",
		url + "/unit-test/registry/template/test/child-0",
	})
}

func (suite *TestSuiteTemplate) TestListDeref(c *C) {
	// Write something
	url := "zk://" + strings.Join(Hosts(), ",")
	zk, err := namespace.Dial(context.Background(), url)
	c.Assert(err, IsNil)
	c.Log(zk)
	defer zk.Close()

	k := "/unit-test/registry/template/test-deref"
	p := namespace.NewPath(k)

	err = zk.Delete(p)

	// Write new data
	v := []byte("test")
	_, err = zk.Put(p, v, false)
	c.Assert(err, IsNil)

	// write children
	for i := 0; i < 5; i++ {
		cp := p.Sub(fmt.Sprintf("child-%d", i))
		cv := []byte(fmt.Sprintf("host-%d:%d", i, 8000+i))
		_, err = zk.Put(cp, cv, false)
		c.Assert(err, IsNil)
	}

	// Generate the auth token required by the server.
	token := auth.NewToken(1*time.Hour).Add("secure", 1)
	header := http.Header{}
	token.SetHeader(header, testutil.PrivateKeyFunc)

	// Note we use the get function
	suite.template = "{{range list \"zk://" + k + "\"}}\n{{get .}}{{end}}"
	ctx := resource.ContextPutHttpHeader(context.Background(), header)
	applied, err := template.Execute(ctx, fmt.Sprintf("http://localhost:%d/secure", suite.port))
	c.Assert(err, IsNil)
	c.Log(string(applied))

	l := strings.Split(string(applied), "\n")[1:]
	c.Assert(l, DeepEquals, []string{
		"host-4:8004",
		"host-3:8003",
		"host-2:8002",
		"host-1:8001",
		"host-0:8000",
	})

	// Note we use the get function and a pipe to get the host
	suite.template = "{{range list \"zk://" + k + "\"}}\n{{get . | host}}{{end}}"
	applied, err = template.Execute(ctx, fmt.Sprintf("http://localhost:%d/secure", suite.port))
	c.Assert(err, IsNil)
	c.Log(string(applied))

	l = strings.Split(string(applied), "\n")[1:]
	c.Assert(l, DeepEquals, []string{
		"host-4",
		"host-3",
		"host-2",
		"host-1",
		"host-0",
	})
}

func (suite *TestSuiteTemplate) TestTemplateExecuteWithVarBlock(c *C) {
	// Write something
	url := "zk://" + strings.Join(Hosts(), ",")
	zk, err := namespace.Dial(context.Background(), url)
	c.Assert(err, IsNil)
	c.Log(zk)

	k := "/unit-test/registry/template/test-template-execute-vars/PG_PASS"
	p := namespace.NewPath(k)

	err = zk.Delete(p)

	// Write new data
	v := []byte("password")
	_, err = zk.Put(p, v, false)
	c.Assert(err, IsNil)

	zk.Close()

	// Now use the value in zk in the template

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
user: "{{env "USER"}}"  # Getting the environment variable and use that as value.
password: "{{ get "zk:///unit-test/registry/template/test-template-execute-vars/PG_PASS" }}"
{{end}}

{{define "host"}}
label: appserver
name: myhost
port: {{.port}}  # Here we allow the application to pass in a context that's refereceable.
{{end}}

{
   "image" : "repo/myapp:` + "{{my `app.version`}}-{{my `app.build`}}" + `",
   "host" : "` + "{{my `host.name`}}" + `",{{/* use this for comment in JSON :) */}}
   "user" : "` + "{{my `app.user`}}" + `",
   "password" : "` + "{{my `app.password`}}" + `",
   "port" : "` + "{{my `host.port`}}" + `"
}`

	// The url is a template too.
	url = "http://localhost:{{.port}}/secure"

	data := map[string]interface{}{
		"Name": "test",
		"Age":  20,
		"port": suite.port,
	}
	ctx := template.ContextPutTemplateData(resource.ContextPutHttpHeader(context.Background(), header), data)

	text, err := template.Execute(ctx, url)
	c.Assert(err, IsNil)
	c.Log(string(text))
	obj := make(map[string]string)
	err = json.Unmarshal(text, &obj)
	c.Assert(err, IsNil)
	c.Assert(obj["image"], Equals, "repo/myapp:1.2-1234")
	c.Assert(obj["user"], Equals, os.Getenv("USER"))
	c.Assert(obj["host"], Equals, "myhost")
	c.Assert(obj["port"], Equals, fmt.Sprintf("%d", suite.port))
	c.Assert(obj["password"], Equals, string(v))
}

package zk

import (
	"fmt"
	"github.com/conductant/gohm/pkg/auth"
	"github.com/conductant/gohm/pkg/registry"
	"github.com/conductant/gohm/pkg/resource"
	"github.com/conductant/gohm/pkg/server"
	"github.com/conductant/gohm/pkg/template"
	"github.com/conductant/gohm/pkg/testutil"
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
	"net/http"
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
	zk, err := registry.Dial(context.Background(), url)
	c.Assert(err, IsNil)
	c.Log(zk)
	defer zk.Close()

	k := fmt.Sprintf("/unit-test/registry/template/%d/test", time.Now().Unix())
	p := registry.NewPath(k)
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
	zk, err := registry.Dial(context.Background(), url)
	c.Assert(err, IsNil)
	c.Log(zk)
	defer zk.Close()

	k := "/unit-test/registry/template/test"
	p := registry.NewPath(k)

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
	zk, err := registry.Dial(context.Background(), url)
	c.Assert(err, IsNil)
	c.Log(zk)
	defer zk.Close()

	k := "/unit-test/registry/template/test-deref"
	p := registry.NewPath(k)

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

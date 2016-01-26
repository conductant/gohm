package template

import (
	"github.com/conductant/gohm/pkg/auth"
	"github.com/conductant/gohm/pkg/server"
	"github.com/conductant/gohm/pkg/testutil"
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestSource(t *testing.T) { TestingT(t) }

type TestSuiteSource struct {
	template     string
	stop         chan<- int
	stopped      <-chan error
	templateFile string
}

var templateFileContent = "this is some test template written to disk"

var _ = Suite(&TestSuiteSource{})

func (suite *TestSuiteSource) SetUpSuite(c *C) {
	suite.stop, suite.stopped = server.NewService().
		ListenPort(7891).
		WithAuth(server.Auth{VerifyKeyFunc: testutil.PublicKeyFunc}.Init()).
		Route(server.ServiceMethod{UrlRoute: "/template", HttpMethod: server.GET, AuthScope: server.AuthScopeNone}).
		To(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte(suite.template))
	}).
		Route(server.ServiceMethod{UrlRoute: "/secure", HttpMethod: server.GET, AuthScope: server.AuthScope("secure")}).
		To(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte(suite.template))
	}).Start()

	suite.templateFile = os.TempDir() + "/test-template"
	err := ioutil.WriteFile(suite.templateFile, []byte(templateFileContent), 0644)
	c.Assert(err, IsNil)
}

func (suite *TestSuiteSource) TearDownSuite(c *C) {
	suite.stop <- 1
	<-suite.stopped
	os.Remove(suite.templateFile)
}

func (suite *TestSuiteSource) TestStringSource(c *C) {
	source := "string://{.FirstName}{.LastName}"
	ctx := context.Background()
	t, err := Source(ctx, source)
	c.Assert(err, IsNil)
	c.Assert(string(t), DeepEquals, "{.FirstName}{.LastName}")
}

func (suite *TestSuiteSource) TestFileSource(c *C) {
	source := "file://" + suite.templateFile
	ctx := context.Background()
	t, err := Source(ctx, source)
	c.Assert(err, IsNil)
	c.Assert(string(t), DeepEquals, templateFileContent)
}

func (suite *TestSuiteSource) TestHttpSource(c *C) {
	suite.template = "this-template"
	source := "http://localhost:7891/template"
	ctx := context.Background()
	t, err := Source(ctx, source)
	c.Assert(err, IsNil)
	c.Assert(string(t), DeepEquals, suite.template)
}

func (suite *TestSuiteSource) TestHttpSourceWithToken(c *C) {
	suite.template = "secure-template"
	source := "http://localhost:7891/secure"

	token := auth.NewToken(1*time.Hour).Add("secure", 1)
	ctx := context.Background()
	header := http.Header{}
	token.SetHeader(header, testutil.PrivateKeyFunc)
	ctx = ContextPutHttpHeader(ctx, header)

	t, err := Source(ctx, source)
	c.Assert(err, IsNil)
	c.Assert(string(t), DeepEquals, suite.template)
}

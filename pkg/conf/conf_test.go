package conf

import (
	"github.com/conductant/gohm/pkg/auth"
	"github.com/conductant/gohm/pkg/encoding"
	"github.com/conductant/gohm/pkg/flag"
	"github.com/conductant/gohm/pkg/resource"
	"github.com/conductant/gohm/pkg/server"
	"github.com/conductant/gohm/pkg/testutil"
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestConf(t *testing.T) { TestingT(t) }

type TestSuiteConf struct {
	port         int
	template1    string
	template2    string
	stop         chan<- int
	stopped      <-chan error
	templateFile string
}

var _ = Suite(&TestSuiteConf{port: 7997})

func (suite *TestSuiteConf) SetUpSuite(c *C) {
	suite.stop, suite.stopped = server.NewService().
		ListenPort(suite.port).
		WithAuth(server.Auth{VerifyKeyFunc: testutil.PublicKeyFunc}.Init()).
		Route(server.Endpoint{UrlRoute: "/template1", HttpMethod: server.GET, AuthScope: server.AuthScope("secure")}).
		To(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte(suite.template1))
		c.Log("/template1 called")
	}).
		Route(server.Endpoint{UrlRoute: "/template2", HttpMethod: server.GET, AuthScope: server.AuthScope("secure")}).
		To(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte(suite.template2))
		c.Log("/template2 called")
	}).
		Start()
}

func (suite *TestSuiteConf) TearDownSuite(c *C) {
	suite.stop <- 1
	<-suite.stopped
}

func args(s string) []string {
	return strings.Split(s, " ")
}

type testAddr struct {
	Addr1 string `json:"addr1" yaml:"addr1"`
	Addr2 string `json:"addr2" yaml:"addr2"`
}

type testCommand struct {
	Conf

	String   string   `json:"string" yaml:"string"`
	Int      int      `json:"int" yaml:"int"`
	Template string   `json:"template" yaml:"template" conf:"format"`
	Eval     string   `json:"eval" yaml:"eval" conf:"eval"`
	Addr     testAddr `json:"addr" yaml:"addr"`
}

// Tests overlaying one config on top of another via multiple flags
func (suite *TestSuiteConf) TestConfOverlay(c *C) {

	suite.template1 = `
string: this is a literal string
int: 25
addr:
  addr1: 1234 post
  addr2: sf, ca 94019
`
	suite.template2 = `
string: this is a override literal string
int: 50
template: '{{ format "hello/{{.Domain}}/{{.Service}}" }}'  # { is special char in yaml, must quote the string with '
bar: 27
`

	test := new(testCommand)

	test.OnDoneExecuteLayer = func(cf *Conf, url string, result []byte, err error) {
		c.Log("Processed:", url, ".... result=", string(result))
	}

	test.OnDoneUnmarshalLayer = func(cf *Conf, url string, err error) {
		c.Log("Processed:", url, ".... model=", cf.Model())
	}

	test.OnDoneUnmarshal = func(cf *Conf, obj interface{}) {
		c.Log("Got unmarshaled object = ", cf.model, "===object=>", obj)
	}

	fs := flag.GetFlagSet("test", test)
	err := fs.Parse(args("-conf.url=http://localhost:7997/template1 -conf.url=http://localhost:7997/template2"))

	header := http.Header{}
	token := auth.NewToken(1*time.Hour).Add("secure", 1)
	token.SetHeader(header, testutil.PrivateKeyFunc)

	ctx := resource.ContextPutHttpHeader(context.Background(), header)
	ctx = ContextPutConfigDataType(ctx, encoding.ContentTypeYAML)

	fm := NewFuncMap().Bind("format").To(escapeFunc("format")).Build()
	c.Assert(fm["format"], Not(IsNil))

	err = Configure(ctx, test.Conf, test, fm)
	c.Assert(err, IsNil)

	c.Log("test=", test)

	c.Assert(test.Urls, DeepEquals, []string{"http://localhost:7997/template1", "http://localhost:7997/template2"})
	c.Assert(test.String, Equals, "this is a override literal string")
	c.Assert(test.Int, Equals, 50)
	c.Assert(test.Addr.Addr1, Equals, "1234 post")
	c.Assert(test.Addr.Addr2, Equals, "sf, ca 94019")
	c.Assert(test.Template, Equals, "{{format \"hello/{{.Domain}}/{{.Service}}\"}}")
}

package server

import (
	"github.com/conductant/gohm/pkg/auth"
	"github.com/conductant/gohm/pkg/testutil"
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
	"net/http"
	"testing"
	"time"
)

func TestServer(t *testing.T) { TestingT(t) }

// Server for end to end testing
type testE2EServer struct {
	c *C
}

func testApi(route string, method HttpMethod, scope AuthScope) ServiceMethod {
	return ServiceMethod{UrlRoute: route, HttpMethod: method, AuthScope: scope}
}

var testGetApiFromContext = testApi("/api-from-context", GET, "api-from-context")

func (this *testE2EServer) testGetApiFromContext(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
	this.c.Log("testGetApiFromContext called.")

	sm := ApiForFunc(ctx, this.testGetApiFromContext)
	this.c.Assert(sm, DeepEquals, testGetApiFromContext)

	sm2 := ApiForScope(ctx)
	this.c.Assert(sm2, DeepEquals, testGetApiFromContext)

	resp.Write([]byte("ok"))
}

var testAnonymousFunc = testApi("/anon-get", GET, AuthScopeNone)
var testSimpleGet = testApi("/simple-get", GET, AuthScopeNone)

func (this *testE2EServer) testSimpleGet(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
	this.c.Log("testSimpleGet called.")
	resp.Write([]byte("ok"))
	return
}

type TestSuiteServer struct {
	server     *testE2EServer
	stopServer chan<- bool
}

var _ = Suite(&TestSuiteServer{})

func (suite *TestSuiteServer) SetUpSuite(c *C) {
	suite.server = &testE2EServer{c: c}
	// Set up the server
	suite.stopServer, _ = NewService().
		WithAuth(
		Auth{
			VerifyKeyFunc: testutil.PublicKeyFunc,
		}.Init()).
		ListenPort(7890).
		Route(testSimpleGet).To(suite.server.testSimpleGet).
		Route(testGetApiFromContext).To(suite.server.testGetApiFromContext).
		Route(testAnonymousFunc).To(
		func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
			c.Log("testAnonymousFunc called.")

			sm := ApiForScope(ctx)
			c.Assert(sm, DeepEquals, testAnonymousFunc)

			resp.Write([]byte("ok"))
		}).
		Start()
}

func (suite *TestSuiteServer) TearDownSuite(c *C) {
	suite.stopServer <- true
}

func (suite *TestSuiteServer) TestNoAuthToken(c *C) {
	testutil.Get(c, "http://localhost:7890/simple-get",
		func(resp *http.Response, body []byte) {
			c.Assert(resp.StatusCode, Equals, http.StatusOK)
			c.Assert(body, DeepEquals, []byte("ok"))
		})
	testutil.Get(c, "http://localhost:7890/anon-get",
		func(resp *http.Response, body []byte) {
			c.Assert(resp.StatusCode, Equals, http.StatusOK)
			c.Assert(body, DeepEquals, []byte("ok"))
		})
	testutil.Get(c, "http://localhost:7890/api-from-context",
		func(resp *http.Response, body []byte) {
			c.Assert(resp.StatusCode, Equals, http.StatusUnauthorized)
		})
}

func (suite *TestSuiteServer) TestWithAuthToken(c *C) {
	token := auth.NewToken(1*time.Hour).Add("api-from-context", true).Add("unrelated-scope", 1)

	testutil.GetWithRequest(c, "http://localhost:7890/simple-get",
		func(req *http.Request) *http.Request {
			token.SetHeader(req.Header, testutil.PrivateKeyFunc)
			return req
		},
		func(resp *http.Response, body []byte) {
			c.Assert(resp.StatusCode, Equals, http.StatusOK)
			c.Assert(body, DeepEquals, []byte("ok"))
		})
	testutil.GetWithRequest(c, "http://localhost:7890/api-from-context",
		func(req *http.Request) *http.Request {
			token.SetHeader(req.Header, testutil.PrivateKeyFunc)
			return req
		},
		func(resp *http.Response, body []byte) {
			c.Assert(resp.StatusCode, Equals, http.StatusOK)
			c.Assert(body, DeepEquals, []byte("ok"))
		})

	// Wrong scope not authorized to call this method
	token = auth.NewToken(1*time.Hour).Add("wrong-scope", true)
	testutil.GetWithRequest(c, "http://localhost:7890/api-from-context",
		func(req *http.Request) *http.Request {
			token.SetHeader(req.Header, testutil.PrivateKeyFunc)
			return req
		},
		func(resp *http.Response, body []byte) {
			c.Assert(resp.StatusCode, Equals, http.StatusUnauthorized)
		})
}

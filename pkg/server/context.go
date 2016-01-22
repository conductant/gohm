package server

import (
	"github.com/conductant/gohm/pkg/auth"
	"golang.org/x/net/context"
	"net/http"
	"runtime"
)

type httpRequestKey int
type httpResponseKey int
type httpStreamerKey int
type apiDiscoveryKey int
type engineKey int

var (
	HttpRequestContextKey  httpRequestKey  = 1
	HttpResponseContextKey httpResponseKey = 2
	HttpStreamerContextKey httpStreamerKey = 3
	EngineContextKey       engineKey       = 4
	ApiDiscoveryContextKey apiDiscoveryKey = 4
)

type serverContext struct {
	context.Context
	token  *auth.Token
	req    *http.Request
	resp   http.ResponseWriter
	engine *engine
}

func (this *serverContext) Value(key interface{}) interface{} {
	switch key.(type) {
	case string:
		if this.token != nil && this.token.HasKey(key.(string)) {
			return this.token.Get(key.(string))
		}
	case httpRequestKey:
		return this.req
	case httpResponseKey:
		return this.resp
	case httpStreamerKey, engineKey, apiDiscoveryKey:
		return this.engine
	default:
	}
	return this.Context.Value(key)
}

func HttpRequestFromContext(ctx context.Context) *http.Request {
	if v, ok := (ctx.Value(HttpRequestContextKey)).(*http.Request); ok {
		return v
	}
	return nil
}

func HttpResponseFromContext(ctx context.Context) http.ResponseWriter {
	if v, ok := (ctx.Value(HttpResponseContextKey)).(http.ResponseWriter); ok {
		return v
	}
	return nil
}

func HttpStreamerFromContext(ctx context.Context) Streamer {
	if v, ok := (ctx.Value(HttpStreamerContextKey)).(Streamer); ok {
		return v
	}
	return nil
}

// If the scope of the caller of this function is the scope of a function bound to the ServiceMethod of the api,
// then return that Api according to the binding.  Otherwise, return NotDefined.
func ApiForScope(ctx context.Context) ServiceMethod {
	if engine, ok := (ctx.Value(EngineContextKey)).(*engine); ok {
		if pc, _, _, ok := runtime.Caller(1); ok {
			return engine.apiFromPC(pc)
		}
	}
	return ServiceMethod{}
}

func ApiForFunc(ctx context.Context, f func(context.Context, http.ResponseWriter, *http.Request)) ServiceMethod {
	if engine, ok := (ctx.Value(ApiDiscoveryContextKey)).(ApiDiscovery); ok {
		return engine.ApiForFunc(f)
	}
	return ServiceMethod{}
}

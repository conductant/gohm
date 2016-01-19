package server

import (
	"github.com/conductant/gohm/pkg/auth"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"reflect"
	"runtime"
)

type httpRequestKey int
type httpResponseKey int
type httpStreamerKey int
type engineKey int

var (
	HttpRequestContextKey  httpRequestKey  = 1
	HttpResponseContextKey httpResponseKey = 2
	HttpStreamerContextKey httpStreamerKey = 3
	EngineContextKey       engineKey       = 4
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
		if this.token.HasKey(key.(string)) {
			return this.token.Get(key.(string))
		}
	case httpRequestKey:
		return this.req
	case httpResponseKey:
		return this.resp
	case httpStreamerKey, engineKey:
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
	// Get the caller
	if pc, _, _, ok := runtime.Caller(1); ok {
		callingFunc := runtime.FuncForPC(pc).Name()
		if engine, ok := (ctx.Value(EngineContextKey)).(*engine); ok {

			log.Println("cf=", callingFunc, "funcs=", engine.functionNames)

			if binding, exists := engine.functionNames[callingFunc]; exists {
				return binding.Api
			}
		}
	}
	return ServiceMethod{}
}

func ApiForFunc(ctx context.Context, f func(context.Context, http.ResponseWriter, *http.Request)) ServiceMethod {
	pc := reflect.ValueOf(f).Pointer()
	callingFunc := runtime.FuncForPC(pc).Name()
	if engine, ok := (ctx.Value(EngineContextKey)).(*engine); ok {
		if binding, exists := engine.functionNames[callingFunc]; exists {
			return binding.Api
		}
	}
	return ServiceMethod{}
}

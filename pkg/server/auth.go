package server

import (
	"fmt"
	"github.com/conductant/gohm/pkg/auth"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"time"
)

func ReadPublicKeyPemFile(filename string) func() []byte {
	return func() []byte {
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			panic(fmt.Errorf("Error reading public key pem file %s: %s", filename, err.Error()))
		}
		return bytes
	}
}

func DisableAuth() AuthManager {
	a := Auth{IsAuthOnFunc: AuthOff}
	return a.Init()
}

var (
	AuthOff = func() bool { return false }
	AuthOn  = func() bool { return true }
)

func (data Auth) Init() AuthManager {
	var s Auth = data
	if s.IsAuthOnFunc == nil {
		panic(fmt.Errorf("IsAuthOnFunc not set."))
	}
	if s.VerifyKeyFunc == nil && s.IsAuthOnFunc() {
		panic(fmt.Errorf("Public key file input function not set."))
	}

	if s.GetTimeFunc == nil {
		s.GetTimeFunc = time.Now
	}
	if s.ErrorRenderFunc == nil {
		s.ErrorRenderFunc = DefaultErrorRenderer
	}
	if s.InterceptAuthFunc == nil {
		s.InterceptAuthFunc = func(a bool, ctx context.Context) (bool, context.Context) {
			return a, ctx
		}
	}
	return &s
}

func (this *Auth) IsAuthOn() bool {
	return this.IsAuthOnFunc()
}

func (this *Auth) IsAuthorized(scope AuthScope, req *http.Request) (bool, error) {
	authed := false
	token, err := auth.TokenFromHttp(req, this.VerifyKeyFunc, this.GetTimeFunc)
	if err != nil {
		return false, err
	}
	authed = token.HasKey(string(scope))
	return authed, err
}

func (this *Auth) interceptAuth(authed bool, ctx context.Context) (bool, context.Context) {
	return this.InterceptAuthFunc(authed, ctx)
}

// Best-effort to extract the token from http request.  Note that if auth is not on, token is nil.
func (this *Auth) GetToken(req *http.Request) (*auth.Token, error) {
	if !this.IsAuthOn() {
		return nil, nil
	}
	return auth.TokenFromHttp(req, this.VerifyKeyFunc, this.GetTimeFunc)
}

func (this *Auth) renderError(resp http.ResponseWriter, req *http.Request, message string, code int) {
	this.ErrorRenderFunc(resp, req, message, code)
}

package main

import (
	"flag"
	"fmt"
	"github.com/conductant/gohm/pkg/runtime"
	"github.com/conductant/gohm/pkg/server"
	"github.com/conductant/gohm/pkg/version"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	currentWorkingDir, _ = os.Getwd()
	port                 = flag.Int("port", runtime.EnvInt("EXAMPLE_PORT", 5050), "Server listening port")
	publicKey            = flag.String("auth.public.key", "", "Public key file in PEM format")
)

func MustNot(err error) {
	if err != nil {
		panic(err)
	}
}

func loadPublicKeyFromFile() []byte {
	if *publicKey == "" {
		panic(fmt.Errorf("No public key file specified."))
	}
	bytes, err := ioutil.ReadFile(*publicKey)
	MustNot(err)
	return bytes
}

func startServer(port int) <-chan error {

	reverseProxy := func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
		hostPort := server.GetUrlParameter(req, "host_port")
		server.NewReverseProxy().SetForwardHostPort(hostPort).Strip("/"+hostPort).ServeHTTP(resp, req)
	}

	key := loadPublicKeyFromFile()
	_, stopped := server.NewService().
		WithAuth(
		server.Auth{
			VerifyKeyFunc: func() []byte { return key },
		}.Init()).
		ListenPort(port).
		Route(
		server.Endpoint{
			UrlRoute:    "/{host_port}/{url:.*}",
			HttpMethods: []server.HttpMethod{server.GET, server.POST, server.PUT, server.PATCH, server.DELETE},
			AuthScope:   server.AuthScopeNone,
		}).
		To(reverseProxy).
		Route(
		server.Endpoint{
			UrlRoute:    "/secure/{host_port}/{url:.*}",
			HttpMethods: []server.HttpMethod{server.GET, server.POST, server.PUT, server.PATCH, server.DELETE},
			AuthScope:   server.AuthScope("secure"),
		}).
		To(reverseProxy).
		Start()
	return stopped
}

func main() {

	flag.Parse()

	buildInfo := version.BuildInfo()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", buildInfo.Notice())
		fmt.Fprintf(os.Stderr, "flags:\n")
		flag.PrintDefaults()
	}

	glog.Infoln(buildInfo.Notice())
	buildInfo.HandleFlag()

	stopped := startServer(*port)

	<-stopped
	glog.Infoln("Bye")
}

package main

import (
	"flag"
	"fmt"
	"github.com/conductant/gohm/pkg/runtime"
	"github.com/conductant/gohm/pkg/server"
	"github.com/conductant/gohm/pkg/testutil"
	"github.com/conductant/gohm/pkg/version"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"net/http"
	"os"
)

var (
	currentWorkingDir, _ = os.Getwd()
	port                 = flag.Int("port", runtime.EnvInt("EXAMPLE_PORT", 5050), "Server listening port")
)

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

	proxy := make(chan bool)
	stop, stopped := server.NewService().
		WithAuth(
		server.Auth{
			VerifyKeyFunc: testutil.PublicKeyFunc,
		}.Init()).
		ListenPort(*port).
		Route(
		server.ServiceMethod{
			UrlRoute:   "/info",
			HttpMethod: server.GET,
			AuthScope:  server.AuthScopeNone,
		}).
		To(
		func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
			glog.Infoln("Showing version info.")
			err := server.Marshal(req, buildInfo, resp)
			if err != nil {
				panic(err)
			}
		}).
		Route(
		server.ServiceMethod{
			UrlRoute:   "/quitquitquit",
			HttpMethod: server.POST,
			AuthScope:  server.AuthScope("quitquitquit"),
		}).
		To(
		func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
			glog.Infoln("Stopping the server....")
			proxy <- true
		}).
		OnShutdown(
		func() error {
			glog.Infoln("Executing user custom shutdown...")
			return nil
		}).
		Start()

	// For stopping the server
	go func() {
		<-proxy
		stop <- 1
	}()

	<-stopped
	glog.Infoln("Bye")
}

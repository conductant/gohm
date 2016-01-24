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
	key := loadPublicKeyFromFile()
	// For implementing shutdown
	proxy := make(chan bool)
	stop, stopped := server.NewService().
		WithAuth(
		server.Auth{
			VerifyKeyFunc: func() []byte { return key },
		}.Init()).
		ListenPort(port).
		Route(
		server.ServiceMethod{
			UrlRoute:   "/info",
			HttpMethod: server.GET,
			AuthScope:  server.AuthScopeNone,
		}).
		To(
		func(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
			glog.Infoln("Showing version info.")
			err := server.Marshal(req, version.BuildInfo(), resp)
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

	// Two server cores running on different ports.  Note that quitquitquit will
	// only shutdown the server requested but no the other one.  Kernel signals
	// will shutdown both.
	stopped1 := startServer(*port)
	stopped2 := startServer(*port + 1)

	for range []int{1, 2} {
		select {
		case <-stopped1:
		case <-stopped2:
		}
	}
	glog.Infoln("Bye")
}

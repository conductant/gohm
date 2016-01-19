package server

import (
	"fmt"
	"github.com/golang/glog"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

func Start(port int, endpoint http.Handler, onShutdown func() error, timeout time.Duration) (chan<- bool, <-chan int) {
	var wg sync.WaitGroup
	shutdownc := make(chan io.Closer, 1)
	go HandleSignals(shutdownc, timeout)

	shutdown := onShutdown
	if shutdown == nil {
		shutdown = func() error { return nil }
	}

	glog.Infoln("Starting server")
	apiDone := make(chan bool)
	RunServer(&http.Server{
		Handler: endpoint,
		Addr:    fmt.Sprintf(":%d", port),
	}, apiDone)

	// Here is a list of shutdown hooks to execute when receiving the OS signal
	shutdown_tasks := ShutdownSequence{
		ShutdownHook(func() error {
			err := shutdown()
			wg.Done()
			return err
		}),
		ShutdownHook(func() error {
			apiDone <- true
			glog.Infoln("Stopped endpoint")
			wg.Done()
			return nil
		}),
	}

	// Pid file
	if pid, pidErr := savePidFile(fmt.Sprintf("%d", port)); pidErr == nil {
		shutdown_tasks = append(shutdown_tasks,
			ShutdownHook(func() error {
				os.Remove(pid)
				glog.Infoln("Removed pid file:", pid)
				wg.Done()
				return nil
			}))
	}

	shutdownc <- shutdown_tasks
	count := len(shutdown_tasks)
	wg.Add(count)
	completed := make(chan int)
	go func() {
		wg.Wait()
		completed <- count
	}()
	return apiDone, completed
}

// Runs the http server.  This server offers more control than the standard go's default http server
// in that when a 'true' is sent to the stop channel, the listener is closed to force a clean shutdown.
// The return value is a channel that can be used to block on.  An error is received if server shuts
// down in error; or a nil is received on a clean signalled shutdown.
func RunServer(server *http.Server, stop <-chan bool) <-chan error {
	protocol := "tcp"
	// e.g. 0.0.0.0:80 or :80 or :8080
	if match, _ := regexp.MatchString("[a-zA-Z0-9\\.]*:[0-9]{2,}", server.Addr); !match {
		protocol = "unix"
	}

	listener, err := net.Listen(protocol, server.Addr)
	if err != nil {
		panic(err)
	}

	stoppedChan := make(chan error)
	glog.Infoln("Starting", protocol, "listener at", server.Addr)
	if protocol == "unix" {
		updateDomainSocketPermissions(server.Addr)
	}

	// This will be set to true if a shutdown signal is received. This allows us to detect
	// if the server stop is intentional or due to some error.
	fromSignal := false

	// The main goroutine where the server listens on the network connection
	go func(fromSignal *bool) {
		// Serve will block until an error (e.g. from shutdown, closed connection) occurs.
		err := server.Serve(listener)
		if !*fromSignal {
			glog.Warningln("Warning: server stops due to error", err)
		}
		stoppedChan <- err
	}(&fromSignal)

	// Another goroutine that listens for signal to close the network connection
	// on shutdown.  This will cause the server.Serve() to return.
	go func(fromSignal *bool) {
		select {
		case <-stop:
			listener.Close()
			*fromSignal = true // Intentionally stopped from signal
			return
		}
	}(&fromSignal)
	return stoppedChan
}

func savePidFile(args ...string) (string, error) {
	cmd := filepath.Base(os.Args[0])
	pidFile, err := os.Create(fmt.Sprintf("%s-%s.pid", cmd, strings.Join(args, "-")))
	if err != nil {
		return "", err
	}
	defer pidFile.Close()
	fmt.Fprintf(pidFile, "%d", os.Getpid())
	return pidFile.Name(), nil
}

func updateDomainSocketPermissions(filename string) (err error) {
	_, err = os.Lstat(filename)
	if err != nil {
		return
	}
	return os.Chmod(filename, 0777)
}

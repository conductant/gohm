package server

import (
	"fmt"
	"github.com/golang/glog"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type ShutdownSequence []io.Closer

func ShutdownHook(h func() error) closeWrapper {
	return closeWrapper{run: h}
}

type closeWrapper struct {
	run func() error
}

func (w closeWrapper) Close() error {
	return w.run()
}

// Implements io.Closer
func (s ShutdownSequence) Close() (err error) {
	for _, cl := range s {
		if err1 := cl.Close(); err == nil && err1 != nil {
			err = err1
		}
	}
	return
}

func exitf(pattern string, args ...interface{}) {
	if !strings.HasSuffix(pattern, "\n") {
		pattern = pattern + "\n"
	}
	fmt.Fprintf(os.Stderr, pattern, args...)
	os.Exit(1)
}

func HandleSignals(shutdownc <-chan io.Closer, timeout time.Duration) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	signal.Notify(c, syscall.SIGINT)
	for {
		sig := <-c
		sysSig, ok := sig.(syscall.Signal)
		if !ok {
			glog.Fatal("Not a unix signal")
		}
		switch sysSig {
		case syscall.SIGHUP:
		case syscall.SIGINT:
			glog.Warningln("Got SIGTERM: shutting down")
			donec := make(chan bool)
			go func() {
				cl := <-shutdownc
				if err := cl.Close(); err != nil {
					exitf("Error shutting down: %v", err)
				}
				donec <- true
			}()
			select {
			case <-donec:
				glog.Infoln("Shut down completed.")
				os.Exit(0)
			case <-time.After(timeout):
				exitf("Timeout shutting down. Exiting uncleanly.")
			}
		default:
			glog.Fatal("Received another signal, should not happen.")
		}
	}
}

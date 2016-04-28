package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/conductant/gohm/pkg/command"
	"github.com/conductant/gohm/pkg/fuse"
	"github.com/conductant/gohm/pkg/runtime"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	command.Register("serve", func() (command.Module, command.ErrorHandling) {
		return &server{
			Debug: true,
		}, command.PanicOnError
	})
}

type server struct {
	Mount string `flag:"m,Mount point"`
	Debug bool   `flag:"debug,True to log debug messages."`
}

func (t *server) Help(w io.Writer) {
	fmt.Fprintln(w, "A simple FUSE driver backed by in-memory map.")
}

func (t *server) Run(args []string, w io.Writer) error {
	if t.Debug {
		log.SetLevel(log.DebugLevel)
	}

	stop := make(chan interface{})

	fromKernel := make(chan os.Signal)

	// kill -9 is SIGKILL and is uncatchable.
	signal.Notify(fromKernel, syscall.SIGHUP)  // 1
	signal.Notify(fromKernel, syscall.SIGINT)  // 2
	signal.Notify(fromKernel, syscall.SIGQUIT) // 3
	signal.Notify(fromKernel, syscall.SIGABRT) // 6
	signal.Notify(fromKernel, syscall.SIGTERM) // 15

	go func() {
		<-fromKernel
		log.Infoln("Stopping.")
		stop <- 1
		close(stop)
	}()

	return fuse.Serve(t.Mount, fuse.NewMapBackend(map[string]interface{}{}), stop)
}

func (t *server) Close() error {
	return nil
}

func main() {
	runtime.Main()
}

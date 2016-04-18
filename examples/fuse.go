package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/conductant/gohm/pkg/command"
	"github.com/conductant/gohm/pkg/fuse"
	"github.com/conductant/gohm/pkg/runtime"
	"io"
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
	return fuse.Serve(t.Mount, fuse.NewMapBackend(map[string]interface{}{}), nil)
}

func (t *server) Close() error {
	return nil
}

func main() {
	runtime.Main()
}

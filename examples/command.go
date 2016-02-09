package main

import (
	"flag"
	"fmt"
	"github.com/conductant/gohm/pkg/command"
	"github.com/conductant/gohm/pkg/runtime"
	_ "github.com/conductant/gohm/pkg/version"
	"github.com/golang/glog"
	"io"
)

func init() {
	command.Register("test", func() (command.Verb, command.ErrorHandling) {
		return &test{
			Bool: true,
		}, command.PanicOnError
	})
}

type test struct {
	String string
	Int    int
	Bool   bool
}

func (t *test) Help(w io.Writer) {
	fmt.Fprintln(w, "A simple test module")
}

func (t *test) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&t.String, "s", "default", "A string flag")
	fs.IntVar(&t.Int, "i", 10, "An int flag")
}

func (t *test) Run(args []string, w io.Writer) error {
	glog.Infoln("Initial value:", t)
	glog.Infoln("Got args:", args)
	return nil
}

func (t *test) Close() error {
	return nil
}

func main() {
	runtime.Main()
}

package main

import (
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
			String: "This is the default for the flag",
			Bool:   true,
		}, command.PanicOnError
	})
}

type test struct {
	String string `flag:"s,A string flag"`
	Int    int    `flag:"i,An int flag"`
	Bool   bool
}

func (t *test) Help(w io.Writer) {
	fmt.Fprintln(w, "A simple test module")
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

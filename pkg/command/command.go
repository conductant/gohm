package command

import (
	"flag"
	"fmt"
	cf "github.com/conductant/gohm/pkg/flag"
	"io"
	"os"
	"sort"
	"sync"
)

type ErrorHandling flag.ErrorHandling

const (
	ContinueOnError = ErrorHandling(flag.ContinueOnError)
	PanicOnError    = ErrorHandling(flag.PanicOnError)
	ExitOnError     = ErrorHandling(flag.ExitOnError)
)

var (
	lock     sync.Mutex
	verbs    = map[string]func() (Verb, ErrorHandling){}
	policies = map[string]flag.ErrorHandling{}
)

func Register(verb string, commandFunc func() (Verb, ErrorHandling)) {
	lock.Lock()
	defer lock.Unlock()
	verbs[verb] = commandFunc
	policies[verb] = flag.PanicOnError // default
}

// Verb helps with building command-line applications of the form
// <command> <verb> <flags...>
type Verb interface {
	io.Closer

	Help(io.Writer)
	Run([]string, io.Writer) error
}

func ListVerbs() []string {
	lock.Lock()
	defer lock.Unlock()

	l := []string{}
	for v, _ := range verbs {
		l = append(l, v)
	}
	sort.Strings(l)
	return l
}

func VisitVerbs(f func(string, Verb)) {
	lock.Lock()
	defer lock.Unlock()

	for k, vf := range verbs {
		v, _ := vf()
		f(k, v)
	}
}

func GetVerb(key string) (Verb, bool) {
	lock.Lock()
	defer lock.Unlock()

	cf, has := verbs[key]
	if has {
		v, p := cf()
		policies[key] = flag.ErrorHandling(p)
		return v, true
	}
	return nil, false
}

func Run(key string, verb Verb, args []string, w io.Writer) {
	flagSet := flag.NewFlagSet(key, flag.ContinueOnError)
	flagSet.Usage = func() {
		verb.Help(os.Stderr)
		flagSet.SetOutput(os.Stderr)
		flagSet.PrintDefaults()
	}
	cf.RegisterFlags(key, verb, flagSet)
	err := flagSet.Parse(args)
	if err != nil {
		handle(err, flag.ExitOnError)
	} else {
		handle(verb.Run(flagSet.Args(), w), policies[key])
		handle(verb.Close(), policies[key])
	}
}

func handle(err error, handling flag.ErrorHandling) {
	if err != nil {
		switch handling {
		case flag.ContinueOnError:
		case flag.PanicOnError:
			panic(err)
		case flag.ExitOnError:
			fmt.Fprintf(os.Stderr, "Error:", err)
			os.Exit(2)
		}
	}
}

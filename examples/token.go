package main

import (
	"flag"
	"fmt"
	"github.com/conductant/gohm/pkg/auth"
	"github.com/conductant/gohm/pkg/version"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

var (
	currentWorkingDir, _ = os.Getwd()
	privateKey           = flag.String("token.private.key", "example_key", "Private key file in PEM format")
	ttl                  = flag.Duration("token.ttl", 1*time.Hour, "TTL for token")
	scopes               = new(scopes_t)
)

type scopes_t []string

func (this *scopes_t) String() string {
	return strings.Join(*this, ",")
}
func (this *scopes_t) Set(value string) error {
	*this = append(*this, value)
	return nil
}

func MustNot(err error) {
	if err != nil {
		panic(err)
	}
}

func loadPrivateKeyFromFile() []byte {
	bytes, err := ioutil.ReadFile(*privateKey)
	MustNot(err)
	return bytes
}

func main() {
	flag.Var(scopes, "A", "Auth scope")
	flag.Parse()

	buildInfo := version.BuildInfo()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", buildInfo.Notice())
		fmt.Fprintf(os.Stderr, "flags:\n")
		flag.PrintDefaults()
	}
	buildInfo.HandleFlag()

	token := auth.NewToken(*ttl)
	for _, scope := range *scopes {
		token.Add(scope, 1)
	}
	signed, err := token.SignedString(loadPrivateKeyFromFile)
	MustNot(err)
	fmt.Print(signed)
}

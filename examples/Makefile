all: clean build-example-server build-token-generator

clean:
	-rm server
	-rm token
	-rm proxy

# Common makefile that extracts git version info and generates the LDFLAGS variable.
include ../hack/make/version.mk

KEY_NAME?=example_key

gen-key-pair:
	openssl genrsa -out $(KEY_NAME) 4096
	openssl rsa -pubout -in $(KEY_NAME) -out "$(KEY_NAME).pub"

BUILD_LABEL=example
BUILD_NUMBER=0

build-token-generator:
	${GODEP} go build -ldflags "$(LDFLAGS)" token.go

build-example-server:
	${GODEP} go build -ldflags "$(LDFLAGS)" server.go

build-example-reverse-proxy:
	${GODEP} go build -ldflags "$(LDFLAGS)" proxy.go

build-example-command:
	${GODEP} go build -ldflags "$(LDFLAGS)" command.go

example-run-server: gen-key-pair
	${GODEP} go run -ldflags "$(LDFLAGS)" server.go --logtostderr --port=5050 --auth.public.key=./example_key.pub

example-run-reverse-proxy: gen-key-pair
	${GODEP} go run -ldflags "$(LDFLAGS)" proxy.go --logtostderr --port=5050 --auth.public.key=./example_key.pub

example-run-command-version:
	${GODEP} go run -ldflags "$(LDFLAGS)" command.go version

EXAMPLE_SERVER_PORT?=5050
example-server-quitquitquit:
	curl -i -X POST -H "Authorization: Bearer `./token -token.private.key=example_key -A quitquitquit`" \
	http://localhost:$(EXAMPLE_SERVER_PORT)/quitquitquit



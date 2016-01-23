all: build-example-server build-token-generator

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

example-run-server: gen-key-pair
	${GODEP} go run -ldflags "$(LDFLAGS)" server.go --logtostderr --port=5050 --auth.public.key=./example_key.pub

example-server-quitquitquit:
	curl -i -X POST -H "Authorization: Bearer `./token -token.private.key=example_key -A quitquitquit`" \
	http://localhost:5050/quitquitquit

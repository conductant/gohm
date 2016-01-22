all: build-example-server

# Common makefile that extracts git version info and generates the LDFLAGS variable.
include ../hack/make/version.mk

BUILD_LABEL=example
BUILD_NUMBER=0

build-example-server:
	${GODEP} go build -ldflags "$(LDFLAGS)" server.go

run-example-server:
	${GODEP} go run -ldflags "$(LDFLAGS)" server.go --logtostderr

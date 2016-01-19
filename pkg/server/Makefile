all: test-server

test-server:
	${GODEP} go test ./...  -logtostderr -check.vv -v ${TEST_ARGS}

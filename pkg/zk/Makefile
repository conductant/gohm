all: test-zk

# Local zookeeper running in a container using Docker Tools
ZK_HOSTS?=192.168.99.100:2181


test-zk:
	ZK_HOSTS=$(ZK_HOSTS) \
	${GODEP} go test ./... -check.vv -v -logtostderr ${TEST_ARGS}

test-node:
	ZK_HOSTS=$(ZK_HOSTS) \
	${GODEP} go test ./... -check.f TestSuiteNode -check.vv -v -logtostderr ${TEST_ARGS}

test-registry:
	ZK_HOSTS=$(ZK_HOSTS) \
	${GODEP} go test ./... -check.f TestSuiteRegistry -check.vv -v -logtostderr ${TEST_ARGS}

test-template:
	ZK_HOSTS=$(ZK_HOSTS) \
	${GODEP} go test ./... -check.f TestSuiteTemplate -check.vv -v -logtostderr ${TEST_ARGS}

test-client:
	ZK_HOSTS=$(ZK_HOSTS) \
	${GODEP} go test ./... -check.f TestSuiteClient -check.vv -v -logtostderr ${TEST_ARGS}

test-source:
	ZK_HOSTS=$(ZK_HOSTS) \
	${GODEP} go test ./... -check.f TestSuiteSource -check.vv -v -logtostderr ${TEST_ARGS}

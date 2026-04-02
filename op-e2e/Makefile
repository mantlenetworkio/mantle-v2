# nproc is for linux, sysctl for Mac and then fallback to 4 if neither is available
num_cores := $(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)

# Generally, JUNIT_FILE is set in CI but may be specified to an arbitrary file location to emulate CI locally
# If JUNIT_FILE is set, JSON_LOG_FILE should also be set
ifdef JUNIT_FILE
	go_test = OP_TESTLOG_DISABLE_COLOR=true OP_E2E_DISABLE_PARALLEL=false gotestsum --format=testname --junitfile=$(JUNIT_FILE) --jsonfile=$(JSON_LOG_FILE) -- -failfast
  # Note: -parallel must be set to match the number of cores in the resource class
	go_test_flags = -timeout=60m -parallel=$(num_cores)
else
	go_test = go test
	go_test_flags = -v
endif

test: pre-test test-ws
.PHONY: test

test-external-%: pre-test
	make -C ./external_$*/
	$(go_test) $(go_test_flags) --externalL2 ./external_$*/

test-ws: pre-test
	$(go_test) $(go_test_flags) ./system/... ./e2eutils/... ./opgeth/... ./interop/...
.PHONY: test-ws

test-actions: pre-test
	$(go_test) $(go_test_flags) ./actions/...
.PHONY: test-actions

test-http: pre-test
	OP_E2E_USE_HTTP=true $(go_test) $(go_test_flags) ./system/... ./e2eutils/... ./opgeth/... ./interop/...
.PHONY: test-http

# Run engine-API tests against op-reth.
# Requires OP_E2E_L2_BIN to point to an op-reth binary.
# Example:
#   make test-opreth OP_E2E_L2_BIN=/path/to/op-reth
test-opreth: pre-test
	@if [ -z "$(OP_E2E_L2_BIN)" ]; then \
		echo "ERROR: OP_E2E_L2_BIN must be set to the op-reth binary path" >&2; \
		exit 1; \
	fi
	OP_E2E_L2_BIN=$(OP_E2E_L2_BIN) $(go_test) $(go_test_flags) ./opreth/...
.PHONY: test-opreth

# Run Mantle-specific engine-API tests against op-reth.
# Requires OP_E2E_L2_BIN to point to an op-reth binary.
test-mantleopreth: pre-test
	@if [ -z "$(OP_E2E_L2_BIN)" ]; then \
		echo "ERROR: OP_E2E_L2_BIN must be set to the op-reth binary path" >&2; \
		exit 1; \
	fi
	OP_E2E_L2_BIN=$(OP_E2E_L2_BIN) $(go_test) $(go_test_flags) ./mantleopreth/...
.PHONY: test-mantleopreth

# Run all reth engine-API tests (opreth + mantleopreth).
test-reth: test-opreth test-mantleopreth
.PHONY: test-reth

test-cannon: pre-test
	OP_E2E_CANNON_ENABLED=true $(go_test) $(go_test_flags) ./faultproofs
.PHONY: test-cannon

test-fault-proofs: pre-test
	$(go_test) $(go_test_flags) ./faultproofs
.PHONY: test-faultproofs

cannon-prestates:
	make -C .. cannon-prestates
.PHONY: cannon-prestates


pre-test: pre-test-cannon
.PHONY: pre-test

pre-test-cannon:
	@if [ ! -e ../op-program/bin ]; then \
		make cannon-prestates; \
	fi
.PHONY: pre-test-cannon


clean:
	rm -r ../.devnet
	rm -r ../op-program/bin
.PHONY: clean

fuzz:
	parallel --tag -N1 "go test -run NOTAREALTEST -tags cgo_test -v -fuzztime 10s -fuzz {} ./opgeth" ::: FuzzFjordCostFunction FuzzFastLzGethSolidity FuzzFastLzCgo

ifndef CONTRACT
gen-binding:
	$(error CONTRACT is required, usage: make gen-binding CONTRACT=OPContractsManager)
else
gen-binding:
	cd ../packages/contracts-bedrock && just build
	./scripts/gen-binding.sh $(CONTRACT)
endif
.PHONY: gen-binding

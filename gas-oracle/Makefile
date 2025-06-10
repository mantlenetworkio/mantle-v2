GITCOMMIT := $(shell git rev-parse HEAD)
GITDATE := $(shell git show -s --format='%ct')
VERSION := v0.0.0

LDFLAGSSTRING +=-X main.GitCommit=$(GITCOMMIT)
LDFLAGSSTRING +=-X main.GitDate=$(GITDATE)
LDFLAGSSTRING +=-X main.GitVersion=$(VERSION)
LDFLAGS :=-ldflags "$(LDFLAGSSTRING)"

CONTRACTS_PATH := "../packages/contracts-bedrock/artifacts/contracts"

gas-oracle:
	env GO111MODULE=on go build -v $(LDFLAGS) -o ./bin/gas-oracle ./main.go

clean:
	rm bin/gas-oracle

test:
	go test -v ./...

lint:
	golangci-lint run ./...

abi:
	cat $(CONTRACTS_PATH)/L2/GasPriceOracle.sol/GasPriceOracle.json \
		| jq '{abi,bytecode}' \
		> abis/GasPriceOracle.json

binding: abi
	$(eval temp := $(shell mktemp))

	cat abis/GasPriceOracle.json \
		| jq -r .bytecode > $(temp)

	cat abis/GasPriceOracle.json \
		| jq .abi \
		| abigen --pkg bindings \
		--abi - \
		--out bindings/gaspriceoracle.go \
		--type GasPriceOracle \
		--bin $(temp)

	rm $(temp)

.PHONY: \
	gas-oracle \
	clean \
	test \
	lint \
	abi \
	binding

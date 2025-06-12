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


.PHONY: \
	gas-oracle \
	clean \
	test \
	lint \

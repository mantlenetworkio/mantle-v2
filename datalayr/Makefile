
# Build

binaries: maybe-compile-el  maybe-compile-dl
	$(MAKE) -C ./dl-disperser dl-disperser
	$(MAKE) -C ./dl-node dl-node
	$(MAKE) -C ./dl-retriever dl-retriever

.PHONY: images
images: maybe-compile-el maybe-compile-dl maybe-compile-rollup
	docker compose build

.PHONY: maybe-compile-el
maybe-compile-el:
	./on-change.sh contracts/eignlayr-contracts/src data/eigenlayer_hash "make compile-el"

.PHONY: compile-el
compile-el:
	cd contracts && ./compile.sh compile-el

.PHONY: maybe-compile-dl
maybe-compile-dl:
	./on-change.sh contracts/datalayr-contracts/src data/datalayr_hash "make compile-dl"

.PHONY: compile-dl
compile-dl:
	cd contracts && ./compile.sh compile-dl

.PHONY: compile-and-test-dl
compile-and-test-dl:
	cd contracts && ./compile.sh compile-and-test-dl

# Unit Test
.PHONY: unit-test
unit-test:
	ginkgo common/crypto/bls


# Local Infrastructure

setup-geth:
	cd integration/geth-node && sudo rm -rf data && ./run.sh setup 40 ./secret/private-keys.txt ./secret/geth-account-password

# setup-geth:
# 	cd integration/geth-node && sudo rm -rf data && ./setup.sh 40 ./secret/private-keys.txt

update-host:
	cd integration && ./update-host.sh

deploy-chain: destroy-chain update-host
	cd integration && docker compose up -d
	cd integration && ./wait-for http://0.0.0.0:8000 -- echo "GraphQL up" \
		&& ./wait-for http://0.0.0.0:8545 -- echo "Geth up"

deploy-chain-n-explorer: destroy-chain-n-explorer update-host
	cd integration && docker compose -f docker-compose-with-explorer.yml up -d
	cd integration && ./wait-for http://0.0.0.0:8000 -- echo "GraphQL up" \
		&& ./wait-for http://0.0.0.0:8545 -- echo "Geth up" \
		&& ./wait-for http://0.0.0.0:4000 -- echo "Explorer up"

destroy-chain:
	cd integration \
		&& docker compose down

destroy-chain-n-explorer:
	cd integration \
		&& docker compose -f docker-compose-with-explorer.yml down

# Local Deployment

new-experiment:
	cd integration && ./new-exp.sh

deploy-experiment: binaries new-experiment
	cd integration && go run ./deploy/cmd

deploy-nodes: destroy-nodes
	cd integration && ./bin.sh start-detached

destroy-nodes:
	cd integration && ./bin.sh stop

deploy-nodes-docker: destroy-nodes-docker
	cd integration && ./docker.sh start

destroy-nodes-docker:
	cd integration && ./docker.sh stop

# Integration Test

test-new-infra: binaries deploy-chain new-experiment
	ginkgo ./integration

test-new-contract: binaries new-experiment
	ginkgo ./integration


deploy-rollup:
	cd integration && ./new-exp.sh config.rollup.yaml
	cd integration && go run ./scripts

test-new-nodes: binaries
	ginkgo ./integration -v -- --new-experiment=false


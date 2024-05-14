-include Makefile.overrides

PREFIX = "ghcr.io/layr-labs/datalayr/"
IMAGE ?= subgraph-starter:v0.0.3
FULL_IMAGE_TAG = "${PREFIX}${IMAGE}"

all: image push

image:
	@echo "Building image ${FULL_IMAGE_TAG}"
	docker build -t ${FULL_IMAGE_TAG} ./

push:
	docker push ${FULL_IMAGE_TAG}

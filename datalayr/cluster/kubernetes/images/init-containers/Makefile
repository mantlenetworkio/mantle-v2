-include Makefile.overrides

PREFIX = "ghcr.io/layr-labs/datalayr/"
IMAGE ?= init-container-data-loader:point-3000
FULL_IMAGE_TAG = "${PREFIX}${IMAGE}"

all: image push

image:
	@echo "Building image ${FULL_IMAGE_TAG}"
	docker build -t ${FULL_IMAGE_TAG} ./

push:
	docker push ${FULL_IMAGE_TAG}

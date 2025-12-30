# Install forge (foundry) for building contracts
FROM --platform=$BUILDPLATFORM alpine:3.21 AS builder_foundry

ARG TARGETARCH

RUN apk add --no-cache curl bash git jq

COPY ./op-deployer/pkg/deployer/forge/version.json /tmp/op-deployer-versions.json

# Verify and install forge
RUN <<-EOF
    #!/bin/bash
    set -e

    FOUNDRY_VERSION="$(jq -r '.forge' /tmp/op-deployer-versions.json)"
    FOUNDRY_PLATFORM=alpine_${TARGETARCH}
    FOUNDRY_URL="https://github.com/foundry-rs/foundry/releases/download/${FOUNDRY_VERSION}/foundry_${FOUNDRY_VERSION}_${FOUNDRY_PLATFORM}.tar.gz"
    DOWNLOAD_PATH="/tmp/forge.tar.gz"

    # Download the artifact archive
    curl -fL -o ${DOWNLOAD_PATH} "${FOUNDRY_URL}"

    # Ensure the checksum matches our expectations
    ACTUAL_SHA="$(sha256sum "${DOWNLOAD_PATH}" | awk '{print $1}')"
    EXPECTED_SHA="$(jq -r --arg arch "${FOUNDRY_PLATFORM}" '.checksums[$arch]' /tmp/op-deployer-versions.json)"

    [ "${ACTUAL_SHA}" == "${EXPECTED_SHA}" ] || (echo "Forge checksum mismatch: Expected $EXPECTED_SHA, got $ACTUAL_SHA" && exit 1)

    # Extract the archive
    tar -xzf "${DOWNLOAD_PATH}" -C /usr/local/bin forge
EOF

FROM --platform=$BUILDPLATFORM golang:1.24.10-alpine3.21 AS builder

ARG VERSION=v0.0.0

RUN apk add --no-cache make gcc musl-dev linux-headers git jq bash just curl

# Install mise to manage svm-rs (like CircleCI does)
# Copy mise installation script and mise.toml first
COPY ./ops/scripts/install_mise.sh /tmp/install_mise.sh
COPY ./mise.toml /tmp/mise.toml

RUN chmod +x /tmp/install_mise.sh && \
    /tmp/install_mise.sh && \
    export PATH="$HOME/.local/bin:$PATH" && \
    mise --version

# Copy forge from builder_foundry stage
COPY --from=builder_foundry /usr/local/bin/forge /usr/local/bin/forge

# build op-node with the shared go.mod & go.sum files
COPY . /app

WORKDIR /app

RUN go mod download

# Install svm-rs via mise and pre-install solc versions (like CircleCI install-solc-compilers does)
# mise.toml is needed for the svm-rs alias, and we need to trust it
RUN export PATH="$HOME/.local/bin:$PATH" && \
    mise trust && \
    mise install svm-rs@0.5.19 && \
    ln -sf ~/.local/share/mise/installs/svm-rs/0.5.19/bin/svm /usr/local/bin/svm && \
    for version in 0.8.15 0.8.30 0.5.17; do \
      svm which $version || svm install $version; \
    done

# Build contracts before generating artifacts
WORKDIR /app/packages/contracts-bedrock
RUN just build-no-tests

# Generate forge-artifacts for op-deployer (required for embedded.go)
WORKDIR /app/op-deployer
RUN just copy-contract-artifacts

ARG TARGETOS TARGETARCH

WORKDIR /app/op-node

RUN make op-node VERSION="$VERSION" GOOS=$TARGETOS GOARCH=$TARGETARCH

FROM alpine:3.21

COPY --from=builder /app/op-node/bin/op-node /usr/local/bin

CMD ["op-node"]

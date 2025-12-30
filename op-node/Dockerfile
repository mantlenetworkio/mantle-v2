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

FROM --platform=$BUILDPLATFORM golang:1.24.10-bookworm AS builder

ARG VERSION=v0.0.0

# Install dependencies (Debian-based, similar to CircleCI's cimg/base:2024.01)
RUN apt-get update && apt-get install -y --no-install-recommends \
    make gcc g++ git jq bash curl ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Install just
RUN curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to /usr/local/bin

# Copy forge from builder_foundry stage
COPY --from=builder_foundry /usr/local/bin/forge /usr/local/bin/forge

# build op-node with the shared go.mod & go.sum files
COPY . /app

WORKDIR /app

# Manually download and install solc binaries from binaries.soliditylang.org (amd64) or hardhat mirror (arm64)
# Using Debian-based image (like CircleCI's cimg/base) ensures glibc compatibility for solc binaries
ARG TARGETARCH
RUN <<-EOF
    #!/bin/bash
    set -e

    # Determine the list.json URL and base download URL based on architecture
    if [ "${TARGETARCH}" = "amd64" ]; then
        LIST_URL="https://binaries.soliditylang.org/linux-amd64/list.json"
        BASE_URL="https://binaries.soliditylang.org/linux-amd64"
    elif [ "${TARGETARCH}" = "arm64" ]; then
        LIST_URL="https://solc-linux-arm64-mirror.hardhat.org/linux/aarch64/list.json"
        BASE_URL="https://solc-linux-arm64-mirror.hardhat.org/linux/aarch64"
    else
        echo "Unsupported architecture: ${TARGETARCH}"
        exit 1
    fi

    echo "Fetching solc version list from ${LIST_URL}"

    # Download the list.json
    curl -fL -o /tmp/solc-list.json "${LIST_URL}" || {
        echo "Failed to download solc version list from ${LIST_URL}"
        exit 1
    }

    # Create svm directory structure
    mkdir -p /root/.svm

    # Download and install each solc version
    for version in 0.8.15 0.8.30 0.5.17; do
        VERSION_DIR="/root/.svm/${version}"
        mkdir -p "${VERSION_DIR}"
        SOLC_BINARY="${VERSION_DIR}/solc-${version}"

        # Get the path from list.json
        # For amd64: "solc-linux-amd64-v0.8.15+commit.e87f0b89"
        # For arm64: "solc-v0.8.15"
        SOLC_PATH=$(jq -r --arg v "${version}" '.releases[$v] // empty' /tmp/solc-list.json)

        if [ -z "${SOLC_PATH}" ]; then
            echo "Failed to find solc ${version} in list.json"
            exit 1
        fi

        # Download from the appropriate base URL
        SOLC_URL="${BASE_URL}/${SOLC_PATH}"

        echo "Downloading solc ${version} from ${SOLC_URL}"
        if ! curl -fL -o "${SOLC_BINARY}" "${SOLC_URL}"; then
            echo "Failed to download solc ${version} from ${SOLC_URL}"
            exit 1
        fi

        chmod +x "${SOLC_BINARY}"
        echo "Installed solc ${version} to ${SOLC_BINARY}"

        # Verify the binary works
        if ! "${SOLC_BINARY}" --version > /dev/null 2>&1; then
            echo "Warning: solc ${version} binary may not be executable or compatible"
        fi
    done

    # Clean up
    rm -f /tmp/solc-list.json
EOF

# Build contracts before generating artifacts
WORKDIR /app/packages/contracts-bedrock
RUN just build-no-tests

RUN go mod download

# Generate forge-artifacts for op-deployer (required for embedded.go)
WORKDIR /app/op-deployer
RUN just copy-contract-artifacts

ARG TARGETOS

WORKDIR /app/op-node

RUN make op-node VERSION="$VERSION" GOOS=$TARGETOS GOARCH=$TARGETARCH

FROM alpine:3.21

COPY --from=builder /app/op-node/bin/op-node /usr/local/bin

CMD ["op-node"]

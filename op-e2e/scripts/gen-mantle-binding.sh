#!/usr/bin/env bash

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "${REPO_ROOT}"

# Mantle-specific contracts that differ from upstream OP
CONTRACTS=(
  GasPriceOracle
  L1Block
  L2OutputOracle
  OptimismPortal
  SystemConfig
)

if [[ $# -ge 1 ]]; then
  CONTRACTS=("$@")
fi

OUTPUT_DIR="op-e2e/mantlebindings"
mkdir -p "${OUTPUT_DIR}"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT

for CONTRACT in "${CONTRACTS[@]}"; do
  ARTIFACT_DIR="packages/contracts-bedrock/forge-artifacts/${CONTRACT}.sol"
  ARTIFACT_PATH="${ARTIFACT_DIR}/${CONTRACT}.json"

  if [[ ! -f "${ARTIFACT_PATH}" ]]; then
    echo "error: artifact not found at ${ARTIFACT_PATH}. Run the contracts build first." >&2
    exit 1
  fi

  OUTPUT_BASENAME="$(echo "${CONTRACT}" | tr '[:upper:]' '[:lower:]')"
  OUTPUT_PATH="${OUTPUT_DIR}/${OUTPUT_BASENAME}.go"

  ABI_PATH="${TMPDIR}/${CONTRACT}.abi.json"
  BIN_PATH="${TMPDIR}/${CONTRACT}.bin"

  jq '.abi' "${ARTIFACT_PATH}" > "${ABI_PATH}"
  jq -r '.bytecode.object' "${ARTIFACT_PATH}" > "${BIN_PATH}"

  echo "generating ${OUTPUT_PATH} ..."
  abigen --pkg mantlebindings --type "${CONTRACT}" --abi "${ABI_PATH}" --bin "${BIN_PATH}" --out "${OUTPUT_PATH}"
  gofmt -w "${OUTPUT_PATH}"
done

echo "done."

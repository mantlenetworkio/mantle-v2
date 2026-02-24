#!/usr/bin/env bash
set -euo pipefail

verify_flag=""
if [ -n "${DEPLOY_VERIFY:-}" ]; then
  verify_flag="--verify"
fi

echo "> Deploying contracts"
forge script -vvv scripts/deploy/Deploy.s.sol:Deploy --rpc-url "$DEPLOY_ETH_RPC_URL" --broadcast --private-key "$DEPLOY_PRIVATE_KEY" $verify_flag

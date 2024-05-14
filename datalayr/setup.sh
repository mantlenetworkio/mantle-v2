#!/bin/bash


# Clone the submodules
git submodule update --init --recursive

# Install subgraph
cd subgraph && yarn && cd ..

# Create necessary .env files
echo '
# EIGENLAYER DEPLOYMENT
RPC_URL=http://18.216.241.146:31000
PRIVATE_KEY=0x54b9e633e01bc3a105561e25a473671fbb6ba21ff5842ddc9b62760c6f508f71

# CHAIN CONFIG
GETH_CHAINID=40525
GETH_UNLOCK_ADDRESS=3aa273f6c6df3a8498ebdcdd241a2575e08cde64
' > ./cluster/.env

echo '
GETH_CHAINID=40525
GETH_UNLOCK_ADDRESS=3aa273f6c6df3a8498ebdcdd241a2575e08cde64
EXPERIMENT=""
HOST_IP=0.0.0.0' > ./integration/.env

# Clone g1, g2 points from s3
mkdir -p ./integration/data/kzg
wget --no-check-certificate --no-proxy https://datalayr-testnet.s3.amazonaws.com/g1.point.3000 -O ./integration/data/kzg/g1.point
wget --no-check-certificate --no-proxy https://datalayr-testnet.s3.amazonaws.com/g2.point.3000 -O ./integration/data/kzg/g2.point

# Run setup script for geth-node
make setup-geth
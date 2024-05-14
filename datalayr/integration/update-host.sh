#!/usr/bin/env bash

if ! which docker 2>&1 > /dev/null; then
    echo "Please install 'docker' first"
    exit 1
fi

if ! which jq 2>&1 > /dev/null; then
    echo "Please install 'jq' first"
    exit 1
fi

# Create the graph-node container
docker compose up --no-start graph-node

# Start graph-node so we can inspect it
docker compose start graph-node

# Identify the container ID
CONTAINER_ID=$(docker container ls | grep graph-node | cut -d' ' -f1)

# Inspect the container to identify the host IP address
HOST_IP=$(docker inspect "$CONTAINER_ID" | jq -r .[0].NetworkSettings.Networks[].Gateway)

echo "Host IP: $HOST_IP"

# Update .env
sed -i"back" "s/HOST_IP=.*/HOST_IP=\"${HOST_IP}\"/" .env && rm .envback

function stop_graph_node {
    # Ensure graph-node is stopped
    docker compose stop graph-node
}

trap stop_graph_node EXIT

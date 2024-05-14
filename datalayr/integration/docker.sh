#!/bin/bash

function start {
    source .env
    docker compose -f data/${EXPERIMENT}/docker-compose.yml up -d

    waiters=""
    for FILE in $(ls data/${EXPERIMENT}/envs/dis*.env); do 
        source $FILE
        ./wait-for 0.0.0.0:${DL_DISPERSER_GRPC_PORT} -- echo "Disperser up" &
        waiters="$waiters $!"
    done

    for FILE in $(ls data/${EXPERIMENT}/envs/dln*.env); do 
        source $FILE
        ./wait-for 0.0.0.0:${DL_NODE_GRPC_PORT} -- echo "Node up" &
        waiters="$waiters $!"
    done

    for waiter in $waiters; do 
        wait $waiter
    done

}

function stop {
    source .env
    docker compose -f data/${EXPERIMENT}/docker-compose.yml down
}

function logs {
    source .env
    docker compose -f data/${EXPERIMENT}/docker-compose.yml logs -f
}


case "$1" in
    help)
        cat <<-EOF
        Docker experiment tool
EOF
        ;;
    start)
        start ${@:2} ;;    
    stop)
        stop ${@:2} ;;
    logs)
        logs ${@:2} ;;
    *)
esac
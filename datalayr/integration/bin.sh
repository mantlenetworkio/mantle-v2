#!/bin/bash

pids=""
function kill_processes {
    echo "STOP"
    for pid in $pids; do
        echo "killing process $pid"
        kill $pid
    done
}

function start_trap {

    trap kill_processes SIGINT

    source .env
    for FILE in $(ls data/${EXPERIMENT}/envs/dis*.env); do
        set -a
        source $FILE
        set +a
        ../dl-disperser/bin/dl-disperser &

        pid="$!"
        pids="$pids $pid"
    done

    for FILE in $(ls data/${EXPERIMENT}/envs/dln*.env); do
        set -a
        source $FILE
        set +a
        ../dl-node/bin/dl-node &

        pid="$!"
        pids="$pids $pid"
    done

    for FILE in $(ls data/${EXPERIMENT}/envs/ret*.env); do
        set -a
        source $FILE
        set +a
        ../dl-retriever/bin/dl-retriever &

        pid="$!"
        pids="$pids $pid"

        ./wait-for 0.0.0.0:${DL_RETRIEVER_GRPC_PORT} -- echo "Retriever up"
    done

    for pid in $pids; do
        wait $pid
    done

}

function start_detached {

    source .env

    pids=""
    waiters=""
    pid_file="data/${EXPERIMENT}/pids"

    if [[ -f "$pid_file" ]]; then
        echo "Processes still running. Run ./bin.sh stop"
        return
    fi

    mkdir -p data/${EXPERIMENT}/logs

    for FILE in $(ls data/${EXPERIMENT}/envs/dis*.env); do
        set -a
        source $FILE
        set +a
        id=$(basename $FILE | tr -d -c 0-9)
        ../dl-disperser/bin/dl-disperser > data/${EXPERIMENT}/logs/dis${id}.log 2>&1 &

        pid="$!"
        pids="$pids $pid"

        ./wait-for 0.0.0.0:${DL_DISPERSER_GRPC_PORT} -- echo "Disperser up" &
        waiters="$waiters $!"
    done

    for FILE in $(ls data/${EXPERIMENT}/envs/dln*.env); do
        set -a
        source $FILE
        set +a
        id=$(basename $FILE | tr -d -c 0-9)
        ../dl-node/bin/dl-node > data/${EXPERIMENT}/logs/dln${id}.log 2>&1 &

        pid="$!"
        pids="$pids $pid"

        ./wait-for 0.0.0.0:${DL_NODE_GRPC_PORT} -- echo "Node up" &
        waiters="$waiters $!"
    done

    for FILE in $(ls data/${EXPERIMENT}/envs/ret*.env); do
        set -a
        source $FILE
        set +a
        id=$(basename $FILE | tr -d -c 0-9)
        ../dl-retriever/bin/dl-retriever > data/${EXPERIMENT}/logs/ret${id}.log 2>&1 &

        pid="$!"
        pids="$pids $pid"

        ./wait-for 0.0.0.0:${DL_RETRIEVER_GRPC_PORT} -- echo "Retriever up" &
        waiters="$waiters $!"
    done

    echo $pids > $pid_file

    for waiter in $waiters; do
        wait $waiter
    done
}


function stop_detached {

    source .env

    pid_file="data/${EXPERIMENT}/pids"

    # pid_file="data/pids"
    pids=$(cat $pid_file)

    kill_processes

    rm -f $pid_file
}



case "$1" in
    help)
        cat <<-EOF
        Binary experiment tool
EOF
        ;;
    start)
        start_trap ${@:2} ;;
    start-detached)
        start_detached ${@:2} ;;
    stop)
        stop_detached ${@:2} ;;
    *)
esac

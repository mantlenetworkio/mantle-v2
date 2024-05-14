#!/bin/bash

function start_cluster {
	if [ "$#" -ne 2 ]; then
		echo "need number dl-node and dl-disperser"
		exit 0
	fi
	docker compose -f docker-compose-cluster.yml up -d
}

function stop_local {
	docker compose -f docker-compose-cluster.yml down
}

case "$1" in
    help)
        cat <<-EOF
        prometheus grafana tool
EOF
        ;;
    start-cluster)
        start_cluster ${@:2} ;;    
    stop)
        stop_cluster ${@:2} ;;
    *)
esac

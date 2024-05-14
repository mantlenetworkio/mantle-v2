#!/bin/bash

pids=""
function kill_processes {
    echo "STOP"
    for pid in $pids; do
        echo "killing process $pid"
        kill $pid
    done
}

trap kill_processes SIGINT

if [ "$#" -ne 1 ]; then
	echo "need log dir"
	exit 0
fi

log_dir=$1

num_disperser=4
nodesYaml=nodes.yaml
kubectl get nodes -o yaml > $nodesYaml

ext_ips=""
for ((id=0; id<$num_disperser; id++)); do
	ext_ip=$(cat $nodesYaml | yq ".items[$id] | .status.addresses | .[] | select(.type == \"ExternalIP\") | .address")
	int_ip=$(cat $nodesYaml | yq ".items[$id] | .status.addresses | .[] | select(.type == \"InternalIP\") | .address")

	# if use kind, a local k8s simulator, ext ip is empty, then map to itself
	if [ -z ${ext_ip} ]; then
		ext_ips="$ext_ips ${int_ip}"
	else
		ext_ips="$ext_ips ${ext_ip}"
	fi
done

echo $ext_ips

i=1
for ip in $ext_ips; do
	port=$((32000 + $i))
	echo $ip $i $port
	./bin/traffic-gen --hostname=$ip --grpc-port=$port --timeout=500s --store-duration=1 --data-size 1000000 --live-threshold=0.9 --adv-threshold=0.4 --idle-period=0 --idle-period-std=0 --number=1 --log.level-std=info --log.level-file=info --log.path ./${log_dir}/log$i &
	pid="$!"
	pids="$pids $pid"

	((i+=2))
done

#./bin/traffic-gen --hostname=$ip1 --grpc-port=32000 --timeout=30s --store-duration=1 --data-size 3000000 --live-threshold=0.9 --adv-threshold=0.5 --idle-period=0 --idle-period-std=0 --number=1 --log.level-std=info --log.level-file=info --log.path ./log1 &
#pid="$!"
#pids="$pids $pid"

echo $pids > pids

for pid in $pids; do
		wait $pid
done

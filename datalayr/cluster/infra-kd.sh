#!/bin/bash

set -e

# source .env

# Deploy chain infrastructure

# Deploys sealer and relay using dev credentials
function deploy_dev_chain {
	# Generate sealer kustomization
	rm -rf kustomize/overlay/${ENVIRONMENT}/chain
	mkdir -p kustomize/overlay/${ENVIRONMENT}/chain

	pushd kustomize/overlay/${ENVIRONMENT}/chain
		kustomize create --resources ../../../base/chain/,../../../base/chain/config/dev-chain
	popd

	# Deploy sealer components first
	kustomize build ./kustomize/overlay/${ENVIRONMENT}/chain | kubectl apply -f -
	echo "sealer resources deployed. Wait for getting ready..."
	sleep 60

	# Generate sealer config for the relay nodes
	rm -rf kustomize/overlay/${ENVIRONMENT}/config/geth
	mkdir -p kustomize/overlay/${ENVIRONMENT}/config/geth

	pushd kustomize/overlay/${ENVIRONMENT}/config/geth
		kustomize create --namespace relay --labels app:relay
		enode=$(kubectl exec -i geth-0 -- sh -c 'geth attach --exec "admin.nodeInfo.enode" /geth-data/geth.ipc')

		parsedEnode=${enode:1}
		parsedEnode="${parsedEnode%@*}@geth-svc.default.svc.cluster.local:30303"

		kustomize edit add configmap sealer-config \
			--from-literal SEALER_ENODE_ENV=$parsedEnode
		echo "wrote sealer enode ($parsedEnode) under kustomize/overlay/${ENVIRONMENT}/config/geth "
	popd
}

# Deploys sealer and relay using dev credentials
function deploy_prod_chain {
	# Generate sealer kustomization
	rm -rf kustomize/overlay/${ENVIRONMENT}/chain
	mkdir -p kustomize/overlay/${ENVIRONMENT}/chain

	pushd kustomize/overlay/${ENVIRONMENT}/chain
		kustomize create --resources ../../../base/chain/,../../../base/chain/config/prod-chain
	popd

	# Deploy sealer components first
	kustomize build ./kustomize/overlay/${ENVIRONMENT}/chain | kubectl apply -f -
	echo "sealer resources deployed. Wait for getting ready..."
	sleep 60

	# Generate sealer config for the relay nodes
	rm -rf kustomize/overlay/${ENVIRONMENT}/config/geth
	mkdir -p kustomize/overlay/${ENVIRONMENT}/config/geth

	pushd kustomize/overlay/${ENVIRONMENT}/config/geth
		kustomize create --namespace relay --labels app:relay
		enode=$(kubectl exec -i geth-0 -- sh -c 'geth attach --exec "admin.nodeInfo.enode" /geth-data/geth.ipc')

		parsedEnode=${enode:1}
		parsedEnode="${parsedEnode%@*}@geth-svc.default.svc.cluster.local:30303"

		kustomize edit add configmap sealer-config \
			--from-literal SEALER_ENODE_ENV=$parsedEnode
		echo "wrote sealer enode ($parsedEnode) under kustomize/overlay/${ENVIRONMENT}/config/geth "
	popd
}

function run_dev_relay {
	if [ "$#" -ne 2 ]; then
		echo "need number of relay"
		echo "need number of eigenda relay"
		exit 0
	fi
	num_relay=$1
	num_eigenda_relay=$2

	rm -rf kustomize/overlay/${ENVIRONMENT}/relay
	mkdir -p kustomize/overlay/${ENVIRONMENT}/relay

	pushd kustomize/overlay/${ENVIRONMENT}/relay
		kustomize create \
			--resources=../../../base/relay,../../../base/chain/config/dev-chain,../config/geth  \
			--namespace=relay
		kustomize edit set replicas relay=${num_relay}
	popd

	# Deploy relay component
	kustomize build ./kustomize/overlay/${ENVIRONMENT}/relay | kubectl apply -f -
	echo "Applied relay services. Wait for relay to start "
	kubectl -n relay rollout status statefulset relay

	# Generate the relay services
	gen_relay_service

	# Gen the relay configmap for the EigenDA components
	gen_chain_relay_cm ${num_eigenda_relay}

	### Chain is deployed at this point, deploy eigenlayer contracts now
}

function run_prod_relay {
	if [ "$#" -ne 2 ]; then
		echo "need number of relay"
		echo "need number of eigenda relay"
		exit 0
	fi
	num_relay=$1
	num_eigenda_relay=$2

	rm -rf kustomize/overlay/${ENVIRONMENT}/relay
	mkdir -p kustomize/overlay/${ENVIRONMENT}/relay

	pushd kustomize/overlay/${ENVIRONMENT}/relay
		kustomize create \
			--resources=../../../base/relay,../../../base/chain/config/prod-chain,,../config/geth \
			--namespace=relay
		kustomize edit set replicas relay=${num_relay}
	popd

	# Deploy relay component
	kustomize build ./kustomize/overlay/${ENVIRONMENT}/relay | kubectl apply -f -
	echo "Applied relay services. Wait for relay to start "
	kubectl -n relay rollout status statefulset relay

	# Generate the relay services
	gen_relay_service

	# Gen the relay configmap for the EigenDA components
	gen_chain_relay_cm ${num_eigenda_relay}

	### Chain is deployed at this point, deploy eigenlayer contracts now
}

function gen_relay_service {
	outPath="kustomize/overlay/${ENVIRONMENT}/relay/tmp"
	mkdir -p $outPath

	podsYaml="${outPath}/relaysDeployState.yaml"

	kubectl get pods -l app=relay -n relay -o yaml > $podsYaml

	num_svc=$(cat $podsYaml | yq '.items' | yq length)

	port=31500
	for ((id=0; id<$num_svc; id++)); do
		p2p_port=${port}
		geth_http_port=$(($port+1))
		((port+=2))

		echo "p2p_port $p2p_port"
		echo "http_port $geth_http_port"

		hostname=$(cat $podsYaml | yq ".items[$id].spec.hostname")
		# hostname is identical configMap name

		# yq inplace udpate
		svc_name="relay-svc-${id}"
		yq  "
			.metadata.name = \"$svc_name\" |
			.spec.selector.\"statefulset.kubernetes.io/pod-name\" = \"$hostname\" |
			.spec.ports.[] |= (
				with(select(.name == \"p2p\" );
					.port = ${p2p_port} |
					.nodePort = ${p2p_port} |
					.targetPort = ${p2p_port}
				) |
				with(select(.name == \"http\" );
					.port = 8545 |
					.nodePort = ${geth_http_port} |
					.targetPort = 8545
				)
		 	)" kustomize/base/relay/service.yaml | kubectl apply -f -
	done

	rm -rf kustomize/overlay/${ENVIRONMENT}/relay/tmp
}

function label_relay_services {
	# First 8 relays are dedicated to the testnets
	kubectl label pod relay-8 --overwrite service=rpc -n relay
	kubectl label pod relay-9 --overwrite service=blocks -n relay
	kubectl label pod relay-10 --overwrite service=blobs -n relay
}

function gen_chain_relay_cm {
	if [ "$#" -ne 1 ]; then
		echo "need number of relay node for eigenda cluster"
		exit 0
	fi

	# get svc node
	num_relay=$(kubectl get svc -n relay -l app=relay --no-headers -o custom-columns=":metadata.name" | wc -l)
	echo "num relay" $num_relay
	# although any ext ip is fine, but offload to all nodes
	# assume number node is always greater than or equal to number relay

	rm -rf kustomize/overlay/${ENVIRONMENT}/config/relay
	mkdir -p kustomize/overlay/${ENVIRONMENT}/config/relay

	pushd kustomize/overlay/${ENVIRONMENT}/config/relay
		mkdir -p "./tmp"
		nodes_yaml="./tmp/relay-nodes.yaml"
		kubectl get nodes -o yaml > $nodes_yaml

		kustomize create --labels app:relay
		for ((id=0; id<$num_relay; id++)); do
			relay_yaml="./tmp/relays-svc-$id.yaml"

			kubectl get svc "relay-svc-${id}" -n relay -o yaml > $relay_yaml

			ext_ip=$(cat $nodes_yaml | yq ".items[$id] | .status.addresses | .[] | select(.type == \"ExternalIP\") | .address")

			http_node_port=$(cat $relay_yaml | yq ".spec.ports.[] | select(.name == \"http\") | .nodePort")
			p2p_node_port=$(cat $relay_yaml | yq ".spec.ports.[] | select(.name == \"p2p\") | .nodePort")

			echo "HTTP_NODE_PORT_${id}=${ext_ip}:${http_node_port}" >> chain-relay-node-port.env
			echo "P2P_NODE_PORT_${id}=${ext_ip}:${p2p_node_port}" >> chain-relay-node-port.env
		done

		kustomize edit add configmap chain-relay-node-port --from-env-file chain-relay-node-port.env
	popd

	rm -rf kustomize/overlay/${ENVIRONMENT}/config/relay/tmp
}

# This only needs to be run once and is left here for reference. To deploy
# blockscout just copy paste `kustomize/overlay/devnet-frontend/blockscout`
# and change the appropriate parameters.
function deploy_blockscout {
	if [ "$#" -ne 4 ]; then
		echo "need clientId, clientSecret, github_users file name, rpc_url"
		exit 0
	fi

	rm -rf kustomize/overlay/${ENVIRONMENT}/blockscout
	mkdir -p kustomize/overlay/${ENVIRONMENT}/blockscout
	mkdir -p kustomize/overlay/${ENVIRONMENT}/blockscout/config

	# Generate oauth2 config
	gen_oauth2_config blockscout $1 $2 $3

	# Patch the rpc url
	rpc_url=$4
	mkdir -p kustomize/overlay/${ENVIRONMENT}/blockscout/patches
	cat <<EOF >kustomize/overlay/${ENVIRONMENT}/blockscout/patches/rpc.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: blockscout
spec:
  template:
    spec:
      containers:
        - name: blockscout-node
          env:
            - name: ETHEREUM_JSONRPC_HTTP_URL
              value: ${rpc_url}
EOF

	pushd kustomize/overlay/${ENVIRONMENT}/blockscout
		kustomize create --resources ../../../base/blockscout/
		kustomize edit add configmap blockscout-oauth2 --from-env-file config/blockscout-oauth2.env
		kustomize edit add patch StatefulSet --path patches/rpc.yaml
	popd

	kustomize build ./kustomize/overlay/${ENVIRONMENT}/blockscout | kubectl apply -f -
}

function gen_oauth2_config {
	if [ "$#" -ne 4 ]; then
		echo "need app name, clientId, clientSecret, github_users file name"
		exit 0
	fi

	app=$1
	clientId=$2
	clientSecret=$3
	users=$(cat $4 | jq -r 'join(",")')

	configMapOutput="kustomize/overlay/${ENVIRONMENT}/${app}/config/${app}-oauth2.env"

	set +e
		kubectl get cm "${app}-oauth2"
		failed=$?
	set -e

	if [ $failed == 0 ]; then
		cookieSecret=$(kubectl get cm "${app}-oauth2" --output=yaml | yq .data.OAUTH2_PROXY_COOKIE_SECRET)
	else
		cookieSecret=$(dd if=/dev/urandom bs=32 count=1 2>/dev/null | base64 | tr -d -- '\n' | tr -- '+/' '-_'; echo)
	fi

	echo "OAUTH2_PROXY_GITHUB_USERS=${users}" > $configMapOutput
	echo "OAUTH2_PROXY_CLIENT_ID=${clientId}" >> $configMapOutput
	echo "OAUTH2_PROXY_CLIENT_SECRET=${clientSecret}" >> $configMapOutput
	echo "OAUTH2_PROXY_COOKIE_SECRET=${cookieSecret}" >> $configMapOutput
}

function update_oauth2_users {
	gen_oauth2_config
	kubectl delete po -l oauth2
}

### MISC ###

function locate_pod_by_port {
	if [ "$#" -ne 1 ]; then
		echo "need port-number"
		exit 0
	fi

	port_number=$1
	echo "pod number:"
	kubectl get svc | grep "$port_number" | awk '{ print $1 }' | awk -F- '{print $NF}'
}

function print_var {
    prefix=$1
    name=$2
    value=$3
    echo "${prefix}_${name}=${value}"
}


function delete_all {
	set +e
	kustomize build ./kustomize/overlay/${ENVIRONMENT}/sealer | kubectl delete -f -
	kustomize build ./kustomize/overlay/${ENVIRONMENT}/relay | kubectl delete -f -
	# pvc needs explicit deletion
	kubectl delete pvc --selector 'owner in (geth)'
	kubectl delete pvc --selector 'owner in (relay)' -n relay

	set -e
}

function check_cluster_name {
	if [ "$#" -ne 1 ]; then
		echo "Need cluster name"
		exit 0
	fi

	cluster_path=$1
	if [ ! -d "${cluster_path}" ]; then
		echo "cluster directory not exists. $cluster_path"
		exit 0
	fi
}


if [ "$1" == "help" ] || [ "$2" == "help" ]  ; then
	cat <<-EOF
	# Deploy chain infra
    0.  deploy-dev-chain or deploy-prod-chain
    1.  run-dev-relay or run-prod-relay

    # Analysis
    1.  locate-pod-by-port

    # Clean up
    1.  delete-all (sts,svc,cm,pvc,secret)
EOF
	exit 0
fi

cluster_name=$1
cluster_region=$2

# Try connecting
aws eks --region $2 update-kubeconfig --name $1

cluster_path="kustomize/overlay/${cluster_name}"
check_cluster_name ${cluster_path}

ENVIRONMENT=$cluster_name

case "$3" in
	deploy-dev-chain)
			deploy_dev_chain ${@:4} ;;
	deploy-prod-chain)
			deploy_prod_chain ${@:4} ;;
	run-dev-relay)
			run_dev_relay  ${@:4} ;;
	run-prod-relay)
			run_prod_relay  ${@:4} ;;
	deploy-blockscout)
			deploy_blockscout ${@:4} ;;
 	gen-relay-svc)
			gen_relay_service ${@:4} ;;
	label-relay)
		label_relay_services  ${@:4} ;;
	gen-chain-relay-cm)
		gen_chain_relay_cm ${@:4} ;;
	update-subgraph)
			update_subgraph ${@:4} ;;
	gen-graph-svc)
			gen_graph_service ${@:4} ;;
	locate-pod-by-port)
			locate_pod_by_port ${@:4} ;;
	gen-oauth2-config)
		gen_oauth2_config ${@:4};;
	update-oauth2-users)
			update_oauth2_users ${@:4};;
	delete-all)
	    delete_all ;;
	*)
		echo "unknown cmd $3"
esac

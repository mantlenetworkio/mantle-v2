#!/bin/bash
set -e

### Deploy EiganDA infrastructure

function deploy_eigenda {
	# Deploy eigenlayer

	# Gen configs

	# Init participants

	# dl-node ip map

	# Deploy services

	# Update subgraph

	# Deploy graph
	kustomize build ./$cluster_path/graph | kubectl apply -f -

	# Deploy eigenda components
	kustomize build ./$cluster_path/eigenda | kubectl apply -f -
}

## Deploy Eigenlayer

function deploy_eigenlayer {

	# Remove old files
	rm -rf $cluster_path/eigenda/disperser/env
	rm -rf $cluster_path/eigenda/retriever/env
	rm -rf $cluster_path/eigenda/node/env

	rm -f $cluster_path/config/networks.json
	rm -f $cluster_path/config/participants.json

	rm -f $cluster_path/deploy.log
	rm -f $cluster_path/config.lock.yaml

	# Do deployment
	pushd ../integration
		go run deploy/cmd/main.go --path=../cluster/$cluster_path
	popd

	# Arrange files
	mkdir -p $cluster_path/eigenda/disperser/env
	mkdir -p $cluster_path/eigenda/retriever/env
	mkdir -p $cluster_path/eigenda/node/env

	for FILE in $(ls $cluster_path/envs/dln*.env); do
		echo "file  $FILE"
		id=$(basename $FILE | tr -d -c 0-9)
        cp $FILE $cluster_path/eigenda/node/env/dl-node-${id}.env
    done

	for FILE in $(ls $cluster_path/envs/dis*.env); do
		id=$(basename $FILE | tr -d -c 0-9)
        cp $FILE $cluster_path/eigenda/disperser/env/dl-disperser-${id}.env
    done

	for FILE in $(ls $cluster_path/envs/ret*.env); do
	id=$(basename $FILE | tr -d -c 0-9)
			cp $FILE $cluster_path/eigenda/retriever/env/dl-retriever-${id}.env
	done

	rm -rf $cluster_path/envs

	rm -rf $cluster_path/config/graph
	mkdir -p $cluster_path/config/graph

	cp ../subgraph/networks.json $cluster_path/config/graph/networks.json
	cp ../subgraph/subgraph.yaml $cluster_path/config/graph/subgraph.yaml

	mv $cluster_path/participants.json $cluster_path/config/participants.json
	mv $cluster_path/addresses.json $cluster_path/config/addresses.json
}

## Generate Config Maps

function gen_config_maps {

	gen_subgraph_config
	gen_component_config
	gen_ipmap_config

}

function gen_subgraph_config {

	rm -f $cluster_path/config/graph/kustomization.yaml

	pushd $cluster_path/config/graph
		kustomize create --labels app:graph
		kustomize edit add configmap subgraph-config --from-file subgraph.yaml
		kustomize edit add configmap graph-network-config --from-file networks.json
	popd
	echo "Wrote kustomize files into $cluster_path/config/graph. Nothing applied"
}


function gen_component_config {

	num_dis=$(ls $cluster_path/eigenda/disperser/env/dl-disperser-*.env | wc -l )
	num_dln=$(ls $cluster_path/eigenda/node/env/dl-node-*.env | wc -l )
	num_ret=$(ls $cluster_path/eigenda/retriever/env/dl-retriever-*.env | wc -l )

	rm -f $cluster_path/eigenda/disperser/kustomization.yaml
	rm -f $cluster_path/eigenda/retriever/kustomization.yaml
	rm -f $cluster_path/eigenda/node/kustomization.yaml

	# Gen patches first
	# Generate disperser volume patches here
	pushd $cluster_path/eigenda/disperser
		kustomize create --resources ../../../../base/disperser/
		kustomize edit set replicas dl-disperser=$num_dis
		for ((ind=0; ind<$num_dis; ind++)); do
			kustomize edit add patch --kind StatefulSet --group apps --version v1 --name dl-disperser --patch "- op: add
  path: /spec/template/spec/initContainers/0/volumeMounts/${ind}
  value:
    name: dl-disperser-${ind}
    mountPath: /dl-disperser-${ind}
- op: add
  path: /spec/template/spec/volumes/${ind}
  value:
    name: dl-disperser-${ind}
    configMap:
      name: dl-disperser-${ind}"
		done
	popd

	# Generate retriever volume patches here
	pushd $cluster_path/eigenda/retriever
		kustomize create --resources ../../../../base/retriever/
		kustomize edit set replicas dl-retriever=$num_ret
		for ((ind=0; ind<$num_ret; ind++)); do
			kustomize edit add patch --kind StatefulSet --group apps --version v1 --name dl-retriever --patch "- op: add
  path: /spec/template/spec/initContainers/0/volumeMounts/${ind}
  value:
    name: dl-retriever-${ind}
    mountPath: /dl-retriever-${ind}
- op: add
  path: /spec/template/spec/volumes/${ind}
  value:
    name: dl-retriever-${ind}
    configMap:
      name: dl-retriever-${ind}"
		done
	popd

	# Generate node volume patches here
	pushd $cluster_path/eigenda/node
		kustomize create --resources ../../../../base/node/
		kustomize edit set replicas dl-node=$num_dln
		for ((ind=0; ind<$num_dln; ind++)); do
			kustomize edit add patch --kind StatefulSet --group apps --version v1 --name dl-node --patch "- op: add
  path: /spec/template/spec/initContainers/0/volumeMounts/${ind}
  value:
    name: dl-node-${ind}
    mountPath: /dl-node-${ind}
- op: add
  path: /spec/template/spec/volumes/${ind}
  value:
    name: dl-node-${ind}
    configMap:
      name: dl-node-${ind}"
		done
	popd

	# DISPERSER
	pushd $cluster_path/eigenda/disperser/env/
		rm -f kustomization.yaml
		kustomize create --labels app:dl-disperser
		for ((ind=0; ind<$num_dis; ind++)); do
			kustomize edit add configmap dl-disperser-${ind} --from-env-file dl-disperser-${ind}.env
		done
		echo "generatorOptions:
  disableNameSuffixHash: true" >> kustomization.yaml
	popd

	# RETRIEVER
	pushd $cluster_path/eigenda/retriever/env
		rm -f kustomization.yaml
		kustomize create --labels app:dl-retriever
		for ((ind=0; ind<$num_ret; ind++)); do
			kustomize edit add configmap dl-retriever-${ind} --from-env-file dl-retriever-${ind}.env
		done
		echo "generatorOptions:
  disableNameSuffixHash: true" >> kustomization.yaml
	popd

	# DLNS
	pushd $cluster_path/eigenda/node/env
		rm -f kustomization.yaml
		kustomize create --labels app:dl-node
		for ((ind=0; ind<$num_dln; ind++)); do
			kustomize edit add configmap dl-node-${ind} --from-env-file dl-node-${ind}.env
		done
		echo "generatorOptions:
  disableNameSuffixHash: true" >> kustomization.yaml
	popd

	echo "wrote kustomize files into $cluster_path/eigenda/. Nothing applied"
}

function gen_ipmap_config {
	rm -rf $cluster_path/config/node
	mkdir -p $cluster_path/config/node

	ip_map_file="$cluster_path/config/node/dl-nodes-ip-map.env"
	nodesYaml="$cluster_path/config/node/nodes.yaml"

	kubectl get nodes -o yaml > $nodesYaml
	num_node=$(cat $nodesYaml | yq '.items' | yq length)

	for ((id=0; id<$num_node; id++)); do
		ext_ip=$(cat $nodesYaml | yq ".items[$id] | .status.addresses | .[] | select(.type == \"ExternalIP\") | .address")
		int_ip=$(cat $nodesYaml | yq ".items[$id] | .status.addresses | .[] | select(.type == \"InternalIP\") | .address")

		# if use kind, a local k8s simulator, ext ip is empty, then map to itself
		if [ -z ${ext_ip} ]; then
			echo "Int_IP.${int_ip}=${int_ip}" >> $ip_map_file
		else
			echo "Int_IP.${int_ip}=${ext_ip}" >> $ip_map_file
		fi
	done

	pushd $cluster_path/config/node/
		kustomize create --labels app:dl-node
		kustomize edit add configmap dl-nodes-ip-map --from-env-file dl-nodes-ip-map.env
	popd

	echo "wrote kustomize files into $cluster_path/config/node/. Nothing applied"

	# Delete tmp files
	rm $cluster_path/config/node/nodes.yaml
}

## Kustomize helper

function create_node_selector_patch {
	if [ "$#" -ne 4 ]; then
		echo "need patch directory"
		echo "need node selector flag (true/false)"
		echo "need node label name"
		echo "need app name"
		exit 0
	fi

	rm -rf $1/patches

    if [ "$2" = true ]; then
        mkdir -p $1/patches
        cat <<EOF > $1/patches/node-selector.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: $3
spec:
  template:
    spec:
      nodeSelector:
        app: $4
EOF

		echo "patchesStrategicMerge:
- ./patches/node-selector.yaml" >> $1/kustomization.yaml
    fi
}

## Run Graph

function run_graph {
	if [ "$#" -ne 3 ]; then
		echo "need number of graph"
		echo "need subgraph starter image"
		echo "need node selector flag (true/false)"
		exit 0
	fi

	mkdir -p ./$cluster_path/graph
	num_graph=$1
	subgraph_starter_image=$2
	use_node_selector=$3

	cat <<EOF >$cluster_path/graph/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
replicas:
- count: $num_graph
  name: graph
resources:
- ../config/relay
- ../config/graph
- ../../../base/graph
images:
- name: subgraph-starter
  name: ghcr.io/layr-labs/datalayr/subgraph-starter
  newTag: $subgraph_starter_image
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
EOF
	create_node_selector_patch $cluster_path/graph $use_node_selector graph graph

	kustomize build ./$cluster_path/graph | kubectl apply -f -
  	echo "Wrote graph kustomize into $cluster_path/graph/kustomization.yaml. And Applied"

	echo "Wait for graph to start..."

	run_graph_service $num_graph
}

function run_graph_service {
  	num_promised_graph=$1

	mkdir -p $cluster_path/graph/tmp

	podsYaml="$cluster_path/graph/tmp/graphsDeployState.yaml"
	kubectl get pods -l app=graph -o yaml > $podsYaml

	num_svc=$(cat $podsYaml | yq '.items' | yq length)
	if [ "${num_promised_graph}" -ne ${num_svc} ]; then
		echo "Not all Pods are ready. Retry Later. Promised ${num_promised_graph}. Actual ${num_svc}"
		exit 0
	fi

	port=32500
	for ((id=0; id<$num_svc; id++)); do
	  http_port=${port}
		((port+=1))
		echo "http_port $http_port"

		hostname=$(cat $podsYaml | yq ".items[$id].spec.hostname")
		svc_name="graph-svc-${id}"

   	yq "
		.metadata.name = \"$svc_name\" |
		.spec.selector.\"statefulset.kubernetes.io/pod-name\" = \"$hostname\" |
		.spec.ports.[] |= (
			with(select(.name == \"http\" );
				.port = 8000 |
				.nodePort = ${http_port} |
				.targetPort = 8000
			)
		 )
		 " kustomize/base/graph/service.yaml | kubectl apply -f -
	done

	echo "Wrote nothing to filesystem. Applied ${num_svc} svc based on kustomize/base/graph/service.yaml"

	rm -rf $cluster_path/graph/tmp
}

## Run EigenDA

function run_eigenda {
	if [ "$#" -ne 5 ]; then
		echo "need number of graph. Should be identical to argument for run-graph"
		echo "need number of relay. Should be less than or equal to total number of relay"
		echo "need image tag"
		echo "need data loader tag"
		echo "need node selector flag (true/false)"
		exit 0
	fi

	mkdir -p ./$cluster_path/eigenda
	mkdir -p ./$cluster_path/eigenda/patches

	num_graph=$1
	num_relay=$2
	image_tag=$3
	data_loader_tag=$4
	use_node_selector=$5

	cat <<EOF >$cluster_path/eigenda/patches/num-graphs.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: dl-disperser
spec:
  template:
    spec:
      containers:
        - name: dl-disperser
          env:
            - name: NUM_CLUSTER_GRAPH
              value: "${num_graph}"
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: dl-retriever
spec:
  template:
    spec:
      containers:
        - name: dl-retriever
          env:
            - name: NUM_CLUSTER_GRAPH
              value: "${num_graph}"
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: dl-node
spec:
  template:
    spec:
      containers:
        - name: dl-node
          env:
            - name: NUM_CLUSTER_GRAPH
              value: "${num_graph}"
EOF

	echo "Add Env of NUM_CLUSTER_GRAPH=${num_graph} into Eigenda Patch. So it can map to graph. Nothing Applied"

	cat <<EOF >$cluster_path/eigenda/patches/num-relays.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: dl-disperser
spec:
  template:
    spec:
      containers:
        - name: dl-disperser
          env:
            - name: NUM_CHAIN_RELAY
              value: "${num_relay}"
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: dl-retriever
spec:
  template:
    spec:
      containers:
        - name: dl-retriever
          env:
            - name: NUM_CHAIN_RELAY
              value: "${num_relay}"
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: dl-node
spec:
  template:
    spec:
      containers:
        - name: dl-node
          env:
            - name: NUM_CHAIN_RELAY
              value: "${num_relay}"
EOF


	echo "Add Env of NUM_CHAIN_RELAY=${num_relay} into Eigenda Patch. So they can map to relay. Nothing Applied"


cat <<EOF >$cluster_path/eigenda/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
images:
- name: ghcr.io/layr-labs/datalayr/dl-disperser
  newTag: $image_tag
- name: ghcr.io/layr-labs/datalayr/dl-retriever
  newTag: $image_tag
- name: ghcr.io/layr-labs/datalayr/dl-node
  newTag: $image_tag
- name: ghcr.io/layr-labs/datalayr/init-container-data-loader
  newTag: $data_loader_tag
patchesStrategicMerge:
- ./patches/num-graphs.yaml
- ./patches/num-relays.yaml
resources:
- ../config/relay
- ../config/node
- ./disperser
- ./disperser/env
- ./node
- ./node/env
- ./retriever/env
- ./retriever
EOF

	create_node_selector_patch $cluster_path/eigenda/disperser $use_node_selector dl-disperser disperser
	create_node_selector_patch $cluster_path/eigenda/node $use_node_selector dl-node node
	create_node_selector_patch $cluster_path/eigenda/retriever $use_node_selector dl-retriever retriever

	kustomize build $cluster_path/eigenda | kubectl apply -f -
	echo "Wrote Nothing for all eigenda components: disperser, node, retriever. Applied in run time"
	echo "Waiting for all pods to start..."

	# kubectl rollout status sts dl-disperser
	# kubectl rollout status sts dl-retriever
	# kubectl rollout status sts dl-node

	run_disperser_service
	run_retriever_service
	run_node_service
}


function run_node_service {
	mkdir -p $cluster_path/eigenda/node/tmp

	podsYaml="$cluster_path/eigenda/node/tmp/nodesDeployState.yaml"
	kubectl get pods -l app=dl-node -o yaml > $podsYaml

	num_svc=$(cat $podsYaml | yq '.items' | yq length)
	for ((id=0; id<$num_svc; id++)); do
		hostname=$(cat $podsYaml | yq ".items[$id].spec.hostname")
		hostIP=$(cat $podsYaml | yq ".items[$id].status.hostIP")

		# the output in alphabetic order, not associated with id
		index=$(printf "%s\n" "${hostname##*-}")
		# hostname is identical configMap name
		kubectl get cm $hostname -o yaml > $cluster_path/eigenda/node/tmp/cm.yaml
		grpcPort=$(cat $cluster_path/eigenda/node/tmp/cm.yaml | yq '.data.DL_NODE_GRPC_PORT')
		metricsPort=$(cat $cluster_path/eigenda/node/tmp/cm.yaml | yq '.data.DL_NODE_METRICS_PORT')
		# yq inplace udpate
		svc_name="dl-node-svc-${index}"
		metrics_svc_name="dl-node-metrics-svc-${index}"

    	yq  "
			.metadata.name = \"$svc_name\" |
			.spec.selector.\"statefulset.kubernetes.io/pod-name\" = \"$hostname\" |
			.spec.ports.[] |= (
				with(select(.name == \"grpc\" );
					.port = ${grpcPort} |
					.nodePort = ${grpcPort} |
					.targetPort = ${grpcPort}
				)
		 )
		 " kustomize/base/node/service.yaml | kubectl apply -f -

			yq  "
			.metadata.name = \"$metrics_svc_name\" |
			.spec.selector.\"statefulset.kubernetes.io/pod-name\" = \"$hostname\" |
			.spec.ports.[] |= (
				with(select(.name == \"metrics\");
					.port = $metricsPort |
					.targetPort = $metricsPort
				)
		 )
		 " kustomize/base/node/service-metrics.yaml | kubectl apply -f -

	done

	rm -rf $cluster_path/eigenda/node/tmp
}

function run_disperser_service {
	mkdir -p $cluster_path/eigenda/disperser/tmp

	podsYaml="$cluster_path/eigenda/disperser/tmp/nodesDeployState.yaml"
	kubectl get pods -l app=dl-disperser -o yaml > $podsYaml

	num_svc=$(cat $podsYaml | yq '.items' | yq length)
	for ((id=0; id<$num_svc; id++)); do
		hostname=$(cat $podsYaml | yq ".items[$id].spec.hostname")
		hostIP=$(cat $podsYaml | yq ".items[$id].status.hostIP")

		# the output in alphabetic order, not associated with id
		index=$(printf "%s\n" "${hostname##*-}")
		# hostname is identical configMap name
		kubectl get cm $hostname -o yaml > $cluster_path/eigenda/disperser/tmp/cm.yaml
		grpcPort=$(cat $cluster_path/eigenda/disperser/tmp/cm.yaml | yq '.data.DL_DISPERSER_GRPC_PORT')
		metricsPort=$(cat $cluster_path/eigenda/disperser/tmp/cm.yaml | yq '.data.DL_DISPERSER_METRICS_PORT')
		# yq inplace udpate
		svc_name="dl-disperser-svc-${index}"
		metrics_svc_name="dl-disperser-metrics-svc-${index}"

    	yq  "
			.metadata.name = \"$svc_name\" |
			.spec.selector.\"statefulset.kubernetes.io/pod-name\" = \"$hostname\" |
			.spec.ports.[] |= (
				with(select(.name == \"grpc\" );
					.port = ${grpcPort} |
					.nodePort = ${grpcPort} |
					.targetPort = ${grpcPort}
				)
		 )
		 " kustomize/base/disperser/service.yaml | kubectl apply -f -

			yq  "
			.metadata.name = \"$metrics_svc_name\" |
			.spec.selector.\"statefulset.kubernetes.io/pod-name\" = \"$hostname\" |
			.spec.ports.[] |= (
				with(select(.name == \"metrics\");
					.port = $metricsPort |
					.targetPort = $metricsPort
				)
		 )
		 " kustomize/base/disperser/service-metrics.yaml | kubectl apply -f -
	done

	rm -rf $cluster_path/eigenda/disperser/tmp
}


function run_retriever_service {
	mkdir -p $cluster_path/eigenda/retriever/tmp

	podsYaml="$cluster_path/eigenda/retriever/tmp/nodesDeployState.yaml"
	kubectl get pods -l app=dl-retriever -o yaml > $podsYaml

	num_svc=$(cat $podsYaml | yq '.items' | yq length)
	for ((id=0; id<$num_svc; id++)); do
		hostname=$(cat $podsYaml | yq ".items[$id].spec.hostname")
		hostIP=$(cat $podsYaml | yq ".items[$id].status.hostIP")

		# the output in alphabetic order, not associated with id
		index=$(printf "%s\n" "${hostname##*-}")
		# hostname is identical configMap name
		kubectl get cm $hostname -o yaml > $cluster_path/eigenda/retriever/tmp/cm.yaml
		grpcPort=$(cat $cluster_path/eigenda/retriever/tmp/cm.yaml | yq '.data.DL_RETRIEVER_GRPC_PORT')
		# yq inplace udpate
		svc_name="dl-retriever-svc-${index}"

    	yq  "
			.metadata.name = \"$svc_name\" |
			.spec.selector.\"statefulset.kubernetes.io/pod-name\" = \"$hostname\" |
			.spec.ports.[] |= (
				with(select(.name == \"grpc\" );
					.port = ${grpcPort} |
					.nodePort = ${grpcPort} |
					.targetPort = ${grpcPort}
				)
		 )
		 " kustomize/base/retriever/service.yaml | kubectl apply -f -
	done

	rm -rf $cluster_path/eigenda/retriever/tmp
}

## Run Prometheus

function run_prometheus {
	#if [ "$#" -ne 1 ]; then
		#echo "need prometheus port number"
		#exit 0
	#fi
	mkdir -p $cluster_path/prometheus
	#prometheus_port=$1

	gen_prometheus_config

	cat <<EOF >$cluster_path/prometheus/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
replicas:
- count: 1
  name: prometheus
resources:
- ../../../base/prometheus
- ../config/prometheus
EOF
	kustomize build $cluster_path/prometheus | kubectl apply -f -
	echo "Wrote prometheus kustomize into $cluster_path/prometheus/kustomization.yaml. And Applied"

	echo "Wait for prometheus to start..."
	kubectl rollout status deployment prometheus

	gen_prometheus_service $prometheus_port
}

function gen_prometheus_config {
	rm -rf $cluster_path/config/prometheus
	mkdir -p $cluster_path/config/prometheus
	mkdir -p $cluster_path/prometheus/tmp

	#nodes_svc_yaml="$cluster_path/prometheus/tmp/nodes_svc.yaml"
	#kubectl get service -l app=dl-node -o yaml > $nodes_svc_yaml
	#num_node_svc=$(cat ${nodes_svc_yaml} | yq '.items' | yq length)

	#dispersers_svc_yaml="$cluster_path/prometheus/tmp/dispersers_svc.yaml"
	#kubectl get service -l app=dl-disperser -o yaml > $dispersers_svc_yaml
	#num_disperser_svc=$(cat ${dispersers_svc_yaml} | yq '.items' | yq length)

	prom_file="$cluster_path/config/prometheus/prometheus.yml"
	pushd ../integration/prom-grafana/
		go run prometheus-yaml-gen/cmd/main.go --path=prometheus-yaml-gen/template.yml > prometheus.yml
		mv prometheus.yml ../../cluster/${prom_file}
	popd

	pushd $cluster_path/config/prometheus
		kustomize create --labels app:prometheus
		kustomize edit add configmap prometheus-config --from-file prometheus.yml
	popd
}

function gen_prometheus_service {
	port=31333
	yq  "
		.spec.ports.[] |= (
			with(select(.name == \"http\" );
				.nodePort = ${port}
			)
		)" kustomize/base/prometheus/service.yaml | kubectl apply -f -
}

## Teardown

function delete_eigenda {
	set +e
	kustomize build $cluster_path/eigenda | kubectl delete -f -
	# pvc needs explicit deletion
	kubectl delete pvc --selector 'owner in (dl-disperser, dl-node, dl-retriever)'
	# svc need explicit deletion since no yaml are generated
	kubectl delete  svc --selector 'app in (dl-node, dl-disperser, dl-retriever)'
	set -e
}

function delete_all {
	delete_eigenda
	set +e
	kustomize build ./$cluster_path/graph | kubectl delete -f -
	# pvc needs explicit deletion
	kubectl delete pvc --selector 'owner in (graph)'
	# svc need explicit deletion since no yaml are generated
	kubectl delete svc --selector 'owner in (graph)'
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

	if [ ! -f "${cluster_path}/config.yaml" ]; then
		echo "cluster directory config does not exists. $cluster_path/config.yaml"
		exit 0
	fi

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




if [ "$1" == "help" ] || [ "$2" == "help" ] ; then
	cat <<-EOF
		# Deploy eigenda infa
		1. deploy-eigenlayer
		2. gen-config-maps
		3. run-graph
		4. run-eigenda
		5. run-prometheus

		# Analysis
		1.  locate-pod-by-port

		# Cleanup
		1. delete-eigenda (sts,svc,cm,pvc for disperser,node,retriever)
EOF
	exit 0
fi

cluster_name=$1
cluster_region=$2

# Try connecting
aws eks --region $2 update-kubeconfig --name $1

cluster_path="kustomize/overlay/${cluster_name}"
check_cluster_name ${cluster_path}

case "$3" in
	deploy-eigenda)
			deploy_eigenda ${@:4} ;;
	deploy-eigenlayer)
			deploy_eigenlayer ${@:4} ;;
	gen-config-maps)
		gen_config_maps ${@:4} ;;
	run-graph)
		  run_graph ${@:4} ;;
	run-eigenda)
		  run_eigenda ${@:4} ;;
	run-prometheus)
		  run_prometheus ${@:4} ;;
	run-graph-svc)
			run_graph_service ${@:4} ;;
	run-nodes-svcs)
			run_node_service ${@:4} ;;
	run-dispersers-svcs)
			run_disperser_service ${@:4} ;;
	run-retrievers-svcs)
			run_retriever_service ${@:4} ;;
	locate-pod-by-port)
			locate_pod_by_port ${@:4} ;;
	delete-eigenda)
			delete_eigenda ;;
	delete-all)
		  delete_all ;;
	*)
		echo "unknown cmd $3"
esac

#!/bin/bash

set -e

source .env

function new_eigenda_cluster {
	if [ "$#" -ne 1 ]; then
		echo "used for creating a new k8s eigenda cluster"
		echo "need new cluster name"
		exit 0
	fi

	name=$1

	# create cluster dir
	mkdir -p ./kustomize/overlay/$name
	mkdir -p ./kustomize/overlay/$name/config

	echo "ENVIRONMENT=$name" > ./kustomize/overlay/$name/.env
}

function new_dev_cluster {
	if [ "$#" -ne 1 ]; then
		echo "used for creating a new k8s dev cluster"
		echo "need new cluster name"
		exit 0
	fi
	name=$1
	# create cluster dir
	mkdir -p ./kustomize/overlay/$name
	# kustomize sealer sts
	mkdir -p ./kustomize/overlay/$name/sealer
	# kustomize relay sts
	mkdir -p ./kustomize/overlay/$name/relay
	echo "ENVIRONMENT=$name" > ./kustomize/overlay/$name/.env

	cat <<EOF >kustomize/overlay/$name/sealer/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- ../../../base/sealer/
EOF

	#
	cat <<-EOF
		# next steps
		1. use infra-kd.sh to deploy chain related infrastructure resources on the new cluster
    2. use eigenda-kd.sh to deploy eigenda resources on the new cluster
EOF
}

if [ "$1" == "help" ] || [ "$2" == "help" ]  ; then
	cat <<-EOF
		# Deploy chain infra
		1. new-dev-cluster   (usage: init-cluster.sh staging new-dev-cluster)

		# Deploy eigenda infra
		1. new-eigenda-cluster   (usage: init-cluster.sh staging new-eigenda-cluster)

EOF
	exit 0
fi

function check_cluster_not_exist {
	if [ "$#" -ne 1 ]; then
		echo "Need cluster name"
		exit 0
	fi

	cluster_path=$1
	if [ -d "${cluster_path}" ]; then
		echo "cluster directory already exists. $cluster_path"
		exit 0
	fi
}

cluster_name=$1
cluster_path="kustomize/overlay/${cluster_name}"
check_cluster_not_exist ${cluster_path}
ENVIRONMENT=${cluster_name}

case "$2" in
	new-dev-cluster)
			new_dev_cluster $cluster_name ;;
	new-eigenda-cluster)
			new_eigenda_cluster $cluster_name ;;
	*)
		echo "unknown cmd $2"
esac

#!/bin/bash

set -e

# Deploy shared infrastructure

### MONITORING ###

function deploy_cloudwatch {
	source .env

	if [ "$#" -ne 2 ]; then
		echo "need cluster name and region"
		exit 0
	fi

	ClusterName=$1
	LogRegion=$2
	FluentBitHttpPort='2020'
	FluentBitReadFromHead='Off'
	[[ ${FluentBitReadFromHead} = 'On' ]] && FluentBitReadFromTail='Off'|| FluentBitReadFromTail='On'
	[[ -z ${FluentBitHttpPort} ]] && FluentBitHttpServer='Off' || FluentBitHttpServer='On'

	inputPath="kubernetes/monitoring/cloudwatch.yaml"
	outputPath="deployments/data/${ENVIRONMENT}/monitoring/cloudwatch.yaml"

	mkdir -p deployments/data/${ENVIRONMENT}/monitoring

	cat $inputPath | sed 's/{{cluster_name}}/'${ClusterName}'/;s/{{region_name}}/'${LogRegion}'/;s/{{http_server_toggle}}/"'${FluentBitHttpServer}'"/;s/{{http_server_port}}/"'${FluentBitHttpPort}'"/;s/{{read_from_head}}/"'${FluentBitReadFromHead}'"/;s/{{read_from_tail}}/"'${FluentBitReadFromTail}'"/'\
	> $outputPath

	kubectl apply -f $outputPath
}

case "$1" in help)
    cat <<-EOF
		# Monitoring
		1.  deploy-cloudwatch
EOF
							;;
	deploy-cloudwatch)
		  deploy_cloudwatch ${@:2} ;;
	*)
		echo "unknown cmd $1"
esac

#!/bin/bash

template=${1:-config.full.yaml}

dt=$(date '+%dD-%mM-%YY-%HH-%MM-%SS')
mkdir -p "data/${dt}"

cp $template data/${dt}/config.yaml

# Update config
source .env
sed -i".back" "s/\${HOST_IP}/${HOST_IP}/" data/${dt}/config.yaml && rm data/${dt}/config.yaml.back

export EXPERIMENT=$dt

# Update .env
sed -i"back" "s/EXPERIMENT=.*/EXPERIMENT=\"${EXPERIMENT}\"/" ./.env && rm ./.envback
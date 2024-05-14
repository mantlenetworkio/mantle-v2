#!/bin/bash

function push_w_tag {
    change_tags $1

    make images
    docker compose push

    change_tags "latest"
}

function change_tags {
    services="disperser node retriever sequencer challenger"
    tag=$1

    for service in $services; do
        image_url=$(cat docker-compose.yml | yq ".services.$service.image")
        readarray -d : -t image_url_parts <<< "$image_url"
        yq -i ".services.$service.image = \"${image_url_parts[0]}:$tag\"" docker-compose.yml
    done
}

case "$1" in
    push-w-tag)
        push_w_tag ${@:2} ;;
    *)
        tput setaf 1
        echo "Unknown subcommand" $1
        echo "./images.sh help"
        tput sgr0 ;;
esac

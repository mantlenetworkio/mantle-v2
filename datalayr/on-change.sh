#!/bin/bash

mkdir -p data

loc=$1
file=$2
command=$3

if [[ ! -f "$file" ]]; then
    touch $file
fi

prev_hash=$(cat $file)
hash=$(find $loc -type f -print0 | sort -z | xargs -0 md5sum | md5sum | head -c 32)

echo "Prev hash: ${prev_hash}, New Hash: ${hash}"

if [[ "$prev_hash" != "$hash" ]]; then
    echo "Do not match"
    echo $hash > $file
    eval $command
fi

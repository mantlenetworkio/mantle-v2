#!/bin/bash

token=$1
new_users=${@:2}

all_users=$(kubectl get cm docs-oauth2 -o yaml | yq '.data.OAUTH2_PROXY_GITHUB_USERS')

for username in $new_users; do

    curl \
        -X PUT \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer $token" \
        https://api.github.com/repos/Layr-Labs/datalayr/collaborators/$username \
        -d '{"permission":"pull"}'

    curl \
        -X PUT \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer $token" \
        https://api.github.com/repos/Layr-Labs/eignlayr-contracts/collaborators/$username \
        -d '{"permission":"pull"}'

    curl \
        -X PUT \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer $token" \
        https://api.github.com/repos/Layr-Labs/datalayr-rollup-example-contracts/collaborators/$username \
        -d '{"permission":"pull"}'

    all_users="$all_users,$username"

done


echo $all_users > github_users
# ./deploy-chain.sh chain update-oauth2-users

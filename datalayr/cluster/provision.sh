#!/bin/bash

function gen_priv_key {
    priv_key=$(cast wallet new)
    delimiter=" "
    priv_key_concat=$priv_key$delimiter
    words=()
    while [[ $priv_key_concat ]]; do
        words+=( "${priv_key_concat%%" "*}" )
        priv_key_concat=${priv_key_concat#*" "}
    done
    echo ${words[6]}
}

function provision_weth {
    
    source .env

    pk=$1
    pushd ../contracts/eignlayr-contracts > /dev/null
        pkUint=$(h2d $pk)
        echo $pkUint > data/recipient
        # --broadcast
        forge script script/Allocate.s.sol:ProvisionWeth --rpc-url $RPC_URL --private-key $PRIVATE_KEY --broadcast -vvvv
        rm -f recipient
    popd > /dev/null
}

function h2d {
    x=$(echo "ibase=16; ${1^^}"|bc)
    y=""
    for (( i=0; i<=${#x}; i++ ))
    do
        if [[ "0123456789" == *"${x:i:1}"* ]]; then
            y=$y${x:i:1}
        fi
    done
    echo $y
}

function gen_n_prov {
    source .env

    folder=deployments/${ENVIRONMENT}/data
    mkdir -p $folder

    now=$(date +%s)
    file=$folder/pks.$now
    touch $file
    
    for (( i=0; i<$1; i++ ))
    do
        sk=$(gen_priv_key)
        provision_weth $sk 
        echo $sk >> $file
    done
}

case "$1" in
    help)
        cat <<-EOF
        Deploy tool
EOF
        ;;
    gen-priv-key)
        gen_priv_key ${@:2} ;;
    provision-weth)
        provision_weth ${@:2} ;;
    gen-n-prov)
        gen_n_prov ${@:2} ;;
esac
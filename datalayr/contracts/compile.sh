#!/bin/bash

function compile_el {
    mkdir -p data

    pushd eignlayr-contracts > /dev/null
    forge clean
    forge build
    popd > /dev/null

    contracts="ERC20PresetFixedSupply BLSPublicKeyCompendium BLSRegistry InvestmentManager EigenLayrDelegation InvestmentManager InvestmentStrategyBase"
    for contract in $contracts; do
        create_binding eignlayr-contracts $contract ../common/contracts/bindings
    done
}

function compile_dl {
    mkdir -p data

    cp -r eignlayr-contracts datalayr-contracts/lib/

    pushd datalayr-contracts > /dev/null
		forge clean
        forge build
    popd > /dev/null

    rm -rf datalayr-contracts/lib/eignlayr-contracts

    contracts="DataLayrServiceManager DataLayrChallengeUtils DataLayrChallenge"
    for contract in $contracts; do
        create_binding datalayr-contracts $contract ../common/contracts/bindings
    done

    for FILE in datalayr-contracts/out/DataLayr*.sol; do
            contract=$(basename -s .sol $FILE)

            if [[ "$contract" != *".t"* ]]; then
                create_binding datalayr-contracts $contract ../common/contracts/bindings
            fi
    done

}

function compile_and_test_dl {
    compile_dl

    cp -r eignlayr-contracts datalayr-contracts/lib/

    pushd datalayr-contracts > /dev/null
        forge test --match-test testLoopConfirmDataStore
    popd > /dev/null

    rm -rf datalayr-contracts/lib/eignlayr-contracts
}

function create_binding {
    contract_dir=$1
    contract=$2
    binding_dir=$3
    echo $contract
    mkdir -p $binding_dir/${contract}
    contract_json="$contract_dir/out/${contract}.sol/${contract}.json"
    solc_abi=$(cat ${contract_json} | jq -r '.abi')
    solc_bin=$(cat ${contract_json} | jq -r '.bytecode.object')

    echo ${solc_abi} > data/tmp.abi
    echo ${solc_bin} > data/tmp.bin

    rm -f $binding_dir/${contract}/binding.go
    abigen --bin=data/tmp.bin --abi=data/tmp.abi --pkg=contract${contract} --out=$binding_dir/${contract}/binding.go
}

case "$1" in
    compile-el)
        compile_el ;;
    compile-dl)
        compile_dl ;;
    compile-and-test-dl)
        compile_and_test_dl ;;
    *)
        tput setaf 1
        echo "Unknown subcommand" $1
        echo "./compile.sh help"
        tput sgr0 ;;
esac

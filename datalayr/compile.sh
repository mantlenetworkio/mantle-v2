#!/bin/bash

function compile_interface {
    compile_interface_ "interfaceDL/disperse.proto"
    compile_interface_ "interfaceRetrieverServer/server.proto"
}

function compile_interface_ {
    protoc --go_out=. --go_opt=paths=source_relative  --go-grpc_out=. --go-grpc_opt=paths=source_relative common/interfaces/${1}
}

case "$1" in
    compile-interface)
        compile_interface ;;
    *)
        tput setaf 1
        echo "Unknown subcommand" $1
        echo "./local.sh help"
        tput sgr0 ;;
esac

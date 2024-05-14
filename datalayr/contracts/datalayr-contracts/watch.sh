#!/bin/bash

function start {
    watch &

    echo "Started watching eignlayr-contracts"    
}

function stop {
    pids=$(ps -ef | grep ./watch.sh | tr -s ' ' | cut -d ' ' -f2)
    pids_rev=""
    echo $pids
    for pid in $pids; do 
        pids_rev="$pid $pids_rev"
    done

    echo "this"
    echo $$
    for pid in $pids_rev; do 
        
        if [[ $pid != $$ ]]; then
            echo "that"
            echo $pid
            kill -9 $pid
        fi
    done
}

function watch {
    cp -r ../eignlayr-contracts/ ./lib
    chsum1=""
    while [[ true ]]
    do
        chsum2=`find ../eignlayr-contracts/ -type f -print0 | sort -z | xargs -0 md5sum | md5sum | head -c 32`
        if [[ $chsum1 != $chsum2 ]] ; then           
            if [ -n "$chsum1" ]; then
                cp -r ../eignlayr-contracts/ ./lib
            fi
            chsum1=$chsum2
        fi
        sleep 2
    done
}

case "$1" in
    start)
        start ;;
    stop)
        stop ;;
    *)
        tput setaf 1
        echo "Unknown subcommand" $1
        echo "./watch.sh help"
        tput sgr0 ;;
esac

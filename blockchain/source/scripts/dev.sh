#!/usr/bin/env bash

trap cleanup EXIT

pids=()

blockTime=$1
if [[ -n "$blockTime" ]]; then
    blockTime="-b $blockTime"
fi

cleanup() {
  for pid in "${pids[@]}"
  do
    echo "killing $pid"
    kill $pid
  done
  pids=()
}

rpc_running() {
  nc -z localhost "$1"
}

run() {
    port=$1;
    chain=$2
    if rpc_running $port; then
        echo "rpc is running on $port"
    else
        set -x
        node_modules/.bin/ganache-cli $blockTime --allowUnlimitedContractSize -l100000000 -d sonm --port=$port -i $port >> ./$port.log &
        { set +x; } 2>/dev/null
        pids+=($!)
        echo "started ganache on $port (pid $!)"
    fi
}

run 8545 main
run 8525 side

wait

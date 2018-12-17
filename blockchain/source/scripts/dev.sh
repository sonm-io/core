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
        echo "running node_modules/.bin/ganache-cli $blockTime -d sonm --port=$port -i $port >> ./$port.log &"
        node_modules/.bin/ganache-cli $blockTime -d sonm --port=$port -i $port >> ./$port.log &
        pids+=($!)
        echo "started ganache on $port (pid $!)"
    fi
}

run 8545 main
run 8525 side

wait

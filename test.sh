#!/bin/bash
if [ "$1" != "1" ]
then
    echo Get data via proxy ...
    for i in `seq 1 100`
    do 
        curl -sS -o /dev/null -X POST -H "Content-type:application/json" --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x1b4", true],"id":1}' 'http://localhost:3000' & 
    done
    wait
else
    echo Get data directly from node ...
    for i in `seq 1 100`
    do
        curl -sS -o /dev/null -X POST -H "Content-type:application/json" --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x1b4", true],"id":1}' 'https://testnet.tomochain.com' &
    done
    wait
fi

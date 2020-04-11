#!/bin/bash

for CMD in `ls $GOPATH/src/github.com/ebakus/ebakus-block-explorer-backend/cmd`; do
    echo "* Building ebakus_$CMD"
    GO111MODULE=on go build -o $GOPATH/bin/ebakus_$CMD $GOPATH/src/github.com/ebakus/ebakus-block-explorer-backend/cmd/$CMD
done

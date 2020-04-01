#!/bin/bash

for CMD in `ls $GOPATH/src/bitbucket.org/pantelisss/ebakus_server/cmd`; do
    echo "* Building ebakus_$CMD"
    GO111MODULE=on go build -o $GOPATH/bin/ebakus_$CMD $GOPATH/src/bitbucket.org/pantelisss/ebakus_server/cmd/$CMD
done

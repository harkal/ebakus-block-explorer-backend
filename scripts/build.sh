#!/bin/bash

cd $GOPATH/bin

for CMD in `ls $GOPATH/src/bitbucket.org/pantelisss/ebakus_server/cmd`; do
    echo "* Building ebakus_$CMD"
    go build -o ebakus_$CMD bitbucket.org/pantelisss/ebakus_server/cmd/$CMD
done

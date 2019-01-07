while [ true ]; do
 sleep 1
 $GOPATH/bin/ebakus_crawler fetchblocks --config $GOPATH/src/bitbucket.org/pantelisss/ebakus_server/configs/default.config.yaml
done

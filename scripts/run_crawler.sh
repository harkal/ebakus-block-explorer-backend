while [ true ]; do
 sleep 1
 /root/go/bin/ebakus_crawler fetchblocks --config /root/go/src/bitbucket.org/pantelisss/ebakus_server/default.yaml 
done

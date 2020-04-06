while :; do
    case $1 in
        -l|--local) local=true
        ;;
        *) break
    esac
    shift
done

# while [ true ]; do
#   sleep 1

  # check if a local ebakus instance is running, and connect to it
  ipc_arg=""
  if [ -n "${local+set}" ]; then
    ebakus_pid=$(pgrep -f "ebakus --datadir")
    if [ -n "$ebakus_pid" ]; then
      ipc_path=$(ps -o args -p $ebakus_pid | awk '{print $2}' | awk -F= '{print $2}')
      if [ -n "$ipc_path" ]; then
        ipc_arg="--ipc $ipc_path/ebakus.ipc"
      fi
    fi
  fi

  $GOPATH/bin/ebakus_crawler checkforks --config $GOPATH/src/bitbucket.org/pantelisss/ebakus_server/configs/default.config.yaml $ipc_arg
# done

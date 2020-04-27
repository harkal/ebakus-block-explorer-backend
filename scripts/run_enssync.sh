while :; do
    case $1 in
        -l|--local) local=true
        ;;
        *) break
    esac
    shift
done

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

$GOPATH/bin/ebakus_crawler enssync --config $GOPATH/src/github.com/ebakus/ebakus-block-explorer-backend/configs/default.config.yaml $ipc_arg

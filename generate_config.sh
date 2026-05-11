# 两个参数，一个是节点代码文件的根位置，另外一个是输出的目录位置
NODE_ROOT=$1
OUTPUT_PATH=$2

# 检查第一个参数是否为空
if [ -z "$NODE_ROOT" ]; then
    echo "Usage: $0 <NODE_ROOT> <OUTPUT_PATH:default=.>"
    exit 1
fi

# 如果第二个参数为空，给它赋一个默认值
if [ -z "$OUTPUT_PATH" ]; then
    OUTPUT_PATH="."
fi

#判断 NODE_ROOT 目录是否存在
if [ ! -d $NODE_ROOT ]; then
    echo "Node directory ($NODE_ROOT) not found!"
    exit 1
fi

#判断 OUTPUT_PATH 目录是否存在
if [ ! -d $OUTPUT_PATH ]; then
    echo "Output directory ($OUTPUT_PATH) not found!"
    exit 1
fi
# 获取目录下的JSON配置文件,并写入到 config.json 文件中
config_path="$OUTPUT_PATH/config.json"

# 写入基础配置
cat <<EOL >$config_path
{
  "GrpcPort": 50051,
  "HttpPort": 8081,
  "MaxEpochDelay ": 1,
  "MaxUnprocessedTaskPoolSize": 100,
  "MaxPendingSchedulePoolSize": 100,
  "MaxScheduledTasksPoolSize": 100,
  "MaxCommitSlotItemPoolSize": 100,
  "MaxGrpcRequestPoolSize": 200,
  "DefaultSlotSize": 100,
  "LogPath": "logs/",
  "FiscoBcosHost": "127.0.0.1",
  "FiscoBcosPort": 20200,
  "GroupID": "group0",
  "PrivateKey": "145e247e170ba3afd6ae97e88f00dbc976c2345d511b0f6713355d19d8b80b58",
  "TLSCaFile": "./ChainUpper/ca.crt",
  "TLSCertFile": "./ChainUpper/sdk.crt",
  "TLSKeyFile": "./ChainUpper/sdk.key",
  "QueueBufferSize": 100000,
  "WorkerCount": 3,
  "BatchSize": 1,
  "ABMStockDataDir": "/root/rappa/stockdata",
  "ABMStockParamDir": "/root/rappa/stockdata/params",
  "ABMV2MaxConcurrency": 4,
  "ErasureCodeParamN": 9,
  "ErasureCodeParamK": 6,
  "DEBUG": false ,
  "BHNodeAddressMap": {
EOL

# 写入节点配置
NODE_ID_LINE=""
NODE_IP_LINE=""
GRPC_PORT_LINE=""
for node_file in $(ls $NODE_ROOT); do
  if [ -d "$NODE_ROOT/$node_file" ]; then
    node_config_path="$NODE_ROOT/$node_file/RappaExecutor/config.json"
    if [ -f $node_config_path -a -r $node_config_path ]; then
        echo "Get config from NODE($node_file) successfully."
        # 这里使用sed格式化输出
        # 使用awk命令获取需要的节点地址信息
        NODE_ID_LINE=$(awk -F'[:, ]+' "/NODE_ID/{print \$3}" $node_config_path)
        NODE_IP_LINE=$(awk "/NODE_IP/{print}" $node_config_path | sed "s/NODE_IP/NodeIPAddress/")
        GRPC_PORT_LINE=$(awk "/GRPC_PORT/{print}" $node_config_path | sed "s/GRPC_PORT/NodeGrpcPort/;s/,\$//" )
        cat <<EOL >>$config_path
    "$NODE_ID_LINE": {
    $NODE_IP_LINE
    $GRPC_PORT_LINE
    },
EOL
    else
        echo "NODE($node_file) config not found!"
    fi
  fi
done

# 处理末尾格式
sed -i '$s/,//' $config_path
cat<<EOL >>$config_path
  }
}
EOL

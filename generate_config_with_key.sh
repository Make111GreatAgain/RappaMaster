#!/bin/bash

NODE_ROOT=$1
OUTPUT_PATH=$2

if [ -z "$NODE_ROOT" ]; then
    echo "Usage: $0 <NODE_ROOT> <OUTPUT_PATH:default=.>"
    exit 1
fi

if [ -z "$OUTPUT_PATH" ]; then
    OUTPUT_PATH="."
fi

if [ ! -d "$NODE_ROOT" ]; then
    echo "Node directory ($NODE_ROOT) not found!"
    exit 1
fi

if [ ! -d "$OUTPUT_PATH" ]; then
    echo "Output directory ($OUTPUT_PATH) not found!"
    exit 1
fi

config_path="$OUTPUT_PATH/config.json"
address_tmp=$(mktemp)
keys_tmp=$(mktemp)

trap 'rm -f "$address_tmp" "$keys_tmp"' EXIT

> "$address_tmp"
> "$keys_tmp"

# 复制 host 密钥到 RappaMaster/cert 目录
script_dir="$(cd "$(dirname "$0")" && pwd)"
executor_path="$(cd "$script_dir/../RappaExecutor" && pwd)"
master_cert_dir="$script_dir/cert"

echo "Copying host keys from RappaExecutor to RappaMaster/cert..."
mkdir -p "$master_cert_dir"

if [ -f "$executor_path/host_sk.key" ]; then
    cp "$executor_path/host_sk.key" "$master_cert_dir/"
    echo "  - Copied host_sk.key"
else
    echo "  - Warning: host_sk.key not found in $executor_path"
fi

if [ -f "$executor_path/host_pk.key" ]; then
    cp "$executor_path/host_pk.key" "$master_cert_dir/"
    echo "  - Copied host_pk.key"
else
    echo "  - Warning: host_pk.key not found in $executor_path"
fi

echo ""

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
  "CertPath": "cert",
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
  "ErasureCodeParamN": 9,
  "ErasureCodeParamK": 6,
  "Database": {
    "username": "root",
    "password": "520@111zz",
    "host": "127.0.0.1",
    "port": 3306,
    "dbname": "db_rappa",
    "timeout": "5s",
    "maxIdleConns": 10,
    "maxOpenConns": 100,
    "maxLifetime": "1h"
  },
  "IsAutoMigrate": true,
  "IsRecovery": true,
  "DEBUG": false,
  "BHNodeAddressMap": {
EOL

for node_file in $(ls $NODE_ROOT); do
  if [ -d "$NODE_ROOT/$node_file" ]; then
    node_config_path="$NODE_ROOT/$node_file/RappaExecutor/config.json"
    if [ -f $node_config_path -a -r $node_config_path ]; then
        echo "Get config from NODE($node_file) successfully."
        NODE_ID_LINE=$(awk -F'[:, ]+' "/NODE_ID/{print \$3}" $node_config_path)
        NODE_IP_LINE=$(awk "/NODE_IP/{print}" $node_config_path | sed "s/NODE_IP/NodeIPAddress/")
        GRPC_PORT_LINE=$(awk "/GRPC_PORT/{print}" $node_config_path | sed "s/GRPC_PORT/NodeGrpcPort/;s/,\$//" )
        cat <<EOL >>$address_tmp
    "$NODE_ID_LINE": {
    $NODE_IP_LINE
    $GRPC_PORT_LINE
    },
EOL

        cert_dir="$NODE_ROOT/$node_file/RappaExecutor/certs"
        spec_key_file="$cert_dir/node_spec_pk.key"
        bls_key_file="$cert_dir/node_bls_pk.key"

        if [ ! -f "$spec_key_file" -o ! -r "$spec_key_file" ]; then
            echo "NODE($node_file) spec public key not found!"
            exit 1
        fi
        if [ ! -f "$bls_key_file" -o ! -r "$bls_key_file" ]; then
            echo "NODE($node_file) bls public key not found!"
            exit 1
        fi

        SPEC_KEY=$(tr -d '\r\n' < "$spec_key_file")
        BLS_KEY=$(tr -d '\r\n' < "$bls_key_file")

        if [ -z "$SPEC_KEY" -o -z "$BLS_KEY" ]; then
            echo "NODE($node_file) key content is empty!"
            exit 1
        fi

        cat <<EOL >>$keys_tmp
    "$NODE_ID_LINE": {
      "spKey": "$SPEC_KEY",
      "blsKey": "$BLS_KEY"
    },
EOL
    else
        echo "NODE($node_file) config not found!"
    fi
  fi
done

if [ -s "$address_tmp" ]; then
    sed -i '$s/,//' "$address_tmp"
    cat "$address_tmp" >>$config_path
fi

cat <<EOL >>$config_path
  },
  "BHNodeKeyMap": {
EOL

if [ -s "$keys_tmp" ]; then
    sed -i '$s/,//' "$keys_tmp"
    cat "$keys_tmp" >>$config_path
fi

cat <<EOL >>$config_path
  }
}
EOL
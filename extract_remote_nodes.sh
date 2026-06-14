#!/bin/bash

# 参数说明:
# 1: NODE_ROOT (包含 node0, node1... 的目录)
# 2: REMOTE_IP (远程机器的 IP 地址)
# 3: PORT_OFFSET (端口偏移量，例如 30000，使 1234 变成 31234)
# 4: OUTPUT_FILE (可选，默认为 remote_nodes_snippet.json)

NODE_ROOT="${1:-}"
REMOTE_IP="${2:-}"
PORT_OFFSET="${3:-30000}"
OUTPUT_FILE="${4:-remote_nodes_snippet.json}"

if [[ -z "$NODE_ROOT" || -z "$REMOTE_IP" ]]; then
    echo "使用方法: $0 <NODE_ROOT> <REMOTE_IP> [PORT_OFFSET] [OUTPUT_FILE]"
    echo "参数说明:"
    echo "  NODE_ROOT   : 节点根目录 (例如 ../RappaExecutor/nodes)"
    echo "  REMOTE_IP   : 部署这些节点的机器 IP"
    echo "  PORT_OFFSET : 端口偏移量 (默认为 30000，1234 -> 31234)"
    echo "  OUTPUT_FILE : 输出文件名 (默认为 remote_nodes_snippet.json)"
    exit 1
fi

if [[ ! -d "$NODE_ROOT" ]]; then
    echo "错误: 未找到节点目录 $NODE_ROOT"
    exit 1
fi

echo "正在从 $NODE_ROOT 提取远程节点配置..."
echo "远程 IP: $REMOTE_IP"
echo "端口偏移: $PORT_OFFSET"

address_snippet=$(mktemp)
keys_snippet=$(mktemp)

trap 'rm -f "$address_snippet" "$keys_snippet"' EXIT

# 初始化片段文件
echo "  \"BHNodeAddressMap\": {" > "$address_snippet"
echo "  \"BHNodeKeyMap\": {" > "$keys_snippet"

# 获取所有节点并按数字排序
nodes=$(find "$NODE_ROOT" -maxdepth 1 -type d -name "node*" | sort -V)

for node_path in $nodes; do
    node_name=$(basename "$node_path")
    node_config="$node_path/RappaExecutor/config.json"
    
    if [[ -f "$node_config" ]]; then
        # 提取 NODE_ID 和原始 GRPC_PORT
        node_id=$(grep '"NODE_ID":' "$node_config" | sed 's/.*: \([0-9]*\),*/\1/')
        orig_port=$(grep '"GRPC_PORT":' "$node_config" | sed 's/.*: \([0-9]*\),*/\1/')
        
        if [[ -z "$node_id" || -z "$orig_port" ]]; then
            echo "警告: 无法从 $node_config 提取信息，跳过 $node_name"
            continue
        fi
        
        # 计算映射后的端口
        mapped_port=$((orig_port + PORT_OFFSET))
        
        # 提取公钥
        cert_dir="$node_path/RappaExecutor/certs"
        spec_key_file="$cert_dir/node_spec_pk.key"
        bls_key_file="$cert_dir/node_bls_pk.key"
        
        if [[ ! -f "$spec_key_file" || ! -f "$bls_key_file" ]]; then
            echo "警告: $node_name 缺少公钥文件，跳过"
            continue
        fi
        
        spec_key=$(tr -d '\r\n ' < "$spec_key_file")
        bls_key=$(tr -d '\r\n ' < "$bls_key_file")
        
        # 写入地址片段
        cat <<EOF >>"$address_snippet"
    "$node_id": {
      "NodeIPAddress": "$REMOTE_IP",
      "NodeGrpcPort": $mapped_port
    },
EOF

        # 写入密钥片段
        cat <<EOF >>"$keys_snippet"
    "$node_id": {
      "spKey": "$spec_key",
      "blsKey": "$bls_key"
    },
EOF
        echo "  - 已提取 $node_name (ID: $node_id, Port: $orig_port -> $mapped_port)"
    fi
done

# 去除最后一个逗号
sed -i '$s/,//' "$address_snippet"
sed -i '$s/,//' "$keys_snippet"

# 结束 JSON 片段
echo "  }," >> "$address_snippet"
echo "  }" >> "$keys_snippet"

# 合并到输出文件
{
    echo "{"
    cat "$address_snippet"
    echo ","
    cat "$keys_snippet"
    echo "}"
} > "$OUTPUT_FILE"

echo "-----------------------------------------------"
echo "提取完成！配置片段已保存至: $OUTPUT_FILE"
echo "您可以打开此文件并将其内容复制到 Master 的 config.json 中。"
echo "-----------------------------------------------"

#!/bin/bash
#
# DHV Volume Usage Query Script for Linux
# Usage: ./volume-usage.sh <volume-id>
#

set -e

# 默认配置
VOLUMES_DIR="${VOLUMES_DIR:-/opt/nomad/volumes}"
PLUGIN_PATH="${PLUGIN_PATH:-./mkfs}"

# 使用说明
usage() {
    cat <<EOF
Usage: $0 <volume-id>

Query disk usage for a dynamic host volume.

Arguments:
  volume-id         The Nomad volume ID to query

Environment Variables:
  VOLUMES_DIR       Directory where volumes are stored (default: /opt/nomad/volumes)
  PLUGIN_PATH       Path to the mkfs plugin binary (default: ./mkfs)

Platform:
  Linux only. Windows is not supported.

Examples:
  # Query by volume ID
  $0 abc123-def456-ghi789

  # With custom volumes directory
  VOLUMES_DIR=/mnt/volumes $0 abc123-def456-ghi789

  # Query by exact path
  DHV_CREATED_PATH=/opt/nomad/volumes/abc123-def456-ghi789 $PLUGIN_PATH usage
EOF
    exit 1
}

# 检查参数
if [ $# -eq 0 ]; then
    echo "Error: Volume ID is required"
    echo ""
    usage
fi

VOLUME_ID="$1"

# 检查插件是否存在
if [ ! -f "$PLUGIN_PATH" ]; then
    echo "Error: Plugin not found at $PLUGIN_PATH"
    exit 1
fi

# 检查卷目录是否存在
VOLUME_PATH="$VOLUMES_DIR/$VOLUME_ID"
if [ ! -d "$VOLUME_PATH" ]; then
    echo "Error: Volume directory not found: $VOLUME_PATH"
    exit 1
fi

# 调用插件
DHV_VOLUMES_DIR="$VOLUMES_DIR" \
DHV_VOLUME_ID="$VOLUME_ID" \
"$PLUGIN_PATH" usage

# 如果需要格式化输出，可以添加 jq
# "$PLUGIN_PATH" usage | jq -r '"Used: \(.used_bytes) bytes, Total: \(.total_bytes) bytes, Usage: \((.used_bytes / .total_bytes * 100) | round)%"'

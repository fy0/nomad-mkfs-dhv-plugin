#!/bin/bash
#
# DHV Volume Manager - Query and manage dynamic host volumes on Linux
#

set -e

# 默认配置
VOLUMES_DIR="${VOLUMES_DIR:-/opt/nomad/volumes}"
PLUGIN_PATH="${PLUGIN_PATH:-./mkfs}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 使用说明
usage() {
    cat <<EOF
DHV Volume Manager

Usage: $0 <command> [options]

Commands:
  list              List all volumes in the volumes directory
  usage <volume-id> Query disk usage for a specific volume
  usage-all         Query disk usage for all volumes
  info <volume-id>  Show detailed information about a volume

Environment Variables:
  VOLUMES_DIR       Directory where volumes are stored (default: /opt/nomad/volumes)
  PLUGIN_PATH       Path to the mkfs plugin binary (default: ./mkfs)

Platform:
  Linux only. Windows is not supported.

Examples:
  # List all volumes
  $0 list

  # Query usage for a specific volume
  $0 usage abc123-def456-ghi789

  # Query usage for all volumes
  $0 usage-all

  # Show volume info
  $0 info abc123-def456-ghi789
EOF
    exit 1
}

# 检查插件是否存在
check_plugin() {
    if [ ! -f "$PLUGIN_PATH" ]; then
        echo -e "${RED}Error: Plugin not found at $PLUGIN_PATH${NC}"
        exit 1
    fi
}

# 格式化字节数
format_bytes() {
    local bytes=$1
    if [ $bytes -lt 1024 ]; then
        echo "${bytes}B"
    elif [ $bytes -lt 1048576 ]; then
        echo "$(awk "BEGIN {printf \"%.2f\", $bytes/1024}")KB"
    elif [ $bytes -lt 1073741824 ]; then
        echo "$(awk "BEGIN {printf \"%.2f\", $bytes/1048576}")MB"
    else
        echo "$(awk "BEGIN {printf \"%.2f\", $bytes/1073741824}")GB"
    fi
}

# 列出所有卷
cmd_list() {
    echo -e "${GREEN}=== Volumes in $VOLUMES_DIR ===${NC}"

    if [ ! -d "$VOLUMES_DIR" ]; then
        echo -e "${RED}Error: Volumes directory not found: $VOLUMES_DIR${NC}"
        exit 1
    fi

    local count=0
    for dir in "$VOLUMES_DIR"/*; do
        if [ -d "$dir" ]; then
            local volume_id=$(basename "$dir")
            local mounted="No"

            # 检查是否挂载
            if mount | grep -q "$dir"; then
                mounted="Yes"
            fi

            echo -e "  Volume ID: ${YELLOW}$volume_id${NC}"
            echo -e "    Path: $dir"
            echo -e "    Mounted: $mounted"
            echo ""

            count=$((count + 1))
        fi
    done

    if [ $count -eq 0 ]; then
        echo "  No volumes found."
    else
        echo -e "${GREEN}Total: $count volumes${NC}"
    fi
}

# 查询单个卷的使用情况
cmd_usage() {
    local volume_id="$1"

    if [ -z "$volume_id" ]; then
        echo -e "${RED}Error: Volume ID is required${NC}"
        usage
    fi

    check_plugin

    local volume_path="$VOLUMES_DIR/$volume_id"
    if [ ! -d "$volume_path" ]; then
        echo -e "${RED}Error: Volume directory not found: $volume_path${NC}"
        exit 1
    fi

    echo -e "${GREEN}=== Volume Usage: $volume_id ===${NC}"

    # 调用插件获取使用情况
    local result=$(DHV_VOLUMES_DIR="$VOLUMES_DIR" DHV_VOLUME_ID="$volume_id" "$PLUGIN_PATH" usage 2>/dev/null)

    if [ $? -eq 0 ]; then
        # 解析 JSON 输出
        if command -v jq &> /dev/null; then
            local used=$(echo "$result" | jq -r '.used_bytes')
            local total=$(echo "$result" | jq -r '.total_bytes')
            local percent=$(awk "BEGIN {printf \"%.2f\", $used/$total*100}")

            echo -e "  Used:  $(format_bytes $used) (${used} bytes)"
            echo -e "  Total: $(format_bytes $total) (${total} bytes)"
            echo -e "  Usage: ${YELLOW}${percent}%${NC}"
        else
            echo "$result"
            echo ""
            echo -e "${YELLOW}Tip: Install 'jq' for formatted output${NC}"
        fi
    else
        echo -e "${RED}Error: Failed to query volume usage${NC}"
        exit 1
    fi
}

# 查询所有卷的使用情况
cmd_usage_all() {
    echo -e "${GREEN}=== All Volumes Usage ===${NC}"

    if [ ! -d "$VOLUMES_DIR" ]; then
        echo -e "${RED}Error: Volumes directory not found: $VOLUMES_DIR${NC}"
        exit 1
    fi

    check_plugin

    local count=0
    for dir in "$VOLUMES_DIR"/*; do
        if [ -d "$dir" ]; then
            local volume_id=$(basename "$dir")

            echo ""
            cmd_usage "$volume_id"

            count=$((count + 1))
        fi
    done

    if [ $count -eq 0 ]; then
        echo "  No volumes found."
    fi
}

# 显示卷详细信息
cmd_info() {
    local volume_id="$1"

    if [ -z "$volume_id" ]; then
        echo -e "${RED}Error: Volume ID is required${NC}"
        usage
    fi

    local volume_path="$VOLUMES_DIR/$volume_id"
    if [ ! -d "$volume_path" ]; then
        echo -e "${RED}Error: Volume directory not found: $volume_path${NC}"
        exit 1
    fi

    echo -e "${GREEN}=== Volume Info: $volume_id ===${NC}"
    echo -e "  Volume ID: ${YELLOW}$volume_id${NC}"
    echo -e "  Path: $volume_path"

    # 检查挂载状态
    local mount_info=$(mount | grep "$volume_path" || true)
    if [ -n "$mount_info" ]; then
        echo -e "  Status: ${GREEN}Mounted${NC}"
        echo -e "  Mount Info: $mount_info"
    else
        echo -e "  Status: ${RED}Not mounted${NC}"
    fi

    # 查找镜像文件
    local image_pattern="$VOLUMES_DIR/*-${volume_id:0:8}-*.img"
    local image_file=$(ls $image_pattern 2>/dev/null | head -1 || true)
    if [ -n "$image_file" ]; then
        echo -e "  Image File: $image_file"
        local image_size=$(stat -f%z "$image_file" 2>/dev/null || stat -c%s "$image_file" 2>/dev/null || echo "unknown")
        if [ "$image_size" != "unknown" ]; then
            echo -e "  Image Size: $(format_bytes $image_size)"
        fi
    fi

    # 如果挂载了，显示使用情况
    if [ -n "$mount_info" ]; then
        echo ""
        cmd_usage "$volume_id"
    fi
}

# 主函数
main() {
    if [ $# -eq 0 ]; then
        usage
    fi

    local command="$1"
    shift

    case "$command" in
        list)
            cmd_list "$@"
            ;;
        usage)
            cmd_usage "$@"
            ;;
        usage-all)
            cmd_usage_all "$@"
            ;;
        info)
            cmd_info "$@"
            ;;
        *)
            echo -e "${RED}Error: Unknown command: $command${NC}"
            echo ""
            usage
            ;;
    esac
}

main "$@"

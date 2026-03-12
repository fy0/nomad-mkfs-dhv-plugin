#!/bin/bash
#
# Quick usage query example (Linux only)
# Usage: ./quick-usage.sh <volume-id>
#

VOLUME_ID="${1:-}"
VOLUMES_DIR="${VOLUMES_DIR:-/opt/nomad/volumes}"

if [ -z "$VOLUME_ID" ]; then
    echo "Usage: $0 <volume-id>"
    echo ""
    echo "Example:"
    echo "  $0 abc123-def456-ghi789"
    echo ""
    echo "Environment Variables:"
    echo "  VOLUMES_DIR - Volume storage directory (default: /opt/nomad/volumes)"
    exit 1
fi

# Find plugin binary
if [ -f "./mkfs" ]; then
    PLUGIN="./mkfs"
elif [ -f "/usr/local/bin/mkfs" ]; then
    PLUGIN="/usr/local/bin/mkfs"
else
    echo "Error: mkfs plugin not found"
    exit 1
fi

# Query usage
echo "Querying volume: $VOLUME_ID"
echo ""

DHV_VOLUMES_DIR="$VOLUMES_DIR" \
DHV_VOLUME_ID="$VOLUME_ID" \
"$PLUGIN" usage | jq -r '"Used: \(.used_bytes) bytes (\((.used_bytes/1024/1024) | round)MB)\nTotal: \(.total_bytes) bytes (\((.total_bytes/1024/1024) | round)MB)\nUsage: \((.used_bytes/.total_bytes*100) | round)%"'

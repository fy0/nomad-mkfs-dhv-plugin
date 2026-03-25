# Nomad Dynamic Host Volume Plugin

A Go implementation of a Nomad Dynamic Host Volume Plugin for HashiCorp Nomad 1.10.0+, which supports the new Dynamic Host Volumes feature.

## Overview

This plugin enables Nomad to dynamically create and manage filesystem volumes on host machines. It creates volume images as files on the host machine and mounts them for use by Nomad jobs.

> [!IMPORTANT]
> This plugin is still actively being worked on. \
> Expect possible changes or reworks of how this plugin functions and how the config is structured.
>
> Additionally, this README isn't always up-to-date, so not everything mentioned here might work as described.

> [!CAUTION]
> Windows is not supported.
> The plugin relies on Unix-style tooling such as `mkfs.*`, `mount`, `/proc/mounts`, and `du -sb`, and is currently intended for Linux hosts.

## Features

- Create and mount filesystem images dynamically
- Support for multiple filesystem types (ext4, xfs, etc.)
- Native Go implementation minimizing external dependencies
- Configurable via volume parameters
- Full support for Nomad's Dynamic Host Volume API

## Installation

### Build from source

```bash
# Clone the repository
git clone https://github.com/mwantia/nomad-mkfs-dhv-plugin.git
cd nomad-mkfs-dhv-plugin

# Build the plugin
task build
```

## Usage

### Volume Configuration

Create a volume configuration in Nomad:

```hcl
name      = "<volume-name>"
node_id   = "<nomad-node-id>"
plugin_id = "mkfs"
type      = "host"

capacity_min = "500MB"
capacity_max = "500MB"

parameters {
    filesystem = "ext4"
    sparse     = true  # Default, can be set to false for pre-allocated files
}

```

### Available Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `filesystem` | Filesystem type to create (ext4, xfs, btrfs, etc.) | `ext4` |
| `sparse` | Create sparse file (recommended for better performance) | `true` |
| `block_size` | Block size for the filesystem (filesystem-specific) | `""` (auto) |
| `mount_options` | Additional mount options | `""` |
| `read_only` | Mount as read-only | `false` |

## Volume Usage Monitoring

This plugin includes a custom `usage` command to query disk usage for volumes. Note that this is **not** part of the official Nomad DHV specification, but a useful extension for monitoring.

This plugin also includes a custom `resize` command for growing existing ext4-backed volumes. `resize` is Linux-only, supports growth only, and does not support shrinking.

### Quick Usage

```bash
# Query usage by volume ID
DHV_VOLUMES_DIR=/opt/nomad/volumes \
DHV_VOLUME_ID=abc123-def456-ghi789 \
./mkfs usage

# Query usage by mount path
DHV_CREATED_PATH=/opt/nomad/volumes/abc123-def456-ghi789 \
./mkfs usage

# Grow an ext4 volume image to 2 GiB
DHV_VOLUMES_DIR=/opt/nomad/volumes \
DHV_VOLUME_ID=abc123-def456-ghi789 \
DHV_RESIZE_BYTES=2147483648 \
./mkfs resize
```

### Using Helper Scripts

The `scripts/` directory contains convenience scripts:

#### Volume Manager (Linux)

```bash
# Make script executable
chmod +x scripts/volume-manager.sh

# List all volumes
./scripts/volume-manager.sh list

# Query usage for a specific volume
./scripts/volume-manager.sh usage <volume-id>

# Grow a specific ext4 volume to the requested byte size
./scripts/volume-manager.sh resize <volume-id> <size-bytes>

# Query usage for all volumes
./scripts/volume-manager.sh usage-all

# Show detailed volume information
./scripts/volume-manager.sh info <volume-id>
```

#### Simple Usage Query (Linux)

```bash
chmod +x scripts/volume-usage.sh
./scripts/volume-usage.sh <volume-id>
```

### Integration Examples

#### Prometheus Node Exporter

Create a textfile collector script:

```bash
#!/bin/bash
# /etc/node_exporter/volume-usage-collector.sh

VOLUMES_DIR="/opt/nomad/volumes"
PLUGIN="/usr/local/bin/mkfs"
OUTPUT="/var/lib/node_exporter/textfile_collector/volume_usage.prom"

echo "# HELP nomad_dhv_volume_used_bytes Volume used bytes" > "$OUTPUT"
echo "# TYPE nomad_dhv_volume_used_bytes gauge" >> "$OUTPUT"
echo "# HELP nomad_dhv_volume_total_bytes Volume total bytes" >> "$OUTPUT"
echo "# TYPE nomad_dhv_volume_total_bytes gauge" >> "$OUTPUT"

for dir in "$VOLUMES_DIR"/*; do
    if [ -d "$dir" ]; then
        volume_id=$(basename "$dir")
        result=$(DHV_VOLUMES_DIR="$VOLUMES_DIR" DHV_VOLUME_ID="$volume_id" "$PLUGIN" usage 2>/dev/null)

        if [ $? -eq 0 ]; then
            used=$(echo "$result" | jq -r '.used_bytes')
            total=$(echo "$result" | jq -r '.total_bytes')

            echo "nomad_dhv_volume_used_bytes{volume_id=\"$volume_id\"} $used" >> "$OUTPUT"
            echo "nomad_dhv_volume_total_bytes{volume_id=\"$volume_id\"} $total" >> "$OUTPUT"
        fi
    fi
done
```

Schedule with cron:

```bash
*/5 * * * * /etc/node_exporter/volume-usage-collector.sh
```

#### Telegraf Integration

```toml
[[inputs.exec]]
  commands = ["/usr/local/bin/volume-usage-wrapper.sh"]
  timeout = "5s"
  data_format = "json"
  name_suffix = "_dhv"
  json_string_fields = ["volume_id"]
  tag_keys = ["volume_id"]
```

### Response Format

The `usage` command returns JSON:

```json
{
  "used_bytes": 1234567,
  "total_bytes": 524288000
}
```

The `resize` command returns JSON:

```json
{
  "path": "/opt/nomad/volumes/abc123-def456-ghi789",
  "old_bytes": 1073741824,
  "new_bytes": 2147483648,
  "filesystem": "ext4",
  "mounted": true
}
```

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DHV_VOLUMES_DIR` | Yes* | Directory where volumes are stored |
| `DHV_VOLUME_ID` | Yes* | Volume ID to query |
| `DHV_CREATED_PATH` | Yes* | Alternative: Direct path to volume mount |

\* Either `DHV_VOLUMES_DIR` + `DHV_VOLUME_ID` OR `DHV_CREATED_PATH` is required.

## Development

### Prerequisites

- Go 1.19 or higher
- HashiCorp Nomad 1.10.0 or later

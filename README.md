# Nomad Dynamic Host Volume Plugin

A Go implementation of a Nomad Dynamic Host Volume Plugin for HashiCorp Nomad 1.10.0+, which supports the new Dynamic Host Volumes feature.

## Overview

This plugin enables Nomad to dynamically create and manage filesystem volumes on host machines. It creates volume images as files on the host machine and mounts them for use by Nomad jobs.

> [!IMPORTANT]
> This plugin is still actively being worked on. \
> Expect possible changes or reworks of how this plugin functions and how the config is structured.
>
> Additionally, this README isn't always up-to-date, so not everything mentioned here might work as described.

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

## Development

### Prerequisites

- Go 1.19 or higher
- HashiCorp Nomad 1.10.0 or later

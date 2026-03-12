# DHV Volume Usage Query Guide

> Note: Windows is not supported. This plugin depends on Unix-style `mkfs`, mount, and filesystem inspection tooling and is intended for Linux hosts.

本指南说明如何使用命令行直接调用 DHV 插件查询卷使用情况。

## 前言

`usage` 命令**不是** Nomad DHV 官方规范的一部分，而是本插件的扩展功能。Nomad 只会调用 `fingerprint`、`create` 和 `delete` 命令。`usage` 命令需要通过命令行手动调用或集成到监控系统中。

## 快速开始

### 方法 1：直接调用（需要设置环境变量）

```bash
# 通过卷 ID 查询
DHV_VOLUMES_DIR=/opt/nomad/volumes \
DHV_VOLUME_ID=abc123-def456-ghi789 \
./mkfs usage

# 通过直接路径查询
DHV_CREATED_PATH=/opt/nomad/volumes/abc123-def456-ghi789 \
./mkfs usage
```

### 方法 2：使用配置文件

```bash
# 1. 复制配置模板
cp examples/usage-query.env my-config.env

# 2. 编辑配置文件，设置你的卷 ID
nano my-config.env

# 3. 加载配置并运行
source my-config.env
./mkfs usage
```

### 方法 3：使用包装脚本（推荐）

```bash
# 安装脚本到系统
make install-scripts

# 使用命令查询单个卷
nomad-dhv-usage <volume-id>

# 使用管理器查看所有卷
nomad-dhv-manager list
nomad-dhv-manager usage <volume-id>
nomad-dhv-manager usage-all
```

### 方法 4：使用简单脚本

```bash
# 给脚本执行权限
chmod +x examples/quick-usage.sh

# 运行
./examples/quick-usage.sh <volume-id>
```

## 环境变量说明

### 必需参数（二选一）

**选项 A：通过卷目录 + ID**
```bash
DHV_VOLUMES_DIR=/opt/nomad/volumes  # 卷存储目录
DHV_VOLUME_ID=abc123-def456-ghi789  # Nomad 卷 ID
```

**选项 B：通过直接路径**
```bash
DHV_CREATED_PATH=/opt/nomad/volumes/abc123-def456-ghi789  # 卷挂载路径
```

### 可选参数

```bash
# 插件目录（通常不需要设置）
DHV_PLUGIN_DIR=/opt/nomad/plugins

# 节点信息（usage 命令不需要，但设置了也不影响）
DHV_NODE_ID=<node-id>
DHV_NAMESPACE=default
```

## 输出格式

### JSON 格式

```json
{
  "used_bytes": 1234567,
  "total_bytes": 524288000
}
```

### 格式化输出（需要 jq）

```bash
# 人类可读格式
./mkfs usage | jq -r '"Used: \(.used_bytes) bytes\nTotal: \(.total_bytes) bytes\nUsage: \((.used_bytes/.total_bytes*100) | round)%"'

# 输出示例：
# Used: 1234567 bytes
# Total: 524288000 bytes
# Usage: 0.24%
```

## 使用场景

### 1. 手动运维

```bash
# 检查特定卷的使用情况
nomad-dhv-manager usage <volume-id>

# 查看所有卷
nomad-dhv-manager list

# 查看卷详细信息
nomad-dhv-manager info <volume-id>
```

### 2. Shell 脚本

```bash
#!/bin/bash
VOLUME_ID="abc123-def456-ghi789"
RESULT=$(DHV_VOLUMES_DIR=/opt/nomad/volumes DHV_VOLUME_ID=$VOLUME_ID ./mkfs usage)

USED=$(echo "$RESULT" | jq -r '.used_bytes')
TOTAL=$(echo "$RESULT" | jq -r '.total_bytes')
PERCENT=$(awk "BEGIN {printf \"%.2f\", $USED/$TOTAL*100}")

if (( $(echo "$PERCENT > 80" | bc -l) )); then
    echo "WARNING: Volume $VOLUME_ID is ${PERCENT}% full"
    # 发送告警...
fi
```

### 3. Cron 定时任务

```bash
# 每 5 分钟检查一次
*/5 * * * * /usr/local/bin/volume-check.sh >> /var/log/volume-usage.log 2>&1
```

### 4. Prometheus 监控

参考 README.md 中的 Prometheus Node Exporter 示例。

### 5. Nomad Job

```hcl
job "volume-monitor" {
  type = "batch"

  periodic {
    cron             = "*/5 * * * *"
    prohibit_overlap = true
  }

  group "monitor" {
    task "check-usage" {
      driver = "exec"

      config {
        command = "/usr/local/bin/mkfs"
        args    = ["usage"]
      }

      env {
        DHV_VOLUMES_DIR = "/opt/nomad/volumes"
        DHV_VOLUME_ID   = "${NOMAD_META_volume_id}"
      }

      template {
        data = <<EOH
{{ range service "volume-monitor" }}
volume_id={{ .ID }}
{{ end }}
EOH
        destination = "local/env.txt"
        env         = true
      }
    }
  }
}
```

## 故障排查

### 错误：Volume path does not exist

```bash
# 检查卷目录是否存在
ls -la /opt/nomad/volumes/<volume-id>

# 检查环境变量是否正确
echo $DHV_VOLUMES_DIR
echo $DHV_VOLUME_ID
```

### 错误：Volume is not mounted

```bash
# 检查挂载状态
mount | grep <volume-id>

# 卷可能已被删除或未正确挂载
nomad volume status -type=host <volume-id>
```

### 错误：failed to run du

```bash
# 确保 du 命令可用
which du

# 检查权限
ls -la /opt/nomad/volumes/<volume-id>
```

### 调试模式

```bash
# 查看详细日志（插件会输出到 stderr）
DHV_VOLUMES_DIR=/opt/nomad/volumes \
DHV_VOLUME_ID=abc123 \
./mkfs usage 2>&1 | tee debug.log
```

## 最佳实践

1. **使用包装脚本** - 避免每次都手动设置环境变量
2. **集成监控系统** - 自动采集使用量数据
3. **设置告警阈值** - 在容量不足时及时通知
4. **定期审计** - 检查未使用的卷并清理
5. **使用配置文件** - 统一管理卷配置

## 与 Nomad 集成

虽然 `usage` 命令不是 Nomad 标准的一部分，但可以通过以下方式与 Nomad 集成：

### 1. 通过 Nomad Variables 存储使用量

```bash
#!/bin/bash
RESULT=$(./mkfs usage)
nomad var put -namespace=default volumes/usage/@{volume_id} \
  used_bytes=$(echo "$RESULT" | jq -r '.used_bytes') \
  total_bytes=$(echo "$RESULT" | jq -r '.total_bytes') \
  updated_at=$(date -Iseconds)
```

### 2. 通过 Nomad Job 定期采集

创建一个 periodic batch job 定期运行 usage 命令并将结果写入日志或外部系统。

### 3. 通过 HTTP API 暴露

创建一个小型 HTTP 服务包装 usage 命令，供 Nomad 或其他系统调用。

## 相关文档

- [README.md](./README.md) - 插件完整文档
- [Nomad DHV Specification](https://developer.hashicorp.com/nomad/docs/other-specifications/volume/host)
- [Nomad Volume Commands](https://developer.hashicorp.com/nomad/commands/volume)

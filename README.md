# sshbox

轻量级SSH连接管理工具，支持Tag分组和快速连接。

## 安装

```bash
go build -o sshbox .
```

或从 [Releases](https://github.com/user/sshbox/releases) 下载预编译二进制。

## 快速开始

```bash
# 添加连接
sshbox add myserver --host 192.168.1.1 --user root --password mypass

# 列出所有连接
sshbox list

# 连接服务器
sshbox connect myserver
```

## 命令参考

### sshbox add

添加SSH连接配置。

```bash
sshbox add <name> --host <host> --user <user> --password <pass> [flags]
```

| 参数 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `--host` | 是 | - | 主机地址 |
| `--port` | 否 | 22 | 端口号 |
| `--user` | 是 | - | 登录用户名 |
| `--password` | 是 | - | 登录密码 |
| `--tag` | 否 | - | 标签，多个用逗号分隔 |
| `--notes` | 否 | - | 备注信息 |

```bash
sshbox add prod-web-01 \
    --host 192.168.1.101 \
    --port 22 \
    --user root \
    --password mypass \
    --tag production,web \
    --notes "生产环境Web服务器"
```

### sshbox list

列出所有SSH连接。

```bash
sshbox list [--tag <tag>] [--search <keyword>] [--format table|json]
```

```bash
# 按标签筛选
sshbox list --tag production

# 搜索连接
sshbox list --search web

# JSON格式输出
sshbox list --format json
```

### sshbox connect

连接SSH服务器，打开交互式终端。

```bash
sshbox connect <name>
```

### sshbox edit

修改SSH连接配置。只更新指定的字段。

```bash
sshbox edit <name> [--host <host>] [--port <port>] [--user <user>] [--password <pass>] [--tag <tag>] [--notes <notes>]
```

```bash
sshbox edit prod-web-01 --host 192.168.1.200
sshbox edit prod-web-01 --password newpass
```

### sshbox rm

删除SSH连接（会提示确认）。

```bash
sshbox rm <name>
```

### sshbox show

查看连接详情（密码显示为 `********`）。

```bash
sshbox show <name>
```

输出示例：
```
名称：    prod-web-01
主机：    192.168.1.101
端口：    22
用户：    root
密码：    ********
标签：    production, web
备注：    生产环境Web服务器
创建时间：2026-06-22 10:00:00
更新时间：2026-06-22 10:00:00
```

### sshbox tags

列出所有标签及其连接数。

```bash
sshbox tags
```

### sshbox export

导出连接配置到JSON。

```bash
sshbox export                           # 输出到stdout
sshbox export --output backup.json      # 输出到文件
sshbox export --tag production          # 只导出指定标签
```

### sshbox import

从JSON文件导入连接配置。

```bash
sshbox import backup.json               # 覆盖模式
sshbox import backup.json --merge       # 合并模式（保留现有连接）
```

## 配置文件

配置文件 `config.json` 存放在可执行文件同目录下，便于迁移和备份。

目录结构：
```
sshbox/
├── sshbox          # 可执行文件（Linux/macOS）
├── sshbox.exe      # 可执行文件（Windows）
└── config.json     # 配置文件
```

## 安全

- 密码使用 AES-256-CFB 加密存储
- 主密钥存储在系统密钥环（Windows Credential Manager / macOS Keyring / Linux Secret Service）
- 配置文件权限设为 600（仅用户可读写）

## 开发

```bash
# 构建
go build -o sshbox .

# 运行测试
go test ./...

# 运行单个测试
go test -run TestAddCommand ./cmd/
```

## License

MIT

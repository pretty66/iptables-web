# iptables-web 使用说明

本文面向需要部署和日常使用 iptables-web 的管理员，介绍系统功能、安装方式、配置方法以及 Web/REST 接口的基本操作流程。

## 1. 功能概览

- **跨协议规则管理**：同时封装 `iptables` 与 `ip6tables`，在页面上可随时切换 IPv4/IPv6。
- **嵌入式 Web UI**：单一二进制内置静态界面，可查看规则、插入/追加/删除条目、刷新计数、导入导出等。
- **REST 接口**：所有页面操作都由 HTTP 接口驱动，可以通过脚本调用。
- **命令执行助手**：提供对底层 `iptables` 命令的封装，支持批量刷新、查看原始执行语句等。

> 系统仅支持 Linux，且需要具备对宿主机 iptables/ip6tables 的执行权限（通常为 root 或特权容器）。

## 2. 前置条件

| 条件              | 说明                                                           |
| ----------------- | -------------------------------------------------------------- |
| 操作系统          | Linux（内核需包含 netfilter/iptables 支持）。                  |
| 运行权限          | 需要 root 或等效权限；Docker 需要 `--privileged --net=host`。  |
| 依赖命令          | `iptables`、`iptables-save`、`iptables-restore`（IPv6 同理）。 |
| Go 环境（仅编译） | Go 1.16+（以 `go.mod` 为准，建议使用 go env 中的版本）。       |

## 3. 部署方式

### 3.1 Docker 运行（推荐）

```bash
docker run -d \
  --name iptables-web \
  --privileged=true \
  --net=host \
  -e IPT_WEB_USERNAME=admin \
  -e IPT_WEB_PASSWORD=admin \
  -e IPT_WEB_ADDRESS=:10001 \
  -p 10001:10001 \
  pretty66/iptables-web:latest
```

- `--privileged --net=host` 是为了让容器具有操作宿主机防火墙的能力。
- `IPT_WEB_ADDRESS` 默认为 `:10001`（监听全部网卡），也可以指定 `127.0.0.1:10001` 仅供本机访问。
- 镜像 tag 请根据发布版本替换，若使用非官方 registry，请自行更改。

### 3.2 二进制部署

```bash
git clone https://github.com/pretty66/iptables-web.git
cd iptables-web
make release   # 生成 build/iptables-server（需要 Go 环境）
./iptables-server -a :10001 -u admin -p admin
```

后台运行可配合 `nohup`/`systemd`/`supervisor` 等工具。若要注入构建信息，请使用默认 `Makefile` 中的 `-ldflags`。

## 4. 配置说明

| 参数       | CLI 标志 | 环境变量           | 默认值   | 说明                |
| ---------- | -------- | ------------------ | -------- | ------------------- |
| 监听地址   | `-a`     | `IPT_WEB_ADDRESS`  | `:10001` | HTTP 服务绑定地址。 |
| 登录用户名 | `-u`     | `IPT_WEB_USERNAME` | `admin`  | Basic Auth 用户名。 |
| 登录密码   | `-p`     | `IPT_WEB_PASSWORD` | `admin`  | Basic Auth 密码。   |

优先级：命令行参数 > 环境变量 > 默认值。所有接口都采用 Basic Auth 认证，请务必在生产环境中覆盖默认凭据，并通过 HTTPS/反向代理保护流量。

## 5. 运行与监控

访问 `http://<host>:10001`，浏览器会弹出 Basic Auth 弹窗。默认凭据 `admin/admin` 登录后进入页面。

- 若系统缺失 iptables/ip6tables，可在日志中看到 `exec [...] err`，需安装相应软件包。
- Docker 模式请确认宿主机的 iptables 命令可在容器中调用（通常由基础镜像提供）。

## 6. Web 界面操作指南

1. **协议切换**：页面顶部“IPv4/IPv6”单选按钮决定所有请求使用的协议，切换后会自动刷新当前表。
2. **表/链浏览**：Tab 包含 `raw/mangle/nat/filter`，点击后可查看系统链和自定义链的列表，支持展开目录快速跳转。
3. **链操作按钮**：
   - `插入`：调用 `iptables -t <table> -I <chain>`。
   - `添加`：调用 `iptables -t <table> -A <chain>`。
   - `清零计数`：执行 `-Z`。
   - `清空规则`：执行 `-F` 指定链。
   - `刷新`：重新获取该链的规则。
   - `查看命令`：显示 `iptables-save` 输出中的对应行。
4. **全局操作**（右侧浮动按钮）：
   - 清空全表/当前表规则。
   - 清空自定义空链。
   - 清零计数（全部/当前表）。
   - 查看当前表命令（`iptables-save -t <table>`）。
   - 执行任意命令（直接传递给 `iptables`/`ip6tables`）。
   - 导入/导出规则（底层使用 `iptables-save/restore`，导入时写入 0600 权限的临时文件）。

## 7. REST 接口速查

所有接口均需 Basic Auth，并接受一个可选的 `protocol` 参数（`ipv4`/`ipv6`，默认 `ipv4`）。

| 路径                     | 方法 | 参数                   | 说明                                                      |
| ------------------------ | ---- | ---------------------- | --------------------------------------------------------- |
| `/version`               | GET  | -                      | 返回当前命令版本字符串。                                  |
| `/listRule`              | POST | `table`, `chain`       | 查询链列表或单链规则。                                    |
| `/listExec`              | POST | `table`, `chain`       | 返回 `iptables-save` 输出，若指定链则只保留包含链名的行。 |
| `/flushRule`             | POST | `table`, `chain`       | 清空指定表/链，均为空则遍历全部表。                       |
| `/flushMetrics`          | POST | `table`, `chain`, `id` | 清空计数，`id` 为空表示整链/整表。                        |
| `/deleteRule`            | POST | `table`, `chain`, `id` | 删除指定序号的规则。                                      |
| `/getRuleInfo`           | POST | `table`, `chain`, `id` | 返回 `iptables-save` 中对应行。                           |
| `/flushEmptyCustomChain` | POST | -                      | 删除自定义空链。                                          |
| `/export`                | POST | `table`, `chain`       | 导出规则文本，便于备份。                                  |
| `/import`                | POST | `rule`                 | 导入规则文本（iptables-restore）。                        |
| `/exec`                  | POST | `args`                 | 直接执行 `iptables` 子命令。                              |

## 8. 常见问题

1. **提示 “ipv6 iptables not available”**：宿主机没有 `ip6tables` 或运行用户无权限；可忽略（IPv4 仍可用）或安装 `ip6tables`。
2. **规则修改未生效**：确认宿主机内核未启用 nftables/iptables 混合模式，或在命令执行前先 `iptables -L` 验证是否可用。
3. **导入失败**：检查规则内容是否包含 IPv4/IPv6 混用或模块依赖缺失（如 `conntrack`），日志会输出 `iptables-restore` 错误。

## 9. 进一步学习

- `docs/iptables-command-reference.md`：包含常用 iptables 命令及场景示例。
- `Makefile`：展示内置构建参数，可用于自定义打包。

# iptables 管理程序

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pretty66/iptables-web)](https://github.com/pretty66/iptables-web/blob/master/go.mod)

<p>
  <a href="README.md" style="padding:6px 14px;border:1px solid #1E9FFF;border-radius:4px;text-decoration:none;margin-right:10px;">中文</a>
  <a href="README_en.md" style="padding:6px 14px;border:1px solid #1E9FFF;border-radius:4px;text-decoration:none;background-color:#1E9FFF;color:#fff;">English</a>
</p>

### iptables-web 是一个轻量级的 iptables/ip6tables Web 管理平台，集成前端界面与 REST API，单二进制即可部署，适合日常运维与学习使用。

![web](./docs/iptables-web.png)

## 目录

- [功能概览](#功能概览)
- [前置条件](#前置条件)
- [安装部署](#安装部署)
  - [Docker（推荐）](#docker推荐)
  - [二进制部署](#二进制部署)
- [配置项](#配置项)
- [运行与监控](#运行与监控)
- [Web 界面操作](#web-界面操作)
- [REST 接口速查](#rest-接口速查)
- [常见问题](#常见问题)
- [附加文档](#附加文档)
- [License](#license)

## 功能概览

- **跨协议管理**：原生支持 `iptables` 与 `ip6tables`，页面/接口均可一键切换 IPv4/IPv6。
- **嵌入式 UI**：内置静态资源，无需额外 Web 服务器即可浏览链、插入/删除规则、导入导出。
- **REST API**：所有操作都暴露为 HTTP 接口，便于脚本化和二次集成。
- **命令执行助手**：在页面直接运行底层命令或查看 `iptables-save` 输出，随时校验规则。

> 注意：仅支持 Linux 系统，并需要具备对宿主机 iptables/ip6tables 的执行权限（root 或特权容器）。

## 前置条件

| 条件              | 说明                                                             |
| ----------------- | ---------------------------------------------------------------- |
| 操作系统          | Linux（内核需启用 netfilter/iptables）。                         |
| 权限              | Root 或具备 CAP_NET_ADMIN；Docker 需 `--privileged --net=host`。 |
| 依赖命令          | `iptables`、`iptables-save`、`iptables-restore`；IPv6 同理。     |
| Go 环境（构建时） | Go 1.19+（以 `go.mod` 为准）。                                   |

## 安装部署

### Docker（推荐）

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

- `--privileged --net=host` 让容器拥有修改宿主机防火墙的能力。
- `IPT_WEB_ADDRESS` 默认为 `:10001`，可改为 `127.0.0.1:10001` 限制访问范围。
- 镜像 tag 可按发布版本/自建仓库调整。

### 二进制部署

```bash
git clone https://github.com/pretty66/iptables-web.git
cd iptables-web
make release   # 需要 Go 环境
./iptables-server -a :10001 -u admin -p admin
```

后台运行可结合 `nohup`、`systemd`、`supervisor`。`Makefile` 默认通过 `-ldflags` 注入构建版本信息。

## 配置项

| 说明       | CLI 标志 | 环境变量           | 默认值   |
| ---------- | -------- | ------------------ | -------- |
| 监听地址   | `-a`     | `IPT_WEB_ADDRESS`  | `:10001` |
| 登录用户名 | `-u`     | `IPT_WEB_USERNAME` | `admin`  |
| 登录密码   | `-p`     | `IPT_WEB_PASSWORD` | `admin`  |

优先级：命令行 > 环境变量 > 默认值。所有接口使用 Basic Auth，请务必修改默认凭据，并建议在生产环境通过 HTTPS/反向代理加固。

## 运行与监控

启动成功后会输出：

```
listen address: :10001
Build Version:  <commit>  Date:  <yyyy-mm-dd hh:mm:ss>
```

访问 `http://<host>:10001`，输入 Basic Auth 凭据即可进入界面。若日志提示缺少 `ip6tables`，说明宿主机未安装对应命令，可忽略或自行安装。

## Web 界面操作

1. **协议切换**：页面顶部 “IPv4/IPv6” 单选按钮控制所有 API 使用的协议，切换会自动刷新当前表。
2. **表/链浏览**：选项卡包含 `raw/mangle/nat/filter`，点击即可查看原生链与自定义链，还可通过右侧目录快速跳转。
3. **链级操作**：
   - `插入`：执行 `iptables -t <table> -I <chain> ...`。
   - `添加`：执行 `iptables -t <table> -A <chain> ...`。
   - `清零计数`：`iptables -Z` 针对链或单条规则。
   - `清空规则`：`iptables -F <chain>`。
   - `刷新/查看命令`：重新拉取链或显示 `iptables-save` 中对应语句。
4. **全局按钮**（右侧浮动）：
   - 清空所有/当前表规则。
   - 清零所有/当前表计数。
   - 清空自定义空链。
   - 查看当前表命令、执行任意命令。
   - 导入/导出规则（底层使用 `iptables-save/restore`，导入文件以 0600 权限保存于临时目录）。

## REST 接口速查

所有接口均需 Basic Auth，可选 `protocol` 参数（`ipv4`/`ipv6`，默认 `ipv4`）。

| 路径                     | 方法 | 参数                   | 说明                                        |
| ------------------------ | ---- | ---------------------- | ------------------------------------------- |
| `/version`               | GET  | -                      | 查看底层命令版本。                          |
| `/listRule`              | POST | `table`, `chain`       | 查询链列表/单链规则。                       |
| `/listExec`              | POST | `table`, `chain`       | 返回 `iptables-save` 输出或包含指定链的行。 |
| `/flushRule`             | POST | `table`, `chain`       | 清空表/链规则，均为空则遍历所有表。         |
| `/flushMetrics`          | POST | `table`, `chain`, `id` | 清零计数；`id` 为空表示整链/整表。          |
| `/deleteRule`            | POST | `table`, `chain`, `id` | 删除指定序号的规则。                        |
| `/getRuleInfo`           | POST | `table`, `chain`, `id` | 返回 `iptables-save` 中指定规则。           |
| `/flushEmptyCustomChain` | POST | -                      | 删除所有空的自定义链。                      |
| `/export`                | POST | `table`, `chain`       | 导出规则文本。                              |
| `/import`                | POST | `rule`                 | 导入规则文本（iptables-restore）。          |
| `/exec`                  | POST | `args`                 | 直接执行命令参数。                          |

## 常见问题

1. **提示 “ipv6 iptables not available”**：宿主机缺少 `ip6tables` 或权限不足，可仅使用 IPv4。
2. **Basic Auth 弹窗反复出现**：确认访问地址正确，或检查用户名/密码是否更新。
3. **规则无效**：核对命令输出是否报错，确认未混用 nftables/iptables，必要时在宿主机直接运行命令验证。
4. **导入失败**：通常因规则格式或模块缺失导致，查看日志中的 `iptables-restore` 错误信息即可定位。

## 附加文档

- [docs/usage-guide.md](docs/usage-guide.md)：使用说明（与本 README 内容一致，可单独阅读）。
- [docs/iptables-command-reference.md](docs/iptables-command-reference.md)：iptables/ip6tables 命令与示例。

## License

iptables-web is under the Apache 2.0 license. See the [LICENSE](./LICENSE) file for details.

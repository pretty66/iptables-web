# iptables 常用命令参考

本文整理了在 Linux 环境下使用 `iptables`/`ip6tables` 的常见命令、参数说明以及典型使用场景，便于初学者查询与实践。

## 1. 基础概念

- **表（Table）**：按功能划分。常见的 `raw`（连接跟踪前）、`mangle`（数据包修改）、`nat`（地址转换）、`filter`（包过滤）。
- **链（Chain）**：数据包在每张表中的处理路径，例如 `INPUT`、`OUTPUT`、`FORWARD`、`PREROUTING`、`POSTROUTING`。自定义链可由用户创建。
- **规则（Rule）**：由匹配条件与目标动作组成，匹配顺序从上到下执行。
- **目标（Target）**：动作，例如 `ACCEPT`、`DROP`、`REJECT`、`LOG`、`SNAT`、`DNAT` 等。

IPv6 使用 `ip6tables` 命令，语法与 IPv4 基本一致，仅在地址/模块支持上存在差异。

## 2. 命令结构

```bash
iptables [-t 表] COMMAND [链] [匹配条件] [-j 目标]
```

常用全局选项：

| 参数 | 说明 |
| --- | --- |
| `-t` | 指定表，默认 `filter`。 |
| `-L` | 列出链规则，配合 `-n`（数字显示）、`-v`（统计信息）、`--line-numbers`（显示序号）。 |
| `-A` / `-I` / `-D` / `-R` | 分别表示追加、插入、删除、替换。 |
| `-F` / `-Z` / `-X` | 清空规则、清零计数、删除自定义链。 |
| `-P` | 设置链的默认策略（仅系统链）。 |
| `-j` | 指定目标动作。 |
| `-m` | 启用匹配模块，例如 `state`/`conntrack`/`limit` 等。 |

## 3. 常用命令速查

### 3.1 查看现有规则

```bash
iptables -L -n -v --line-numbers
iptables -t nat -L -n -v
ip6tables -t filter -L INPUT -n
```

### 3.2 设置默认策略

```bash
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT
```

### 3.3 允许 SSH / Web 等端口

```bash
iptables -A INPUT -p tcp --dport 22 -s 10.0.0.0/24 -m state --state NEW,ESTABLISHED -j ACCEPT
iptables -A OUTPUT -p tcp --sport 22 -m state --state ESTABLISHED -j ACCEPT
iptables -A INPUT -p tcp --dport 80 -m state --state NEW -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -m state --state NEW -j ACCEPT
```

### 3.4 拒绝或限制流量

```bash
iptables -A INPUT -p tcp --dport 25 -j REJECT --reject-with icmp-port-unreachable
iptables -A INPUT -p tcp --dport 22 -m state --state NEW -m recent --set --name SSH
iptables -A INPUT -p tcp --dport 22 -m state --state NEW -m recent --update --seconds 60 --hitcount 5 --name SSH -j DROP
```

### 3.5 NAT / 端口转发

```bash
# 内网访问 Internet，使用 SNAT
iptables -t nat -A POSTROUTING -s 192.168.0.0/24 -o eth0 -j SNAT --to-source 203.0.113.10
# 动态地址环境（例如 PPPoE）可使用 MASQUERADE
iptables -t nat -A POSTROUTING -s 10.10.0.0/16 -o ppp0 -j MASQUERADE

# DNAT：将外网流量转发至内网主机
iptables -t nat -A PREROUTING -d 203.0.113.10/32 -p tcp --dport 2222 -j DNAT --to-destination 192.168.0.10:22
iptables -A FORWARD -p tcp -d 192.168.0.10 --dport 22 -m state --state NEW,ESTABLISHED,RELATED -j ACCEPT
iptables -A FORWARD -p tcp -s 192.168.0.10 --sport 22 -m state --state ESTABLISHED -j ACCEPT
```

### 3.6 透明代理 / 端口劫持

```bash
# 将 80 端口流量转交给本机 8080（如 HTTP 代理）
iptables -t nat -A PREROUTING -i eth0 -p tcp --dport 80 -j REDIRECT --to-ports 8080
```

### 3.7 记录与速率限制

```bash
iptables -A INPUT -p tcp --dport 22 -m limit --limit 3/min -j LOG --log-prefix "SSH attempt: "
iptables -A INPUT -p icmp -m limit --limit 1/s --limit-burst 5 -j ACCEPT
iptables -A INPUT -p icmp -j DROP
```

### 3.8 保存与恢复

```bash
iptables-save > /etc/iptables/rules.v4
ip6tables-save > /etc/iptables/rules.v6
iptables-restore < /etc/iptables/rules.v4
ip6tables-restore < /etc/iptables/rules.v6
```

## 4. 场景示例

### 场景 A：只允许特定网段访问 SSH

```bash
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT
iptables -A INPUT -i lo -j ACCEPT
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -A INPUT -p tcp --dport 22 -s 192.168.1.0/24 -m state --state NEW -j ACCEPT
iptables -A INPUT -j LOG --log-prefix "DROP INPUT: "
```

### 场景 B：双网卡网关的 SNAT 与防火墙

```bash
# 开启内核转发
sysctl -w net.ipv4.ip_forward=1

# NAT 公网出口
iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE

# 转发策略
iptables -A FORWARD -i eth1 -o eth0 -m state --state NEW,ESTABLISHED,RELATED -j ACCEPT
iptables -A FORWARD -i eth0 -o eth1 -m state --state ESTABLISHED,RELATED -j ACCEPT

# 阻断外网主动访问内网
iptables -A FORWARD -i eth0 -o eth1 -j DROP
```

### 场景 C：IPv6 入站仅开放 80/443

```bash
ip6tables -P INPUT DROP
ip6tables -P FORWARD DROP
ip6tables -P OUTPUT ACCEPT
ip6tables -A INPUT -i lo -j ACCEPT
ip6tables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
ip6tables -A INPUT -p tcp --dport 80 -j ACCEPT
ip6tables -A INPUT -p tcp --dport 443 -j ACCEPT
```

### 场景 D：限制单 IP 并发连接数

```bash
iptables -A INPUT -p tcp --syn --dport 80 -m connlimit --connlimit-above 50 --connlimit-mask 32 -j REJECT
```

## 5. 故障排查建议

1. **规则与 nftables 冲突**：在使用新内核时确认是否启用了 `iptables-nft`，可通过 `iptables -V` 查看，必要时安装 `iptables-legacy`。
2. **模块不可用**：例如 `-m conntrack` 提示错误，需加载 `nf_conntrack` 模块或安装对应包。
3. **顺序问题**：iptables 规则按顺序匹配，建议使用 `--line-numbers` 查看并使用 `-I`/`-R` 调整。
4. **调试**：利用 `LOG` 目标输出到 `dmesg`/`/var/log/messages`，协助定位被丢弃的数据包。
5. **持久化**：在系统重启后需重新加载规则，建议配合 `iptables-save` + `systemd`/`/etc/rc.local` 或发行版提供的 `netfilter-persistent`。

## 6. 进阶阅读

- `man iptables` / `man ip6tables`
- `man iptables-extensions`
- Netfilter.org 官方文档

结合 iptables-web，可以在图形界面中执行上述命令、观察效果并快速导出/导入规则，降低学习成本。

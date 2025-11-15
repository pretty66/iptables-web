# iptables Web Management
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pretty66/iptables-web)](https://github.com/pretty66/iptables-web/blob/master/go.mod)

### iptables-web is a lightweight management console for both `iptables` and `ip6tables`. It bundles a UI, REST API, and utilities into a single binary that fits daily operations as well as learning scenarios.

![web](./docs/iptables-web.png)

## Table of Contents
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [Docker (recommended)](#docker-recommended)
  - [Binary](#binary)
- [Configuration](#configuration)
- [Running & Monitoring](#running--monitoring)
- [Web UI Guide](#web-ui-guide)
- [REST API](#rest-api)
- [FAQ](#faq)
- [Additional Docs](#additional-docs)
- [License](#license)

## Features

- **Dual-stack management** – first-class support for `iptables` and `ip6tables`; switch protocols via UI/REST.
- **Embedded UI** – static assets are baked into the binary, allowing rule browsing, insertion/deletion, and import/export without extra servers.
- **REST-first** – every UI action has an HTTP endpoint for automation and integration.
- **Command helper** – execute raw commands or inspect `iptables-save` results directly in the console.

> Only Linux is supported. The service must run with privileges that allow managing the host firewall (root or an equivalent capability; Docker requires privileged mode).

## Prerequisites

| Requirement | Description |
| --- | --- |
| OS | Linux with netfilter/iptables enabled. |
| Privileges | Root or `CAP_NET_ADMIN`; Docker needs `--privileged --net=host`. |
| Runtime deps | `iptables`, `iptables-save`, `iptables-restore` (IPv6 counterparts included automatically). |
| Go toolchain (build only) | Go 1.19+ (follow `go.mod`). |

## Installation

### Docker (recommended)

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

- `--privileged --net=host` lets the container manipulate the host firewall.
- `IPT_WEB_ADDRESS` defaults to `:10001`; change to `127.0.0.1:10001` to limit exposure.
- Swap the image tag to match your release or registry.

### Binary

```bash
git clone https://github.com/pretty66/iptables-web.git
cd iptables-web
make release   # requires Go
./iptables-server -a :10001 -u admin -p admin
```

Use `nohup`, `systemd`, or `supervisor` to keep it in the background. The default `Makefile` injects build metadata through `-ldflags`.

## Configuration

| Description | CLI flag | Env | Default |
| --- | --- | --- | --- |
| Listen address | `-a` | `IPT_WEB_ADDRESS` | `:10001` |
| Username | `-u` | `IPT_WEB_USERNAME` | `admin` |
| Password | `-p` | `IPT_WEB_PASSWORD` | `admin` |

Priority: CLI > env vars > defaults. Since the service uses Basic Auth for every endpoint, change the credentials in production and place it behind HTTPS/a reverse proxy if possible.

## Running & Monitoring

On startup you should see:

```
listen address: :10001
Build Version:  <commit>  Date:  <yyyy-mm-dd hh:mm:ss>
```

Open `http://<host>:10001` and authenticate via Basic Auth. If the log reports missing `ip6tables`, the binary is absent on the host—install it or operate in IPv4-only mode.

## Web UI Guide

1. **Protocol switch** – radio buttons (IPv4/IPv6) at the top decide which backend to call; switching refreshes the current table.
2. **Tables/chains** – tabs cover `raw/mangle/nat/filter`; visualize system/custom chains and navigate via the right-side directory.
3. **Chain actions** – insert (`-I`), append (`-A`), zero counters (`-Z`), flush (`-F`), refresh, and view raw commands (`iptables-save` snippet).
4. **Global actions** – clear rules or counters (all/current table), delete empty custom chains, inspect current table output, run arbitrary commands, and import/export rules (save/restore with temporary files created using mode 0600).

## REST API

All endpoints require Basic Auth. Optional `protocol` can be `ipv4` (default) or `ipv6`.

| Path | Method | Params | Description |
| --- | --- | --- | --- |
| `/version` | GET | - | Underlying binary version string. |
| `/listRule` | POST | `table`, `chain` | List chains or a single chain's rules. |
| `/listExec` | POST | `table`, `chain` | Return `iptables-save` output or lines containing the chain. |
| `/flushRule` | POST | `table`, `chain` | Flush the specified table/chain; empty values flush every table. |
| `/flushMetrics` | POST | `table`, `chain`, `id` | Reset counters for a rule, chain, or entire table. |
| `/deleteRule` | POST | `table`, `chain`, `id` | Delete a rule by its line number. |
| `/getRuleInfo` | POST | `table`, `chain`, `id` | Fetch the `iptables-save` line for the rule. |
| `/flushEmptyCustomChain` | POST | - | Remove all empty custom chains. |
| `/export` | POST | `table`, `chain` | Export rules as text. |
| `/import` | POST | `rule` | Import rule text via `iptables-restore`. |
| `/exec` | POST | `args` | Execute arbitrary `iptables` arguments. |

## FAQ

1. **“ipv6 iptables not available”** – the host lacks `ip6tables` or privileges. IPv4 still works; install `ip6tables` if needed.
2. **Authentication prompt loops** – double-check the URL and credentials.
3. **Rules not applied** – run the command directly on the host to confirm there are no nftables conflicts or syntax errors.
4. **Import failures** – inspect logs for `iptables-restore` errors (often module dependencies or IPv4/IPv6 mismatch).

## Additional Docs

- [docs/usage-guide.md](docs/usage-guide.md) – Chinese usage guide (mirrors this README).
- [docs/iptables-command-reference.md](docs/iptables-command-reference.md) – iptables/ip6tables command cheatsheet.
- [docs/dev-plan.md](docs/dev-plan.md) – development workflow.

## License

iptables-web is released under the Apache 2.0 License. See [LICENSE](./LICENSE) for details.

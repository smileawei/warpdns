# warpdns

将本地 UDP DNS 查询转换并转发到自定义 DoH (DNS-over-HTTPS) 上游的轻量代理。
A lightweight proxy that converts local UDP DNS queries into custom DNS-over-HTTPS (DoH) requests.

---

## 中文

### 简介
`warpdns` 监听本地 UDP DNS 端口,把每一条查询封装成 RFC 8484 的 DoH 请求转发到任意自定义上游。常用于:
- 给老程序 / 系统接入只支持 DoH 的解析服务
- 自建 DoH 网关的客户端入口
- 注入固定 EDNS Client Subnet 影响上游 GeoDNS 解析结果

### 特性
- ✅ 完全自定义 DoH 上游 (URL / URI 路径 / 方法 / Header / SNI)
- ✅ 支持 `POST application/dns-message` 与 `GET ?dns=base64url` 两种方式
- ✅ 可配置注入 EDNS Client Subnet (ECS)
- ✅ TOML 配置文件
- ✅ 单二进制,基于 distroless 的最小镜像 (~10MB),以 `nonroot` 用户运行

### 镜像
预构建镜像每次推送 `main` 与 tag 都会发布到:
- Docker Hub: [`smileawei/warpdns`](https://hub.docker.com/r/smileawei/warpdns)
- GHCR: [`ghcr.io/smileawei/warpdns`](https://github.com/smileawei/warpdns/pkgs/container/warpdns)

可用 tag: `latest`、`v0.1.0` 等版本号、`main`、`sha-xxxxxxx`。

### 快速开始 — Docker Compose
```bash
# 只需要这两个文件
curl -O https://raw.githubusercontent.com/smileawei/warpdns/main/docker-compose.yml
curl -o config.toml https://raw.githubusercontent.com/smileawei/warpdns/main/config.example.toml

# 编辑 config.toml 改为你的 DoH 上游
docker compose up -d
docker compose logs -f
```
验证:
```bash
dig @127.0.0.1 -p 1053 example.com
```

### 快速开始 — Docker
```bash
docker run -d --name warpdns \
  -p 1053:1053/udp \
  -v $(pwd)/config.toml:/etc/warpdns/config.toml:ro \
  --restart unless-stopped \
  smileawei/warpdns:latest
```

### 本地构建(开发者)
```bash
git clone git@github.com:smileawei/warpdns.git
cd warpdns
docker build -t warpdns:dev .
```

### 配置说明 (`config.toml`)
| 字段 | 说明 | 默认值 |
| --- | --- | --- |
| `listen` | UDP 监听地址 | `0.0.0.0:1053` |
| `upstream.url` | DoH 基址 (scheme + host[:port]) | 必填 |
| `upstream.path` | DoH URI 路径 | `/dns-query` |
| `upstream.method` | `POST` 或 `GET` | `POST` |
| `upstream.timeout` | 请求超时 | `10s` |
| `upstream.headers` | 任意自定义 HTTP Header | 无 |
| `upstream.insecure_skip_verify` | 跳过 TLS 校验 (生产慎用) | `false` |
| `upstream.server_name` | TLS SNI 覆盖 | 取自 url |
| `ecs.enabled` | 是否注入 EDNS Client Subnet | `false` |
| `ecs.subnet` | 注入的 CIDR,如 `1.2.3.0/24` | — |

完整示例见 `config.example.toml`。

### 监听特权端口 (53)
默认 `1053` 是为了在不需要 root 的情况下运行。若要监听 53:
- 在容器外做端口映射 `-p 53:1053/udp` (推荐)
- 或修改 `config.toml` 中 `listen = "0.0.0.0:53"` 并让容器使用 `cap_add: [NET_BIND_SERVICE]`

---

## English

### Overview
`warpdns` listens on a local UDP DNS port and forwards every query as an RFC 8484 DoH request to a fully-custom upstream. Useful for:
- Letting legacy software / OSes talk to DoH-only resolvers
- Acting as a client gateway in front of your own DoH endpoint
- Pinning an EDNS Client Subnet to influence upstream GeoDNS answers

### Features
- ✅ Fully customizable DoH upstream (URL / URI path / method / headers / SNI)
- ✅ Both `POST application/dns-message` and `GET ?dns=base64url` modes
- ✅ Configurable EDNS Client Subnet injection
- ✅ TOML config
- ✅ Single static binary on a distroless image (~10MB), running as `nonroot`

### Images
Prebuilt images are published on every push to `main` and on tags:
- Docker Hub: [`smileawei/warpdns`](https://hub.docker.com/r/smileawei/warpdns)
- GHCR: [`ghcr.io/smileawei/warpdns`](https://github.com/smileawei/warpdns/pkgs/container/warpdns)

Available tags: `latest`, semver like `v0.1.0`, `main`, `sha-xxxxxxx`.

### Quick start — Docker Compose
```bash
# Only these two files are needed
curl -O https://raw.githubusercontent.com/smileawei/warpdns/main/docker-compose.yml
curl -o config.toml https://raw.githubusercontent.com/smileawei/warpdns/main/config.example.toml

# Edit config.toml to point at your DoH upstream
docker compose up -d
docker compose logs -f
```
Test:
```bash
dig @127.0.0.1 -p 1053 example.com
```

### Quick start — Docker
```bash
docker run -d --name warpdns \
  -p 1053:1053/udp \
  -v $(pwd)/config.toml:/etc/warpdns/config.toml:ro \
  --restart unless-stopped \
  smileawei/warpdns:latest
```

### Build locally (developers)
```bash
git clone git@github.com:smileawei/warpdns.git
cd warpdns
docker build -t warpdns:dev .
```

### Configuration (`config.toml`)
| Field | Description | Default |
| --- | --- | --- |
| `listen` | UDP listen address | `0.0.0.0:1053` |
| `upstream.url` | DoH base URL (scheme + host[:port]) | required |
| `upstream.path` | DoH URI path | `/dns-query` |
| `upstream.method` | `POST` or `GET` | `POST` |
| `upstream.timeout` | Request timeout | `10s` |
| `upstream.headers` | Arbitrary HTTP headers | none |
| `upstream.insecure_skip_verify` | Skip TLS verification (use with care) | `false` |
| `upstream.server_name` | TLS SNI override | from url |
| `ecs.enabled` | Inject EDNS Client Subnet | `false` |
| `ecs.subnet` | CIDR to inject, e.g. `1.2.3.0/24` | — |

See `config.example.toml` for a full template.

### Binding privileged port 53
The default of `1053` lets the binary run without root. To use port 53:
- Map the port outside the container: `-p 53:1053/udp` (recommended)
- Or set `listen = "0.0.0.0:53"` in `config.toml` and grant the container `cap_add: [NET_BIND_SERVICE]`

---

## License
MIT

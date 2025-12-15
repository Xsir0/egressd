# egressd

> A lightweight, configurable HTTP egress proxy with access control.

`egressd` 是一个用 Go 编写的 **通用 HTTP 出网代理服务**，用于在受限网络环境中，安全、可控地转发 HTTP/HTTPS 请求。  

---

## ✨ Features

- 🚪 HTTP 出网转发（Relay / Forward）
- 🔒 来源 IP / CIDR 白名单控制
- 🌐 目标 Host 白名单
- 📦 请求 Body 大小限制
- ⚙️ YAML 配置文件
- 🧱 Middleware 架构，易扩展
- 🚀 高并发、低资源占用（Go 原生 HTTP）

---

## 📦 Typical Use Cases

- 在 **无法直接访问外网** 的环境中提供统一出网入口
- 为内部服务提供 **受控的 HTTP 转发能力**
- 作为 AI / API 请求的 **网络中继层**
- 构建轻量级的 **Egress Gateway**

---

## 📁 Project Structure

```text
.
├── cmd
│   └── main.go              # 程序入口
├── config.yaml              # 示例配置文件
├── go.mod
├── go.sum
└── internal
    ├── config               # 配置加载与校验
    │   └── config.go
    └── httpserver           # HTTP Server & Middleware
        ├── middleware.go
        ├── middleware_host.go
        ├── middleware_ip.go
        ├── proxy_handler.go
        └── upstream
            └── upstream.go  # 上游请求转发逻辑
```
---

## ⚙️ Configuration

示例 config.yaml

```yaml
listen_addr: "0.0.0.0:8080"
log_level: "info"

# 最大请求体大小
max_body_size: "10MB"

# 上游请求超时（秒）
upstream_timeout_seconds: 30

# 允许访问的目标 Host 白名单
allowed_hosts:
  - api.openai.com
  - generativelanguage.googleapis.com

# 允许访问的来源 IP / CIDR
allowed_ips:
  - 127.0.0.1
  - 192.168.0.0/16
```

---

## 🚀 Getting Started

```bash
1️⃣ Build

go build -o egressd ./cmd

2️⃣ Run

./egressd -config ./config.yaml
```

---

## 🔁 Request Flow

客户端请求 → egressd → 校验 → 转发 → 返回响应

```text
Client
  |
  |  HTTP Request
  v
egressd
  ├─ IP Access Control
  ├─ Host Whitelist
  ├─ Body Size Limit
  └─ Forward to Upstream
        |
        v
     Target Host
```

---

## 🔐 Access Control

来源 IP / CIDR

支持以下格式：
```yaml
allowed_ips:
  - 10.0.0.1
  - 10.0.0.0/8
```
请求将基于以下顺序解析客户端 IP：
1. X-Forwarded-For
2. X-Real-IP
3. RemoteAddr

---

目标 Host 白名单

仅允许访问明确声明的目标 Host：
```yaml
allowed_hosts:
  - example.com
```
防止被用作 开放代理（Open Proxy）。

---
## 📦 Body Size Limit

支持人类可读格式：
```yaml
max_body_size: "100KB"
max_body_size: "10MB"
max_body_size: "1.5GB"
```

内部统一转换为字节数后进行限制。

---

## 🧱 Middleware Design

egressd 使用标准 Go middleware 模式：

```go
type Middleware func(http.Handler) http.Handler
```

你可以非常容易地扩展：
	•	认证（API Key / Token）
	•	限流（Rate Limit）
	•	审计日志（Audit Log）
	•	Metrics（Prometheus）

---


## 🛣 Roadmap
	•	Request / Response Logging
	•	Rate Limiting
	•	Authentication Middleware
	•	Metrics / Observability
	•	TLS Termination
	•	HTTP/2 / HTTP/3 Support

---

## 🤝 Contributing

欢迎 issue / PR。
请保持代码风格简洁、职责清晰。

⸻

## 📄 License

MIT License


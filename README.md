# go-proxy-pool

一个面向 Decodo residential proxy 的 Go pkg，负责：

- 建模 Decodo `user:pass` backconnect 配置
- 生成可直接用于 `httpcloak`、`net/http` 和 `SOCKS5` 的代理 URL
- 管理按业务 key 复用的 sticky session
- 支持 100+ 国家、8 个城市、50 个美国州的端点预设

## 安装

```bash
go env -w GOPRIVATE=github.com/VectorSprint/*
go get github.com/VectorSprint/go-proxy-pool@latest
```

## 导入路径

```go
import "github.com/VectorSprint/go-proxy-pool/pkg/decodo"

import httpcloakadapter "github.com/VectorSprint/go-proxy-pool/pkg/decodo/adapter/httpcloak"
import nethttpadapter "github.com/VectorSprint/go-proxy-pool/pkg/decodo/adapter/nethttp"
```

## 快速开始

```go
auth, err := decodo.NewAuth("my-proxy-user", "my-proxy-password")
if err != nil {
  return err
}

cfg := decodo.Config{
  Auth: auth,
  Targeting: decodo.Targeting{
    Country: "us",
    City:    "new_york",
  },
  Session: decodo.Session{
    Type:            decodo.SessionTypeSticky,
    ID:              "account-1",
    DurationMinutes: 30,
  },
}

proxyURL, err := cfg.ProxyURL()
if err != nil {
  return err
}
```

生成结果：

```text
http://user-my-proxy-user-country-us-city-new_york-session-account-1-sessionduration-30:my-proxy-password@gate.decodo.com:7000
```

## HTTP/HTTPS 代理接入

### httpcloak

```go
proxy, err := httpcloakadapter.ProxyString(cfg)
if err != nil {
  return err
}

client := client.NewClient("chrome-latest")
client.SetProxy(proxy)
```

### net/http

```go
proxyFunc, err := nethttpadapter.ProxyFunc(cfg)
if err != nil {
  return err
}

transport := &http.Transport{
  Proxy: proxyFunc,
}

httpClient := &http.Client{
  Transport: transport,
}
```

## SOCKS5 代理接入

SOCKS5 使用 `gate.decodo.com:7000`，地理位置通过 username 参数指定。

### httpcloak + SOCKS5

```go
proxy, err := httpcloakadapter.ProxyStringSOCKS5(cfg)
if err != nil {
  return err
}

// socks5h://user-my-proxy-user-country-us-city-new_york-session-account-1-sessionduration-30:my-proxy-password@gate.decodo.com:7000
```

### net/http + SOCKS5

net/http 原生不支持 SOCKS5，需要配合第三方 dialer 使用：

```go
proxyURL, err := nethttpadapter.ProxyURLSOCKS5(cfg)
if err != nil {
  return err
}

// proxyURL 可用于 golang.org/x/net/proxy
```

## 端点预设 (Endpoint Presets)

内置 100+ 国家、8 个城市、50 个美国州的端点配置，自动根据 targeting 选择正确的端点和端口。

### 自动应用预设

```go
cfg := decodo.Config{
  Auth: auth,
  Targeting: decodo.Targeting{
    Country: "us",
    City:    "new_york",
  },
  Session: decodo.Session{
    Type:            decodo.SessionTypeSticky,
    DurationMinutes: 30,
  },
}

// 自动选择 city.decodo.com:21000 和 sticky port range
cfg.ApplyPreset()
```

### 支持的 targeting 级别

| 级别 | 示例 | 端点 |
|------|------|------|
| 国家 | `Country: "us"` | us.decodo.com |
| 城市 | `City: "new_york"` | city.decodo.com |
| 美国州 | `State: "us_california"` | state.decodo.com |

支持的端点详情见 [Decodo 文档](https://help.decodo.com/docs/residential-proxy-endpoints-and-ports)。

## Sticky Session Pool

```go
pool, err := decodo.NewPool(decodo.PoolOptions{
  Config: decodo.Config{
    Auth: auth,
    Session: decodo.Session{
      Type:            decodo.SessionTypeSticky,
      DurationMinutes: 30,
    },
  },
  FailureThreshold: 3,
})
if err != nil {
  return err
}

lease, err := pool.Get("account-1")
if err != nil {
  return err
}

proxy := lease.ProxyURL
```

### 随机端口分配

启用 `RandomPort` 可以在 sticky port range 内随机选择端口，降低被检测风险：

```go
pool, err := decodo.NewPool(decodo.PoolOptions{
  Config: decodo.Config{
    Auth:         auth,
    EndpointSpec: caEndpoint, // ca.decodo.com:20001-29999
    Session: decodo.Session{
      Type:            decodo.SessionTypeSticky,
      DurationMinutes: 30,
    },
  },
  RandomPort:       true,
  Rand:             rand.New(rand.NewSource(time.Now().UnixNano())),
  FailureThreshold: 3,
})
```

### 强制换 IP

```go
if err := pool.Rotate("account-1"); err != nil {
  return err
}
```

### 失败驱动轮换

```go
if err := pool.ReportFailure("account-1", decodo.FailureCause{
  StatusCode: 429,
}); err != nil {
  return err
}
```

## username 和 password 填法

### username

填 **Decodo dashboard 里的原始 proxy user**，不要填 `user-` 前缀：

```
my-proxy-user        ✓ 正确
user-my-proxy-user   ✗ 错误
```

### password

填 **原始密码**，不要做任何拼接。

## Dedicated Endpoint

支持 Decodo dedicated endpoint 配置：

```go
caEndpoint, err := decodo.NewEndpointSpec("ca.decodo.com", 20000, decodo.PortRange{
  Start: 20001,
  End:   29999,
})
if err != nil {
  return err
}

cfg := decodo.Config{
  Auth:         auth,
  EndpointSpec: caEndpoint,
}
```

选择规则：

- rotating session：自动使用 `RotatingPort`
- sticky session：自动使用 `StickyPortRange.Start`
- sticky pool：在范围内为不同 key 分配可用 sticky port

## examples 目录

- `examples/nethttp-basic`：标准库 `net/http` 接入
- `examples/pool-basic`：sticky session pool 的获取与失败轮换
- `examples/httpcloak-proxy-string`：为 `httpcloak` 生成可直接 `SetProxy(...)` 的字符串
- `examples/dedicated-endpoint`：使用 Decodo dedicated endpoint
- `examples/random-port`：随机端口分配和端点预设示例

## 开发检查

```bash
task test
task lint
task check
```

## 测试覆盖率

| Package | Coverage |
|---------|----------|
| `pkg/decodo` | 84.4% |
| `pkg/decodo/adapter/httpcloak` | 84.0% |
| `pkg/decodo/adapter/nethttp` | 86.4% |

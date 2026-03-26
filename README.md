# go-proxy-pool

一个面向 Decodo residential proxy 的 Go pkg，负责：

- 建模 Decodo `user:pass` backconnect 配置
- 生成可直接用于 `httpcloak` 和 `net/http` 的代理 URL
- 管理按业务 key 复用的 sticky session

## 当前标准化状态

当前项目已经具备这些常见 Go 项目要素：

- 正式 module path：`github.com/VectorSprint/go-proxy-pool`
- `CHANGELOG.md`
- `README.md`
- MIT `LICENSE`
- `.gitignore`
- `Taskfile.yml`
- 基础 GitHub Actions 测试 workflow
- 基础 GitHub Actions lint workflow
- 仓库级 `.golangci.yml` lint 配置
- 导出 API 的 Go doc 注释
- 可运行示例与单元测试

## 安装

私有仓库场景下，建议先配置：

```bash
go env -w GOPRIVATE=github.com/VectorSprint/*
```

然后在其他项目中拉取：

```bash
go get github.com/VectorSprint/go-proxy-pool@latest
```

## 开发检查

本仓库使用仓库根目录下的 `.golangci.yml` 作为统一 lint 配置，并通过 `Taskfile.yml` 暴露本地开发命令。

先按 <https://taskfile.dev/> 安装 `task`，然后在仓库根目录执行：

```bash
task test
task lint
task check
```

其中 `task check` 会串行执行测试和 lint。

GitHub Actions 的 `lint` workflow 仍然使用同一套仓库配置。

## 导入路径

主入口：

```go
import "github.com/VectorSprint/go-proxy-pool/pkg/decodo"
```

适配器：

```go
import httpcloakadapter "github.com/VectorSprint/go-proxy-pool/pkg/decodo/adapter/httpcloak"
import nethttpadapter "github.com/VectorSprint/go-proxy-pool/pkg/decodo/adapter/nethttp"
```

## username 和 password 应该怎么填

### username

这里填的是 **Decodo dashboard 里看到的原始 proxy user**。

你应该填：

```text
my-proxy-user
```

你不应该填：

```text
user-my-proxy-user
user-my-proxy-user-country-us
```

原因是：

- `user-` 前缀由这个 pkg 自动加上
- `country`、`city`、`session`、`sessionduration` 等参数由 `Targeting` 和 `Session` 自动拼装

### password

这里填的是 **Decodo dashboard 里对应 proxy user 的原始密码**，不要做任何拼接。

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

生成结果类似：

```text
http://user-my-proxy-user-country-us-city-new_york-session-account-1-sessionduration-30:my-proxy-password@gate.decodo.com:7000
```

## 接入 httpcloak

```go
proxy, err := httpcloakadapter.ProxyString(cfg)
if err != nil {
  return err
}

client := client.NewClient("chrome-latest")
client.SetProxy(proxy)
```

## 接入 net/http

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

## Sticky session pool

```go
pool, err := decodo.NewPool(decodo.PoolOptions{
  Config: decodo.Config{
    Auth: auth,
    Session: decodo.Session{
      Type:            decodo.SessionTypeSticky,
      DurationMinutes: 30,
    },
  },
  FailureThreshold: 2,
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

如果同一个 key 需要强制换 IP：

```go
if err := pool.Rotate("account-1"); err != nil {
  return err
}
```

如果上层请求失败并希望按阈值轮换：

```go
if err := pool.ReportFailure("account-1", decodo.FailureCause{
  StatusCode: 429,
}); err != nil {
  return err
}
```

## 当前边界

- 当前以 Decodo `user:pass` backconnect 为主
- 默认端点为 `gate.decodo.com:7000`
- 失败驱动轮换由上层显式调用 `ReportFailure`
- 还没有接管真实 HTTP 请求执行

## examples 目录

仓库里提供了几个更接近真实使用的示例：

- `examples/nethttp-basic`：标准库 `net/http` 接入
- `examples/pool-basic`：sticky session pool 的获取与失败轮换
- `examples/httpcloak-proxy-string`：为 `httpcloak` 生成可直接 `SetProxy(...)` 的字符串
- `examples/dedicated-endpoint`：使用 Decodo dedicated endpoint、rotating port 和 sticky port range

## endpoint 和 port 选择

当前项目除了默认的 `gate.decodo.com:7000` backconnect 方案外，也支持显式建模 Decodo 文档里的 dedicated endpoint：

```go
caEndpoint, err := decodo.NewEndpointSpec("ca.decodo.com", 20000, decodo.PortRange{
  Start: 20001,
  End:   29999,
})
if err != nil {
  return err
}
```

然后接入到配置里：

```go
cfg := decodo.Config{
  Auth:         auth,
  EndpointSpec: caEndpoint,
}
```

选择规则如下：

- rotating session：如果没显式指定 `Port`，自动使用 `RotatingPort`
- sticky session：如果没显式指定 `Port`，自动使用 `StickyPortRange.Start`
- sticky pool：如果配置了 `StickyPortRange`，pool 会在范围内为不同 key 分配可用 sticky port
- 如果你显式设置了 `Config.Port`，则优先使用该端口

这意味着你现在可以表达：

- 默认 backconnect：`gate.decodo.com:7000`
- dedicated rotating endpoint：例如 `ca.decodo.com:20000`
- dedicated sticky endpoint：例如 `ca.decodo.com:20001-29999` 范围中的 sticky 端口

## Go docs

当前已经补齐导出 API 的 Go doc 注释。

本地可以这样看：

```bash
go doc github.com/VectorSprint/go-proxy-pool/pkg/decodo
go doc github.com/VectorSprint/go-proxy-pool/pkg/decodo/adapter/httpcloak
go doc github.com/VectorSprint/go-proxy-pool/pkg/decodo/adapter/nethttp
```

如果未来仓库改为公开仓库，也可以直接使用 `pkg.go.dev` 查看文档。

## 发布前建议

当前正式 module path 为：

```text
github.com/VectorSprint/go-proxy-pool
```

建议用语义化版本发布，例如：

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin main --tags
```

首个版本建议包含：

- 当前 `decodo` 主包
- `httpcloak` / `net/http` adapter
- examples
- `CHANGELOG.md`

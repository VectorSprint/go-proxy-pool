# go-proxy-pool

一个面向 Decodo residential proxy 的 Go pkg，负责：

- 建模 Decodo `user:pass` backconnect 配置
- 生成可直接用于 `httpcloak` 和 `net/http` 的代理 URL
- 管理按业务 key 复用的 sticky session

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

## 发布前建议

当前正式 module path 为：

```text
github.com/VectorSprint/go-proxy-pool
```

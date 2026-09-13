**语言:** [English](README.md) | 简体中文

# Airymax Go SDK

[![Version](https://img.shields.io/badge/version-0.1.1-5a6b7e)](https://atomgit.com/openairymax/sdk-go)
[![License](https://img.shields.io/badge/license-AGPL--3.0+Apache--2.0-4a90d9)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)

> [Airymax](https://atomgit.com/openairymax) AI 智能体运行时平台的官方 Go 开发工具包。
> [sdk](https://atomgit.com/openairymax/sdk) 管理仓聚合的叶子仓之一。

---

## 概述

**Airymax Go SDK**（导入路径 `github.com/spharx/agentrt/sdk/go/agentrt`，包名 `agentrt`）提供符合 Go 惯用法的 Airymax 运行时接口。它与其他语言 SDK 共享相同的 API 架构，并基于 Go 标准库实现 HTTP/JSON 栈，因此**零外部依赖**。并发通过 goroutine 与 `context.Context` 表达，客户端暴露类型化错误码体系（`agentrt/errors.go`）。

基于该 SDK 构建的 Agent 应用是**运行时租户**：通过 SDK 调用系统能力，而非直接访问运行时内部。

## API 架构

Go SDK 遵循统一的双层结构：

- **HTTP 客户端层** —— `client.Client` 实现 `client.APIClient` 接口（`Get` / `Post` / `Put` / `Delete`），带连接池、指数退避 + 抖动的重试和 Bearer 令牌认证。同时提供 `Health` 与 `Metrics` 探针，以及测试用 `MockClient`。
- **模块管理器层** —— 四个业务域管理器，分别覆盖运行时的一个资源面：

```
client.NewClient(...)                    # HTTP 客户端（client.APIClient）
├── task.NewTaskManager(api)             # 任务生命周期：提交 / 等待 / 取消 / 列表
├── memory.NewMemoryManager(api)         # 记忆：写入 / 检索 / 更新 / 分层
├── session.NewSessionManager(api)       # 会话：创建 / 上下文 KV / 关闭
└── skill.NewSkillManager(api)           # 技能：加载 / 执行 / 注册 / 检索
```

## 目录结构

```
sdk-go/
├── agentrt/
│   ├── agentrt.go              # Version / Author / License 常量
│   ├── config.go               # Config / ConfigOption / 环境变量加载
│   ├── protocol.go             # 协议处理
│   ├── errors.go               # 错误类型与错误码常量
│   ├── client/
│   │   ├── client.go           # APIClient / Client / Health / Metrics
│   │   └── mock.go             # 测试用 MockClient
│   ├── modules/                # 业务模块管理器
│   │   ├── modules.go          # 模块导出
│   │   ├── base_manager.go     # 泛型 BaseManager + 资源转换器
│   │   ├── task/               # TaskManager（含 benchmark 测试）
│   │   ├── memory/             # MemoryManager
│   │   ├── session/            # SessionManager
│   │   └── skill/              # SkillManager
│   ├── plugin/                 # 插件系统
│   ├── syscall/                # 系统调用绑定
│   ├── telemetry/              # 内置分布式追踪（Tracer / Span）
│   ├── types/                  # 枚举与领域模型
│   └── utils/                  # 工具函数
├── go.mod                      # Go 模块：github.com/spharx/agentrt/sdk/go
└── README.md                   # 本文件
```

## 前置条件

SDK 是运行中 AgentRT 运行时的客户端，请先安装运行时：

```bash
curl -fsSL "https://api.atomgit.com/api/v5/repos/openairymax/agentrt/contents/scripts/install.sh?ref=main" | python3 -c 'import json,sys,base64;sys.stdout.buffer.write(base64.b64decode(json.load(sys.stdin)["content"]))' | bash
```

（另有兼容入口 `https://atomgit.com/openairymax/agentrt/releases/download/latest/install.sh`。）随后启动网关，例如 `airymaxrt start`。

### 端点解析

端点按以下顺序解析：

1. 显式选项，如 `agentrt.WithEndpoint("http://127.0.0.1:8080")`；
2. `AGENTRT_ENDPOINT` 环境变量；
3. 内置默认值 `http://127.0.0.1:18789`。

其他环境变量（均由 `agentrt.NewConfigFromEnv()` 消费）：`AGENTRT_TIMEOUT`、`AGENTRT_MAX_RETRIES`、`AGENTRT_RETRY_DELAY`、`AGENTRT_API_KEY`（以 `Authorization: Bearer <key>` 发送）、`AGENTRT_DEBUG`、`AGENTRT_LOG_LEVEL`、`AGENTRT_MAX_CONNECTIONS`、`AGENTRT_USER_AGENT`。

> 网关默认端口为 **8080**；SDK 内置默认值 `18789` 仅在既未传选项也未设置 `AGENTRT_ENDPOINT` 时生效。请按你的部署显式传入端点。

## 安装

```bash
go get github.com/spharx/agentrt/sdk/go/agentrt
```

**环境要求：** Go >= 1.22。**零外部依赖** —— 仅使用 Go 标准库（`net/http`、`encoding/json`、`context`、`sync` 等）。

## 快速入门

### 客户端与任务管理器

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/spharx/agentrt/sdk/go/agentrt"
    "github.com/spharx/agentrt/sdk/go/agentrt/client"
    "github.com/spharx/agentrt/sdk/go/agentrt/modules/task"
)

func main() {
    c, err := client.NewClient(agentrt.WithEndpoint("http://127.0.0.1:8080"))
    if err != nil {
        log.Fatal(err)
    }
    defer c.Close()

    tasks := task.NewTaskManager(c)
    ctx := context.Background()

    submitted, err := tasks.Submit(ctx, "analyze quarterly metrics")
    if err != nil {
        log.Fatal(err)
    }

    result, err := tasks.Wait(ctx, submitted.ID, 60*time.Second)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Task %s finished: %s\n", result.ID, result.Output)
}
```

### 记忆与会话

```go
memory := memory.NewMemoryManager(c)
mem, err := memory.Write(ctx, "user prefers concise answers", types.MemoryLayerL2)
hits, err := memory.Search(ctx, "user preferences", 5)

sessions := session.NewSessionManager(c)
sess, err := sessions.Create(ctx, "user-123")
err = sessions.SetContext(ctx, sess.ID, "locale", "zh-CN")
```

### 技能

```go
skills := skill.NewSkillManager(c)
loaded, err := skills.Load(ctx, "web-search")
res, err := skills.Execute(ctx, "web-search", map[string]interface{}{"query": "airymax"})
```

### 配置

```go
config := agentrt.NewConfig(
    agentrt.WithEndpoint("http://127.0.0.1:8080"),
    agentrt.WithTimeout(30*time.Second),
    agentrt.WithMaxRetries(3),
    agentrt.WithAPIKey("your-api-key"),
    agentrt.WithUserAgent("my-app/1.0"),
    agentrt.WithDebug(true),
)
c, _ := client.NewClientWithConfig(config)

// 或完全从环境变量加载：
config, _ := agentrt.NewConfigFromEnv()
```

## 构建与测试

```bash
# 运行全部测试
go test ./...

# 运行基准测试
go test -bench=. ./...

# 运行指定模块
go test ./agentrt/modules/task/...
```

## 许可证

采用 **AGPL v3 + Apache 2.0** 双许可证（SPDX: `AGPL-3.0-or-later OR Apache-2.0`）。详见 [LICENSE](LICENSE)。

Copyright (c) 2025-2026 **SPHARX Ltd.** All Rights Reserved.

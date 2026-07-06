**语言:** [English](README.md) | 简体中文

# Airymax Go SDK

[![Version](https://img.shields.io/badge/version-0.1.1-5a6b7e)](https://atomgit.com/openairymax/sdk-go)
[![License](https://img.shields.io/badge/license-AGPL--3.0+Apache--2.0-4a90d9)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)

> [Airymax](https://atomgit.com/openairymax/airymaxhub) AI 智能体运行时平台的官方 Go 开发工具包。
> [sdk](https://atomgit.com/openairymax/sdk) 管理仓聚合的叶子仓之一。

---

## 概述

**Airymax Go SDK**（`agentrt`）提供符合 Go 惯用法的 Airymax 运行时接口。它与其他语言 SDK 共享相同的双层 API 架构，并基于 Go 标准库实现 HTTP/JSON 栈，因此**无任何外部运行时依赖**。并发通过 goroutine 与 `context.Context` 表达，客户端暴露与运行时响应码一一对应的类型化错误码体系。

基于该 SDK 构建的 Agent 应用是**运行时租户**：通过 SDK 调用系统能力，而非直接访问内核内部。

## 双层 API 架构

每个 Airymax SDK 都提供顶层 `AgentRTClient`，内嵌四个资源客户端，分别覆盖运行时的一个平面：

```
AgentRTClient
├── CognitionClient   # 认知平面：任务 / 循环 / 推理
├── SafetyClient      # 安全平面：审计 / 沙箱 / 策略
├── ToolClient        # 工具平面：注册 / 调用 / 编排
└── ChatClient        # 对话平面：LLM 路由 / 会话 / 流式
```

Go 客户端通过 `client.Cognition`、`client.Safety`、`client.Tool`、`client.Chat` 暴露这些平面，底层是带连接池、重试与退避的 HTTP 传输。

## 目录结构

```
sdk-go/
├── agentrt/
│   ├── agentrt.go              # Version / Author / License 常量
│   ├── config.go               # Config / ConfigOption / 环境变量加载
│   ├── protocol.go             # 协议处理
│   ├── errors.go               # 错误类型与错误码常量
│   ├── client/
│   │   ├── client.go           # APIClient / Client / ClientConfig
│   │   └── mock.go             # 测试用 MockClient
│   ├── modules/                # 业务模块管理器
│   │   ├── modules.go          # 模块导出
│   │   ├── base_manager.go     # BaseManager
│   │   ├── task/               # TaskManager（含 benchmark_test）
│   │   ├── memory/             # MemoryManager
│   │   ├── session/            # SessionManager
│   │   └── skill/              # SkillManager
│   ├── plugin/                 # 插件系统
│   ├── syscall/                # 系统调用绑定
│   ├── telemetry/              # OpenTelemetry 追踪
│   ├── types/                  # 枚举与领域模型
│   └── utils/                  # 工具函数
├── go.mod                      # Go 模块：github.com/spharx/agentrt/sdk/go
└── README.md                   # 本文件
```

## 上下游依赖

### 上游

- **运行时**：通过 HTTP 和 JSON-RPC 2.0 连接到运行中的 Airymax / AgentRT 实例（`gateway_d`）。
- **协议**：使用平台 `protocols/` 中定义的 AgentsIPC 协议。
- **配置**：依次从函数式选项、环境变量（`AGENTRT_ENDPOINT`、`AGENTRT_TIMEOUT`、`AGENTRT_API_KEY`）、默认值 `http://127.0.0.1:18789` 解析。

### 下游

- **Agent 应用**：用户编写的 Agent 导入 `agentrt` 成为运行时租户。
- **示例**：平台 `ecosystem/examples/` 中的参考 Agent。

## 安装

```bash
go get github.com/spharx/agentrt/sdk/go
```

**环境要求：** Go >= 1.22。**无外部运行时依赖** —— 仅使用 Go 标准库（`net/http`、`encoding/json`、`context`、`sync`）。

## 快速入门

### 客户端初始化

```go
import "agentrt"

client, err := agentrt.NewClient("http://localhost:18789")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

clientWithKey, _ := agentrt.NewClientWithAPIKey("http://localhost:18789", "your-api-key")
```

### 认知平面 —— 任务

```go
ctx := context.Background()

task, err := client.Cognition.SubmitTask(ctx, "analyze this data")
result, err := client.Cognition.Wait(ctx, task.ID, 30*time.Second)
```

### 工具平面 —— 调用

```go
res, err := client.Tool.Invoke(ctx, "web-scraper", map[string]any{"url": "https://example.com"})
```

### 配置

```go
config := agentrt.NewConfig(
    agentrt.WithEndpoint("http://localhost:18789"),
    agentrt.WithTimeout(30*time.Second),
    agentrt.WithMaxRetries(3),
    agentrt.WithAPIKey("your-api-key"),
    agentrt.WithUserAgent("my-app/1.0"),
    agentrt.WithDebug(true),
)
client, _ := agentrt.NewClientWithConfig(config)

// 或完全从环境变量加载：
config = agentrt.NewConfigFromEnv()
```

### 完整示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "agentrt"
)

func main() {
    client, err := agentrt.NewClient("http://localhost:18789")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    task, err := client.Cognition.SubmitTask(ctx, "analyze sales data")
    if err != nil {
        log.Fatal(err)
    }
    result, err := client.Cognition.Wait(ctx, task.ID, 60*time.Second)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Task result: %s\n", result.Output)
}
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

## 分支策略

本叶子仓在 **`feature/official-hubs-01`** 分支上开发。聚合管理仓 `sdk` 仅使用 `main` 分支。

## 许可证

采用 **AGPL v3 + Apache 2.0** 双许可证（SPDX: `AGPL-3.0-or-later OR Apache-2.0`）。详见 [LICENSE](LICENSE)。

Copyright (c) 2025-2026 **SPHARX Ltd.** All Rights Reserved.

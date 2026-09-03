**Language:** English | [简体中文](README_zh.md)

# Airymax Go SDK

[![Version](https://img.shields.io/badge/version-0.1.9-5a6b7e)](https://atomgit.com/openairymax/sdk-go)
[![License](https://img.shields.io/badge/license-AGPL--3.0+Apache--2.0-4a90d9)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)

> Official Go development kit for the [Airymax](https://atomgit.com/openairymax/airymaxhub) AI Agent Runtime Platform.
> One of the leaf repositories aggregated by the [sdk](https://atomgit.com/openairymax/sdk) management repo.

---

## Overview

The **Airymax Go SDK** (`agentrt`) provides an idiomatic Go interface to the Airymax runtime. It shares the same double-layer API architecture as the other language SDKs and leans on Go's standard library for the HTTP/JSON stack, so it ships with **no external runtime dependencies**. Concurrency is expressed through goroutines and `context.Context`, and the client surfaces a typed error-code system that maps 1:1 with the runtime's response codes.

Agent applications built on this SDK are **runtime tenants**: they invoke system capabilities through the SDK rather than touching kernel internals directly.

## Double-Layer API Architecture

Every Airymax SDK ships a top-level `AgentRTClient` that nests four resource clients, each covering one plane of the runtime:

```
AgentRTClient
├── CognitionClient   # Cognition plane: tasks / loops / inference
├── SafetyClient      # Safety plane: audit / sandbox / policy
├── ToolClient        # Tool plane: register / invoke / orchestrate
└── ChatClient        # Chat plane: LLM routing / sessions / streaming
```

The Go client exposes these as `client.Cognition`, `client.Safety`, `client.Tool`, and `client.Chat`, backed by a connection-pooled HTTP transport with retry and backoff.

## Directory Structure

```
sdk-go/
├── agentrt/
│   ├── agentrt.go              # Version / Author / License constants
│   ├── config.go               # Config / ConfigOption / env-var loading
│   ├── protocol.go             # Protocol handling
│   ├── errors.go               # Error types + error-code constants
│   ├── client/
│   │   ├── client.go           # APIClient / Client / ClientConfig
│   │   └── mock.go             # MockClient for tests
│   ├── modules/                # Domain module managers
│   │   ├── modules.go          # Module exports
│   │   ├── base_manager.go     # BaseManager
│   │   ├── task/               # TaskManager (+ benchmark_test)
│   │   ├── memory/             # MemoryManager
│   │   ├── session/            # SessionManager
│   │   └── skill/              # SkillManager
│   ├── plugin/                 # Plugin system
│   ├── syscall/                # Syscall bindings
│   ├── telemetry/              # OpenTelemetry tracing
│   ├── types/                  # Enums + domain models
│   └── utils/                  # Helpers
├── go.mod                      # Go module: github.com/spharx/agentrt/sdk/go
└── README.md                   # This file
```

## Upstream & Downstream Dependencies

### Upstream

- **Runtime**: Connects to a running Airymax / AgentRT instance (`gateway_d`) over HTTP and JSON-RPC 2.0.
- **Protocol**: Speaks the AgentsIPC protocol defined in the platform `protocols/` tree.
- **Configuration**: Resolved from functional options, then environment variables (`AGENTRT_ENDPOINT`, `AGENTRT_TIMEOUT`, `AGENTRT_API_KEY`), then a `http://127.0.0.1:18789` default.

### Downstream

- **Agent applications**: User-written agents import `agentrt` to become runtime tenants.
- **Examples**: Reference agents in the platform `ecosystem/examples/`.

## Installation

```bash
go get github.com/spharx/agentrt/sdk/go
```

**Requirements:** Go >= 1.22. **No external runtime dependencies** — only the Go standard library (`net/http`, `encoding/json`, `context`, `sync`).

## Quick Start

### Client initialization

```go
import "agentrt"

client, err := agentrt.NewClient("http://localhost:18789")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

clientWithKey, _ := agentrt.NewClientWithAPIKey("http://localhost:18789", "your-api-key")
```

### Cognition plane — tasks

```go
ctx := context.Background()

task, err := client.Cognition.SubmitTask(ctx, "analyze this data")
result, err := client.Cognition.Wait(ctx, task.ID, 30*time.Second)
```

### Tool plane — invocation

```go
res, err := client.Tool.Invoke(ctx, "web-scraper", map[string]any{"url": "https://example.com"})
```

### Configuration

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

// Or load entirely from the environment:
config = agentrt.NewConfigFromEnv()
```

### Complete example

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

## Build & Test

```bash
# Run all tests
go test ./...

# Run benchmarks
go test -bench=. ./...

# Run a specific module
go test ./agentrt/modules/task/...
```

## Branch Strategy

This leaf repository is developed on **`develop/hubs-01`**; its `main` is a release snapshot. The aggregating `sdk` management repo develops directly on `main`.

## License

Dual-licensed under **AGPL v3 + Apache 2.0** (SPDX: `AGPL-3.0-or-later OR Apache-2.0`). See [LICENSE](LICENSE) for the full text.

Copyright (c) 2025-2026 **SPHARX Ltd.** All Rights Reserved.

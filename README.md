**Language:** English | [简体中文](README_zh.md)

# Airymax Go SDK

[![Version](https://img.shields.io/badge/version-0.1.1-5a6b7e)](https://atomgit.com/openairymax/sdk-go)
[![License](https://img.shields.io/badge/license-AGPL--3.0+Apache--2.0-4a90d9)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)

> Official Go development kit for the [Airymax](https://atomgit.com/openairymax) AI Agent Runtime Platform.
> One of the leaf repositories aggregated by the [sdk](https://atomgit.com/openairymax/sdk) management repo.

---

## Overview

The **Airymax Go SDK** (import path `github.com/spharx/agentrt/sdk/go/agentrt`, package name `agentrt`) provides an idiomatic Go interface to the Airymax runtime. It shares the same API architecture as the other language SDKs and leans on Go's standard library for the HTTP/JSON stack, so it ships with **zero external dependencies**. Concurrency is expressed through goroutines and `context.Context`, and the client surfaces a typed error-code system (`agentrt/errors.go`).

Agent applications built on this SDK are **runtime tenants**: they invoke system capabilities through the SDK rather than touching runtime internals directly.

## API Architecture

The Go SDK follows the common two-layer layout:

- **HTTP client layer** — `client.Client` implements the `client.APIClient` interface (`Get` / `Post` / `Put` / `Delete`) with connection pooling, retry with exponential backoff + jitter, and Bearer-token authentication. It also exposes `Health` and `Metrics` probes and a `MockClient` for tests.
- **Module manager layer** — four domain managers, one per resource plane of the runtime:

```
client.NewClient(...)                    # HTTP client (client.APIClient)
├── task.NewTaskManager(api)             # Task lifecycle: submit / wait / cancel / list
├── memory.NewMemoryManager(api)         # Memory: write / search / update / layers
├── session.NewSessionManager(api)       # Sessions: create / context kv / close
└── skill.NewSkillManager(api)           # Skills: load / execute / register / search
```

## Directory Structure

```
sdk-go/
├── agentrt/
│   ├── agentrt.go              # Version / Author / License constants
│   ├── config.go               # Config / ConfigOption / env-var loading
│   ├── protocol.go             # Protocol handling
│   ├── errors.go               # Error types + error-code constants
│   ├── client/
│   │   ├── client.go           # APIClient / Client / Health / Metrics
│   │   └── mock.go             # MockClient for tests
│   ├── modules/                # Domain module managers
│   │   ├── modules.go          # Module exports
│   │   ├── base_manager.go     # Generic BaseManager + resource converters
│   │   ├── task/               # TaskManager (+ benchmark tests)
│   │   ├── memory/             # MemoryManager
│   │   ├── session/            # SessionManager
│   │   └── skill/              # SkillManager
│   ├── plugin/                 # Plugin system
│   ├── syscall/                # Syscall bindings
│   ├── telemetry/              # Built-in distributed tracing (Tracer / Span)
│   ├── types/                  # Enums + domain models
│   └── utils/                  # Helpers
├── go.mod                      # Go module: github.com/spharx/agentrt/sdk/go
└── README.md                   # This file
```

## Prerequisites

The SDK is a client of a running AgentRT runtime. Install the runtime first:

```bash
curl -fsSL "https://api.atomgit.com/api/v5/repos/openairymax/agentrt/contents/scripts/install.sh?ref=main" | python3 -c 'import json,sys,base64;sys.stdout.buffer.write(base64.b64decode(json.load(sys.stdin)["content"]))' | bash
```

(A compatibility entry `https://atomgit.com/openairymax/agentrt/releases/download/latest/install.sh` is also available.) Then start the gateway, e.g. `airymaxrt start`.

### Endpoint Resolution

The endpoint is resolved in this order:

1. explicit options, e.g. `agentrt.WithEndpoint("http://127.0.0.1:8080")`;
2. `AGENTRT_ENDPOINT` environment variable;
3. built-in default `http://127.0.0.1:18789`.

Other environment variables (all consumed by `agentrt.NewConfigFromEnv()`): `AGENTRT_TIMEOUT`, `AGENTRT_MAX_RETRIES`, `AGENTRT_RETRY_DELAY`, `AGENTRT_API_KEY` (sent as `Authorization: Bearer <key>`), `AGENTRT_DEBUG`, `AGENTRT_LOG_LEVEL`, `AGENTRT_MAX_CONNECTIONS`, `AGENTRT_USER_AGENT`.

> The gateway's default port is **8080**; the SDK's built-in default of `18789` applies only when neither an option nor `AGENTRT_ENDPOINT` is set. Pass the endpoint explicitly to match your deployment.

## Installation

```bash
go get github.com/spharx/agentrt/sdk/go/agentrt
```

**Requirements:** Go >= 1.22. **Zero external dependencies** — only the Go standard library (`net/http`, `encoding/json`, `context`, `sync`, ...).

## Quick Start

### Client and task manager

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

### Memory and sessions

```go
memory := memory.NewMemoryManager(c)
mem, err := memory.Write(ctx, "user prefers concise answers", types.MemoryLayerL2)
hits, err := memory.Search(ctx, "user preferences", 5)

sessions := session.NewSessionManager(c)
sess, err := sessions.Create(ctx, "user-123")
err = sessions.SetContext(ctx, sess.ID, "locale", "zh-CN")
```

### Skills

```go
skills := skill.NewSkillManager(c)
loaded, err := skills.Load(ctx, "web-search")
res, err := skills.Execute(ctx, "web-search", map[string]interface{}{"query": "airymax"})
```

### Configuration

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

// Or load entirely from the environment:
config, _ := agentrt.NewConfigFromEnv()
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

## License

Dual-licensed under **AGPL v3 + Apache 2.0** (SPDX: `AGPL-3.0-or-later OR Apache-2.0`). See [LICENSE](LICENSE) for the full text.

Copyright (c) 2025-2026 **SPHARX Ltd.** All Rights Reserved.

# 🏗️ Architecture Guide

This document provides a comprehensive overview of the Expose architecture, design patterns, and implementation details.

## Table of Contents

- [Overview](#overview)
- [Architecture Diagram](#architecture-diagram)
- [Project Structure](#project-structure)
- [Core Components](#core-components)
- [Data Flow](#data-flow)
- [Design Patterns](#design-patterns)
- [Provider Implementation](#provider-implementation)
- [Concurrency Model](#concurrency-model)
- [Error Handling](#error-handling)
- [Testing Strategy](#testing-strategy)

---

## Overview

Expose is built with a clean, modular architecture following Go best practices. The system is designed around three core principles:

1. **Provider Abstraction** - Support multiple tunnel providers through a common interface
2. **Separation of Concerns** - Clear boundaries between CLI, business logic, and providers
3. **Concurrency Safety** - Thread-safe operations with proper synchronization

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     CLI Layer                            │
│  (cmd/expose, internal/cli)                             │
│  - Command parsing                                       │
│  - Flag handling                                         │
│  - User interaction                                      │
└──────────────────┬───────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────┐
│                  Service Layer                           │
│  (internal/tunnel)                                       │
│  - Lifecycle management                                  │
│  - Provider coordination                                 │
│  - State management                                      │
└──────────────────┬───────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────┐
│                 Provider Layer                           │
│  (internal/provider)                                     │
│  - LocalTunnel implementation                            │
│  - Cloudflare implementation                             │
│  - Connection management                                 │
│  - Traffic proxying                                      │
└─────────────────────────────────────────────────────────┘
```

---

## Architecture Diagram

```
                     User
                       |
                       v
            ┌──────────────────┐
            │   CLI Interface  │
            │   (Cobra)        │
            └────────┬─────────┘
                     |
         ┌───────────┴────────────┐
         |                        |
    ┌────v─────┐           ┌─────v──────┐
    │  Config  │           │   Tunnel   │
    │  Manager │           │   Service  │
    └──────────┘           └─────┬──────┘
                                 |
                    ┌────────────┴────────────┐
                    |                         |
           ┌────────v─────────┐    ┌─────────v────────┐
           │   LocalTunnel    │    │   Cloudflare     │
           │   Provider       │    │   Provider       │
           └────────┬─────────┘    └─────────┬────────┘
                    |                        |
              ┌─────v──────┐          ┌──────v─────┐
              │ TCP Pool   │          │ cloudflared│
              │ Management │          │   Process  │
              └─────┬──────┘          └──────┬─────┘
                    |                        |
                    └────────┬───────────────┘
                             |
                    ┌────────v──────────┐
                    │  Localhost:PORT   │
                    │  (Your Server)    │
                    └───────────────────┘
```

---

## Project Structure

```
expose/
├── cmd/
│   └── expose/
│       └── main.go              # Entry point
│
├── internal/
│   ├── cli/                     # CLI commands layer
│   │   ├── root.go             # Root command setup
│   │   ├── init.go             # 'init' command
│   │   ├── tunnel.go           # 'tunnel' command
│   │   ├── config.go           # 'config' command
│   │   └── *_test.go           # CLI tests
│   │
│   ├── config/                  # Configuration management
│   │   ├── config.go           # Config struct & operations
│   │   └── config_test.go      # Config tests
│   │
│   ├── tunnel/                  # Core tunnel service
│   │   ├── provider.go         # Provider interface
│   │   ├── service.go          # Service implementation
│   │   ├── manager.go          # (Future) Multi-tunnel manager
│   │   └── *_test.go           # Service tests
│   │
│   ├── provider/                # Tunnel provider implementations
│   │   ├── localtunnel.go      # LocalTunnel provider
│   │   ├── cloudflare.go       # Cloudflare provider
│   │   └── *_test.go           # Provider tests
│   │
│   └── version/                 # Version information
│       ├── version.go          # Version string
│       └── version_test.go     # Version tests
│
├── docs/                        # Documentation
│   ├── GETTING_STARTED.md
│   ├── ARCHITECTURE.md         # This file
│   └── ONBOARDING.md
│
├── .expose.yml                  # Example config
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
└── README.md                    # Project README
```

### Package Responsibilities

| Package | Responsibility | Dependencies |
|---------|---------------|--------------|
| `cmd/expose` | Application entry point | `internal/cli` |
| `internal/cli` | Command-line interface | `internal/config`, `internal/tunnel`, `internal/provider` |
| `internal/config` | Configuration file I/O | `gopkg.in/yaml.v3` |
| `internal/tunnel` | Tunnel lifecycle & service | `internal/provider` |
| `internal/provider` | Provider implementations | `net`, `http`, `os/exec` |
| `internal/version` | Version info | None |

---

## Core Components

### 1. CLI Layer (`internal/cli`)

**Purpose**: Handle user interaction and command execution.

**Key Files**:
- `root.go` - Sets up Cobra root command
- `tunnel.go` - Implements `expose tunnel` command
- `init.go` - Implements `expose init` command
- `config.go` - Implements `expose config` command

**Responsibilities**:
- Parse command-line arguments
- Load configuration
- Coordinate with service layer
- Display output to user
- Handle signals (Ctrl+C)

**Example Flow**:
```go
// tunnel.go
func runTunnelCmd(cmd *cobra.Command, _ []string) error {
    // 1. Load config
    cfg, err := config.Load("")
    
    // 2. Get port from flag or config
    port := getPort(cmd, cfg)
    
    // 3. Select provider
    providerName := getProvider(cmd)
    
    // 4. Run tunnel
    return runTunnel(port, providerName)
}
```

### 2. Config Management (`internal/config`)

**Purpose**: Manage project configuration files.

**Key Types**:
```go
type Config struct {
    Project string `yaml:"project"`
    Port    int    `yaml:"port"`
}
```

**Operations**:
- `Load(path)` - Read config from file
- `Init()` - Create default config
- `List()` - Return all config values
- `Get(key)` - Retrieve specific value

**File Format** (`.expose.yml`):
```yaml
project: my-app
port: 3000
```

### 3. Tunnel Service (`internal/tunnel`)

**Purpose**: Manage tunnel lifecycle and coordinate providers.

**Key Types**:

```go
// Provider interface - abstraction for all tunnel providers
type Provider interface {
    Connect(ctx context.Context, localPort int) (string, error)
    Close() error
    IsConnected() bool
    PublicURL() string
    Name() string
}

// Service - wraps a provider with lifecycle management
type Service struct {
    provider Provider
    ready    chan struct{}
    mu       sync.RWMutex
    started  bool
    closed   bool
}
```

**Service Lifecycle**:

1. **Creation**: `NewService(provider)`
2. **Starting**: `Start(ctx, port)` - connects provider
3. **Ready Signal**: `Ready()` channel closes when ready
4. **Active**: Tunnel forwards traffic
5. **Shutdown**: `Close()` - cleanup resources

**Thread Safety**:
- Uses `sync.RWMutex` for state protection
- Prevents multiple starts or closes
- Safe concurrent access to status methods

### 4. Provider Layer (`internal/provider`)

**Purpose**: Implement tunnel provider-specific logic.

#### LocalTunnel Provider

**Key Types**:
```go
type localTunnel struct {
    publicURL      string
    localPort      int
    tunnelPort     int
    tunnelHost     string
    connected      bool
    mu             sync.RWMutex
    connections    []net.Conn
    maxConnections int
    ctx            context.Context
    cancel         context.CancelFunc
    httpClient     *http.Client
    serverAPIEndpoint string
}
```

**Connection Flow**:

```
1. Request Tunnel
   ├─> HTTP GET to localtunnel.me/?new
   ├─> Receive: {id, url, port, max_conn_count}
   └─> Store tunnel configuration

2. Open Connection Pool
   ├─> Dial TCP to localtunnel.me:PORT
   ├─> Repeat for maxConnections (typically 10)
   └─> Start handler goroutine for each

3. Handle Connections (per connection)
   ├─> Read request from tunnel
   ├─> Dial localhost:PORT
   ├─> Bidirectional copy (io.Copy)
   ├─> Clear deadline for reuse
   └─> Repeat until shutdown

4. Shutdown
   ├─> Cancel context
   ├─> Close all connections
   └─> Clear state
```

**Important Implementation Details**:

1. **Connection Pooling**: Maintains 10 concurrent TCP connections
2. **Deadline Management**: Sets deadlines per request, clears after
3. **Graceful Shutdown**: Context-based cancellation
4. **Error Recovery**: Logs errors but continues serving

#### Cloudflare Provider

**Key Types**:
```go
type Cloudflare struct {
    cmd       *exec.Cmd
    mu        sync.RWMutex
    publicURL string
    RequestTunnel func(ctx context.Context, port int, timeout time.Duration) 
                      (string, *exec.Cmd, error)
}
```

**Connection Flow**:

```
1. Start cloudflared Process
   ├─> exec: cloudflared tunnel --url http://localhost:PORT
   ├─> Capture stderr output
   └─> Parse for URL pattern

2. Extract Public URL
   ├─> Regex: https://[a-z0-9-]+\.trycloudflare\.com
   ├─> Wait up to 30 seconds
   └─> Return URL + process handle

3. Traffic Forwarding
   └─> Handled by cloudflared (external process)

4. Shutdown
   ├─> Kill cloudflared process
   └─> Clear state
```

---

## Data Flow

### Complete Request Flow (LocalTunnel)

```
Internet Request
      ↓
┌─────────────────┐
│ localtunnel.me  │
│   (Server)      │
└────────┬────────┘
         │ TCP
         ↓
┌─────────────────┐
│  Connection #N  │ ← One of 10 pooled connections
│  (goroutine)    │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│  proxyRequest   │
│  - Set deadline │
│  - Dial local   │
│  - Bidirectional│
│    copy         │
│  - Clear deadline│
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│ localhost:3000  │
│ (Your Server)   │
└─────────────────┘
         │
         ↓ (Response)
┌─────────────────┐
│  io.Copy back   │
│  to tunnel      │
└─────────────────┘
```

### Complete Request Flow (Cloudflare)

```
Internet Request
      ↓
┌─────────────────┐
│   Cloudflare    │
│   Edge Network  │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│  cloudflared    │ ← External process
│   (Local)       │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│ localhost:3000  │
│ (Your Server)   │
└─────────────────┘
```

---

## Design Patterns

### 1. Strategy Pattern (Provider Selection)

Different tunnel providers are interchangeable through the `Provider` interface:

```go
var provider tunnel.Provider

switch providerName {
case "cloudflare":
    provider = provider.NewCloudFlare()
default:
    provider = provider.NewLocalTunnel(nil)
}

service := tunnel.NewService(provider)
```

### 2. Facade Pattern (Service Layer)

The `Service` wraps provider complexity with a simple API:

```go
service := tunnel.NewService(provider)
service.Start(ctx, port)
<-service.Ready()
fmt.Println(service.PublicURL())
service.Close()
```

### 3. Template Method (Provider Interface)

Common lifecycle defined by interface, implementation varies:

```go
type Provider interface {
    Connect(ctx context.Context, localPort int) (string, error)
    Close() error
    IsConnected() bool
    PublicURL() string
    Name() string
}
```

### 4. Dependency Injection

Providers are injected into service:

```go
// Allows for testing with mock providers
func NewService(p Provider) *Service {
    return &Service{provider: p}
}
```

### 5. Observer Pattern (Ready Channel)

Service signals readiness through channel:

```go
select {
case <-svc.Ready():
    fmt.Println("Ready!")
case err := <-errChan:
    return err
}
```

---

## Provider Implementation

### Implementing a New Provider

To add a new tunnel provider:

1. **Create Provider File**:
```go
// internal/provider/myprovider.go
package provider

type MyProvider struct {
    // fields
}

func NewMyProvider() *MyProvider {
    return &MyProvider{}
}
```

2. **Implement Interface**:
```go
func (m *MyProvider) Connect(ctx context.Context, localPort int) (string, error) {
    // 1. Establish tunnel connection
    // 2. Get public URL
    // 3. Start traffic forwarding
    return publicURL, nil
}

func (m *MyProvider) Close() error {
    // Cleanup resources
    return nil
}

func (m *MyProvider) IsConnected() bool {
    // Check connection status
    return true
}

func (m *MyProvider) PublicURL() string {
    return m.publicURL
}

func (m *MyProvider) Name() string {
    return "MyProvider"
}
```

3. **Add to CLI**:
```go
// internal/cli/tunnel.go
switch providerName {
case "cloudflare":
    svc = tunnel.NewService(provider.NewCloudFlare())
case "myprovider":
    svc = tunnel.NewService(provider.NewMyProvider())
default:
    svc = tunnel.NewService(provider.NewLocalTunnel(nil))
}
```

4. **Add Tests**:
```go
// internal/provider/myprovider_test.go
func TestMyProvider_Connect(t *testing.T) {
    // Test implementation
}
```

---

## Concurrency Model

### Thread Safety Principles

1. **Mutex Protection**: All shared state protected by `sync.RWMutex`
2. **Channel Communication**: Goroutines coordinate via channels
3. **Context Cancellation**: Graceful shutdown using `context.Context`

### LocalTunnel Concurrency

```go
// Main goroutine
├─> Start() - setup & connect
│
├─> Connection Pool (10 goroutines)
│   ├─> handleConnection #1
│   ├─> handleConnection #2
│   ├─> ...
│   └─> handleConnection #10
│       │
│       └─> Per request
│           ├─> proxyRequest goroutine #1 (tunnel→local)
│           └─> proxyRequest goroutine #2 (local→tunnel)
│
└─> Shutdown
    ├─> ctx.Cancel()
    └─> Close all connections
```

### Critical Section Example

```go
// Read lock for reading shared state
func (lt *localTunnel) PublicURL() string {
    lt.mu.RLock()
    defer lt.mu.RUnlock()
    return lt.publicURL
}

// Write lock for modifying shared state
func (lt *localTunnel) Connect(ctx context.Context, localPort int) (string, error) {
    lt.mu.Lock()
    lt.localPort = localPort
    lt.ctx, lt.cancel = context.WithCancel(ctx)
    lt.mu.Unlock()
    // ... rest of logic
}
```

### Deadline Bug Fix

**Problem**: TCP connection deadlines persist across requests, causing reuse issues.

**Solution**: Clear deadline after each request:

```go
// In proxyRequest()
wg.Wait() // Wait for bidirectional copy

// IMPORTANT: Clear deadline for connection reuse
_ = tunnelConn.SetDeadline(time.Time{})

return nil
```

---

## Error Handling

### Error Propagation

Errors flow up through layers:

```
Provider Error
    ↓
Service.Start() error
    ↓
runTunnel() error
    ↓
runTunnelCmd() error
    ↓
main() - prints to stderr
```

### Error Wrapping

Use `fmt.Errorf` with `%w` for context:

```go
if err := svc.Start(ctx, port); err != nil {
    return fmt.Errorf("failed to start tunnel: %w", err)
}
```

### Error Categories

1. **Configuration Errors**: Invalid config, missing file
2. **Network Errors**: Connection failures, timeouts
3. **Provider Errors**: API failures, process crashes
4. **User Errors**: Invalid input, wrong flags

---

## Testing Strategy

### Test Coverage Areas

1. **Unit Tests**: Individual functions and methods
2. **Integration Tests**: Component interactions
3. **Mock Testing**: External dependencies (HTTP, exec)

### Test Structure

```go
func TestComponent_Method(t *testing.T) {
    // Arrange
    setup := createTestSetup()
    
    // Act
    result, err := setup.Method()
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### Mocking Providers

```go
type mockProvider struct {
    connectFunc func(ctx context.Context, port int) (string, error)
}

func (m *mockProvider) Connect(ctx context.Context, port int) (string, error) {
    if m.connectFunc != nil {
        return m.connectFunc(ctx, port)
    }
    return "http://mock.test", nil
}
```

### Test Best Practices

1. **Table-driven tests** for multiple scenarios
2. **Timeout contexts** prevent hanging tests
3. **Cleanup with defer** ensure resources released
4. **Mock external dependencies** for reliability

---

## Performance Considerations

### LocalTunnel Optimizations

1. **Connection Pooling**: 10 concurrent connections reduce latency
2. **Bidirectional Copy**: Parallel upload/download for full duplex
3. **Deadline Management**: Prevents resource leaks from hung connections
4. **Connection Reuse**: Same TCP conn handles multiple requests

### Memory Management

- Connection pool bounded to 10 (configurable)
- Goroutines cleaned up on context cancellation
- No unbounded buffers or queues
- HTTP client reused across requests

### Latency Profile

```
Request Latency =
  Network to Tunnel Server +
  Tunnel Server Processing +
  Local Network +
  Application Processing +
  Response Path (reverse)

Typical: 50-200ms overhead
```

---

## Future Improvements

### Planned Enhancements

1. **Multi-Tunnel Manager**: Run multiple tunnels simultaneously
2. **Health Checks**: Monitor tunnel connectivity
3. **Metrics**: Request counts, latency tracking
4. **Custom Domains**: Support custom subdomains
5. **TLS Termination**: Handle HTTPS locally
6. **Load Balancing**: Distribute across multiple backends

### Extension Points

- New providers via `Provider` interface
- Custom authentication handlers
- Middleware for request/response inspection
- Plugin system for extensibility

---

## References

- [Provider Interface](../internal/tunnel/provider.go)
- [LocalTunnel Implementation](../internal/provider/localtunnel.go)
- [Cloudflare Implementation](../internal/provider/cloudflare.go)
- [Service Layer](../internal/tunnel/service.go)
- [Contributing Guide](../CONTRIBUTING.md)

---

**Questions?** Open an issue or discussion on [GitHub](https://github.com/kernelshard/expose/issues).

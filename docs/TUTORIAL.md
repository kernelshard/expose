---
title: How Tunnels Work
description: Understand the protocol behind Expose by building a simplified tunnel client from scratch in Go.
---

# How Tunnels Work

This guide explains how reverse-proxy tunnels work by walking you through building a simplified version of Expose from scratch. By the end you'll understand every moving part of the real codebase.

!!! note "Who this is for"
    This is for developers who want to understand how tunneling works under the hood — whether to contribute to Expose or just to satisfy curiosity. For daily usage, see [Getting Started](GETTING_STARTED.md).

---

## Part 1 — What Is a Tunnel?

A tunnel solves one problem: **your localhost is not reachable from the internet** (NAT, firewall). A public server acts as a relay:

```
Internet user
     │
     ▼
Public server (e.g. localtunnel.me)
     │  TCP connection maintained by your tunnel client
     ▼
Your machine (localhost:3000)
```

The flow for every request:

1. User hits `https://xyz.loca.lt`
2. Public server forwards request down the open TCP connection to your client
3. Your client proxies it to `localhost:3000`
4. Response travels back the same path

### LocalTunnel Protocol

```
1. Register    POST https://localtunnel.me/?new
               ← { id, url, port }

2. Connect     net.Dial("tcp", "localtunnel.me:PORT")

3. Per request  request  → [TCP conn] → dial localhost:3000
                response ← [TCP conn] ← local server
```

---

## Part 2 — Basic Tunnel Client

Let's build a minimal, working tunnel client step by step.

### Register a tunnel

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type TunnelInfo struct {
    URL  string `json:"url"`
    Port int    `json:"port"`
}

func requestTunnel() (*TunnelInfo, error) {
    resp, err := http.Get("https://localtunnel.me/?new")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var info TunnelInfo
    if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
        return nil, err
    }

    fmt.Printf("Tunnel assigned: %s (relay port %d)\n", info.URL, info.Port)
    return &info, nil
}
```

### Connect to the relay server

```go
import "net"

func connectTunnel(info *TunnelInfo) (net.Conn, error) {
    address := fmt.Sprintf("localtunnel.me:%d", info.Port)
    return net.Dial("tcp", address)
}
```

### Proxy a single request

```go
import "io"

func proxyRequest(tunnelConn net.Conn, localPort int) error {
    localConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", localPort))
    if err != nil {
        return err
    }
    defer localConn.Close()

    done := make(chan error, 2)

    go func() { _, err := io.Copy(localConn, tunnelConn); done <- err }()
    go func() { _, err := io.Copy(tunnelConn, localConn); done <- err }()

    return <-done // return when either direction closes
}
```

### Wire it together

```go
func main() {
    localPort := 3000

    info, err := requestTunnel()
    if err != nil {
        panic(err)
    }

    fmt.Printf("Public URL:     %s\n", info.URL)
    fmt.Printf("Forwarding to:  localhost:%d\n", localPort)

    conn, err := connectTunnel(info)
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    for {
        if err := proxyRequest(conn, localPort); err != nil {
            fmt.Printf("error: %v\n", err)
            break
        }
    }
}
```

Try it:

```bash
python -m http.server 3000   # local server
go run basic-tunnel.go       # your tunnel client
```

---

## Part 3 — Provider Abstraction

The real Expose supports multiple providers (LocalTunnel, Cloudflare, self-hosted). We use Go interfaces to make switching seamless.

### The Provider interface

```go
// internal/tunnel/provider.go
package tunnel

import "context"

type Provider interface {
    Connect(ctx context.Context, localPort int) (string, error)
    Close() error
    IsConnected() bool
    PublicURL() string
    Name() string
}
```

Any struct implementing these five methods can be a provider — no other changes needed in the CLI or service layer.

### LocalTunnel implementation

```go
// internal/provider/localtunnel.go
type LocalTunnel struct {
    publicURL string
    localPort int
    conn      net.Conn
    mu        sync.RWMutex
}

func (lt *LocalTunnel) Connect(ctx context.Context, localPort int) (string, error) {
    lt.mu.Lock()
    lt.localPort = localPort
    lt.mu.Unlock()

    info, err := lt.requestTunnel()
    if err != nil {
        return "", err
    }

    conn, err := net.Dial("tcp", fmt.Sprintf("localtunnel.me:%d", info.Port))
    if err != nil {
        return "", err
    }

    lt.mu.Lock()
    lt.conn = conn
    lt.publicURL = info.URL
    lt.mu.Unlock()

    go lt.handleConnections()
    return info.URL, nil
}
```

!!! info "Why `sync.RWMutex`?"
    The tunnel connection is accessed from multiple goroutines (the handler loop and the CLI layer reading the URL). A `RWMutex` allows concurrent reads while serializing writes — the standard Go pattern for shared state.

### Service wrapper

```go
// internal/tunnel/service.go
type Service struct {
    provider Provider
    ready    chan struct{}
    mu       sync.RWMutex
    started  bool
}

func NewService(p Provider) *Service {
    return &Service{
        provider: p,
        ready:    make(chan struct{}),
    }
}

func (s *Service) Start(ctx context.Context, localPort int) error {
    s.mu.Lock()
    if s.started {
        s.mu.Unlock()
        return fmt.Errorf("already started")
    }
    s.started = true
    s.mu.Unlock()

    if _, err := s.provider.Connect(ctx, localPort); err != nil {
        return err
    }

    close(s.ready) // signal readiness to callers
    return nil
}

func (s *Service) Ready() <-chan struct{} { return s.ready }
func (s *Service) PublicURL() string     { return s.provider.PublicURL() }
func (s *Service) ProviderName() string  { return s.provider.Name() }
func (s *Service) Close() error          { return s.provider.Close() }
```

Usage:

```go
provider := NewLocalTunnel()
svc := NewService(provider)

go svc.Start(ctx, 3000)

<-svc.Ready()
fmt.Println("Tunnel ready:", svc.PublicURL())
```

---

## Part 4 — Building the CLI

Expose uses [Cobra](https://github.com/spf13/cobra) for the CLI layer.

### Root command

```go
// internal/cli/root.go
var rootCmd = &cobra.Command{
    Use:   "expose",
    Short: "Expose localhost to the internet",
}

func Execute() error {
    return rootCmd.Execute()
}
```

### Tunnel command

```go
// internal/cli/tunnel.go
var tunnelCmd = &cobra.Command{
    Use:   "tunnel",
    Short: "Start a tunnel to localhost",
    RunE:  runTunnel,
}

func init() {
    tunnelCmd.Flags().IntVarP(&port, "port", "p", 0, "Local port (overrides config)")
    tunnelCmd.Flags().StringVarP(&providerName, "provider", "P", "localtunnel", "Provider")
    rootCmd.AddCommand(tunnelCmd)
}

func runTunnel(cmd *cobra.Command, _ []string) error {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Graceful shutdown on Ctrl+C
    go func() {
        sig := make(chan os.Signal, 1)
        signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
        <-sig
        fmt.Println("\nShutting down...")
        cancel()
    }()

    cfg, _ := config.Load("")
    if port == 0 {
        port = cfg.Port
    }

    svc := tunnel.NewService(providerForName(providerName))

    errCh := make(chan error, 1)
    go func() { errCh <- svc.Start(ctx, port) }()

    select {
    case <-svc.Ready():
        fmt.Printf("✓ Public URL:      %s\n", svc.PublicURL())
        fmt.Printf("✓ Forwarding to:   http://localhost:%d\n", port)
        fmt.Printf("✓ Provider:        %s\n", svc.ProviderName())
        fmt.Println("  Press Ctrl+C to stop")
    case err := <-errCh:
        return err
    }

    <-ctx.Done()
    return svc.Close()
}
```

---

## Part 5 — Advanced Patterns

### Connection pooling

LocalTunnel maintains 10 concurrent TCP connections so multiple requests can be handled in parallel without waiting for each connection to finish:

```go
func (lt *LocalTunnel) openPool(port int) error {
    for i := 0; i < 10; i++ {
        conn, err := net.Dial("tcp", fmt.Sprintf("localtunnel.me:%d", port))
        if err != nil {
            return err
        }
        go lt.handleConn(conn)
    }
    return nil
}
```

### Cloudflare provider — subprocess wrapping

The Cloudflare provider doesn't open raw TCP connections. Instead it shells out to the `cloudflared` binary and parses the public URL from its stderr:

```go
func (c *Cloudflare) Connect(ctx context.Context, localPort int) (string, error) {
    cmd := exec.CommandContext(ctx, "cloudflared", "tunnel",
        "--url", fmt.Sprintf("http://localhost:%d", localPort),
        "--no-autoupdate",
    )

    stderr, _ := cmd.StderrPipe()
    if err := cmd.Start(); err != nil {
        return "", fmt.Errorf("cloudflared not found: %w", err)
    }

    // Parse the public URL from output like:
    // "Your quick Tunnel has been created! Visit it at: https://..."
    scanner := bufio.NewScanner(stderr)
    for scanner.Scan() {
        if match := urlPattern.FindString(scanner.Text()); match != "" {
            c.publicURL = match
            return match, nil
        }
    }

    return "", fmt.Errorf("failed to get tunnel URL from cloudflared")
}
```

---

## Summary

| Concept | Where in Expose |
|---|---|
| `Provider` interface | `internal/tunnel/provider.go` |
| `Service` lifecycle | `internal/tunnel/service.go` |
| LocalTunnel TCP pool | `internal/provider/localtunnel.go` |
| Cloudflare subprocess | `internal/provider/cloudflare.go` |
| CLI commands (Cobra) | `internal/cli/` |
| Config file (YAML) | `internal/config/config.go` |

**Next steps:**

- Read [Architecture](ARCHITECTURE.md) for the full concurrency model
- Set up a dev environment via [Onboarding](ONBOARDING.md)
- Add your own provider by implementing the `Provider` interface

---

**Questions?** [Open a discussion on GitHub](https://github.com/kernelshard/expose/discussions).

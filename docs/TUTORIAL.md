---
title: How Tunnels Work: Build an ngrok Alternative in Go
description: Understand how reverse proxy tunnels like ngrok and Expose work under the hood. Build a lightweight, pure Go reverse-proxy tunnel client from first principles.
---

# How Tunnels Work: Build an ngrok Alternative in Go

Reverse-proxy tunnels allow services running on your local machine (`localhost`) to be reachable from anywhere on the public internet, bypassing NATs, CGNATs, and restrictive local firewalls.

This tutorial breaks down the mechanics of tunneling from first principles. By the end, you will understand how tools like **ngrok**, **LocalTunnel**, and **Expose** work under the hood, and how to build a working tunnel client in pure Go.

!!! tip "Public Sandboxes vs. Self-Hosted Relays"
    In this tutorial, we will use the free, open **LocalTunnel protocol** (`localtunnel.me`) as a zero-setup sandbox so you can run the code immediately without spinning up a cloud server. 
    
    In production, **Expose** also includes its own built-in, pure-Go private relay server (`expose server`) for complete data privacy, custom subdomains, and zero third-party reliance.

---

## Part 1 — What Is a Reverse Proxy Tunnel? (NAT Traversal)

When you run a web server on `localhost:3000`, external visitors cannot connect to it because your computer sits behind a router using Network Address Translation (NAT) with a private IP address (e.g., `192.168.1.X`).

A **reverse proxy tunnel** solves this by flipping the direction of connection: **your local machine initiates an outbound TCP connection to a public relay server.** Because stateful firewalls automatically permit outbound traffic and its return packets, this outbound socket bypasses NAT and firewall restrictions without requiring port forwarding or router configuration.

```text
 ┌────────────────┐          ┌───────────────────────┐          ┌──────────────┐
 │ Internet User  │ ──HTTP─► │  Public Relay Server  │ ◄──TCP── │ Local Client │ ──HTTP─► localhost:3000
 └────────────────┘          │  (e.g. VPS / Relay)   │          └──────────────┘
                             └───────────────────────┘           (Expose Client)
```

### The Request Lifecycle

1. **Registration:** Your client contacts the relay server and requests a public endpoint (e.g., `https://my-app.relay.com`).
2. **Persistent Control Connection:** The client establishes an outbound TCP connection to the relay and keeps it open.
3. **Incoming Traffic:** A public user visits `https://my-app.relay.com`.
4. **Data Multiplexing:** The relay forwards raw HTTP bytes down the active TCP connection to your client.
5. **Local Proxying:** Your client dials `localhost:3000`, writes the HTTP bytes, reads the local server's response, and pumps it back to the relay.

---

## Part 2 — Building a Minimal Tunnel Client

Let's build a standalone, working tunnel client in Go in under 60 lines of code.

### Step 1: Request a Tunnel Endpoint

First, we ask the relay for an assigned subdomain and a dedicated TCP data port:

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

    return &info, nil
}
```

### Step 2: Establish the TCP Relay Connection

Next, we open an outbound TCP socket to the relay host on the assigned port:

```go
import "net"

func dialTunnel(info *TunnelInfo) (net.Conn, error) {
    address := fmt.Sprintf("localtunnel.me:%d", info.Port)
    return net.Dial("tcp", address)
}
```

### Step 3: Bidirectional Socket Proxying

When the relay sends an incoming HTTP request down the connection, we dial our local application and stream raw bytes using full-duplex piping with `io.Copy`:

```go
import (
    "fmt"
    "io"
    "net"
)

func proxyRequest(tunnelConn net.Conn, localPort int) error {
    localConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", localPort))
    if err != nil {
        return fmt.Errorf("local server offline: %w", err)
    }
    defer localConn.Close()

    // Dual-direction stream pump: visitor -> local, local -> visitor
    done := make(chan error, 2)
    go func() { _, err := io.Copy(localConn, tunnelConn); done <- err }()
    go func() { _, err := io.Copy(tunnelConn, localConn); done <- err }()

    // Return when either stream ends; defer closes sockets and unblocks the sibling routine
    return <-done
}
```

### Step 4: Putting It All Together

Here is the complete runnable program:

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    localPort := 3000

    info, err := requestTunnel()
    if err != nil {
        log.Fatalf("Failed to register tunnel: %v", err)
    }

    fmt.Printf("✓ Public URL:    %s\n", info.URL)
    fmt.Printf("✓ Forwarding to: http://localhost:%d\n", localPort)

    for {
        conn, err := dialTunnel(info)
        if err != nil {
            log.Printf("Tunnel connection error: %v", err)
            break
        }

        if err := proxyRequest(conn, localPort); err != nil {
            log.Printf("Proxy stream closed: %v", err)
        }
        conn.Close()
    }
}
```

To test it:
```bash
# Terminal 1: Start any local web server
python3 -m http.server 3000

# Terminal 2: Run your tunnel client
go run main.go
```

!!! note "Handling Traffic Concurrency"
    This introductory script handles requests sequentially. In **Part 5**, we show how Expose uses pre-warmed connection pools to serve concurrent requests simultaneously.

---

## Part 3 — The Provider Abstraction Pattern

Production systems like Expose support multiple relays (LocalTunnel, Cloudflare, private VPS). Moving from our standalone script to a production codebase introduces two crucial Go patterns:
- **`context.Context`**: For dial timeouts, cancellation, and leak-free shutdowns that our 50-line script omitted.
- **`Provider` Interfaces**: To switch transports with zero changes to CLI or business logic.

### The `Provider` Interface

```go
// internal/tunnel/provider.go
package tunnel

import "context"

type Provider interface {
    // Connect starts the tunnel for the given local port and returns the public URL
    Connect(ctx context.Context, localPort int) (string, error)
    // Close terminates the tunnel and cleans up socket descriptors
    Close() error
    // IsConnected reports whether the tunnel transport is currently alive
    IsConnected() bool
    // PublicURL returns the public HTTP address assigned by the relay
    PublicURL() string
    // Name returns the provider identifier (e.g. "localtunnel", "selfhosted", "cloudflare")
    Name() string
}
```

### Thread-Safe Service Lifecycle

In real applications, multiple goroutines query tunnel state concurrently (e.g., CLI printers, health probes, shutdown hooks). 

Expose wraps the active provider in a `Service` struct synchronized by `sync.RWMutex` and channels:

```go
// internal/tunnel/service.go
package tunnel

import (
    "context"
    "fmt"
    "sync"
)

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
        return fmt.Errorf("tunnel already started")
    }
    s.started = true
    s.mu.Unlock()

    if _, err := s.provider.Connect(ctx, localPort); err != nil {
        return err
    }

    close(s.ready) // Deterministically signals CLI and caller routines that the URL is live
    return nil
}

func (s *Service) Ready() <-chan struct{} { return s.ready }
func (s *Service) PublicURL() string     { return s.provider.PublicURL() }
func (s *Service) Close() error          { return s.provider.Close() }
```

---

## Part 4 — Modern CLI Ergonomics with Cobra

Expose uses [Cobra](https://github.com/spf13/cobra) to offer an intuitive CLI supporting positional arguments (`expose tunnel 3000`), configuration defaults, and custom flags:

```go
// internal/cli/tunnel.go
package cli

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "github.com/spf13/cobra"
)

func NewTunnelCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "tunnel [port]",
        Short: "Expose local server via tunnel",
        Args:  cobra.MaximumNArgs(1),
        RunE:  runTunnel,
    }

    cmd.Flags().IntP("port", "p", 0, "Local port to expose")
    cmd.Flags().StringP("provider", "P", "localtunnel", "Provider: localtunnel, cloudflare, selfhosted")
    cmd.Flags().StringP("server", "s", "", "Self-hosted server relay (e.g. tunnel.mysite.com:7890)")
    cmd.Flags().StringP("subdomain", "d", "", "Request a custom subdomain")

    return cmd
}

func runTunnel(cmd *cobra.Command, args []string) error {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Listen for SIGINT / SIGTERM for clean teardown
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-sigCh
        fmt.Println("\nGracefully shutting down tunnel...")
        cancel()
    }()

    port := 3000 // Resolved from args, flags, or .expose.yml
    svc := tunnel.NewService(selectedProvider)

    go svc.Start(ctx, port)

    select {
    case <-svc.Ready():
        fmt.Printf("✓ Public URL:    %s\n", svc.PublicURL())
        fmt.Printf("✓ Forwarding to: http://localhost:%d\n", port)
    case <-ctx.Done():
        return ctx.Err()
    }

    <-ctx.Done()
    return svc.Close()
}
```

---

## Part 5 — Advanced Tunneling Patterns

Production tunnel clients must handle connection drops, high traffic spikes, and custom routing. Expose solves these through three core architectural patterns:

### 1. Pre-Warmed Connection Pooling

Dialing a new TCP connection on every incoming request introduces 50–150ms of latency. To achieve high throughput, the client maintains a pool of pre-warmed idle TCP sockets ready to handle requests instantly:

```go
func (lt *LocalTunnel) maintainPool(port int, maxConn int) {
    for i := 0; i < maxConn; i++ {
        go func() {
            for lt.IsConnected() {
                conn, err := net.Dial("tcp", fmt.Sprintf("localtunnel.me:%d", port))
                if err != nil {
                    time.Sleep(time.Second)
                    continue
                }
                lt.handleTraffic(conn) // Handles request and automatically replenishes socket
            }
        }()
    }
}
```

### 2. Native Self-Hosted Relays & Heartbeat Ping-Pong

Rather than relying on third-party servers, Expose implements its own pure-Go relay protocol (`internal/server`):

1. **JSON Control Channel:** The client opens a single persistent TCP connection to the server (`expose server`) and sends a `protocol.TunnelRequest`:
   ```json
   {"type": "control", "subdomain": "my-api"}
   ```
2. **Heartbeat Eviction:** The client transmits a lightweight `PING\n` every 20 seconds. If silent packet loss or a Wi-Fi drop occurs, the server's read deadline (60s) evicts the zombie registration to prevent stale port binding.
3. **On-Demand Data Sockets:** When an external HTTP visitor arrives at the relay, the server signals the client over the control channel to establish a separate data socket for that request.

### 3. Subprocess Orchestration (Cloudflare Quick Tunnels)

When using Cloudflare, Expose does not dial raw TCP sockets. Instead, it manages the lifecycle of the official `cloudflared` binary as a child process and streams logs:

```go
cmd := exec.CommandContext(ctx, "cloudflared", "tunnel", "--url", fmt.Sprintf("http://localhost:%d", port))
stderr, _ := cmd.StderrPipe()
// Read stderr stream to extract the generated trycloudflare.com URL
```

---

## Architecture Summary

| Layer | Responsibility | Location in Expose |
| :--- | :--- | :--- |
| **`Provider` Interface** | Common contract across all tunnel transports | [`internal/tunnel/provider.go`](file:///Users/sam/code/opensource/expose/internal/tunnel/provider.go) |
| **`Service` Wrapper** | Channel synchronization & graceful shutdown | [`internal/tunnel/service.go`](file:///Users/sam/code/opensource/expose/internal/tunnel/service.go) |
| **Self-Hosted Server** | Pure-Go relay, subdomain routing, & heartbeat timeouts | [`internal/server/server.go`](file:///Users/sam/code/opensource/expose/internal/server/server.go) |
| **Self-Hosted Client** | JSON control plane, deferred local dial, & custom subdomains | [`internal/provider/selfhosted.go`](file:///Users/sam/code/opensource/expose/internal/provider/selfhosted.go) |
| **LocalTunnel Client** | Public zero-config TCP socket connection pool | [`internal/provider/localtunnel.go`](file:///Users/sam/code/opensource/expose/internal/provider/localtunnel.go) |
| **Cloudflare Client** | Subprocess supervisor for `cloudflared` | [`internal/provider/cloudflare.go`](file:///Users/sam/code/opensource/expose/internal/provider/cloudflare.go) |
| **CLI & Commands** | Cobra flags, positional port parsing, & signal traps | [`internal/cli/`](file:///Users/sam/code/opensource/expose/internal/cli/) |

---

## Frequently Asked Questions (FAQ)

??? question "Why not just use traditional port forwarding?"
    Port forwarding requires administrative access to the local router, exposes your public home IP address, and fails under Carrier-Grade NAT (CGNAT) where your router lacks a dedicated public IPv4 address. Reverse tunnels bypass NAT transparently by establishing an outbound TCP socket to a public relay.

??? question "How does Expose differ from ngrok?"
    While ngrok is a closed-source SaaS requiring accounts and enforcing usage limits, Expose is an Apache 2.0-licensed, pure-Go tool. It supports zero-signup public relays (LocalTunnel, Cloudflare) as well as its own built-in, private self-hosted relay server (`expose server`).

??? question "How does socket teardown work in full-duplex proxies?"
    When either the visitor disconnects or the local server finishes sending the HTTP response, `io.Copy` unblocks with `io.EOF` or a closed-socket error. The returned channel immediately executes deferred socket closures (`localConn.Close()` and `tunnelConn.Close()`), unblocking the sibling routine and preventing leaked goroutines.

---

## Next Steps

- **[Self-Hosted Server Guide](SELF_HOSTED.md):** Deploy your own private relay server on a $5/mo VPS.
- **[Architecture Deep Dive](ARCHITECTURE.md):** Explore detailed state machines and concurrency models.
- **[Contributing Guide](CONTRIBUTING.md):** Learn how to implement your own custom tunnel provider.

# 📚 Complete Tutorial: Building a Tunnel CLI from Scratch

This tutorial will walk you through understanding Expose by building a simplified version from scratch. You'll learn the core concepts and implementation patterns used in the project.

## Table of Contents

- [Part 1: Understanding Tunnels](#part-1-understanding-tunnels)
- [Part 2: Building a Basic Tunnel](#part-2-building-a-basic-tunnel)
- [Part 3: Adding Provider Abstraction](#part-3-adding-provider-abstraction)
- [Part 4: Building the CLI](#part-4-building-the-cli)
- [Part 5: Advanced Features](#part-5-advanced-features)

---

## Part 1: Understanding Tunnels

### What is a Tunnel?

A tunnel exposes your local server to the internet by:

1. **Creating a connection** to a public server
2. **Forwarding requests** from public URL to localhost
3. **Returning responses** back through the tunnel

```
Internet → Public Server → Tunnel → localhost:3000
```

### How LocalTunnel Works

```
1. Your App:     localhost:3000
                      ↑
2. Tunnel Client:     | (Your code)
                      |
3. TCP Connection: ───┴──→ localtunnel.me:PORT
                      
4. Public URL:    https://xyz.loca.lt → Your App
```

### The Flow

```go
// 1. Request a tunnel
POST localtunnel.me/?new
Response: {id, url, port}

// 2. Open TCP connection
conn = net.Dial("tcp", "localtunnel.me:PORT")

// 3. For each request:
request  → [tunnel conn] → dial localhost:3000 → local server
response ← [tunnel conn] ← read response ← local server
```

---

## Part 2: Building a Basic Tunnel

Let's build a minimal tunnel client step by step.

### Step 1: Request a Tunnel

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
    
    fmt.Printf("Got tunnel: %s (port %d)\n", info.URL, info.Port)
    return &info, nil
}
```

### Step 2: Connect to Tunnel Server

```go
import "net"

func connectTunnel(info *TunnelInfo) (net.Conn, error) {
    address := fmt.Sprintf("localtunnel.me:%d", info.Port)
    conn, err := net.Dial("tcp", address)
    if err != nil {
        return nil, err
    }
    
    fmt.Println("Connected to tunnel server")
    return conn, nil
}
```

### Step 3: Proxy Requests

```go
import "io"

func proxyRequest(tunnelConn net.Conn, localPort int) error {
    // Connect to local server
    localConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", localPort))
    if err != nil {
        return err
    }
    defer localConn.Close()
    
    // Copy data bidirectionally
    done := make(chan error, 2)
    
    // Tunnel → Local
    go func() {
        _, err := io.Copy(localConn, tunnelConn)
        done <- err
    }()
    
    // Local → Tunnel
    go func() {
        _, err := io.Copy(tunnelConn, localConn)
        done <- err
    }()
    
    // Wait for one to finish
    return <-done
}
```

### Step 4: Put It Together

```go
func main() {
    localPort := 3000
    
    // 1. Request tunnel
    info, err := requestTunnel()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Public URL: %s\n", info.URL)
    fmt.Printf("Forwarding to: localhost:%d\n", localPort)
    
    // 2. Connect to tunnel
    conn, err := connectTunnel(info)
    if err != nil {
        panic(err)
    }
    defer conn.Close()
    
    // 3. Handle requests forever
    for {
        if err := proxyRequest(conn, localPort); err != nil {
            fmt.Printf("Error: %v\n", err)
            break
        }
    }
}
```

### Try It!

```bash
# Start a local server
python -m http.server 3000

# In another terminal, run your tunnel
go run basic-tunnel.go
```

You should see:
```
Got tunnel: https://xyz.loca.lt (port 12345)
Public URL: https://xyz.loca.lt
Forwarding to: localhost:3000
Connected to tunnel server
```

Visit the public URL in your browser!

---

## Part 3: Adding Provider Abstraction

Now let's make it extensible with interfaces.

### Step 1: Define Provider Interface

```go
// provider.go
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

### Step 2: Implement LocalTunnel Provider

```go
// localtunnel.go
package tunnel

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net"
    "net/http"
    "sync"
)

type LocalTunnel struct {
    publicURL string
    localPort int
    conn      net.Conn
    mu        sync.RWMutex
}

func NewLocalTunnel() *LocalTunnel {
    return &LocalTunnel{}
}

func (lt *LocalTunnel) Connect(ctx context.Context, localPort int) (string, error) {
    lt.mu.Lock()
    lt.localPort = localPort
    lt.mu.Unlock()
    
    // Request tunnel
    info, err := lt.requestTunnel()
    if err != nil {
        return "", err
    }
    
    lt.mu.Lock()
    lt.publicURL = info.URL
    lt.mu.Unlock()
    
    // Connect to tunnel server
    conn, err := lt.dialTunnel(info.Port)
    if err != nil {
        return "", err
    }
    
    lt.mu.Lock()
    lt.conn = conn
    lt.mu.Unlock()
    
    // Start handling requests
    go lt.handleConnection()
    
    return info.URL, nil
}

func (lt *LocalTunnel) requestTunnel() (*TunnelInfo, error) {
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

func (lt *LocalTunnel) dialTunnel(port int) (net.Conn, error) {
    address := fmt.Sprintf("localtunnel.me:%d", port)
    return net.Dial("tcp", address)
}

func (lt *LocalTunnel) handleConnection() {
    defer lt.conn.Close()
    
    for {
        if err := lt.proxyRequest(); err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }
    }
}

func (lt *LocalTunnel) proxyRequest() error {
    lt.mu.RLock()
    localPort := lt.localPort
    tunnelConn := lt.conn
    lt.mu.RUnlock()
    
    localConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", localPort))
    if err != nil {
        return err
    }
    defer localConn.Close()
    
    var wg sync.WaitGroup
    wg.Add(2)
    
    go func() {
        defer wg.Done()
        io.Copy(localConn, tunnelConn)
    }()
    
    go func() {
        defer wg.Done()
        io.Copy(tunnelConn, localConn)
    }()
    
    wg.Wait()
    return nil
}

func (lt *LocalTunnel) Close() error {
    lt.mu.Lock()
    defer lt.mu.Unlock()
    
    if lt.conn != nil {
        return lt.conn.Close()
    }
    return nil
}

func (lt *LocalTunnel) IsConnected() bool {
    lt.mu.RLock()
    defer lt.mu.RUnlock()
    return lt.conn != nil
}

func (lt *LocalTunnel) PublicURL() string {
    lt.mu.RLock()
    defer lt.mu.RUnlock()
    return lt.publicURL
}

func (lt *LocalTunnel) Name() string {
    return "LocalTunnel"
}
```

### Step 3: Create Service Wrapper

```go
// service.go
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
        return fmt.Errorf("already started")
    }
    s.started = true
    s.mu.Unlock()
    
    _, err := s.provider.Connect(ctx, localPort)
    if err != nil {
        return err
    }
    
    close(s.ready)
    return nil
}

func (s *Service) Ready() <-chan struct{} {
    return s.ready
}

func (s *Service) PublicURL() string {
    return s.provider.PublicURL()
}

func (s *Service) ProviderName() string {
    return s.provider.Name()
}

func (s *Service) Close() error {
    return s.provider.Close()
}
```

### Step 4: Use It

```go
func main() {
    ctx := context.Background()
    
    // Create provider
    provider := NewLocalTunnel()
    
    // Wrap in service
    svc := NewService(provider)
    
    // Start in background
    go svc.Start(ctx, 3000)
    
    // Wait for ready
    <-svc.Ready()
    
    fmt.Printf("Tunnel ready!\n")
    fmt.Printf("Provider: %s\n", svc.ProviderName())
    fmt.Printf("Public URL: %s\n", svc.PublicURL())
    
    // Keep running
    select {}
}
```

---

## Part 4: Building the CLI

Let's add a proper CLI with Cobra.

### Step 1: Install Cobra

```bash
go get github.com/spf13/cobra
```

### Step 2: Create Root Command

```go
// cmd/root.go
package cmd

import (
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "tunnel",
    Short: "Expose localhost to the internet",
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    rootCmd.AddCommand(startCmd)
}
```

### Step 3: Create Start Command

```go
// cmd/start.go
package cmd

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/spf13/cobra"
    "your-module/tunnel"
)

var startCmd = &cobra.Command{
    Use:   "start",
    Short: "Start tunnel",
    RunE:  runStart,
}

var (
    port     int
    provider string
)

func init() {
    startCmd.Flags().IntVarP(&port, "port", "p", 3000, "Local port")
    startCmd.Flags().StringVarP(&provider, "provider", "P", "localtunnel", "Provider")
}

func runStart(cmd *cobra.Command, args []string) error {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Handle Ctrl+C
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-sigChan
        fmt.Println("\nShutting down...")
        cancel()
    }()
    
    // Create provider
    var p tunnel.Provider
    switch provider {
    case "localtunnel":
        p = tunnel.NewLocalTunnel()
    default:
        return fmt.Errorf("unknown provider: %s", provider)
    }
    
    // Create service
    svc := tunnel.NewService(p)
    
    // Start in background
    errChan := make(chan error, 1)
    go func() {
        errChan <- svc.Start(ctx, port)
    }()
    
    // Wait for ready
    select {
    case <-svc.Ready():
        fmt.Printf("✓ Tunnel started\n")
        fmt.Printf("✓ Public URL: %s\n", svc.PublicURL())
        fmt.Printf("✓ Forwarding to: http://localhost:%d\n", port)
    case err := <-errChan:
        return err
    }
    
    // Wait for shutdown
    <-ctx.Done()
    
    // Cleanup
    return svc.Close()
}
```

### Step 4: Main Entry Point

```go
// main.go
package main

import (
    "fmt"
    "os"
    
    "your-module/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Step 5: Build and Run

```bash
# Build
go build -o tunnel .

# Run
./tunnel start --port 3000
```

---

## Part 5: Advanced Features

### Feature 1: Connection Pooling

Instead of one connection, maintain a pool:

```go
type LocalTunnel struct {
    // ...
    connections []net.Conn
    maxConns    int
}

func (lt *LocalTunnel) Connect(ctx context.Context, localPort int) (string, error) {
    // ... request tunnel ...
    
    // Open connection pool
    for i := 0; i < lt.maxConns; i++ {
        conn, err := lt.dialTunnel(info.Port)
        if err != nil {
            return "", err
        }
        lt.connections = append(lt.connections, conn)
        
        // Start handler for each
        go lt.handleConnection(conn)
    }
    
    return info.URL, nil
}
```

### Feature 2: Timeout Management

Set deadlines to prevent hung connections:

```go
func (lt *LocalTunnel) proxyRequest(tunnelConn net.Conn) error {
    // Set deadline
    deadline := time.Now().Add(30 * time.Second)
    tunnelConn.SetDeadline(deadline)
    localConn.SetDeadline(deadline)
    
    // ... proxy ...
    
    // Clear deadline for reuse
    tunnelConn.SetDeadline(time.Time{})
    
    return nil
}
```

### Feature 3: Configuration Files

Add YAML config support:

```go
// config.go
type Config struct {
    Port     int    `yaml:"port"`
    Provider string `yaml:"provider"`
}

func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}
```

### Feature 4: Multiple Providers

Add Cloudflare support:

```go
// cloudflare.go
type Cloudflare struct {
    cmd       *exec.Cmd
    publicURL string
}

func (c *Cloudflare) Connect(ctx context.Context, localPort int) (string, error) {
    // Start cloudflared process
    cmd := exec.CommandContext(
        ctx,
        "cloudflared",
        "tunnel",
        "--url", fmt.Sprintf("http://localhost:%d", localPort),
    )
    
    // Capture output to extract URL
    stderr, _ := cmd.StderrPipe()
    cmd.Start()
    
    // Parse URL from output
    scanner := bufio.NewScanner(stderr)
    for scanner.Scan() {
        line := scanner.Text()
        if match := urlRegex.FindString(line); match != "" {
            c.publicURL = match
            return match, nil
        }
    }
    
    return "", fmt.Errorf("failed to get URL")
}
```

---

## Summary

You've learned:

1. ✅ How tunneling works at a protocol level
2. ✅ How to build a basic tunnel client
3. ✅ How to use interfaces for extensibility
4. ✅ How to build a CLI with Cobra
5. ✅ Advanced features like connection pooling and timeouts

### Next Steps

- Study the full Expose codebase
- Add your own provider implementation
- Contribute to the project!

---

## Additional Resources

- [Expose GitHub](https://github.com/kernelshard/expose)
- [LocalTunnel Protocol](https://github.com/localtunnel/localtunnel)
- [Cloudflare Tunnel Docs](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/)
- [Go Networking](https://go.dev/doc/effective_go#concurrency)

---

**Questions?** Open an issue or discussion on GitHub!

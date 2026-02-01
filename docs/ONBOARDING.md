# 👋 Onboarding Guide for Contributors

Welcome to the Expose project! This guide will help you get up to speed and make your first contribution.

## Table of Contents

- [Welcome](#welcome)
- [Prerequisites](#prerequisites)
- [Development Setup](#development-setup)
- [Understanding the Codebase](#understanding-the-codebase)
- [Making Your First Change](#making-your-first-change)
- [Testing](#testing)
- [Code Style & Standards](#code-style--standards)
- [Pull Request Process](#pull-request-process)
- [Getting Help](#getting-help)

---

## Welcome!

Thank you for your interest in contributing to Expose! Whether you're fixing a bug, adding a feature, or improving documentation, your contribution is valued.

### What is Expose?

Expose is a CLI tool that creates secure tunnels to expose local development servers to the internet. It supports multiple providers (LocalTunnel, Cloudflare) and requires zero configuration to get started.

### Project Goals

- **Simplicity**: Easy to use, minimal dependencies
- **Reliability**: Well-tested, production-ready code
- **Extensibility**: Support for multiple tunnel providers
- **Performance**: Efficient connection handling and low overhead

---

## Prerequisites

Before you start, make sure you have:

### Required

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Git** - Version control
- **A code editor** - VS Code, GoLand, or your favorite editor

### Recommended

- **Make** (optional) - For running common tasks
- **Docker** (optional) - For testing in isolated environments
- **cloudflared** (optional) - For testing Cloudflare provider

### Check Your Setup

```bash
# Verify Go installation
go version
# Expected: go version go1.21 or higher

# Verify Git installation
git --version
# Expected: git version 2.x or higher
```

---

## Development Setup

### Step 1: Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork:

```bash
git clone https://github.com/YOUR_USERNAME/expose.git
cd expose
```

3. Add upstream remote:

```bash
git remote add upstream https://github.com/kernelshard/expose.git
```

### Step 2: Install Dependencies

```bash
# Download dependencies
go mod download

# Verify everything works
go mod verify
```

### Step 3: Build the Project

```bash
# Build the binary
go build -o expose ./cmd/expose

# Run it
./expose --version
```

You should see:

```
expose version X.Y.Z
```

### Step 4: Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/config
```

### Step 5: Set Up Your Editor

#### VS Code

1. Install the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.go)
2. The project includes `.vscode/settings.json` (if not, create it):

```json
{
  "go.testFlags": ["-v"],
  "go.coverOnSave": true,
  "go.lintTool": "golangci-lint",
  "editor.formatOnSave": true
}
```

#### GoLand

1. Open the project directory
2. GoLand will automatically detect it as a Go project
3. Enable "Format on save" in Settings → Tools → Actions on Save

---

## Understanding the Codebase

### Quick Tour (5 minutes)

Let's walk through the key files:

#### 1. Entry Point

```bash
# Look at the main function
cat cmd/expose/main.go
```

This is simple - it just calls `cli.Execute()`.

#### 2. CLI Layer

```bash
# See how commands are structured
cat internal/cli/root.go
cat internal/cli/tunnel.go
```

The CLI uses [Cobra](https://github.com/spf13/cobra) for command parsing.

#### 3. Provider Interface

```bash
# The core abstraction
cat internal/tunnel/provider.go
```

This interface allows us to support multiple tunnel providers.

#### 4. LocalTunnel Implementation

```bash
# The main provider
cat internal/provider/localtunnel.go
```

This is the most complex file - it manages TCP connections and proxying.

#### 5. Configuration

```bash
# Simple YAML config
cat internal/config/config.go
```

Handles `.expose.yml` files.

### Code Organization

```
expose/
├── cmd/expose/          → Entry point
├── internal/
│   ├── cli/            → Commands (tunnel, init, config)
│   ├── config/         → Configuration management
│   ├── tunnel/         → Service layer
│   ├── provider/       → Provider implementations
│   └── version/        → Version info
└── docs/               → Documentation
```

### Key Concepts

#### 1. Provider Pattern

All tunnel providers implement the same interface:

```go
type Provider interface {
    Connect(ctx context.Context, localPort int) (string, error)
    Close() error
    IsConnected() bool
    PublicURL() string
    Name() string
}
```

#### 2. Service Wrapper

The `Service` wraps a provider and adds lifecycle management:

```go
service := tunnel.NewService(provider)
service.Start(ctx, port)
<-service.Ready()
fmt.Println(service.PublicURL())
```

#### 3. Connection Pooling (LocalTunnel)

LocalTunnel maintains 10 concurrent TCP connections to handle requests efficiently.

---

## Making Your First Change

Let's make a simple change to get familiar with the workflow.

### Example 1: Add a New Config Field

Let's add a `timeout` field to the config.

#### Step 1: Update Config Struct

Edit `internal/config/config.go`:

```go
type Config struct {
    Project string `yaml:"project"`
    Port    int    `yaml:"port"`
    Timeout int    `yaml:"timeout"` // Add this line
}
```

#### Step 2: Update Init Function

Still in `config.go`, update `Init()`:

```go
cfg := &Config{
    Project: projectName,
    Port:    3000,
    Timeout: 30, // Add default timeout
}
```

#### Step 3: Update Get Method

Add timeout case in `Get()`:

```go
func (c *Config) Get(key string) (interface{}, error) {
    switch key {
    case "project":
        return c.Project, nil
    case "port":
        return c.Port, nil
    case "timeout":
        return c.Timeout, nil // Add this case
    default:
        return nil, fmt.Errorf("unknown config key: %s", key)
    }
}
```

#### Step 4: Add Tests

Edit `internal/config/config_test.go`:

```go
func TestConfig_Get_Timeout(t *testing.T) {
    cfg := &Config{
        Project: "test",
        Port:    3000,
        Timeout: 30,
    }
    
    val, err := cfg.Get("timeout")
    if err != nil {
        t.Fatalf("Get failed: %v", err)
    }
    
    if val != 30 {
        t.Errorf("got timeout %v, want 30", val)
    }
}
```

#### Step 5: Test Your Changes

```bash
# Run tests
go test ./internal/config

# Build to make sure it compiles
go build ./cmd/expose

# Try it out
./expose init
cat .expose.yml  # Should see timeout: 30
./expose config get timeout  # Should print 30
```

### Example 2: Add a CLI Flag

Let's add a `--timeout` flag to the tunnel command.

#### Step 1: Add Flag

Edit `internal/cli/tunnel.go`:

```go
func newTunnelCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "tunnel",
        Short: "Expose local server via tunnel",
        RunE:  runTunnelCmd,
    }
    
    cmd.Flags().StringP("provider", "P", "localtunnel", "Tunnel provider")
    cmd.Flags().IntP("port", "p", 0, "Local port to expose")
    cmd.Flags().IntP("timeout", "t", 30, "Connection timeout in seconds") // Add this
    
    return cmd
}
```

#### Step 2: Use the Flag

In `runTunnelCmd`:

```go
timeout, err := cmd.Flags().GetInt("timeout")
if err != nil {
    return fmt.Errorf("invalid timeout flag: %w", err)
}

fmt.Printf("Using timeout: %d seconds\n", timeout)
```

#### Step 3: Test

```bash
go build ./cmd/expose
./expose tunnel --help  # Should show --timeout flag
```

---

## Testing

### Test Organization

```
internal/
├── config/
│   ├── config.go
│   └── config_test.go      ← Tests for config.go
├── provider/
│   ├── localtunnel.go
│   ├── localtunnel_test.go ← Tests for localtunnel.go
│   └── ...
```

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/config

# With coverage
go test -cover ./...

# Verbose output
go test -v ./internal/config

# Run specific test
go test -run TestConfig_Load ./internal/config
```

### Writing Tests

#### Example: Simple Unit Test

```go
func TestConfig_Load(t *testing.T) {
    // Create temp config file
    tmpFile := createTempFile(t, "project: test\nport: 3000\n")
    defer os.Remove(tmpFile)
    
    // Load it
    cfg, err := Load(tmpFile)
    if err != nil {
        t.Fatalf("Load failed: %v", err)
    }
    
    // Check values
    if cfg.Port != 3000 {
        t.Errorf("got port %d, want 3000", cfg.Port)
    }
}
```

#### Example: Table-Driven Test

```go
func TestConfig_Get(t *testing.T) {
    cfg := &Config{Project: "test", Port: 3000}
    
    tests := []struct {
        key     string
        want    interface{}
        wantErr bool
    }{
        {"project", "test", false},
        {"port", 3000, false},
        {"invalid", nil, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.key, func(t *testing.T) {
            got, err := cfg.Get(tt.key)
            if (err != nil) != tt.wantErr {
                t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

#### Example: Mock Provider

```go
type mockProvider struct {
    connectErr error
    publicURL  string
}

func (m *mockProvider) Connect(ctx context.Context, port int) (string, error) {
    if m.connectErr != nil {
        return "", m.connectErr
    }
    return m.publicURL, nil
}

func (m *mockProvider) Close() error { return nil }
func (m *mockProvider) IsConnected() bool { return true }
func (m *mockProvider) PublicURL() string { return m.publicURL }
func (m *mockProvider) Name() string { return "mock" }
```

### Test Best Practices

1. **Use table-driven tests** for multiple scenarios
2. **Clean up resources** with `defer`
3. **Use meaningful test names** - `TestFunction_Scenario`
4. **Test error cases** as well as happy paths
5. **Mock external dependencies** (HTTP, exec, etc.)
6. **Keep tests fast** - no unnecessary sleeps

---

## Code Style & Standards

### Go Style Guide

We follow the [official Go style guide](https://go.dev/doc/effective_go). Key points:

#### 1. Formatting

```bash
# Format your code
go fmt ./...

# Or use gofmt
gofmt -w .
```

#### 2. Naming Conventions

```go
// Good
func ProcessRequest()      // Exported: PascalCase
func handleConnection()    // Unexported: camelCase
var MaxConnections = 10    // Exported constant
var defaultTimeout = 30    // Unexported

// Avoid
func process_request()     // No underscores
func HandleConnection()    // Don't export unless needed
```

#### 3. Error Handling

```go
// Good
if err != nil {
    return fmt.Errorf("failed to connect: %w", err)
}

// Avoid
if err != nil {
    panic(err)  // Don't panic in library code
}
```

#### 4. Comments

```go
// Package config provides configuration file management for Expose.
package config

// Load reads and parses the configuration file at the given path.
// If path is empty, it uses the default ".expose.yml" file.
func Load(path string) (*Config, error) {
    // Implementation
}
```

#### 5. Struct Tags

```go
type Config struct {
    Project string `yaml:"project"`  // Lowercase for YAML
    Port    int    `yaml:"port"`
}
```

### Linting

```bash
# Install golangci-lint
# See: https://golangci-lint.run/usage/install/

# Run linter
golangci-lint run

# Fix auto-fixable issues
golangci-lint run --fix
```

### Commit Messages

Use conventional commits:

```
feat: add timeout configuration option
fix: clear TCP deadline after each request
docs: add architecture guide
test: add tests for config package
refactor: simplify error handling in tunnel service
```

Format:
```
<type>: <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `style`, `chore`

---

## Pull Request Process

### Step 1: Create a Branch

```bash
# Get latest changes
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feat/add-timeout-config
```

### Step 2: Make Changes

```bash
# Edit files
# Add tests
# Run tests
go test ./...

# Format code
go fmt ./...
```

### Step 3: Commit Changes

```bash
git add .
git commit -m "feat: add timeout configuration option"
```

### Step 4: Push to Your Fork

```bash
git push origin feat/add-timeout-config
```

### Step 5: Create Pull Request

1. Go to GitHub
2. Click "New Pull Request"
3. Select your branch
4. Fill out the template:

```markdown
## Description
Adds timeout configuration option to control connection timeouts.

## Changes
- Added `timeout` field to Config struct
- Added `--timeout` flag to tunnel command
- Updated documentation

## Testing
- Added unit tests for config timeout
- Manually tested with various timeout values

## Related Issues
Closes #123
```

### Step 6: Address Review Comments

```bash
# Make requested changes
# Commit them
git commit -m "fix: address review comments"
git push origin feat/add-timeout-config
```

### PR Checklist

Before submitting:

- [ ] Tests pass (`go test ./...`)
- [ ] Code is formatted (`go fmt ./...`)
- [ ] No linting errors
- [ ] Tests added for new features
- [ ] Documentation updated
- [ ] Commit messages follow convention
- [ ] Branch is up to date with main

---

## Getting Help

### Resources

1. **Documentation**
   - [Getting Started](GETTING_STARTED.md)
   - [Architecture Guide](ARCHITECTURE.md)
   - [Contributing Guide](../CONTRIBUTING.md)

2. **Code Examples**
   - Look at existing tests for examples
   - Study similar features in the codebase

3. **Community**
   - [GitHub Discussions](https://github.com/kernelshard/expose/discussions)
   - [GitHub Issues](https://github.com/kernelshard/expose/issues)

### Asking Questions

When asking for help:

1. **Search existing issues** - Your question might be answered
2. **Provide context** - What are you trying to do?
3. **Include code** - Show what you've tried
4. **Share errors** - Include full error messages

Good question format:

```markdown
**What I'm trying to do:**
Add a new provider for XYZ tunnel service.

**What I've tried:**
- Created xyz.go with Provider implementation
- Added tests but getting error...

**Error:**
```
panic: nil pointer dereference
```

**Code:**
[Link to your branch or paste code]
```

### Common Issues

#### Issue: Tests Fail on CI

**Solution**: Run tests locally with race detector:
```bash
go test -race ./...
```

#### Issue: Import Cycle

**Solution**: Review package dependencies. Use interfaces to break cycles.

#### Issue: Can't Build

**Solution**: Clean and rebuild:
```bash
go clean -cache
go mod tidy
go build ./cmd/expose
```

---

## Next Steps

### After Your First PR

1. 🎯 **Pick another issue** - Look for "good first issue" label
2. 📚 **Read Architecture doc** - Understand the design deeply
3. 🔨 **Try a bigger feature** - Add a new provider, improve error handling
4. 📖 **Improve docs** - Documentation is always appreciated

### Ideas for Contributions

**Easy**:
- Add more config options
- Improve error messages
- Add CLI flag aliases
- Write more tests
- Fix typos in docs

**Medium**:
- Add new CLI commands
- Improve logging
- Add metrics/stats
- Better error handling
- Performance optimizations

**Advanced**:
- Add new tunnel provider
- Implement connection health checks
- Add custom domain support
- Multi-tunnel management
- Plugin system

---

## Thank You!

Thank you for contributing to Expose! Every contribution, no matter how small, helps make the project better for everyone.

**Happy coding!** 🚀

---

**Questions?** Feel free to:
- Open a [GitHub Discussion](https://github.com/kernelshard/expose/discussions)
- Comment on an existing issue
- Reach out to maintainers

We're here to help! 👋

# 🚀 Expose

![Expose Social Preview](docs/social-preview.png)

[![Downloads](https://img.shields.io/github/downloads/kernelshard/expose/total?color=blue&logo=github)](https://github.com/kernelshard/expose/releases)
[![Documentation](https://img.shields.io/badge/docs-mkdocs--material-blue.svg)](https://kernelshard.github.io/expose/)
[![Tests](https://github.com/kernelshard/expose/actions/workflows/test.yml/badge.svg)](https://github.com/kernelshard/expose/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kernelshard/expose)](https://goreportcard.com/report/github.com/kernelshard/expose)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://github.com/kernelshard/expose/blob/main/LICENSE)

> **The open-source, pure Go (Golang) alternative to ngrok.**
> Expose your local `localhost` server to the internet with zero config, zero signup, and zero hassle.

**Expose** is a modern Go-based tunneling tool that lets you share your local development environment with the world. Perfect for testing webhooks, mobile debugging, or showing off your work on `localtunnel` or `Cloudflare` networks.

## 🆚 Why Expose?

| Feature | 🚀 **Expose** | ngrok | Cloudflare Tunnel | LocalTunnel (Node) |
| :--- | :---: | :---: | :---: | :---: |
| **No Signup Required** | ✅ | ❌ | ❌ | ✅ |
| **Single Binary** | ✅ | ✅ | ✅ | ❌ (requires Node.js) |
| **Open Source** | ✅ | ❌ | ✅ | ✅ |
| **Self-Hosted Server** | ✅ | ❌ | ❌ | ❌ |
| **Language** | Go | Go | Go | JS |
| **Free Custom Domains** | 🚧 (Planned) | 💲 Paid | ✅ | ❌ |

## ✨ Features

- 🌐 **Multiple Providers**: LocalTunnel, Cloudflare, or your own **self-hosted server**.
- 🏠 **Self-Hosted**: Run `expose server` on any $5 VPS — full control, no third parties.
- ⚡ **Zero Friction**: No accounts, no auth tokens, just run and go.
- 📦 **Lightweight**: A single <10MB binary with no external dependencies.
- 🔧 **Developer Friendly**: Written in pure Go, easy to contribute to.

---

## 🚀 Quick Start

### 1. Install

```bash
# macOS & Linux (Homebrew — recommended)
brew install kernelshard/tap/expose

# Or via one-line shell installer
curl -fsSL https://raw.githubusercontent.com/kernelshard/expose/main/install.sh | sh
```

### 2. Expose a Local Port

```bash
# Expose any local port instantly (zero config required!)
expose tunnel 3000
```

> No signup, API tokens, or configuration files needed. Your localhost is instantly live on a secure public HTTPS URL.

### ⚡ Other Providers

```bash
# Use Cloudflare Quick Tunnels
expose tunnel -P cloudflare -p 8080

# Or connect to your private self-hosted server
expose tunnel 3000 --server=tunnel.mysite.com:7890
```

---

### 📦 Installation

#### 1. Homebrew (macOS & Linux)

```bash
brew install kernelshard/tap/expose
```

#### 2. Shell Script (macOS & Linux)

Install the latest pre-compiled binary with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/kernelshard/expose/main/install.sh | sh
```

You can customize the destination directory or pin a specific version:

```bash
# Install a specific version
VERSION=v0.5.0 curl -fsSL https://raw.githubusercontent.com/kernelshard/expose/main/install.sh | sh

# Install into a custom directory (defaults to /usr/local/bin or ~/.local/bin)
BINDIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/kernelshard/expose/main/install.sh | sh
```

#### 3. Pre-built Binaries
Download the latest binary for your operating system (macOS, Linux, Windows) from the [Releases Page](https://github.com/kernelshard/expose/releases). Extract the archive and place `expose` in your `$PATH`.

#### 4. Using Go Install

Requires Go 1.21+:

```bash
go install github.com/kernelshard/expose/cmd/expose@latest
```

### From Source

```bash
git clone https://github.com/kernelshard/expose.git
cd expose
go build -o expose ./cmd/expose
./expose --version
```

---

## 📖 Usage

### Initialize Configuration

```bash
$ expose init
✓ Config created: .expose.yml
```

Creates `.expose.yml` in current directory:

```yaml
project: expose
port: 3000
```

### Start Tunnel

```bash
# Use config port
$ expose tunnel
✓ Tunnel (LocalTunnel) started for localhost:3000
✓ Public URL: https://quick-mammals-sing.loca.lt
✓ Forwarding to http://localhost:3000
✓ Provider: LocalTunnel
✓ Press Ctrl+C to stop

# Override port
$ expose tunnel --port 8080
```

### Manage Configuration

```bash
# List all config values
$ expose config list
project: expose
port: 3000

# Get specific value
$ expose config get port
3000

$ expose config get project
expose
```

---

## ✅ Tested Locally

```bash
$ expose --version
expose version v0.1.2 (commit: d30c483, built: 2025-11-10)

$ expose init
✓ Config created: .expose.yml (project: expose, port: 3000)

$ python3 -m http.server 3000 &
Serving HTTP on 0.0.0.0 port 3000...

$ expose tunnel
🚀 Tunnel[LocalTunnel] started for localhost:3000
✓ Public URL: https://ripe-garlics-add.loca.lt
✓ Forwarding to: http://localhost:3000
✓ Provider: LocalTunnel
Press Ctrl+C to stop

$ curl https://ripe-garlics-add.loca.lt
<!DOCTYPE HTML>...  # Works!
```

**Tested on:** Go 1.23, macOS 14, Ubuntu 22.04

---

## 🏗 Architecture

```text
expose/
├── cmd/expose/       # CLI entry point
├── internal/
│   ├── cli/          # Cobra commands (thin layer)
│   ├── config/       # YAML config management
│   ├── provider/     # Tunnel provider implementations
│   ├── server/       # Self-hosted tunnel server
│   ├── tunnel/       # Service layer (business logic)
│   └── version/      # Version metadata
└── .expose.yml       # User config (add to .gitignore per project)
```

**Design principles:**
- **Interface-driven** — `Provider` interface supports multiple tunnel backends
- **Clean separation** — CLI → Service → Provider (no circular deps)
- **Testable** — Real file tests, injectable service layer

---

## ⚠️ Known Limitations
- **One tunnel per process** — Each `expose tunnel` command runs independently (can run multiple on different ports)
- **No persistence** — Public URLs change on restart
- **CLI-only** — No web UI or dashboard yet

See [GitHub Issues](https://github.com/kernelshard/expose/issues) for roadmap.

---

## 🧪 Development

### Prerequisites
- Go 1.23+
- Git

### Setup

```bash
git clone https://github.com/kernelshard/expose.git
cd expose
go mod download
```

### Run Tests

```bash
# Run all tests with race detector
go test ./... -v -race -cover

# Check coverage for specific packages
go test ./internal/config -cover
go test ./internal/tunnel -cover
go test ./internal/provider -cover
```

### Build

```bash
go build -o expose ./cmd/expose
./expose --version
```

### Run Locally

```bash
# Without installing
go run cmd/expose/main.go tunnel

# Test with live server
python3 -m http.server 3000  # Terminal 1
./expose tunnel              # Terminal 2
```

---

## 🤝 Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development workflow
- Branch strategy
- Testing requirements
- Code style guidelines

---

## 📝 License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

---

**Made with ❤️ by [@kernelshard](https://github.com/kernelshard)**

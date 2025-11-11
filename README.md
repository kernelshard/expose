# 🚀 Expose

[![Tests](https://github.com/kernelshard/expose/actions/workflows/test.yml/badge.svg)](https://github.com/kernelshard/expose/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kernelshard/expose)](https://goreportcard.com/report/github.com/kernelshard/expose)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/kernelshard/expose/blob/main/LICENSE)

> Minimal CLI tool to expose your local dev server to the internet

**Expose** lets you share your `localhost` with the world — perfect for testing webhooks, demoing work, or debugging on mobile devices. Built as a lightweight alternative to ngrok, powered by LocalTunnel.

## ✨ Features

- 🌐 **Instant public URLs** — Share localhost with one command
- ⚡ **Zero signup** — No accounts, no registration required
- � **Config management** — Save port settings per project
- 📦 **Single binary** — No Node.js, Python, or runtime dependencies
- 🧪 **Production-tested** — 75%+ test coverage, CI/CD pipeline

---

## 🚀 Quick Start

```bash
# Install
go install github.com/kernelshard/expose/cmd/expose@latest

# Initialize config
expose init

# Start tunnel
expose tunnel
```

---

## 📦 Installation

### Using Go Install

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
│   ├── provider/     # Tunnel provider interface
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

- **LocalTunnel only** — ngrok/Cloudflare support planned for v0.2.0
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

MIT License - see [LICENSE](LICENSE) for details.

---

**Made with ❤️ by [@kernelshard](https://github.com/kernelshard)**

# expose

> Expose localhost to the internet. Minimal. Fast. Open Source.

**expose** is a lightweight Golang CLI that makes sharing your local development server effortless.

## ✨ Features

- 🚀 **One Command**: Expose your local server instantly
- ⚙️ **Zero Config**: Works out of the box with sensible defaults
- 🔒 **Privacy First**: Self-hostable, no vendor lock-in
- 🎯 **Minimal**: Single binary, no runtime dependencies

## 📦 Installation

```
# Clone the repository
git clone https://github.com/yourusername/expose.git
cd expose

# Build
go build -o expose cmd/expose/main.go

# Optional: Install globally
go install github.com/yourusername/expose/cmd/expose@latest
```

## 🚀 Quick Start

```
# 1. Initialize configuration
expose init

# 2. Expose your local server
expose tunnel

# 3. Access via http://localhost:8080
```

## 📖 Usage

### Initialize Project

Create a `.expose.yml` configuration file:

```
expose init
```

This generates:
```
project: my-app
default_port: 3000
```

### Expose Local Server

Start exposing your local development server:

```
# Use default port from config
expose tunnel

# Specify custom port
expose tunnel --port 8080
```

## ⚙️ Configuration

Edit `.expose.yml` to customize settings:

```
project: "my-awesome-app"
default_port: 3000
```

## 🏗️ Architecture

```
expose/
├── cmd/expose/          # CLI entry point
└── internal/
    ├── cli/             # Command implementations
    └── config/          # Configuration management
```

**Design Principles:**
- Idiomatic Go code
- Clean architecture
- Minimal dependencies
- Easy to contribute

## 🛠️ Development

```
# Install dependencies
go mod download

# Run locally
go run cmd/expose/main.go init

# Build
go build -o expose cmd/expose/main.go

# Format code
go fmt ./...
```

## 🗺️ Roadmap

- [x] Basic tunnel functionality
- [ ] Localtunnel/ngrok-style public URLs
- [ ] Branch-aware environment switching
- [ ] PR preview environments
- [ ] Custom tunnel server support

## 🤝 Contributing

Contributions welcome! This project follows:
- Standard Go conventions
- Commit message format: `type: description`
- Clean, tested, documented code

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- Go standard library - Minimal and powerful

---

**Status:** Early development - contributions welcome!

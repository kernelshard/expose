# 📋 Documentation Index

Welcome to the Expose documentation! This index will help you find what you need.

## 🚀 Getting Started

**New to Expose?** Start here:

- **[Getting Started Guide](GETTING_STARTED.md)** - Installation, first tunnel, common use cases
  - Installation methods
  - Your first tunnel in 3 steps
  - Testing webhooks and mobile devices
  - Troubleshooting basics

## 🏗️ Understanding Expose

**Want to understand how it works?**

- **[Architecture Guide](ARCHITECTURE.md)** - Deep dive into the design and implementation
  - System architecture
  - Project structure
  - Core components explained
  - Data flow diagrams
  - Concurrency model
  - Provider implementation guide

- **[Tutorial: Building a Tunnel](TUTORIAL.md)** - Learn by building a simplified version
  - Understanding tunnels
  - Building from scratch
  - Provider abstraction
  - CLI development
  - Advanced features

## 👥 Contributing

**Want to contribute?**

- **[Onboarding Guide](ONBOARDING.md)** - Complete contributor onboarding
  - Development setup
  - Understanding the codebase
  - Making your first change
  - Testing guide
  - Code style standards
  - Pull request process

- **[Contributing Guidelines](../CONTRIBUTING.md)** - Project contribution guidelines
  - Code of conduct
  - How to report bugs
  - Feature requests
  - Development workflow

## 🔧 Advanced Usage

**Already using Expose?**

- **[Advanced Usage Guide](ADVANCED_USAGE.md)** - Power user features
  - Configuration management
  - Provider selection
  - Performance tuning
  - Security considerations
  - Integration examples
  - Tips & tricks

## 📖 Quick Reference

### Common Commands

```bash
# Initialize config
expose init

# Start tunnel (uses config)
expose tunnel

# Start with specific port
expose tunnel --port 8080

# Use Cloudflare provider
expose tunnel --provider cloudflare

# View configuration
expose config list
expose config get port

# Help
expose --help
expose tunnel --help
```

### File Structure

```
expose/
├── cmd/expose/          → Entry point
├── internal/
│   ├── cli/            → Commands
│   ├── config/         → Configuration
│   ├── tunnel/         → Service layer
│   ├── provider/       → Providers
│   └── version/        → Version info
└── docs/               → Documentation (you are here!)
```

### Configuration File (`.expose.yml`)

```yaml
project: my-app
port: 3000
```

## 🎯 Quick Navigation

### By Task

| I want to... | Go to... |
|--------------|----------|
| Install Expose | [Getting Started → Installation](GETTING_STARTED.md#installation) |
| Create my first tunnel | [Getting Started → Your First Tunnel](GETTING_STARTED.md#your-first-tunnel) |
| Test webhooks | [Getting Started → Webhooks](GETTING_STARTED.md#testing-webhooks) |
| Understand the architecture | [Architecture Guide](ARCHITECTURE.md) |
| Add a new feature | [Onboarding → Making Changes](ONBOARDING.md#making-your-first-change) |
| Write tests | [Onboarding → Testing](ONBOARDING.md#testing) |
| Build from scratch | [Tutorial](TUTORIAL.md) |
| Use advanced features | [Advanced Usage](ADVANCED_USAGE.md) |
| Troubleshoot issues | [Advanced Usage → Troubleshooting](ADVANCED_USAGE.md#troubleshooting) |
| Contribute code | [Onboarding Guide](ONBOARDING.md) |

### By Role

#### User
1. [Getting Started](GETTING_STARTED.md)
2. [Advanced Usage](ADVANCED_USAGE.md)

#### Developer/Contributor
1. [Onboarding Guide](ONBOARDING.md)
2. [Architecture Guide](ARCHITECTURE.md)
3. [Tutorial](TUTORIAL.md)
4. [Contributing Guidelines](../CONTRIBUTING.md)

#### Maintainer
1. [Architecture Guide](ARCHITECTURE.md)
2. [Contributing Guidelines](../CONTRIBUTING.md)

## 🔍 Detailed Contents

### Getting Started Guide

- What is Expose?
- Installation (3 methods)
- Your first tunnel (3 steps)
- Common use cases
  - Testing webhooks
  - Mobile device testing
  - Demoing to clients
- Configuration tips
- Troubleshooting
- Next steps

### Architecture Guide

- System overview
- Architecture diagrams
- Project structure
- Core components
  - CLI layer
  - Config management
  - Tunnel service
  - Provider layer
- Data flow
- Design patterns
- Provider implementation
- Concurrency model
- Error handling
- Testing strategy
- Performance considerations

### Tutorial

- Part 1: Understanding tunnels
- Part 2: Building a basic tunnel
- Part 3: Adding provider abstraction
- Part 4: Building the CLI
- Part 5: Advanced features

### Onboarding Guide

- Prerequisites
- Development setup (step by step)
- Understanding the codebase
- Making your first change
- Testing guide
- Code style & standards
- Pull request process
- Getting help
- Next steps

### Advanced Usage Guide

- Configuration management
- Provider selection
- Performance tuning
- Security considerations
- Troubleshooting guide
- Integration examples
  - GitHub webhooks
  - Stripe webhooks
  - Mobile testing
  - API development
- Tips & tricks

## 📚 Additional Resources

### External Links

- [Expose GitHub Repository](https://github.com/kernelshard/expose)
- [Issue Tracker](https://github.com/kernelshard/expose/issues)
- [Discussions](https://github.com/kernelshard/expose/discussions)
- [LocalTunnel Documentation](https://github.com/localtunnel/localtunnel)
- [Cloudflare Tunnel Documentation](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/)

### Go Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)
- [Go Standard Library](https://pkg.go.dev/std)

### Tools

- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Go Testing](https://pkg.go.dev/testing)

## ❓ FAQ

### General

**Q: Do I need to sign up for anything?**
A: No! Expose requires zero signup or authentication.

**Q: Is Expose free?**
A: Yes, completely free and open source.

**Q: What providers are supported?**
A: LocalTunnel and Cloudflare Tunnel, with more coming soon.

### Technical

**Q: Can I use custom domains?**
A: Not currently, but it's on the roadmap.

**Q: Is it safe for production?**
A: No, Expose is for development and testing only.

**Q: How many concurrent connections?**
A: LocalTunnel supports ~10, Cloudflare supports 100+.

### Contributing

**Q: How can I contribute?**
A: Read the [Onboarding Guide](ONBOARDING.md) and check [open issues](https://github.com/kernelshard/expose/issues).

**Q: What's a good first contribution?**
A: Look for issues labeled "good first issue" or improve documentation.

**Q: Who maintains Expose?**
A: Check the [contributors page](https://github.com/kernelshard/expose/graphs/contributors).

## 🤝 Getting Help

### Questions?

1. Check the relevant guide above
2. Search [existing issues](https://github.com/kernelshard/expose/issues)
3. Ask in [Discussions](https://github.com/kernelshard/expose/discussions)
4. Open a new issue

### Found a Bug?

1. Check if it's already [reported](https://github.com/kernelshard/expose/issues)
2. Include:
   - Expose version (`expose --version`)
   - Operating system
   - Steps to reproduce
   - Expected vs actual behavior
   - Logs/errors

### Want a Feature?

1. Check [existing requests](https://github.com/kernelshard/expose/issues?q=is%3Aissue+label%3Aenhancement)
2. Open a feature request with:
   - Use case description
   - Proposed solution
   - Alternatives considered

---

## 📝 Documentation Updates

This documentation is continuously improved. Last updated: January 2026.

**Found an issue in the docs?** Please [open an issue](https://github.com/kernelshard/expose/issues) or submit a PR!

---

**Happy tunneling!** 🚀

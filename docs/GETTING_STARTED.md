# 🚀 Getting Started with Expose

Welcome to **Expose**! This guide will help you get up and running in minutes.

## Table of Contents

- [What is Expose?](#what-is-expose)
- [Installation](#installation)
- [Your First Tunnel](#your-first-tunnel)
- [Common Use Cases](#common-use-cases)
- [Next Steps](#next-steps)

---

## What is Expose?

**Expose** is a lightweight CLI tool that creates secure tunnels to expose your local development server to the internet. It's perfect for:

- 🔗 **Testing webhooks** from services like GitHub, Stripe, or Twilio
- 📱 **Mobile testing** your localhost on real devices
- 🎨 **Demoing projects** to clients or teammates without deployment
- 🔍 **Debugging** remote services that need to reach your local environment

Unlike other tunneling tools, Expose:
- ✅ Requires **zero signup or authentication**
- ✅ Works with **multiple providers** (LocalTunnel, Cloudflare)
- ✅ Is a **single binary** with no runtime dependencies
- ✅ Supports **project-based configuration** for easy reuse

---

## Installation

### Option 1: Using Go Install (Recommended)

If you have Go installed (1.21+):

```bash
go install github.com/kernelshard/expose/cmd/expose@latest
```

Verify the installation:

```bash
expose --version
```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/kernelshard/expose.git
cd expose

# Build the binary
go build -o expose ./cmd/expose

# Move to your PATH (optional)
sudo mv expose /usr/local/bin/

# Verify
expose --version
```

### Option 3: Download Pre-built Binary

Visit the [releases page](https://github.com/kernelshard/expose/releases) and download the appropriate binary for your system.

---

## Your First Tunnel

Let's expose a local server in 3 simple steps:

### Step 1: Start Your Local Server

First, make sure you have a local server running. For example:

```bash
# Python
python -m http.server 3000

# Node.js
npx http-server -p 3000

# Or any framework on port 3000
```

### Step 2: Initialize Configuration

In your project directory, run:

```bash
expose init
```

This creates a `.expose.yml` file:

```yaml
project: my-project
port: 3000
```

**Pro tip**: Add `.expose.yml` to your git repository so your team can use the same settings!

### Step 3: Start the Tunnel

```bash
expose tunnel
```

You'll see output like:

```
🚀 Tunnel[LocalTunnel] started for localhost:3000
✓ Public URL: https://quick-mammals-sing.loca.lt
✓ Forwarding to: http://localhost:3000
✓ Provider: LocalTunnel
Press Ctrl+C to stop
```

**That's it!** 🎉 Your local server is now accessible at the public URL.

### Stopping the Tunnel

Press `Ctrl+C` in the terminal to stop the tunnel:

```
^C
Shutting down...
✓ Tunnel closed
```

---

## Common Use Cases

### Testing Webhooks

Many services like GitHub, Stripe, or Twilio need to send HTTP requests to your server. Use Expose to give them a public URL:

```bash
# Start your webhook receiver on port 4000
node webhook-server.js

# In another terminal, expose it
expose tunnel --port 4000

# Use the public URL in your webhook configuration
# Example: https://your-tunnel.loca.lt/webhook
```

### Mobile Device Testing

Test your responsive design on real devices:

```bash
# Start your dev server
npm run dev  # or whatever starts your local server

# Expose it
expose tunnel

# Open the public URL on your phone/tablet
# Example: https://brave-lions-jump.loca.lt
```

### Demo to Client

Share your work-in-progress without deploying:

```bash
# Start your app
npm start

# Create tunnel with specific port
expose tunnel --port 3000

# Share the URL with your client
```

### Using Different Providers

Expose supports multiple tunnel providers:

#### LocalTunnel (Default)
- No installation required
- No signup needed
- Good for quick testing

```bash
expose tunnel --provider localtunnel
```

#### Cloudflare Tunnel
- More reliable for production demos
- Requires `cloudflared` binary installed
- Better performance

```bash
# Install cloudflared first
# macOS: brew install cloudflare/cloudflare/cloudflared
# Linux: https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation

# Use Cloudflare provider
expose tunnel --provider cloudflare
```

---

## Configuration Tips

### Override Port

You can override the config file port with a flag:

```bash
expose tunnel --port 8080
```

### Per-Project Configuration

Create `.expose.yml` in each project:

```yaml
# Frontend project
project: my-frontend
port: 3000
```

```yaml
# Backend API
project: my-api
port: 4000
```

### View Configuration

Check your current config:

```bash
expose config list
```

Get a specific value:

```bash
expose config get port
```

---

## Troubleshooting

### "Config not found" Error

```bash
Error: config not found (run 'expose init' first)
```

**Solution**: Run `expose init` in your project directory first.

### "Port already in use" Error

**Solution**: Either stop the process using that port or choose a different port:

```bash
expose tunnel --port 8080
```

### Tunnel Connection Fails

**Solution**: Try a different provider:

```bash
# If LocalTunnel fails, try Cloudflare
expose tunnel --provider cloudflare
```

### Cloudflare Provider Not Working

**Solution**: Make sure `cloudflared` is installed:

```bash
# macOS
brew install cloudflare/cloudflare/cloudflared

# Linux
wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64
chmod +x cloudflared-linux-amd64
sudo mv cloudflared-linux-amd64 /usr/local/bin/cloudflared
```

---

## Next Steps

Now that you're familiar with the basics:

1. 📖 **[Architecture Guide](ARCHITECTURE.md)** - Understand how Expose works under the hood
2. 🛠️ **[Developer Guide](CONTRIBUTING.md)** - Contribute to the project
3. 🎯 **[Advanced Usage](ADVANCED_USAGE.md)** - Learn advanced features and tips
4. 🐛 **[Troubleshooting](TROUBLESHOOTING.md)** - Common issues and solutions

---

## Quick Reference

```bash
# Initialize project config
expose init

# Start tunnel (uses config)
expose tunnel

# Start tunnel with specific port
expose tunnel --port 8080

# Use Cloudflare provider
expose tunnel --provider cloudflare

# Short flags
expose tunnel -P cloudflare -p 3000

# View config
expose config list
expose config get port

# Help
expose --help
expose tunnel --help
```

---

**Need help?** Open an issue on [GitHub](https://github.com/kernelshard/expose/issues) or check the [FAQ](FAQ.md).

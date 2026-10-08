---
title: Getting Started
description: Install Expose and run your first tunnel in under 2 minutes.
---

# Getting Started

**Expose** is a lightweight CLI tool that creates secure tunnels so your `localhost` is reachable from the internet — perfect for webhook testing, mobile debugging, or sharing work-in-progress.

---

## Installation

=== "Go Install (Recommended)"
    Requires **Go 1.21+**. This installs the latest release directly into your `$GOPATH/bin`:
    ```bash
    go install github.com/kernelshard/expose/cmd/expose@latest
    ```

=== "Pre-built Binary"
    Download the binary for your OS from the [Releases page](https://github.com/kernelshard/expose/releases):
    ```bash
    # macOS / Linux — make executable and move to PATH
    chmod +x expose-linux-amd64
    sudo mv expose-linux-amd64 /usr/local/bin/expose
    ```
    Windows users: add the `.exe` file to a directory in `%PATH%`.

=== "Build from Source"
    ```bash
    git clone https://github.com/kernelshard/expose.git
    cd expose
    go build -o expose ./cmd/expose

    # Optional: move to PATH
    sudo mv expose /usr/local/bin/
    ```

Verify the installation:

```bash
expose --version
```

---

## Your First Tunnel

### Step 1 — Start a local server

You need something running on localhost. Any of these work:

```bash
python -m http.server 3000   # Python 3
npx http-server -p 3000      # Node.js
npm run dev                  # Any framework
```

### Step 2 — Initialize project config

In your project directory:

```bash
expose init
```

This creates a `.expose.yml` file:

```yaml
project: my-project
port: 3000
```

!!! tip "Commit `.expose.yml` to git"
    Share your tunnel settings with your team — everyone gets the same port and project name automatically.

### Step 3 — Start the tunnel

```bash
expose tunnel
```

Output:

```text
✓ Tunnel (LocalTunnel) started for localhost:3000
✓ Public URL: https://quick-mammals-sing.loca.lt
✓ Forwarding to: http://localhost:3000
✓ Provider: LocalTunnel
  Press Ctrl+C to stop
```

**That's it.** Your local server is now publicly accessible. :tada:

Press ++ctrl+c++ to stop the tunnel:

```text
^C
Shutting down...
✓ Tunnel closed
```

---

## Common Use Cases

### Testing Webhooks

Services like GitHub, Stripe, and Twilio need to reach your local handler via HTTP. Use Expose to give them a public URL:

```bash
# Start your webhook receiver
node webhook-server.js        # e.g. listening on port 4000

# In another terminal
expose tunnel -p 4000

# Use the public URL in your service's webhook settings
# e.g. https://your-tunnel.loca.lt/webhook
```

### Mobile Device Testing

Test responsive design on real devices without deploying:

```bash
expose tunnel
# Open the public URL on your phone or tablet
```

### Client Demo

Share work-in-progress without a staging environment:

```bash
# Use Cloudflare for a more reliable connection during demos
expose tunnel -P cloudflare -p 3000
```

---

## Provider Options

### LocalTunnel (default)

No dependencies, no signup — just works. Great for quick webhook tests.

```bash
expose tunnel
# same as: expose tunnel --provider localtunnel
```

**Limitations**: shared infrastructure, occasional connection drops on long sessions.

### Cloudflare Tunnel

More reliable for demos and longer sessions. Requires `cloudflared` to be installed:

=== "macOS"
    ```bash
    brew install cloudflare/cloudflare/cloudflared
    ```

=== "Linux"
    ```bash
    wget -q https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64
    sudo mv cloudflared-linux-amd64 /usr/local/bin/cloudflared
    sudo chmod +x /usr/local/bin/cloudflared
    ```

Then:

```bash
expose tunnel --provider cloudflare
# short form: expose tunnel -P cloudflare
```

---

## Configuration Reference

### View config

```bash
expose config list        # all settings
expose config get port    # a specific key
```

### Override port on the fly

The `-p` / `--port` flag overrides whatever is set in `.expose.yml`:

```bash
expose tunnel -p 8080
```

### Per-project configs

Each project directory can have its own `.expose.yml`:

```yaml
# ~/projects/frontend/.expose.yml
project: frontend
port: 3000
```

```yaml
# ~/projects/api/.expose.yml
project: api
port: 4000
```

Switch between them by `cd`-ing into the right directory before running `expose tunnel`.

---

## Troubleshooting

### `Error: config not found`

```text
Error: config not found (run 'expose init' first)
```

**Fix**: Run `expose init` in your project directory.

### Port already in use

```bash
# Find what's using the port
lsof -i :3000

# Kill it or pick a different port
expose tunnel -p 8080
```

### Tunnel disconnects frequently

LocalTunnel can be unstable. Switch to Cloudflare:

```bash
expose tunnel -P cloudflare
```

### `cloudflared: command not found`

See [Cloudflare Tunnel](#cloudflare-tunnel) installation steps above.

---

## Quick Reference

```bash
expose init                          # Create .expose.yml in current directory
expose tunnel                        # Start tunnel (uses config)
expose tunnel -p 8080                # Override port
expose tunnel -P cloudflare          # Use Cloudflare provider
expose tunnel -P cloudflare -p 3000  # Cloudflare on a specific port
expose config list                   # Show current config
expose config get port               # Get a single config value
expose --help                        # Help for any command
expose tunnel --help
```

---

**Need help?** [Open an issue on GitHub](https://github.com/kernelshard/expose/issues) — we're happy to help.

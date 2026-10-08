---
title: Advanced Usage
description: Multi-project setups, provider selection, performance, security, and integration examples for Expose.
---

# Advanced Usage

This guide covers power-user features, real-world integrations, and best practices for getting the most out of Expose.

---

## Configuration Management

### Multiple projects

Each project keeps its own `.expose.yml` — just `cd` into the project and run `expose tunnel`:

```bash
cd ~/projects/frontend
expose init    # creates .expose.yml with port: 3000

cd ~/projects/api
expose init    # creates .expose.yml with port: 4000
```

### View and inspect config

```bash
expose config list        # print all settings
expose config get port    # get a single key
```

### Override at runtime

Any flag overrides the config file for that run only:

```bash
expose tunnel -p 8080            # different port
expose tunnel -P cloudflare      # different provider
```

---

## Provider Selection

| Provider | Best For | Requires |
|----------|----------|----------|
| **LocalTunnel** | Quick webhook tests, short sessions | Nothing |
| **Cloudflare** | Client demos, longer sessions | `cloudflared` binary |
| **Self-Hosted** | Full control, no rate limits | Your own VPS |

### LocalTunnel

```bash
expose tunnel               # default
expose tunnel -P localtunnel  # explicit
```

!!! warning "Stability"
    LocalTunnel's shared infrastructure can drop connections during long sessions. If you need reliability, use Cloudflare.

### Cloudflare Tunnel

=== "macOS"
    ```bash
    brew install cloudflare/cloudflare/cloudflared
    ```

=== "Linux"
    ```bash
    wget -q https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64
    sudo mv cloudflared-linux-amd64 /usr/local/bin/cloudflared
    sudo chmod +x /usr/local/bin/cloudflared
    cloudflared --version
    ```

Then start the tunnel:

```bash
expose tunnel -P cloudflare -p 3000
```

### Self-Hosted Server

Run the Expose server component on any VPS (a $5/mo instance is plenty):

```bash
# On your VPS — one-time setup
expose server \
  --domain=tunnel.mysite.com \
  --control-port=7890 \
  --public-port=8080
```

Then connect from your local machine:

```bash
expose tunnel --server=tunnel.mysite.com:7890 -p 3000
```

!!! tip "Advantages of self-hosting"
    - No rate limits
    - Traffic stays on your infrastructure
    - Custom domain support
    - No third-party ToS restrictions

---

## Performance

### Latency benchmarks

Typical tunnel overhead for a simple HTTP request:

| Provider | Added Latency |
|----------|---------------|
| Cloudflare | ~20–80 ms |
| LocalTunnel | ~50–150 ms |
| Self-Hosted (same region) | ~5–30 ms |

### Connection pooling

LocalTunnel maintains **10 concurrent connections** by default. This means up to 10 requests can be proxied in parallel without queuing.

### Reduce latency

1. **Choose the closest provider**: Cloudflare's anycast routes to the nearest PoP.
2. **Optimize your local server**: enable compression, minimize response sizes.
3. **Self-host in the same region** as your users for the lowest round-trip time.

### Memory footprint

```
Base:            ~10 MB
Per connection:  ~1 MB
Typical total:   20–30 MB
```

---

## Security

!!! danger "Expose is for development only"
    `expose tunnel` creates a **public URL** that routes directly to your `localhost`. Never expose production services, databases, or admin interfaces this way.

### What you should never expose

```bash
# ❌ DO NOT expose
expose tunnel -p 5432   # PostgreSQL
expose tunnel -p 27017  # MongoDB
expose tunnel -p 6379   # Redis
expose tunnel -p 22     # SSH
```

### Add authentication to your local app

Even in development, protect your tunnel with a simple token check:

```javascript
// Express middleware example
app.use((req, res, next) => {
    const token = req.headers['x-dev-token'];
    if (token !== process.env.DEV_TOKEN) {
        return res.status(401).json({ error: 'Unauthorized' });
    }
    next();
});
```

### Log all tunnel traffic

```javascript
app.use((req, res, next) => {
    console.log(`[${new Date().toISOString()}] ${req.method} ${req.url} — ${req.ip}`);
    next();
});
```

### Stop tunnels when done

Always press ++ctrl+c++ when you're finished. Don't leave tunnels running overnight or over the weekend.

---

## Troubleshooting

### `Connection refused`

Your local server isn't running.

```bash
# Start your server first, then run expose tunnel
npm start
```

### `Port already in use`

```bash
# Find the process
lsof -i :3000
# or
netstat -tuln | grep 3000

# Kill it or switch ports
kill -9 <PID>
expose tunnel -p 8080
```

### Tunnel disconnects frequently

LocalTunnel can be unstable. Switch to Cloudflare:

```bash
expose tunnel -P cloudflare
```

### `cloudflared: command not found`

See [Cloudflare Tunnel](#cloudflare-tunnel) installation above.

### Slow responses

1. Check your local server first: `curl -o /dev/null -s -w "%{time_total}s" http://localhost:3000`
2. Try Cloudflare: `expose tunnel -P cloudflare`
3. Self-host closer to your users.

---

## Integration Examples

### GitHub Webhooks

```bash
# 1. Start your handler
node webhook-server.js    # listens on port 4000

# 2. Start tunnel
expose tunnel -p 4000

# 3. Add the public URL in GitHub:
# Settings → Webhooks → Add webhook
# URL: https://brave-lions-jump.loca.lt/webhook
# Content type: application/json
```

### Stripe Webhooks

```bash
expose tunnel -p 3000

# Stripe Dashboard → Developers → Webhooks → Add endpoint
# URL: https://quick-birds-sing.loca.lt/stripe/webhook
```

### Slack App Development

```bash
python bot.py             # start your bot on port 5000
expose tunnel -p 5000

# Slack App config → Event Subscriptions → Request URL:
# https://smart-dogs-play.loca.lt/slack/events
```

### Docker Containers

```bash
docker run -p 3000:80 my-app
expose tunnel -p 3000
```

### Automation Script

```bash title="start-dev.sh"
#!/usr/bin/env bash
set -e
npm start &
sleep 2
expose tunnel -p 3000
```

```bash
chmod +x start-dev.sh
./start-dev.sh
```

### Multiple Tunnels at Once

Run multiple terminal sessions:

```bash
# Terminal 1
cd ~/projects/frontend && expose tunnel   # port 3000

# Terminal 2
cd ~/projects/api && expose tunnel        # port 4000
```

---

## Best Practices

| ✅ Do | ❌ Don't |
|-------|---------|
| Use for development and testing only | Expose production services |
| Stop the tunnel when done | Leave tunnels running indefinitely |
| Add auth to your local app | Expose unprotected admin UIs |
| Use Cloudflare for important demos | Use LocalTunnel for critical demos |
| Monitor your access logs | Trust unknown incoming traffic |

---

**Need more help?** [Open an issue on GitHub](https://github.com/kernelshard/expose/issues).

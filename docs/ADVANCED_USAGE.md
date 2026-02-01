# 🔧 Advanced Usage Guide

Master advanced features and techniques for using Expose effectively.

## Table of Contents

- [Configuration Management](#configuration-management)
- [Provider Selection](#provider-selection)
- [Performance Tuning](#performance-tuning)
- [Security Considerations](#security-considerations)
- [Troubleshooting](#troubleshooting)
- [Integration Examples](#integration-examples)
- [Tips & Tricks](#tips--tricks)

---

## Configuration Management

### Multiple Projects

Manage different configurations for different projects:

```bash
# Frontend project
cd ~/projects/my-frontend
expose init
# Edit .expose.yml: port: 3000

# Backend project
cd ~/projects/my-api
expose init
# Edit .expose.yml: port: 4000
```

Each project remembers its settings!

### Environment-Specific Configs

Create different configs for different environments:

```bash
# Development
cp .expose.yml .expose.dev.yml
# Staging
cp .expose.yml .expose.staging.yml

# Use specific config (future feature)
expose tunnel --config .expose.staging.yml
```

### Viewing Configuration

```bash
# List all settings
expose config list

# Get specific value
expose config get port
expose config get project
```

---

## Provider Selection

### Choosing the Right Provider

| Provider | Best For | Pros | Cons |
|----------|----------|------|------|
| LocalTunnel | Quick testing, webhooks | No signup, fast setup | Less reliable |
| Cloudflare | Demos, production | Very reliable, fast | Requires cloudflared |

### LocalTunnel

**Use when:**
- Testing webhooks quickly
- No installation preferences
- Short-lived tunnels

```bash
expose tunnel --provider localtunnel
# or just
expose tunnel
```

**Limitations:**
- Occasional connection drops
- May require reconnection
- Shared infrastructure

### Cloudflare Tunnel

**Use when:**
- Demoing to clients
- Need reliability
- Longer tunnel sessions

```bash
# Install cloudflared first
# macOS
brew install cloudflare/cloudflare/cloudflared

# Linux
wget -q https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64
sudo mv cloudflared-linux-amd64 /usr/local/bin/cloudflared
sudo chmod +x /usr/local/bin/cloudflared

# Use it
expose tunnel --provider cloudflare
```

**Benefits:**
- Cloudflare's global network
- Better performance
- More reliable connections

---

## Performance Tuning

### Connection Pooling (LocalTunnel)

LocalTunnel maintains 10 concurrent connections by default. This is optimal for most use cases.

### Reducing Latency

1. **Choose nearby provider**:
   - LocalTunnel routes through nearest server
   - Cloudflare uses global network

2. **Optimize local server**:
   ```bash
   # Use production mode
   NODE_ENV=production npm start
   
   # Enable compression
   # Check your framework docs
   ```

3. **Monitor performance**:
   ```bash
   # Check connection
   curl -w "@curl-format.txt" -o /dev/null -s https://your-tunnel.loca.lt
   ```

### Memory Usage

Expose is lightweight:
- Base memory: ~10MB
- Per connection: ~1MB
- Total typical: 20-30MB

---

## Security Considerations

### Important: Expose is for Development Only

⚠️ **Warning**: Expose creates a public URL to your localhost. Use only for development and testing.

### Best Practices

1. **Never expose sensitive data**:
   ```bash
   # ❌ Don't expose production databases
   # ❌ Don't expose admin panels
   # ❌ Don't expose internal tools
   ```

2. **Use authentication in your app**:
   ```javascript
   // Add auth to your local app
   app.use((req, res, next) => {
       const token = req.headers['x-api-token'];
       if (token !== process.env.DEV_TOKEN) {
           return res.status(401).send('Unauthorized');
       }
       next();
   });
   ```

3. **Use short-lived tunnels**:
   ```bash
   # Start tunnel
   expose tunnel
   
   # Stop when done (Ctrl+C)
   ```

4. **Monitor access**:
   ```javascript
   // Log all requests
   app.use((req, res, next) => {
       console.log(`${req.method} ${req.url} from ${req.ip}`);
       next();
   });
   ```

### Network Security

```bash
# Expose only specific port
expose tunnel --port 3000

# Don't expose:
# - Port 22 (SSH)
# - Port 5432 (PostgreSQL)
# - Port 27017 (MongoDB)
# - Any database port
```

---

## Troubleshooting

### Common Issues

#### 1. "Connection refused"

**Problem**: Local server not running

**Solution**:
```bash
# Start your local server first
npm start  # or your framework's command

# Then start tunnel
expose tunnel
```

#### 2. "Port already in use"

**Problem**: Another process using the port

**Solution**:
```bash
# Find process
lsof -i :3000
# or
netstat -tuln | grep 3000

# Kill it
kill -9 <PID>

# Or use different port
expose tunnel --port 8080
```

#### 3. Tunnel disconnects frequently

**Problem**: LocalTunnel instability

**Solution**:
```bash
# Try Cloudflare instead
expose tunnel --provider cloudflare

# Or restart tunnel
# Press Ctrl+C, then restart
expose tunnel
```

#### 4. "cloudflared not found"

**Problem**: Cloudflare provider requires cloudflared

**Solution**:
```bash
# macOS
brew install cloudflare/cloudflare/cloudflared

# Linux
wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64
sudo mv cloudflared-linux-amd64 /usr/local/bin/cloudflared
sudo chmod +x /usr/local/bin/cloudflared

# Verify
cloudflared --version
```

#### 5. Slow response times

**Causes**:
- Network latency
- Local server slow
- Heavy request load

**Solutions**:
```bash
# 1. Check local server performance
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:3000

# 2. Try different provider
expose tunnel --provider cloudflare

# 3. Optimize your local app
# Enable caching, optimize database queries, etc.
```

### Debug Mode

Currently not available, but you can:

```bash
# Monitor requests in your app
app.use((req, res, next) => {
    console.log(`[${new Date().toISOString()}] ${req.method} ${req.url}`);
    next();
});

# Watch expose output
expose tunnel 2>&1 | tee expose.log
```

---

## Integration Examples

### Webhook Testing

#### GitHub Webhooks

```bash
# 1. Start your webhook receiver
node webhook-server.js

# 2. Start tunnel
expose tunnel --port 4000

# 3. Copy the public URL
# Example: https://brave-lions-jump.loca.lt

# 4. Add to GitHub
# Settings → Webhooks → Add webhook
# URL: https://brave-lions-jump.loca.lt/webhook
# Content type: application/json
```

#### Stripe Webhooks

```bash
# 1. Start your Stripe webhook handler
npm start

# 2. Start tunnel
expose tunnel --port 3000

# 3. Add to Stripe Dashboard
# Developers → Webhooks → Add endpoint
# URL: https://quick-birds-sing.loca.lt/stripe/webhook
```

### Mobile Testing

```bash
# 1. Start your dev server
npm run dev

# 2. Start tunnel
expose tunnel

# 3. Open URL on your phone
# Example: https://happy-cats-run.loca.lt

# 4. Test responsive design
# Check different screen sizes, orientations, etc.
```

### API Development

```bash
# 1. Start your API
go run main.go

# 2. Expose it
expose tunnel --port 8080

# 3. Share with frontend team
# They can point to: https://your-tunnel.loca.lt/api
```

### Slack Bot Development

```bash
# 1. Start your Slack bot
python bot.py

# 2. Expose webhook endpoint
expose tunnel --port 5000

# 3. Configure in Slack App settings
# Event Subscriptions → Request URL
# https://smart-dogs-play.loca.lt/slack/events
```

---

## Tips & Tricks

### 1. Quick Testing

```bash
# One-liner to expose Python server
python -m http.server 8000 & expose tunnel --port 8000

# Stop with: kill %1
```

### 2. Multiple Tunnels

```bash
# Terminal 1
cd ~/frontend
expose tunnel --port 3000

# Terminal 2
cd ~/backend
expose tunnel --port 4000
```

### 3. Share with Team

```bash
# Start tunnel
expose tunnel

# Share the URL in Slack/email
# Team members can access your local work!
```

### 4. Test from Different Locations

```bash
# Start tunnel
expose tunnel

# Open URL on:
# - Your phone
# - Colleague's computer
# - Different browser
# All see your local server!
```

### 5. Automate with Scripts

```bash
# start-dev.sh
#!/bin/bash
npm start &
sleep 2
expose tunnel --port 3000
```

```bash
chmod +x start-dev.sh
./start-dev.sh
```

### 6. Use with Docker

```bash
# Start Docker container
docker run -p 3000:80 my-app

# Expose it
expose tunnel --port 3000
```

### 7. Test Payment Flows

```bash
# Start checkout page
npm start

# Expose
expose tunnel

# Use URL in payment provider testing
# Stripe, PayPal, etc. can send webhooks to it
```

### 8. Browser DevTools

```bash
# Start tunnel
expose tunnel

# Open in browser with DevTools
# Network tab shows all requests
# Console shows errors
```

---

## Comparison with Alternatives

| Feature | Expose | ngrok | localtunnel (CLI) |
|---------|--------|-------|-------------------|
| Signup | ❌ None | ✅ Required | ❌ None |
| Cost | Free | Free tier limited | Free |
| Providers | 2+ | 1 | 1 |
| Config files | ✅ Yes | ✅ Yes | ❌ No |
| Binary | Single | Single | Node.js required |
| Custom domains | ❌ No | ✅ Yes (paid) | ❌ No |

---

## Performance Benchmarks

Typical latency overhead:

```
LocalTunnel:    50-150ms
Cloudflare:     20-80ms
Direct access:  0ms (baseline)
```

Connection capacity:

```
LocalTunnel:    ~10 concurrent connections
Cloudflare:     ~100+ concurrent connections
```

---

## Best Practices Summary

✅ **Do:**
- Use for development and testing
- Stop tunnel when done
- Add authentication to your app
- Use Cloudflare for important demos
- Monitor access logs

❌ **Don't:**
- Expose production services
- Leave tunnels running indefinitely
- Expose sensitive data
- Share tunnel URLs publicly
- Use for production traffic

---

## Next Level

Want to contribute or extend Expose?

- Read the [Architecture Guide](ARCHITECTURE.md)
- Check the [Contributing Guide](../CONTRIBUTING.md)
- Join discussions on GitHub

---

**Need help?** Open an issue on [GitHub](https://github.com/kernelshard/expose/issues).

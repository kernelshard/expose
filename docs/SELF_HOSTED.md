---
title: Self-Hosted Server
description: Run your own Expose server on a VPS for full control over your tunneling infrastructure.
---

# Self-Hosted Server

`expose server` turns any Linux VPS into your own private tunnel relay — no third parties, no rate limits, no shared infrastructure.

---

## Option A: Docker Compose (Recommended)

The fastest path. Caddy handles TLS automatically.

```bash
git clone https://github.com/kernelshard/expose.git
cd expose/deploy
cp .env.example .env
# Edit .env: set DOMAIN=tunnel.yourdomain.com
docker compose up -d
```

See [`deploy/README.md`](https://github.com/kernelshard/expose/blob/main/deploy/README.md) for DNS setup and environment variables.

---

## Option B: Standalone Binary

### Install

```bash
wget https://github.com/kernelshard/expose/releases/latest/download/expose-linux-amd64
chmod +x expose-linux-amd64
sudo mv expose-linux-amd64 /usr/local/bin/expose
```

### Open firewall ports

```bash
sudo ufw allow 7890/tcp   # control port (clients)
sudo ufw allow 8080/tcp   # public port (inbound HTTP)
sudo ufw reload
```

### Start the server

```bash
expose server --domain=tunnel.mysite.com --control-port=7890 --public-port=8080
```

### Run as a service

```ini title="/etc/systemd/system/expose-server.service"
[Unit]
Description=Expose Tunnel Server
After=network.target

[Service]
Environment="EXPOSE_SERVER_TOKEN=your-secret-token"
ExecStart=/usr/local/bin/expose server --domain=tunnel.mysite.com --control-port=7890 --public-port=8080
Restart=on-failure
RestartSec=5
User=nobody

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now expose-server
```

### HTTPS with Caddy

```caddyfile title="/etc/caddy/Caddyfile"
tunnel.mysite.com {
    reverse_proxy localhost:8080
}
```

```bash
sudo systemctl reload caddy
```

Caddy provisions a Let's Encrypt certificate automatically.

---

## Connect from your machine

```bash
expose tunnel --server=tunnel.mysite.com:7890 -p 3000
```

```text
✓ Tunnel (SelfHosted) started for localhost:3000
✓ Public URL: http://swift-falcon-4821.tunnel.mysite.com:8080
✓ Provider: SelfHosted
  Press Ctrl+C to stop
```

To use a fixed subdomain instead of a random one:

```bash
expose tunnel --server=tunnel.mysite.com:7890 --subdomain=my-app -p 3000
```

Subdomains must be lowercase alphanumeric with hyphens, 1–63 characters. If a subdomain is already taken, the server rejects the request.

---

## Security & Authentication

By default, authentication is **completely optional** — the server works out of the box without any setup. 

However, if your VPS has a public IP, **configuring an auth token is strongly recommended** to prevent port scanners and strangers from using your server as a free tunnel relay.

### Option 1: Shared Auth Token (Recommended for Public VPS)

When configured, the server requires connecting clients to supply a matching token during handshake. If no token is configured, the server remains open for anyone you share the address with. Token comparison is timing-attack safe (`crypto/subtle.ConstantTimeCompare`).

#### 1. On your VPS (Server):
Set `--token` or the `EXPOSE_SERVER_TOKEN` environment variable. Using the environment variable is recommended for production (systemd or Docker) to avoid exposing credentials in `ps aux`:

```bash
# Via flag
expose server --domain=tunnel.mysite.com --token=your-secret-token

# Or via environment variable
export EXPOSE_SERVER_TOKEN=your-secret-token
expose server --domain=tunnel.mysite.com
```

#### 2. On your Machine (Client):
Pass the `--token` flag when opening your tunnel:

```bash
expose tunnel 3000 --server=tunnel.mysite.com:7890 --token=your-secret-token
```

Or export it in your shell environment (`~/.zshrc` or `~/.bashrc`):
```bash
export EXPOSE_SERVER_TOKEN=your-secret-token
expose tunnel 3000 --server=tunnel.mysite.com:7890
```

> **Note**: If a token is configured on the server and a client attempts to connect without a token or with an invalid token, the connection is immediately rejected:
> ```text
> server error: unauthorized: invalid or missing auth token
> ```

### Option 2: IP Whitelisting (Firewall)

If your development machine has a static IP address, you can also restrict access to the control port (`7890`) using UFW:

```bash
sudo ufw allow from <YOUR_IP> to any port 7890
sudo ufw deny 7890
```

---

## Troubleshooting

**Client can't connect**
```bash
ss -tlnp | grep 7890   # verify server is listening
sudo ufw status        # check firewall
```

**Public port unreachable**
```bash
curl -I http://tunnel.mysite.com:8080
systemctl status expose-server
```

**Tunnel keeps dropping**  
Check server logs: `sudo journalctl -u expose-server -f`

---

**Need help?** [Open an issue on GitHub](https://github.com/kernelshard/expose/issues).

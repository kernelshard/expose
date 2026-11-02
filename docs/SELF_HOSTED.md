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

## Security

Lock the control port to your IP so only you can register tunnels:

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

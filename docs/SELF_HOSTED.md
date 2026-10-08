---
title: Self-Hosted Server
description: Run your own Expose server on a VPS for full control over your tunneling infrastructure — no rate limits, no third parties.
---

# Self-Hosted Server

Expose ships a built-in server mode (`expose server`) that turns any VPS into your own private tunneling relay. Traffic flows entirely through your infrastructure — no third-party routing, no rate limits, no terms-of-service surprises.

!!! tip "When to self-host"
    Self-hosting makes sense when you need persistent tunnels, custom domains, or want to avoid shared infrastructure. For quick one-off webhook tests, [LocalTunnel](GETTING_STARTED.md#localtunnel-default) is simpler.

---

## How It Works

```
Your machine                 Your VPS                    Internet
─────────────               ─────────────               ─────────────
expose tunnel  ──TCP──►  expose server  ◄──HTTP──  Browser / Service
localhost:3000            control-port: 7890
                          public-port:  8080
                          domain: tunnel.mysite.com
```

1. `expose server` listens on two ports:
    - **Control port** (`7890` by default) — your tunnel clients connect here
    - **Public port** (`8080` by default) — inbound HTTP traffic from the internet
2. `expose tunnel --server` connects your local machine to the control port
3. Requests arriving at `http://tunnel.mysite.com:8080` are forwarded through the tunnel to `localhost:3000`

---

## Server Setup (VPS)

### Prerequisites

- Any Linux VPS ($5/mo on DigitalOcean, Hetzner, Linode, etc.)
- Go 1.21+ **or** a pre-built binary
- Firewall open on your chosen control port and public port

### Install on the VPS

=== "Go Install"
    ```bash
    go install github.com/kernelshard/expose/cmd/expose@latest
    ```

=== "Pre-built Binary"
    ```bash
    # Download the linux-amd64 binary from the releases page
    wget https://github.com/kernelshard/expose/releases/latest/download/expose-linux-amd64
    chmod +x expose-linux-amd64
    sudo mv expose-linux-amd64 /usr/local/bin/expose
    expose --version
    ```

### Open firewall ports

```bash
# UFW (Ubuntu / Debian)
sudo ufw allow 7890/tcp   # control port
sudo ufw allow 8080/tcp   # public port
sudo ufw reload
```

### Start the server

```bash
expose server \
  --domain=tunnel.mysite.com \
  --control-port=7890 \
  --public-port=8080
```

Output:

```text
🚀 Expose server starting...
✓ Control port: 7890  (clients connect here)
✓ Public port:  8080  (inbound HTTP traffic)
✓ Domain:       tunnel.mysite.com
```

### Server flags reference

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--domain` | `-d` | `localhost` | Domain used in public URLs |
| `--control-port` | `-c` | `7890` | Port tunnel clients connect to |
| `--public-port` | `-p` | `8080` | Port that receives inbound HTTP traffic |

---

## Client Connection

On your local machine, connect to your VPS with the `--server` flag:

```bash
expose tunnel --server=tunnel.mysite.com:7890 -p 3000
```

Short form using `-s`:

```bash
expose tunnel -s tunnel.mysite.com:7890 -p 3000
```

Output:

```text
✓ Tunnel (SelfHosted) started for localhost:3000
✓ Public URL: http://tunnel.mysite.com:8080
✓ Forwarding to: http://localhost:3000
✓ Provider: SelfHosted
  Press Ctrl+C to stop
```

!!! note
    When `--server` is provided, Expose automatically selects the `selfhosted` provider — you don't need to pass `-P selfhosted` separately.

---

## Run as a Background Service (systemd)

To keep the server running after you disconnect from SSH, create a systemd service:

```bash
sudo nano /etc/systemd/system/expose-server.service
```

```ini title="/etc/systemd/system/expose-server.service"
[Unit]
Description=Expose Self-Hosted Tunnel Server
After=network.target

[Service]
ExecStart=/usr/local/bin/expose server \
  --domain=tunnel.mysite.com \
  --control-port=7890 \
  --public-port=8080
Restart=on-failure
RestartSec=5
User=nobody

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable expose-server
sudo systemctl start expose-server

# Check status
sudo systemctl status expose-server

# Watch logs
sudo journalctl -u expose-server -f
```

---

## Putting It Behind a Domain

If you want clean URLs like `https://tunnel.mysite.com` instead of `http://tunnel.mysite.com:8080`, put Nginx or Caddy in front.

=== "Caddy (easiest — auto HTTPS)"
    ```caddyfile title="/etc/caddy/Caddyfile"
    tunnel.mysite.com {
        reverse_proxy localhost:8080
    }
    ```

    ```bash
    sudo systemctl reload caddy
    ```
    Caddy automatically provisions a Let's Encrypt TLS certificate.

=== "Nginx"
    ```nginx title="/etc/nginx/sites-available/expose"
    server {
        listen 80;
        server_name tunnel.mysite.com;

        location / {
            proxy_pass http://localhost:8080;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_cache_bypass $http_upgrade;
        }
    }
    ```

    ```bash
    sudo ln -s /etc/nginx/sites-available/expose /etc/nginx/sites-enabled/
    sudo nginx -t && sudo systemctl reload nginx

    # Add TLS with Certbot
    sudo certbot --nginx -d tunnel.mysite.com
    ```

---

## Security Recommendations

!!! warning "Restrict access to the control port"
    The control port (`7890`) should only be reachable by trusted machines. Consider blocking it from the public internet and using SSH tunneling to connect your client.

```bash
# Only allow your home/office IP on the control port
sudo ufw allow from <YOUR_IP> to any port 7890
sudo ufw deny 7890
```

Alternatively, connect through SSH port forwarding:

```bash
# Forward local port 7890 to VPS port 7890 over SSH
ssh -L 7890:localhost:7890 user@your-vps -N &

# Then use localhost as the server address
expose tunnel --server=localhost:7890 -p 3000
```

---

## Troubleshooting

### Client can't connect to control port

```bash
# Verify the server is listening
ss -tlnp | grep 7890

# Check firewall
sudo ufw status
```

### Public port not reachable

```bash
# Test from outside the VPS
curl -I http://tunnel.mysite.com:8080

# Verify the server process is up
systemctl status expose-server
```

### Connection drops after idle

Add a keepalive to your SSH session (if using SSH tunnel forwarding):

```bash
# ~/.ssh/config
Host your-vps
    ServerAliveInterval 30
    ServerAliveCountMax 3
```

---

**Need help?** [Open an issue on GitHub](https://github.com/kernelshard/expose/issues).

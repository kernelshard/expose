# Self-Hosting Expose with Docker Compose

Deploy your own private tunnel relay on any VPS (DigitalOcean, Hetzner, AWS Lightsail, etc.) in under 60 seconds.

```text
[ Developer Laptop ] ─── TCP :7890 ───────────────────────► [ expose-server :7890 ]
                                                                     ▲
[ Public Visitor   ] ─── HTTPS :443 ──► [ Caddy ] ── :8080 ──────────┘
```

---

## 1. Prerequisites & DNS Setup

To route dynamic subdomains (`cool-tiger-42.tunnel.yourdomain.com`) to your server, add two DNS **A records** pointing to your VPS IP:

| Type | Name / Host | Value | Purpose |
|---|---|---|---|
| **A** | `tunnel` | `<YOUR_VPS_IP>` | Base server address |
| **A** | `*.tunnel` | `<YOUR_VPS_IP>` | Wildcard for all dynamic client tunnels |

> **Tip (Free Wildcard HTTPS)**: If your domain uses Cloudflare, set the proxy status to **Proxied (Orange Cloud)** for both records. Cloudflare will automatically terminate wildcard SSL at their edge with zero API keys or rate limits.

---

## 2. Quickstart

### Step 1: Clone the Repository
```bash
git clone https://github.com/kernelshard/expose.git
cd expose/deploy
```

### Step 2: Configure Environment
Copy the example environment file and set your domain:
```bash
cp .env.example .env
nano .env
```

Set `DOMAIN` to match your relay address:
```env
DOMAIN=tunnel.yourdomain.com
CONTROL_PORT=7890
```

### Step 3: Start the Stack
```bash
docker compose up -d
```

Check server logs:
```bash
docker compose logs -f expose-server
```

---

## 3. Connecting From Your Laptop

Now start a tunnel pointing to your self-hosted server:

```bash
expose tunnel 3000 --server=tunnel.yourdomain.com:7890
```

You'll get an instant, secure public URL:
```text
🚀 Tunnel started!
✓ Public URL: https://cool-tiger-42.tunnel.yourdomain.com
✓ Forwarding to: http://localhost:3000
```

---

## 4. Standalone Deployment (Existing Reverse Proxy)

If you already have Nginx, Traefik, or Caddy running on your server, you can run `expose-server` standalone:

```bash
docker compose up -d expose-server
```

Then point your existing reverse proxy to `localhost:8080` and forward the `Host` header.

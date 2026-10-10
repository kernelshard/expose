---
title: Expose — Go-Based Reverse Proxy & Tunneling CLI (Ngrok Alternative)
description: The open-source, single-binary Go (Golang) alternative to ngrok. Expose local ports via LocalTunnel, Cloudflare, or a self-hosted server — zero signup, no rate limits.
keywords:
  - expose
  - ngrok alternative
  - golang ngrok alternative
  - go tunnel
  - reverse proxy
  - tunneling
  - localtunnel
  - cloudflare tunnel
  - self-hosted tunnel
  - webhook testing
  - golang cli
hide:
  - toc
---

<div class="expose-hero" markdown>
<span class="expose-hero__badge">🚀 Open Source · Zero Signup · Single Binary</span>

# Expose Local → Internet

Share your localhost with the world in seconds. No account. No auth tokens. No nonsense.

<div class="expose-hero__actions" markdown>
[Get Started](GETTING_STARTED.md){ .expose-btn .expose-btn--primary }
[View on GitHub](https://github.com/kernelshard/expose){ .expose-btn .expose-btn--secondary }
</div>
</div>

<div class="expose-badges" markdown>
[![Downloads](https://img.shields.io/github/downloads/kernelshard/expose/total?color=5C6BC0&logo=github&labelColor=1a1f35&style=flat-square)](https://github.com/kernelshard/expose/releases)
[![Tests](https://img.shields.io/github/actions/workflow/status/kernelshard/expose/test.yml?label=tests&style=flat-square&labelColor=1a1f35&color=26C6DA)](https://github.com/kernelshard/expose/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kernelshard/expose)](https://goreportcard.com/report/github.com/kernelshard/expose)
[![License: MIT](https://img.shields.io/badge/License-MIT-5C6BC0.svg?style=flat-square&labelColor=1a1f35)](https://github.com/kernelshard/expose/blob/main/LICENSE)
</div>

<div class="expose-stats" markdown>
<div class="expose-stat">
  <span class="expose-stat__value">&lt;10MB</span>
  <span class="expose-stat__label">Binary Size</span>
</div>
<div class="expose-stat">
  <span class="expose-stat__value">3</span>
  <span class="expose-stat__label">Providers</span>
</div>
<div class="expose-stat">
  <span class="expose-stat__value">&gt;80%</span>
  <span class="expose-stat__label">Test Coverage</span>
</div>
<div class="expose-stat">
  <span class="expose-stat__value">0</span>
  <span class="expose-stat__label">Signups Needed</span>
</div>
</div>

---

## Quick Start

=== "Shell Script (macOS & Linux)"
    ```bash
    curl -fsSL https://raw.githubusercontent.com/kernelshard/expose/main/install.sh | sh
    expose init
    expose tunnel
    ```

=== "Go Install"
    ```bash
    go install github.com/kernelshard/expose/cmd/expose@latest
    expose init
    expose tunnel
    ```

=== "Download Binary"
    Download for your platform from the [Releases page](https://github.com/kernelshard/expose/releases), then:
    ```bash
    chmod +x expose
    expose init
    expose tunnel
    ```

=== "Build from Source"
    ```bash
    git clone https://github.com/kernelshard/expose.git
    cd expose
    go build -o expose ./cmd/expose
    ./expose init && ./expose tunnel
    ```

Output when running `expose tunnel`:

```text
✓ Tunnel (LocalTunnel) started for localhost:3000
✓ Public URL: https://quick-mammals-sing.loca.lt
✓ Forwarding to: http://localhost:3000
  Press Ctrl+C to stop
```

!!! tip "Override port on the fly"
    ```bash
    expose tunnel -p 8080           # LocalTunnel on port 8080
    expose tunnel -P cloudflare     # Use Cloudflare instead
    expose tunnel --server host:7890 -p 3000  # Self-hosted
    ```

---

## Why Expose?

<div class="expose-features" markdown>

<div class="expose-feature-card" markdown>
<span class="expose-feature-card__icon">🌐</span>
<div class="expose-feature-card__title">3 Providers, 1 Tool</div>
<p class="expose-feature-card__desc">LocalTunnel for quick tests, Cloudflare for reliability, or your own server for full control.</p>
</div>

<div class="expose-feature-card" markdown>
<span class="expose-feature-card__icon">🏠</span>
<div class="expose-feature-card__title">Self-Hosted Mode</div>
<p class="expose-feature-card__desc">Run <code>expose server</code> on any $5 VPS. Your traffic, your rules — no third-party inspection.</p>
</div>

<div class="expose-feature-card" markdown>
<span class="expose-feature-card__icon">⚡</span>
<div class="expose-feature-card__title">Zero Friction</div>
<p class="expose-feature-card__desc">No accounts, no API keys, no dashboards. Just install and run — three commands from zero to tunnel.</p>
</div>

<div class="expose-feature-card" markdown>
<span class="expose-feature-card__icon">📦</span>
<div class="expose-feature-card__title">Single Binary</div>
<p class="expose-feature-card__desc">One file under 10MB, statically compiled. Copy it anywhere, it just works. No runtime needed.</p>
</div>

<div class="expose-feature-card" markdown>
<span class="expose-feature-card__icon">🔧</span>
<div class="expose-feature-card__title">Developer First</div>
<p class="expose-feature-card__desc">Project-based config files, short flags, clean output, and >80% test coverage. Built by devs for devs.</p>
</div>

<div class="expose-feature-card" markdown>
<span class="expose-feature-card__icon">🔓</span>
<div class="expose-feature-card__title">Fully Open Source</div>
<p class="expose-feature-card__desc">MIT licensed. Audit the code, extend it, run it yourself. No proprietary lock-in.</p>
</div>

</div>

---

## Expose vs. Alternatives

| Feature | **Expose** | ngrok | Cloudflare Tunnel | LocalTunnel (Node) |
| :--- | :---: | :---: | :---: | :---: |
| No Signup | ✅ | ❌ | ❌ | ✅ |
| Single Binary | ✅ | ✅ | ✅ | ❌ Node.js |
| Open Source | ✅ | ❌ | ✅ | ✅ |
| Self-Hosted Server | ✅ | ❌ | ❌ | ❌ |
| Multiple Providers | ✅ | ❌ | ❌ | ❌ |
| Free | ✅ | Free tier | Free | ✅ |
| Custom Domains | 🚧 Planned | 💲 Paid | ✅ | ❌ |

---

## Common Use Cases

=== "Webhook Testing"
    ```bash
    # Start your webhook handler on port 4000
    node webhook-server.js

    # Expose it
    expose tunnel -p 4000

    # Paste the public URL into GitHub/Stripe/Twilio settings
    ```

=== "Mobile Testing"
    ```bash
    # Start your dev server
    npm run dev

    # Expose it
    expose tunnel

    # Open the public URL on your phone or tablet
    ```

=== "Client Demo"
    ```bash
    # Start your app
    npm start

    # Use Cloudflare for better reliability during demos
    expose tunnel -P cloudflare -p 3000
    ```

=== "Self-Hosted"
    ```bash
    # 1. One-time setup on your VPS
    expose server --domain=tunnel.mysite.com \
                  --control-port=7890 \
                  --public-port=8080

    # 2. Connect from anywhere
    expose tunnel --server=tunnel.mysite.com:7890 -p 3000
    ```

!!! note "Self-hosted advantage"
    With self-hosted mode, traffic stays entirely on your infrastructure — no rate limits, no third-party routing, no terms-of-service surprises.

---

## Explore the Docs

- 📖 **[Getting Started](GETTING_STARTED.md)** — Installation, first tunnel, and troubleshooting
- 🏠 **[Self-Hosted Server](SELF_HOSTED.md)** — Run your own server on a VPS
- 🔬 **[How Tunnels Work](TUTORIAL.md)** — Deep dive into the protocol and implementation
- 🛠️ **[Advanced Usage](ADVANCED_USAGE.md)** — Multi-project setups, security, and integrations
- 🏗️ **[Architecture](ARCHITECTURE.md)** — Concurrent Go architecture and design patterns
- 👋 **[Onboarding](ONBOARDING.md)** — Set up a dev environment and contribute

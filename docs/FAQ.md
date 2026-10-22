---
title: Frequently Asked Questions (FAQ) — Expose
description: Comprehensive FAQ for Expose. Compare Expose vs ngrok, learn about zero-signup tunnels, self-hosting relays on a VPS, webhook testing, and pure-Go architecture.
keywords:
  - expose faq
  - ngrok alternative faq
  - go tunnel questions
  - self hosted tunnel faq
  - webhook testing localhost
  - expose vs ngrok
  - localtunnel alternative
  - zero signup tunnel
---

# Frequently Asked Questions (FAQ)

Everything you need to know about **Expose** — how it works, how it compares to ngrok, self-hosting on your own VPS, and architecture details.

---

## 🚀 General & Comparison

### How does Expose compare to ngrok?

Expose was built from first principles as a lightweight, open-source alternative to ngrok. The primary differences are:

| Feature | 🚀 **Expose** | ngrok |
| :--- | :--- | :--- |
| **Account Required** | **Zero signups** — run immediately | Mandatory account & auth token |
| **Pricing** | **100% Free & Open Source** (Apache 2.0) | Proprietary with paid tiers |
| **Self-Hosting** | **Built-in relay** (`expose server`) | SaaS only (relays run by ngrok) |
| **Binary Footprint** | **&lt;10MB** (single static Go binary) | ~30MB proprietary executable |
| **Runtime Dependencies** | **None** (zero CGO, zero runtime) | Closed-source daemon |
| **Custom Domains** | Free with self-hosted server | Requires paid subscription |

!!! tip "When to use Expose"
    Expose is ideal when you need an instant public URL without creating an account, want complete privacy by routing traffic through your own VPS, or need a clean, scriptable CLI for CI/CD pipelines.

---

### Does Expose require an account or credit card?

**No.** Expose requires zero signup, no email registration, and no API authentication tokens. You can install the binary and run `expose tunnel 3000` immediately out of the box.

---

### Is Expose related to the PHP Expose tool (`beyondcode/expose`)?

**No.** Expose by `@kernelshard` is an independent, pure-Go implementation. 

* **No PHP or Composer**: Unlike `beyondcode/expose` (which requires a PHP 8+ runtime and global Composer dependencies), Expose is a single statically compiled Go binary with **zero external dependencies**.
* **Zero CGO**: Compiles to native machine code on macOS, Linux, and Windows.
* **Architecture**: Built from scratch using Go's standard networking library (`net`, `net/http`), native connection pooling, and atomic subdomain routing.

---

### Is Expose completely open source?

**Yes.** Expose is distributed under the permissive **Apache License 2.0**. The client, CLI, and self-hosted relay server are 100% open source, free to use, audit, and modify for commercial and personal projects.

---

## 🏠 Self-Hosting & Privacy

### Can I run my own private tunnel relay server?

**Yes.** Expose ships with a built-in server mode in the exact same binary:

```bash
expose server --domain=tunnel.mysite.com --control-port=7890 --public-port=8080
```

Any client can then connect to your private gateway:

```bash
expose tunnel 3000 --server=tunnel.mysite.com:7890 --subdomain=my-app
```

Traffic flows strictly between your local computer and your VPS — no third-party relays, no telemetry, and no commercial rate limits. For step-by-step VPS instructions, see the [Self-Hosted Server Guide](SELF_HOSTED.md).

---

### Can I request a custom subdomain?

**Yes.** When connecting to a self-hosted server, use the `--subdomain` flag:

```bash
expose tunnel 3000 --server=tunnel.mysite.com:7890 --subdomain=my-feature
```

Your tunnel will be accessible at `http://my-feature.tunnel.mysite.com:8080`.

!!! note "RFC 1123 Validation & Collision Protection"
    Subdomains must comply with RFC 1123 DNS standards (1–63 lowercase alphanumeric characters and hyphens). If a requested subdomain is already in use by an active tunnel, the server safely rejects the request to prevent traffic hijacking.

---

### What tunnel providers does Expose support?

Expose provides an interface-driven architecture supporting three interchangeable backends:

1. **LocalTunnel** (`-P localtunnel`, default): Zero-configuration public relay ideal for quick tests.
2. **Cloudflare Quick Tunnels** (`-P cloudflare`): Routes through Cloudflare's global edge network via `cloudflared`.
3. **Self-Hosted** (`-P selfhosted` or `--server`): Connects to your own private `expose server`.

---

## 🔒 Networking & Security

### Do I need to configure router port forwarding or open inbound ports?

**No.** Expose uses **reverse tunneling**. Your local machine initiates an **outbound** TCP connection to the relay server. Because firewalls and NAT routers permit outbound traffic and its stateful responses, your local app becomes reachable over the public internet without port forwarding or modifying router rules.

```text
Localhost:3000 <──HTTP── Expose Client ──TCP (Outbound)──► Relay Server ◄──HTTP── Internet Visitor
```

---

### Does Expose work behind CGNAT (Carrier-Grade NAT) and corporate firewalls?

**Yes.** Because the tunnel connection is established outbound on standard TCP ports, Expose functions seamlessly across cellular data networks (4G/5G mobile tethering), hotel Wi-Fi, CGNAT ISPs, and restrictive corporate firewalls.

---

### Is Expose suitable for webhook testing (Stripe, Twilio, GitHub, PayPal)?

**Yes.** Webhook testing is one of the primary use cases for Expose:

1. Start your local webhook handler (e.g. on port `8080`).
2. Run `expose tunnel 8080`.
3. Copy the generated public HTTPS URL (e.g. `https://cool-app.loca.lt/api/webhook`).
4. Paste the URL into your Stripe, Twilio, or GitHub developer dashboard.

All external webhook events will be delivered in real-time to your local code for instant debugging and breakpoints.

---

### What happens if my local app isn't running yet when I start the tunnel?

Expose features **deferred local dialing**:
* The tunnel establishes the public endpoint immediately so you can configure your webhook URL.
* If your local development server is still booting, Expose displays:
  ```text
  Forwarding to: http://localhost:3000 (Waiting for local app to start...)
  ```
* Once your local app starts listening, Expose automatically detects it and begins routing traffic without requiring a tunnel restart.

---

## 📦 Installation & Troubleshooting

### How do I install Expose?

=== "Homebrew (macOS & Linux)"
    ```bash
    brew install kernelshard/tap/expose
    ```

=== "One-Line Shell Installer"
    ```bash
    curl -fsSL https://raw.githubusercontent.com/kernelshard/expose/main/install.sh | sh
    ```

=== "Go Install (Go 1.22+)"
    ```bash
    go install github.com/kernelshard/expose/cmd/expose@latest
    ```

=== "Pre-Built Binary"
    Download the latest executable for your OS and architecture from the [GitHub Releases](https://github.com/kernelshard/expose/releases).

---

### Why does my browser show a "Friendly Reminder" or warning page when using LocalTunnel?

The default public LocalTunnel relay displays a one-time anti-phishing interstitial page for web visitors. 
To bypass this completely:
* Send a `Bypass-Tunnel-Reminder: true` header in API/webhook requests.
* Use Cloudflare Quick Tunnels: `expose tunnel -P cloudflare 3000`.
* Or run your own private server: `expose server --domain=...`.

---

### Can I save default settings per project?

**Yes.** Run `expose init` in your project root to generate a `.expose.yml` file:

```yaml
project: my-app
port: 3000
```

Once created, you can simply run `expose tunnel` without specifying a port number.

---

### Where can I report bugs or request features?

Open an issue or join discussions on our [GitHub Repository](https://github.com/kernelshard/expose/issues). Contributions, bug reports, and feedback are always welcome!

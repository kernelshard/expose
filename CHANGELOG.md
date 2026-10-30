# Changelog

All notable changes to Expose will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

## [0.6.1] - 2026-10-31

### Fixed
- **LocalTunnel Relay Handshake Timeout & Fallback Retry**: Bumped LocalTunnel provider API timeout from 10s to 12s and added an immediate fallback retry attempt to withstand upstream traffic congestion on localtunnel.me ([#104](https://github.com/kernelshard/expose/issues/104), [#105](https://github.com/kernelshard/expose/pull/105)).

## [0.6.0] - 2026-10-30

### Added
- **Turnkey Docker Self-Hosting Stack**: Official multi-stage `Dockerfile` and `deploy/docker-compose.yml` bundled with Caddy reverse proxy, enabling one-command self-hosting (`docker compose up -d`) on any VPS or homelab with zero external dependencies ([#66](https://github.com/kernelshard/expose/issues/66), [#99](https://github.com/kernelshard/expose/pull/99)).
- **Automated Wildcard HTTPS via Caddy**: Bundled Caddy container configured with `tls internal` and upstream `Host` preservation to terminate HTTPS on ports 80/443 and route traffic to the internal Expose server port 8080, delivering free edge wildcard SSL with zero DNS API keys or Let's Encrypt rate limits ([#66](https://github.com/kernelshard/expose/issues/66), [#99](https://github.com/kernelshard/expose/pull/99)).
- **Minimal Non-Root Container Image**: Lightweight Alpine 3.21 runtime image (~16.7 MB) running as unprivileged `USER expose` with statically compiled binary (`CGO_ENABLED=0`, `-s -w`) ([#66](https://github.com/kernelshard/expose/issues/66), [#99](https://github.com/kernelshard/expose/pull/99)).
- **Animated Terminal Demo GIF**: Embedded high-framerate VHS recording into the README hero section showcasing instant tunnel startup, readable subdomain assignment, and live HTTP webhook forwarding ([#87](https://github.com/kernelshard/expose/issues/87), [#97](https://github.com/kernelshard/expose/pull/97), [#98](https://github.com/kernelshard/expose/pull/98)).

### Removed
- **Legacy Tunnel Manager Prototype**: Removed deprecated `internal/tunnel/manager.go` and its unit tests (-484 lines) dating back to v0.1.0 in favor of the clean Provider + Service pattern ([#100](https://github.com/kernelshard/expose/issues/100), [#101](https://github.com/kernelshard/expose/pull/101)).

### Documentation
- **Quickstart Usability Split**: Reorganized README quickstart into distinct Public Tunnel (zero setup) vs Self-Hosted Relay paths for clearer onboarding ([#102](https://github.com/kernelshard/expose/pull/102)).
- **Turnkey Docker Self-Hosting Documentation**: Updated `README.md`, `docs/index.md`, `docs/SELF_HOSTED.md`, `llms.txt`, and added `deploy/README.md` featuring Docker Compose as the recommended VPS installation path with DNS A-record topology and environment configuration templates ([#66](https://github.com/kernelshard/expose/issues/66), [#99](https://github.com/kernelshard/expose/pull/99)).

## [0.5.4] - 2026-10-25

### Added
- **Memorable Adjective-Noun Subdomains**: Replaced 8-character random hex subdomains (`a1b2c3d4`) on the self-hosted relay with human-friendly, readable names (e.g., `swift-falcon-4821`). Built using a curated dictionary of 67 adjectives and 80 nature/animal nouns combined with a 4-digit cryptographic random number via `crypto/rand`, yielding over 51 million unique combinations while strictly enforcing RFC 1123 DNS compliance ([#62](https://github.com/kernelshard/expose/issues/62), [#94](https://github.com/kernelshard/expose/pull/94)).

### Documentation
- **Readable Subdomain Documentation**: Updated the self-hosted deployment guide, FAQ, README, and reference manuals (`llms.txt`, `llms-full.txt`) documenting default memorable subdomain generation and custom subdomain rules ([#95](https://github.com/kernelshard/expose/pull/95)).

## [0.5.3] - 2026-10-22

### Performance
- **Zero-Allocation Stream Buffer Recycling**: Implemented `internal/pool` with `sync.Pool` to recycle 32KB copy buffers during bidirectional proxy streaming on both the relay server and self-hosted client. Stored buffer pointers (`*[]byte`) instead of slice values to eliminate Go interface boxing heap allocations, dropping heap allocations by 95.8% (34,209 B/op $\rightarrow$ 1,444 B/op) and boosting throughput by 26% ([#68](https://github.com/kernelshard/expose/issues/68), [#84](https://github.com/kernelshard/expose/pull/84)).

### Fixed
- **Symmetric Socket Teardown & Descriptor Leak Prevention**: Coordinated bidirectional proxy copy goroutines in `ServeHTTP` using `sync.WaitGroup`, ensuring that when either the public visitor or local tunnel client disconnects (EOF), the opposite socket is immediately closed. This prevents orphaned half-open sockets, blocked goroutines, and file descriptor exhaustion under continuous client churn ([#58](https://github.com/kernelshard/expose/issues/58), [#85](https://github.com/kernelshard/expose/pull/85)).
- **Connection Variable Naming Clarity**: Renamed ambiguous `clientConn` to `visitorConn` in the relay server proxy pipeline to clearly delineate internet visitor traffic from internal client tunnel sockets.

## [0.5.2] - 2026-10-18

### Security
- **Control Plane Handshake Read Deadline & Payload Limits**: Added a 5-second handshake deadline on port 7890 to mitigate Slowloris denial-of-service, capped handshake payloads at 4KB using `io.LimitReader` to prevent unbounded memory allocation (OOM), and immediately reset the deadline upon successful handshake so pooled data sockets remain healthy ([#56](https://github.com/kernelshard/expose/issues/56), [#81](https://github.com/kernelshard/expose/pull/81)).

### Fixed
- **RFC 1123 Hostname Validation & Namespace Protection**: Enforced strict RFC 1123 rules for custom subdomains (1–63 lowercase alphanumeric characters and hyphens, no dots or spaces), blocked reserved infrastructure keywords (`admin`, `api`, `localhost`, `www`, `root`, `internal`, `dashboard`), and added upfront `400 Bad Request` validation for public visitors ([#57](https://github.com/kernelshard/expose/issues/57), [#82](https://github.com/kernelshard/expose/pull/82)).

### Documentation
- **Feature Matrix & Subdomain Guidelines**: Updated comparison tables in README and docs to reflect partial custom subdomain support on the self-hosted relay, and documented RFC 1123 subdomain format rules and reserved namespaces in the self-hosted guide.

## [0.5.1] - 2026-10-15

### Added
- **Heartbeat Ping-Pong & Zombie Connection Eviction**: Periodic 20s client `PING` and 60s server deadline to actively prune dead tunnel connections upon silent network drops ([#73](https://github.com/kernelshard/expose/issues/73), [#79](https://github.com/kernelshard/expose/pull/79)).
- **Custom Subdomain Flag**: Added `--subdomain` / `-d` CLI flag support for self-hosted relay registration with collision detection ([#75](https://github.com/kernelshard/expose/issues/75), [#76](https://github.com/kernelshard/expose/pull/76)).

### Fixed
- **Keep-Alive Control Connection**: Kept the control connection socket open indefinitely until client termination, preventing premature tunnel teardown ([#72](https://github.com/kernelshard/expose/issues/72), [#74](https://github.com/kernelshard/expose/pull/74)).
- **Deferred Local Connection Dialing**: Delayed dialing `localhost` until incoming visitor request bytes actually arrive over the tunnel, avoiding connection spam and resource leaks ([#59](https://github.com/kernelshard/expose/issues/59), [#78](https://github.com/kernelshard/expose/pull/78)).

### Documentation & UX
- **Quickstart Usability**: Streamlined README with dedicated copy-paste blocks for Homebrew installation and direct command execution.

---

## [0.5.0] - 2026-10-12

### Added
- **Multi-Tenant Subdomain Routing**: Isolated tunnel connection pools per subdomain, ensuring multiple concurrent client tunnels can register and route traffic independently without shared FIFO connection collision ([#33](https://github.com/kernelshard/expose/issues/33), [#51](https://github.com/kernelshard/expose/pull/51)).
- **Homebrew Tap Support**: Automated formula publication to [`kernelshard/homebrew-tap`](https://github.com/kernelshard/homebrew-tap) via GoReleaser, enabling `brew install kernelshard/tap/expose` ([#60](https://github.com/kernelshard/expose/issues/60), [#61](https://github.com/kernelshard/expose/pull/61)).
- **LLM & Search Engine Optimization**: Added `/llms.txt`, `/llms-full.txt`, `robots.txt`, and Open Graph / JSON-LD schema metadata for improved AI discoverability and search indexing ([#48](https://github.com/kernelshard/expose/issues/48), [#49](https://github.com/kernelshard/expose/pull/49)).

### Changed
- **Relicensed to Apache License 2.0**: Relicensed the project from MIT to the Apache License 2.0 to provide explicit patent grant protection (Section 3), trademark reservations (Section 6), and greater legal certainty for enterprise adoption ([#52](https://github.com/kernelshard/expose/issues/52), [#53](https://github.com/kernelshard/expose/pull/53)).

---

## [0.4.2] - 2026-10-10

### Fixed
- **LocalTunnel Connection Replenishment & Proxy Hang**: Workers now run in a loop to replenish connections upon completion and immediately close the tunnel socket on EOF, resolving 30s hangs and subsequent 502 Bad Gateway errors ([#45](https://github.com/kernelshard/expose/issues/45), [#46](https://github.com/kernelshard/expose/pull/46)).

### Added
- **CLI Ergonomics**: Made `.expose.yml` optional (defaults to port 3000) and added support for positional port arguments like `expose tunnel 8080` ([#44](https://github.com/kernelshard/expose/pull/44)).
- **One-Line Shell Installer**: Automated POSIX installation script (`install.sh`) for macOS and Linux supporting architecture auto-detection (`x86_64`, `arm64`), SHA-256 checksum verification, custom directory configuration (`BINDIR`), version pinning (`VERSION`), and graceful permissions handling ([#39](https://github.com/kernelshard/expose/issues/39), [#42](https://github.com/kernelshard/expose/pull/42)).
- **Documentation Site**: Dedicated documentation website powered by Material for MkDocs hosted at `https://kernelshard.github.io/expose/` ([#40](https://github.com/kernelshard/expose/issues/40), [#41](https://github.com/kernelshard/expose/pull/41)).
- **Self-Hosted Deployment Guide**: Comprehensive end-to-end documentation covering architecture, server CLI flags, systemd service configuration, reverse proxy setup (Caddy and Nginx), and firewall security.
- **Theme & Design**: Custom indigo/cyan palette, modern hero section, feature grid, dark/light mode toggle, and responsive typography using Inter and JetBrains Mono.
- **Automated Docs CI/CD**: GitHub Actions workflow deploying the documentation site to GitHub Pages on pushes to `main` and releases.

---

## [0.4.1] - 2026-10-08

### Fixed
- **Server Shutdown Detection**: Use `net.ErrClosed` for accurate listener shutdown detection, preventing swallowed network errors.
- **Subdomain Collision Protection**: Atomically check and reject duplicate subdomain registrations using `sync.Map.LoadOrStore`.
- **Self-Hosted Client Resilience**: Added `sync.RWMutex` to eliminate state data races, exponential dial backoff to prevent CPU spin loops, and context-aware dialing for clean cancellation.
- **Proxy Lifecycle & 502 Handling**: Synchronized bidirectional proxy pumps with `sync.WaitGroup` for symmetric connection teardown, and added 502 Bad Gateway response when local target is offline.

## [0.4.0] - 2026-10-01

### Added
- **Awesome-Go Quality Readiness**: Achieved $\ge$ 80% statement test coverage across all core packages (`internal/server` 91.7%, `internal/provider` 80.5%, `internal/config` 91.9%, `internal/tunnel` 85.1%).
- **Server Tests**: End-to-end HTTP hijacking, data piping, and startup collision tests.
- **Provider Tests**: Comprehensive unit tests for `selfhosted`, `localtunnel`, and `cloudflare` providers.

### Fixed
- **Server Listener Race Condition**: Eliminated concurrent background listener binding and added `Server.Ready()` channel synchronization for deterministic, zero-race startup.
- **Test Flakiness**: Replaced non-deterministic `time.Sleep` calls with channel primitives.

---

## [0.3.0] - 2026-02-01

### Added
- **Automated Releases**: Binaries for Windows, Linux, and macOS are now automatically built on every tag.
- **Documentation**: New "Social Preview" banner, Comparison Table, and improved SEO guide.
- **Community**: Added "Help us Grow" section to Contributing guide.

---

## [0.2.0] - 2025-11-29

### Added
- **Cloudflare Tunnel** support (`expose tunnel -P cloudflare`)
- `--provider/-P` flag (localtunnel, cloudflare)
- Full test coverage for provider + service layers


## [0.1.2] - 2025-11-10

### Added
- Config management commands (`config list`, `config get`)
- Service layer with thread-safe tunnel management
- `--version` flag with commit and build date metadata

### Changed
- Improved error messages for tunnel lifecycle
- Better context cancellation handling

### Fixed
- Race conditions in Service.Start()
- Graceful shutdown on Ctrl+C

---

## [0.1.1] - 2025-11-09

### Added
- LocalTunnel provider integration
- 6 unit tests for Service layer (75%+ coverage)
- Provider interface for extensibility

### Changed
- Refactored tunnel command to use Service layer
- Separated CLI logic from business logic

---

## [0.1.0] - 2025-11-07

### Added
- Initial release
- `expose init` - Create `.expose.yml` config
- `expose tunnel` - Start local reverse proxy
- Cobra CLI framework
- GitHub Actions CI/CD (test.yml)
- Basic test coverage (tunnel package)

---

[Unreleased]: https://github.com/kernelshard/expose/compare/v0.6.1...HEAD
[0.6.1]: https://github.com/kernelshard/expose/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/kernelshard/expose/compare/v0.5.4...v0.6.0
[0.5.4]: https://github.com/kernelshard/expose/compare/v0.5.3...v0.5.4
[0.5.3]: https://github.com/kernelshard/expose/compare/v0.5.2...v0.5.3
[0.5.2]: https://github.com/kernelshard/expose/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/kernelshard/expose/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/kernelshard/expose/compare/v0.4.2...v0.5.0
[0.4.2]: https://github.com/kernelshard/expose/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/kernelshard/expose/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/kernelshard/expose/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/kernelshard/expose/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/kernelshard/expose/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/kernelshard/expose/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/kernelshard/expose/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/kernelshard/expose/releases/tag/v0.1.0

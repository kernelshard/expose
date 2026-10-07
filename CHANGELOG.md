# Changelog

All notable changes to Expose will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

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

[Unreleased]: https://github.com/kernelshard/expose/compare/v0.4.1...HEAD
[0.4.1]: https://github.com/kernelshard/expose/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/kernelshard/expose/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/kernelshard/expose/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/kernelshard/expose/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/kernelshard/expose/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/kernelshard/expose/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/kernelshard/expose/releases/tag/v0.1.0

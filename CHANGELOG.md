# Changelog

All notable changes to Expose will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]
### Planned for v0.2.0

### Planned for v0.2.0
- ngrok provider support
- Custom subdomains
- HTTPS support

## [v0.2.0] - 2025-11-29

### Added
- **Cloudflare Tunnel** support (`expose tunnel -P cloudflare`) [#21]
- `--provider/-P` flag (localtunnel, cloudflare)
- test coverage for provider + service layers

### Changed
- Bump version to v0.2.0
- Update README with providers table + examples
## [v0.2.0] - 2025-11-29

### Added
- **Cloudflare Tunnel** support (`expose tunnel -P cloudflare`) [#12]
- `--provider/-P` flag (localtunnel, cloudflare)
- Full test coverage for provider + service layers

### Changed
- Bump version to v0.2.0
- Update README with providers table + examples

---

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

[v0.1.2]: https://github.com/kernelshard/expose/compare/v0.2.0...v0.1.2
[0.1.1]: https://github.com/kernelshard/expose/releases/tag/v0.1.1
[0.1.0]: https://github.com/kernelshard/expose/releases/tag/v0.1.0

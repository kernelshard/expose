# Contributing to Expose

Thanks for your interest in contributing! This guide will help you get started.

---

## Development Setup

### Prerequisites
- Go 1.23+
- Git

### Clone and Build

```
git clone https://github.com/kernelshard/expose.git
cd expose
go mod download
go build -o expose ./cmd/expose
./expose --version
```

---

## Workflow

### Branch Strategy

1. **Always branch from `develop`** (not `main`)
   ```
   git checkout develop
   git pull origin develop
   git checkout -b feature/your-feature-name
   ```

2. **Branch naming:**
   - Features: `feat/add-ngrok-provider`
   - Fixes: `fix/config-load-error`
   - Docs: `docs/update-readme`
   - Tests: `test/add-provider-tests`

3. **Create PR to `develop`** (not `main`)
   - Base branch: `develop`
   - Title: Use commit prefix (`feat:`, `fix:`, `docs:`, `test:`)
   - Description: Reference issue number, explain changes

4. **After merge:** We merge `develop` → `main` only for releases

---

## Code Standards

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add ngrok provider support
fix: resolve config file not found error
docs: update installation instructions
test: add tunnel service tests
```

### Testing Requirements

- ✅ Write tests for new features
- ✅ Minimum **75% coverage**
- ✅ Run with race detector: `go test -race ./...`
- ✅ Use table-driven tests for multiple cases

**Example:**

```
func TestConfigLoad(t *testing.T) {
    tests := []struct {
        name    string
        config  string
        wantErr bool
    }{
        {"valid config", "port: 3000", false},
        {"invalid yaml", "port:", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test body
        })
    }
}
```

### Code Style

- Run `go fmt` before committing
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Keep CLI layer thin (delegate to service layer)
- Export struct fields only when needed (YAML marshaling)
- Return errors early

**File organization:**

```
internal/
├── cli/          # Cobra commands (newXXXCmd() factory pattern)
├── config/       # Config CRUD operations
├── provider/     # Provider implementations
├── tunnel/       # Service layer (business logic)
└── version/      # Version metadata
```

---

## Running Tests

```
# All tests with race detector
go test ./... -v -race -cover

# Specific package coverage
go test ./internal/config -cover
go test ./internal/tunnel -cover

# Coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Testing Locally

```
# Build
go build -o expose ./cmd/expose

# Run tunnel
./expose tunnel

# Test with HTTP server
python3 -m http.server 3000  # Terminal 1
./expose tunnel              # Terminal 2
curl <public-url>            # Terminal 3
```

---

## Pull Request Checklist

Before submitting:

- [ ] Tests pass: `go test ./... -race -cover`
- [ ] Code formatted: `go fmt ./...`
- [ ] Coverage ≥ 75%
- [ ] Commit messages follow convention
- [ ] PR description includes issue reference
- [ ] Branch is up to date with `develop`

---

## Need Help?

- **Issues:** [github.com/kernelshard/expose/issues](https://github.com/kernelshard/expose/issues)
- **Discussions:** Comment on relevant issue

---

**Made with ❤️ by contributors like you.**

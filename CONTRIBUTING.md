# Contributing to Bold Brew

Thanks for your interest in contributing! This guide will help you get set up and productive quickly.

## Prerequisites

- **Go 1.25+**
- **Homebrew** (for testing package operations)
- **Podman** (optional, for containerized builds)
- **golangci-lint** (for local linting)

## Getting Started

```bash
# Clone the repository
git clone https://github.com/Valkyrie00/bold-brew.git
cd bold-brew

# Install dependencies
make deps

# Build locally
make build-local

# Run the app
./bbrew
```

## Development Workflow

### Build & Run

```bash
make build-local          # Build binary
make run                  # Build and run
make clean                # Remove build artifacts
```

### Quality Checks

```bash
make test                 # Run tests
make test-coverage        # Run tests with coverage report
make quality-local        # Run golangci-lint
make security             # Run all security checks (govulncheck + gosec)
```

### Containerized Builds

```bash
make container-build-image    # Build the container image
make build                    # Build inside container
make quality                  # Run linter inside container
make release-snapshot         # Test release build
```

## Project Structure

```
bold-brew/
├── cmd/bbrew/               # Application entry point and CLI flags
├── internal/
│   ├── models/              # Data models
│   │   ├── package.go       # Unified Package type (formulae, casks, MAS)
│   │   ├── formula.go       # Homebrew formula JSON structure
│   │   ├── cask.go          # Homebrew cask JSON structure
│   │   ├── sort.go          # Sort mode enum
│   │   └── vulnerability.go # CVE vulnerability model
│   ├── services/            # Business logic
│   │   ├── app.go           # Application orchestrator and state
│   │   ├── brew.go          # Homebrew command execution
│   │   ├── dataprovider.go  # Data fetching, caching, and merging
│   │   ├── input.go         # Keyboard event handlers
│   │   ├── search.go        # Search, filter, and sort logic
│   │   ├── brewfile.go      # Brewfile parsing and loading
│   │   ├── export.go        # Brewfile export generation
│   │   ├── vulns.go         # brew vulns integration
│   │   ├── mas.go           # Mac App Store (mas) support
│   │   ├── flatpak.go       # Flatpak support
│   │   ├── cache.go         # XDG-compliant file caching
│   │   ├── command.go       # Streaming command executor
│   │   └── selfupdate.go    # Version check
│   └── ui/                  # Terminal UI layer
│       ├── layout.go        # Grid layout orchestration
│       ├── writer.go        # Thread-safe io.Writer for tview
│       ├── components/      # UI components (table, details, help, etc.)
│       └── theme/           # Color theme definitions
├── .github/workflows/       # CI/CD (quality, security, release, test-install)
├── site/                    # Website source (bold-brew.com)
├── examples/                # Example Brewfiles
├── .goreleaser.yaml         # Release configuration
└── Makefile                 # Build automation
```

## Architecture

The application follows a layered architecture:

```
┌─────────────────────────────────────────┐
│  cmd/bbrew (CLI entry point)            │
├─────────────────────────────────────────┤
│  services (business logic)              │
│  ┌──────────┬──────────┬──────────────┐ │
│  │ AppService│ Input   │ DataProvider │ │
│  │ (state)  │ (keys)  │ (fetch/cache)│ │
│  └──────────┴──────────┴──────────────┘ │
├─────────────────────────────────────────┤
│  ui (tview components + layout)         │
├─────────────────────────────────────────┤
│  models (shared data types)             │
└─────────────────────────────────────────┘
```

Key design decisions:
- **Services use `io.Writer`** for output — decoupled from tview, testable
- **`ThreadSafeWriter`** bridges services to the UI via `QueueUpdateDraw`
- **`DataProvider`** handles all Homebrew API/CLI data with parallel fetching
- **Models are value types** — `Package` is the unified view for all package sources

## Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add vulnerability scanning (v key)
fix: detect cask outdated by version comparison
refactor: decouple brew service from tview
docs: update README with new keybindings
ci: add test job to quality pipeline
chore: update dependencies
```

Scope is optional but helpful: `feat(brewfile): support mas entries`

## Pull Request Process

1. Create a feature branch from `main` (`feat/...`, `fix/...`, `refactor/...`)
2. Make your changes with tests where applicable
3. Ensure `make test quality-local` passes with 0 issues
4. Push and open a PR — CI will run lint, tests, security, and build
5. The build job gates on lint + tests passing

## Release Process

Releases are automated via GoReleaser:
1. Tag a semver version: `git tag v2.4.0`
2. Push the tag: `git push origin v2.4.0`
3. CI validates (lint + test), then builds and publishes:
   - GitHub Release with binaries (macOS/Linux, amd64/arm64)
   - Homebrew formula update in `Valkyrie00/homebrew-bbrew`
   - Changelog generated from conventional commits

## Tips

- Run `bbrew` with `-f examples/dev-tools.brewfile` to test Brewfile mode
- The app caches data in `$XDG_CACHE_HOME/bbrew/` — delete to force fresh fetch
- Use `make test-coverage` to generate `coverage.html` for visual coverage inspection

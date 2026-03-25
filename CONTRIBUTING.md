# Contributing to PhotoMeta

Thank you for your interest in contributing to PhotoMeta! This document provides guidelines and instructions for contributing to the project.

## 🛠 Prerequisites

Before you start, ensure you have the following installed:

*   **Go** (version 1.21 or higher)
*   **Git**
*   **Git LFS** (Required for test images ~200MB) - See [Installation](https://git-lfs.github.io/)
*   **GCC** (Required for Fyne GUI) - See [Fyne Prerequisites](https://developer.fyne.io/started/)
*   **golangci-lint** (Required for linting):
    ```bash
    # Linux/macOS
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    
    # Windows (PowerShell)
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    ```
    See also: [Installation Guide](https://golangci-lint.run/usage/install/#local-installation)

## 🚀 Getting Started

1.  **Fork the repository** on GitHub.
2.  **Clone your fork** locally:
    ```bash
    git clone https://github.com/your-username/photometa.git
    cd photometa
    ```
3.  **Download test images** (stored via Git LFS):
    ```bash
    git lfs install
    git lfs pull
    ```
4.  **Create a branch** for your feature or bug fix:
    ```bash
    git checkout -b feature/my-new-feature
    ```

## 🏗 Build and Run

### CLI & Server
The standard build does not require CGO unless specialized libraries are added.
```bash
go build ./cmd/photometa
```

### GUI
The GUI mode requires CGO and the Fyne toolkit dependencies.
```bash
go build -tags gui ./cmd/photometa
```

The resulting executable `photometa` (or `photometa.exe` on Windows) will be created in the project root directory.

## 🧪 Testing

Testing is a core part of our development process. **Every new feature, improvement, or bug fix must include corresponding tests** to ensure stability and prevent regressions.

Run all tests using `go test`. NOTE: GUI tests may require a graphical environment or mock drivers.

```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...
```

## 🔎 Check Code Style & Linting

We follow standard Go coding conventions. Please ensure your code is formatted correctly.

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run linter (REQUIRED before Pull Request)
golangci-lint run ./...
```

## 📐 Project Structure

This project follows the **Hexagonal Architecture (Ports and Adapters)** pattern:

| Directory | Purpose |
|-----------|---------|
| `cmd/photometa` | Application entry point (Composition Root) |
| `internal/adapter` | Driving adapters (CLI, GUI, HTTP Server) |
| `internal/analyzer` | Application service layer with metadata filler |
| `internal/domain` | Domain models (ImageFile, Metadata, etc.) |
| `internal/fake` | Test doubles for unit testing |
| `internal/format` | EXIF, IPTC, XMP parsers and format detection |
| `internal/platform` | Infrastructure (logger, locale, assets, version) |
| `internal/port` | Interface definitions (ImageAnalyzer, Logger) |
| `integration/` | Integration tests (cross-layer testing) |
| `docs/img/` | Sample images used for integration testing |

### Versioning

The project uses **Semantic Versioning (SemVer)**. Version information is injected at build time via ldflags.

| Variable | Description | Example |
|----------|-------------|---------|
| `version.Version` | Semantic version | `1.2.3` |
| `version.Commit` | Git commit hash | `abc1234` |
| `version.Date` | Build timestamp | `2026-03-20` |

**Build with version:**
```bash
go build -ldflags="-X github.com/DementorAK/photometa/internal/platform/version.Version=1.2.3 \
  -X github.com/DementorAK/photometa/internal/platform/version.Commit=$(git rev-parse --short HEAD) \
  -X github.com/DementorAK/photometa/internal/platform/version.Date=$(date +%Y-%m-%d)" ./cmd/photometa
```

**CLI usage:**
```bash
./photometa version          # Human-readable output
./photometa version --json   # JSON for CI/CD scripts
```

### Testing Infrastructure

*   `internal/fake/` — Contains test doubles (e.g., `FakeLogger`) that implement interfaces from `internal/port`. Use these in your unit tests to avoid external dependencies.
*   `integration/` — Contains tests that verify interactions between multiple layers (e.g., Analyzer → Format parsers → Domain models). Run with `go test ./integration/...`.
*   `docs/img/` — Sample images used for integration testing. Images are stored via Git LFS due to size (~100MB). Run `git lfs pull` after cloning to download them.

Please respect this separation of concerns when adding new features.

## 📝 Pull Request Process

1.  **Verify that all new or changed code is covered by tests.**
2.  Update the `README.md` or documentation in `docs/` with details of changes to the interface, this includes new environment variables, exposed ports, useful file locations and container parameters.
3.  Increase the version numbers in any examples files and the README.md to the new version that this Pull Request would represent.
4.  You may merge the Pull Request in once you have the sign-off of two other developers, or if you do not have permission to do that, you may request the second reviewer to merge it for you.

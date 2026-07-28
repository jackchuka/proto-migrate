# Proto-Migrate

Protocol Buffer migrations done right - AST-aware, safe, and blazing fast

[![Test](https://github.com/jackchuka/proto-migrate/actions/workflows/test.yml/badge.svg)](https://github.com/jackchuka/proto-migrate/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/jackchuka/proto-migrate.svg)](https://pkg.go.dev/github.com/jackchuka/proto-migrate)
[![Release](https://img.shields.io/github/v/release/jackchuka/proto-migrate?sort=semver)](https://github.com/jackchuka/proto-migrate/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Proto-Migrate provides automated, AST-aware transformations for Protocol Buffer schemas. It rewrites package names, imports, service names, and language-specific options while ensuring the resulting protobuf graph remains valid and backward-compatible.

## Why Proto-Migrate?

When evolving Protocol Buffer schemas across large codebases:

- **Package reorganization** becomes error-prone with manual find-and-replace
- **Import paths** need updating across hundreds of files
- **Language-specific options** (go_package, java_package) require consistent updates
- **Service renames** must maintain backward compatibility
- **Validation** is crucial to ensure changes don't break compilation

Proto-Migrate automates these transformations safely and reliably.

## Features

- 🔧 **AST-based transformations** - Precise modifications without regex fragility
- 📦 **Smart import resolution** - Updates import paths and language-specific options
- 🎯 **Multiple rule types** - Package renames, service renames, custom regex patterns
- ⚡ **Terraform-like workflow** - Plan → Diff → Apply with dry-run capabilities
- 🔍 **Built-in validation** - Ensures changes maintain compilation and compatibility
- 🚀 **Performance optimized** - Concurrent processing with configurable parallelism
- 🤖 **CI/CD ready** - JSON output, exit codes, and GitHub Actions support

## Installation

### From Source

```bash
# Requires Go 1.22+
go install github.com/jackchuka/proto-migrate/cmd/proto-migrate@latest
```

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/jackchuka/proto-migrate/releases).

### Verify Installation

```bash
proto-migrate version
```

## Quick Start

```bash
# 1. Initialize a migration config
proto-migrate init > .proto-migrate.yaml

# 2. Edit .proto-migrate.yaml with your rules

# 3. Preview changes (dry-run)
proto-migrate plan

# 4. View detailed diff
proto-migrate diff

# 5. Apply transformations
proto-migrate apply
```

## Configuration

Create a `.proto-migrate.yaml` file:

```yaml
# Define source and target directories
source: proto/oldpackage/v1
target: proto/newpackage/v1

# Exclude patterns (glob syntax)
excludes:
  - "*_test.proto"
  - "**/internal/**"
  - "vendor/**"

# Transformation rules
rules:
  # Package rename
  - kind: package
    from: oldpackage.v1
    to: newpackage.v1

  # Service rename
  - kind: service
    from: OldService
    to: NewService

  # Option updates (go_package, java_package, etc.)
  - kind: option
    from: oldpackage
    to: newpackage

  # Custom regex transformations
  - kind: regexp
    pattern: "oldpackage\\.v1\\."
    replace: "newpackage.v1."
```

### Rule Types

| Rule Kind | Description                 | Example                              |
| --------- | --------------------------- | ------------------------------------ |
| `package` | Renames protobuf packages   | `oldpkg.v1` → `newpkg.v1`            |
| `service` | Renames service definitions | `OldSvc` → `NewSvc`                  |
| `option`  | Updates file options        | Updates `go_package`, `java_package` |
| `regexp`  | Custom pattern matching     | Any regex pattern                    |

## Commands

### Core Commands

| Command | Description                             |
| ------- | --------------------------------------- |
| `init`  | Generate a starter configuration file   |
| `plan`  | Preview changes without modifying files |
| `diff`  | Show unified diff of pending changes    |
| `apply` | Execute transformations and write files |

### Command Examples

```bash
# Initialize configuration
proto-migrate init > .proto-migrate.yaml

# Preview changes
proto-migrate plan --config=.proto-migrate.yaml

# Show colorized diff
proto-migrate diff --color

# Apply
proto-migrate apply

# Apply with external dependency vendoring
proto-migrate apply --vendor-deps
```

### Global Flags

| Flag            | Description                     | Default     |
| --------------- | ------------------------------- | ----------- |
| `--config`      | Path to configuration file      | Auto-detect |
| `--concurrency` | Number of parallel workers      | CPU count   |
| `--vendor-deps` | Copy external protos to vendor/ | `false`     |

## Advanced Usage

### Working with Multiple Configs

```bash
# Process multiple migration configs
for config in migrations/*.yaml; do
  proto-migrate apply --config="$config"
done
```

### Programmatic Usage

```go
package main

import (
    "github.com/jackchuka/proto-migrate/pkg/protomigrate"
)

func main() {
    err := protomigrate.Run(protomigrate.Config{
        ConfigPath: ".proto-migrate.yaml",
        VendorDeps: true,
    })
    if err != nil {
        panic(err)
    }
}
```

## Comparison with Alternatives

| Feature                      | sed/awk | Buf CLI        | **Proto-Migrate** |
| ---------------------------- | ------- | -------------- | ----------------- |
| AST-aware transformations    | ❌      | ✅ (read-only) | ✅                |
| Automated rewrites           | ❌      | ❌             | ✅                |
| Import path resolution       | ❌      | ❌             | ✅                |
| Custom transformation rules  | Manual  | ❌             | ✅                |
| Compilation validation       | Manual  | ✅             | ✅                |
| Vendor external dependencies | ❌      | ❌             | ✅                |
| Dry-run capability           | ❌      | ❌             | ✅                |
| Batch operations             | Limited | ❌             | ✅                |

### Development

```bash
# Clone the repository
git clone https://github.com/jackchuka/proto-migrate
cd proto-migrate

# Install dependencies
go mod download

# Run tests
make test

# Run linters
make lint

# Build binary
make build
```

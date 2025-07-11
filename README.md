# Proto-Migrate

_A declarative toolkit for refactoring and migrating Protocol Buffers._

[![Go Reference](https://pkg.go.dev/badge/github.com/jackchuka/proto-migrate.svg)](https://pkg.go.dev/github.com/jackchuka/proto-migrate)
[![CI](https://github.com/jackchuka/proto-migrate/actions/workflows/ci.yml/badge.svg)](https://github.com/jackchuka/proto-migrate/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

proto-migrate rewrites package names, imports, service names, language-specific
options, and more—then validates that the resulting graph **still compiles and
is backward-compatible**.

---

## Features

- **AST-level rewrites** – no fragile regex munging.
- **Import resolver** – adjusts `import` paths, `go_package`, `java_package`,
  `swift_prefix`, … and can optionally vendor third-party protos.
- **Plan / Diff / Apply / Validate** sub-commands modelled after `terraform`.
- **Buf & `protoc` hooks** – lint and breaking-change checks run automatically.
- **Plugin architecture** – register your own `Rule` without forking.
- **CI-friendly** – JSON logs, deterministic output, atomic writes.

---

## Quick Start

```console
# Install (Go 1.22+)
go install github.com/jackchuka/proto-migrate/cmd/proto-migrate@latest

# Scaffold a migration plan
proto-migrate init > .proto-sync.yaml

# Dry run – see what will change
proto-migrate plan

# View a colorised unified diff
proto-migrate diff

# Apply the changes and ensure everything still compiles
proto-migrate apply --update-options --vendor-deps
```

---

## Configuration file

```yaml
# .proto-sync.yaml
source: proto/oldpackage/v1
target: proto/newpackage/v1

excludes:
  - "*ignore*.proto"
  - "*private*.proto"

rules:
  - kind: package
    from: oldpackage.v1
    to: newpackage.v1

  - kind: service
    from: OldService
    to: NewService

  - kind: regexp
    pattern: "oldpackage\\.v1\\."
    replace: "newpackage.v1."
```

---

## CLI commands

| Command    | Description                                                |
| ---------- | ---------------------------------------------------------- |
| `plan`     | Loads plan, prints summary (no writes).                    |
| `diff`     | Shows a unified diff of all pending rewrites.              |
| `apply`    | Executes the plan, rewrites files atomically.              |
| `validate` | Compiles with `protoc` and `buf lint/breaking`.            |
| `watch`    | Watches source directory and re-runs `plan` incrementally. |

Common flags:

```
--config           Path to .proto-sync.yaml   (default: auto-detect)
--proto_path       Extra -I paths for protoc
--vendor-deps      Copy missing externals to vendor/
--update-options   Rewrite go/java/swift options that embed old paths
--log-json         Machine-readable output
--concurrency N    Parallel file visits        (default: #CPU)
```

---

## CI integration

Add `.github/workflows/proto-migrate.yml`:

```yaml
name: proto-migrate

on: [pull_request]

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.22" }
      - run: go install github.com/jackchuka/proto-migrate/cmd/proto-migrate@latest
      - run: proto-migrate diff --exit-code # fails build if diffs exist
```

---

## Comparison

|                    | sed/awk | Buf `breaking` | **proto-migrate** |
| ------------------ | ------- | -------------- | ----------------- |
| AST-aware          | ✗       | ✓ (read-only)  | **✓**             |
| Automated rewrites | ✗       | ✗              | **✓**             |
| Import resolution  | ✗       | ✗              | **✓**             |
| Custom rules       | Manual  | ✗              | **✓**             |
| Compile validation | Manual  | ✗              | **✓**             |
| Vendor externals   | ✗       | ✗              | **✓**             |

---

## Contributing

1. Fork and create a feature branch.
2. Run `make test vet lint`.
3. Submit a PR with a concise title and description.
4. Ensure CI is green.

---

## License

Apache-2.0 © jackchuka

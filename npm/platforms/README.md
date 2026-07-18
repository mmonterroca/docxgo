# Platform binary packages

This directory contains platform-specific npm packages for `@mmonterroca/docxgo`.

Each subdirectory (`darwin-arm64`, `linux-x64`, etc.) becomes a separate npm package
containing the pre-built `docxgo` binary for that platform.

## How it works

1. **CI builds** (`release.yml`) → creates Go binaries for all platforms
2. **npm publish** (`npm-publish.yml`) → triggered by GH Release:
   - Builds each platform binary
   - Generates `package.json` for each platform package
   - Publishes `@mmonterroca/docxgo-{platform}` packages
   - Publishes the main `@mmonterroca/docxgo` package with matching versions

## Platform packages

| Package | OS | Arch |
|---------|------|------|
| `@mmonterroca/docxgo-darwin-arm64` | macOS | Apple Silicon (M1+) |
| `@mmonterroca/docxgo-darwin-x64` | macOS | Intel x64 |
| `@mmonterroca/docxgo-linux-x64` | Linux | x64 |
| `@mmonterroca/docxgo-linux-arm64` | Linux | ARM64 |
| `@mmonterroca/docxgo-win32-x64` | Windows | x64 |

## Local development

For local development, build the binary manually:

```bash
go build -o npm/bin/docxgo ./cmd/docxgo
```

The `resolveBinary()` function checks `npm/bin/docxgo` before trying platform packages.

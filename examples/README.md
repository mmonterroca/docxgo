# go-docx Examples

This directory contains practical, runnable examples demonstrating go-docx capabilities.

## Current Examples

### basic/ - Basic Document Creation
Simple example showing document creation, paragraphs, text runs, and tables in v2.

**Run:**
```bash
cd basic && go run main.go
```

**Features:**
- Document creation
- Adding paragraphs
- Text formatting (bold, italic, color, size)
- Basic tables

## Legacy v1 Examples

For v1 examples, see [`../legacy/v1/examples/`](../legacy/v1/examples/):
- `01_hello` - Hello World
- `02_formatted` - Formatted text
- `03_toc` - Table of Contents
- `v030_demo` - Professional document demo

> **Note**: v1 examples use the legacy API. For new projects, use v2 examples above.

## Requirements

- Go 1.21 or higher
- Microsoft Word or LibreOffice to view generated files

## More Examples Coming Soon

We're actively developing more v2 examples:
- Advanced text formatting
- Complex tables
- Headers and footers
- Images and drawings
- Styles and themes

## Contributing Examples

Have a great example? We'd love to include it! See [CONTRIBUTING.md](../CONTRIBUTING.md).

## Documentation

- [API Reference](https://pkg.go.dev/github.com/SlideLang/go-docx)
- [Design Document](../docs/V2_DESIGN.md)
- [Migration Guide](../MIGRATION.md) - Converting v1 examples to v2

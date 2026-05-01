# docxgo

Production-grade Microsoft Word .docx (OOXML) file manipulation in Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/mmonterroca/docxgo/v2.svg)](https://pkg.go.dev/github.com/mmonterroca/docxgo/v2)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/mmonterroca/docxgo)](https://goreportcard.com/report/github.com/mmonterroca/docxgo)

## Overview

**docxgo** is a powerful, clean-architecture library for creating Microsoft Word documents in Go. Built with production-grade code quality, comprehensive documentation, and modern design patterns.

### Key Features

- ✅ **Clean Architecture** - Interface-based design, dependency injection, separation of concerns
- ✅ **Type Safety** - No `interface{}`, explicit error handling throughout
- ✅ **Builder Pattern** - Fluent API for easy document construction
- ✅ **Thread-Safe** - Concurrent access supported with atomic operations
- ✅ **Production Ready** - EXCELLENT error handling, comprehensive validation
- ✅ **Well Documented** - Complete godoc, examples, and architecture docs
- ✅ **Template / Mail Merge** - Placeholder detection, data merging, batch document generation
- ✅ **In-Memory Images** - Insert images from byte slices without touching the file system
- ✅ **Open Source** - MIT License, use in commercial and private projects

---

## Status

**Current Version**: v2.4.0 (Stable)
**Stability**: Production Ready
**Released**: April 2026
**Test Coverage**: 50.7%

**Latest Features**: In-memory Image API (v2.4.0), Template / Mail Merge Engine (v2.3.0), Theme System (v2.1.0+), Round-trip Style Preservation (v2.2.1)

> **Note**: This library underwent a complete architectural rewrite in 2024-2025, implementing clean architecture principles, comprehensive testing, and modern Go practices. Version 2.0.0 released October 2025, with theme system added in v2.1.0, continued improvements through v2.2.x, template/mail merge engine in v2.3.0, and in-memory image insertion plus merged-cell round-trip fixes in v2.4.0.

---

## Installation

```bash
go get github.com/mmonterroca/docxgo/v2
```

### Requirements

- Go 1.21 or higher
- No external C dependencies
- Works on Linux, macOS, Windows

---

## Quick Start

### Option 1: Simple API (Direct Domain Interfaces)

```go
package main

import (
    "log"
    docx "github.com/mmonterroca/docxgo/v2"
)

func main() {
    // Create document
    doc := docx.NewDocument()

    // Add paragraph with formatted text
    para, _ := doc.AddParagraph()
    run, _ := para.AddRun()
    run.SetText("Hello, World!")
    run.SetBold(true)
    run.SetColor(docx.Red)

    // Save document
    if err := doc.SaveAs("simple.docx"); err != nil {
        log.Fatal(err)
    }
}
```

### Option 2: Builder API (Fluent, Chainable - Recommended)

```go
package main

import (
    "log"
    docx "github.com/mmonterroca/docxgo/v2"
    "github.com/mmonterroca/docxgo/v2/domain"
)

func main() {
    // Create builder with options
    builder := docx.NewDocumentBuilder(
        docx.WithTitle("My Report"),
        docx.WithAuthor("John Doe"),
        docx.WithDefaultFont("Calibri"),
        docx.WithDefaultFontSize(22), // 11pt in half-points
        docx.WithPageSize(docx.A4),
        docx.WithMargins(docx.NormalMargins),
    )

    // Add content using fluent API
    builder.AddParagraph().
        Text("Project Report").
        Bold().
        FontSize(16).
        Color(docx.Blue).
        Alignment(domain.AlignmentCenter).
        End()

    builder.AddParagraph().
        Text("This is bold text").Bold().
        Text(" and this is ").
        Text("colored text").Color(docx.Red).FontSize(14).
        End()

    // Build and save
    doc, err := builder.Build()
    if err != nil {
        log.Fatal(err)
    }

    if err := doc.SaveAs("report.docx"); err != nil {
        log.Fatal(err)
    }
}
```

### Option 3: Read and Modify Existing Documents 🆕

```go
package main

import (
    "log"
    docx "github.com/mmonterroca/docxgo/v2"
)

func main() {
    // Open existing document
    doc, err := docx.OpenDocument("template.docx")
    if err != nil {
        log.Fatal(err)
    }

    // Read existing content
    paragraphs := doc.Paragraphs()
    for _, para := range paragraphs {
        // Modify existing text
        runs := para.Runs()
        for _, run := range runs {
            if run.Text() == "PLACEHOLDER" {
                run.SetText("Updated Value")
                run.SetBold(true)
            }
        }
    }

    // Add new content
    newPara, _ := doc.AddParagraph()
    newRun, _ := newPara.AddRun()
    newRun.SetText("This paragraph was added by the reader")

    // Save modified document
    if err := doc.SaveAs("modified.docx"); err != nil {
        log.Fatal(err)
    }
}
```

### More Examples

See the [`examples/`](examples/) directory for 15 comprehensive examples:

- **[01_basic](examples/01_basic/)** - Simple document with builder pattern
- **[02_intermediate](examples/02_intermediate/)** - Professional product catalog
- **[03_toc](examples/03_toc/)** - Table of contents generation
- **[04_fields](examples/04_fields/)** - TOC, page numbers, hyperlinks
- **[05_styles](examples/05_styles/)** - Built-in style system (40+ styles)
- **[06_sections](examples/06_sections/)** - Page layout, margins, orientation
- **[07_advanced](examples/07_advanced/)** - All features combined
- **[08_images](examples/08_images/)** - Image insertion and positioning
- **[09_advanced_tables](examples/09_advanced_tables/)** - Cell merging, nested tables, styling
- **[10_paragraph_spacing](examples/10_paragraph_spacing/)** - Line and paragraph spacing
- **[11_multi_section](examples/11_multi_section/)** - Multi-section layouts
- **[12_read_and_modify](examples/12_read_and_modify/)** - Read and modify existing documents
- **[13_themes](examples/13_themes/)** - Theme system with 7 preset themes
- **[14_mail_merge](examples/14_mail_merge/)** - Template engine with mail merge and batch generation
- **[15_external_template](examples/15_external_template/)** - Mail merge with external Word template (MERGEFIELD)

---

## Architecture

This library follows clean architecture principles with clear separation of concerns:

```
github.com/mmonterroca/docxgo/v2/
├── domain/          # Core interfaces (public API)
│   ├── document.go  # Document interface
│   ├── paragraph.go # Paragraph interface
│   ├── run.go       # Run interface
│   ├── table.go     # Table interfaces
│   ├── section.go   # Section interfaces
│   ├── image.go     # Image interfaces
│   └── style.go     # Style interfaces
│
├── internal/        # Internal implementations
│   ├── core/        # Core domain implementations
│   ├── manager/     # Service managers (ID, media, style, relationship)
│   ├── reader/      # Document reading/parsing
│   ├── serializer/  # XML serialization
│   ├── writer/      # .docx file writing
│   └── xml/         # OOXML structures
│
├── pkg/             # Public utilities
│   ├── errors/      # Structured error types
│   ├── constants/   # OOXML constants
│   └── color/       # Color utilities
│
├── themes/          # Theme system (7 preset themes)
│
└── examples/        # 13 usage examples
    ├── 01_basic/    ... 13_themes/
```

### Design Principles

1. **Interface Segregation** - Small, focused interfaces
2. **Dependency Injection** - No global state
3. **Explicit Errors** - Errors returned immediately, not silently ignored
4. **Immutability** - Defensive copies to prevent external mutation
5. **Type Safety** - Strong typing, no `interface{}`
6. **Thread Safety** - Concurrent access supported
7. **Documentation** - Every public method documented

---

## Features

### ✅ Fully Implemented

**Core Document Structure**

- Document creation with metadata (title, author, subject, keywords)
- Paragraphs with comprehensive formatting
- Text runs with character-level styling
- Tables with rows, cells, and styling
- Sections with page layout control

**Text Formatting**

- Bold, italic, underline, strikethrough
- Font color (RGB), size, and family
- Highlight colors (15 options)
- Alignment (left, center, right, justify)
- Line spacing (single, 1.5, double, custom)
- Indentation (left, right, first-line, hanging)

**Advanced Table Features** (Phase 9 - Complete)

- **Cell Merging**: Horizontal (colspan) and vertical (rowspan)
- **Nested Tables**: Tables within table cells
- **8 Built-in Styles**: Normal, Grid, Plain, MediumShading, LightShading, Colorful, Accent1, Accent2
- Row height control
- Cell width and alignment
- Borders and shading

**Images & Media** (Phase 8 - Complete)

- **9 Image Formats**: PNG, JPEG, GIF, BMP, TIFF, SVG, WEBP, ICO, EMF
- Inline and floating images
- Custom dimensions (pixels, inches, EMUs)
- Positioning (left, center, right, custom coordinates)
- Automatic format detection
- Relationship management

**Fields & Dynamic Content** (Phase 6 - Complete)

- **Table of Contents (TOC)**: Auto-generated with styles
- **Page Numbers**: Current page, total pages
- **Hyperlinks**: External URLs and internal bookmarks
- **StyleRef**: Dynamic text from heading styles
- **Date/Time**: Document creation/modification dates
- **Custom Fields**: Extensible field system

**Headers & Footers** (Phase 6 - Complete)

- Default, first page, and even/odd page headers/footers
- Page numbering in footers
- Dynamic content with fields
- Per-section customization

**Styles System** (Phase 6 - Complete)

- **40+ Built-in Styles**: All standard Word paragraph styles
- **Character Styles**: For inline formatting
- **Custom Styles**: Create and apply user-defined styles
- Style inheritance and cascading

**Builder Pattern** (Phase 6.5 - Complete)

- Fluent API for easy document construction
- Error accumulation (no intermediate error checking)
- Chainable methods for all operations
- Functional options for configuration

**Quality & Reliability** (Phase 11 - Complete)

- **EXCELLENT Error Handling**: Structured errors with rich context
- Comprehensive validation at every layer
- Thread-safe ID generation (atomic counters)
- **50.7% Test Coverage** (improvement plan ready: → 95%)
- **0 Linter Warnings** (golangci-lint with 30+ linters)
- Complete godoc documentation

### Planned Features

- Mail merge and templates
- Comments and change tracking
- Custom XML parts
- Advanced drawing shapes
- Document comparison
- Content controls

---

## Error Handling

All operations return explicit errors - no silent failures. The error system was rated **EXCELLENT** in Phase 11 review:

```go
// Structured errors with full context
para, err := doc.AddParagraph()
if err != nil {
    // Error contains: operation, code, message, and context
    // Example: "operation=Document.AddParagraph | code=VALIDATION_ERROR | ..."
    log.Fatal(err)
}

// Validation errors with detailed information
err := run.SetSize(10000) // Invalid size
if err != nil {
    // Returns: ValidationError with field, value, and constraint details
    var validationErr *errors.ValidationError
    if errors.As(err, &validationErr) {
        fmt.Printf("Field '%s' failed: %s\n", validationErr.Field, validationErr.Message)
    }
}

// Builder pattern accumulates errors
builder := docx.NewDocumentBuilder()
builder.AddParagraph().
    Text("Hello").
    FontSize(9999). // Invalid - error recorded
    Bold().
    End()

// All errors surface at Build()
doc, err := builder.Build()
if err != nil {
    // Returns first accumulated error with full context
    log.Fatal(err)
}
```

**Error System Features**:

- ✅ **DocxError**: Structured errors with operation context
- ✅ **ValidationError**: Domain-specific validation errors
- ✅ **BuilderError**: Error accumulation for fluent API
- ✅ **7 Error Codes**: Well-defined error categories
- ✅ **10+ Helper Functions**: Easy error creation
- ✅ **100% Best Practices**: Proper wrapping, context, no panics

See [docs/ERROR_HANDLING.md](docs/ERROR_HANDLING.md) for comprehensive review.

---

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package
go test -v ./internal/core

# Run benchmarks
go test -bench=. ./...
```

**Current Test Coverage**: 50.7%

See [docs/COVERAGE_ANALYSIS.md](docs/COVERAGE_ANALYSIS.md) for detailed coverage analysis.

---

## Documentation

### 📖 Complete Documentation Suite

**For Users:**

- **[V2 API Guide](docs/V2_API_GUIDE.md)** - Complete v2 API reference with examples ⭐ START HERE
- **[Implementation Status](docs/IMPLEMENTATION_STATUS.md)** - What's implemented, what's planned
- **[Examples](examples/README.md)** - 13 working code examples
- **[Migration Guide](MIGRATION.md)** - Migrating from v1 to v2

**For Developers:**

- **[V2 Design](docs/V2_DESIGN.md)** - Architecture, design decisions, and roadmap
- **[Contributing Guide](CONTRIBUTING.md)** - How to contribute
- **[Error Handling](docs/ERROR_HANDLING.md)** - Complete error system review
- **[Coverage Analysis](docs/COVERAGE_ANALYSIS.md)** - Test coverage report

**Quick Links:**

- [API Reference (pkg.go.dev)](https://pkg.go.dev/github.com/mmonterroca/docxgo/v2)
- [Documentation Index](docs/README.md) - Complete documentation guide
- [Credits](CREDITS.md) - Project history and contributors

---

## Performance

Optimized for real-world usage:

- **Pre-allocated slices** with sensible defaults (paragraphs: 32, tables: 8)
- **Thread-safe atomic counters** for ID generation
- **Lazy loading** of relationships and media
- **Efficient string building** for text extraction
- **Memory-conscious** defensive copies only when necessary

**Benchmarks** (coming in Phase 11.5):

- Simple document creation: target < 1ms
- Complex document (100 paragraphs, 10 tables): target < 50ms
- Image insertion: target < 5ms per image

---

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for:

- Code of conduct
- Development workflow (Git Flow)
- Testing requirements
- Pull request process

### Quick Contribution Guide

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Ensure tests pass (`go test ./...`)
5. Commit your changes (follow commit message conventions)
6. Push to your fork
7. Open a Pull Request

---

## License

**MIT License**

This means:

- ✅ **Free to use** in any project (commercial or personal)
- ✅ **No copyleft** - modifications don't need to be shared
- ✅ **Permissive** - do almost anything you want
- ✅ **Attribution required** - keep copyright notices

See [LICENSE](LICENSE) for full text.

### Copyright

```
Copyright (C) 2024-2026 Misael Monterroca
Copyright (C) 2022-2024 fumiama (original enhancements)
Copyright (C) 2020-2022 Gonzalo Fernández-Victorio (original library)
```

See [CREDITS.md](CREDITS.md) for complete project history.

---

## Credits & History

This project evolved through multiple stages:

1. **gonfva/docxlib** (2020-2022) - Original library by Gonzalo Fernández-Victorio
2. **fumiama/go-docx** (2022-2024) - Enhanced fork with images, tables, shapes
3. **mmonterroca/docxgo v1** (2023-2024) - Professional features (headers, TOC, links)
4. **mmonterroca/docxgo v2** (2024-2026) - Complete architectural rewrite

**Current Maintainer**: Misael Monterroca (misael@monterroca.com)
**GitHub**: [@mmonterroca](https://github.com/mmonterroca)

**V2 Rewrite**:

- All development phases completed
- 6,646+ lines of production code
- 1,500+ lines of documentation
- Clean architecture implementation
- Production-grade quality

For complete project genealogy, see [CREDITS.md](CREDITS.md).

---

## Roadmap

### Release History

- **v2.4.0** (April 2026 - Current Stable)
  - In-memory image API: `AddImageFromBytes`, `AddImageFromBytesWithSize`, `AddImageFromBytesWithPosition`
  - Round-trip preservation of `w:gridSpan` and `w:vMerge` for merged table cells
  - CLI handler now embeds base64 images without temp files

- **v2.3.0** (February 2026)
  - Template / Mail Merge engine (`pkg/template/`)
  - MERGEFIELD support with real Word templates and `«»` delimiters
  - Paragraph mutation APIs (`ClearRuns`, `RemoveRun`, `InsertRunAt`)
  - Examples 14 and 15 (mail merge + external Word templates)

- **v2.2.2** (February 2026)
  - Table style borders fix (TableGrid, PlainTable1 now render with visible borders)
  - Grid column width calculation from cell widths
  - Comprehensive documentation overhaul
  - Restored 13_themes example

- **v2.2.1** (January 2026)
  - Hyperlink RelationshipID preservation during round-trip
  - Drawing position serialization fix
  - Internal hyperlink anchor support
  - Custom styles preservation during round-trip

- **v2.2.0** (January 2026)
  - Table cell border serialization with explicit properties

- **v2.1.1** (November 2025)
  - Go module path fix (/v2 suffix)

- **v2.1.0** (October 2025)
  - Complete theme system with 7 preset themes
  - Custom theme support (Clone, WithColors, WithFonts, WithSpacing)

- **v2.0.1** (November 2025)
  - Go module path fix (/v2 suffix)

- **v2.0.0** (October 2025)
  - Production-ready stable release
  - Clean architecture implementation
  - Document reading and modification
  - Comprehensive documentation

- **v2.0.0-beta** (October 2025)
  - Initial beta release

### Planned Features

- Comments and change tracking
- Content controls / form fields
- Custom XML parts
- Advanced drawing shapes
- Document comparison
- Header/footer reading round-trip

See [docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md) for detailed status.

---

## Support & Community

- **Issues**: [GitHub Issues](https://github.com/mmonterroca/docxgo/issues)
- **Discussions**: [GitHub Discussions](https://github.com/mmonterroca/docxgo/discussions)
- **Email**: misael@monterroca.com
- **Documentation**: [pkg.go.dev](https://pkg.go.dev/github.com/mmonterroca/docxgo/v2)

### Reporting Bugs

Please include:

- Go version (`go version`)
- OS and architecture
- Minimal reproduction code
- Expected vs actual behavior

---

## Related Projects

- **[UniOffice](https://github.com/unidoc/unioffice)** - Commercial, comprehensive OOXML library
- **[gingfrederik/docx](https://github.com/gingfrederik/docx)** - Write-only docx library
- **[Office Open XML](http://www.ecma-international.org/publications/standards/Ecma-376.htm)** - OOXML specification

### Why Choose docxgo?

- ✅ **Free & Open Source** - MIT License, no restrictions
- ✅ **Clean Architecture** - Production-grade code quality
- ✅ **Feature Complete** - 95% of planned features implemented
- ✅ **EXCELLENT Error Handling** - Structured errors, rich context
- ✅ **Well Documented** - Complete godoc, examples, architecture docs
- ✅ **Active Development** - Regular updates, responsive to issues
- ✅ **Modern Go** - Follows current best practices (Go 1.21+)
- ✅ **Builder Pattern** - Fluent API for easy document construction

**Comparison**:

- **UniOffice** - Commercial ($$$), more features, heavier
- **gingfrederik/docx** - Write-only, simpler, less features
- **docxgo** - Free, balanced features, production-ready

---

**Made with ❤️ by [Misael Monterroca](https://github.com/mmonterroca)**

_Star ⭐ this repo if you find it useful!_

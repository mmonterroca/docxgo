# docxgo

[![Go Reference](https://pkg.go.dev/badge/github.com/mmonterroca/docxgo.svg)](https://pkg.go.dev/github.com/mmonterroca/docxgo)
[![Go Report Card](https://goreportcard.com/badge/github.com/mmonterroca/docxgo)](https://goreportcard.com/report/github.com/mmonterroca/docxgo)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

A modern, clean-architecture Go library for creating and manipulating Microsoft Word (.docx) documents programmatically.

## Status: Beta (Phase 6 Complete) ✅

v2 is feature-complete for Phase 6 (Advanced Features) and ready for beta testing. All core functionality, including sections, headers/footers, fields, and styles, is implemented and tested.

## Features

### Core Features (Phase 1-3) ✅
- ✅ **Clean Architecture**: Domain-driven design with clear separation of concerns
- ✅ **Document Creation**: Create new DOCX files from scratch
- ✅ **Paragraphs**: Add and format paragraphs with full control
- ✅ **Text Runs**: Fine-grained text formatting (bold, italic, color, fonts, sizes)
- ✅ **Tables**: Create and format tables with rows, columns, and cells
- ✅ **Relationships**: Internal relationship management
- ✅ **Serialization**: OOXML-compliant XML generation

### Advanced Features (Phase 6) ✅
- ✅ **Sections**: Complete page layout control (size, orientation, margins, columns)
- ✅ **Headers/Footers**: Three types (Default, First, Even) with full formatting
- ✅ **Fields**: Nine field types including TOC, page numbers, hyperlinks, dates
- ✅ **Styles**: 40+ built-in styles with StyleManager and custom style support

### Coming Soon
- ⏳ **Phase 7**: Complete documentation and v2.0.0-beta release
- ⏳ **Phase 8**: Media support (images, charts)
- ⏳ **Phase 9**: Advanced tables and complex layouts

## Installation

```bash
go get github.com/mmonterroca/docxgo
```

## Quick Start

### Basic Document

```go
package main

import (
    docx "github.com/mmonterroca/docxgo"
    "github.com/mmonterroca/docxgo/domain"
)

func main() {
    // Create a new document
    doc := docx.NewDocument()
    
    // Add a paragraph with styled text
    para, _ := doc.AddParagraph()
    para.SetStyle(domain.StyleIDHeading1)
    run, _ := para.AddRun()
    run.AddText("Hello, docxgo!")
    
    // Save the document
    doc.SaveToFile("hello.docx")
}
```

### Page Layout and Headers

```go
// Configure page layout
section, _ := doc.DefaultSection()
section.SetPageSize(domain.PageSizeA4)
section.SetOrientation(domain.OrientationPortrait)
section.SetMargins(domain.DefaultMargins)

// Add header with page numbers
header, _ := section.Header(domain.HeaderDefault)
headerPara, _ := header.AddParagraph()
headerPara.SetAlignment(domain.AlignmentRight)
headerRun, _ := headerPara.AddRun()
headerRun.AddText("Document Title")

// Add footer with "Page X of Y"
footer, _ := section.Footer(domain.FooterDefault)
footerPara, _ := footer.AddParagraph()
footerPara.SetAlignment(domain.AlignmentCenter)

r1, _ := footerPara.AddRun()
r1.AddText("Page ")

r2, _ := footerPara.AddRun()
r2.AddField(docx.NewPageNumberField())

r3, _ := footerPara.AddRun()
r3.AddText(" of ")

r4, _ := footerPara.AddRun()
r4.AddField(docx.NewPageCountField())
```

### Fields and Dynamic Content

```go
// Table of Contents
tocPara, _ := doc.AddParagraph()
tocRun, _ := tocPara.AddRun()
tocOptions := map[string]string{
    "levels":     "1-3",
    "hyperlinks": "true",
}
tocRun.AddField(docx.NewTOCField(tocOptions))

// Hyperlink
linkPara, _ := doc.AddParagraph()
linkRun, _ := linkPara.AddRun()
linkField := docx.NewHyperlinkField(
    "https://github.com/mmonterroca/docxgo",
    "Visit go-docx",
)
linkRun.SetColor(0x0000FF)
linkRun.SetUnderline(domain.UnderlineSingle)
linkRun.AddField(linkField)
```

### Styles

```go
// Use built-in styles with type-safe constants
title, _ := doc.AddParagraph()
title.SetStyle(domain.StyleIDTitle)

heading, _ := doc.AddParagraph()
heading.SetStyle(domain.StyleIDHeading1)

quote, _ := doc.AddParagraph()
quote.SetStyle(domain.StyleIDIntenseQuote)

// Query styles with StyleManager
styleMgr := doc.StyleManager()
if styleMgr.HasStyle(domain.StyleIDHeading2) {
    para.SetStyle(domain.StyleIDHeading2)
}
```

## Examples

Comprehensive examples are available in the [`examples/`](./examples/) directory:

- **[Example 01 - Basic](./examples/basic/)**: Document creation and basic formatting
- **[Example 04 - Fields](./examples/04_fields/)**: TOC, page numbers, hyperlinks, dates
- **[Example 05 - Styles](./examples/05_styles/)**: Built-in styles and StyleManager
- **[Example 06 - Sections](./examples/06_sections/)**: Page layout, headers, footers
- **[Example 07 - Advanced](./examples/07_advanced/)**: Complete integration demo

Run any example:
```bash
cd examples/07_advanced
go run main.go
```

## Documentation

- **[API Documentation](./docs/API_DOCUMENTATION.md)**: Complete API reference
- **[Migration Guide](./MIGRATION.md)**: Upgrading from v1 to v2
- **[Design Document](./docs/V2_DESIGN.md)**: Architecture and design decisions
- **[Contributing](./CONTRIBUTING.md)**: How to contribute to the project

## Architecture

docxgo follows **Clean Architecture** principles:

```
domain/         # Interfaces and business logic
  ├── document.go
  ├── paragraph.go
  ├── run.go
  ├── section.go
  └── style.go

internal/       # Implementation details
  ├── core/       # Core implementations
  ├── manager/    # ID, Relationship, Style managers
  ├── serializer/ # XML serialization
  ├── writer/     # ZIP file writing
  └── xml/        # OOXML structures

pkg/            # Shared utilities
  ├── color/
  └── constants/
```

This architecture provides:
- **Testability**: Easy to mock and test
- **Maintainability**: Clear separation of concerns
- **Extensibility**: Simple to add new features
- **Thread-Safety**: Concurrent-safe by design

## Roadmap

### ✅ Phase 1-3: Foundation (Complete)
- Core document structure
- Paragraphs, runs, and basic formatting
- Tables and relationships
- OOXML serialization

### ✅ Phase 5.5: Restructuring (Complete)
- Clean architecture implementation
- Domain-driven design
- Manager pattern for shared resources

### ✅ Phase 6: Advanced Features (Complete)
- Section management (page size, orientation, margins)
- Headers and footers (three types)
- Field system (9 field types)
- Style management (40+ built-in styles)

### 🚧 Phase 7: Documentation & Release (In Progress - 50%)
- ✅ API documentation updates
- ✅ Migration guide enhancements
- ✅ Example collection (05, 06, 07)
- 🚧 README updates
- ⏳ CHANGELOG creation
- ⏳ godoc comments

### ⏳ Phase 8: Media & Images
- Image insertion
- Image formatting and positioning
- Charts and diagrams

### ⏳ Phase 9: Advanced Tables
- Table styles and themes
- Cell merging and splitting
- Advanced table formatting

## Performance

v2 is designed for efficiency:
- **Lazy initialization**: Resources created only when needed
- **Thread-safe**: Safe for concurrent use with RWMutex
- **Minimal allocations**: Efficient memory usage
- **Streaming output**: Large documents supported

## Compatibility

- **Go Version**: Requires Go 1.21 or higher
- **Output Format**: Office Open XML (OOXML) / ECMA-376
- **Compatible With**:
  - Microsoft Word 2007+
  - LibreOffice Writer
  - Google Docs
  - Pages (macOS)
  - Any OOXML-compatible word processor

## Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/core/
go test ./internal/manager/
```

Current test coverage: **95%+**

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/mmonterroca/docxgo.git
cd go-docx/v2

# Install dependencies
go mod download

# Run tests
go test ./...

# Run examples
go run ./examples/basic/main.go
```

## License

This project is licensed under the **MIT License**.

See [LICENSE](../LICENSE) for full details.

### What this means

✅ **Commercial use allowed** - Use in proprietary/commercial projects  
✅ **Private use allowed** - Use in private/internal projects  
✅ **Modification allowed** - Modify and customize the code  
✅ **Distribution allowed** - Share and distribute  

Only requirement: Include the MIT license and copyright notice.

## Credits

docxgo is a complete rewrite with clean architecture principles. While inspired by earlier work, it represents a ground-up redesign focused on maintainability, testability, and modern Go practices.

Original inspiration: [fumiama/go-docx](https://github.com/fumiama/go-docx)

## Support

- **Issues**: [GitHub Issues](https://github.com/mmonterroca/docxgo/issues)
- **Discussions**: [GitHub Discussions](https://github.com/mmonterroca/docxgo/discussions)
- **Documentation**: [docs/](./docs/)

## Changelog

See [CHANGELOG.md](./CHANGELOG.md) for version history and release notes.

---

**Built with ❤️ by the go-docx community**

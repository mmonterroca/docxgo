# Release Notes - v2.0.0-beta

## 🎉 Major Rewrite: go-docx v2

This is a **complete rewrite** of go-docx with improved architecture, comprehensive error handling, and full OOXML compliance.

### ⚠️ Breaking Changes

This is a **major version update** (v1.x → v2.x) with breaking API changes. Existing code using v1.x will need to be updated.

**Migration Guide**: See [docs/V2_API_GUIDE.md](docs/V2_API_GUIDE.md) and [MIGRATION.md](MIGRATION.md)

### ✨ New Features

#### Document Structure
- **Section Management**: Full support for multiple sections with independent page settings
- **Headers & Footers**: Per-section headers/footers with first-page and odd/even variants
- **Page Layout**: Configurable page sizes, orientations, margins, and columns
- **Fields Support**: TOC, page numbers, hyperlinks, dates, and custom fields

#### Content Creation
- **Advanced Tables**: Cell merging (colspan/rowspan), vertical alignment, shading, borders
- **Images**: Inline and floating images with precise positioning
- **Styles**: Built-in Word styles (Normal, Heading1-9, Title, etc.) with custom style support
- **Rich Text**: Bold, italic, underline, colors, fonts, sizes, highlights

#### Architecture
- **Domain-Driven Design**: Clean separation of domain, serialization, and infrastructure
- **Error Handling**: Comprehensive error types with context and validation
- **Type Safety**: Strong typing throughout with proper interfaces
- **Memory Efficient**: Streaming ZIP writer, optimized serialization

### 🐛 Bug Fixes

- Fixed character encoding issues in XML serialization
- Corrected relationship ID management for images and hyperlinks
- Resolved section break serialization order
- Fixed table cell border rendering
- Corrected bookmark generation for TOC functionality

### 📚 Documentation

- Complete API documentation with examples
- 11 working examples covering all major features
- Migration guide from v1.x
- Architecture documentation
- Error handling patterns

### 🔧 Developer Experience

- **CI/CD**: GitHub Actions with golangci-lint v2.5.0, tests, and examples validation
- **Code Quality**: Strict linting for production code, lenient for examples
- **Testing**: Comprehensive unit tests with >80% coverage
- **Examples**: All examples include README and validate against Word

### 📦 Installation

```bash
go get github.com/mmonterroca/docxgo@v2.0.0-beta
```

### 🚀 Quick Start

```go
package main

import (
    "github.com/mmonterroca/docxgo/v2"
    "github.com/mmonterroca/docxgo/domain"
)

func main() {
    doc := docx.NewDocument()
    
    para, _ := doc.AddParagraph()
    run, _ := para.AddRun()
    run.SetText("Hello, World!")
    run.SetBold(true)
    run.SetFontSize(14) // points
    
    doc.SaveAs("hello.docx")
}
```

### 📖 Examples

See [examples/](examples/) directory:
- `01_basic` - Simple document creation
- `02_intermediate` - Paragraphs, runs, formatting
- `03_toc` - Table of Contents
- `04_fields` - Page numbers, dates, hyperlinks
- `05_styles` - Built-in and custom styles
- `06_sections` - Multi-section documents with headers/footers
- `07_advanced` - Complex layouts
- `08_images` - Image insertion and positioning
- `09_advanced_tables` - Table features (merging, styling)
- `11_multi_section` - Multiple sections with different layouts

### 🔗 Compatibility

- **Go Version**: 1.23+
- **OOXML**: Office Open XML (ISO/IEC 29500)
- **Word Compatibility**: Microsoft Word 2007+, LibreOffice, Google Docs

### 🙏 Credits

Based on the original [fumiama/go-docx](https://github.com/fumiama/go-docx) library.  
Completely rewritten with new architecture and features by [@mmonterroca](https://github.com/mmonterroca).

### 📄 License

MIT License - See [LICENSE](LICENSE) file for details.

### 🐛 Known Issues

- TOC requires manual "Update Field" in Word (F9) - inherent OOXML limitation
- Complex table borders may require fine-tuning
- Some advanced Word features not yet implemented (see roadmap)

### 🗺️ Roadmap

See [docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md) for planned features.

---

**Full Changelog**: https://github.com/mmonterroca/docxgo/compare/v1.0.0...v2.0.0-beta

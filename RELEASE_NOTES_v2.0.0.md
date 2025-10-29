# Release Notes - v2.0.0 (Stable)

## 🎉 go-docx v2.0.0 - Production Ready Release

**Release Date**: October 29, 2025

We're excited to announce **v2.0.0 stable**, the production-ready release of go-docx! This version includes all features from the beta release plus major new capabilities for **reading and modifying existing documents**.

### 🆕 What's New Since Beta

#### 🚀 Phase 10: Document Reading (NEW!)

The biggest addition to v2.0.0 is full support for **reading and modifying existing .docx files**:

- ✅ **Open existing documents** - `docx.OpenDocument("file.docx")`
- ✅ **Read document structure** - Access paragraphs, runs, tables, and styles
- ✅ **Modify existing content** - Edit text, change formatting, update table cells
- ✅ **Preserve document styles** - Maintains Title, Subtitle, Headings, Quote, Normal styles
- ✅ **Add new content** - Insert new paragraphs, runs, and sections into existing documents
- ✅ **Round-trip capability** - Create → Save → Open → Modify → Save workflow

**Example: Read and Modify Documents**

```go
package main

import (
    "log"
    docx "github.com/mmonterroca/docxgo"
)

func main() {
    // Open existing document
    doc, err := docx.OpenDocument("template.docx")
    if err != nil {
        log.Fatal(err)
    }
    
    // Read and modify existing content
    paragraphs := doc.Paragraphs()
    for _, para := range paragraphs {
        runs := para.Runs()
        for _, run := range runs {
            if run.Text() == "PLACEHOLDER" {
                run.SetText("Updated Value")
                run.SetBold(true)
                run.SetColor(docx.Blue)
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

See [`examples/12_read_and_modify/`](examples/12_read_and_modify/) for complete examples.

#### 🐛 Critical Bug Fixes

- **Fixed style preservation** - Paragraph styles (Title, Subtitle, Heading1-9, Quote, Normal, ListParagraph) are now correctly preserved when reading and modifying documents
- **Fixed README API examples** - Corrected all Quick Start examples to show accurate API signatures and usage patterns
- **Enhanced example 12** - Now demonstrates both editing existing content AND adding new content

#### 📚 Documentation Improvements

- **Updated README.md** - Fixed incorrect API examples, separated Simple API vs Builder API clearly
- **Added Option 3** - New Quick Start section showing document reading/modification workflow
- **Updated implementation status** - Phase 10 now marked as "60% complete - core features working"
- **Corrected example count** - All documentation now references 11 working examples (was 9)

---

## ✨ Complete Feature List

### Document Structure
- **Document Management** - Create, open, read, modify, and save .docx files
- **Section Support** - Multiple sections with independent page settings
- **Headers & Footers** - Per-section headers/footers with first-page and odd/even variants
- **Page Layout** - Configurable page sizes, orientations, margins, and columns
- **Fields** - TOC, page numbers, hyperlinks, dates, and custom fields

### Content Creation
- **Paragraphs** - Full formatting with alignment, indentation, spacing, and styles
- **Text Runs** - Bold, italic, underline, colors, fonts, sizes, highlights
- **Advanced Tables** - Cell merging (colspan/rowspan), vertical alignment, shading, borders, 8 built-in styles
- **Images** - Inline and floating images (9 formats: PNG, JPEG, GIF, BMP, TIFF, SVG, WEBP, ICO, EMF)
- **Styles** - 40+ built-in Word styles (Normal, Heading1-9, Title, Subtitle, Quote, etc.)

### Architecture
- **Domain-Driven Design** - Clean separation of concerns (domain, internal, pkg)
- **Error Handling** - Comprehensive error types with context and validation
- **Type Safety** - Strong typing throughout, no `interface{}`
- **Thread Safety** - Concurrent access supported with RWMutex
- **Memory Efficient** - Streaming ZIP writer, optimized serialization

### Developer Experience
- **Two APIs** - Simple API (direct interfaces) and Builder API (fluent, chainable)
- **11 Working Examples** - Comprehensive examples covering all major features
- **Excellent Error Handling** - Rated EXCELLENT in Phase 11 review
- **50.7% Test Coverage** - Comprehensive unit tests (improvement plan ready → 95%)
- **Zero Linter Warnings** - golangci-lint with 30+ linters

---

## 📦 Installation

### Using go get

```bash
go get github.com/mmonterroca/docxgo@v2.0.0
```

### Import in your code

```go
import (
    docx "github.com/mmonterroca/docxgo"
    "github.com/mmonterroca/docxgo/domain"
)
```

---

## 🚀 Quick Start

### Option 1: Simple API (Direct Domain Interfaces)

```go
package main

import (
    "log"
    docx "github.com/mmonterroca/docxgo"
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
    docx "github.com/mmonterroca/docxgo"
    "github.com/mmonterroca/docxgo/domain"
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
    docx "github.com/mmonterroca/docxgo"
)

func main() {
    // Open existing document
    doc, err := docx.OpenDocument("template.docx")
    if err != nil {
        log.Fatal(err)
    }
    
    // Modify existing content
    paragraphs := doc.Paragraphs()
    for _, para := range paragraphs {
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
    newRun.SetText("Added by the reader")
    
    // Save modified document
    if err := doc.SaveAs("modified.docx"); err != nil {
        log.Fatal(err)
    }
}
```

---

## 📖 Examples

See [examples/](examples/) directory for 11 comprehensive examples:

| Example | Description | Key Features |
|---------|-------------|--------------|
| `01_basic` | Simple document creation | Document, paragraph, run basics |
| `02_intermediate` | Formatted content | Font styles, colors, alignment |
| `03_toc` | Table of Contents | TOC generation, headings, bookmarks |
| `04_fields` | Dynamic fields | Page numbers, dates, hyperlinks |
| `05_styles` | Paragraph styles | Built-in Word styles (40+ styles) |
| `06_sections` | Multi-section documents | Headers, footers, page setup |
| `07_advanced` | Complex layouts | Multiple sections, page breaks |
| `08_images` | Image insertion | 9 formats, inline/floating positioning |
| `09_advanced_tables` | Table features | Cell merging, nested tables, 8 styles |
| `11_multi_section` | Section management | Different layouts per section |
| `12_read_and_modify` 🆕 | Document reading | Open, read, modify, save workflow |

### Running Examples

```bash
cd examples/01_basic
go run main.go

# Or run all examples at once
cd examples
bash run_all_examples.sh
```

All examples generate valid .docx files that open correctly in Microsoft Word, LibreOffice, and Google Docs.

---

## ⚠️ Breaking Changes from v1.x

This is a **major version update** with complete API redesign. v1.x code will not work with v2.0.0.

### Key Differences

| v1.x | v2.0.0 |
|------|--------|
| Concrete types | Interface-based design |
| Silent failures | Explicit error returns |
| Global state | Dependency injection |
| No validation | Full input validation |
| Limited features | Comprehensive OOXML support |
| No error context | Rich error context |

### Migration Resources

- **[MIGRATION.md](MIGRATION.md)** - Step-by-step migration guide
- **[docs/V2_API_GUIDE.md](docs/V2_API_GUIDE.md)** - Complete API reference
- **[examples/](examples/)** - 11 working examples
- **[docs/ERROR_HANDLING.md](docs/ERROR_HANDLING.md)** - Error handling patterns

---

## 🔗 Compatibility

- **Go Version**: 1.23+
- **OOXML**: Office Open XML (ISO/IEC 29500)
- **Microsoft Word**: 2007+ (Windows/Mac)
- **LibreOffice**: 6.0+ (all platforms)
- **Google Docs**: Full compatibility
- **Operating Systems**: Linux, macOS, Windows

---

## 📊 Test Coverage & Quality

- ✅ **50.7% Test Coverage** - Comprehensive unit tests with improvement plan ready
- ✅ **0 Linter Warnings** - golangci-lint with 30+ linters
- ✅ **100% Valid OOXML** - All generated documents pass strict schema validation
- ✅ **11/11 Examples Working** - All examples generate valid documents
- ✅ **Excellent Error Handling** - Rated EXCELLENT in Phase 11 review

### CI/CD

- GitHub Actions with automated testing
- golangci-lint v2.5.0 for code quality
- Strict linting for production code
- Example validation against Word

---

## 🐛 Known Issues & Limitations

### TOC Field Update Required
- **Issue**: Table of Contents requires manual "Update Field" (F9) in Word
- **Cause**: Inherent OOXML limitation - TOC content is generated by Word, not the library
- **Workaround**: Press F9 in Word after opening the document
- **Status**: Cannot be fixed in library (Word behavior)

### Complex Table Borders
- **Issue**: Some advanced border configurations may require fine-tuning
- **Impact**: Borders appear but might not match exact expectations
- **Workaround**: Use built-in table styles or adjust border settings
- **Status**: Will improve in future versions

### Advanced Features Not Yet Implemented
- Comments and change tracking (planned for v2.1.0)
- Custom XML parts (planned for v2.1.0)
- Advanced drawing shapes (planned for v2.2.0)
- Mail merge and templates (planned for v2.3.0)
- Document comparison (planned for v2.3.0)
- Content controls (planned for v2.4.0)

See [docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md) for complete roadmap.

---

## 🗺️ Roadmap

### v2.1.0 (Q1 2026) - Enhanced Reading
- Complete Phase 10 (Document Reading to 100%)
- Read headers/footers from existing documents
- Read images and complex tables
- Read document properties and metadata
- Comments and change tracking

### v2.2.0 (Q2 2026) - Advanced Content
- Custom XML parts
- Advanced drawing shapes
- Enhanced image manipulation
- Content controls

### v2.3.0 (Q3 2026) - Automation
- Mail merge support
- Document templates
- Document comparison
- Batch processing utilities

See [docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md) for detailed roadmap.

---

## 📚 Documentation

### Main Documentation
- **[README.md](README.md)** - Project overview and quick start
- **[MIGRATION.md](MIGRATION.md)** - v1.x → v2.0.0 migration guide
- **[CHANGELOG.md](CHANGELOG.md)** - Complete version history

### Technical Documentation
- **[docs/V2_DESIGN.md](docs/V2_DESIGN.md)** - Architecture and design decisions
- **[docs/V2_API_GUIDE.md](docs/V2_API_GUIDE.md)** - Complete API reference
- **[docs/ERROR_HANDLING.md](docs/ERROR_HANDLING.md)** - Error handling patterns
- **[docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md)** - Feature status and roadmap

### Examples & Guides
- **[examples/README.md](examples/README.md)** - Example overview and usage
- **[examples/v2_README.md](examples/v2_README.md)** - v2-specific examples

---

## 🙏 Credits

**Based on**: [fumiama/go-docx](https://github.com/fumiama/go-docx) library

**v2.0.0 Complete Rewrite**: [@mmonterroca](https://github.com/mmonterroca) and contributors

**Contributors**: See [CONTRIBUTORS](CONTRIBUTORS) file

**Special Thanks**: All beta testers and early adopters who provided feedback

---

## 📄 License

**MIT License** - See [LICENSE](LICENSE) file for details.

You are free to:
- ✅ Use commercially
- ✅ Modify
- ✅ Distribute
- ✅ Private use

---

## 🔧 Support & Community

### Getting Help
- 📖 **Documentation**: Start with [README.md](README.md) and [docs/](docs/)
- 💬 **Issues**: Report bugs on [GitHub Issues](https://github.com/mmonterroca/docxgo/issues)
- 📧 **Questions**: Open a discussion on [GitHub Discussions](https://github.com/mmonterroca/docxgo/discussions)

### Contributing
We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Changelog
Full changelog: [v2.0.0-beta...v2.0.0](https://github.com/mmonterroca/docxgo/compare/v2.0.0-beta...v2.0.0)

---

## 🎉 Thank You!

Thank you for using go-docx v2.0.0! We're excited to see what you build with it.

If you find this library useful, please:
- ⭐ Star the repository on GitHub
- 📣 Share with your colleagues
- 🐛 Report issues you encounter
- 💡 Suggest features you'd like to see

Happy coding! 🚀

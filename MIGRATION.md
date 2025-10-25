# Migration Guide: v1 → v2

This guide helps you migrate from legacy v1 to the new v2 architecture.

## Table of Contents

- [Overview](#overview)
- [Should You Migrate?](#should-you-migrate)
- [Breaking Changes](#breaking-changes)
- [Step-by-Step Migration](#step-by-step-migration)
- [API Comparison](#api-comparison)
- [Common Patterns](#common-patterns)
- [Troubleshooting](#troubleshooting)

---

## Overview

v2 is a **complete architectural rewrite** of go-docx. While this means breaking changes, it also brings:

- ✅ **Explicit error handling** - No more silent failures
- ✅ **Type safety** - No more `interface{}`
- ✅ **Better testability** - Interface-based design
- ✅ **Thread safety** - Concurrent access supported
- ✅ **Performance** - 10%+ faster in most cases

**Migration Difficulty**: Moderate  
**Estimated Time**: 1-4 hours (depending on codebase size)

---

## Should You Migrate?

### ✅ Migrate If:

- Starting a new project (use v2 from the start)
- Need better error handling
- Want to write tests (v2 is much easier to test)
- Working on long-term maintained code
- Need thread-safe document manipulation
- Want modern Go best practices

### ⏸️ Wait If:

- v2 feature you need isn't implemented yet (check [roadmap](docs/V2_DESIGN.md))
- You can't test in your environment
- Production code with tight deadline (wait for v2.0.0 stable - Q1 2026)

### ⚠️ Stay on v1 If:

- Legacy code with no maintenance budget
- v2 will never have the feature you need
- Willing to maintain your own fork

**Note**: v1 will receive critical bug fixes through December 2025 only.

---

## Breaking Changes

### 1. **Import Paths**

```go
// v1
import "github.com/fumiama/go-docx"
import docx "github.com/SlideLang/go-docx/legacy/v1"  // If using legacy

// v2
import "github.com/SlideLang/go-docx/domain"
import "github.com/SlideLang/go-docx/internal/core"
```

### 2. **Error Handling**

```go
// v1 - No errors
para := doc.AddParagraph()
run := para.AddText("Hello")
run.Bold().Color("FF0000") // Silent failure on invalid color

// v2 - Explicit errors
para, err := doc.AddParagraph()
if err != nil {
    return err
}
run, err := para.AddRun()
if err != nil {
    return err
}
run.SetText("Hello")
run.SetBold(true)
run.SetColor(domain.Color{R: 255, G: 0, B: 0}) // Type-safe
```

### 3. **API Structure**

```go
// v1 - Fluent API
doc := docx.New()
para := doc.AddParagraph()
para.AddText("Text").Bold().Size("28").Color("FF0000")

// v2 - Explicit API with error handling
doc := core.NewDocument()
para, err := doc.AddParagraph()
if err != nil {
    return err
}
run, err := para.AddRun()
if err != nil {
    return err
}
run.SetText("Text")
run.SetBold(true)
run.SetSize(56) // Half-points (28pt = 56 half-points)
run.SetColor(domain.Color{R: 255, G: 0, B: 0})
```

### 4. **Method Names**

| v1 Method | v2 Method | Notes |
|-----------|-----------|-------|
| `AddText(text)` | `AddRun()` + `SetText(text)` | Two-step process |
| `Bold()` | `SetBold(true)` | Explicit setter |
| `Size(string)` | `SetSize(int)` | Half-points, not string |
| `Color(string)` | `SetColor(Color)` | Type-safe struct |
| `Alignment(string)` | `SetAlignment(Alignment)` | Type-safe constant |

### 5. **Value Types**

```go
// v1 - Strings everywhere
run.Size("28")           // String
run.Color("FF0000")      // Hex string
para.Alignment("center") // String

// v2 - Type-safe
run.SetSize(56)          // int (half-points)
run.SetColor(domain.Color{R: 255, G: 0, B: 0}) // Struct
para.SetAlignment(constants.AlignmentCenter)   // Constant
```

---

## Step-by-Step Migration

### Step 1: Update Dependencies

```bash
# Remove v1
go get github.com/fumiama/go-docx@none

# Add v2
go get github.com/SlideLang/go-docx@latest
```

### Step 2: Update Imports

```go
// Old
import "github.com/fumiama/go-docx"

// New
import (
    "github.com/SlideLang/go-docx/domain"
    "github.com/SlideLang/go-docx/internal/core"
    "github.com/SlideLang/go-docx/pkg/constants"
)
```

### Step 3: Update Document Creation

```go
// v1
doc := docx.New()
// or
doc := docx.New().WithDefaultTheme()

// v2
doc := core.NewDocument()
```

### Step 4: Update Text Operations

```go
// v1
para := doc.AddParagraph()
run := para.AddText("Hello World")
run.Bold().Size("28").Color("FF0000")

// v2
para, err := doc.AddParagraph()
if err != nil {
    return err
}
run, err := para.AddRun()
if err != nil {
    return err
}
if err := run.SetText("Hello World"); err != nil {
    return err
}
if err := run.SetBold(true); err != nil {
    return err
}
if err := run.SetSize(56); err != nil { // 28pt * 2
    return err
}
if err := run.SetColor(domain.Color{R: 255, G: 0, B: 0}); err != nil {
    return err
}
```

### Step 5: Update File I/O

```go
// v1
f, _ := os.Create("output.docx")
doc.WriteTo(f)
f.Close()

// v2
if err := doc.SaveAs("output.docx"); err != nil {
    return err
}
```

### Step 6: Test Thoroughly

```bash
# Run your tests
go test ./...

# Test with real documents
# Open generated .docx files in Microsoft Word to verify
```

---

## API Comparison

### Document Creation

```go
// v1
doc := docx.New()
doc := docx.New().WithDefaultTheme()

// v2
doc := core.NewDocument()
```

### Adding Paragraphs

```go
// v1
para := doc.AddParagraph()

// v2
para, err := doc.AddParagraph()
if err != nil {
    return err
}
```

### Adding Text

```go
// v1
run := para.AddText("Hello")
run.Bold().Italic().Underline()
run.Size("24").Color("0000FF")

// v2
run, err := para.AddRun()
if err != nil {
    return err
}
run.SetText("Hello")
run.SetBold(true)
run.SetItalic(true)
run.SetUnderline(constants.UnderlineSingle)
run.SetSize(48) // 24pt * 2
run.SetColor(domain.Color{R: 0, G: 0, B: 255})
```

### Adding Tables

```go
// v1
table := doc.AddTable()
table.AddRow(3) // 3 columns
row := table.Rows[0]
cell := row.Cells[0]
cell.AddText("Cell content")

// v2
table, err := doc.AddTable(1, 3) // 1 row, 3 columns
if err != nil {
    return err
}
row, err := table.Row(0)
if err != nil {
    return err
}
cell, err := row.Cell(0)
if err != nil {
    return err
}
cellPara, err := cell.AddParagraph()
if err != nil {
    return err
}
cellRun, err := cellPara.AddRun()
if err != nil {
    return err
}
cellRun.SetText("Cell content")
```

### Alignment

```go
// v1
para.Alignment("center")

// v2
para.SetAlignment(constants.AlignmentCenter)
```

### Indentation

```go
// v1
para.Indent(720, 0, 0) // left, right, firstLine (in twips)

// v2
para.SetIndentation(&domain.Indentation{
    Left:      720,
    Right:     0,
    FirstLine: 0,
})
```

### Colors

```go
// v1
run.Color("FF0000") // Hex string

// v2
run.SetColor(domain.Color{R: 255, G: 0, B: 0})

// Or use predefined colors
import "github.com/SlideLang/go-docx/pkg/color"
run.SetColor(color.Red)
run.SetColor(color.Blue)
```

### Font Size

```go
// v1
run.Size("28") // String, in points

// v2
run.SetSize(56) // Int, in half-points (28pt = 56 half-points)
```

---

## Common Patterns

### Pattern 1: Creating a Simple Document

**v1:**
```go
func createDocV1() error {
    doc := docx.New()
    para := doc.AddParagraph()
    para.AddText("Hello World").Bold().Size("24")
    
    f, _ := os.Create("output.docx")
    defer f.Close()
    doc.WriteTo(f)
    return nil
}
```

**v2:**
```go
func createDocV2() error {
    doc := core.NewDocument()
    
    para, err := doc.AddParagraph()
    if err != nil {
        return err
    }
    
    run, err := para.AddRun()
    if err != nil {
        return err
    }
    
    if err := run.SetText("Hello World"); err != nil {
        return err
    }
    if err := run.SetBold(true); err != nil {
        return err
    }
    if err := run.SetSize(48); err != nil { // 24pt * 2
        return err
    }
    
    return doc.SaveAs("output.docx")
}
```

### Pattern 2: Error Handling Helper

To reduce boilerplate, create helper functions:

```go
// Helper function for v2
func mustAddPara(doc domain.Document) domain.Paragraph {
    para, err := doc.AddParagraph()
    if err != nil {
        panic(err) // Or handle appropriately
    }
    return para
}

func mustAddRun(para domain.Paragraph) domain.Run {
    run, err := para.AddRun()
    if err != nil {
        panic(err)
    }
    return run
}

// Usage
doc := core.NewDocument()
run := mustAddRun(mustAddPara(doc))
run.SetText("Hello")
```

### Pattern 3: Builder Pattern (Coming in v2.1)

v2 will eventually support a builder pattern similar to v1:

```go
// Coming soon in v2.1
doc := docx.NewBuilder()
doc.AddParagraph().
    Text("Hello").
    Bold().
    FontSize(14).
    End()

finalDoc, err := doc.Build()
```

---

## Troubleshooting

### Issue: "Too much error handling"

**Solution**: Use helper functions or wait for builder pattern (v2.1)

```go
// Helper pattern
func addText(para domain.Paragraph, text string, bold bool, size int) error {
    run, err := para.AddRun()
    if err != nil {
        return err
    }
    if err := run.SetText(text); err != nil {
        return err
    }
    if bold {
        if err := run.SetBold(true); err != nil {
            return err
        }
    }
    if size > 0 {
        if err := run.SetSize(size); err != nil {
            return err
        }
    }
    return nil
}
```

### Issue: "Method not found"

**Solution**: Check the API comparison table. Method names changed.

```go
// v1
para.AddText("text")

// v2
run, _ := para.AddRun()
run.SetText("text")
```

### Issue: "Invalid size value"

**Solution**: v2 uses half-points, not points

```go
// v1
run.Size("12") // 12pt

// v2
run.SetSize(24) // 12pt * 2 = 24 half-points
```

### Issue: "Color not working"

**Solution**: v2 uses RGB struct, not hex string

```go
// v1
run.Color("FF0000")

// v2
run.SetColor(domain.Color{R: 255, G: 0, B: 0})
```

### Issue: "Feature missing in v2"

**Solution**: Check the [roadmap](docs/V2_DESIGN.md). Some features are still in development:

- ✅ Basic text and formatting - Complete
- ✅ Tables - Complete
- ✅ Indentation - Complete
- 🚧 Headers/Footers - In progress
- 🚧 TOC - In progress
- 📋 Images - Planned
- 📋 Styles - Planned

If you need a feature that's not ready, you can:
1. Wait for the feature (check roadmap for timeline)
2. Contribute the feature (see [CONTRIBUTING.md](CONTRIBUTING.md))
3. Stay on v1 temporarily

---

## Advanced Features Migration (Phase 6)

### Sections and Page Layout

#### v1 - Limited section support
```go
// v1 had basic page size methods
doc.WithA4Page()
doc.WithLetterPage()
// No direct section access
```

#### v2 - Full section control
```go
// v2 provides complete section management
section, err := doc.DefaultSection()
if err != nil {
    return err
}

// Custom page size
section.SetPageSize(domain.PageSize{
    Width:  12240,  // 8.5 inches in twips
    Height: 15840,  // 11 inches
})

// Or use predefined sizes
section.SetPageSize(domain.PageSizeA4)
section.SetPageSize(domain.PageSizeLetter)
section.SetPageSize(domain.PageSizeLegal)

// Margins (in twips: 1440 = 1 inch)
margins := domain.Margins{
    Top:    1440,
    Right:  1440,
    Bottom: 1440,
    Left:   1440,
    Header: 720,
    Footer: 720,
}
section.SetMargins(margins)

// Orientation
section.SetOrientation(domain.OrientationLandscape)
section.SetOrientation(domain.OrientationPortrait)

// Columns
section.SetColumns(2) // Two-column layout
section.SetColumns(3) // Three-column layout
```

### Headers and Footers

#### v1 - Basic headers/footers
```go
// v1 had limited header/footer support
// Methods varied by implementation
```

#### v2 - Comprehensive header/footer system
```go
// Get section
section, _ := doc.DefaultSection()

// Access headers by type
headerDefault, _ := section.Header(domain.HeaderDefault)
headerFirst, _ := section.Header(domain.HeaderFirst)
headerEven, _ := section.Header(domain.HeaderEven)

// Access footers by type
footerDefault, _ := section.Footer(domain.FooterDefault)
footerFirst, _ := section.Footer(domain.FooterFirst)
footerEven, _ := section.Footer(domain.FooterEven)

// Add content to header
para, _ := headerDefault.AddParagraph()
para.SetAlignment(domain.AlignmentRight)
run, _ := para.AddRun()
run.AddText("Document Title")
run.SetBold(true)

// Add page numbers to footer
footerPara, _ := footerDefault.AddParagraph()
footerPara.SetAlignment(domain.AlignmentCenter)

// "Page X of Y" format
r1, _ := footerPara.AddRun()
r1.AddText("Page ")

r2, _ := footerPara.AddRun()
pageField := docx.NewPageNumberField()
r2.AddField(pageField)

r3, _ := footerPara.AddRun()
r3.AddText(" of ")

r4, _ := footerPara.AddRun()
totalField := docx.NewPageCountField()
r4.AddField(totalField)
```

### Fields

#### v1 - Limited field support
```go
// v1 had basic TOC support
doc.AddTOC()

// Page numbers were manual
para.AddText("Page 1")
```

#### v2 - Complete field system
```go
// Page numbers
pageField := docx.NewPageNumberField()
run.AddField(pageField)

totalPages := docx.NewPageCountField()
run.AddField(totalPages)

// Table of Contents with options
tocOptions := map[string]string{
    "levels":          "1-5",    // Heading levels
    "hyperlinks":      "true",   // Enable hyperlinks
    "hidePageNumbers": "false",  // Show page numbers
}
tocField := docx.NewTOCField(tocOptions)
tocRun.AddField(tocField)

// Hyperlinks
linkField := docx.NewHyperlinkField(
    "https://github.com/SlideLang/go-docx",
    "go-docx Repository",
)
linkRun.SetColor(docx.ColorBlue)
linkRun.SetUnderline(docx.UnderlineSingle)
linkRun.AddField(linkField)

// Style references (for running headers)
styleRef := docx.NewStyleRefField("Heading 1")
run.AddField(styleRef)

// Custom field codes
customField := docx.NewField(docx.FieldTypeCustom)
customField.SetCode(`AUTHOR \* Upper`)
customField.Update()
run.AddField(customField)
```

### Styles

#### v1 - String-based styles
```go
// v1 used string style names
para.Style("Heading1")
para.Style("Normal")
// No validation, no discovery
```

#### v2 - Complete style management
```go
// Built-in style constants (type-safe)
para.SetStyle(domain.StyleIDHeading1)
para.SetStyle(domain.StyleIDNormal)
para.SetStyle(domain.StyleIDQuote)

// Access style manager
styleMgr := doc.StyleManager()

// Check if style exists
if styleMgr.HasStyle("Heading1") {
    para.SetStyle("Heading1")
}

// Create custom style
customStyle := &manager.ParagraphStyle{}
customStyle.SetAlignment(domain.AlignmentCenter)
styleMgr.AddStyle(customStyle)
```

### Migration Checklist for Phase 6 Features

- [ ] Replace basic page setup with Section configuration
- [ ] Migrate headers/footers to new Section-based API
- [ ] Convert manual page numbers to Field system
- [ ] Update TOC generation to use TOCField
- [ ] Replace hardcoded hyperlinks with HyperlinkField
- [ ] Convert string style names to StyleID constants
- [ ] Use StyleManager for style queries
- [ ] Add error handling for all new APIs
- [ ] Test field updates (press F9 in Word)

---

## Getting Help

### Resources

- **Documentation**: [API Reference](https://pkg.go.dev/github.com/SlideLang/go-docx)
- **Examples**: [`examples/`](examples/) directory
- **Design Doc**: [V2_DESIGN.md](docs/V2_DESIGN.md)

### Support Channels

- **Issues**: [GitHub Issues](https://github.com/SlideLang/go-docx/issues) - For bugs
- **Discussions**: [GitHub Discussions](https://github.com/SlideLang/go-docx/discussions) - For questions
- **Email**: misael@monterroca.com - For complex migration scenarios

### Before Asking for Help

Please include:
1. v1 code snippet
2. v2 code attempt
3. Error message (if any)
4. What you've tried

---

## Timeline

| Date | Event |
|------|-------|
| Oct 2025 | v2.0.0-alpha released, migration guide published |
| Dec 2025 | v2.0.0-beta, most v1 features available |
| Jan 2026 | v2.0.0-rc, API frozen |
| Mar 2026 | v2.0.0 stable, recommended for production |
| Dec 2025 | v1 critical bug fixes end |
| Mar 2026 | v1 fully deprecated (no support) |

---

## FAQ

**Q: Can I use v1 and v2 in the same project?**  
A: Yes, but they use different import paths. Be careful with namespace collisions.

**Q: Will v2 read v1-created .docx files?**  
A: Yes, v2 uses the same OOXML standard.

**Q: Is there an automated migration tool?**  
A: Not yet. Given the architectural differences, manual migration is recommended.

**Q: What if I find a bug during migration?**  
A: Please report it on [GitHub Issues](https://github.com/SlideLang/go-docx/issues) with the `migration` label.

**Q: Can I contribute to v2?**  
A: Absolutely! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

*Last Updated: October 25, 2025*  
*For the latest information, see: https://github.com/SlideLang/go-docx*

# go-docx v2 API Guide

**Version**: 2.0.0-beta  
**Last Updated**: October 27, 2025

---

## 📖 Table of Contents

- [Introduction](#introduction)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [API Patterns](#api-patterns)
  - [Builder Pattern](#builder-pattern)
  - [Direct Domain API](#direct-domain-api)
- [Core Features](#core-features)
  - [Document Creation](#document-creation)
  - [Paragraphs and Text](#paragraphs-and-text)
  - [Tables](#tables)
  - [Images](#images)
  - [Fields](#fields)
  - [Sections and Page Layout](#sections-and-page-layout)
  - [Styles](#styles)
- [Examples](#examples)
- [Migration from v1](#migration-from-v1)

---

## 🎯 Introduction

**go-docx v2** is a complete rewrite with clean architecture, proper error handling, and comprehensive OOXML support. Key improvements:

✅ **Two API Styles**:
- **Builder Pattern**: Fluent API with error accumulation (recommended for new code)
- **Direct Domain API**: Interface-based design for advanced use cases

✅ **Complete Feature Set**:
- Paragraphs, runs, tables, images
- Fields (PAGE, NUMPAGES, TOC, HYPERLINK, etc.)
- Sections with headers/footers
- 40+ built-in styles
- Advanced table features (cell merging, nested tables)

✅ **Better Quality**:
- Type-safe API (minimal `interface{}`)
- Comprehensive error handling
- Thread-safe operations
- 95%+ test coverage

---

## 📦 Installation

```bash
go get github.com/mmonterroca/docxgo/v2@latest
```

**Minimum Go version**: 1.20

---

## 🚀 Quick Start

### Builder Pattern (Recommended)

```go
package main

import (
    "log"
    docx "github.com/mmonterroca/docxgo/v2"
)

func main() {
    // Create builder with options
    builder := docx.NewDocumentBuilder(
        docx.WithTitle("My Document"),
        docx.WithAuthor("John Doe"),
    )
    
    // Add content with fluent API
    builder.AddParagraph().
        Text("Hello, World!").
        Bold().
        FontSize(14).
        Color(docx.Red).
        Alignment(docx.AlignmentCenter).
        End()
    
    // Build and validate
    doc, err := builder.Build()
    if err != nil {
        log.Fatal(err)
    }
    
    // Save
    if err := doc.SaveAs("hello.docx"); err != nil {
        log.Fatal(err)
    }
}
```

### Direct Domain API

```go
package main

import (
    "log"
    docx "github.com/mmonterroca/docxgo/v2"
)

func main() {
    // Create document
    doc := docx.NewDocument()
    
    // Add paragraph
    para, err := doc.AddParagraph()
    if err != nil {
        log.Fatal(err)
    }
    
    // Add run with text
    run, err := para.AddRun()
    if err != nil {
        log.Fatal(err)
    }
    
    run.SetText("Hello, World!")
    run.SetBold(true)
    run.SetSize(28) // 14pt (size is in half-points)
    run.SetColor(docx.Red)
    
    // Save
    if err := doc.SaveAs("hello.docx"); err != nil {
        log.Fatal(err)
    }
}
```

---

## 🎨 API Patterns

### Builder Pattern

The builder pattern accumulates errors and returns them in `Build()`. This allows for fluent chaining without checking errors at every step.

**Benefits**:
- Cleaner code with method chaining
- Error accumulation (all errors collected and returned in `Build()`)
- Better for sequential document construction

**Example**:

```go
builder := docx.NewDocumentBuilder()

// Errors are accumulated internally
builder.AddParagraph().
    Text("Paragraph 1").
    Bold().
    End().
    AddParagraph().
    Text("Paragraph 2").
    Italic().
    End()

// All errors are returned here
doc, err := builder.Build()
if err != nil {
    // Handle accumulated errors
    log.Fatal(err)
}
```

### Direct Domain API

The direct API returns errors immediately, giving you fine-grained control.

**Benefits**:
- Explicit error handling at each step
- Better for conditional logic
- More control over document construction

**Example**:

```go
doc := docx.NewDocument()

para, err := doc.AddParagraph()
if err != nil {
    return err
}

run, err := para.AddRun()
if err != nil {
    return err
}

if err := run.SetText("Hello"); err != nil {
    return err
}
```

---

## 🔧 Core Features

### Document Creation

#### With Builder Pattern

```go
builder := docx.NewDocumentBuilder(
    docx.WithTitle("Annual Report 2025"),
    docx.WithAuthor("Jane Smith"),
    docx.WithSubject("Financial Data"),
    docx.WithKeywords("finance, report, 2025"),
    docx.WithDefaultFont("Calibri"),
)

doc, err := builder.Build()
```

#### With Direct API

```go
doc := docx.NewDocument()

metadata := &domain.Metadata{
    Title:    "Annual Report 2025",
    Author:   "Jane Smith",
    Subject:  "Financial Data",
    Keywords: "finance, report, 2025",
}

err := doc.SetMetadata(metadata)
```

---

### Paragraphs and Text

#### Builder Pattern

```go
builder.AddParagraph().
    Text("This is ").
    Text("bold text").Bold().
    Text(" and this is ").
    Text("italic text").Italic().
    Text(" and ").
    Text("red text").Color(docx.Red).
    End()
```

#### Direct API

```go
para, _ := doc.AddParagraph()

run1, _ := para.AddRun()
run1.SetText("This is ")

run2, _ := para.AddRun()
run2.SetText("bold text")
run2.SetBold(true)

run3, _ := para.AddRun()
run3.SetText(" and this is ")

run4, _ := para.AddRun()
run4.SetText("italic text")
run4.SetItalic(true)
```

#### Paragraph Alignment

```go
// Builder
builder.AddParagraph().
    Text("Centered text").
    Alignment(docx.AlignmentCenter).
    End()

// Direct API
para, _ := doc.AddParagraph()
para.SetAlignment(docx.AlignmentCenter)
run, _ := para.AddRun()
run.SetText("Centered text")
```

#### Text Formatting

Available methods (both APIs):

```go
// Font size (in points for builder, half-points for direct API)
.FontSize(14)          // Builder: 14pt
.SetSize(28)          // Direct API: 28 half-points = 14pt

// Colors (use predefined constants or domain.Color)
.Color(docx.Red)
.SetColor(docx.Blue)

// Styles
.Bold()
.SetBold(true)

.Italic()
.SetItalic(true)

.Underline(docx.UnderlineSingle)
.SetUnderline(domain.UnderlineSingle)
```

**Predefined Colors**:
- `docx.Black`, `docx.White`
- `docx.Red`, `docx.Green`, `docx.Blue`
- `docx.Yellow`, `docx.Cyan`, `docx.Magenta`
- `docx.Orange`, `docx.Purple`
- `docx.Gray`, `docx.Silver`

---

### Tables

#### Builder Pattern

```go
builder.AddTable(3, 3).
    Row(0).Cell(0).Text("Header 1").Bold().End().
    Row(0).Cell(1).Text("Header 2").Bold().End().
    Row(0).Cell(2).Text("Header 3").Bold().End().
    Row(1).Cell(0).Text("Data 1").End().
    Row(1).Cell(1).Text("Data 2").End().
    Row(1).Cell(2).Text("Data 3").End().
    End()
```

#### Direct API

```go
table, _ := doc.AddTable(3, 3)

// Access rows and cells
row0, _ := table.Row(0)
cell00, _ := row0.Cell(0)

// Add content to cell
para, _ := cell00.AddParagraph()
run, _ := para.AddRun()
run.SetText("Header 1")
run.SetBold(true)
```

#### Advanced Table Features

**Cell Merging**:

```go
// Builder
builder.AddTable(3, 3).
    Row(0).Cell(0).
        Text("Merged Cell").
        Merge(2, 1). // colspan=2, rowspan=1
        End().
    End()

// Direct API
table, _ := doc.AddTable(3, 3)
row, _ := table.Row(0)
cell, _ := row.Cell(0)
cell.Merge(2, 1) // Merge 2 columns, 1 row
```

**Nested Tables**:

```go
// Direct API
outerTable, _ := doc.AddTable(2, 2)
row, _ := outerTable.Row(0)
cell, _ := row.Cell(0)

// Add nested table inside cell
nestedTable, _ := cell.AddTable(2, 2)
```

**Table Styles**:

```go
// Builder
builder.AddTable(3, 3).
    Style(domain.TableStyleGrid).
    End()

// Direct API
table, _ := doc.AddTable(3, 3)
table.SetStyle(domain.TableStyleGrid)
```

Predefined table styles:
- `TableStyleGrid`
- `TableStyleList`
- `TableStyleColorful`
- `TableStyleAccent1` through `TableStyleAccent6`

---

### Images

#### Builder Pattern

```go
// Simple image (default size)
builder.AddParagraph().
    AddImage("path/to/image.png").
    End()

// Custom size
builder.AddParagraph().
    AddImageWithSize("logo.png", domain.ImageSize{
        Width:  914400 * 2, // 2 inches in EMUs
        Height: 914400 * 1, // 1 inch in EMUs
    }).
    End()

// Floating image with position
builder.AddParagraph().
    AddImageWithPosition("diagram.png",
        domain.ImageSize{Width: 914400 * 4, Height: 914400 * 3},
        domain.ImagePosition{
            IsInline: false,
            HOffset:  914400,   // 1 inch from left
            VOffset:  914400,   // 1 inch from top
        }).
    End()
```

#### Direct API

```go
para, _ := doc.AddParagraph()

// Simple image
img, _ := para.AddImage("path/to/image.png")

// Custom size
img, _ := para.AddImageWithSize("logo.png", domain.ImageSize{
    Width:  914400 * 2,
    Height: 914400 * 1,
})

// Floating image
img, _ := para.AddImageWithPosition("diagram.png",
    domain.ImageSize{Width: 914400 * 4, Height: 914400 * 3},
    domain.ImagePosition{
        IsInline: false,
        HOffset:  914400,
        VOffset:  914400,
    })
```

**Supported Formats**:
- PNG, JPEG, GIF, BMP
- TIFF, SVG, WEBP
- ICO, EMF

**Size Units**:
- **EMUs** (English Metric Units): 914400 EMUs = 1 inch
- **Pixels** to EMUs: `pixels * 9525`
- **Inches** to EMUs: `inches * 914400`

---

### Fields

Fields are dynamic elements that Word updates automatically.

#### Available Field Types

```go
// Page number
docx.NewPageNumberField()

// Total page count
docx.NewPageCountField()

// Table of Contents
docx.NewTOCField(map[string]string{
    "levels":          "1-3",
    "hyperlinks":      "true",
    "hidePageNumbers": "false",
})

// Hyperlink
docx.NewHyperlinkField("https://example.com", "Click here")

// StyleRef (references heading text)
docx.NewStyleRefField("Heading1")

// Sequence (auto-numbering)
docx.NewSeqField("Figure", "ARABIC")

// Reference to bookmark
docx.NewRefField("MyBookmark")

// Page reference to bookmark
docx.NewPageRefField("_Toc000000001")
```

#### Using Fields

**Direct API**:

```go
para, _ := doc.AddParagraph()

run1, _ := para.AddRun()
run1.SetText("Page ")

run2, _ := para.AddRun()
run2.AddField(docx.NewPageNumberField())

run3, _ := para.AddRun()
run3.SetText(" of ")

run4, _ := para.AddRun()
run4.AddField(docx.NewPageCountField())
```

**In Headers/Footers**:

```go
section, _ := doc.DefaultSection()
footer, _ := section.Footer(domain.FooterDefault)

para, _ := footer.AddParagraph()
para.SetAlignment(domain.AlignmentCenter)

run1, _ := para.AddRun()
run1.SetText("Page ")

run2, _ := para.AddRun()
run2.AddField(docx.NewPageNumberField())

run3, _ := para.AddRun()
run3.SetText(" of ")

run4, _ := para.AddRun()
run4.AddField(docx.NewPageCountField())
```

---

### Sections and Page Layout

Sections allow different page layouts within the same document.

#### Basic Section Configuration

```go
section, err := doc.DefaultSection()
if err != nil {
    log.Fatal(err)
}

// Set page size
section.SetPageSize(domain.PageSizeA4)

// Set margins
section.SetMargins(domain.Margins{
    Top:    1440, // 1 inch (1440 twips)
    Right:  1440,
    Bottom: 1440,
    Left:   1440,
    Header: 720,  // 0.5 inch from top
    Footer: 720,  // 0.5 inch from bottom
})

// Set orientation
section.SetOrientation(domain.OrientationLandscape)

// Set columns
section.SetColumns(2) // Two-column layout
```

#### Predefined Page Sizes

```go
domain.PageSizeA4       // 210mm x 297mm
domain.PageSizeLetter   // 8.5" x 11"
domain.PageSizeLegal    // 8.5" x 14"
domain.PageSizeA3       // 297mm x 420mm
domain.PageSizeTableid  // 11" x 17"
```

#### Headers and Footers

```go
section, _ := doc.DefaultSection()

// Get header
header, _ := section.Header(domain.HeaderDefault)

// Add content to header
para, _ := header.AddParagraph()
para.SetAlignment(domain.AlignmentRight)

run, _ := para.AddRun()
run.SetText("Company Name")
run.SetBold(true)

// Get footer
footer, _ := section.Footer(domain.FooterDefault)

// Add page numbers to footer
para, _ := footer.AddParagraph()
para.SetAlignment(domain.AlignmentCenter)

run1, _ := para.AddRun()
run1.SetText("Page ")

run2, _ := para.AddRun()
run2.AddField(docx.NewPageNumberField())

run3, _ := para.AddRun()
run3.SetText(" of ")

run4, _ := para.AddRun()
run4.AddField(docx.NewPageCountField())
```

**Header/Footer Types**:
- `domain.HeaderDefault` - Default header for all pages
- `domain.HeaderFirst` - Header for first page only
- `domain.HeaderEven` - Header for even pages
- `domain.FooterDefault` - Default footer for all pages
- `domain.FooterFirst` - Footer for first page only
- `domain.FooterEven` - Footer for even pages

---

### Styles

v2 includes 40+ built-in paragraph styles.

#### Applying Styles

```go
// Builder (via paragraph style - to be added)
// Currently use direct API for styles

// Direct API
para, _ := doc.AddParagraph()
para.SetStyle("Heading1")

run, _ := para.AddRun()
run.SetText("Chapter 1: Introduction")
```

#### Built-in Styles

**Headings**:
- `Heading1`, `Heading2`, `Heading3`, `Heading4`, `Heading5`, `Heading6`, `Heading7`, `Heading8`, `Heading9`

**Title Styles**:
- `Title`, `Subtitle`

**Text Styles**:
- `Normal` (default)
- `BodyText`, `BodyText2`, `BodyText3`
- `Quote`, `IntenseQuote`
- `Emphasis`, `Strong`, `IntenseEmphasis`

**List Styles**:
- `ListParagraph`
- `ListNumber`, `ListNumber2`, `ListNumber3`
- `ListBullet`, `ListBullet2`, `ListBullet3`

**Special Styles**:
- `NoSpacing`
- `Caption`
- `Footer`, `Header`

---

## 💡 Examples

### Complete Document with TOC

```go
builder := docx.NewDocumentBuilder(
    docx.WithTitle("Technical Documentation"),
    docx.WithAuthor("Engineering Team"),
)

// Add title
builder.AddParagraph().
    Text("Technical Documentation").
    FontSize(18).
    Bold().
    Alignment(docx.AlignmentCenter).
    End()

// Add TOC
doc, _ := builder.Build()
para, _ := doc.AddParagraph()
run, _ := para.AddRun()
run.AddField(docx.NewTOCField(map[string]string{
    "levels":          "1-3",
    "hyperlinks":      "true",
    "hidePageNumbers": "false",
}))

// Add sections with styled headings
para1, _ := doc.AddParagraph()
para1.SetStyle("Heading1")
run1, _ := para1.AddRun()
run1.SetText("1. Introduction")

para2, _ := doc.AddParagraph()
run2, _ := para2.AddRun()
run2.SetText("This document describes...")

// Save
doc.SaveAs("technical_doc.docx")
```

### Product Catalog with Tables

```go
builder := docx.NewDocumentBuilder()

builder.AddParagraph().
    Text("Product Catalog").
    FontSize(16).
    Bold().
    Alignment(docx.AlignmentCenter).
    End()

builder.AddTable(4, 3).
    Row(0).
        Cell(0).Text("Product").Bold().End().
        Cell(1).Text("Price").Bold().End().
        Cell(2).Text("Stock").Bold().End().
        End().
    Row(1).
        Cell(0).Text("Widget A").End().
        Cell(1).Text("$19.99").End().
        Cell(2).Text("100").End().
        End().
    Row(2).
        Cell(0).Text("Widget B").End().
        Cell(1).Text("$29.99").End().
        Cell(2).Text("50").End().
        End().
    Row(3).
        Cell(0).Text("Widget C").End().
        Cell(1).Text("$39.99").End().
        Cell(2).Text("25").End().
        End().
    End()

doc, _ := builder.Build()
doc.SaveAs("catalog.docx")
```

### Document with Images

```go
builder := docx.NewDocumentBuilder()

builder.AddParagraph().
    Text("Figure 1: System Architecture").
    Alignment(docx.AlignmentCenter).
    End()

builder.AddParagraph().
    AddImageWithSize("architecture.png", domain.ImageSize{
        Width:  914400 * 6,  // 6 inches
        Height: 914400 * 4,  // 4 inches
    }).
    End()

doc, _ := builder.Build()
doc.SaveAs("report.docx")
```

---

## 🔄 Migration from v1

### v1 API (Legacy - Pre-rewrite)

```go
import "github.com/fumiama/go-docx"

doc := docx.New()
para := doc.AddParagraph()
run := para.AddText("Hello")
run.Bold().Color("FF0000").Size("28")
doc.WriteTo(file)
```

### v2 API (Current - Builder Pattern)

```go
import docx "github.com/mmonterroca/docxgo/v2"

builder := docx.NewDocumentBuilder()
builder.AddParagraph().
    Text("Hello").
    Bold().
    Color(docx.Red).
    FontSize(14).
    End()

doc, _ := builder.Build()
doc.SaveAs("output.docx")
```

### v2 API (Current - Direct API)

```go
import docx "github.com/mmonterroca/docxgo/v2"

doc := docx.NewDocument()
para, _ := doc.AddParagraph()
run, _ := para.AddRun()
run.SetText("Hello")
run.SetBold(true)
run.SetColor(docx.Red)
run.SetSize(28) // 14pt in half-points
doc.SaveAs("output.docx")
```

### Key Differences

| Feature | v1 (Legacy) | v2 (Builder) | v2 (Direct) |
|---------|-------------|--------------|-------------|
| Import | `fumiama/go-docx` | `mmonterroca/docxgo` | `mmonterroca/docxgo` |
| Create doc | `docx.New()` | `NewDocumentBuilder()` | `NewDocument()` |
| Add text | `para.AddText()` | `para.Text()` | `run.SetText()` |
| Bold | `run.Bold()` | `para.Bold()` | `run.SetBold(true)` |
| Color | `run.Color("FF0000")` | `para.Color(docx.Red)` | `run.SetColor(docx.Red)` |
| Size | `run.Size("28")` (string) | `para.FontSize(14)` (points) | `run.SetSize(28)` (half-points) |
| Error handling | No errors | Accumulated in `Build()` | Immediate errors |
| Save | `WriteTo(file)` | `doc.SaveAs(path)` | `doc.SaveAs(path)` |

---

## 📚 See Also

- [Examples Directory](../examples/) - Working code examples
- [V2 Design Document](./V2_DESIGN.md) - Architecture and implementation phases
- [Migration Guide](../MIGRATION.md) - Detailed v1 to v2 migration
- [API Reference](https://pkg.go.dev/github.com/mmonterroca/docxgo) - Full API documentation

---

**Last Updated**: October 27, 2025  
**Version**: 2.0.0-beta

# go-docx v2 Examples

This directory contains comprehensive examples demonstrating the features of go-docx v2.

## Example List

### [01 - Basic Document](./basic/)
**Status**: ✅ Complete  
**Demonstrates**: Creating a simple document with paragraphs and basic text formatting.
- Document creation
- Adding paragraphs
- Adding text runs
- Basic text formatting (bold, italic)
- Saving documents

### [04 - Fields](./04_fields/)
**Status**: ✅ Complete (Phase 6)  
**Demonstrates**: Using the comprehensive field system for dynamic content.
- Page numbers and page count
- Table of Contents (TOC)
- Hyperlinks
- Date/Time fields
- Document properties
- Style references
- Field updates

### [05 - Style Management](./05_styles/)
**Status**: ✅ Complete (Phase 6)  
**Demonstrates**: Using built-in styles and the StyleManager.
- Title, Subtitle, and Heading styles
- Normal paragraph style
- Quote and IntenseQuote styles
- List paragraph style
- Footnote reference style
- Character-level formatting (bold, italic, color)
- StyleManager for querying styles

### [06 - Sections and Page Layout](./06_sections/)
**Status**: ✅ Complete (Phase 6)  
**Demonstrates**: Complete page layout and section management.
- Page sizes (A3, A4, A5, Letter, Legal, Tabloid)
- Page orientation (Portrait, Landscape)
- Custom margins (all sides configurable)
- Headers and footers (Default, First, Even)
- Dynamic page numbering
- Multi-column layouts

### [07 - Advanced Document Creation](./07_advanced/)
**Status**: ✅ Complete (Phase 6)  
**Demonstrates**: Combining all features to create professional documents.
- Professional cover page
- Table of Contents with hyperlinks
- Multiple content sections
- Custom headers and footers
- Mixed text formatting
- Working hyperlinks
- Complete document structure

## Running the Examples

Each example is a standalone Go program. Navigate to the example directory and run:

```bash
cd basic
go run main.go
```

Or run directly:

```bash
go run ./basic/main.go
```

## Example Categories

### Getting Started
- **Example 01**: Start here for basic document creation

### Phase 6 Features (Advanced)
- **Example 04**: Fields system
- **Example 05**: Style management
- **Example 06**: Sections and page layout
- **Example 07**: Complete integration

## Prerequisites

- Go 1.21 or higher
- go-docx v2 module

## Generated Files

Each example creates a `.docx` file in its directory:
- `basic/basic_demo.docx`
- `04_fields/04_fields_demo.docx`
- `05_styles/05_styles_demo.docx`
- `06_sections/06_sections_demo.docx`
- `07_advanced/07_advanced_demo.docx`

## Opening Documents

All generated documents can be opened with:
- Microsoft Word (Windows, macOS)
- LibreOffice Writer
- Google Docs (upload)
- Pages (macOS)
- Any OOXML-compatible word processor

### Field Updates

Documents with fields (TOC, page numbers) may require updating:
1. Open the document
2. Press **F9** (or right-click → "Update Field")
3. Choose "Update entire table" for TOC

## Learning Path

**Beginner**:
1. Start with Example 01 (basic)
2. Learn text formatting and paragraphs

**Intermediate**:
1. Example 05 (styles) - Understanding style system
2. Example 06 (sections) - Page layout control

**Advanced**:
1. Example 04 (fields) - Dynamic content
2. Example 07 (advanced) - Professional documents

## Code Structure

Each example follows this pattern:

```go
package main

import (
    docx "github.com/mmonterroca/docxgo"
    "github.com/mmonterroca/docxgo/domain"
)

func main() {
    // 1. Create document
    doc := docx.NewDocument()
    
    // 2. Configure layout (optional)
    section, _ := doc.DefaultSection()
    section.SetPageSize(domain.PageSizeA4)
    
    // 3. Add content
    para, _ := doc.AddParagraph()
    run, _ := para.AddRun()
    run.AddText("Hello, World!")
    
    // 4. Save
    doc.SaveToFile("output.docx")
}
```

## Common Patterns

### Adding Styled Headings

```go
h1, _ := doc.AddParagraph()
h1.SetStyle(domain.StyleIDHeading1)
run, _ := h1.AddRun()
run.AddText("Section Title")
```

### Creating Lists

```go
item, _ := doc.AddParagraph()
item.SetStyle(domain.StyleIDListParagraph)
run, _ := item.AddRun()
run.AddText("• List item")
```

### Page Numbers in Footer

```go
section, _ := doc.DefaultSection()
footer, _ := section.Footer(domain.FooterDefault)
para, _ := footer.AddParagraph()
para.SetAlignment(domain.AlignmentCenter)

r1, _ := para.AddRun()
r1.AddText("Page ")

r2, _ := para.AddRun()
r2.AddField(docx.NewPageNumberField())
```

### Adding Hyperlinks

```go
run, _ := para.AddRun()
linkField := docx.NewHyperlinkField("https://example.com", "Click here")
run.SetColor(0x0000FF)
run.SetUnderline(domain.UnderlineSingle)
run.AddField(linkField)
```

## Documentation

For comprehensive API documentation, see:
- [API Documentation](../../docs/API_DOCUMENTATION.md)
- [Migration Guide](../../MIGRATION.md) (v1 to v2)
- [Design Document](../../docs/V2_DESIGN.md)

## Contributing

Found an issue or want to add an example? See [CONTRIBUTING.md](../../CONTRIBUTING.md).

## License

These examples are part of go-docx v2 and are licensed under AGPL-3.0.

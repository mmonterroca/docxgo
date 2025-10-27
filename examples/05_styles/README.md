# Example 05: Style Management

This example demonstrates the comprehensive style management system in go-docx v2.

## Features Demonstrated

### Built-in Styles
- **Title and Subtitle**: Document title styles
- **Headings**: Heading1 through Heading9
- **Normal**: Default body text style
- **Quote styles**: Quote and IntenseQuote
- **List styles**: ListParagraph, ListBullet, ListNumber
- **Reference styles**: FootnoteReference, FootnoteText

### Style Management
- Using `StyleManager` to query styles
- Applying styles with type-safe constants (`domain.StyleID*`)
- Checking style availability with `HasStyle()`
- Character-level formatting (bold, italic, color)

## Running the Example

```bash
go run main.go
```

This will create `05_styles_demo.docx` demonstrating various built-in styles.

## Key Concepts

### Type-Safe Style IDs

v2 provides constants for all built-in styles:

```go
// Heading styles
para.SetStyle(domain.StyleIDHeading1)
para.SetStyle(domain.StyleIDHeading2)

// Text styles
para.SetStyle(domain.StyleIDNormal)
para.SetStyle(domain.StyleIDQuote)

// List styles
para.SetStyle(domain.StyleIDListParagraph)
```

### Style Manager

Access the style manager to query available styles:

```go
styleMgr := doc.StyleManager()

if styleMgr.HasStyle(domain.StyleIDHeading1) {
    para.SetStyle(domain.StyleIDHeading1)
}
```

### Character Styles

Apply character-level formatting to runs:

```go
run.SetBold(true)
run.SetItalic(true)
run.SetColor(0xFF0000) // Red
run.SetFontSize(14)
run.SetFontFamily("Arial")
```

## Output

The generated document includes:
- Styled title and headings
- Normal body paragraphs
- Quoted text with special formatting
- List items
- Footnote references
- Mixed character formatting in paragraphs

## Next Steps

- See [Example 06](../06_sections) for section and page layout management
- See [Example 07](../07_advanced) for combining all advanced features
- Read [API Documentation](../../../docs/API_DOCUMENTATION.md) for complete style reference

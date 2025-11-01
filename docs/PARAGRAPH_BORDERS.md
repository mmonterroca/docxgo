# Paragraph Borders Feature

## Overview

Added support for paragraph borders (top, bottom, left, right) to allow decorative lines and boxes around paragraphs.

## New API Methods

### Paragraph Interface

```go
// Get all borders
Borders() ParagraphBorders

// Set all borders at once
SetBorders(borders ParagraphBorders) error

// Set individual borders
SetBorderTop(border BorderStyle) error
SetBorderBottom(border BorderStyle) error
SetBorderLeft(border BorderStyle) error
SetBorderRight(border BorderStyle) error
```

### Types

```go
// ParagraphBorders represents borders for a paragraph
type ParagraphBorders struct {
    Top    BorderStyle
    Bottom BorderStyle
    Left   BorderStyle
    Right  BorderStyle
}

// BorderStyle (already existed for tables)
type BorderStyle struct {
    Style BorderLineStyle
    Width int   // Width in eighths of a point
    Color Color
}

// BorderLineStyle constants
const (
    BorderNone   BorderLineStyle = iota
    BorderSingle
    BorderDotted
    BorderDashed
    BorderDouble
    BorderTriple
    BorderThick
)
```

## Usage Examples

### Bottom Border (Decorative Line)

Perfect for adding visual separation under headings:

```go
h1, _ := doc.AddParagraph()
h1.SetStyle(domain.StyleIDHeading1)
h1Run, _ := h1.AddRun()
h1Run.AddText("Section Title")

// Add decorative line below heading
h1.SetBorderBottom(domain.BorderStyle{
    Style: domain.BorderSingle,
    Width: 6,
    Color: domain.Color{R: 0, G: 122, B: 204}, // Blue
})
```

### Full Box Border

Create a box around important text:

```go
para, _ := doc.AddParagraph()
run, _ := para.AddRun()
run.AddText("Important note with border on all sides")

para.SetBorders(domain.ParagraphBorders{
    Top: domain.BorderStyle{
        Style: domain.BorderSingle,
        Width: 4,
        Color: domain.Color{R: 128, G: 128, B: 128},
    },
    Bottom: domain.BorderStyle{
        Style: domain.BorderSingle,
        Width: 4,
        Color: domain.Color{R: 128, G: 128, B: 128},
    },
    Left: domain.BorderStyle{
        Style: domain.BorderSingle,
        Width: 4,
        Color: domain.Color{R: 128, G: 128, B: 128},
    },
    Right: domain.BorderStyle{
        Style: domain.BorderSingle,
        Width: 4,
        Color: domain.Color{R: 128, G: 128, B: 128},
    },
})
```

### Different Border Styles

```go
// Dashed border
para.SetBorderBottom(domain.BorderStyle{
    Style: domain.BorderDashed,
    Width: 6,
    Color: domain.Color{R: 200, G: 50, B: 50},
})

// Dotted border
para.SetBorderBottom(domain.BorderStyle{
    Style: domain.BorderDotted,
    Width: 6,
    Color: domain.Color{R: 200, G: 50, B: 50},
})

// Double border (elegant)
para.SetBorderBottom(domain.BorderStyle{
    Style: domain.BorderDouble,
    Width: 6,
    Color: domain.Color{R: 0, G: 0, B: 0},
})
```

## Implementation Details

### Files Modified

1. **domain/paragraph.go**
   - Added `ParagraphBorders` type
   - Added border methods to `Paragraph` interface

2. **internal/core/paragraph.go**
   - Added `borders` field to paragraph struct
   - Implemented all border methods

3. **internal/xml/paragraph.go**
   - Added `ParagraphBorders` XML structure
   - Reused existing `Border` type from table.go

4. **internal/serializer/serializer.go**
   - Added border serialization in `serializeProperties`
   - Added `hasBorders()`, `serializeBorder()`, `borderStyleToString()` helper methods

### XML Output

Generates standard OOXML paragraph border markup:

```xml
<w:p>
    <w:pPr>
        <w:pBdr>
            <w:bottom w:val="single" w:color="007ACC" w:sz="6"/>
        </w:pBdr>
    </w:pPr>
    <w:r>
        <w:t>Paragraph text</w:t>
    </w:r>
</w:p>
```

## Use Cases

1. **Section Separators**: Add decorative lines under section headings
2. **Callout Boxes**: Create bordered boxes for important notes or warnings
3. **Visual Hierarchy**: Use different border styles to indicate different types of content
4. **Technical Documents**: Add professional styling to architecture documents
5. **Table of Contents**: Add lines between major sections

## Test Example

See `examples/13_themes/test_borders/main.go` for a complete example demonstrating:
- Bottom border on headings
- Full box borders
- Different border styles (single, dashed, dotted, double, thick, triple)

Run the test:
```bash
cd examples/13_themes/test_borders
go run main.go
```

## Notes

- Border widths are in eighths of a point (same as table borders)
- A width of 8 = 1 point
- Setting `BorderNone` removes the border
- Borders are compatible with all Word versions that support OOXML
- Borders do not affect layout spacing (use `SetSpacingBefore/After` for that)

## Integration with Themes

The technical architecture example (`examples/13_themes/04_tech_architecture/`) demonstrates how to use borders with themes:

```go
h1.SetBorderBottom(domain.BorderStyle{
    Style: domain.BorderSingle,
    Width: 6,
    Color: colors.Primary, // Uses theme color
})
```

This creates consistent, theme-aware decorative lines throughout the document.

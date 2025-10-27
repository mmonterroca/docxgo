# Example 06: Section and Page Layout Management

This example demonstrates comprehensive section management, page layout configuration, and headers/footers in go-docx v2.

## Features Demonstrated

### Page Layout
- **Page sizes**: A3, A4, A5, Letter, Legal, Tabloid
- **Orientation**: Portrait and Landscape
- **Margins**: Customizable margins for all sides
- **Columns**: Single, two-column, multi-column layouts

### Headers and Footers
- **Header types**: Default, First, Even
- **Footer types**: Default, First, Even
- **Dynamic fields**: Page numbers, page count
- **Formatting**: Alignment, styles, colors

## Running the Example

```bash
go run main.go
```

This will create `06_sections_demo.docx` with configured page layout and headers/footers.

## Key Concepts

### Getting the Default Section

Every document has a default section:

```go
section, err := doc.DefaultSection()
if err != nil {
    log.Fatal(err)
}
```

### Configuring Page Size

Use predefined constants or custom dimensions:

```go
// Predefined sizes
section.SetPageSize(domain.PageSizeA4)
section.SetPageSize(domain.PageSizeLetter)
section.SetPageSize(domain.PageSizeLegal)

// Custom size (width and height in twips)
section.SetPageSize(domain.PageSize{
    Width:  12240,  // 8.5 inches
    Height: 15840,  // 11 inches
})
```

### Setting Orientation

```go
section.SetOrientation(domain.OrientationPortrait)
section.SetOrientation(domain.OrientationLandscape)
```

### Configuring Margins

Margins are specified in twips (1440 twips = 1 inch):

```go
margins := domain.Margins{
    Top:    1440, // 1 inch
    Right:  1440,
    Bottom: 1440,
    Left:   1440,
    Header: 720,  // 0.5 inch
    Footer: 720,
}
section.SetMargins(margins)
```

### Adding Headers

Access headers by type and add content:

```go
header, err := section.Header(domain.HeaderDefault)
if err != nil {
    log.Fatal(err)
}

para, _ := header.AddParagraph()
para.SetAlignment(domain.AlignmentRight)
run, _ := para.AddRun()
run.AddText("Document Title")
run.SetBold(true)
```

### Adding Footers with Page Numbers

Create dynamic page numbering:

```go
footer, err := section.Footer(domain.FooterDefault)
if err != nil {
    log.Fatal(err)
}

para, _ := footer.AddParagraph()
para.SetAlignment(domain.AlignmentCenter)

// "Page X of Y" format
r1, _ := para.AddRun()
r1.AddText("Page ")

r2, _ := para.AddRun()
pageField := docx.NewPageNumberField()
r2.AddField(pageField)

r3, _ := para.AddRun()
r3.AddText(" of ")

r4, _ := para.AddRun()
totalField := docx.NewPageCountField()
r4.AddField(totalField)
```

### Column Layouts

Configure multi-column sections:

```go
section.SetColumns(1) // Single column (default)
section.SetColumns(2) // Two-column layout
section.SetColumns(3) // Three-column layout
```

## Units

### Twips

Most measurements in Word documents use **twips** (twentieth of a point):
- 1 inch = 1440 twips
- 1 cm = 567 twips
- 1 point = 20 twips

### Common Conversions

```go
// 1 inch margins
margin := 1440

// 2.54 cm (1 inch)
margin := 1440

// 0.5 inch
margin := 720

// Custom page size (8.5" × 11")
pageSize := domain.PageSize{
    Width:  12240, // 8.5 × 1440
    Height: 15840, // 11 × 1440
}
```

## Output

The generated document includes:
- A4 portrait page with 1-inch margins
- Right-aligned header with title
- Center-aligned footer with "Page X of Y"
- Multiple pages demonstrating consistent layout
- List of available page sizes and options

## Next Steps

- See [Example 07](../07_advanced) for combining sections with fields and styles
- See [Example 04](../04_fields) for more field types
- Read [API Documentation](../../../docs/API_DOCUMENTATION.md) for complete section reference

# Example 07: Advanced Document Creation

This comprehensive example demonstrates how to combine all Phase 6 features to create professional, complex documents programmatically.

## Features Demonstrated

### Complete Integration
- **Page Layout**: A4, portrait, 1-inch margins
- **Headers**: Custom header with document title
- **Footers**: Dynamic page numbers (Page X of Y)
- **Cover Page**: Professional title page
- **Table of Contents**: Auto-generated with hyperlinks
- **Multiple Sections**: Introduction, Features, Examples, Conclusion
- **Style Variety**: Title, Subtitle, Heading1-3, Normal, Quote, List
- **Character Formatting**: Bold, italic, color, combined formatting
- **Hyperlinks**: Working links to external resources
- **Fields**: Page numbers, page count, TOC, hyperlinks

## Running the Example

```bash
go run main.go
```

This will create `07_advanced_demo.docx` - a complete, professional document.

## Document Structure

### 1. Cover Page
- Centered title with Title style
- Subtitle with Subtitle style
- Author information with italic formatting

### 2. Table of Contents
- Auto-generated from heading levels 1-3
- Hyperlinked entries (clickable navigation)
- Instructions for updating in Word

### 3. Introduction Section
- Heading 1: "Introduction"
- Heading 2: "Purpose"
- Normal paragraphs with descriptive text

### 4. Features Overview
- Detailed breakdown of Phase 6 capabilities
- Lists with ListParagraph style
- Organized subsections (H2 level)

### 5. Code Examples
- Hyperlink demonstration (clickable link)
- Quote styling with IntenseQuote
- Mixed character formatting in single paragraph

### 6. Conclusion
- Summary of capabilities
- Next steps for users
- Centered, bold final message

## Key Implementation Details

### Modular Structure

The code is organized into focused functions:

```go
// Main flow
setupHeader(header)
setupFooter(footer)
addCoverPage(doc)
addTableOfContents(doc)
addIntroduction(doc)
addFeatures(doc)
addExamples(doc)
addConclusion(doc)
```

### Header Configuration

```go
func setupHeader(header domain.Header) {
    para, _ := header.AddParagraph()
    para.SetAlignment(domain.AlignmentRight)
    
    run, _ := para.AddRun()
    run.AddText("go-docx v2 • Advanced Features Demo")
    run.SetFontSize(10)
    run.SetColor(0x4472C4) // Blue
}
```

### Footer with Fields

```go
func setupFooter(footer domain.Footer) {
    para, _ := footer.AddParagraph()
    para.SetAlignment(domain.AlignmentCenter)
    
    r1, _ := para.AddRun()
    r1.AddText("Page ")
    
    r2, _ := para.AddRun()
    r2.AddField(docx.NewPageNumberField())
    
    r3, _ := para.AddRun()
    r3.AddText(" of ")
    
    r4, _ := para.AddRun()
    r4.AddField(docx.NewPageCountField())
}
```

### Table of Contents

```go
tocOptions := map[string]string{
    "levels":          "1-3",    // Include H1, H2, H3
    "hyperlinks":      "true",   // Enable clickable links
    "hidePageNumbers": "false",  // Show page numbers
}
tocField := docx.NewTOCField(tocOptions)
run.AddField(tocField)
```

### Hyperlinks

```go
linkField := docx.NewHyperlinkField(
    "https://github.com/mmonterroca/docxgo",
    "go-docx GitHub repository",
)
run.SetColor(0x0000FF)
run.SetUnderline(domain.UnderlineSingle)
run.AddField(linkField)
```

## Best Practices Demonstrated

1. **Separation of Concerns**: Each section in its own function
2. **Consistent Styling**: Use style constants throughout
3. **Error Handling**: Check errors from API calls (abbreviated for clarity)
4. **Field Updates**: Include instructions for users to update TOC
5. **Professional Layout**: Proper spacing, alignment, formatting
6. **Reusability**: Functions can be adapted for your documents

## Opening the Document

When you open `07_advanced_demo.docx` in Microsoft Word:

1. **Update TOC**: Right-click the Table of Contents → "Update Field" (or press F9)
2. **Navigate**: Click TOC entries to jump to sections
3. **Click Links**: Hyperlinks are clickable
4. **Review Layout**: Notice headers, footers, page numbers on every page

## Adapting for Your Use Case

### Corporate Report
- Change header to company logo/name
- Add more sections with your content
- Customize colors to brand colors
- Add tables or images

### Academic Paper
- Add bibliography section
- Include footnotes/endnotes
- Use different heading styles
- Add figure captions

### User Manual
- Expand TOC levels to 4-5
- Add step-by-step procedures
- Include numbered lists
- Add troubleshooting sections

## Output

The generated document is approximately 4-5 pages and includes:
- Professional cover page
- Interactive table of contents
- Multiple content sections
- Consistent headers and footers
- Dynamic page numbering
- Working hyperlinks
- Various text formatting styles

## Next Steps

- Review individual examples for specific features
- Read [API Documentation](../../../docs/API_DOCUMENTATION.md)
- Check [MIGRATION.md](../../../MIGRATION.md) for v1→v2 transition
- Explore [V2_DESIGN.md](../../../docs/V2_DESIGN.md) for architecture

## Performance Note

This example prioritizes clarity over performance. In production:
- Add comprehensive error handling
- Use buffered operations for large documents
- Consider memory usage with many fields
- Test with real-world document sizes

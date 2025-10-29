# Example 04: Fields

This example demonstrates how to use fields in Word documents using go-docx v2.

## What are Fields?

Fields are dynamic elements in Word documents that can automatically calculate and display values. They are updated when:
- The document is opened in Word
- The user presses F9 (Update Field)
- The document is printed or converted to PDF

## Field Types Demonstrated

### 1. Page Numbers
- `PAGE`: Current page number
- `NUMPAGES`: Total page count
- Used in headers/footers for page numbering

### 2. Table of Contents (TOC)
- Automatically generated from heading styles
- Supports custom heading levels (e.g., 1-3)
- Includes hyperlinks to sections
- Updates automatically when headings change

### 3. Hyperlinks
- `HYPERLINK`: Links to external URLs or internal bookmarks
- Displays custom text while linking to a URL
- Can be styled with color and underline

### 4. Style References
- `STYLEREF`: References text with a specific style
- Useful for running headers (e.g., chapter titles)

### 5. Sequence Numbers
- `SEQ`: Automatic numbering for figures, tables, equations
- Maintains separate counters for different categories

### 6. Date/Time
- `DATE`: Current date
- `TIME`: Current time
- Supports custom formatting

## Running the Example

```bash
cd examples/04_fields
go run main.go
```

This will create `fields_example.docx` with:
- A Table of Contents at the beginning
- Multiple sections with headings
- Page numbers in the footer ("Page X of Y")
- Sample hyperlinks
- Various content to demonstrate field functionality

## Code Structure

```go
// Create page number field
pageField := docx.NewPageNumberField()
run.AddField(pageField)

// Create TOC with custom options
tocOptions := map[string]string{
    "levels":     "1-3",  // Include heading levels 1-3
    "hyperlinks": "true", // Enable hyperlinks
}
tocField := docx.NewTOCField(tocOptions)

// Create hyperlink
hyperlinkField := docx.NewHyperlinkField(
    "https://example.com",
    "Example Link",
)
run.AddField(hyperlinkField)
```

## Important Notes

### Updating Fields
Fields display placeholder values until updated in Word. To update:
1. Open the document in Microsoft Word
2. Press **Ctrl+A** (Select All)
3. Press **F9** (Update Fields)
4. Or right-click on a field and select "Update Field"

### Field Codes
You can view the underlying field codes in Word:
- Press **Alt+F9** to toggle field code display
- Right-click a field and select "Toggle Field Codes"

### TOC Customization
The TOC field supports many switches:
- `\o "1-3"`: Use heading levels 1-3
- `\h`: Include hyperlinks
- `\z`: Hide tab leader in Web Layout
- `\u`: Use outline levels
- `\n`: Hide page numbers
- `\p`: Use paragraph formatting

## Advanced Usage

### Custom Field Codes
```go
field := docx.NewField(docx.FieldTypeCustom)
field.SetCode(`STYLEREF "Heading 1" \* MERGEFORMAT`)
```

### Complex TOC
```go
tocOptions := map[string]string{
    "levels":          "1-5",    // Deeper hierarchy
    "hidePageNumbers": "true",   // No page numbers
    "hideTabLeader":   "true",   // No dots
}
```

### Cross-References
```go
// Reference a bookmark
refField := docx.NewField(docx.FieldTypeRef)
refField.SetCode(`REF MyBookmark \h`)
```

## See Also

- [Example 01: Hello World](../01_hello/) - Basic document creation
- [Example 02: Formatted Text](../02_formatted/) - Text formatting
- [Example 03: Table of Contents](../03_toc/) - TOC generation
- [API Documentation](../../docs/API_DOCUMENTATION.md)
- [OOXML Field Reference](http://officeopenxml.com/WPfields.php)

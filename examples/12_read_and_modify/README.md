# Example 12: Read and Modify Documents

This example demonstrates the complete document read/modify/write workflow in go-docx v2.

## What This Example Demonstrates

### 1. **Document Creation**
Creates a comprehensive showcase document featuring:
- Title and subtitle formatting
- Multiple paragraph styles (Heading1, Heading2, Heading3, Quote, Normal)
- Rich text formatting (bold, italic, underline, colors, font sizes)
- Tables with merged cells and styling
- List paragraphs

### 2. **Document Reading**
- Opens an existing .docx file using `docx.OpenDocument()`
- Parses and reconstructs the document structure
- Provides access to all document elements

### 3. **Content Inspection**
- Traverses paragraphs and tables
- Extracts text content
- Reads table dimensions (rows/columns)
- Gathers document statistics

### 4. **Document Modification**
- Adds new paragraphs to existing documents
- Creates new runs with custom formatting
- Adds tables with styled cells
- Applies paragraph and character styles
- Preserves existing content and formatting

### 5. **Save As Different File**
- Saves modified document with a new filename
- Preserves all original and new content
- Allows side-by-side comparison of original vs modified

## How to Run

```bash
cd examples/12_read_and_modify
go run main.go
```

## Output Files

This example generates two files:

1. **`12_showcase_original.docx`** - The original comprehensive document with all features
2. **`12_modified_document.docx`** - The modified version with added section 5

Compare these files in Microsoft Word or LibreOffice to see the modifications clearly.

## Code Walkthrough

### Step 1: Create Showcase Document

```go
builder := docx.NewDocumentBuilder(
    docx.WithTitle("Document Showcase"),
    docx.WithAuthor("go-docx v2"),
)

builder.AddParagraph().
    Text("Document Showcase - All Features").
    Style(domain.StyleIDTitle).
    End()

// ... add more content
doc, _ := builder.Build()
doc.SaveAs("12_showcase_original.docx")
```

### Step 2: Read Document

```go
doc, err := docx.OpenDocument("12_showcase_original.docx")
if err != nil {
    log.Fatal(err)
}
```

### Step 3: Inspect Content

```go
paragraphs := doc.Paragraphs()
tables := doc.Tables()

fmt.Printf("Paragraphs: %d\n", len(paragraphs))
fmt.Printf("Tables: %d\n", len(tables))

for i, para := range paragraphs {
    fmt.Printf("%d. %s\n", i+1, para.Text())
}
```

### Step 4: Modify Document

```go
// Add new paragraph
newPara, _ := doc.AddParagraph()
newPara.SetStyle(domain.StyleIDHeading1)
run, _ := newPara.AddRun()
run.SetText("5. Modifications (Added by Reader)")
run.SetColor(docx.Blue)

// Add new table
table, _ := doc.AddTable(3, 2)
table.SetStyle(domain.TableStyleMediumShading)
// ... populate table
```

### Step 5: Save Modified Version

```go
err := doc.SaveAs("12_modified_document.docx")
if err != nil {
    log.Fatal(err)
}
```

## API Features Demonstrated

| Feature | API Method | Example Usage |
|---------|-----------|---------------|
| Open document | `docx.OpenDocument()` | `doc, err := docx.OpenDocument("file.docx")` |
| Get paragraphs | `doc.Paragraphs()` | `paras := doc.Paragraphs()` |
| Get tables | `doc.Tables()` | `tables := doc.Tables()` |
| Add paragraph | `doc.AddParagraph()` | `para, _ := doc.AddParagraph()` |
| Add run | `para.AddRun()` | `run, _ := para.AddRun()` |
| Set text | `run.SetText()` | `run.SetText("Hello")` |
| Set style | `para.SetStyle()` | `para.SetStyle(domain.StyleIDHeading1)` |
| Set color | `run.SetColor()` | `run.SetColor(docx.Blue)` |
| Add table | `doc.AddTable()` | `table, _ := doc.AddTable(3, 2)` |
| Get row | `table.Row()` | `row, _ := table.Row(0)` |
| Get cell | `row.Cell()` | `cell, _ := row.Cell(0)` |

## Real-World Use Cases

This pattern is useful for:

1. **Template Processing** - Read template documents and fill in placeholders
2. **Document Merging** - Combine multiple documents into one
3. **Content Updates** - Update specific sections while preserving formatting
4. **Automated Reports** - Generate reports by modifying base documents
5. **Document Analysis** - Extract and analyze content from existing documents
6. **Format Conversion** - Read, transform, and save in different formats

## Technical Notes

### Reader Implementation Status (Phase 10)

The reader infrastructure is currently **55% complete**:

- ✅ **Working**: Paragraph reading, run reconstruction, table hydration, image reconstruction
- 🚧 **In Progress**: Style reading, section reading, field reading
- ⏳ **Planned**: Header/footer reading, advanced table features

### Limitations

Current known limitations:

- Complex nested tables may not fully reconstruct
- Some advanced table styles may lose fidelity
- Custom styles require full Phase 10 implementation
- Header/footer modification not yet supported

### Document Fidelity

The reader aims to preserve:
- ✅ Text content
- ✅ Basic formatting (bold, italic, underline, color)
- ✅ Paragraph styles
- ✅ Table structure and content
- 🚧 Complex styling (partial)
- ⏳ Advanced features (planned)

## Expected Output

When you run this example, you'll see:

```
📝 Step 1: Creating comprehensive showcase document...
   ✅ Created: 12_showcase_original.docx

📖 Step 2: Reading the document back...
   ✅ Document loaded successfully

🔍 Step 3: Inspecting document content...
   📊 Document statistics:
      • Paragraphs: 25
      • Tables: 1

   📝 First 3 paragraphs:
      1. "Document Showcase - All Features"
      2. "This document demonstrates all capabilities of go-docx v2"
      3. ""

   📋 First table:
      • Rows: 4
      • Columns: 3

✏️  Step 4: Modifying the document...
   ✅ Modifications applied

💾 Step 5: Saving modified document as '12_modified_document.docx'...
   ✅ Saved: 12_modified_document.docx

🎉 Complete! Compare the original and modified documents to see the changes.

Generated files:
  📄 12_showcase_original.docx (original showcase)
  📄 12_modified_document.docx (modified version)
```

## Next Steps

After running this example:

1. Open both generated files in Word/LibreOffice
2. Compare the original and modified versions
3. Verify that section 5 was added correctly
4. Check that all original content was preserved
5. Examine the new table and formatting

## Learn More

- See `docs/V2_DESIGN.md` for architecture details
- See `docs/V2_API_GUIDE.md` for complete API reference
- See other examples for specific feature demonstrations

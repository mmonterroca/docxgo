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
Demonstrates **two types of modifications**:

#### A. **Editing Existing Content**
- Change text color of existing runs (title → blue)
- Apply formatting to existing text (subtitle → italic)
- Update paragraph styles (Heading1 → Heading2)
- Modify text content of existing runs
- Update table cell values and formatting

#### B. **Adding New Content**
- Add new paragraphs to existing documents
- Create new runs with custom formatting
- Add tables with styled cells
- Apply paragraph and character styles
- Preserve existing content and formatting

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

### Step 4A: Edit Existing Content

```go
// Get existing paragraphs
paragraphs := doc.Paragraphs()

// Modify title color
if len(paragraphs) > 0 {
    title := paragraphs[0]
    runs := title.Runs()
    if len(runs) > 0 {
        runs[0].SetColor(docx.Blue) // Change color
    }
}

// Find and update specific paragraph
for _, para := range paragraphs {
    if para.Text() == "1. Text Formatting Capabilities" {
        runs := para.Runs()
        if len(runs) > 0 {
            runs[0].SetText("1. Text Formatting (MODIFIED)")
            runs[0].SetColor(docx.Red)
        }
        break
    }
}

// Update table cell value
tables := doc.Tables()
if len(tables) > 0 {
    row, _ := tables[0].Row(2)
    cell, _ := row.Cell(2)
    paras := cell.Paragraphs()
    if len(paras) > 0 {
        runs := paras[0].Runs()
        if len(runs) > 0 {
            runs[0].SetText("$35.00 (UPDATED)")
            runs[0].SetBold(true)
        }
    }
}
```

### Step 4B: Add New Content

```go
// Add new paragraph
newPara, _ := doc.AddParagraph()
newPara.SetStyle(domain.StyleIDHeading1)
run, _ := newPara.AddRun()
run.SetText("5. Modifications (Added by Reader)")
run.SetColor(docx.Purple)

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

| Category | API Method | Example Usage |
|---------|-----------|---------------|
| **Opening** | | |
| Open document | `docx.OpenDocument()` | `doc, err := docx.OpenDocument("file.docx")` |
| **Reading** | | |
| Get paragraphs | `doc.Paragraphs()` | `paras := doc.Paragraphs()` |
| Get tables | `doc.Tables()` | `tables := doc.Tables()` |
| Get runs | `para.Runs()` | `runs := para.Runs()` |
| Get text | `para.Text()` | `text := para.Text()` |
| Get table row | `table.Row()` | `row, _ := table.Row(0)` |
| Get cell | `row.Cell()` | `cell, _ := row.Cell(0)` |
| **Editing** | | |
| Set text | `run.SetText()` | `run.SetText("New text")` |
| Set color | `run.SetColor()` | `run.SetColor(docx.Blue)` |
| Set bold | `run.SetBold()` | `run.SetBold(true)` |
| Set italic | `run.SetItalic()` | `run.SetItalic(true)` |
| Set style | `para.SetStyle()` | `para.SetStyle(domain.StyleIDHeading1)` |
| **Adding** | | |
| Add paragraph | `doc.AddParagraph()` | `para, _ := doc.AddParagraph()` |
| Add run | `para.AddRun()` | `run, _ := para.AddRun()` |
| Add table | `doc.AddTable()` | `table, _ := doc.AddTable(3, 2)` |

## Real-World Use Cases

This pattern is useful for:

1. **Template Processing** - Read template documents and fill in placeholders with dynamic data
2. **Content Updates** - Update specific sections (prices, dates, names) while preserving formatting
3. **Document Merging** - Combine multiple documents into one master document
4. **Automated Reports** - Generate reports by modifying base documents with current data
5. **Batch Processing** - Apply consistent changes across multiple documents
6. **Document Analysis** - Extract and analyze content from existing documents
7. **Format Conversion** - Read, transform, and save in different formats
8. **Version Updates** - Update document versions by modifying specific sections

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
   → Modifying existing paragraphs...
   → Modifying existing table...
   → Adding new section...
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

1. **Open both generated files** in Word/LibreOffice
2. **Compare the original and modified versions** to see:
   - Title is now **blue** (was black)
   - Subtitle is now **italic** (was normal)
   - Section 1 heading text changed to "1. Text Formatting **(MODIFIED)**" in **red**
   - Table price updated from $30.00 to **$35.00 (UPDATED)** in green
   - New section 5 added with content summary
3. **Verify all original content was preserved** (no data loss)
4. **Examine the modification summary** in section 5

## Learn More

- See `docs/V2_DESIGN.md` for architecture details
- See `docs/V2_API_GUIDE.md` for complete API reference
- See other examples for specific feature demonstrations

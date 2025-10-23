# go-docx Examples

This directory contains practical, runnable examples demonstrating the capabilities of the go-docx library.

## Running the Examples

Each example is in its own directory. To run any example:

```bash
# Navigate to an example directory
cd 01_hello

# Run the example
go run main.go

# The generated .docx file will be created in that directory
```

## Available Examples

### 01_hello - Hello World
The simplest possible document - creates a basic "Hello, World!" document.

**Run:**
```bash
cd 01_hello && go run main.go
```

### 02_formatted - Formatted Text
Shows various text formatting options including bold, italic, colors, and sizes.

**Run:**
```bash
cd 02_formatted && go run main.go
```

### 03_toc - Table of Contents
Professional document with automatic Table of Contents, headings, and page numbers.

**Run:**
```bash
cd 03_toc && go run main.go
```

**After opening in Word:**
- Press `Ctrl+A` then `F9` to update all fields
- Or right-click TOC → "Update Field" → "Update entire table"

### v030_demo - Professional Document (v0.3.0) ⭐ NEW
Comprehensive 12-page demo showcasing all v0.3.0 features:
- ✅ Cover page with modern Calibri typography
- ✅ Table of Contents with 5 chapters
- ✅ Headers and footers with page numbers (Page X of Y)
- ✅ Hyperlinks to external resources (GitHub, Microsoft Docs, Go.dev)
- ✅ Tables with version history and style specifications
- ✅ Code examples with proper formatting
- ✅ Paragraph indentation examples

**Run:**
```bash
cd v030_demo && go run main.go && open v030_demo.docx
```

**Features Demonstrated:**
- `AddPageNumberFooter()` - Page X of Y footer
- `AddHyperlinkField()` - External and internal links
- `AddSmartHeading()` - TOC-compatible headings
- `AddSeqField()` - Auto-numbered tables
- `Paragraph.Indent()` - Bullet list indentation
- Professional table layouts

## Testing All Examples

Run all examples sequentially:

```bash
for d in 01_hello 02_formatted 03_toc v030_demo; do
  (cd $d && go run main.go)
done
```

## Requirements

- Go 1.20 or higher
- Microsoft Word or LibreOffice to view generated files

## More Information

- [Complete API Documentation](../docs/API_DOCUMENTATION.md)
- [Project README](../README.md)
- [Demo Test](../demo_test.go)

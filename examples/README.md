# go-docx Examples

This directory contains practical, runnable examples demonstrating go-docx v2 capabilities.

## ✅ Working Examples (v2)

### [01_basic/](./01_basic/) - Basic Builder Pattern
**Status**: ✅ Fully functional  
**Demonstrates**: Document creation with fluent builder API
- DocumentBuilder with options (title, author, font, margins)
- Fluent paragraph building (Text, Bold, Italic, Color, FontSize, Alignment)
- Mixed formatting in paragraphs
- Simple table creation
- Predefined color constants

**Run:**
```bash
cd 01_basic && go run main.go
```

### [02_intermediate/](./02_intermediate/) - Product Catalog
**Status**: ✅ Fully functional  
**Demonstrates**: Professional multi-section document
- Professional document layout
- Multiple sections with headings
- Product tables (3 tables with different products)
- Mixed text formatting
- Document metadata (title, author, subject)

**Run:**
```bash
cd 02_intermediate && go run main.go
```

### [08_images/](./08_images/) - Image Insertion
**Status**: ✅ Fully functional  
**Demonstrates**: Complete image handling
- 9 image formats (PNG, JPEG, GIF, BMP, TIFF, SVG, WEBP, ICO, EMF)
- Inline and floating images
- Custom dimensions (pixels, inches, EMUs)
- Positioning (left, center, right, custom coordinates)
- Automatic format detection

**Run:**
```bash
cd 08_images && go run main.go
```

### [09_advanced_tables/](./09_advanced_tables/) - Advanced Table Features
**Status**: ✅ Fully functional  
**Demonstrates**: Complete table manipulation
- Cell merging (horizontal colspan, vertical rowspan)
- Nested tables (tables within cells)
- 8 built-in table styles
- Row height control
- Cell alignment and styling

**Run:**
```bash
cd 09_advanced_tables && go run main.go
```

### [basic/](./basic/) - Simple API Example
**Status**: ✅ Fully functional  
**Demonstrates**: Direct domain API (non-builder)
- Simple document creation
- Basic paragraphs and text runs
- Text formatting
- Basic tables

**Run:**
```bash
cd basic && go run main.go
```

## ⚠️ Examples Under Renovation

The following examples use APIs that are being refactored:
- `04_fields/` - Fields system (TOC, page numbers, hyperlinks) - **Needs v2 API update**
- `05_styles/` - Style management - **Needs v2 API update**
- `06_sections/` - Sections and page layout - **Needs v2 API update**
- `07_advanced/` - Advanced integration - **Needs v2 API update**

These will be updated to match the v2 API in a future release.

## Requirements

- Go 1.21 or higher
- Microsoft Word or LibreOffice to view generated files

## Quick Test - Compile All Working Examples

```bash
# From the examples directory
cd 01_basic && go build && cd ../02_intermediate && go build && cd ../08_images && go build && cd ../09_advanced_tables && go build && cd ../basic && go build && cd ..
echo "✅ All working examples compiled successfully!"
```

## Documentation

- [V2 Examples README](./v2_README.md) - Detailed v2 example documentation
- [API Reference](https://pkg.go.dev/github.com/mmonterroca/docxgo)
- [Design Document](../docs/V2_DESIGN.md)
- [Migration Guide](../MIGRATION.md) - Converting v1 to v2

## Contributing Examples

Have a great example? We'd love to include it! See [CONTRIBUTING.md](../CONTRIBUTING.md).

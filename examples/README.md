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

### [03_toc/](./03_toc/) - Table of Contents (NEW!)
**Status**: ✅ Fully functional  
**Demonstrates**: Automatic Table of Contents generation
- Cover page with Title and Subtitle styles
- TOC field configured for Heading 1 and Heading 2
- Placeholder result so the TOC looks polished before updating
- Chapters, sub-sections, and appendix content driven by heading styles

**Run:**
```bash
cd 03_toc && go run main.go
```

### [04_fields/](./04_fields/) - Fields System (NEW!)
**Status**: ✅ Fully functional  
**Demonstrates**: Complete field system
- Page numbers and page count fields
- Table of Contents (TOC) with custom options
- Hyperlinks to external URLs
- Headers and footers
- Page breaks

**Run:**
```bash
cd 04_fields && go run main.go
```

### [05_styles/](./05_styles/) - Style Management (NEW!)
**Status**: ✅ Fully functional  
**Demonstrates**: Built-in style system
- 40+ built-in paragraph styles (Normal, Heading1-9, Title, Subtitle, etc.)
- Character-level formatting (bold, italic, color, font size)
- Mixed formatting within paragraphs
- Quote and list paragraph styles

**Run:**
```bash
cd 05_styles && go run main.go
```

### [06_sections/](./06_sections/) - Sections and Page Layout (NEW!)
**Status**: ✅ Fully functional  
**Demonstrates**: Advanced page layout
- Custom page sizes (A4, Letter, Legal, etc.)
- Page orientation (portrait, landscape)
- Custom margins
- Headers and footers with dynamic fields
- Multi-page documents

**Run:**
```bash
cd 06_sections && go run main.go
```

### [07_advanced/](./07_advanced/) - Advanced Integration (NEW!)
**Status**: ✅ Fully functional  
**Demonstrates**: All Phase 6 features combined
- Professional cover page
- Table of Contents with hyperlinks
- Headers and footers
- Multiple heading levels
- Mixed formatting and styles
- Page numbers (Page X of Y)
- Hyperlinks
- Quotes and emphasis

**Run:**
```bash
cd 07_advanced && go run main.go
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

### [10_paragraph_spacing/](./10_paragraph_spacing/) - Paragraph Spacing (NEW!)
**Status**: ✅ Fully functional  
**Demonstrates**: Line and paragraph spacing controls
- Set spacing before and after paragraphs (twips)
- Configure exact vs. at-least line spacing rules
- Mix typography blocks for comparison
- Save finished document ready for inspection

**Run:**
```bash
cd 10_paragraph_spacing && go run main.go
```

### [11_multi_section/](./11_multi_section/) - Multi-Section Layouts (NEW!)
**Status**: ✅ Fully functional  
**Demonstrates**: Independent layouts per section
- Section breaks (Next Page, Continuous)
- Per-section headers and footers
- Portrait ↔ landscape transitions
- Unique margin and column settings per section
- Dynamic page numbering maintained across sections

**Run:**
```bash
cd 11_multi_section && go run main.go
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

## Testing All Examples

### Quick Test - Verify Compilation
Run the included test script to verify all examples compile:

```bash
./test_all.sh
```

This will test all 9 working examples and report results.

---

### Complete Validation - Generate and Validate Documents

#### 🚀 Run All Examples
Execute all examples and generate .docx files:

```bash
./run_all_examples.sh
```

**Output**: 8 .docx files in their respective directories

#### 🔍 Validate OOXML Integrity
Verify the generated documents are valid:

```bash
./validate_docx.sh
```

**Checks**:
- ✅ ZIP structure integrity
- ✅ Required OOXML files present
- ✅ Valid Office Open XML format

#### 📊 View Validation Results
See detailed validation reports:
- [VALIDATION_COMPLETE.md](./VALIDATION_COMPLETE.md) - Full validation report
- [VALIDATION_RESULTS.md](./VALIDATION_RESULTS.md) - Detailed feature checklist

---

### Generated Documents

After running `./run_all_examples.sh`, you'll have:

```
examples/
├── 01_basic/01_basic_builder.docx (4.1KB)
├── 02_intermediate/02_intermediate_builder.docx (4.7KB)
├── 03_toc/03_toc_demo.docx (4.3KB)
├── 04_fields/fields_example.docx (4.0KB)
├── 05_styles/05_styles_demo.docx (3.7KB)
├── 06_sections/06_sections_demo.docx (3.9KB)
├── 07_advanced/07_advanced_demo.docx (4.6KB)
├── 08_images/08_images_output.docx (4.2KB)
├── 09_advanced_tables/09_advanced_tables_output.docx (4.8KB)
└── 11_multi_section/11_multi_section_demo.docx (4.4KB)
```

**All documents are ready to open in**:
- Microsoft Word (Windows/macOS)
- LibreOffice Writer
- Google Docs
- Any OOXML-compatible word processor

## Documentation

- [V2 Examples README](./v2_README.md) - Detailed v2 example documentation
- [API Reference](https://pkg.go.dev/github.com/mmonterroca/docxgo)
- [Design Document](../docs/V2_DESIGN.md)
- [Migration Guide](../MIGRATION.md) - Converting v1 to v2

## Contributing Examples

Have a great example? We'd love to include it! See [CONTRIBUTING.md](../CONTRIBUTING.md).

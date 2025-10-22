# go-docx API Documentation

**Professional API Documentation for DOCX Document Generation**

[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-AGPL%203.0-blue.svg)](../LICENSE)

## 📖 Table of Contents

- [Introduction](#introduction)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Complete API Reference](#complete-api-reference)
  - [Document Base](#document-base)
  - [Paragraphs and Text](#paragraphs-and-text)
  - [Tables](#tables)
  - [Images and Drawings](#images-and-drawings)
  - [Table of Contents (TOC)](#table-of-contents-toc)
  - [Bookmarks and References](#bookmarks-and-references)
  - [Word Fields](#word-fields)
  - [Styles and Formatting](#styles-and-formatting)
  - [Shapes and Canvas](#shapes-and-canvas)
- [Practical Examples](#practical-examples)
- [How to Test and Run](#how-to-test-and-run)
- [Advanced Use Cases](#advanced-use-cases)
- [Troubleshooting](#troubleshooting)

---

## 🎯 Introduction

**go-docx** is a high-performance Go library for creating and manipulating Microsoft Word (.docx) documents programmatically. This is an enhanced version (SlideLang fork) that includes professional features for technical documentation.

### Key Features

✅ **Professional Document Generation**
- Create documents from scratch with full formatting
- Dynamic Table of Contents with navigation
- Native heading styles (H1-H4)
- Cross-references and bookmarks

✅ **Rich Content**
- Paragraphs with advanced formatting (bold, italic, color, size)
- Tables with customizable borders
- Images (inline and anchored)
- Shapes and vector drawings

✅ **Advanced Functionality**
- Word fields (PAGE, NUMPAGES, TOC, PAGEREF, SEQ, REF)
- Automatic numbering for figures and tables
- Paragraph indentation and alignment
- Headers and footers

✅ **ECMA-376 (Office Open XML) Compliant**

---

## 📦 Installation

### Option 1: Using `go get`

```bash
go get github.com/SlideLang/go-docx@latest
```

### Option 2: Add to `go.mod`

```go
require github.com/SlideLang/go-docx v0.1.0-slidelang
```

Then run:

```bash
go mod download
```

### Verify Installation

```bash
go mod tidy
go test github.com/SlideLang/go-docx
```

---

## 🚀 Quick Start

### Minimal Example: "Hello World"

```go
package main

import (
    "os"
    "github.com/fumiama/go-docx"
)

func main() {
    // Create a new document
    doc := docx.New()
    
    // Add a paragraph with text
    para := doc.AddParagraph()
    para.AddText("Hello, World!")
    
    // Set page size
    doc.WithA4Page()
    
    // Save the document
    f, _ := os.Create("hello.docx")
    defer f.Close()
    doc.WriteTo(f)
}
```

### Run the Example

```bash
go run main.go
```

Open `hello.docx` in Microsoft Word to see the result.

---

## 📚 Complete API Reference

### Document Base

#### Create a New Document

```go
doc := docx.New()
```

#### Set Default Theme

```go
doc := docx.New().WithDefaultTheme()
```

#### Set Page Size

```go
// A4 size (important: add at the END, after all content)
doc.WithA4Page()
```

**⚠️ Important:** Always call `WithA4Page()` at the end, after adding all content.

#### Parse an Existing Document

```go
readFile, err := os.Open("document.docx")
if err != nil {
    panic(err)
}
defer readFile.Close()

fileinfo, _ := readFile.Stat()
size := fileinfo.Size()

doc, err := docx.Parse(readFile, size)
if err != nil {
    panic(err)
}
```

#### Save the Document

```go
// Option 1: To a file
f, err := os.Create("output.docx")
if err != nil {
    panic(err)
}
defer f.Close()
doc.WriteTo(f)

// Option 2: To any io.Writer
var buf bytes.Buffer
doc.WriteTo(&buf)
```

---

### Paragraphs and Text

#### Add Paragraph

```go
para := doc.AddParagraph()
```

#### Add Text with Formatting

```go
// Simple text
para.AddText("This is simple text")

// Text with chained formatting
run := para.AddText("Formatted text")
run.Bold()                    // Bold
run.Italic()                  // Italic
run.Color("FF0000")           // Red color
run.Size("28")                // Size 14pt (28 = 14*2)
run.Underline("single")       // Underline
run.Highlight("yellow")       // Yellow highlight
run.Strike(true)              // Strikethrough

// Combination of formats
para.AddText("Important").Bold().Color("FF0000").Size("32")
```

#### Paragraph Alignment

```go
// Horizontal alignment
para.Justification("left")    // Left (default)
para.Justification("center")  // Center
para.Justification("right")   // Right
para.Justification("both")    // Justified
```

#### Indentation

```go
// Indentation in twips (1440 twips = 1 inch = 2.54 cm)
// 720 twips = 0.5 inch = 1.27 cm

// Left indentation only
para.Indent(720, 0, 0)

// First line indentation
para.Indent(720, 360, 0)

// Hanging indentation (for lists)
para.Indent(720, 0, 360)
```

#### Page Break

```go
para := doc.AddParagraph()
para.AddPageBreaks()
```

#### Paragraph Styles

```go
para.Style("Heading1")  // Heading 1
para.Style("Heading2")  // Heading 2
para.Style("Heading3")  // Heading 3
para.Style("Heading4")  // Heading 4
para.Style("Normal")    // Normal
```

---

### Tables

#### Create Simple Table

```go
// Table with 3 rows x 4 columns
table := doc.AddTable(3, 4, 0, nil)

// Access a cell
cell := table.TableRows[0].TableCells[0]

// Add text to cell
para := cell.AddParagraph()
para.AddText("Cell content")
```

#### Create Table with Specific Dimensions

```go
// Dimensions in twips
rowHeights := []int64{400, 400, 400}  // Height of each row
colWidths := []int64{2000, 3000, 2500, 2000}  // Width of each column
tableWidth := int64(9500)  // Total table width

table := doc.AddTableTwips(rowHeights, colWidths, tableWidth, nil)
```

#### Customize Table Borders

```go
borderColors := &docx.APITableBorderColors{
    Top:     "0000FF",  // Blue
    Left:    "FF0000",  // Red
    Bottom:  "00FF00",  // Green
    Right:   "FFFF00",  // Yellow
    InsideH: "000000",  // Black (horizontal inside)
    InsideV: "000000",  // Black (vertical inside)
}

table := doc.AddTable(3, 4, 0, borderColors)
```

#### Table Alignment

```go
table.Justification("center")  // Center table
table.Justification("left")    // Align left
table.Justification("right")   // Align right
```

#### Cell Shading

```go
cell := table.TableRows[0].TableCells[0]
cell.Shade("clear", "auto", "D9E2F3")  // Light blue background
```

---

### Images and Drawings

#### Add Image from File

```go
para := doc.AddParagraph()
run, err := para.AddInlineDrawingFrom("image.png")
if err != nil {
    panic(err)
}
```

#### Add Image from Bytes

```go
imageData, err := os.ReadFile("image.jpg")
if err != nil {
    panic(err)
}

para := doc.AddParagraph()
run, err := para.AddInlineDrawing(imageData)
```

#### Anchored Image (with text wrapping)

```go
para := doc.AddParagraph()
run, err := para.AddAnchorDrawingFrom("diagram.png")
```

#### Resize Image

```go
// Size in EMUs (English Metric Units)
// 914400 EMUs = 1 inch
width := int64(914400 * 3)   // 3 inches
height := int64(914400 * 2)  // 2 inches

// For inline image
if drawing, ok := run.Children[0].(*docx.Drawing); ok {
    if drawing.Inline != nil {
        drawing.Inline.Size(width, height)
    }
}
```

---

### Table of Contents (TOC)

#### Add Basic TOC

```go
opts := docx.DefaultTOCOptions()
opts.Title = "Table of Contents"
opts.Depth = 3  // Show up to H3
opts.PageNumbers = true
opts.Hyperlinks = true

err := doc.AddTOC(opts)
```

#### Custom TOC Options

```go
opts := docx.TOCOptions{
    Title:       "Contents",
    Depth:       4,            // H1-H4
    PageNumbers: true,         // Show page numbers
    Hyperlinks:  true,         // Clickable hyperlinks
    RightAlign:  true,         // Align numbers to the right
    TabLeader:   "dot",        // Dotted line
}
doc.AddTOC(opts)
```

#### Add Heading with TOC Entry

```go
// Add heading that will appear in the TOC
h1 := doc.AddHeadingWithTOC("1. Introduction", 1, 1)
h1.Style("Heading1")

h2 := doc.AddHeadingWithTOC("1.1 Background", 2, 2)
h2.Style("Heading2")
```

**Parameters:**
- `text`: Heading text
- `level`: Heading level (1-4)
- `tocIndex`: Unique index for TOC (increment for each heading)

#### Smart Heading

```go
// Add heading with automatic style and bookmark
para := doc.AddSmartHeading("2. Methodology", 1)
```

---

### Bookmarks and References

#### Add Simple Bookmark

```go
para := doc.AddParagraph()
para.AddText("This section is important")
para.AddBookmark("important_section")
```

#### Bookmark for TOC

```go
para := doc.AddParagraph()
para.AddText("Chapter 1: Introduction")
para.AddTOCBookmark("Chapter 1", 1)  // tocNumber is the index in TOC
```

#### Add Reference to Bookmark

```go
para := doc.AddParagraph()
para.AddText("See section ")
para.AddRefField("important_section", true)  // true = with hyperlink
para.AddText(" for more details.")
```

#### Reference with Page Number

```go
para := doc.AddParagraph()
para.AddText("See page ")
para.AddPageRefField("_Toc000000001", true)
para.AddText(" for more information.")
```

---

### Word Fields

Word fields are dynamic elements that update automatically.

#### TOC Field (Table of Contents)

```go
para := doc.AddParagraph()
para.AddTOCField(3, true, true)  // depth=3, hyperlinks=true, pageNumbers=true
```

#### PAGE Field (Current Page Number)

```go
para := doc.AddParagraph()
para.AddText("Page ")
para.AddPageField()
```

#### NUMPAGES Field (Total Pages)

```go
para := doc.AddParagraph()
para.AddText("Total pages: ")
para.AddNumPagesField()
```

#### SEQ Field (Automatic Numbering)

```go
// For figures
figurePara := doc.AddParagraph()
figurePara.AddText("Figure ")
figurePara.AddSeqField("Figure", "ARABIC")  // Figure 1, Figure 2, ...
figurePara.AddText(": System diagram")

// For tables
tablePara := doc.AddParagraph()
tablePara.AddText("Table ")
tablePara.AddSeqField("Table", "ARABIC")  // Table 1, Table 2, ...
tablePara.AddText(": Experimental results")
```

**Available formats:**
- `"ARABIC"`: 1, 2, 3, ...
- `"ROMAN"`: I, II, III, ...
- `"roman"`: i, ii, iii, ...
- `"ALPHABETIC"`: A, B, C, ...
- `"alphabetic"`: a, b, c, ...

#### REF Field (Reference to Bookmark)

```go
para := doc.AddParagraph()
para.AddText("As mentioned in ")
para.AddRefField("bookmark_introduction", true)  // true = hyperlink
```

#### Page Format in Footer

```go
footer := doc.AddParagraph()
footer.AddText("Page ")
footer.AddPageField()
footer.AddText(" of ")
footer.AddNumPagesField()
footer.Justification("center")
```

---

### Styles and Formatting

#### Text Formatting in Run

```go
run := para.AddText("Text with multiple formats")

// Color
run.Color("FF0000")  // Red (hexadecimal format)

// Size (in half-points: 24 = 12pt, 28 = 14pt, 32 = 16pt)
run.Size("28")

// Bold
run.Bold()

// Italic
run.Italic()

// Underline
run.Underline("single")      // single, double, dotted, dash, wave, etc.

// Highlight
run.Highlight("yellow")      // yellow, green, cyan, magenta, blue, red

// Strikethrough
run.Strike(true)

// Shading
run.Shade("clear", "auto", "FFFF00")  // val, color, fill

// Character spacing (in twips)
run.Spacing(100)

// Font
run.Font("Arial", "Arial", "Arial", "default")
```

#### Predefined Paragraph Styles

```go
para.Style("Heading1")    // Heading 1 (16pt, blue, bold)
para.Style("Heading2")    // Heading 2 (13pt, blue, bold)
para.Style("Heading3")    // Heading 3 (12pt, dark blue, bold)
para.Style("Heading4")    // Heading 4 (11pt, blue, bold+italic)
para.Style("Normal")      // Normal style
```

#### List Numbering

```go
// Set numbering
para.NumPr("1", "0")  // numID, ilvl (list level)

// Set numbering font
para.NumFont("Arial", "Arial", "Arial", "default")

// Set numbering size
para.NumSize("24")  // 12pt
```

---

### Shapes and Canvas

#### Add Inline Shape

```go
// Dimensions in EMUs
width := int64(914400 * 2)   // 2 inches
height := int64(914400 * 1)  // 1 inch

// Create border line
line := &docx.ALine{
    W: 12700,  // Line width
    SolidFill: &docx.ASolidFill{
        SchemeClr: &docx.ASchemeColor{Val: "accent1"},
    },
}

run := para.AddInlineShape(
    width,
    height,
    "Rectangle",     // Name
    "auto",          // Black and white mode
    "rect",          // Shape type (rect, ellipse, triangle, etc.)
    line,            // Line configuration
)
```

#### Available Shapes

Predefined shape types (`prst` parameter):
- `"rect"`: Rectangle
- `"ellipse"`: Ellipse/Circle
- `"triangle"`: Triangle
- `"rtTriangle"`: Right Triangle
- `"parallelogram"`: Parallelogram
- `"trapezoid"`: Trapezoid
- `"diamond"`: Diamond
- `"pentagon"`: Pentagon
- `"hexagon"`: Hexagon
- `"octagon"`: Octagon
- `"star5"`: 5-pointed Star
- `"plus"`: Plus Sign
- `"arrow"`: Arrow

---

## 💡 Practical Examples

### Example 1: Complete Document with TOC

```go
package main

import (
    "os"
    "github.com/fumiama/go-docx"
)

func main() {
    // Create document
    doc := docx.New().WithDefaultTheme()
    
    // Document title
    title := doc.AddParagraph()
    title.AddText("Technical Report").Bold().Size("32").Color("2E75B6")
    title.Justification("center")
    
    doc.AddParagraph()  // Space
    
    // Add TOC
    opts := docx.DefaultTOCOptions()
    opts.Title = "Table of Contents"
    opts.Depth = 3
    doc.AddTOC(opts)
    
    // Page break
    doc.AddParagraph().AddPageBreaks()
    
    // Chapter 1
    h1 := doc.AddHeadingWithTOC("1. Introduction", 1, 1)
    h1.Style("Heading1")
    
    p1 := doc.AddParagraph()
    p1.AddText("This is the introduction content.")
    
    // Subchapter 1.1
    h11 := doc.AddHeadingWithTOC("1.1 Context", 2, 2)
    h11.Style("Heading2")
    
    p2 := doc.AddParagraph()
    p2.AddText("Relevant contextual information.")
    
    // Chapter 2
    h2 := doc.AddHeadingWithTOC("2. Methodology", 1, 3)
    h2.Style("Heading1")
    
    p3 := doc.AddParagraph()
    p3.AddText("Description of the methodology used.")
    
    // Footer with page numbers
    footer := doc.AddParagraph()
    footer.AddText("Page ")
    footer.AddPageField()
    footer.AddText(" of ")
    footer.AddNumPagesField()
    footer.Justification("center")
    
    // Important: add page size at the end
    doc.WithA4Page()
    
    // Save
    f, _ := os.Create("report.docx")
    defer f.Close()
    doc.WriteTo(f)
}
```

### Example 2: Document with Tables and Figures

```go
package main

import (
    "os"
    "github.com/fumiama/go-docx"
)

func main() {
    doc := docx.New().WithDefaultTheme()
    
    // Title
    title := doc.AddParagraph()
    title.AddText("Results Report").Bold().Size("32")
    title.Justification("center")
    
    doc.AddParagraph()
    
    // Data table
    table := doc.AddTable(3, 3, 0, nil)
    
    // Table headers
    headers := []string{"Name", "Value", "Unit"}
    for i, header := range headers {
        cell := table.TableRows[0].TableCells[i]
        cell.Shade("clear", "auto", "4472C4")
        p := cell.AddParagraph()
        p.AddText(header).Bold().Color("FFFFFF")
    }
    
    // Data
    data := [][]string{
        {"Temperature", "25.5", "°C"},
        {"Pressure", "101.3", "kPa"},
    }
    
    for i, row := range data {
        for j, val := range row {
            cell := table.TableRows[i+1].TableCells[j]
            p := cell.AddParagraph()
            p.AddText(val)
        }
    }
    
    // Table caption
    caption := doc.AddParagraph()
    caption.AddText("Table ")
    caption.AddSeqField("Table", "ARABIC")
    caption.AddText(": Experimental data")
    caption.Justification("center")
    
    doc.AddParagraph()
    
    // Add image (if exists)
    imgPara := doc.AddParagraph()
    imgPara.AddInlineDrawingFrom("chart.png")
    
    // Figure caption
    figCaption := doc.AddParagraph()
    figCaption.AddText("Figure ")
    figCaption.AddSeqField("Figure", "ARABIC")
    figCaption.AddText(": Results chart")
    figCaption.Justification("center")
    
    doc.WithA4Page()
    
    f, _ := os.Create("report.docx")
    defer f.Close()
    doc.WriteTo(f)
}
```

### Example 3: Formatted Bullet List

```go
package main

import (
    "os"
    "github.com/fumiama/go-docx"
)

func main() {
    doc := docx.New().WithDefaultTheme()
    
    title := doc.AddParagraph()
    title.AddText("Feature List").Bold().Size("28")
    
    doc.AddParagraph()
    
    // Custom bullet list
    items := []string{
        "High performance and scalability",
        "Easy to use and maintain",
        "Complete documentation",
        "Active community support",
    }
    
    for _, item := range items {
        bullet := doc.AddParagraph()
        bullet.AddText("•  " + item)
        bullet.Indent(720, 0, 0)  // Indent 0.5 inches
    }
    
    doc.WithA4Page()
    
    f, _ := os.Create("list.docx")
    defer f.Close()
    doc.WriteTo(f)
}
```

### Example 4: Cross-References

```go
package main

import (
    "os"
    "github.com/fumiama/go-docx"
)

func main() {
    doc := docx.New().WithDefaultTheme()
    
    // Section 1 with bookmark
    h1 := doc.AddParagraph()
    h1.AddText("1. Introduction").Bold().Size("28")
    h1.AddBookmark("intro")
    
    p1 := doc.AddParagraph()
    p1.AddText("This is the introduction section.")
    
    doc.AddParagraph()
    
    // Section 2 with reference to section 1
    h2 := doc.AddParagraph()
    h2.AddText("2. Development").Bold().Size("28")
    
    p2 := doc.AddParagraph()
    p2.AddText("As mentioned in section ")
    p2.AddRefField("intro", true)  // Reference with hyperlink
    p2.AddText(", it is important to consider...")
    
    // Reference with page number
    p3 := doc.AddParagraph()
    p3.AddText("See page ")
    p3.AddPageRefField("intro", false)
    p3.AddText(" for more details.")
    
    doc.WithA4Page()
    
    f, _ := os.Create("references.docx")
    defer f.Close()
    doc.WriteTo(f)
}
```

---

## 🧪 How to Test and Run

### Method 1: Create a Go Program

1. **Create project directory:**
```bash
mkdir my-docx-project
cd my-docx-project
```

2. **Initialize Go module:**
```bash
go mod init my-project
```

3. **Install dependency:**
```bash
go get github.com/SlideLang/go-docx@latest
```

4. **Create `main.go` file** with any of the examples above.

5. **Run:**
```bash
go run main.go
```

6. **Open the generated `.docx` file** in Microsoft Word or LibreOffice.

### Method 2: Use Existing Tests

```bash
# Clone the repository
git clone https://github.com/SlideLang/go-docx.git
cd go-docx

# Run the demo test
go test -v -run TestEnhancedFeaturesDemo

# The file 'slidelang_enhanced_demo.docx' will be created in the current directory
```

### Method 3: Command Line Program

```bash
# Use the included example program
go run cmd/main/main.go -u
```

### Verify the Generated Document

1. **Open in Microsoft Word**
2. **Update fields:**
   - Right-click on TOC → "Update field" → "Update entire table"
   - Or press `Ctrl+A` (select all) → `F9` (update fields)
3. **View navigation pane:**
   - View → Navigation Pane
   - You'll see the heading structure
4. **Test hyperlinks:**
   - Click on TOC entries to navigate
5. **View field codes:**
   - Press `Alt+F9` to toggle between codes and results

---

## 🔥 Advanced Use Cases

### Case 1: Generate Complete Technical Documentation

```go
func GenerateTechnicalDocumentation(title string, sections []Section) error {
    doc := docx.New().WithDefaultTheme()
    
    // Cover page
    cover := doc.AddParagraph()
    cover.AddText(title).Bold().Size("48").Color("2E75B6")
    cover.Justification("center")
    
    date := doc.AddParagraph()
    date.AddText("Generated on 2025-10-22").Size("20")
    date.Justification("center")
    
    doc.AddParagraph().AddPageBreaks()
    
    // TOC
    opts := docx.DefaultTOCOptions()
    opts.Title = "Table of Contents"
    opts.Depth = 3
    doc.AddTOC(opts)
    
    doc.AddParagraph().AddPageBreaks()
    
    // Generate sections
    tocIndex := 1
    for i, section := range sections {
        // Section heading
        h := doc.AddHeadingWithTOC(
            fmt.Sprintf("%d. %s", i+1, section.Title),
            1,
            tocIndex,
        )
        h.Style("Heading1")
        tocIndex++
        
        // Content
        p := doc.AddParagraph()
        p.AddText(section.Content)
        
        // Subsections
        for j, subsection := range section.Subsections {
            h2 := doc.AddHeadingWithTOC(
                fmt.Sprintf("%d.%d %s", i+1, j+1, subsection.Title),
                2,
                tocIndex,
            )
            h2.Style("Heading2")
            tocIndex++
            
            p2 := doc.AddParagraph()
            p2.AddText(subsection.Content)
        }
        
        // Page break between sections
        if i < len(sections)-1 {
            doc.AddParagraph().AddPageBreaks()
        }
    }
    
    // Footer with page numbers
    footer := doc.AddParagraph()
    footer.AddText("Page ")
    footer.AddPageField()
    footer.AddText(" of ")
    footer.AddNumPagesField()
    footer.Justification("center")
    
    doc.WithA4Page()
    
    f, err := os.Create("technical_documentation.docx")
    if err != nil {
        return err
    }
    defer f.Close()
    
    _, err = doc.WriteTo(f)
    return err
}
```

### Case 2: Report Generator with Templates

```go
type ReportData struct {
    Title       string
    Author      string
    Date        string
    Summary     string
    Sections    []SectionData
    Tables      []TableData
    Images      []ImageData
}

func GenerateReport(data ReportData) error {
    doc := docx.New().WithDefaultTheme()
    
    // Document metadata
    // (this would require extending the API with core properties)
    
    // Cover page
    title := doc.AddParagraph()
    title.AddText(data.Title).Bold().Size("48")
    title.Justification("center")
    
    author := doc.AddParagraph()
    author.AddText("By: " + data.Author).Size("24")
    author.Justification("center")
    
    date := doc.AddParagraph()
    date.AddText(data.Date).Size("20")
    date.Justification("center")
    
    doc.AddParagraph().AddPageBreaks()
    
    // Executive summary
    summaryH := doc.AddParagraph()
    summaryH.AddText("Executive Summary").Bold().Size("32")
    
    summaryP := doc.AddParagraph()
    summaryP.AddText(data.Summary)
    
    doc.AddParagraph().AddPageBreaks()
    
    // TOC
    opts := docx.DefaultTOCOptions()
    opts.Title = "Contents"
    doc.AddTOC(opts)
    
    doc.AddParagraph().AddPageBreaks()
    
    // Sections
    tocIdx := 1
    for _, section := range data.Sections {
        h := doc.AddHeadingWithTOC(section.Title, 1, tocIdx)
        h.Style("Heading1")
        tocIdx++
        
        p := doc.AddParagraph()
        p.AddText(section.Content)
    }
    
    // Add tables
    for i, tableData := range data.Tables {
        table := doc.AddTable(
            len(tableData.Rows),
            len(tableData.Columns),
            0,
            nil,
        )
        
        // Fill table...
        
        caption := doc.AddParagraph()
        caption.AddText("Table ")
        caption.AddSeqField("Table", "ARABIC")
        caption.AddText(": " + tableData.Title)
    }
    
    // Add images
    for _, img := range data.Images {
        imgPara := doc.AddParagraph()
        imgPara.AddInlineDrawingFrom(img.Path)
        
        caption := doc.AddParagraph()
        caption.AddText("Figure ")
        caption.AddSeqField("Figure", "ARABIC")
        caption.AddText(": " + img.Title)
    }
    
    doc.WithA4Page()
    
    f, _ := os.Create("complete_report.docx")
    defer f.Close()
    doc.WriteTo(f)
    
    return nil
}
```

### Case 3: Markdown to DOCX Exporter

```go
func ConvertMarkdownToDOCX(markdownFile string, docxFile string) error {
    // Read and parse Markdown
    content, _ := os.ReadFile(markdownFile)
    
    doc := docx.New().WithDefaultTheme()
    
    // TOC
    opts := docx.DefaultTOCOptions()
    doc.AddTOC(opts)
    doc.AddParagraph().AddPageBreaks()
    
    // Parse and convert content
    // (this is pseudocode, you would need a Markdown parser)
    lines := strings.Split(string(content), "\n")
    tocIdx := 1
    
    for _, line := range lines {
        if strings.HasPrefix(line, "# ") {
            // H1
            h := doc.AddHeadingWithTOC(line[2:], 1, tocIdx)
            h.Style("Heading1")
            tocIdx++
        } else if strings.HasPrefix(line, "## ") {
            // H2
            h := doc.AddHeadingWithTOC(line[3:], 2, tocIdx)
            h.Style("Heading2")
            tocIdx++
        } else if strings.HasPrefix(line, "### ") {
            // H3
            h := doc.AddHeadingWithTOC(line[4:], 3, tocIdx)
            h.Style("Heading3")
            tocIdx++
        } else if strings.HasPrefix(line, "- ") {
            // List
            bullet := doc.AddParagraph()
            bullet.AddText("•  " + line[2:])
            bullet.Indent(720, 0, 0)
        } else if line != "" {
            // Normal paragraph
            p := doc.AddParagraph()
            // Process inline formatting (bold, italic, links)
            processInlineFormatting(p, line)
        }
    }
    
    doc.WithA4Page()
    
    f, _ := os.Create(docxFile)
    defer f.Close()
    doc.WriteTo(f)
    
    return nil
}
```

---

## 🔧 Troubleshooting

### Document won't open in Word

**Problem:** Word says the file is corrupted.

**Solutions:**
1. Make sure to call `WithA4Page()` at the end, after all content
2. Verify that all `Run` elements have a valid parent
3. Don't add empty `RunProperties` (can cause document rejection)

```go
// ❌ Incorrect
doc.WithA4Page()  // At the beginning
// ... add content

// ✅ Correct
// ... add content
doc.WithA4Page()  // At the end
```

### TOC doesn't update

**Problem:** TOC shows "Error! Bookmark not defined" or incorrect page numbers.

**Solutions:**
1. Make sure to add bookmarks to headings with `AddTOCBookmark()`
2. Use `AddHeadingWithTOC()` to create headings automatically
3. In Word, update the TOC: Right-click → Update field → Update entire table
4. Or press `Ctrl+A` then `F9` to update all fields

### Images don't display

**Problem:** Images appear as empty boxes.

**Solutions:**
1. Verify that the image file exists and is accessible
2. Make sure to use compatible formats (PNG, JPEG, GIF)
3. Verify that the file is not corrupted
4. Use correct absolute or relative paths

```go
// ✅ Absolute path
para.AddInlineDrawingFrom("/full/path/image.png")

// ✅ Relative path to execution directory
para.AddInlineDrawingFrom("./images/logo.png")
```

### Heading styles don't work

**Problem:** `para.Style("Heading1")` doesn't apply formatting.

**Solutions:**
1. Use `WithDefaultTheme()` when creating the document
2. Native styles are included in the default theme

```go
// ✅ Correct
doc := docx.New().WithDefaultTheme()
```

### Compilation errors

**Problem:** `cannot find package` or similar.

**Solutions:**
```bash
# Update dependencies
go mod tidy

# Clean cache
go clean -modcache

# Re-download
go get -u github.com/SlideLang/go-docx@latest
```

### Document is too large

**Problem:** The .docx file is too big.

**Solutions:**
1. Compress images before adding them
2. Use efficient image formats (JPEG for photos, PNG for graphics)
3. Avoid adding redundant content
4. Consider splitting the document into multiple files

---

## 📝 Best Practices

### 1. Order of Operations

```go
// ✅ Correct order
doc := docx.New()
doc.WithDefaultTheme()

// ... add all content

doc.WithA4Page()  // AT THE END
```

### 2. Error Handling

```go
// ✅ Always handle errors
f, err := os.Create("document.docx")
if err != nil {
    return fmt.Errorf("error creating file: %w", err)
}
defer f.Close()

_, err = doc.WriteTo(f)
if err != nil {
    return fmt.Errorf("error writing document: %w", err)
}
```

### 3. Code Reusability

```go
// Create helper functions for common operations
func addHeading(doc *docx.Docx, text string, level int, tocIdx int) {
    h := doc.AddHeadingWithTOC(text, level, tocIdx)
    h.Style(fmt.Sprintf("Heading%d", level))
}

func addParagraph(doc *docx.Docx, text string) {
    p := doc.AddParagraph()
    p.AddText(text)
}
```

### 4. Testing

```go
func TestGenerateDocument(t *testing.T) {
    doc := docx.New().WithDefaultTheme()
    
    // Add content
    doc.AddParagraph().AddText("Test")
    doc.WithA4Page()
    
    // Save to buffer for testing
    var buf bytes.Buffer
    _, err := doc.WriteTo(&buf)
    
    if err != nil {
        t.Fatalf("Error generating document: %v", err)
    }
    
    if buf.Len() == 0 {
        t.Error("Empty document")
    }
}
```

---

## 📚 Additional Resources

### Official Documentation

- [README.md](../README.md) - Project overview
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guide
- [PROJECT_STATUS.md](../PROJECT_STATUS.md) - Project status

### Specifications

- [Office Open XML (ECMA-376)](http://www.ecma-international.org/publications/standards/Ecma-376.htm)
- [Microsoft Word Field Codes](https://support.microsoft.com/en-us/office/field-codes-in-word)

### Examples

- [demo_test.go](../demo_test.go) - Complete feature demonstration
- [cmd/main/main.go](../cmd/main/main.go) - Command line program

### Community

- [GitHub Issues](https://github.com/SlideLang/go-docx/issues)
- [GitHub Discussions](https://github.com/SlideLang/go-docx/discussions)

---

## 🤝 Contributing

Found a bug or have an idea to improve the documentation?

1. Open an [Issue](https://github.com/SlideLang/go-docx/issues)
2. Submit a [Pull Request](https://github.com/SlideLang/go-docx/pulls)
3. Join the [Discussions](https://github.com/SlideLang/go-docx/discussions)

---

## 📄 License

AGPL-3.0 - See [LICENSE](../LICENSE) for details.

---

## 🙏 Credits

- **Original:** [gonfva/docxlib](https://github.com/gonfva/docxlib)
- **Upstream:** [fumiama/go-docx](https://github.com/fumiama/go-docx)
- **Enhanced by:** [SlideLang Team](https://github.com/SlideLang)

---

**Questions? Comments?**

Open an [Issue](https://github.com/SlideLang/go-docx/issues) or join the [Discussions](https://github.com/SlideLang/go-docx/discussions).

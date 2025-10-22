# go-docx - SlideLang Enhanced Fork

Professional DOCX generation library for Go with advanced features for technical documentation.

**Fork of**: [fumiama/go-docx](https://github.com/fumiama/go-docx) | **Original**: [gonfva/docxlib](https://github.com/gonfva/docxlib)

## ✨ SlideLang Enhancements

This fork adds professional document generation features needed for **DocLang** and **SlideLang** exporters:

- ✅ **Dynamic Table of Contents** with hyperlinks and page numbers
- ✅ **Bookmarks** for cross-references and TOC navigation
- ✅ **Field Codes** (TOC, PAGEREF, PAGE, NUMPAGES, REF, SEQ)
- ✅ **Native Heading Styles** (H1-H4) with proper outline levels
- ✅ **Auto-numbering** for figures, tables, and sections
- ✅ **Cross-references** between sections
- ✅ **Navigation Pane** support in Microsoft Word

### Why This Fork?

Original library is excellent for basic DOCX manipulation but lacks:
- No Table of Contents support
- No bookmark functionality
- No Word field codes
- No native heading styles with outline levels

This fork completes those gaps for **professional technical documentation**.

## Introduction

> As part of my work for [Basement Crowd](https://www.basementcrowd.com) and [FromCounsel](https://www.fromcounsel.com), we were in need of a basic library to manipulate (both read and write) Microsoft Word documents.
> 
> The difference with other projects is the following:
> - [UniOffice](https://github.com/unidoc/unioffice) is probably the most complete but it is also commercial (you need to pay). It also very complete, but too much for my needs.
> - [gingfrederik/docx](https://github.com/gingfrederik/docx) only allows to write.
> 
> There are also a couple of other projects [kingzbauer/docx](https://github.com/kingzbauer/docx) and [nguyenthenguyen/docx](https://github.com/nguyenthenguyen/docx)
> 
> [gingfrederik/docx](https://github.com/gingfrederik/docx) was a heavy influence (the original structures and the main method come from that project).
> 
> However, those original structures didn't handle reading and extending them was particularly difficult due to Go xml parser being a bit limited including a [6 year old bug](https://github.com/golang/go/issues/9519).
> 
> Additionally, my requirements go beyond the original structure and a hard fork seemed more sensible.
> 
> The plan is to evolve the library, so the API is likely to change according to my company's needs. But please do feel free to send patches, reports and PRs (or fork).
> 
> In the mean time, shared as an example in case somebody finds it useful.

## 🎯 Features

### Core Features (from upstream)
- [x] Parse and save document
- [x] Edit text (color, size, alignment, link, ...)
- [x] Edit picture
- [x] Edit table
- [x] Edit shape
- [x] Edit canvas
- [x] Edit group

### SlideLang Enhanced Features
- [x] **Dynamic Table of Contents** - Auto-generated with F9 update support
- [x] **Bookmarks** - Named references and TOC anchors
- [x] **Field Codes** - TOC, PAGEREF, PAGE, NUMPAGES, REF, SEQ
- [x] **Native Heading Styles** - H1-H4 with proper formatting and outline levels
- [x] **Cross-references** - Link between sections with automatic updates
- [x] **Auto-numbering** - Figures, tables, and equations
- [x] **Navigation Pane** - Word document structure support

## Quick Start
```bash
go run cmd/main/main.go -u
```
And you will see two files generated under `pwd` with the same contents as below.

<table>
	<tr>
		<td align="center"><img src="https://user-images.githubusercontent.com/41315874/223348099-4a6099d2-0fec-4e13-92a7-152c00bc6f6b.png"></td>
		<td align="center"><img src="https://user-images.githubusercontent.com/41315874/223349486-e78ac0f1-c879-4888-9110-ea4db2590241.png"></td>
	</tr>
	<tr>
		<td align="center">p1</td>
		<td align="center">p2</td>
	</tr>
</table>

## 📦 Installation

```bash
go get -d github.com/SlideLang/go-docx@latest
```

Or use in your `go.mod`:
```go
require github.com/SlideLang/go-docx v0.1.0-slidelang
```

## 🚀 Quick Start

### Professional Document with TOC

```go
package main

import (
	"os"
	"github.com/fumiama/go-docx"
)

func main() {
	// Create document with default theme
	doc := docx.New().WithDefaultTheme()
	
	// Add title
	title := doc.AddParagraph()
	title.AddText("Professional Document").Bold().Size("32").Color("2E75B6")
	title.Justification("center")
	
	// Add Table of Contents
	opts := docx.DefaultTOCOptions()
	opts.Title = "Table of Contents"
	opts.Depth = 3
	opts.PageNumbers = true
	opts.Hyperlinks = true
	doc.AddTOC(opts)
	
	// Page break after TOC
	doc.AddParagraph().AddPageBreaks()
	
	// Add content with headings
	h1 := doc.AddHeadingWithTOC("1. Introduction", 1, 1)
	h1.Style("Heading1")
	
	doc.AddParagraph().AddText("This is the introduction content...")
	
	h2 := doc.AddHeadingWithTOC("1.1 Background", 2, 2)
	h2.Style("Heading2")
	
	doc.AddParagraph().AddText("Background information...")
	
	// Add figure with auto-numbering
	fig := doc.AddParagraph()
	fig.AddText("Figure ")
	fig.AddSeqField("Figure", "ARABIC")
	fig.AddText(": System diagram")
	
	// Add page numbers in footer
	footer := doc.AddParagraph()
	footer.AddText("Page ")
	footer.AddPageField()
	footer.AddText(" of ")
	footer.AddNumPagesField()
	footer.Justification("center")
	
	// IMPORTANT: Add page size at the END
	doc.WithA4Page()
	
	// Save document
	f, _ := os.Create("professional.docx")
	defer f.Close()
	doc.WriteTo(f)
}
```
### Parse Document
```go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fumiama/go-docx"
)

func main() {
	readFile, err := os.Open("file2parse.docx")
	if err != nil {
		panic(err)
	}
	fileinfo, err := readFile.Stat()
	if err != nil {
		panic(err)
	}
	size := fileinfo.Size()
	doc, err := docx.Parse(readFile, size)
	if err != nil {
		panic(err)
	}
	fmt.Println("Plain text:")
	for _, it := range doc.Document.Body.Items {
		switch it.(type) {
		case *docx.Paragraph, *docx.Table: // printable
			fmt.Println(it)
		}
	}
}
```

## 📚 API Reference

### Table of Contents

```go
// Create TOC with options
opts := docx.DefaultTOCOptions()
opts.Title = "Contents"
opts.Depth = 3              // Levels H1-H3
opts.PageNumbers = true     // Show page numbers
opts.Hyperlinks = true      // Clickable links
doc.AddTOC(opts)

// Smart heading with TOC entry
h1 := doc.AddHeadingWithTOC("Section Title", level, tocId)
h1.Style("Heading1")
```

### Bookmarks

```go
// Add simple bookmark
para.AddBookmark("section_intro")

// Add TOC-format bookmark
para.AddTOCBookmark(1)  // Creates _Toc000000001

// Generate heading bookmark name
name := GenerateHeadingBookmark("Introduction")
```

### Field Codes

```go
// TOC field
para.AddTOCField(depth, hyperlinks, pageNumbers)

// Page reference
para.AddPageRefField("_Toc000000001", true)

// Page numbers
para.AddPageField()        // Current page
para.AddNumPagesField()    // Total pages

// Sequences (figures, tables)
para.AddSeqField("Figure", "ARABIC")
para.AddSeqField("Table", "ARABIC")

// Cross-references
para.AddRefField("bookmark_name", true)
```

### Heading Styles

```go
// Apply native Word styles
para.Style("Heading1")  // 16pt, Blue, Bold
para.Style("Heading2")  // 13pt, Blue, Bold
para.Style("Heading3")  // 12pt, Dark Blue, Bold
para.Style("Heading4")  // 11pt, Blue, Bold+Italic

// All include proper outlineLevel for TOC
```

### Paragraph Formatting

```go
// Indentation (in twips: 1440 = 1 inch, 720 = 0.5 inch)
para.Indent(720, 0, 0)      // Left indent only
para.Indent(0, 360, 0)      // First line indent
para.Indent(720, 0, 360)    // Hanging indent (bullets)

// Alignment
para.Justification("left")    // or "center", "right", "both"

// Example: Professional bullet list
bullet := doc.AddParagraph()
bullet.AddText("•  Item text")
bullet.Indent(720, 0, 0)  // 0.5 inch indent
```
## 🎓 Examples

See [`demo_test.go`](demo_test.go) for a comprehensive example with:
- Dynamic TOC
- Multiple heading levels
- Bookmarks and cross-references
- Field codes (PAGE, NUMPAGES, SEQ)
- Professional formatting

## 📖 Documentation

- [Initial Plan](docs/initial-plan.md) - Design and implementation roadmap
- [Enhancement Spec](GO_DOCX_ENHANCEMENTS.md) - Detailed feature specifications

## 🔧 Usage in Word

After generating your document:

1. **Update TOC**: Click on TOC → Press F9 → Select "Update entire table"
2. **View Structure**: View → Navigation Pane (shows heading hierarchy)
3. **Update Fields**: Select all (Ctrl+A) → F9 to update all fields
4. **View Field Codes**: Alt+F9 (toggle between codes and results)

## 🤝 Contributing

This is an active fork maintained for SlideLang/DocLang projects. We welcome contributions!

### Git Flow Workflow

We use a simplified Git Flow branching strategy:

- **`master`**: Stable releases only. Tagged with semantic versions (e.g., `v0.1.0-slidelang`)
- **`dev`**: Integration branch for testing features before release
- **Feature branches**: Short-lived branches for specific features or fixes

#### Contributing Process

1. **Fork** the repository to your GitHub account
2. **Clone** your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-docx.git
   cd go-docx
   ```
3. **Add upstream** remote (if not already added):
   ```bash
   git remote add upstream https://github.com/SlideLang/go-docx.git
   ```
4. **Create feature branch** from `dev`:
   ```bash
   git checkout dev
   git pull upstream dev
   git checkout -b feature/your-feature-name
   ```
5. **Make changes** and commit with descriptive messages:
   ```bash
   git add .
   git commit -m "feat: add support for STYLEREF field"
   ```
6. **Push** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```
7. **Open PR** to `dev` branch (not `master`)
8. **Wait for review** and address feedback
9. Once approved, maintainers will merge to `dev`
10. Periodically, `dev` is merged to `master` and tagged

#### Commit Message Convention

We follow conventional commits:
- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `test:` Test additions/changes
- `refactor:` Code refactoring
- `perf:` Performance improvements

#### What We're Looking For

- ✅ Bug fixes
- ✅ Performance improvements
- ✅ Additional field codes (STYLEREF, HYPERLINK, IF, etc.)
- ✅ Extended style support
- ✅ Test coverage improvements
- ✅ Documentation enhancements

## 📄 License

AGPL-3.0. See [LICENSE](LICENSE)

## 🙏 Credits

- Original: [gonfva/docxlib](https://github.com/gonfva/docxlib)
- Upstream: [fumiama/go-docx](https://github.com/fumiama/go-docx)
- Enhanced by: [SlideLang Team](https://github.com/SlideLang)

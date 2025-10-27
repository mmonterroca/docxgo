# Architecture Documentation

## Overview

`go-docx` is a Go library for reading and writing Microsoft Word `.docx` files (ECMA-376 Office Open XML format). This document explains the internal architecture, design decisions, and how the components work together.

## Table of Contents

1. [High-Level Architecture](#high-level-architecture)
2. [Core Components](#core-components)
3. [Data Flow](#data-flow)
4. [File Structure](#file-structure)
5. [Design Patterns](#design-patterns)
6. [Performance Considerations](#performance-considerations)
7. [Extension Points](#extension-points)

---

## High-Level Architecture

### DOCX File Format

A `.docx` file is actually a ZIP archive containing multiple XML files:

```
document.docx
├── [Content_Types].xml          # Defines content types
├── _rels/                       # Package relationships
│   └── .rels
├── docProps/                    # Document properties
│   ├── app.xml
│   └── core.xml
└── word/                        # Main document folder
    ├── document.xml             # Document content
    ├── styles.xml               # Style definitions
    ├── fontTable.xml            # Font table
    ├── header1.xml              # Headers (if present)
    ├── footer1.xml              # Footers (if present)
    ├── _rels/                   # Document relationships
    │   └── document.xml.rels
    ├── media/                   # Embedded media
    │   └── image1.png
    └── theme/
        └── theme1.xml
```

### Library Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         User Code                            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Public API Layer                        │
│  (docx.go, api*.go - New, Parse, AddParagraph, etc.)       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Domain Model Layer                        │
│  (struct*.go - Document, Paragraph, Run, Table, etc.)       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  Serialization Layer                         │
│         (pack.go, unpack.go - ZIP + XML handling)           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      File System                             │
│              (io.Reader, io.Writer, fs.FS)                   │
└─────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. Main Document Structure (`Docx`)

**File**: `docx.go`

```go
type Docx struct {
    Document     Document           // Main document content
    docRelation  Relationships      // Relationship tracking
    headers      map[HeaderFooterType]*Header
    footers      map[HeaderFooterType]*Footer
    sectPr       *SectPr           // Section properties
    media        []Media           // Embedded images/media
    mediaNameIdx map[string]int    // Media lookup index
    rID          uintptr           // Relationship ID counter
    imageID      uintptr           // Image ID counter
    docID        uintptr           // Document ID counter
    slowIDs      map[string]uintptr // Other ID counters
    slowIDsMu    sync.Mutex        // Thread safety
    template     string            // Template name
    tmplfs       fs.FS             // Template file system
    tmpfslst     []string          // Template file list
}
```

**Responsibilities**:
- Root container for all document elements
- Manages relationships between document parts
- Tracks IDs for various elements (images, bookmarks, etc.)
- Handles serialization to/from ZIP archives

**Design Note**: This struct has grown over time and exhibits some "God Object" characteristics. Future refactoring could split responsibilities:
- `RelationshipManager`: Handle relationships and IDs
- `MediaManager`: Manage media files
- `TemplateManager`: Handle templates

### 2. Document Model (`Document`, `Body`)

**File**: `structdoc.go`

```go
type Document struct {
    XMLName xml.Name `xml:"w:document"`
    XMLW    string   `xml:"xmlns:w,attr"`
    XMLR    string   `xml:"xmlns:r,attr,omitempty"`
    // ... other namespace declarations
    Body    Body     `xml:"w:body"`
}

type Body struct {
    Items []interface{}  // Paragraphs, Tables, SectPr, etc.
    file  *Docx         // Back-reference to parent
}
```

**Responsibilities**:
- Represents the `word/document.xml` structure
- Contains all document content (paragraphs, tables, etc.)
- Handles XML marshaling/unmarshaling

**Design Pattern**: Composite pattern - `Body.Items` can contain various element types

### 3. Content Elements

#### Paragraph (`Paragraph`)
**File**: `structpara.go`, `apipara.go`

```go
type Paragraph struct {
    Properties *ParagraphProperties
    Children   []interface{}  // Runs, Hyperlinks, Fields, etc.
    file       *Docx         // Back-reference
}
```

**Responsibilities**:
- Container for inline content (text runs, hyperlinks, fields)
- Formatting properties (alignment, indentation, spacing)
- Style application

#### Run (`Run`)
**File**: `structrun.go`, `apirun.go`

```go
type Run struct {
    RunProperties *RunProperties
    Children      []interface{}  // Text, Tab, Drawing, etc.
    file          *Docx
}
```

**Responsibilities**:
- Smallest unit of formatted text
- Character-level formatting (font, size, color, bold, italic, etc.)
- Contains actual text and inline elements

#### Table (`Table`)
**File**: `structtable.go`, `apitable.go`

```go
type Table struct {
    TableProperties *WTableProperties
    TableGrid       *WTableGrid
    TableRows       []*WTableRow
    file            *Docx
}
```

**Responsibilities**:
- Tabular data structure
- Row/column management
- Table-level formatting (borders, width, alignment)

### 4. Serialization Layer

#### Pack (`pack.go`)

```go
func (f *Docx) pack(zipWriter *zip.Writer) error
```

**Process**:
1. Create file map from template or defaults
2. Marshal document structures to XML
3. Generate dynamic `[Content_Types].xml`
4. Write all files to ZIP archive

**Key Design Decision**: Uses `marshaller` type with `io.WriterTo` interface for efficient streaming:

```go
type marshaller struct {
    data interface{}
}

func (m marshaller) WriteTo(w io.Writer) (int64, error) {
    io.WriteString(w, xml.Header)
    xml.NewEncoder(w).Encode(m.data)
    return 0, nil
}
```

#### Unpack (`unpack.go`)

```go
func unpack(zipReader *zip.Reader) (*Docx, error)
```

**Process**:
1. Read ZIP archive
2. Parse `document.xml.rels` for relationships
3. Unmarshal `word/document.xml` to `Document`
4. Load media files
5. Reconstruct `Docx` structure

### 5. API Layer

The library provides a fluent API for building documents:

```go
doc := docx.New()
para := doc.AddParagraph()
run := para.AddText("Hello")
run.Bold().Color("FF0000").Size("28")
```

**Files**: `api*.go`
- `apipara.go` - Paragraph operations
- `apirun.go` - Run operations
- `apitable.go` - Table operations
- `apifield.go` - Field codes (TOC, page numbers, etc.)
- `apibookmark.go` - Bookmarks
- `apitoc.go` - Table of Contents
- `apiheaderfooter.go` - Headers and footers

**Design Pattern**: Fluent Interface / Method Chaining

Each method returns `*T` to allow chaining:
```go
func (r *Run) Bold() *Run {
    r.ensureRunProperties()
    r.RunProperties.Bold = &Bold{}
    return r  // Enable chaining
}
```

---

## Data Flow

### Creating a New Document

```
User Code
    │
    ├─> docx.New()
    │       │
    │       ├─> newEmptyFile()
    │       │       │
    │       │       └─> Initialize Docx struct
    │       │           ├─> Set up default relationships
    │       │           ├─> Initialize ID counters
    │       │           └─> Load template or use defaults
    │       │
    │       └─> Return *Docx
    │
    ├─> doc.AddParagraph()
    │       │
    │       ├─> Create Paragraph with empty Children
    │       ├─> Append to Document.Body.Items
    │       └─> Return *Paragraph
    │
    ├─> para.AddText("text")
    │       │
    │       ├─> Create Run with Text child
    │       ├─> Append to Paragraph.Children
    │       └─> Return *Run
    │
    ├─> run.Bold().Size("28")
    │       │
    │       ├─> ensureRunProperties()
    │       ├─> Set RunProperties.Bold
    │       ├─> Set RunProperties.Size
    │       └─> Return *Run (chaining)
    │
    └─> doc.WriteTo(file)
            │
            ├─> pack(zipWriter)
            │       │
            │       ├─> Build file map
            │       ├─> Marshal XML structures
            │       ├─> Generate [Content_Types].xml
            │       └─> Write to ZIP
            │
            └─> Close ZIP writer
```

### Reading an Existing Document

```
User Code
    │
    ├─> docx.Parse(reader, size)
    │       │
    │       ├─> zip.NewReader(reader, size)
    │       │
    │       └─> unpack(zipReader)
    │               │
    │               ├─> Read document.xml.rels
    │               ├─> Unmarshal word/document.xml
    │               ├─> Load media files
    │               ├─> Reconstruct relationships
    │               └─> Return *Docx
    │
    └─> Iterate over doc.Document.Body.Items
            │
            ├─> Type switch on item
            │   ├─> *Paragraph
            │   ├─> *Table
            │   └─> *SectPr
            │
            └─> Access content
```

---

## File Structure

### Package Organization

```
go-docx/
├── docx.go              # Main entry point, Docx struct
├── constants.go         # Constants (NEW)
├── errors.go            # Error types (NEW)
├── validation.go        # Validation functions (NEW)
│
├── api*.go              # Public API methods
│   ├── apipara.go      # Paragraph operations
│   ├── apirun.go       # Run operations
│   ├── apitable.go     # Table operations
│   ├── apifield.go     # Field codes
│   ├── apibookmark.go  # Bookmarks
│   ├── apitoc.go       # Table of contents
│   └── apiheaderfooter.go # Headers/footers
│
├── struct*.go           # Data structures
│   ├── structdoc.go    # Document, Body
│   ├── structpara.go   # Paragraph
│   ├── structrun.go    # Run
│   ├── structtable.go  # Table
│   ├── structrel.go    # Relationships
│   ├── structsect.go   # Section properties
│   └── ...
│
├── pack.go              # Serialization (write)
├── unpack.go            # Deserialization (read)
├── helper.go            # Utility functions
├── id.go                # ID generation
├── media.go             # Media handling
├── theme.go             # Theme handling
├── empty.go             # Default template
├── fs.go                # Template file system
│
├── *_test.go            # Tests
├── validation_test.go   # Validation tests (NEW)
│
├── docs/
│   ├── ARCHITECTURE.md  # This file (NEW)
│   ├── API_DOCUMENTATION.md
│   └── initial-plan.md
│
├── examples/            # Example programs
└── xml/                 # Default templates
```

---

## Design Patterns

### 1. Fluent Interface / Method Chaining

**Where**: API layer (`api*.go`)

**Example**:
```go
para.AddText("Hello").Bold().Color("FF0000").Size("28")
```

**Benefits**:
- Intuitive, readable code
- Reduces intermediate variables
- Natural document building flow

**Trade-offs**:
- Cannot return errors during chaining
- Must validate at end (or panic)

### 2. Lazy Initialization

**Where**: Property setters

**Example**:
```go
func (r *Run) ensureRunProperties() {
    if r.RunProperties == nil {
        r.RunProperties = &RunProperties{}
    }
}

func (r *Run) Bold() *Run {
    r.ensureRunProperties()
    r.RunProperties.Bold = &Bold{}
    return r
}
```

**Benefits**:
- Memory efficient (only allocate when needed)
- Cleaner structs (no empty properties in XML)
- Avoids nil pointer panics

### 3. Composite Pattern

**Where**: Document structure (`Body.Items`, `Paragraph.Children`)

**Example**:
```go
type Body struct {
    Items []interface{}  // Can contain Paragraph, Table, SectPr
}

type Paragraph struct {
    Children []interface{}  // Can contain Run, Hyperlink, Field
}
```

**Benefits**:
- Flexible, extensible structure
- Matches XML hierarchy naturally
- Easy to add new element types

**Trade-offs**:
- Type assertions needed when accessing
- Less type safety
- Performance overhead of interface{}

### 4. Back-References

**Where**: All major structs (`file *Docx`)

**Example**:
```go
type Paragraph struct {
    Properties *ParagraphProperties
    Children   []interface{}
    file       *Docx  // Back-reference to parent document
}
```

**Purpose**:
- Access document-level resources (media, IDs)
- Register relationships
- Generate unique IDs

**Trade-offs**:
- Circular references (parent → child → parent)
- Complicates serialization
- Makes testing harder

**Future Improvement**: Replace with interface:
```go
type DocumentContext interface {
    IncreaseID(name string) uint
    AddMedia(m Media) string
    AddRelationship(rel Relationship) string
}
```

### 5. Template Method Pattern

**Where**: `pack()`, `unpack()`

**Example**:
```go
func (f *Docx) pack(zipWriter *zip.Writer) error {
    // 1. Build file map (extensible)
    files := make(map[string]io.Reader)
    
    // 2. Add template files
    // 3. Add document XML
    // 4. Add headers/footers
    // 5. Add media
    
    // 6. Write all files
    for path, r := range files {
        // ... write to ZIP
    }
}
```

**Benefits**:
- Consistent structure
- Easy to extend (add new file types)
- Clear separation of concerns

---

## Performance Considerations

### Memory Allocation

**Pre-allocation with Constants** (NEW):
```go
// Before
p := &Paragraph{
    Children: make([]interface{}, 0, 64),  // Magic number
}

// After
p := &Paragraph{
    Children: make([]interface{}, 0, DefaultParagraphCapacity),
}
```

**Benefits**:
- Reduces slice reallocation
- Tunable via constants
- Documents capacity choices

### String Conversions

**Removed `unsafe.Pointer`** (NEW):
```go
// Before (UNSAFE)
func BytesToString(b []byte) string {
    return *(*string)(unsafe.Pointer(&b))
}

// After (SAFE)
func BytesToString(b []byte) string {
    return string(b)
}
```

**Rationale**:
- Safety over micro-optimizations
- Profile before optimizing
- Modern Go compiler is smart

**Future Optimization**:
If profiling shows string conversion is a bottleneck:
- Use `strings.Builder` for concatenation
- Use `sync.Pool` for reusable buffers
- Avoid unnecessary copies with `io.Writer` interfaces

### ID Generation

**Current**: Simple counter with mutex
```go
func (f *Docx) IncreaseID(name string) uintptr {
    f.slowIDsMu.Lock()
    n := f.slowIDs[name]
    n++
    f.slowIDs[name] = n
    f.slowIDsMu.Unlock()
    return n
}
```

**Consideration**: Use `sync/atomic` for lock-free increments if contention becomes an issue

### XML Marshaling

**Streaming approach** in `pack()`:
```go
type marshaller struct {
    data interface{}
}

func (m marshaller) WriteTo(w io.Writer) (int64, error) {
    io.WriteString(w, xml.Header)
    return 0, xml.NewEncoder(w).Encode(m.data)
}
```

**Benefits**:
- Memory efficient (doesn't buffer entire XML)
- Direct write to ZIP archive
- Handles large documents

---

## Extension Points

### Adding New Element Types

1. **Define struct** in `struct*.go`:
```go
type MyElement struct {
    XMLName xml.Name `xml:"w:myElement"`
    Attr    string   `xml:"w:attr,attr"`
    Content string   `xml:",chardata"`
}
```

2. **Add to parent's `interface{}` slice**:
```go
paragraph.Children = append(paragraph.Children, &MyElement{...})
```

3. **Add marshaling** (automatic via xml tags)

4. **Add API method** in `api*.go`:
```go
func (p *Paragraph) AddMyElement(content string) *MyElement {
    elem := &MyElement{Content: content}
    p.Children = append(p.Children, elem)
    return elem
}
```

### Adding New Document Parts

Example: Adding endnotes

1. **Define structure**:
```go
type Endnotes struct {
    XMLName  xml.Name `xml:"w:endnotes"`
    Endnote  []Endnote
}
```

2. **Add to `Docx`**:
```go
type Docx struct {
    // ...
    endnotes *Endnotes
}
```

3. **Update `pack()`**:
```go
if f.endnotes != nil {
    files["word/endnotes.xml"] = marshaller{data: f.endnotes}
}
```

4. **Update `[Content_Types].xml` generation**:
```go
if f.endnotes != nil {
    buf.WriteString(`<Override PartName="/word/endnotes.xml" .../>`)
}
```

5. **Add relationship**:
```go
f.docRelation.Relationship = append(f.docRelation.Relationship,
    Relationship{
        ID: f.nextRelID(),
        Type: "http://.../endnotes",
        Target: "endnotes.xml",
    })
```

---

## Recent Improvements (v0.4.0)

### 1. Constants (`constants.go`)

**Added**:
- Capacity constants (DefaultParagraphCapacity, etc.)
- Measurement constants (TwipsPerInch, MaxIndentTwips, etc.)
- ID type constants (IDTypeImage, IDTypeDrawing, etc.)
- Alignment constants (AlignLeft, AlignCenter, etc.)
- Underline style constants

**Benefits**:
- No more magic numbers
- Self-documenting code
- Easy to adjust globally

### 2. Error Handling (`errors.go`)

**Added**:
- `DocxError` - Structured error type
- `ValidationError` - Validation-specific errors
- Common error constants (ErrNilDocument, ErrInvalidIndent, etc.)
- Helper functions (IsValidationError, IsDocxError)

**Benefits**:
- Better error messages
- Error type checking with `errors.Is()` and `errors.As()`
- Programmatic error handling

### 3. Validation (`validation.go`)

**Added**:
- `ValidateIndent()` - Check indent ranges
- `ValidateJustification()` - Validate alignment values
- `ValidateUnderline()` - Validate underline styles
- `ValidateColor()` - Basic color validation

**Benefits**:
- Catch errors early
- Prevent invalid documents
- Clear constraint messages

### 4. Code Reduction

**Added helper methods**:
- `Run.ensureRunProperties()`
- `Paragraph.ensureParagraphProperties()`
- `Paragraph.ensureRunProperties()`

**Result**: Removed ~100 lines of repetitive nil checks

### 5. Safety

**Removed `unsafe.Pointer`**:
- Replaced with safe standard conversions
- Added deprecation warnings
- Documented performance trade-offs

---

## Future Architecture Considerations

### 1. Separation of Concerns

**Current Issue**: `Docx` struct has too many responsibilities

**Proposed**:
```go
type Docx struct {
    Document    Document
    Relations   *RelationshipManager
    Media       *MediaManager
    IDs         *IDGenerator
    Template    *TemplateManager
}

type RelationshipManager struct {
    relationships Relationships
    nextID        uint
}

type MediaManager struct {
    items []Media
    index map[string]int
}

type IDGenerator struct {
    mu       sync.RWMutex
    counters map[string]uint
}
```

**Benefits**:
- Single Responsibility Principle
- Easier testing
- Clear boundaries

### 2. Interfaces for Decoupling

**Current**: Concrete types everywhere

**Proposed**:
```go
type Document interface {
    AddParagraph() Paragraph
    AddTable(rows, cols int) Table
    WriteTo(w io.Writer) (int64, error)
}

type Paragraph interface {
    AddText(text string) Run
    Style(name string) Paragraph
    // ...
}
```

**Benefits**:
- Testability (mocking)
- Alternative implementations
- Plugin architecture

### 3. Error Handling in Fluent API

**Current**: No errors in chains

**Options**:

**A. Builder with deferred validation**:
```go
para := doc.AddParagraph()
para.AddText("Hello").Bold().Color("invalid")
err := para.Build()  // Validates here
```

**B. Collect errors**:
```go
run := para.AddText("Hello").Bold().Color("invalid")
if run.Err() != nil {
    // Handle error
}
```

**C. Context with errors**:
```go
ctx := docx.NewContext()
para := ctx.AddParagraph()
para.AddText("Hello").Bold().Color("invalid")
if err := ctx.Err(); err != nil {
    // Handle accumulated errors
}
```

### 4. Versioning

**Recommendation**: Use Go module semantic versioning

```
v1.x.x - Current stable API
v2.x.x - Breaking changes (future refactoring)
```

**Migration path**:
- Keep v1 for compatibility
- Develop v2 in parallel
- Provide migration guide

---

## Testing Strategy

### Current Coverage: ~58%

**Well-covered**:
- Basic document operations
- Text formatting
- Tables
- Fields and bookmarks

**Needs improvement**:
- Edge cases in unmarshaling
- Error conditions
- Large document handling
- Concurrent access

### Recommended Tests

1. **Unit tests**: Individual functions
2. **Integration tests**: Full document creation → save → reload → verify
3. **Validation tests**: ✅ Added in `validation_test.go`
4. **Benchmarks**: Performance critical paths
5. **Fuzz tests**: XML parsing edge cases

---

## Conclusion

The `go-docx` library has evolved significantly with contributions from multiple authors. The recent refactoring (v0.4.0) improves safety, maintainability, and clarity while maintaining backward compatibility.

Key strengths:
- Comprehensive OOXML support
- Fluent, intuitive API
- Active development

Areas for future improvement:
- Architectural refactoring (separate concerns)
- Better error handling
- Interface-based design
- Performance optimizations

This architecture document will be updated as the library evolves.

---

**Last Updated**: October 24, 2025  
**Version**: 0.4.0  
**Contributors**: Misael Monterroca, fumiama, gonfva, and community

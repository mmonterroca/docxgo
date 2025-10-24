# go-docx v2.0 - Clean Architecture Design

**Status**: 🚧 Planning Phase  
**Target**: Q1 2026  
**Breaking Changes**: Yes (major version bump)

---

## 🎯 Goals

### Primary Objectives
1. **Clean Architecture** - Proper separation of concerns
2. **Interface-Based Design** - Testable and extensible
3. **Better Error Handling** - Errors in fluent API
4. **Type Safety** - Less `interface{}`, more concrete types
5. **Performance** - Optimized for real-world usage

### Non-Goals
- Backward compatibility with v1.x (breaking changes allowed)
- Supporting every OOXML feature (focus on 80% use cases)

---

## 🏗️ Architecture Overview

### Layered Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Application Layer                     │
│              (User code using the library)               │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                      API Layer (v2)                      │
│         Builder Pattern + Fluent Interface               │
│     document.Builder, paragraph.Builder, etc.            │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                     Domain Layer                         │
│   Interfaces: Document, Paragraph, Run, Table, etc.     │
│   Implementations: docxDocument, docxParagraph, etc.     │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                  Infrastructure Layer                    │
│    Managers: RelationshipManager, MediaManager, etc.    │
│    Services: IDGenerator, Validator, Serializer         │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                    Persistence Layer                     │
│         pack.go, unpack.go (ZIP + XML handling)         │
└─────────────────────────────────────────────────────────┘
```

---

## 📦 Package Structure

```
github.com/SlideLang/go-docx/v2/
├── docx.go                     # Main entry point
├── builder.go                  # Builder pattern implementation
├── options.go                  # Functional options
│
├── domain/                     # Domain models (interfaces)
│   ├── document.go            # Document interface
│   ├── paragraph.go           # Paragraph interface
│   ├── run.go                 # Run interface
│   ├── table.go               # Table interface
│   ├── style.go               # Style interface
│   └── field.go               # Field interface
│
├── internal/                   # Internal implementations
│   ├── core/                  # Core domain implementations
│   │   ├── document.go        # docxDocument struct
│   │   ├── paragraph.go       # docxParagraph struct
│   │   ├── run.go             # docxRun struct
│   │   └── table.go           # docxTable struct
│   │
│   ├── manager/               # Resource managers
│   │   ├── relationship.go    # RelationshipManager
│   │   ├── media.go           # MediaManager
│   │   ├── id.go              # IDGenerator
│   │   └── style.go           # StyleManager
│   │
│   ├── service/               # Services
│   │   ├── validator.go       # Validation service
│   │   ├── serializer.go      # XML serialization
│   │   └── template.go        # Template service
│   │
│   └── ooxml/                 # OOXML structures (internal)
│       ├── wdocument.go       # w:document
│       ├── wparagraph.go      # w:p
│       ├── wrun.go            # w:r
│       └── wtable.go          # w:tbl
│
├── pkg/                        # Public utilities
│   ├── constants/             # Constants
│   │   ├── alignment.go
│   │   ├── underline.go
│   │   └── measurements.go
│   │
│   └── errors/                # Error types
│       └── errors.go
│
└── examples/                   # Examples for v2
    ├── basic/
    ├── advanced/
    └── migration_from_v1/
```

---

## 🎨 Core Design Patterns

### 1. Interface-Based Design

```go
// domain/document.go
package domain

import "io"

// Document represents a Word document
type Document interface {
    // Builder methods
    AddParagraph() (ParagraphBuilder, error)
    AddTable(rows, cols int) (TableBuilder, error)
    AddSection() (SectionBuilder, error)
    
    // Query methods
    Paragraphs() []Paragraph
    Tables() []Table
    Sections() []Section
    
    // Persistence
    WriteTo(w io.Writer) (int64, error)
    SaveAs(path string) error
    
    // Validation
    Validate() error
}

// Paragraph represents a paragraph
type Paragraph interface {
    AddRun() (RunBuilder, error)
    AddField(fieldType FieldType) (FieldBuilder, error)
    
    // Properties
    Style() Style
    SetStyle(name string) error
    Alignment() Alignment
    SetAlignment(align Alignment) error
    
    // Content
    Text() string
    Runs() []Run
}

// Run represents a text run
type Run interface {
    SetText(text string) error
    Text() string
    
    // Formatting
    SetBold(bold bool) error
    IsBold() bool
    SetColor(color Color) error
    GetColor() Color
    SetFont(font Font) error
    GetFont() Font
}
```

### 2. Builder Pattern with Validation

```go
// builder.go
package docx

// DocumentBuilder builds a document with validation
type DocumentBuilder struct {
    doc    domain.Document
    errors []error
}

// NewDocument creates a new document builder
func NewDocument(opts ...Option) *DocumentBuilder {
    config := defaultConfig()
    for _, opt := range opts {
        opt(config)
    }
    
    return &DocumentBuilder{
        doc: internal.NewDocxDocument(config),
    }
}

// AddParagraph adds a paragraph and returns builder for chaining
func (b *DocumentBuilder) AddParagraph() *ParagraphBuilder {
    para, err := b.doc.AddParagraph()
    if err != nil {
        b.errors = append(b.errors, err)
        return &ParagraphBuilder{err: err} // Error propagates
    }
    return &ParagraphBuilder{para: para, parent: b}
}

// Build validates and returns the document
func (b *DocumentBuilder) Build() (domain.Document, error) {
    if len(b.errors) > 0 {
        return nil, fmt.Errorf("document has %d errors: %w", len(b.errors), b.errors[0])
    }
    
    if err := b.doc.Validate(); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    return b.doc, nil
}

// ParagraphBuilder builds a paragraph
type ParagraphBuilder struct {
    para   domain.Paragraph
    parent *DocumentBuilder
    err    error
}

// Text adds text to the paragraph
func (pb *ParagraphBuilder) Text(text string) *ParagraphBuilder {
    if pb.err != nil {
        return pb // Propagate error
    }
    
    run, err := pb.para.AddRun()
    if err != nil {
        pb.err = err
        pb.parent.errors = append(pb.parent.errors, err)
        return pb
    }
    
    if err := run.SetText(text); err != nil {
        pb.err = err
        pb.parent.errors = append(pb.parent.errors, err)
    }
    
    return pb
}

// Bold makes the text bold
func (pb *ParagraphBuilder) Bold() *ParagraphBuilder {
    if pb.err != nil {
        return pb
    }
    
    // Get last run and make it bold
    runs := pb.para.Runs()
    if len(runs) == 0 {
        pb.err = errors.New("no runs to make bold")
        return pb
    }
    
    if err := runs[len(runs)-1].SetBold(true); err != nil {
        pb.err = err
        pb.parent.errors = append(pb.parent.errors, err)
    }
    
    return pb
}

// End returns to document builder
func (pb *ParagraphBuilder) End() *DocumentBuilder {
    return pb.parent
}
```

### 3. Functional Options Pattern

```go
// options.go
package docx

type Config struct {
    DefaultFont      string
    DefaultFontSize  int
    PageSize         PageSize
    Margins          Margins
    StrictValidation bool
    Logger           Logger
}

type Option func(*Config)

// WithDefaultFont sets the default font
func WithDefaultFont(font string) Option {
    return func(c *Config) {
        c.DefaultFont = font
    }
}

// WithPageSize sets the page size
func WithPageSize(size PageSize) Option {
    return func(c *Config) {
        c.PageSize = size
    }
}

// WithStrictValidation enables strict validation
func WithStrictValidation() Option {
    return func(c *Config) {
        c.StrictValidation = true
    }
}

// Usage:
doc := docx.NewDocument(
    docx.WithDefaultFont("Calibri"),
    docx.WithPageSize(docx.A4),
    docx.WithStrictValidation(),
)
```

### 4. Dependency Injection

```go
// internal/core/document.go
package core

type docxDocument struct {
    // Injected dependencies
    relationMgr  manager.RelationshipManager
    mediaMgr     manager.MediaManager
    idGen        manager.IDGenerator
    styleMgr     manager.StyleManager
    validator    service.Validator
    serializer   service.Serializer
    
    // Domain data
    paragraphs   []domain.Paragraph
    tables       []domain.Table
    sections     []domain.Section
    
    // Configuration
    config       *Config
}

// NewDocxDocument creates a document with injected dependencies
func NewDocxDocument(config *Config) *docxDocument {
    return &docxDocument{
        relationMgr: manager.NewRelationshipManager(),
        mediaMgr:    manager.NewMediaManager(),
        idGen:       manager.NewIDGenerator(),
        styleMgr:    manager.NewStyleManager(),
        validator:   service.NewValidator(config.StrictValidation),
        serializer:  service.NewSerializer(),
        config:      config,
        paragraphs:  make([]domain.Paragraph, 0, 32),
        tables:      make([]domain.Table, 0, 8),
        sections:    make([]domain.Section, 0, 1),
    }
}
```

---

## 🔄 Migration from v1 to v2

### Example: v1 Code

```go
// v1 (old)
import "github.com/fumiama/go-docx"

doc := docx.New()
para := doc.AddParagraph()
run := para.AddText("Hello")
run.Bold().Color("FF0000").Size("28")

file, _ := os.Create("output.docx")
doc.WriteTo(file)
```

### Example: v2 Code

```go
// v2 (new)
import docx "github.com/SlideLang/go-docx/v2"

// Builder pattern with error handling
doc := docx.NewDocument(
    docx.WithDefaultFont("Calibri"),
)

doc.AddParagraph().
    Text("Hello").
    Bold().
    Color(docx.Red).
    FontSize(14).
    End()

// Validate and build
finalDoc, err := doc.Build()
if err != nil {
    log.Fatal(err)
}

// Save
if err := finalDoc.SaveAs("output.docx"); err != nil {
    log.Fatal(err)
}
```

---

## 📊 Implementation Phases

### Phase 1: Foundation (Weeks 1-2)
- [ ] Set up v2 module (`go mod init github.com/SlideLang/go-docx/v2`)
- [ ] Define core interfaces (`domain/`)
- [ ] Create package structure
- [ ] Set up testing framework
- [ ] Design error handling strategy

### Phase 2: Core Domain (Weeks 3-4)
- [ ] Implement Document interface
- [ ] Implement Paragraph interface
- [ ] Implement Run interface
- [ ] Implement Table interface
- [ ] Add basic validation

### Phase 3: Managers (Weeks 5-6)
- [ ] Implement RelationshipManager
- [ ] Implement MediaManager
- [ ] Implement IDGenerator
- [ ] Implement StyleManager
- [ ] Add comprehensive tests

### Phase 4: Builders (Weeks 7-8)
- [ ] Implement DocumentBuilder
- [ ] Implement ParagraphBuilder
- [ ] Implement TableBuilder
- [ ] Add fluent API with error handling
- [ ] Integration tests

### Phase 5: Serialization (Weeks 9-10)
- [ ] Refactor pack/unpack
- [ ] Implement Serializer service
- [ ] OOXML generation
- [ ] Parsing existing documents
- [ ] Validation

### Phase 6: Advanced Features (Weeks 11-12)
- [ ] Headers/Footers (proper)
- [ ] TOC (improved)
- [ ] Fields (complete)
- [ ] Styles (comprehensive)
- [ ] Performance optimization

### Phase 7: Documentation & Release (Weeks 13-14)
- [ ] Complete API documentation
- [ ] Migration guide from v1
- [ ] Examples
- [ ] Benchmarks
- [ ] v2.0.0 release

---

## 🎯 Key Improvements Over v1

### 1. Type Safety
```go
// v1: interface{} everywhere
Children []interface{}

// v2: Concrete types
Runs []Run
Tables []Table
```

### 2. Error Handling
```go
// v1: No errors in fluent API
para.AddText("test").Bold().Color("invalid") // Silent failure

// v2: Errors propagate
doc.AddParagraph().
    Text("test").
    Bold().
    Color("invalid"). // Error recorded
    End()
    
doc, err := builder.Build() // Errors surface here
```

### 3. Testability
```go
// v2: Interface-based, easy to mock
type MockDocument struct {
    mock.Mock
}

func (m *MockDocument) AddParagraph() (ParagraphBuilder, error) {
    args := m.Called()
    return args.Get(0).(ParagraphBuilder), args.Error(1)
}
```

### 4. Separation of Concerns
```go
// v1: God Object
type Docx struct {
    Document, Relations, Media, IDs, Templates... // 15+ fields
}

// v2: Single Responsibility
type docxDocument struct {
    relationMgr RelationshipManager  // Handles relationships
    mediaMgr    MediaManager         // Handles media
    idGen       IDGenerator          // Handles IDs
}
```

### 5. Performance
```go
// v2: Optimizations
- sync.Pool for frequently allocated objects
- Lazy initialization where appropriate
- Efficient serialization with streaming
- Concurrent processing where safe
```

---

## 🧪 Testing Strategy

### Unit Tests
- Every interface has a mock
- Every implementation has tests
- 80%+ coverage target

### Integration Tests
- Full document creation → save → reload
- Validation in Microsoft Word
- Large document handling
- Concurrent access

### Benchmark Tests
```go
BenchmarkDocumentCreation
BenchmarkParagraphAddition
BenchmarkTableGeneration
BenchmarkSerialization
```

### Compatibility Tests
- v1 → v2 migration examples
- Roundtrip tests (create → save → load → verify)

---

## 📝 API Examples

### Basic Document
```go
doc := docx.NewDocument()

doc.AddParagraph().
    Text("Hello, World!").
    FontSize(16).
    Bold().
    Center().
    End()

finalDoc, _ := doc.Build()
finalDoc.SaveAs("hello.docx")
```

### Complex Document
```go
doc := docx.NewDocument(
    docx.WithDefaultFont("Arial"),
    docx.WithPageSize(docx.A4),
)

// Cover page
doc.AddParagraph().
    Text("My Report").
    Style("Heading 1").
    Center().
    End()

// Add table
doc.AddTable(3, 3).
    Row(0).Cell(0).Text("Header 1").Bold().End().
    Row(0).Cell(1).Text("Header 2").Bold().End().
    Row(1).Cell(0).Text("Data 1").End().
    End()

// Save
finalDoc, err := doc.Build()
if err != nil {
    return err
}
return finalDoc.SaveAs("report.docx")
```

### With Error Handling
```go
doc := docx.NewDocument()

doc.AddParagraph().
    Text("Test").
    Color(docx.Red).
    FontSize(999). // Invalid size - error recorded
    End()

_, err := doc.Build() // Errors surface here
if err != nil {
    var validationErr *docx.ValidationError
    if errors.As(err, &validationErr) {
        fmt.Printf("Validation failed: %v\n", validationErr)
    }
}
```

---

## 🚀 Success Criteria

### Must Have
- [ ] Clean, interface-based architecture
- [ ] Proper error handling throughout
- [ ] 80%+ test coverage
- [ ] Complete migration guide
- [ ] Performance ≥ v1
- [ ] Can open/edit files created by v1

### Nice to Have
- [ ] 90%+ test coverage
- [ ] Performance > v1 (10%+ improvement)
- [ ] Plugin system for custom elements
- [ ] Streaming API for large files

---

## 📅 Timeline

- **Planning**: October 2025 (Current)
- **Development**: November 2025 - January 2026
- **Testing**: February 2026
- **Release**: March 2026

---

## 🤝 Team Responsibilities

### Architecture Lead
- Define interfaces
- Review PRs
- Ensure consistency

### Domain Developer
- Implement core interfaces
- Write unit tests

### Infrastructure Developer
- Implement managers/services
- Performance optimization

### QA/Testing
- Write integration tests
- Validate in Microsoft Word
- Performance benchmarks

---

## 📚 References

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Proverbs](https://go-proverbs.github.io/)
- [OOXML Specification](http://www.ecma-international.org/publications/standards/Ecma-376.htm)

---

**Next Steps:**
1. ✅ Review this design doc
2. Create proof-of-concept
3. Define all interfaces
4. Start implementation

**Last Updated**: October 24, 2025  
**Status**: 📝 Design Phase

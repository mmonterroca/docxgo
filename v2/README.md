# go-docx v2

Clean architecture implementation of Microsoft Word .docx (OOXML) file manipulation in Go.

## Status: **Pre-Alpha Development** 🚧

v2 is a complete rewrite focusing on clean architecture, type safety, and testability.

**Current Version:** v2.0.0-alpha  
**Target Release:** Q1 2026

## What's New in v2

### ✨ Clean Architecture
- Interface-based design for all domain entities
- Dependency injection for managers
- Separation of concerns (domain, internal, pkg)
- No god objects or circular dependencies

### 🛡️ Type Safety
- No `interface{}` usage
- Proper error handling (no silent failures)
- Validation on all inputs
- Immutable return types (defensive copies)

### 🧪 Testability
- 100% mockable interfaces
- Comprehensive unit tests
- Benchmark suite
- Thread-safe managers

### 🚀 Performance
- Optimized memory allocations
- Thread-safe ID generation with `atomic`
- Pre-allocated slices with sensible defaults
- Lazy loading where appropriate

## Installation

```bash
go get github.com/SlideLang/go-docx/v2
```

## Quick Start

```go
package main

import (
    "github.com/SlideLang/go-docx/v2/domain"
    "github.com/SlideLang/go-docx/v2/internal/core"
)

func main() {
    // Create document
    doc := core.NewDocument()
    
    // Add paragraph
    para, _ := doc.AddParagraph()
    run, _ := para.AddRun()
    run.SetText("Hello, World!")
    run.SetBold(true)
    run.SetSize(24) // 12pt
    
    // Add table
    table, _ := doc.AddTable(3, 4)
    row, _ := table.Row(0)
    cell, _ := row.Cell(0)
    cellPara, _ := cell.AddParagraph()
    cellRun, _ := cellPara.AddRun()
    cellRun.SetText("Cell content")
    
    // Save (coming soon)
    // doc.SaveAs("output.docx")
}
```

## Architecture

```
v2/
├── domain/          # Core interfaces (public API)
│   ├── document.go  # Document interface
│   ├── paragraph.go # Paragraph interface
│   ├── run.go       # Run interface
│   ├── table.go     # Table interfaces
│   └── section.go   # Section interfaces
│
├── internal/        # Internal implementations
│   ├── core/        # Core domain implementations
│   │   ├── document.go
│   │   ├── paragraph.go
│   │   ├── run.go
│   │   └── table.go
│   └── manager/     # Service managers
│       ├── id.go           # Thread-safe ID generation
│       ├── relationship.go # Relationship management
│       └── media.go        # Media file management
│
├── pkg/             # Public utilities
│   ├── errors/      # Structured error types
│   ├── constants/   # OOXML constants
│   └── color/       # Color utilities
│
└── examples/        # Usage examples
    └── basic/       # Basic example
```

## Features

### ✅ Implemented
- Document creation
- Paragraph management with formatting
- Text runs with character formatting
- Tables with rows and cells
- Indentation (left, right, first line, hanging)
- Alignment (left, center, right, justify)
- Font formatting (bold, italic, underline, color, size)
- Hyperlinks
- Input validation
- Error handling
- Thread-safe managers

### 🚧 In Progress
- XML serialization
- File I/O (.docx reading/writing)
- Sections with headers/footers
- Styles management
- Fields (TOC, page numbers, etc.)

### 📋 Planned
- Images and drawings
- Comments and tracking
- Custom XML
- Advanced table features (merging, borders)
- Builder pattern for fluent API
- Migration guide from v1

## Design Principles

1. **Interface Segregation**: Small, focused interfaces
2. **Dependency Injection**: No global state
3. **Fail Fast**: Errors are returned immediately, not silently ignored
4. **Immutability**: Return defensive copies to prevent external mutation
5. **Type Safety**: Strong typing, no `interface{}`
6. **Thread Safety**: Concurrent access supported via mutexes and atomics
7. **Documentation**: Every public method documented

## Error Handling

```go
// Structured errors with context
para, err := doc.AddParagraph()
if err != nil {
    // Error contains operation, code, and context
    // Example: "operation=Document.AddParagraph | code=VALIDATION_ERROR | ..."
    return err
}

// Validation errors
err := run.SetSize(10000) // Invalid size
// Returns: ValidationError with field, value, and constraint details
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test -v ./internal/core
```

Current test coverage: **95%+**

## Performance

Optimizations include:
- Pre-allocated slices: `DefaultParagraphCapacity = 64`
- Thread-safe atomic counters for ID generation
- Lazy loading of relationships and media
- Efficient string building for text extraction

## Migration from v1

v2 is a breaking change. See [MIGRATION.md](docs/MIGRATION.md) (coming soon) for details.

Key differences:
- All methods return errors (no silent failures)
- Interface-based API (dependency injection)
- No global document state
- Validation on all inputs
- Different package structure

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md)

## License

AGPL-3.0 License - see [LICENSE](../../LICENSE)

## Roadmap

- **Phase 1 (Complete)**: Foundation, interfaces, core implementations
- **Phase 2 (In Progress)**: Managers, XML serialization
- **Phase 3**: File I/O, complete OOXML support
- **Phase 4**: Builder pattern, fluent API
- **Phase 5**: Beta release, migration guide, benchmarks

## Credits

Developed by [SlideLang](https://github.com/SlideLang)

v2 is a complete rewrite with lessons learned from v1, focusing on production-grade code quality.

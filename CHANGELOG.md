# Changelog - go-docx v2

All notable changes to v2 will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned for v2.1.0
- Complete Phase 10: Document Reading to 100%
- Read headers/footers from existing documents
- Read images from existing documents
- Comments and change tracking

---

## [2.0.0] - 2025-10-29

### 🎉 Production Release - v2.0.0 Stable

This is the first stable, production-ready release of go-docx v2!

### Added - Phase 10: Document Reading (NEW! 60% Complete)

#### Document Reading & Modification
- **OpenDocument()** - Open and read existing .docx files
- **Document parsing** - Parse document structure (paragraphs, runs, tables)
- **Content modification** - Edit existing text, formatting, and table cells
- **Style preservation** - Maintains all paragraph styles (Title, Subtitle, Heading1-9, Quote, Normal, ListParagraph)
- **Content addition** - Add new paragraphs, runs, and sections to existing documents
- **Round-trip workflow** - Create → Save → Open → Modify → Save
- **Example 12** - Complete read/modify example with documentation

#### Bug Fixes Since Beta
- **Fixed style preservation** - Paragraph styles now correctly preserved when reading documents (Oct 29, 2025)
  - Added `applyParagraphStyle()` function in `internal/reader/reconstruct.go`
  - Extracts `w:pStyle` from paragraph properties and applies via `para.SetStyle()`
  - All paragraph styles (Title, Subtitle, Heading1-3, Quote, Normal, ListParagraph) now working
- **Enhanced example 12** - Now demonstrates both editing existing content AND adding new content

#### Documentation Improvements Since Beta
- **Fixed README Quick Start** - Corrected all API examples to show accurate signatures
  - Separated Simple API vs Builder API clearly
  - Fixed `WithDefaultFont()` and `WithDefaultFontSize()` signatures
  - Corrected page size constant (`docx.A4` not `docx.PageSizeA4`)
  - Fixed alignment constant (`domain.AlignmentCenter` not `docx.AlignmentCenter`)
  - Added new Option 3: Read and Modify Existing Documents
- **Updated example count** - All documentation now references 11 working examples (was 9)
- **Updated Phase 10 status** - Marked as "60% complete - core features working" (was "Not Started")
- **Fixed error handling examples** - Corrected builder pattern usage in README

### Commits Since Beta (v2.0.0-beta...v2.0.0)
- `3a0832a` - docs: fix README Quick Start with correct API examples and update status
- `19c0e73` - Update documentation: Phase 10 now 60% complete with working reader
- `ca7f1f1` - Fix: Preserve paragraph styles when reading documents
- `3289b54` - Enhance example 12: Add content editing capabilities
- `852c3bf` - Add example 12: Read and modify documents
- `497a94a` - Fix strict OOXML validation and update docs to beta-ready status

### Summary of v2.0.0 Complete Features

#### Core Features from Beta (Carried Forward)
- **Core Interfaces** (domain package)
  - `Document` interface with metadata support
  - `Paragraph` interface with full formatting options
  - `Run` interface for character-level formatting
  - `Table`, `TableRow`, `TableCell` interfaces
  - `Section`, `Header`, `Footer` interfaces
  - `Style` and `Field` interfaces

- **Core Implementations** (internal/core)
  - `document` - Document implementation with validation
  - `paragraph` - Paragraph with alignment, indentation, spacing
  - `run` - Text run with bold, italic, color, font, size
  - `table`, `tableRow`, `tableCell` - Full table support

- **Service Managers** (internal/manager)
  - `IDGenerator` - Thread-safe ID generation with atomic counters
  - `RelationshipManager` - OOXML relationship management
  - `MediaManager` - Media file (images) management

- **Utilities** (pkg)
  - `errors` package - Structured error types
    - `DocxError` with operation and error codes
    - `ValidationError` for input validation
    - `BuilderError` for fluent API error propagation
  - `constants` package - OOXML constants
    - Measurement conversions (twips, EMUs, points)
    - Default capacities for performance
    - OOXML namespaces and relationship types
    - Content types and file paths
    - Validation limits
  - `color` package - Color utilities
    - `FromHex()` - Parse hex color strings
    - `ToHex()` - Convert to hex strings

- **Examples**
  - Basic example demonstrating v2 API usage
  - Shows paragraph, run, table creation
  - Demonstrates formatting and validation

- **Tests**
  - Comprehensive unit tests for core package
  - Tests for validation logic
  - Table operations tests
  - 95%+ code coverage

### Design Decisions
- **Clean Architecture**: Separation of domain (interfaces), internal (implementations), pkg (utilities)
- **Interface-Based**: All domain entities are interfaces for testability
- **Dependency Injection**: No global state, managers injected
- **Error Handling**: All methods return errors, no silent failures
- **Type Safety**: No `interface{}`, strong typing throughout
- **Thread Safety**: Managers use mutexes and atomic operations
- **Validation**: All inputs validated before use
- **Immutability**: Defensive copies returned from getters

### Performance Optimizations
- Pre-allocated slices with capacity hints
- Thread-safe atomic counters (no mutex overhead for IDs)
- Lazy loading of relationships and media
- Efficient string building for text extraction

### Breaking Changes from v1
- Complete API rewrite
- All methods now return errors
- Interface-based design (vs. concrete types)
- Different package structure
- No global document state
- Validation required on all inputs
- Different naming conventions

## [2.0.0-alpha] - 2025-01-XX

First alpha release of v2.

### Status
- ✅ Core domain interfaces defined
- ✅ Basic implementations working
- ✅ Service managers implemented
- ✅ Error handling system in place
- ✅ Constants and utilities ready
- ✅ Tests passing (95%+ coverage)
- 🚧 XML serialization (pending)
- 🚧 File I/O (pending)
- 🚧 Builder pattern (pending)

### Known Limitations
- Cannot save/load .docx files yet (serialization pending)
- Sections not fully implemented
- Fields (TOC, page numbers) not implemented
- Images/drawings not implemented
- No migration guide from v1 yet

### Next Steps (Phase 2)
- Implement XML serialization
- Add file I/O support
- Complete sections implementation
- Add builder pattern for fluent API
- Create migration guide

---

## Version History

- v2.0.0-alpha: Initial pre-release (Phase 1 complete)
- v1.0.0: Legacy version (separate branch)

## Migration from v1

v2 is a complete rewrite and is not backward compatible with v1.

**Key Differences:**

| v1 | v2 |
|----|-----|
| Concrete types | Interfaces |
| Silent failures | Error returns |
| Global state | Dependency injection |
| No validation | Full validation |
| Magic numbers | Named constants |
| `unsafe.Pointer` | Safe conversions |
| Chinese strings | English constants |

See full migration guide (coming soon): `docs/MIGRATION.md`

## Credits

- Original v1: Various contributors
- v2 Clean Architecture Rewrite: Misael Monterroca team (2025)

## License

AGPL-3.0 - See LICENSE file

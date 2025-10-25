# Changelog - go-docx v2

All notable changes to v2 will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added (Phase 1 - Foundation)
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
- v2 Clean Architecture Rewrite: SlideLang team (2025)

## License

AGPL-3.0 - See LICENSE file

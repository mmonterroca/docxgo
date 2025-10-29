# Credits & Project History

## Current Maintainer & Primary Author

**Misael Monterroca**  
Email: misael@monterroca.com  
GitHub: [@mmonterroca](https://github.com/mmonterroca)  
Role: Project Lead, v2 Architect & Lead Developer

## Project Genealogy

This project has evolved through multiple stages, with each contributor adding significant value. We maintain this history to honor all contributions while clarifying the current state of the project.

### 🚀 Version 2.0 (2024-2025) - mmonterroca/docxgo

**Status**: Current Active Development  
**Repository**: https://github.com/mmonterroca/docxgo  
**License**: AGPL-3.0

#### Author
- **Misael Monterroca** - Complete architectural rewrite

#### Major Contributions

**Clean Architecture Implementation**
- Interface-based design with dependency injection
- Separation of concerns (domain, internal, pkg layers)
- Type-safe implementations (eliminated `interface{}` usage)
- Comprehensive error handling throughout the API

**Core Domain**
- Document, Paragraph, Run, Table, Section interfaces
- Builder pattern with fluent API
- Functional options for configuration
- Validation at every layer

**Infrastructure**
- Thread-safe managers (Relationship, Media, ID, Style)
- Atomic ID generation for concurrent access
- Serialization service with XML optimization
- Writer service for .docx file generation

**Testing & Quality**
- 95%+ test coverage
- Integration tests with real document validation
- Benchmark suite for performance tracking
- Mock implementations for all interfaces

**Performance Optimizations**
- Pre-allocated slices with sensible defaults
- Efficient memory management
- Lazy loading where appropriate
- Optimized string building

---

### 📦 Version 1.x Enhanced (2023-2024) - Misael Monterroca Fork

**Repository**: https://github.com/mmonterroca/docxgo (root code, now deprecated)  
**Based on**: fumiama/go-docx

#### Author
- **Misael Monterroca** - Professional document features

#### Major Enhancements

**Headers & Footers**
- `AddHeader()`, `AddFooter()` with types (default, first, even)
- `AddPageNumberFooter()` for automatic page numbering
- `AddDocumentTitleHeader()` for dynamic headers

**Hyperlinks & Fields**
- `AddHyperlinkField()` for external and internal links
- `AddStyleRefField()` for dynamic text from heading styles
- Proper relationship management for links

**Table of Contents**
- `AddTOC()` with comprehensive options
- Configurable depth, page numbers, hyperlinks
- Custom TOC titles and styling

**Paragraph Formatting**
- `Indent()` method for first-line, hanging, and left indents
- Proper OOXML indentation support (twips)

**Documentation**
- Comprehensive API documentation (1,393+ lines)
- Professional examples in `examples/v030_demo/`
- Contributing guidelines with Git Flow

**Testing**
- Extensive test coverage for new features
- Integration tests for complex documents

---

### 🔧 Version 1.x Original (2022-2023) - fumiama/go-docx

**Repository**: https://github.com/fumiama/go-docx  
**Based on**: gonfva/docxlib  
**License**: AGPL-3.0

#### Author
- **fumiama** - Expanded functionality

#### Major Contributions

**Core Features**
- Parse and save Word documents
- Edit text (color, size, alignment, links)
- Edit pictures with image handling
- Edit tables with complex structures
- Edit shapes and drawing objects
- Edit canvas elements
- Edit group objects

**Examples & Demos**
- Command-line demo in `cmd/main/`
- Quick start guide with code examples
- Visual documentation with screenshots

**Package Management**
- Proper Go module structure
- Dependency on `github.com/fumiama/imgsz` for image handling

---

### 📄 Version 0.x Original Library (2020-2022) - gonfva/docxlib

**Repository**: https://github.com/gonfva/docxlib  
**Author**: Gonzalo Fernández-Victorio  
**License**: AGPL-3.0

#### Original Purpose

Created for [Basement Crowd](https://www.basementcrowd.com) and [FromCounsel](https://www.fromcounsel.com) to provide basic Microsoft Word document manipulation in Go.

#### Original Contributions

**Foundation**
- Initial OOXML structure definitions
- Basic document parsing and writing
- Core paragraph and text run handling
- ZIP-based .docx file format support

**Design Philosophy** (from original README)
> "The difference with other projects is the following:
> - UniOffice is probably the most complete but it is also commercial (you need to pay)
> - gingfrederik/docx only allows to write"

**Inspiration**
- Heavily influenced by [gingfrederik/docx](https://github.com/gingfrederik/docx)
- Addressed limitations of Go's XML parser
- Solved specific needs beyond other available libraries

---

## Evolution Timeline

```
2020-2022: gonfva/docxlib
           └─ Basic OOXML manipulation
           └─ Foundation for future work

2022-2023: fumiama/go-docx (fork)
           └─ Enhanced with images, tables, shapes
           └─ Expanded API surface

2023-2024: mmonterroca/docxgo v1 (fork)
           └─ Professional features (headers, TOC, links)
           └─ Comprehensive documentation

2024-2025: mmonterroca/docxgo v2 (complete rewrite)
           └─ Clean architecture
           └─ Production-grade code quality
           └─ Independent project
```

---

## Why v2 is Independent

### Reasons for Independence

1. **Complete Architectural Rewrite**
   - No shared code patterns with v1
   - Different design principles (clean architecture)
   - Breaking changes throughout

2. **Original Fork Inactive**
   - `fumiama/go-docx` has had no updates in months
   - No PR merges from other forks
   - Community fragmentation

3. **Significant Divergence**
   - v2 is ~70% new code
   - Different package structure
   - Different API design
   - Different error handling philosophy

4. **Namespace Clarity**
   - Users need clear distinction between versions
   - Original namespace doesn't reflect current reality
   - Misael Monterroca organization as proper owner

### Attribution Philosophy

We maintain **full transparency** about project history:
- Original authors credited in LICENSE
- This CREDITS.md preserved indefinitely
- AGPL-3.0 license maintained
- Fork history acknowledged in documentation

---

## Related Projects

### Alternatives in the Go Ecosystem

- **[UniOffice](https://github.com/unidoc/unioffice)** - Commercial, comprehensive
- **[gingfrederik/docx](https://github.com/gingfrederik/docx)** - Write-only
- **[kingzbauer/docx](https://github.com/kingzbauer/docx)** - Alternative approach
- **[nguyenthenguyen/docx](https://github.com/nguyenthenguyen/docx)** - Different implementation

### Why Choose mmonterroca/docxgo v2?

- ✅ **Open source** (AGPL-3.0) - no commercial license needed
- ✅ **Clean architecture** - testable, maintainable, extensible
- ✅ **Both read and write** - parse existing + create new documents
- ✅ **Type safe** - no `interface{}`, proper error handling
- ✅ **Well tested** - 95%+ coverage, integration tests
- ✅ **Active development** - regular updates, responsive maintenance
- ✅ **Comprehensive docs** - examples, guides, API reference

---

## Acknowledgments

### Special Thanks

- **Gonzalo Fernández-Victorio** - For creating the foundation that made this work possible
- **fumiama** - For expanding the feature set and maintaining an active fork
- **The Go Team** - For an excellent language and standard library
- **ECMA-376 Authors** - For the OOXML specification

### Community

We welcome contributions from the community. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## License

This project is licensed under **AGPL-3.0** (GNU Affero General Public License v3.0).

See [LICENSE](LICENSE) for full text.

### Copyright Notices

```
Copyright (C) 2024-2025 Misael Monterroca / Misael Monterroca (v2 architecture & development)
Copyright (C) 2022-2024 fumiama (v1 enhancements)
Copyright (C) 2020-2022 Gonzalo Fernández-Victorio (original library)
```

---

## Contact

**Project Lead**: Misael Monterroca  
**Email**: misael@monterroca.com  
**GitHub**: https://github.com/mmonterroca/docxgo  
**Issues**: https://github.com/mmonterroca/docxgo/issues

---

*Last Updated: October 25, 2025*

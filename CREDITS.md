# Credits & Project History

## Current Maintainer & Primary Author

**Misael Monterroca**  
Email: misael@monterroca.com  
GitHub: [@mmonterroca](https://github.com/mmonterroca)  
Role: Project Lead, v2 Architect & Lead Developer

## Project Genealogy

This project has evolved through multiple stages, with each contributor adding significant value. We maintain this history to honor all contributions while clarifying the current state of the project.

Listed newest first. See the timeline below for the chronological view.

### 🚀 Version 2.x (2025-2026) - mmonterroca/docxgo

**Status**: Current Active Development  
**Repository**: https://github.com/mmonterroca/docxgo  
**License**: MIT

#### Author
- **Misael Monterroca** - Complete architectural rewrite

#### Major Contributions

**Clean Architecture Implementation**
- Interface-based design with dependency injection
- Separation of concerns (domain, internal, pkg layers)
- Type-safe public and domain APIs; dynamic values remain only where OOXML polymorphism requires them
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
- Comprehensive unit and round-trip test suite
- Integration tests with real document validation
- Benchmark suite for performance tracking
- Mock implementations for all interfaces

**Performance Optimizations**
- Pre-allocated slices with sensible defaults
- Efficient memory management
- Lazy loading where appropriate
- Optimized string building

---

### 📦 Version 1.x — mmonterroca fork (2025)

**Repository**: https://github.com/mmonterroca/docxgo (root code, now deprecated)  
**Forked from**: the 2025-05 `fumiama/go-docx` HEAD. The upstream history is preserved as repository ancestry — see below.

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

### 🔧 Version 1.x — fumiama/go-docx, upstream (2023-2025)

**Repository**: https://github.com/fumiama/go-docx  
**Forked from**: gonfva/docxlib (MIT)  
**License**: AGPL-3.0 — relicensed by fumiama on 2023-02-24, having received the code under MIT

docxgo v2's Git history descends from this repository through its 2025-05 HEAD,
including the AGPL period. The completed
[provenance audit](docs/PROVENANCE_AUDIT.md) determines that the current
rewritten release contains no protectable AGPL implementation; historical
snapshots retain their original license.

#### Author
- **fumiama** - Expanded functionality (first commit 2023-02-08, last 2025-05-06)

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

### 📄 Version 0.x Original Library (2021-2022) - gonfva/docxlib

**Repository**: https://github.com/gonfva/docxlib  
**Author**: Gonzalo Fernández-Victorio, for Basement Crowd Ltd  
**Forked from**: gingfrederik/docx (2020, MIT)  
**License**: MIT

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
2020       gingfrederik/docx                                   [MIT]
             └─ Original library

2021-2022  gonfva/docxlib  (fork)                              [MIT]
             └─ Basic OOXML manipulation, ZIP/XML foundation

2023       fumiama/go-docx  (fork)      ← relicensed AGPL-3.0 on 2023-02-24
             └─ Images, tables, shapes; expanded API surface
             └─ development continued through 2025-05
                    │
                    │  docxgo's history continues from the 2025-05 HEAD;
                    │  the separate v2 implementation begins in 2025-10
                    ▼
2025-2026  mmonterroca/docxgo v2  (separate implementation in the same history)
             └─ Clean-architecture rewrite; residual overlap is limited to
                standards-constrained declarations, generic idioms, and data
             └─ Distributed under MIT — see the licensing note below
```

docxgo v2 is distributed under the MIT license ([LICENSE](LICENSE)). Its Git
history includes AGPL-era `fumiama/go-docx` snapshots, while the current
implementation is a substantial rewrite. The completed
**[provenance audit](docs/PROVENANCE_AUDIT.md)** records the evidence and MIT
determination. Historical snapshots retain their original licenses.

---

## How v2 diverges from the upstream

These points summarize how docxgo v2 was rewritten from the historical
`fumiama/go-docx` base. The full evidence and licensing determination are in
[docs/PROVENANCE_AUDIT.md](docs/PROVENANCE_AUDIT.md).

1. **Substantial architectural rewrite**
   - Clean-architecture layering; different design principles
   - Breaking changes throughout
   - Residual overlap is standards-constrained or generic, as quantified by the audit

2. **Significant divergence**
   - Different package structure
   - Different API design
   - Different error handling philosophy

3. **Namespace clarity**
   - Users need a clear distinction between versions
   - Original namespace doesn't reflect current reality

### Attribution Philosophy

We maintain **full transparency** about project history:
- Original authors credited in LICENSE
- This CREDITS.md preserved indefinitely
- Fork history acknowledged in documentation
- Licensing provenance and determination recorded in [docs/PROVENANCE_AUDIT.md](docs/PROVENANCE_AUDIT.md)

---

## Related Projects

### Alternatives in the Go Ecosystem

- **[UniOffice](https://github.com/unidoc/unioffice)** - Commercial, comprehensive
- **[gingfrederik/docx](https://github.com/gingfrederik/docx)** - Write-only
- **[kingzbauer/docx](https://github.com/kingzbauer/docx)** - Alternative approach
- **[nguyenthenguyen/docx](https://github.com/nguyenthenguyen/docx)** - Different implementation

---

## Acknowledgments

### Special Thanks

- **gingfrederik** - For the original docx library
- **Gonzalo Fernández-Victorio** and **Basement Crowd Ltd** - For creating the foundation that made this work possible
- **fumiama** - For expanding the feature set and maintaining an active fork
- **The Go Team** - For an excellent language and standard library
- **ECMA-376 Authors** - For the OOXML specification

### Community

We welcome contributions from the community. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## License

docxgo v2 is distributed under the **MIT License** ([LICENSE](LICENSE)). Its
provenance relative to the AGPL-era `fumiama/go-docx` history and the completed
MIT determination are documented in
[docs/PROVENANCE_AUDIT.md](docs/PROVENANCE_AUDIT.md).

### Copyright Notices

```
Copyright (c) 2024-2026 Misael Monterroca (v2 architecture & development)
Copyright (c) 2023 Fumiama Minamoto (源文雨) (MIT-era predecessor notice)
Copyright (c) 2021 Gonzalo Fernandez-Victorio (original library)
Copyright (c) 2021 Basement Crowd Ltd (https://www.basementcrowd.com)
Copyright (c) 2020 gingfrederik (original library)
```

---

## Contact

**Project Lead**: Misael Monterroca  
**Email**: misael@monterroca.com  
**GitHub**: https://github.com/mmonterroca/docxgo  
**Issues**: https://github.com/mmonterroca/docxgo/issues

---

*Last Updated: July 2026*

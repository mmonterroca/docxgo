## v2.2.1 — 2026-01-22

### Fixed
- Hyperlink RelationshipID preservation during round-trip (read -> modify -> save)
- Drawing position serialization: fixed `wp:align` empty value error for floating images; added `UseOffsetX`/`UseOffsetY` flags
- Internal hyperlink anchors: support for anchor-based bookmarks and TOC links
- Custom styles preservation: all 55+ custom styles from templates maintained during round-trip

### Added
- Troubleshooting documentation (`docs/TROUBLESHOOTING_DOCX_VALIDATION.md`)

---

## v2.2.0 — 2026-01-06

### Fixed
- Table cell border serialization: ensure each side serializes with correct style ("single"), width (Sz=4), and color ("FF0000" hex)

### Tests
- Added `TestTableSerializer_CellBorders` validating per-side style, width, and color

### Acknowledgements
- Original fix: [PR #3](https://github.com/mmonterroca/docxgo/pull/3) by @g-mero
- Validation & extension: [PR #4](https://github.com/mmonterroca/docxgo/pull/4) by @Copilot

---

## v2.1.1 — 2025-11-04

### Fixed
- Go module path: re-applied `/v2` suffix to `go.mod` (regression from v2.1.0)

---

## v2.1.0 — 2025-10-31

### Added
- Theme system with 7 preset themes (Corporate, Startup, Modern, Fintech, Academic, TechPresentation, TechDarkMode)
- Theme colors, fonts, and spacing configuration
- Custom theme support (clone, customize, create from scratch)
- `WithTheme()` builder option for one-call theme application
- Example 13: themes showcase

### Fixed
- Style preservation when applying themes
- Font inheritance in themed documents
- Color serialization in OOXML

---

## v2.0.1 — 2025-11-04

### Fixed
- Go module path: added required `/v2` suffix to `go.mod` (`module github.com/mmonterroca/docxgo` -> `module github.com/mmonterroca/docxgo/v2`)

---

## v2.0.0 — 2025-10-29

### Added
- Document reading: `docx.OpenDocument()` to open existing .docx files
- Read document structure (paragraphs, runs, tables, styles)
- Modify existing content (edit text, change formatting, update table cells)
- Round-trip capability (Create -> Save -> Open -> Modify -> Save)
- Example 12: read and modify documents

### Fixed
- Style preservation (Title, Subtitle, Heading1-9, Quote, Normal, ListParagraph) during round-trip
- README API examples corrected for accurate signatures

---

## v2.0.0-beta — 2025-10-28

### Added
- Complete rewrite of go-docx with domain-driven clean architecture
- Builder pattern API with fluent chaining and error accumulation
- Direct domain API for programmatic control
- Section management with independent page settings
- Headers and footers (per-section, first-page, odd/even variants)
- Page layout (sizes, orientations, margins, columns)
- Fields support (TOC, page numbers, hyperlinks, dates, 9 field types)
- Advanced tables (cell merging, vertical alignment, shading, borders, 8 styles)
- Inline and floating images with precise positioning (9 formats)
- 40+ built-in Word styles
- Rich text formatting (bold, italic, underline, colors, fonts, sizes, highlights)
- Comprehensive error types with context and validation
- 11 working examples covering all major features
- Complete API documentation and migration guide from v1.x

### Fixed
- Character encoding issues in XML serialization
- Relationship ID management for images and hyperlinks
- Section break serialization order
- Table cell border rendering
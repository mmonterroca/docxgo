## v2.3.0 — 2026-02-27

### Added

- **Template / Mail Merge** (`pkg/template/`) — new package for document template processing
  - `MergeTemplate()` — replace `{{placeholder}}` tokens with data values, preserving formatting
  - `FindPlaceholders()` / `PlaceholderNames()` — detect all placeholders in body, tables, headers, footers
  - `ValidateTemplate()` — check for missing/unused data keys before merging
  - `ConsolidateRuns()` — merge adjacent runs with identical formatting to heal split placeholders
  - Custom delimiter support (e.g., `${key}`, `«key»` instead of `{{key}}`)
  - Strict mode for missing key error reporting
  - Batch merge support (reopen template for each record)
  - Full MERGEFIELD compatibility with real Microsoft Word templates
- `Paragraph.ClearRuns()`, `Paragraph.RemoveRun(index)`, `Paragraph.InsertRunAt(index)` — paragraph mutation APIs
- `Run.Fields()`, `Run.Breaks()`, `Run.Image()` — run content inspection methods
- `Run.ClearFields()` — clear field instructions from a run
- Example 14: Mail merge invoice template with batch generation
- Example 15: External Word template with MERGEFIELDs and «» delimiters

### Fixed

- `walkParagraphs()` auto-creating phantom headers/footers causing "Word found unreadable content" errors
- MERGEFIELD text duplication after merge due to field instructions persisting in preceding runs

---

## v2.2.2 — 2026-02-26

### Fixed

- Table style borders not rendering in Word: styles like `TableGrid` now emit proper `w:tblPr > w:tblBorders` in `styles.xml` (#15)
- Grid column width calculation: derive widths from first row cells instead of emitting `w:w="0"`

### Added

- `TableStyleDef` interface for table-specific style properties (borders, cell margins)
- `TableLevelBorders` struct with `InsideH`/`InsideV` for inner grid borders
- Example 13_themes restored with full v2 API (7 preset themes)

### Changed

- Examples 01_basic and 02_intermediate now use `Style(domain.TableStyleGrid)` for visible table borders
- Comprehensive documentation overhaul across all docs files (#14)
- CHANGELOG.md now contains complete version history (v2.0.0-beta through v2.2.2)

---

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

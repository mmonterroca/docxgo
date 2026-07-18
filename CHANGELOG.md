## v2.7.2 — 2026-07-18

### Compliance

- **Definitive MIT provenance determination.** `docs/PROVENANCE_AUDIT.md` now records the reproducible conclusion that the current v2 implementation may remain MIT: AGPL applies to the historical upstream snapshots in Git history, not to the current release tree. No project-wide AGPL relicense or further licensing consultation is an open release task.
- **Corrected copyright notices.** The root `LICENSE` preserves the exact MIT-era predecessor notices without implying that fumiama's later AGPL-era work was licensed as MIT. `CREDITS.md`, the README, package godoc, and design documentation now use the same determination.
- **License-complete artifacts.** Platform npm packages and GitHub binary archives now include the root MIT license text. This repairs the notice omission identified in the published v2.7.1 artifacts.

### Fixed

- **Release metadata and documentation** now consistently report v2.7.2 across the Go version constant, CLI examples, API/design/status guides, and npm package metadata.
- **Documentation import paths and branding** now use the public `/v2` module path and the `docxgo` project name.
- **Release automation comments and actions** now match the actual PAT-triggered npm publication path, and Go setup actions are aligned on `actions/setup-go@v7`.

### Compatibility

- No public Go, CLI protocol, or Node.js API changed. This is a backward-compatible compliance, packaging, and documentation patch.

---

## v2.7.1 — 2026-07-18

### Added

- **The two Tech themes are now discoverable through the public API.** `TechPresentation` and `TechDarkMode` existed but were not resolvable by name; they are now returned by `themes.AllThemes()` (which now yields **7** themes), `themes.ThemeNames()`, and `themes.GetTheme()`, and are accepted by the CLI's theme option and the Node.js `ThemeName` type (PR #56). Consumers that relied on `AllThemes()` returning exactly five themes should be aware of the new count.
- **`docs/PROVENANCE_AUDIT.md`** — a reproducible code-provenance record for docxgo v2 relative to the AGPL-licensed `fumiama/go-docx` upstream, with an accompanying comparison script (`docs/provenance/compare_line_overlap.py`).

### Fixed

- **Output branding.** Generated documents reported `go-docx/v2` (and a frozen `go-docx v2.0.0`) as the authoring application in `docProps/app.xml`/`core.xml`; they now report `docxgo`.

### Changed

- **Documentation aligned with actual behavior** (no API or generation changes): image support is stated as PNG/JPEG/GIF — the formats that decode end-to-end (see #55 for the previously over-stated set); thread-safety wording clarified (a single `Document` is not thread-safe, the internal managers are); the package godoc version string corrected; and the public README prose tightened. The license/provenance narrative across README, CREDITS, and the godoc now defers to `docs/PROVENANCE_AUDIT.md` as the single source of truth.

---

## v2.7.0 — 2026-07-17

### Added

- **Command-line interface (`cmd/docxgo`)** — a JSON-RPC binary that exposes docxgo over stdin/stdout, so documents can be created and manipulated from any language (Node.js, Python, shell, AWS Lambda), on any platform (Linux/macOS/Windows, arm64/x64), with zero config, ports, or auth (closes #19, PR #24)
  - Two modes: `exec` (one-shot, ideal for `child_process`) and `rpc` (persistent newline-delimited JSON session, ideal for batch/Lambda usage)
  - 22 methods across the `system.*`, `document.*`, `paragraph.*`, `table.*`, `section.*`, and `template.*` namespaces, including `document.applyPatch` for multi-operation mutation and `system.batch` for pipelining
  - File **and** base64/buffer I/O in both directions, so binaries never need to touch the filesystem
  - Full protocol reference in [docs/CLI_GUIDE.md](docs/CLI_GUIDE.md)
- **Node.js wrapper — `@mmonterroca/docxgo`** — a TypeScript package (CommonJS + ESM) wrapping the CLI binary with three API levels: `DocxgoExec` (synchronous one-shot), `DocxgoRPC` (low-level persistent client), and `DocumentBuilder` (high-level fluent API). Ships full type definitions and resolves a platform-specific binary via `optionalDependencies`. See [npm/README.md](npm/README.md)
- **`document.setLanguage` RPC method** — exposes v2.6.0's `WithLanguage`/`WithLanguageEx` through the CLI and npm wrapper. Also available as a `setLanguage` patch operation, and the current language is reported by `document.inspect`. Honors the same round-trip guard as `Document.SetLanguage` (works on documents created via `document.create`, not ones opened via `document.open`)
- **Release automation** — `.github/workflows/release.yml` builds multi-platform binaries and publishes a GitHub Release on `v*` tags; `.github/workflows/npm-publish.yml` publishes the platform packages and the main npm package with OIDC provenance when a Release is published

### Fixed

- **`docx.Version`** now reports the correct version. It was stale at `2.5.0` (the v2.6.0 release did not bump it), so `docxgo version` and `system.version` reported the wrong value
- **`template.ConsolidateRuns`** now returns an error and stops at the first run-setter failure instead of silently leaving a paragraph partially rebuilt; `MergeTemplate` and `FindPlaceholders` propagate it
- **`document.applyPatch`** error responses now include an `applied` count, so callers can tell how many operations succeeded before a mid-sequence failure. `applyPatch` is documented as **not** atomic — there is no rollback
- **`template.render`** no longer reports `"error"`-severity findings in an otherwise-successful (`ok: true`) response; in non-strict mode all findings that reach a successful response are labeled `"warning"`
- **npm `DocumentBuilder.create()`/`createToFile()`** now track the new document's ID, so `applyPatch`/`inspect`/`saveToBuffer`/etc. can be chained directly after creating a document — no save-and-reopen round-trip, which would otherwise trip `setLanguage`'s round-trip guard
- **npm `DocxgoRPC.close()`/`kill()`** now mark the client closed synchronously, closing a race window where a call issued during shutdown could hang

---

## v2.6.0 — 2026-07-17

### Added

- **`WithLanguage` / `WithLanguageEx`** — set the document's default proofing language, used by Word for spell-checking, grammar-checking, and hyphenation (closes #44)
  - `WithLanguage(lang string)` sets the primary language (BCP 47 tag, e.g. `"es-MX"`)
  - `WithLanguageEx(docx.Language{Val, EastAsia, Bidi})` additionally sets East Asian (CJK) and right-to-left (bidi) script languages; at least one of the three must be non-empty
  - Written as `w:lang` in `word/styles.xml`'s `docDefaults/rPrDefault/rPr` and as `w:themeFontLang` in `word/settings.xml`
  - `NewDocumentBuilder` always starts from a new document, so `WithLanguage`/`WithLanguageEx` always take effect
  - Opening an existing `.docx` via `OpenDocument`/`OpenDocumentFromBytes`/`OpenDocumentFromReader` now hydrates `Language()` from its `styles.xml`, if it declares one
  - `domain.Document` gained `SetLanguage(*Language) error` and `Language() *Language`; `domain.Language` is the new public type (aliased as `docx.Language`). `Language()` returns a defensive copy — mutating it has no effect on the document
  - `SetLanguage` returns an error (rather than silently no-op) on a document opened via `OpenDocument` whose `styles.xml`/`settings.xml` were preserved verbatim for round-trip fidelity, since a language set there could never actually reach the saved file. Use `WithLanguage`/`WithLanguageEx` when building a new document instead

### Changed

- **`domain.Document` interface gained two methods** (`SetLanguage`, `Language`, above). If you implement `domain.Document` directly with your own type — most commonly a hand-written test double — you'll need to add both methods; embedding `domain.Document` in your type is unaffected, since the new methods are promoted automatically. No exported docxgo API accepts a `domain.Document` from a caller (only returns one), so this is consistent with `SetBackgroundColor`/`BackgroundColor`, which were added to this same interface in the `v2.0.1` patch release.

### Fixed

- **`word/styles.xml` element order** — `internal/xml.Styles` now marshals `w:docDefaults` before `w:latentStyles`, matching the `CT_Styles` schema. This was previously backwards but unobservable, since `DocDefaults` was never populated before this change.
- **CI never ran against `master`** — `.github/workflows/ci.yml` triggered on `branches: [ main, dev ]`, but this repo's default/release branch is `master`, not `main`. The Lint/Build/Test job had silently never fired for any push or PR targeting `master` since the workflow was added. Fixed to trigger on `master` and `dev`.

---

## v2.5.0 — 2026-07-04

### Added

- **Run-level formatting on `CellBuilder`** — table cells now support the full fluent formatting set, matching `ParagraphBuilder` (PR #39; `Italic`/`Color`/`FontSize` contributed by @SlashLight in #35)
  - `CellBuilder.Italic()` — italicize the last run in the last paragraph of the cell
  - `CellBuilder.Color(color)` — set the last run's text color
  - `CellBuilder.FontSize(points)` — set the last run's font size (points, converted to half-points internally)
  - `CellBuilder.Underline(style)` — set the last run's underline style
  - `lastRun()` helper extracted and reused by `Bold()`; existing `Bold()` behavior and error messages are unchanged (backward-compatible)

### Fixed

- **Per-part relationship ID resolution for headers/footers** (PR #40, closes #37)
  - Relationship IDs in OOXML are scoped per-part: a header's own `word/_rels/header1.xml.rels` may reuse an ID (e.g. `rId1`) that also exists in `word/_rels/document.xml.rels` for something unrelated. Drawings inside headers/footers now resolve their `r:embed` against that part's own `.rels`, falling back to the document-wide map — fixing wrong or missing media in headers and footers
  - `internal/reader/parser.go` parses each header/footer part's own `.rels` into a `PartRelationships` map on `ParsedPackage`; an unparseable per-part `.rels` is skipped rather than aborting the whole document open
  - `internal/reader/reconstruct.go` adds an `activeRelationships` scope to `reconstructContext`; header/footer hydration shares a single `hydratePartParagraphs` helper

### Tests

- `TestCellBuilder_RunFormatting` / `TestCellBuilder_RunFormattingErrors` — happy-path and error coverage for `Italic`/`Color`/`FontSize`/`Underline` on cells
- `TestOpenDocument_CollidingRelationshipIDsAcrossParts` — a header image whose `rId1` collides with an unrelated document-level `rId1` resolves to the header's own media (plus a save/reopen round-trip guard)
- `TestOpenDocument_MalformedHeaderRelsIsTolerated` — a corrupt per-part `.rels` no longer fails the open

---

## v2.4.0 — 2026-04-30

### Added

- **In-memory image API** (`pkg/builder` + `domain.Paragraph`) — insert images from byte slices without touching the file system (PR #30, closes #29)
  - `ParagraphBuilder.AddImageFromBytes(data, format)` — inline image from bytes
  - `ParagraphBuilder.AddImageFromBytesWithSize(data, format, size)` — with custom dimensions
  - `ParagraphBuilder.AddImageFromBytesWithPosition(data, format, size, pos)` — floating with positioning
  - Matching methods on `domain.Paragraph` (`AddImageFromBytes`, `AddImageFromBytesWithSize`, `AddImageFromBytesWithPosition`)
  - New `internal/core` constructors: `NewImageFromBytes`, `NewImageFromBytesWithSize`, `NewImageFromBytesWithPosition`
  - Format normalization (`JPG` → `jpeg`, leading `.` trimmed) and validation against the supported set
  - Defensive copy of the input byte slice so callers can safely reuse buffers
- `examples/08_images` updated to demonstrate the new in-memory image flow

### Fixed

- **Round-trip preservation of `w:gridSpan` and `w:vMerge`** for merged table cells (PR #26, closes #25)
  - `hydrateTableCell` now parses `<w:tcPr>` and applies horizontal merges through `cell.Merge(span, 1)` so spanned-over cells are correctly marked as `IsHorizontallyMergedContinuation()`
  - Restores `w:vMerge` (`restart` / `continue`) onto reconstructed `domain.TableCell`s
  - `hydrateTable` tracks `colOffset` and recomputes `maxCols` from gridSpan sums so XML cells map to the correct grid columns
  - Numeric parse errors wrapped with `errors.WrapWithContext` (attribute + raw value) for clearer diagnostics

### Changed

- CLI handler (`cmd/docxgo/handlers.go`) `applyImage()` now uses `AddImageFromBytes*` directly for base64 images, eliminating the temp-file write/read round-trip

### Tests

- `TestGridSpanPreservedAfterRoundTrip` — verifies horizontal merge survives save + reopen and that continuation cells are flagged correctly
- `TestVMergePreservedAfterRoundTrip` — verifies vertical merge `restart` / `continue` survives save + reopen
- Unit tests for all three `NewImageFromBytes*` constructors (valid data, empty data, empty/invalid format)
- Builder tests for all three `AddImageFromBytes*` methods (error path validation)

---

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

- Original fix contributed by @g-mero; the historical PR is no longer available
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

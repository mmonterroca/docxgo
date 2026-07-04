# Release Notes - v2.5.0

**Release Date:** July 2026

## Summary

v2.5.0 brings the fluent table `CellBuilder` to full parity with `ParagraphBuilder` for run-level text formatting, and fixes a header/footer image bug caused by relationship IDs colliding across OOXML parts. The cell formatting feature was contributed by [@SlashLight](https://github.com/SlashLight) (#35) and rounded out with an `Underline()` helper for complete parity (#39).

## New Features

### Run-level formatting on `CellBuilder` (PR #39, feature by @SlashLight in #35)

Previously the fluent table API only exposed `Bold()` at the run level, so italic, color, font size, and underline were unreachable for table cells even though the underlying run model already supported them. `CellBuilder` now offers the same run-formatting surface as `ParagraphBuilder`:

- `CellBuilder.Italic()` — italicize the last run in the last paragraph of the cell
- `CellBuilder.Color(color)` — set the last run's text color
- `CellBuilder.FontSize(points)` — set the last run's font size (points, converted to half-points internally)
- `CellBuilder.Underline(style)` — set the last run's underline style

```go
builder.AddTable(1, 1).
    Row(0).Cell(0).
        Text("Formatted cell").
        Bold().
        Italic().
        Color(domain.Color{R: 255, G: 0, B: 0}).
        FontSize(14).
        Underline(domain.UnderlineSingle).
        End().
    End()
```

Highlights:

- A shared `lastRun()` helper backs all five methods; `Bold()` was refactored to use it while preserving its existing error messages (`"no paragraphs in cell"` / `"no runs in paragraph"`), so behavior is unchanged for existing callers.
- Errors follow the established builder pattern: calling a formatter before any `Text()` records an `InvalidState` error that surfaces at `Build()`.

## Bug Fixes

### Per-part relationship ID resolution for headers/footers (PR #40, closes #37)

Relationship IDs in OOXML are scoped per-part. A header's own `word/_rels/header1.xml.rels` may reuse an ID such as `rId1` that also exists in `word/_rels/document.xml.rels` for something unrelated (e.g. a customXml part). Previously, drawings inside headers and footers resolved their `r:embed` against the document-wide relationships, hydrating the wrong media — or none at all.

The reader now:

- Parses each header/footer part's own `.rels` file into a `PartRelationships` map on `ParsedPackage`, keyed by part path (`internal/reader/parser.go`). An unparseable per-part `.rels` is skipped rather than aborting the whole document open.
- Adds an `activeRelationships` scope to `reconstructContext` (`internal/reader/reconstruct.go`). While hydrating a header/footer, relationship IDs resolve against that part's own `.rels`, falling back to the document-wide map. Header and footer hydration now share a single `hydratePartParagraphs` helper.

Regression tests:

- `TestOpenDocument_CollidingRelationshipIDsAcrossParts` — hand-builds a minimal `.docx` where `rId1` is defined independently in both `document.xml.rels` and `header1.xml.rels`, and asserts the header image resolves to the header's own media part (plus a save/reopen round-trip guard).
- `TestOpenDocument_MalformedHeaderRelsIsTolerated` — a corrupt per-part `.rels` no longer fails the open.

## Installation

```bash
go get github.com/mmonterroca/docxgo/v2@v2.5.0
```

## Compatibility

- **Purely additive** for the fluent builder: the new `CellBuilder` methods do not change any existing signature, and `Bold()` retains its previous behavior and error messages.
- **Documents written with earlier versions are unchanged.** The header/footer fix only affects how existing documents are *read*; output is unaffected.
- No changes to the `domain` interfaces, so custom implementations are unaffected.

## Acknowledgements

- PR #35 / #39 (`feature/cell-run-formatting`) — run-level cell formatting, original feature by [@SlashLight](https://github.com/SlashLight)
- PR #40 (`fix/per-part-relationship-ids`) — header/footer relationship ID fix

## Related Issues

- Closes [#37](https://github.com/mmonterroca/docxgo/issues/37) — header/footer images resolve wrong media on colliding relationship IDs
- Supersedes [#35](https://github.com/mmonterroca/docxgo/pull/35) — original cell formatting PR (targeted `dev`; shipped via #39 on `master`)

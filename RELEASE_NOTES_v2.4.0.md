# Release Notes - v2.4.0

**Release Date:** April 2026

## Summary

v2.4.0 ships a new in-memory image API and a critical round-trip fix for merged table cells. The release adds `AddImageFromBytes*` methods on `ParagraphBuilder` (and the underlying `domain.Paragraph` interface) so callers can embed images directly from byte slices without touching the file system, and restores `w:gridSpan` / `w:vMerge` properties when reopening documents — fixing data loss for tables with horizontal or vertical cell merges.

## New Features

### In-Memory Image Insertion (PR #30, closes #29)

Three new methods on `ParagraphBuilder` and `domain.Paragraph` let you insert images straight from a `[]byte` payload:

- `AddImageFromBytes(data []byte, format ImageFormat)` — inline image from raw bytes
- `AddImageFromBytesWithSize(data, format, size)` — with custom dimensions
- `AddImageFromBytesWithPosition(data, format, size, pos)` — floating with positioning

Highlights:

- Format strings are normalized (`JPG` → `jpeg`, leading `.` trimmed) and validated against the supported set; unknown formats fail fast with a clear `InvalidArgument` error.
- The byte slice is defensively copied so callers can safely reuse or mutate their buffer after insertion.
- The CLI handler (`cmd/docxgo`) now routes base64 images through this path, eliminating the temp-file write/read round-trip used previously.

```go
// Inline image from bytes — no file system needed
builder.AddParagraph().
    AddImageFromBytes(pngBytes, domain.ImageFormatPNG).
    End()

// With custom size
builder.AddParagraph().
    AddImageFromBytesWithSize(pngBytes, domain.ImageFormatPNG, domain.NewImageSize(200, 150)).
    End()
```

The `examples/08_images` example now demonstrates the in-memory flow alongside file-based image insertion.

## Bug Fixes

### Preserve `gridSpan` and `vMerge` on document reopen (PR #26, closes #25)

`hydrateTableCell` in `internal/reader/reconstruct.go` previously skipped `<w:tcPr>`, so horizontal (`w:gridSpan`) and vertical (`w:vMerge`) merge metadata was dropped when reading a document. Re-saving the file then emitted extra `<w:tc>` elements and broke the original table layout.

The reader now:

- Parses `<w:tcPr>` and applies horizontal merges through `cell.Merge(span, 1)` so spanned-over cells are correctly marked as `IsHorizontallyMergedContinuation()`.
- Restores `w:vMerge` (`restart` / `continue`) onto the rebuilt `domain.TableCell`.
- Tracks `colOffset` in `hydrateTable` and recomputes `maxCols` from gridSpan sums so XML cells map to the correct grid columns.
- Wraps numeric parse errors with `errors.WrapWithContext` (attribute + raw value) for clearer diagnostics on malformed documents.

Two regression tests cover the round-trip:

- `TestGridSpanPreservedAfterRoundTrip` — 2×3 table with row 0 spanning all 3 columns; asserts `GridSpan() == 3` and that cells 1 and 2 report `IsHorizontallyMergedContinuation()`.
- `TestVMergePreservedAfterRoundTrip` — 3×2 table with a 3-row vertical merge; asserts `VMergeRestart` on row 0 and `VMergeContinue` on rows 1 and 2.

## Internal Changes

- `internal/core` adds `NewImageFromBytes`, `NewImageFromBytesWithSize`, and `NewImageFromBytesWithPosition` constructors with full unit-test coverage (valid data, empty data, empty/invalid format).
- `cmd/docxgo/handlers.go` `applyImage()` now uses the new in-memory APIs directly for base64 payloads instead of writing to a temp file and calling `AddImage(path)`.

## Installation

```bash
go get github.com/mmonterroca/docxgo/v2@v2.4.0
```

## Compatibility

- Backward compatible with v2.3.x — all additions are new methods on existing interfaces. Custom `domain.Paragraph` implementations outside this module will need to add the three `AddImageFromBytes*` methods.
- Documents written with earlier versions are unchanged; only round-trip behavior for merged tables improves.

## Acknowledgements

- PR #26 (`fix/issue-25-gridspan-lost-on-reopen`) — table merge round-trip fix
- PR #30 (`feature/add-image-from-bytes`) — in-memory image API
- Code review feedback from Copilot, Codex, and Gemini Code Assist on both PRs

## Related Issues

- Closes [#25](https://github.com/mmonterroca/docxgo/issues/25) — OpenDocument drops horizontal cell merges
- Closes [#29](https://github.com/mmonterroca/docxgo/issues/29) — AddImageFromBytes

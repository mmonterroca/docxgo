# Release Notes - v2.2.1

**Release Date:** January 2026

## Bug Fixes

This release fixes critical DOCX validation errors that prevented documents from opening in Microsoft Word after round-trip operations (read → modify → save).

### Fixed

- **Hyperlink RelationshipID Preservation** ([#6](https://github.com/mmonterroca/docxgo/issues/6))
  - When reading and re-saving documents with hyperlinks, the library was generating new relationship IDs instead of preserving the originals
  - This caused Word to fail with "relationship not found" errors
  - Now correctly preserves original `rId` references during round-trip

- **Drawing Position Serialization**
  - Fixed `wp:align` empty value error for floating images
  - Images with `posOffset=0` were incorrectly serialized with empty `<wp:align>` elements
  - Added `UseOffsetX`/`UseOffsetY` flags to `ImagePosition` to distinguish between "offset = 0" and "no offset set"

- **Internal Hyperlink Anchors**
  - Added support for anchor-based internal hyperlinks (bookmarks, TOC links)
  - Serializer now correctly handles both `anchor` property and URLs starting with `#`

- **Custom Styles Preservation**
  - Documents with custom styles (modified Heading1, custom BodyCopy, etc.) are now preserved correctly during round-trip
  - All 55+ custom styles from templates are maintained

## New Features

### Troubleshooting Documentation

Added comprehensive troubleshooting guide for DOCX validation errors:
- `docs/TROUBLESHOOTING_DOCX_VALIDATION.md`
- Includes diagnostic workflow, common errors, and solutions

## Installation

```bash
go get github.com/mmonterroca/docxgo/v2@v2.2.1
```

## Usage Example

```go
// Read existing document with custom styles
doc, err := docx.OpenDocument("template.docx")
if err != nil {
    log.Fatal(err)
}

// Modify content - styles are preserved
para := doc.AddParagraph()
para.AddRun().SetText("New content")

// Save - document opens correctly in Word
if err := doc.SaveAs("output.docx"); err != nil {
    log.Fatal(err)
}
```

## Known Limitations

When loading an existing document and adding **new** images or hyperlinks (not modifying existing ones), the new relationships may not be written correctly. This is tracked in [#8](https://github.com/mmonterroca/docxgo/issues/8).

For the primary use case of reading a template, modifying text, and saving, this version works correctly.

## Acknowledgements

Thanks to [@krishnadubagunta](https://github.com/krishnadubagunta) for reporting the issue with detailed examples that helped identify the root cause.

## Files Changed

- `domain/image.go` - Added `UseOffsetX`/`UseOffsetY` flags
- `internal/core/run.go` - Preserve existing relationshipID
- `internal/core/document.go` - Store preserved parts for round-trip
- `internal/reader/reconstruct.go` - Preserve original parts and set UseOffset flags
- `internal/serializer/serializer.go` - Handle anchor hyperlinks, use preserved relationshipID
- `internal/writer/zip.go` - Write preserved parts during round-trip
- `internal/xml/drawing_helper.go` - Use offset flags in serialization
- `internal/xml/paragraph.go` - Add Anchor/History to Hyperlink struct

## Full Changelog

See [PR #7](https://github.com/mmonterroca/docxgo/pull/7) for complete details.

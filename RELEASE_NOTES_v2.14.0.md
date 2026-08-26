# Release Notes - v2.14.0

**Release Date:** August 26, 2026

## Overview

v2.14.0 makes edits to existing Word documents substantially safer. The body round-trip no longer depends on reconstructing every source feature in docxgo's domain model. It now keeps the original `word/document.xml` bytes and applies localized, policy-driven changes only where the API explicitly modified the document.

The practical result is that adding a row or changing a section no longer rewrites unrelated content controls, text boxes, floating table settings, row formatting, opaque properties, or whitespace. A write with no model changes preserves `word/document.xml` byte-for-byte.

This release also adds `CellBuilder.Alignment` and restores compatibility with real-world documents whose table widths use integer-valued decimal syntax such as `9360.0`.

## Upgrade

Go:

```bash
go get github.com/mmonterroca/docxgo/v2@v2.14.0
```

Node.js:

```bash
npm install @mmonterroca/docxgo@2.14.0
```

The complete source diff is available in the [v2.13.0...v2.14.0 comparison](https://github.com/mmonterroca/docxgo/compare/v2.13.0...v2.14.0).

## Highlights

### Lossless `document.xml` round-trip

Opened documents now retain source anchors for original body blocks, tables, rows, table properties, grids, and sections. Each unit has its own fingerprint. When the model is unchanged, the writer uses the original bytes. When the API changes a unit, the writer replaces or merges only that unit.

This fixes the losses reported in #116 when adding a row to a manually authored document:

- text boxes and content controls survive because untouched paragraphs remain raw;
- TOCs wrapped in `w:sdt` survive in their original position;
- floating table positioning and other unmodeled `w:tblPr` children survive;
- existing rows retain their exact borders, run formatting, font sizes, and row properties;
- opaque markup between rows remains attached to the original following row;
- tables and section headings keep their original order.

### Explicit OOXML merge policies

The writer now applies property-specific rules instead of replacing whole property subtrees:

- `w:pgMar` updates the six modeled margins and preserves `gutter`;
- `w:pgSz` updates width, height, and orientation while retaining orthogonal attributes;
- header and footer references match by type and update their relationship ID;
- changing the number of columns removes incompatible explicit `w:col` children and `equalWidth`, while preserving compatible spacing and separator options;
- `w:tblPr` children merge in schema order;
- table borders merge by side, preserving `space`, theme colors, and other compatible attributes;
- table-grid columns merge by position and retain unknown attributes;
- an API removal also removes opaque dependent properties that would contradict it.

### Section breaks stay structural

An embedded `w:sectPr` now remains inside the `w:pPr` of its source paragraph. A section-only edit patches that subtree without introducing a paragraph. If both the paragraph and section change, the writer serializes one paragraph containing the merged section properties. The final section follows the same policies as intermediate sections.

### QName-safe parsing and merging

Elements and attributes are matched by their resolved namespace plus local name. An unrelated namespace reusing a familiar local name can no longer be interpreted as WordprocessingML or as a relationship attribute.

The new `internal/ooxmlmerge` engine retains original prefixes, attributes, children, source offsets, and bytes. It supports preservation, replacement, attribute merging, child merging, and byte-range splicing. Replacements are applied from the end of the source toward the beginning so earlier offsets remain valid.

### `CellBuilder.Alignment`

`CellBuilder.Alignment(domain.Alignment)` applies a horizontal alignment to every paragraph currently in the cell. An empty cell is a no-op. Invalid values follow the existing builder contract and are returned by `Build()`.

Requested by @drkisler in #117.

### Tolerant table widths

The reader accepts decimal strings such as `9360.0` for `w:tblW/@w:w`. Decimal measurements are truncated toward zero before the model validates them. Malformed values, `NaN`, infinities, and values outside the range of `int` remain errors.

This fixes the v2.13.0 regression tracked in #118.

## Identity safety

Source drawing and bookmark IDs are observed before new content is generated. A newly inserted image or bookmark therefore cannot reuse an identifier already present in preserved raw XML.

## Compatibility

- No existing public method, interface, or signature changed.
- `CellBuilder.Alignment` is an additive public method.
- The lossless merge scope is `word/document.xml`. Headers, footers, styles, relationships, and other package parts retain their v2.13.0 mechanisms.
- Explicit API modifications take priority over original XML when the two conflict.
- Modified paragraphs and rows remain atomic units, except that section properties stay attached to their source paragraph.

## Validation

The release candidate is covered by:

- exact byte comparison for no-change writes;
- exact byte comparison for a second write after modification;
- the complete #116 source fixture and row-insertion reproduction;
- section, columns, borders, table grid, drawing, bookmark, namespace, whitespace, and opaque-markup regressions;
- unit and fuzz seeds for the merge engine;
- package reopen and `Validate()` checks;
- a generated two-section Word document with tables, images, headers, footers, and opaque XML, rendered and inspected page by page;
- `go build ./...`, `go test -count=1 ./...`, `go test -race ./...`, `go vet ./...`, formatting, and diff checks.

## Credits

In #116, @djmoch provided the original input and output documents. The first output came from a local `replace` based on v2.12.0, which explained why its table-width loss could not be reproduced against the v2.13.0 tag. After correcting the module version, he supplied a new v2.13.0 output that confirmed the remaining losses, including cell borders and font sizes, while the same input also reproduced the text-box, TOC, and floating-table losses. That follow-up separated the version mismatch from the defects fixed in this release; the first output's table-width symptom is therefore not listed as a v2.13.0 regression.

`CellBuilder.Alignment` was requested by @drkisler in #117.

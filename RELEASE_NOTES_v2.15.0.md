# Release Notes - v2.15.0

**Release Date:** November 1, 2026

## Overview

v2.15.0 adds the paragraph and table controls needed to generate formatted reports: shaded paragraphs, paragraphs that stay with the next one, table header rows that repeat on every page, rows that never split across pages, and a table of contents that shows a pre-rendered result until Word rebuilds it.

## Interface change: read this before upgrading

New methods were added to two exported interfaces:

- `domain.Paragraph`: `Shading`, `SetShading`, `ClearShading`, `KeepNext`, `SetKeepNext`, `KeepLines`, `SetKeepLines`.
- `domain.TableRow`: `RepeatAsHeader`, `SetRepeatAsHeader`, `CantSplit`, `SetCantSplit`.

If you use only docxgo's own types, nothing changes. If your code defines a type that implements `domain.Paragraph` or `domain.TableRow` (typically a hand-written test double), it stops compiling until you add those methods. Embedding `domain.Paragraph` or `domain.TableRow` in your type avoids the change, because the new methods are promoted automatically.

The only exported function that accepts a caller-provided value of either type is `template.ConsolidateRuns(domain.Paragraph)`. Passing it your own `Paragraph` implementation requires the new methods too.

### Why this is a minor release

These interfaces describe the paragraphs and rows that docxgo creates and returns. They are not an extension point: docxgo has no function that takes your implementation and writes it into a document. Adding a method to this kind of interface is how the library has always grown, and earlier minor releases did the same: `domain.Run.SetLanguage` in v2.12.0, and `AddTable`, `Tables` and `Blocks` on `domain.Header` and `domain.Footer` in v2.13.0. No existing method, signature or behavior of the library's own types changed. Releasing this as a major version would force every user to migrate for a change that only affects code implementing these interfaces itself.

## Upgrade

Go:

```bash
go get github.com/mmonterroca/docxgo/v2@v2.15.0
```

Node.js:

```bash
npm install @mmonterroca/docxgo@2.15.0
```

The complete source diff is available in the [v2.14.0...v2.15.0 comparison](https://github.com/mmonterroca/docxgo/compare/v2.14.0...v2.15.0).

## Highlights

### Paragraph shading

`Paragraph.SetShading(Color)` writes `<w:shd w:val="clear" w:fill="RRGGBB"/>` in `w:pPr`, right after `w:pBdr` as the schema requires. `Shading() (Color, bool)` reports whether a shading is set, which tells "unset" apart from black (`Color`'s zero value). `ClearShading()` removes it. The reader hydrates a source `w:shd` with the same clear/solid rules already used for cell shading. `ParagraphBuilder.Shading(Color)` is the builder entry point.

### Keep with next, keep lines together

`SetKeepNext(bool)` and `SetKeepLines(bool)` write `w:keepNext` and `w:keepLines` right after `w:pStyle`. They keep, for example, an image with its caption on one page. An explicit `false` is written as `w:val="false"`, so it overrides what a style sets. The reader hydrates both. `ParagraphBuilder.KeepNext()` is the builder entry point.

### Repeating header rows and unsplittable rows

`TableRow.SetRepeatAsHeader(true)` writes `w:tblHeader` so Word repeats the row at the top of every page the table spans. `SetCantSplit(true)` writes `w:cantSplit` to keep a row on one page. Both live in `w:trPr`, are hydrated by the reader, and survive the round trip.

### Pre-rendered TOC results

A TOC field now keeps a result set with `SetResult` even while it is dirty. Before, serializing replaced it with the "Table of Contents" placeholder. The field stays dirty, so Word still rebuilds it on open, but your entries show until then. A result containing `\n` is written as one run per line, separated by `w:br`.

## Compatibility

- Interface additions on `domain.Paragraph` and `domain.TableRow`, described above.
- No existing method, signature or default behavior was removed or changed. The one behavior difference is that a dirty TOC field with a caller-set result no longer shows the placeholder text.
- Output is unchanged for documents that do not use the new methods.

Included in #128.

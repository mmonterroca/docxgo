# Release Notes - v2.16.0

**Release Date:** November 3, 2026

## Overview

v2.16.0 makes it possible to take content out of an existing document, and fixes a bug that glued words together when an opened document was edited.

Both changes were contributed by [@coffee-realist](https://github.com/coffee-realist), who reported each one with a precise diagnosis ([#124](https://github.com/mmonterroca/docxgo/issues/124), [#126](https://github.com/mmonterroca/docxgo/issues/126)), sent the implementation with thorough tests ([#125](https://github.com/mmonterroca/docxgo/pull/125), [#127](https://github.com/mmonterroca/docxgo/pull/127)), and went beyond what the review asked for. Thank you!

## Interface change: read this before upgrading

`domain.Document` gained one method, `RemoveBlock(index int) error`.

If you use only docxgo's own types, nothing changes. If your code defines a type that implements `domain.Document` (typically a hand-written test double), it stops compiling until you add the method. Embedding `domain.Document` in your type avoids the change, because the new method is promoted automatically.

The functions in `pkg/template` (`MergeTemplate`, `ValidateTemplate`, `ReplaceText`, `FindPlaceholders`, `FindPlaceholdersCustom`, `PlaceholderNames`) accept a caller-provided `domain.Document`, so passing them your own implementation requires the new method too. The v2.6.0 notes said no exported function accepted a `domain.Document`; that was not accurate.

This is a minor release for the same reason as v2.15.0: `domain.Document` describes the documents docxgo creates and returns, not an extension point.

## Upgrade

Go:

```bash
go get github.com/mmonterroca/docxgo/v2@v2.16.0
```

Node.js:

```bash
npm install @mmonterroca/docxgo@2.16.0
```

The complete source diff is available in the [v2.15.0...v2.16.0 comparison](https://github.com/mmonterroca/docxgo/compare/v2.15.0...v2.16.0).

## Highlights

### Removing paragraphs and tables

`Document.RemoveBlock(index)` removes the top-level paragraph or table at a position in `Blocks()`. It covers the cases where an existing document or template needs content taken out: an optional section that doesn't apply, a placeholder left empty after a merge, or selected list items together with the paragraphs and tables that belong to them.

For a document opened from a file, the removed element is left out on save and the rest of `word/document.xml` is written back byte-for-byte. A paragraph that carries a mid-document section break can be removed without losing the section or its properties. When a bookmark or comment range starts in the removed paragraph and ends in a kept one, the unmatched half is dropped instead of being left dangling.

A few things to keep in mind:

- `Blocks()` indices include section breaks and shift after each removal, so remove several blocks in descending index order.
- Section breaks cannot be removed.
- Content the model does not represent, such as a body-level content control (`w:sdt`), is not part of `Blocks()`, so it can't be removed this way. It is still written back as it was.
- Removing the only paragraph between two tables makes Word render them as one table.

### Spaces between words no longer disappear after an edit

Word often stores a space as its own run, for example between a bold word and the regular text after it. docxgo skipped that run when opening the document. Untouched paragraphs were still written back as they were, so nothing showed until the paragraph was edited in any way (even `SetAlignment`), and then the words around the space were glued together. A run holding only a no-break space was lost the same way, and a field code split as `PAGE`, ` `, `\* MERGEFORMAT` was read as `PAGE\* MERGEFORMAT`.

Whitespace-only text is now kept when `xml:space="preserve"` applies to it, set on the element itself or inherited from an ancestor, as XML 1.0 defines it. A no-break space is always kept. Without `preserve`, whitespace-only text is still ignored, as Word does, so editing a paragraph doesn't make an insignificant space visible. Headers, footers, table cells, hyperlinks and field instructions go through the same reader and are fixed as well.

## Compatibility

- One method added to `domain.Document`, described above.
- No existing method, signature or default behavior was removed or changed, apart from the fix above: `Text()` and `Runs()` of an opened paragraph now include whitespace-only runs that carry `xml:space="preserve"`.
- Output is unchanged for documents that don't use `RemoveBlock` and don't contain whitespace-only runs.

Included in #125 and #127, integrated through #131.

# Release Notes - v2.11.0

**Release Date:** July 29, 2026

## Summary

v2.10.0 was merged and published to npm without a code review — CI alone. A
post-hoc review of that release's full diff found ten confirmed defects that
CI's gofmt, `go vet`, `go test -race`, and golangci-lint could not have
caught, because none of them are the kind of defect those tools can see:
they're about what the code *means*, not whether it compiles cleanly or
races. v2.11.0 fixes all ten in one batch.

Two are worth calling out specifically because they're silent data loss: a
`document.replaceText` call could report a header or footer edit as applied
when it was actually discarded on save, and a misspelled parameter name
(`replacement` instead of `replace`) could delete a document's content while
the RPC reported success.

No public Go interface changed. If you're already on v2.10.0, upgrade —
there's no reason to stay on it.

## Fixed

### `document.replaceText`/`MergeTemplate` no longer report header/footer edits that get discarded on save

On a document opened via `document.open`, `WriteTo` writes the original
header/footer XML back verbatim for round-trip fidelity — any in-memory
mutation to a header or footer paragraph never reaches the saved file. Both
`document.replaceText` and `template.MergeTemplate` mutated those paragraphs
anyway and reported success: `replaceText` counted the match as `replaced`,
and `MergeTemplate` treated the placeholder as filled.

Both now recognize a document with preserved header/footer parts and treat
header/footer matches as unwritable: `replaceText` reports them in `skipped`
instead of `replaced`, and `MergeTemplate`'s `StrictMode` now surfaces them
as missing keys rather than silently discarding the write. This is a
document-wide check — if any header or footer part was preserved, every
header/footer match is skipped, not just the one in the section that was
actually touched.

### Strict decoding closes a silent-deletion path in `document.replaceText`

`document.replaceText`, `table.getCell`, and `table.setCell` decoded their
top-level params with plain `json.Unmarshal`, unlike the nested paragraph
items in the same requests, which were already decoded strictly. A
misspelled top-level field silently decoded to Go's zero value —
indistinguishable from a deliberately minimal request.

Concretely: a `document.replaceText` request with `"replacement"` instead of
`"replace"` decoded to an empty `Replace` string, deleted every occurrence of
`find` in the document, and reported `{"replaced": N}` as a success. All
three methods now reject an unrecognized field instead.

### No orphaned content when `paragraph.add`/`table.add`/`document.addContent` reject a request

All three appended to the real document before content that could still fail
validation — most commonly a field constructor rejecting a double quote in a
style name or URL — got the chance to. A rejected request could therefore
leave a stray, half-populated paragraph or table behind even though the
overall call returned an error.

All three now validate against a scratch document first, the same pattern
`paragraph.setText` and `table.setCell` already used. `document.addContent`
validates every item in a content batch before applying any of them, so a
later item's rejection can no longer leave an earlier item's paragraph
behind either.

### `SetIndent` no longer leaves a stale per-side flag behind

v2.10.0 added `SetIndentLeft`/`SetIndentRight`/`SetIndentFirstLine`/
`SetIndentHanging`, each marking a per-side "this was explicitly set" flag so
the serializer can emit an explicit `0` on just that side. `SetIndent` — the
whole-struct call — replaced the indentation values but never cleared those
flags.

The consequence: a paragraph hydrated by the reader (which calls
`SetIndentLeft` for a source `<w:ind>`'s `left` attribute) and then given a
plain `SetIndent` call for a different side kept the stale `indentLeftSet`
flag from the read. The serializer, seeing that flag, emitted an explicit
`w:left="0"` the `SetIndent` call never asked for — silently clobbering the
paragraph's style indentation on the left side. This was a round-trip
regression `SetIndent` introduced against v2.9.1's behavior, in the very
release that added per-side indentation. `SetIndent` now clears all four
flags along with the struct it replaces.

### The serializer never emits both `firstLine` and `hanging` on `<w:ind>`

`SetIndent` already rejects a caller trying to set both `FirstLine` and
`Hanging` at once — Word treats them as mutually exclusive even though the
OOXML schema technically allows both. The per-side setters
`SetIndentFirstLine`/`SetIndentHanging` don't share that check, and the
reader calls them independently per attribute present in a source `<w:ind>`,
so a document whose original XML already carried both attributes could
reach the serializer with both set. The serializer now emits only one —
`hanging` wins, matching what Word itself does when it encounters both.

### TOC field switches are written with a single backslash, not two

`NewTOCField`'s default code and its `\o`/`\h`/`\n`/`\p`/`\z`/`\u` switches
were built as Go raw string literals, each containing two literal
backslashes rather than one escaped backslash. XML chardata escaping never
touches a bare backslash, so the doubled backslash reached
`word/document.xml` completely unchanged — and Word's field-code parser
treats `\\o` as an unrecognized switch, not `\o` with an extra character. A
table of contents field built by docxgo has never actually applied any of
its switches (default heading levels, hyperlinks, page-number hiding, tab
leaders) in any version before this one.

This only affects documents containing a field constructed via `NewTOCField`
(directly, or through `field: {"type": "toc"}` in the RPC layer) — rebuilding
an existing, already-saved document doesn't change anything on its own.

### `Field.SetCode` rejects control characters; the field-code sanitizer can no longer be bypassed

v2.10.0's field-code sanitizer (#85) rejects a `"` that would break out of a
quoted field-code argument, but only in the `New*Field` constructors.
`Field.SetCode` — a lower-level, separately-reachable method on the same
type — accepted any non-blank string verbatim, and the serializer emits
`Code()` straight into `<w:instrText>`. `SetCode` is Go-API-only (no RPC
parameter reaches it), but it was a way to reach the same class of
byte-level corruption the #85 fix exists to prevent, entirely bypassing that
fix.

`SetCode` now rejects a control character or line break — unambiguously
invalid inside a single-line field instruction — while still accepting
balanced quotes, since a well-formed instruction legitimately contains them
(e.g. `STYLEREF "Heading 1"`); a naive quote-count check can't tell that
apart from an injection attempt. The reader, which hydrates a field's code
from a real file's own `<w:instrText>`, uses a new unexported `SetCodeRaw`
path that bypasses this guard — the source file is the ground truth for
what a valid field instruction looks like, so `OpenDocument` won't fail on
an ordinary `.docx` because of this new check.

### Release workflow no longer hard-fails when `RELEASE_PAT` is unset

`release.yml` already had a check that warned when `RELEASE_PAT` was empty
and explained the release would fall back to being authored by
`github-actions[bot]`. But the step that actually creates the GitHub Release
passed `token: ${{ secrets.RELEASE_PAT }}` as a `with:` input — and an
explicitly-passed *empty* `with:` input overrides
`softprops/action-gh-release`'s own default fallback to `github.token`,
unlike an unset environment variable. With the secret unset, the release
step didn't degrade to a bot-authored release as the warning promised — it
failed outright, taking the whole release with it. `token:` now reads
`secrets.RELEASE_PAT || github.token` explicitly.

## Changed

### Docs corrected to match `paragraph.setText`/`table.setCell`'s actual behavior

Both methods replace a paragraph's *content* (its text and runs) — any
paragraph property (style, alignment, indentation, spacing) the request
doesn't mention keeps whatever value the paragraph already had; neither
method resets it. `CLI_GUIDE.md` and the npm `ParagraphSetTextParams` JSDoc
previously described both as full replacements without drawing that
distinction, which reads as "specify everything or it's gone" — the opposite
of the actual behavior.

### `TableSetCellResult`'s npm JSDoc corrected

It said the cell "can grow but never shrink" — the behavior before #84,
which shipped `TableCell.RemoveParagraph` in the very same v2.10.0 release
that made this description false.

## Compatibility

- **No public Go interface changed. No new public Go, CLI protocol, or
  Node.js API surface.**
- **`document.replaceText`/`MergeTemplate` response semantics change for
  documents with preserved headers/footers.** A call that used to report a
  header/footer match as `replaced` now reports it as `skipped` (total match
  count is unchanged); `MergeTemplate`'s `StrictMode` can now fail on a
  document it previously reported as fully merged. Any document without
  preserved headers/footers is unaffected.
- **`document.replaceText`, `table.getCell`, and `table.setCell` now reject
  an unrecognized top-level parameter**, where it used to be silently
  ignored. A well-formed request is unaffected.
- **TOC field code bytes change.** A document built with `NewTOCField` under
  this release has single-backslash switches instead of the doubled-backslash
  bytes every prior version produced. This isn't a behavior any caller could
  have depended on — the doubled-backslash form was never a working TOC
  field in Word.
- **`Field.SetCode` (Go API only — no RPC parameter reaches it) now rejects
  a control character or line break**, where it previously accepted
  anything non-blank.
- **Retroactive note on v2.10.0's `Config` pointerization.** v2.10.0's
  `### Compatibility` section described `docx.Config`'s `DefaultFont`/
  `DefaultFontSize`/`PageSize`/`Margins` fields becoming pointers, but not
  its concrete consequence: this is source-breaking for code that writes to
  a `*docx.Config` directly (for example, a hand-written `docx.Option`), not
  for code that only calls `docx.With*` functions. `c.DefaultFont = "Arial"`
  no longer compiles; see `MIGRATION.md` for the replacement.

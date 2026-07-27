# Release Notes - v2.8.0

**Release Date:** July 27, 2026

## Summary

v2.8.0 fixes two defects in generated document output. One is a silent data
loss: inline images could disappear from a document merely because a template
operation touched their paragraph. The other is paragraph spacing, where
generated documents did not state their own defaults and Microsoft Word
substituted its own.

The spacing fix **changes how existing documents render** and is the reason
this is a minor release rather than a patch — see "Rendering change" below
before upgrading.

## Fixed

### Inline images could be silently deleted by template operations

`template.ConsolidateRuns` merges adjacent runs with identical formatting, to
heal the "split placeholder" problem where Word fragments a `{{token}}` across
several `<w:r>` elements. It did this by rebuilding the paragraph's run list:
clearing it, then re-creating each run and copying the old run's properties
across one by one.

That copy went through the `domain.Run` interface, which has no way to carry a
run's image. Any run holding a picture was therefore dropped the moment
consolidation touched its paragraph — the `<w:drawing>` vanished from
`word/document.xml` and the `word/media/` part was left orphaned.

The blast radius was wider than mail merge:

- `MergeTemplate` — any merge over a document with a picture near mergeable runs.
- `FindPlaceholders` — reads as a pure query, but consolidates internally, so
  **scanning a document for placeholders could delete an image from it**.
- Headers and footers, which is where logos live: a merge touching only the
  body could remove the header logo.

Consolidation no longer rebuilds runs at all. Each merge group's own first run
receives the combined text and the runs it absorbed are removed; runs that
merge with nothing are not touched. Nothing is copied, so nothing can be lost —
and the failure mode is gone structurally, not patched: there is no longer a
field-by-field copy that must be kept in sync as `domain.Run` grows.

Documents already damaged by this cannot be recovered from the saved file; the
image data is gone from `document.xml`. Regenerate them from source.

### Generated documents now state their own paragraph spacing defaults

`word/styles.xml` carried an empty `<w:pPrDefault/>`, so a paragraph that did
not set spacing explicitly inherited nothing from the document and Word applied
its own defaults instead — 8pt space after each paragraph and 1.15 line
spacing. docxgo's model says those defaults are 0pt and single (240 twips), so
generated output did not match the API's own semantics. The same gap made an
explicit `SpacingAfter(0)` indistinguishable from never setting it.

`w:docDefaults` now carries a real `w:pPrDefault` of
`before=0, after=0, line=240, lineRule=auto`, in both the style-manager path
and the no-style-manager fallback.

Reported and diagnosed by **@b52es** in
[#63](https://github.com/mmonterroca/docxgo/pull/63) — thank you. The fix here
takes a different route than that PR proposed: forcing direct `0/0` formatting
onto every paragraph would have won over paragraph styles in the OOXML cascade
and stripped the spacing from every `Heading` in every document. Setting the
document defaults fixes the reported cases without touching styles.

### Exact and at-least line spacing at 240 twips

`SetLineSpacing({Rule: Exact, Value: 240})` and `{Rule: AtLeast, Value: 240}`
emitted no `<w:spacing>` element, because the decision to emit looked only at
the value and 240 is the default. With document defaults now in place, such a
paragraph would have inherited `lineRule="auto"` — silently turning a caller's
exact 12pt line height into automatic spacing. A non-auto rule is now treated
as a departure from the defaults in its own right, whatever its value.

## Rendering change

**Documents generated without explicit spacing will render more compactly than
they did before.**

Previously they inherited Word's defaults (8pt after each paragraph, 1.15 line
spacing) because docxgo stated none. They now render at the 0pt / single
spacing that docxgo's API has always claimed. This is the correction, but it is
visible in every such document.

To keep the previous look, set the spacing explicitly:

```go
para.SetSpacingAfter(160)                                    // 8pt
para.SetLineSpacing(domain.LineSpacing{
    Rule:  domain.LineSpacingAuto,
    Value: 276,                                              // 1.15 lines
})
```

Or apply it once through a style, or via the builder's document-level options.

**Documents opened from disk are unaffected.** `OpenDocument` preserves the
source `word/styles.xml` verbatim and writes it back untouched, for round-trip
fidelity. A `.docx` produced by docxgo v2.7.2 or earlier therefore does not
acquire the new defaults by being opened and re-saved — it must be regenerated.

## Known limitations

- **An explicit `SetSpacingBefore(0)`/`SetSpacingAfter(0)` is still discarded on
  a paragraph that carries a style with non-zero spacing.** The domain model
  cannot distinguish "never set" from "explicitly set to zero", so the
  attribute is omitted and the style's value wins. This release fixes the
  unstyled case; the styled case needs unset-tracking in the domain model and
  is tracked in [#69](https://github.com/mmonterroca/docxgo/issues/69). The
  same limitation applies to `Indentation`.
- **`FindPlaceholders` still mutates the document** it scans, by consolidating
  runs. This release removes the damage that mutation could do, but not the
  mutation itself — tracked in
  [#68](https://github.com/mmonterroca/docxgo/issues/68).

## Changed

### The `dev` integration branch has been retired

Contributions now branch from and target `master`.

`dev` was introduced as a Git Flow integration branch, but since ~v2.4 its only
traffic had been `master → dev` sync merges — it had stopped being a real
integration point while still being the branch `CONTRIBUTING.md` told
contributors to use. That cost real contributions:
[#35](https://github.com/mmonterroca/docxgo/pull/35) targeted `dev`, went stale
there, and had to be re-shipped through `master` as
[#39](https://github.com/mmonterroca/docxgo/pull/39);
[#57](https://github.com/mmonterroca/docxgo/pull/57) was closed; and the author
of [#64](https://github.com/mmonterroca/docxgo/pull/64) declined to target
`dev` at all because it was behind `master`.

Releases are cut by tagging `master`, so the branch was never the release gate
either. Short-lived `integration/<topic>` branches remain available for the
rare case where two in-flight changes must be co-staged before either lands.

## Compatibility

- **No public Go, CLI protocol, or Node.js API changed.** No signature,
  interface, or method was added, removed, or altered.
- **Generated output changes** as described under "Rendering change".
- `template.ConsolidateRuns`, `MergeTemplate`, and `FindPlaceholders` keep
  their existing signatures and behavior for every `domain.Paragraph`
  implementation, including third-party and wrapped ones.

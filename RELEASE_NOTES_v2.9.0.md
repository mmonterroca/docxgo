# Release Notes - v2.9.0

**Release Date:** July 27, 2026

## Summary

v2.9.0 closes out the follow-ups filed during v2.8.0's release: the release
pipeline's npm publish step no longer depends on an event that a bot-authored
release can never fire, `FindPlaceholders` and its siblings stop mutating the
document they scan, an explicit zero spacing is now honored on a styled
paragraph, and a dead code path in the styles writer is removed.

The spacing fix **changes rendering for paragraphs that explicitly zero their
spacing on top of a style** — see "Rendering change" below before upgrading.
Nothing else in this release changes generated output.

## Fixed

### npm publish no longer depends on the GitHub release event

`release.yml` authors the GitHub Release using a `RELEASE_PAT` repository
secret specifically so the `release: published` event fires and
`npm-publish.yml` picks it up automatically. When that secret is empty or
expired, `softprops/action-gh-release` silently falls back to the default
`GITHUB_TOKEN`, and the release ships anyway — authored by
`github-actions[bot]` instead of a real user. GitHub does not let events
created by `GITHUB_TOKEN` trigger other workflows (anti-recursion protection),
so `npm-publish.yml`'s trigger never fires. `release.yml` still reports
success, because creating the release worked fine; nothing surfaces that the
cascade it exists to start did not happen.

This was not hypothetical: every release since at least v2.7.2, including
v2.8.0, needed a manual `workflow_dispatch` to actually reach npm.

`release.yml` now invokes `npm-publish.yml` directly as a `workflow_call` job,
passing the version it already computed as a job output. Publishing is now an
edge in the release job graph rather than a listener on an event that can go
silently dead. The `RELEASE_PAT` emptiness check remains, but as a warning: a
bot-authored release is no longer a publish blocker, only a cosmetic wrong
author on the GitHub Release page.

### FindPlaceholders, PlaceholderNames, and ValidateTemplate no longer mutate the document they scan

All three read as pure queries — none return an error, none hint at taking a
mutable argument. Internally, though, they called `template.ConsolidateRuns`
on every paragraph to heal placeholders Word split across `<w:r>` elements.
v2.8.0 removed the worst consequence of that mutation (it used to silently
drop images); this removes the mutation itself.

The run-grouping logic `ConsolidateRuns` already computed — which adjacent
runs share formatting and can be treated as one span of text — is now a
shared helper. The scanner matches placeholders against each group's
concatenated *virtual* text without touching the paragraph, then translates
the match's offsets back to the real, unmodified run and offset where it
starts and ends.

`template.Location` gains a new field, `EndRunIndex`, since a match spanning
multiple runs now needs to report where it ends as well as where it starts.
`RunIndex`/`StartOffset` point at the run the paragraph actually has, not a
post-consolidation run that no longer exists once the scan is done.

`MergeTemplate` is unaffected and keeps consolidating — there, merging split
runs is necessary to write the replacement text into one place.

### An explicit zero spacing (or line spacing) is now honored on a styled paragraph

v2.8.0 fixed this for the *unstyled* case: a paragraph with no style and no
explicit spacing now correctly inherits `0` from the document's
`w:pPrDefault`. It did not fix the *styled* case, because
`Paragraph.SpacingBefore()`/`SpacingAfter() int` can't tell "the caller never
called this" from "the caller explicitly set it to 0" — so
`SetSpacingAfter(0)` on a paragraph carrying a style with non-zero spacing was
silently dropped, and the style's own value won instead of the caller's
explicit zero.

The same gap applied to line spacing: an explicit
`SetLineSpacing({Rule: Auto, Value: 240})` meant to override a style's
non-auto rule back to single-spacing was indistinguishable from never calling
`SetLineSpacing` at all, since `Auto`/`240` are also the defaults.

The concrete paragraph type now tracks whether `SetSpacingBefore`,
`SetSpacingAfter`, and `SetLineSpacing` were ever called, and the serializer
honors an explicit value even when it happens to be zero (or a default).
`domain.Paragraph` itself is unchanged — the tracking is read via a type
assertion that degrades to the old behavior for any implementation that
doesn't expose it, so third-party or wrapped paragraphs keep working exactly
as before.

**Indentation has the identical gap** — `SetIndent` can't currently express
"I set Left but left Right untouched" either — but fixing it needs a
different shape of change (`SetIndent` takes one struct covering all four
sides, unlike spacing's four independent setters). Tracked separately in
[#76](https://github.com/mmonterroca/docxgo/issues/76).

## Rendering change

**A paragraph that carries a style and explicitly sets its own spacing to 0
will now render at that 0, instead of silently keeping the style's spacing.**

This only affects paragraphs that both (a) have a style with non-zero
spacing, and (b) explicitly call `SetSpacingBefore(0)`, `SetSpacingAfter(0)`,
or `SetLineSpacing` with default-looking values, expecting it to stick. Before
this release, such a call was silently ignored. If your code relied on that
silent ignoring — i.e., you called `SetSpacingAfter(0)` on a styled paragraph
but expected the style's spacing to still show through — remove the call;
that intent can no longer be expressed as "set to 0", since 0 is now always
honored. Paragraphs that never call these setters are unaffected.

## Changed

### Removed the unreachable no-style-manager styles.xml fallback

`ZipWriter.writeDefaultStyles` wrote a hand-maintained raw `word/styles.xml`
string, reached only when `writeStyles` received a `nil` `*Styles`. The only
production caller (`internal/core/document.go`) always passes a real
serialized style manager, so the fallback only ever ran from tests calling
`WriteDocument` directly — never from a generated document. It was a second,
independent source of truth for default paragraph properties that had to be
kept in sync by hand with the serializer; v2.8.0's spacing fix had to update
both copies and add a test pinning that they agreed. `writeStyles` now
returns an error on `nil` instead of degrading to the fallback, making the
invariant explicit. That cross-path drift test is removed along with what it
was guarding.

### CLI: spacingBefore/spacingAfter accept an explicit zero

`paragraph.add`'s `spacingBefore`/`spacingAfter` JSON fields become nullable
integers server-side, so a request body with `"spacingAfter": 0` can be told
apart from omitting the field entirely. Without this, the spacing fix above
would only reach the Go API — the CLI and Node.js wrapper would still have no
way to ask for an explicit zero. Omitting the field is unchanged; passing a
number, including `0`, now always sets it.

## Known limitations

- **Indentation has the same unset-tracking gap the spacing fix closed.** An
  explicit `SetIndent(domain.Indentation{Left: 720, Right: 0})` intended to
  zero out a style's right indent is indistinguishable from never touching
  `Right`, and is silently dropped in favor of the style's value. Tracked in
  [#76](https://github.com/mmonterroca/docxgo/issues/76).

## Compatibility

- **No public Go interface changed.** `domain.Paragraph` is exactly as it was;
  the new `SpacingBeforeSet()`/`SpacingAfterSet()`/`LineSpacingSet()` methods
  live only on the internal concrete type.
- **`template.Location`** gains `EndRunIndex` (additive). No non-test consumer
  of `RunIndex`/`StartOffset`/`EndOffset` exists in this repo (`cmd/docxgo`,
  `npm/`), and their values are unchanged for the common case of a match
  contained in a single run.
- **CLI JSON-RPC:** `spacingBefore`/`spacingAfter` remain optional; omitting
  them behaves as before. Passing an explicit `0` on a styled paragraph now
  behaves differently — see "Rendering change".
- `template.ConsolidateRuns`, `MergeTemplate`, and `FindPlaceholders` keep
  their existing signatures for every `domain.Paragraph` implementation,
  including third-party and wrapped ones.

# Release Notes - v2.9.1

**Release Date:** July 27, 2026

## Summary

v2.9.1 is a same-day patch on v2.9.0. A multi-agent code review of that
release's diff caught two real regressions before they reached many users:
an unintended direct line-spacing override, and an off-by-one in the new
`FindPlaceholders` location reporting. Both are fixed here, along with three
CI/workflow hardening issues the same review surfaced.

No public API changed. If you're already on v2.9.0, upgrade — there's no
reason to stay on it.

## Fixed

### An explicit spacing override no longer clobbers a style's line spacing

v2.9.0 fixed a real bug: `SetSpacingAfter(0)` on a paragraph carrying a style
with non-zero spacing used to be silently dropped, with the style's value
winning over the caller's explicit zero. The fix widened the condition for
emitting a direct `<w:spacing>` element to include "the caller explicitly set
spacingBefore or spacingAfter" — but the code that builds that element filled
in `Line` and `LineRule` unconditionally, regardless of *why* the element was
being emitted.

The consequence: a paragraph that called only `SetSpacingAfter(0)` — never
touching line spacing at all — gained an unintended direct
`w:line="240" w:lineRule="auto"`. If that paragraph's style set a real line
spacing (say, 1.5 lines for body text), the direct formatting silently
overrode it, rendering the paragraph single-spaced instead.

`Line` and `LineRule` are now gated by their own check — whether line spacing
itself was explicitly set or departs from the document defaults — independent
of whether `Before`/`After` triggered the element. Setting only
`spacingAfter` now emits only `w:after`, exactly as intended.

### FindPlaceholders reports the correct Location for the common case

v2.9.0 made `FindPlaceholders` stop mutating the documents it scans (see the
v2.9.0 notes). To find placeholders Word split across multiple runs without
merging them, it translates a match's position in a virtual concatenated text
back to the run and offset that actually contains it.

That translation used the same rule — "offset at or before this run's length
belongs to this run" — for both the start and the end of a match. That rule
is correct for the end (an exclusive offset), but wrong for the start: a
match beginning exactly where one run ends and the next begins would resolve
to the end of the *earlier* run instead of the start of the run that actually
holds it. This isn't an exotic edge case — `"Hello "` + `"{{Name}}"` is an
ordinary way for Word to split a paragraph, with the placeholder living
entirely in the second run. The bug also misattributed placeholders following
a leading empty run.

Concretely, this meant `Location.RunIndex`/`StartOffset` could point at a run
that doesn't contain the placeholder at all, and the natural way to use them —
`runs[loc.RunIndex].Text()[loc.StartOffset:loc.EndOffset]` — could panic with
a slice-bounds error.

The start and end lookups are now distinct: the end uses an inclusive
comparison (as before), the start uses a strict one, which also correctly
skips past empty runs to the run that actually starts the match.
`template.Location`'s doc comment now states the contract precisely: when
`RunIndex == EndRunIndex`, slicing a single run works exactly as it did in
every release before v2.9.0. A difference means the match is genuinely split
across runs — use `FullMatch`, or `MergeTemplate` to replace it.

## Changed

### CI is now resilient to a gap in the release process

`npm/package-lock.json` is regenerated during release prep with
`npm install --package-lock-only`, which necessarily drops the
platform-package entries (`@mmonterroca/docxgo-darwin-arm64` and siblings) —
they don't exist on the npm registry yet at the moment the release commit is
made. `npm ci`, used by CI's Node.js Tests job, requires every declared
`optionalDependency` to already have a resolved entry in the lock, so it
failed on any commit sitting between a release-prep commit and whenever the
lock next got regenerated. This reproduced identically after both the v2.8.0
and v2.9.0 release commits landed on `master`.

CI now uses `npm install`, which reconciles the manifest against the lock
instead of demanding they already agree exactly. This is what CI actually
needs here — correctness of the Node.js wrapper's own test suite, not
lockfile reproducibility for a gap the release process itself produces every
cycle.

### Hardened the release workflows

- `inputs.version` in `npm-publish.yml` — reachable via manual
  `workflow_dispatch` — is now bound through `env:` and validated against a
  semantic-version pattern before being used, instead of being interpolated
  directly into `run:` shell scripts in a job that holds `NPM_TOKEN` and
  `id-token: write`.
- Removed the now-redundant `release: types: [published]` trigger from
  `npm-publish.yml`. v2.9.0 added a `publish-npm` job to `release.yml` that
  calls `npm-publish.yml` directly and unconditionally; keeping the old event
  trigger too meant that once `RELEASE_PAT` is configured (see v2.9.0's
  notes), a single tag would fire the publish workflow twice, racing two
  `npm publish` calls for the same version.
- `release.yml`'s `RELEASE_PAT` emptiness check now reads the secret through
  `env:` rather than expanding it directly into a shell `if` condition.

## Compatibility

- **No public Go, CLI protocol, or Node.js API changed.**
- **Generated output** for a paragraph that both carries a style with
  non-default line spacing *and* explicitly sets only `spacingBefore`/
  `spacingAfter` (never touching line spacing) changes back to what v2.9.0
  intended: the style's line spacing is inherited again, rather than being
  overridden by an unintended direct `0/240/auto`. No other paragraph shape
  is affected.
- **`FindPlaceholders`/`FindPlaceholdersCustom`/`PlaceholderNames`/
  `ValidateTemplate`** report corrected `RunIndex`/`StartOffset` for matches
  sitting exactly at a run boundary or following a leading empty run — a
  narrow shape unlikely to have been observed in the few hours v2.9.0 shipped
  before this fix.

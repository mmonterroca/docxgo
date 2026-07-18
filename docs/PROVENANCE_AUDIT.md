# docxgo v2 — License and provenance determination

**Determination:** **PASS — the current docxgo v2 release may remain MIT.**

**AGPL scope:** the repository contains historical versions of
`fumiama/go-docx` that were distributed under AGPL-3.0. Those historical
versions remain AGPL-3.0. The audited v2.7.1 baseline and the v2.7.2 release
candidate do not contain protectable implementation copied from the AGPL-only
work, so AGPL-3.0 does not attach to the current release merely because the
commits share Git ancestry.

**Audited on:** 2026-07-18

**Release candidate:** `v2.7.2`

**Audited release baseline:** `v2.7.1` at
`9dbac7a4afbda67df296db3c31de52038146ff11`

**Integrated branch base:** `5fd78ad` (same Go implementation as the audited
baseline, plus the v2.7.1 release notes)

**Upstream boundary:** `0c30fd09304b17fdb42b0dcea142962b2f4883a3`

This is the project's technical and open-source compliance determination. It
is not a court opinion and does not replace counsel for a transaction-specific
warranty or indemnity, but legal review is **not an unresolved prerequisite**
to keeping the project MIT and publishing releases with the corrected notices
now present in the repository.

---

## Decision in practical terms

| Material | License determination | Action |
|---|---|---|
| Current source tree (`v2.7.2` candidate) | MIT | Keep `LICENSE` and SPDX headers as MIT |
| Go module `github.com/mmonterroca/docxgo/v2@v2.7.2` | MIT | No AGPL action required |
| Main npm package `@mmonterroca/docxgo@2.7.2` | MIT | No AGPL action required |
| Platform npm and GitHub release binaries | MIT | Include the MIT text in every binary archive/package |
| Commits `70ec491d` through upstream `0c30fd0` | AGPL-3.0 | Do not represent or reuse those historical snapshots as MIT |
| Pre-relicense upstream material through `c983cd71` | MIT | May be used under its preserved MIT notice |

The correct public description is:

> docxgo v2 is an MIT-licensed, substantially rewritten successor whose Git
> history descends from `fumiama/go-docx`. The history includes AGPL-era
> upstream snapshots; the current release does not include protectable AGPL
> implementation.

Do not describe the work as a “clean-room rewrite”: v2 was developed in the
same repository and with knowledge of v1. Also do not claim that AGPL “never
entered the repository”: it is plainly present in historical commits. Neither
statement is needed for the MIT determination.

The v2.7.2 delta was reviewed separately. It changes the version constant,
license and credit text, documentation, and release packaging workflows; it
does not restore or copy implementation from `legacy/v1` or any AGPL-era
commit. The current-tree comparison was rerun against the candidate with the
same result described below.

---

## Why Git ancestry is not the license test

Git ancestry proves where the repository came from; it does not make every
later tree a copy of every earlier tree. Copyright and the AGPL instead ask
whether the material being distributed copies or adapts protected expression
from a covered work:

- [17 U.S.C. §101](https://www.copyright.gov/title17/92chap1.html#101)
  defines a derivative work as one in which a preexisting work is recast,
  transformed, or adapted.
- [17 U.S.C. §102(b)](https://www.copyright.gov/title17/92chap1.html#102)
  excludes ideas, procedures, processes, systems, methods of operation,
  concepts, principles, and discoveries from copyright protection. The
  legislative notes specifically distinguish program expression from program
  methodology.
- AGPL-3.0 §0 defines a modified or “based on” work by copying or adaptation
  that requires copyright permission. Section 5 requires an entire work to be
  AGPL when it is such a covered work; it does not say that common repository
  history alone creates coverage. See the
  [official AGPL-3.0 text](https://www.gnu.org/licenses/agpl-3.0.en.html#section0)
  and [§5](https://www.gnu.org/licenses/agpl-3.0.en.html#section5).

Accordingly, ancestry is an important reason to inspect closely, but the
decision turns on retained protected expression. The audit found none from the
AGPL-only implementation in the current release.

---

## Provenance timeline

| Date | Event | Applicable license to that snapshot |
|---|---|---|
| 2020–2021 | `gingfrederik/docx` → `gonfva/docxlib` | MIT |
| 2023-02-08 | fumiama begins work on the inherited MIT code | MIT |
| 2023-02-24 15:58 +08:00 | Last pre-relicense commit `c983cd71` | MIT |
| 2023-02-24 16:14 +08:00 | `70ec491d` replaces the license and adds AGPL headers | AGPL-3.0 |
| 2023-02 to 2025-05 | fumiama continues development through `0c30fd0` | AGPL-3.0 |
| 2025-10-24 | `fcb23ff` creates the separate `v2/` architecture and domain model | Authored for docxgo v2 |
| 2025-10-25 | `9b140aa`, `c3e91c6`, and `52053d8` add the new core, XML, serializer, and writer layers | Authored for docxgo v2 |
| 2025-10-25 | `aa5b7ce` promotes `v2/` to the repository root and archives v1 | Authored for docxgo v2 |
| 2025-10-26 | `c120cb6` removes the archived v1 implementation from the current tree | Authored for docxgo v2 |
| 2026-07-18 | v2.7.1 release at `9dbac7a` | MIT |

The ancestry is reproducible:

```sh
# Upstream's complete 140-commit history is an ancestor of v2.7.1.
git merge-base --is-ancestor \
  0c30fd09304b17fdb42b0dcea142962b2f4883a3 \
  9dbac7a4afbda67df296db3c31de52038146ff11

git rev-list --count 0c30fd09304b17fdb42b0dcea142962b2f4883a3
# 140

# The license transition itself is directly inspectable.
git diff 70ec491d^ 70ec491d -- LICENSE
```

This proves that historical AGPL snapshots are in the repository. It does not
prove that their protected code remains in the current tree.

---

## Current-tree evidence

### 1. v2 was introduced as a separate implementation

The rewrite is visible in the history rather than inferred from a low overlap
score:

- `fcb23ff` created `v2/domain/*` and a new architecture document.
- `9b140aa` created the new `v2/internal/core`, `manager`, and public utility
  packages.
- `c3e91c6` created the new typed XML model and serializer.
- `52053d8` created the new ZIP writer and public v2 entry point.
- `aa5b7ce` promoted those files to the root. The old implementation was moved
  to `legacy/v1`, then deleted by `c120cb6`.

Relative to upstream `0c30fd0`, the current Go tree records 38,143 added and
8,744 deleted lines. The upstream is a flat, polymorphic OOXML implementation;
v2 separates domain interfaces, core state, readers, serializers, writers, and
managers. This is a different implementation structure, not a mechanical
rename of the upstream packages.

### 2. No non-empty current Go line is inherited from the AGPL period

All 102 tracked `.go` files were checked with `git blame` against the upstream
boundary. Five current Go lines receive an upstream blame assignment:

- one non-empty line is from MIT-era commit `56810d8` and is only a closing
  brace;
- two other MIT-era lines are empty;
- the two lines assigned to AGPL-era commits (`70ec491d` and `daf7190`) are
  empty lines.

Thus **zero non-empty Go source lines** in v2.7.1 are blame-attributed to an
AGPL-era commit. Git's blame matching is not itself a copyright test, but this
result is strong direct evidence that the old implementation was removed.

### 3. No current file is an unchanged AGPL-era blob

The v2.7.1 tree contains 185 tracked files. Comparing its blob hashes against
both the `0c30fd0` tree and the objects introduced in the AGPL-era range yields
no match:

```sh
# Exact blobs shared with the final upstream tree: no output.
comm -12 \
  <(git ls-tree -r 0c30fd0 | awk '{print $3}' | sort -u) \
  <(git ls-tree -r 9dbac7a | awk '{print $3}' | sort -u)

# Current blobs identical to objects introduced in the AGPL range: no output.
comm -12 \
  <(git rev-list --objects 70ec491d^..0c30fd0 | awk '{print $1}' | sort -u) \
  <(git ls-tree -r 9dbac7a | awk '{print $3}' | sort -u)
```

### 4. Exact line overlap is small and non-protectable in character

The reproducible comparison script
[`docs/provenance/compare_line_overlap.py`](provenance/compare_line_overlap.py)
checks whitespace-normalized, non-comment lines of at least 25 characters.
Against all 36 non-test Go files at upstream `0c30fd0`, it reports:

- 11,565 distinctive lines across docxgo's 102 Go files;
- 47 exact matches, or **0.41%**;
- 43 matches are type/field declarations dominated by OOXML names and Go XML
  struct tags;
- four are generic test/append idioms;
- no match is an upstream algorithm or distinctive function body.

The current tree was also checked against the union of distinctive lines first
added during the full AGPL period, excluding every line ever present in the
pre-relicense history. It found 21 exact matches:

- 20 are OOXML field declarations such as `w:val`, `w:color`, `w:top`, and
  `xml:space` represented using Go's required `encoding/xml` tag syntax;
- one is the generic Go statement `items = append(items, item)`.

ECMA-376 defines the OOXML vocabulary and representation requirements; see the
[official ECMA-376 description](https://ecma-international.org/publications-and-standards/standards/ecma-376/).
Go's standard library in turn requires the `name,attr` tag form for named XML
attributes; see the official
[`encoding/xml` mapping rules](https://pkg.go.dev/encoding/xml#Marshal).
Those names and mapping forms are external compatibility constraints, not
creative implementation copied from fumiama.

A whole-text history check (normalized lines of at least 40 characters) found
one additional exact line in the built-in
default Office theme: an `a:outerShdw` parameter line. The same default-theme
line is publicly documented in Office-generated files predating fumiama's
project—for example, in this
[2021 Office XML example](https://learn.microsoft.com/en-us/answers/questions/324367/marker-settings-in-the-style1-xml-file)—and
is data describing an OOXML effect, not fumiama-authored program logic. No
other match appeared in that whole-text check.

### 5. Non-literal structure was checked as well

Textual difference alone would not be sufficient if v2 preserved the
upstream's distinctive structure or organization. It does not:

- upstream uses a flat document model with types such as `WTableCell`,
  `WGridSpan`, `WvMerge`, `WPAnchor`, `RunMergeRule`, and `Kinsoku`; these do
  not appear in v2;
- upstream's `Run` uses polymorphic `Children []interface{}` plus a custom token
  walk and a document back-pointer; v2's `Run` uses typed optional fields and
  ordinary `encoding/xml` mapping without that coupling;
- current implementation logic in `internal/serializer/serializer.go`,
  `internal/serializer/latent_styles.go`, and `internal/writer/zip.go` has no
  exact upstream Go-line overlap;
- the few shared method/type names introduced during the AGPL period are words
  such as `Bold`, `Italic`, `Underline`, `Spacing`, and `Table`: functional
  vocabulary necessary to expose Word-formatting operations, not distinctive
  implementation structure.

This is enough to reject the prior document's assumption that sequential Git
history, by itself, makes the current release a “strong candidate” derivative.
The history warranted the audit; the inspected release content resolves it.

---

## Dependencies and published artifacts

### Go

`go list -m all` reports only `github.com/mmonterroca/docxgo/v2`; the module
has no third-party Go runtime dependency. The public module proxy resolves
`v2.7.1` to release commit `9dbac7a` with checksum
`h1:8OsL3czY5f4TuYIdg7lXSisanAmWoBxJgnb0E2A6Zmc=`. Its module ZIP contains the
185-file release snapshot, not `.git` history. Go documents that module
versions are distributed as ZIP snapshots in the
[Go Modules Reference](https://go.dev/ref/mod#zip-files).

The tracked `04_tech_architecture` example binary was also inspected with
`go version -m`: it was built from post-rewrite docxgo code and records no
external module dependencies. The root MIT license accompanies it in the Go
module ZIP.

### npm

The published main package `@mmonterroca/docxgo@2.7.1` contains the compiled
TypeScript wrapper, source maps, README, package metadata, and the MIT license.
It bundles no third-party JavaScript dependency. Its optional dependencies are
the five docxgo platform binaries, which are built from the audited Go module.

The published v2.7.1 platform packages and GitHub binary archives declare or
derive from MIT code but omit the license text. That is an MIT
notice-packaging defect, **not evidence of AGPL coverage**. The repository's
publish workflows now copy the root `LICENSE` into every platform package and
binary archive. The fix applies to the next published version because npm
releases and existing GitHub release assets are immutable records.

---

## License notice determination

The MIT license permits use, modification, distribution, sublicensing, and
sale, subject to retaining its copyright and permission notice; see the
[OSI MIT text](https://opensource.org/license/mit).

The v2.7.1 notice used an over-broad “2023–2025 fumiama” attribution inside the
MIT file. That wording did not create an MIT grant from fumiama for AGPL-era
work and was misleading even though no such protected work remains in the
release. The repository notice now preserves the exact MIT-era predecessor
line, “Copyright (c) 2023 Fumiama Minamoto (源文雨)”, without labeling the later
2023–2025 AGPL work as MIT. Later authorship remains fully credited in
`CREDITS.md`, but credit is not a relicense.

The current SPDX headers identify the license of the current files. They do
not and cannot retroactively change a historical commit's license.

---

## Guardrails that keep this determination valid

Re-run the audit before a release if any of the following occurs:

1. code or assets are restored from `legacy/v1` or an AGPL-era commit;
2. a commit is cherry-picked from `fumiama/go-docx` after `70ec491d`;
3. current files are regenerated from AGPL upstream sources;
4. a new runtime dependency is added; or
5. a contributor identifies a specific non-generic current passage allegedly
   copied from AGPL-only code.

Ordinary feature work written against ECMA-376, the public API, and current v2
code does not reopen the determination. Preserve this audit, the comparison
script, the corrected MIT notice, and the platform-package license copy step.

---

## Final conclusion

The prior ambiguity came from treating a Git commit graph as if it were a
source-code diff. The repository does descend through an AGPL period,
and those old commits remain AGPL. But the released v2 implementation was
created in a separate tree, the old implementation was removed, no current
file is an AGPL-era blob, no non-empty Go line is blame-inherited from that
period, and the residual exact matches are standard-constrained declarations,
generic idioms, and default OOXML data.

**Project determination: docxgo v2.7.2 and subsequent releases that respect the
guardrails above may be distributed under MIT. No project-wide AGPL relicense
is required, and no further licensing consultation is an open release task.**

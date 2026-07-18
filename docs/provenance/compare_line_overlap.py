#!/usr/bin/env python3
"""Reproduce the docxgo vs fumiama/go-docx line-level provenance comparison
described in docs/PROVENANCE_AUDIT.md.

Usage:
    git clone https://github.com/fumiama/go-docx.git /tmp/fumiama-go-docx
    git -C /tmp/fumiama-go-docx checkout 0c30fd09304b17fdb42b0dcea142962b2f4883a3
    python3 docs/provenance/compare_line_overlap.py /tmp/fumiama-go-docx .

Methodology: a line is "distinctive" if, after stripping comments and
collapsing whitespace, it is at least 25 characters long. The fumiama corpus
is the set of all such lines across every non-test .go file in its tree
(this is the AGPL-era corpus we're checking against). For each docxgo .go
file, we report how many of its distinctive lines also appear verbatim in
that corpus.

This is a coarse, line-set-membership check, not an AST or semantic diff —
by design: it flags any verbatim textual overlap at all (including
schema-dictated struct tags, which are then manually reviewed for whether
they're independently protectable, per docs/PROVENANCE_AUDIT.md Finding 2).
"""
import os
import re
import sys
from pathlib import Path


def normalize(line):
    return re.sub(r"\s+", " ", line.strip())


def is_comment_only(line):
    s = line.strip()
    return s.startswith("//") or s.startswith("/*") or s.startswith("*") or s == "*/"


def strip_block_comments(text):
    return re.sub(r"/\*.*?\*/", "", text, flags=re.DOTALL)


def distinctive_lines(filepath):
    try:
        text = Path(filepath).read_text(encoding="utf-8", errors="replace")
    except OSError:
        return set()
    text = strip_block_comments(text)
    lines = set()
    for raw in text.splitlines():
        if is_comment_only(raw):
            continue
        norm = normalize(raw)
        if len(norm) >= 25:
            lines.add(norm)
    return lines


def collect_corpus(root, include_tests=False):
    corpus = set()
    file_count = 0
    for dirpath, _dirnames, filenames in os.walk(root):
        if ".git" in dirpath.split(os.sep):
            continue
        for fn in filenames:
            if not fn.endswith(".go"):
                continue
            if not include_tests and fn.endswith("_test.go"):
                continue
            corpus |= distinctive_lines(os.path.join(dirpath, fn))
            file_count += 1
    return corpus, file_count


def main():
    if len(sys.argv) != 3:
        print(__doc__)
        sys.exit(1)
    fumiama_root, docxgo_root = sys.argv[1], sys.argv[2]

    fumiama_corpus, fumiama_files = collect_corpus(fumiama_root, include_tests=False)
    print(f"fumiama/go-docx: {fumiama_files} non-test .go files, "
          f"{len(fumiama_corpus)} distinctive lines\n")

    results = []
    total_docxgo_files = 0
    grand_total_lines = 0
    grand_total_matches = 0

    for dirpath, _dirnames, filenames in os.walk(docxgo_root):
        parts = dirpath.split(os.sep)
        if ".git" in parts or ".claude" in parts:
            continue
        for fn in sorted(filenames):
            if not fn.endswith(".go"):
                continue
            fp = os.path.join(dirpath, fn)
            rel = os.path.relpath(fp, docxgo_root)
            total_docxgo_files += 1
            dlines = distinctive_lines(fp)
            if not dlines:
                continue
            matched = dlines & fumiama_corpus
            grand_total_lines += len(dlines)
            grand_total_matches += len(matched)
            if matched:
                results.append((rel, len(matched), len(dlines), matched))

    results.sort(key=lambda r: -(r[1] / r[2]))
    print(f"docxgo: {total_docxgo_files} .go files scanned\n")
    print(f"{'file':<45} {'overlap':>8} {'matched/total':>15}")
    for rel, m, t, _matched in results:
        pct = 100.0 * m / t
        print(f"{rel:<45} {pct:>7.1f}% {m:>6}/{t:<8}")

    print(f"\nTotal distinctive lines across all docxgo .go files: {grand_total_lines}")
    print(f"Total matched lines: {grand_total_matches}")

    print("\n=== Matched line detail per file ===")
    for rel, m, _t, matched in results:
        print(f"\n--- {rel} ({m} matches) ---")
        for line in sorted(matched):
            print(f"  {line}")


if __name__ == "__main__":
    main()

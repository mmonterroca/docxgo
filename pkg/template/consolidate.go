// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package template

import (
	"fmt"

	"github.com/mmonterroca/docxgo/v2/domain"
)

// consolidateErr wraps a run-rebuild failure with a consistent operation prefix.
func consolidateErr(err error) error {
	return fmt.Errorf("template: consolidate runs: %w", err)
}

// runAppender is implemented by paragraph types that can re-attach an
// existing run verbatim (see internal/core.paragraph.AppendRun). Consolidation
// requires it: an untouched run (one that didn't merge with a neighbor) must
// be carried over as-is, since rebuilding it via AddRun + the domain.Run
// setters cannot copy content the interface doesn't expose, such as images.
type runAppender interface {
	AppendRun(domain.Run) error
}

// ConsolidateRuns merges adjacent runs with identical formatting in a paragraph.
// This heals the "split placeholder" problem where Word fragments tokens like
// {{name}} across multiple <w:r> elements due to spell-check, proofing, or
// editing history boundaries.
//
// Only text-only runs (no fields, breaks, or images) are eligible for merging.
// The merged run retains the formatting of the first run in each merge group.
// A run that doesn't merge with any neighbor is carried over unchanged — it
// is never copied through AddRun, so content the domain.Run interface can't
// express field-by-field (namely images) survives consolidation.
//
// This function modifies the paragraph in place via ClearRuns/AddRun. If a
// run setter fails partway through the rebuild, it stops immediately and
// returns that error rather than silently continuing with a partially
// rebuilt paragraph.
//
// It is called automatically by MergeTemplate and FindPlaceholders.
func ConsolidateRuns(para domain.Paragraph) error {
	runs := para.Runs()
	if len(runs) <= 1 {
		return nil
	}

	appender, ok := para.(runAppender)
	if !ok {
		return consolidateErr(fmt.Errorf("paragraph type %T cannot preserve non-mergeable runs (e.g. images) during consolidation", para))
	}

	// Build merged groups: each group is a sequence of adjacent text-only runs
	// with identical formatting that will be combined into a single run.
	// count tracks how many source runs fed the group: a group of 1 is an
	// untouched run and must be re-attached verbatim, not rebuilt.
	type mergedRun struct {
		text  string
		src   domain.Run // first run in the group (provides formatting)
		count int
	}

	merged := make([]mergedRun, 0, len(runs))
	merged = append(merged, mergedRun{text: runs[0].Text(), src: runs[0], count: 1})

	for i := 1; i < len(runs); i++ {
		prev := runs[i-1]
		curr := runs[i]

		// Merge if both are text-only and have identical formatting
		if isTextOnly(prev) && isTextOnly(curr) && formatsEqual(prev, curr) {
			// Append text to the current merge group
			merged[len(merged)-1].text += curr.Text()
			merged[len(merged)-1].count++
		} else {
			// Start a new group
			merged = append(merged, mergedRun{text: curr.Text(), src: curr, count: 1})
		}
	}

	// If nothing was merged, skip the rebuild
	if len(merged) == len(runs) {
		return nil
	}

	// Rebuild the paragraph's runs
	para.ClearRuns()
	for _, m := range merged {
		if m.count == 1 {
			// Untouched run: re-attach the original object instead of copying
			// it into a fresh run, so content AddRun+setters can't express
			// (images) isn't dropped.
			if err := appender.AppendRun(m.src); err != nil {
				return consolidateErr(err)
			}
			continue
		}

		r, err := para.AddRun()
		if err != nil {
			return consolidateErr(err)
		}
		// Copy formatting from the source run, stopping at the first failure
		// instead of continuing with a partially-formatted run.
		if err := r.SetText(m.text); err != nil {
			return consolidateErr(err)
		}
		if err := r.SetFont(m.src.Font()); err != nil {
			return consolidateErr(err)
		}
		if err := r.SetColor(m.src.Color()); err != nil {
			return consolidateErr(err)
		}
		if err := r.SetSize(m.src.Size()); err != nil {
			return consolidateErr(err)
		}
		if err := r.SetBold(m.src.Bold()); err != nil {
			return consolidateErr(err)
		}
		if err := r.SetItalic(m.src.Italic()); err != nil {
			return consolidateErr(err)
		}
		if err := r.SetUnderline(m.src.Underline()); err != nil {
			return consolidateErr(err)
		}
		if err := r.SetStrike(m.src.Strike()); err != nil {
			return consolidateErr(err)
		}
		if err := r.SetHighlight(m.src.Highlight()); err != nil {
			return consolidateErr(err)
		}

		// Re-add fields, breaks from the source run (for non-text runs)
		for _, f := range m.src.Fields() {
			if err := r.AddField(f); err != nil {
				return consolidateErr(err)
			}
		}
		for _, b := range m.src.Breaks() {
			if err := r.AddBreak(b); err != nil {
				return consolidateErr(err)
			}
		}
	}

	return nil
}

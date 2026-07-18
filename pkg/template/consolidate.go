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

// ConsolidateRuns merges adjacent runs with identical formatting in a paragraph.
// This heals the "split placeholder" problem where Word fragments tokens like
// {{name}} across multiple <w:r> elements due to spell-check, proofing, or
// editing history boundaries.
//
// Only text-only runs (no fields, breaks, or images) are eligible for merging.
// The merged run retains the formatting of the first run in each merge group.
//
// This function modifies the paragraph in place via ClearRuns/RemoveRun. If a
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

	// Build merged groups: each group is a sequence of adjacent text-only runs
	// with identical formatting that will be combined into a single run.
	type mergedRun struct {
		text string
		src  domain.Run // first run in the group (provides formatting)
	}

	merged := make([]mergedRun, 0, len(runs))
	merged = append(merged, mergedRun{text: runs[0].Text(), src: runs[0]})

	for i := 1; i < len(runs); i++ {
		prev := runs[i-1]
		curr := runs[i]

		// Merge if both are text-only and have identical formatting
		if isTextOnly(prev) && isTextOnly(curr) && formatsEqual(prev, curr) {
			// Append text to the current merge group
			merged[len(merged)-1].text += curr.Text()
		} else {
			// Start a new group
			merged = append(merged, mergedRun{text: curr.Text(), src: curr})
		}
	}

	// If nothing was merged, skip the rebuild
	if len(merged) == len(runs) {
		return nil
	}

	// Rebuild the paragraph's runs
	para.ClearRuns()
	for _, m := range merged {
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

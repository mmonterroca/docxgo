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
// Each merge group keeps its own first run — the group's combined text is
// written onto that run and the runs it absorbed are removed. Nothing is ever
// copied into a freshly created run, so a run's formatting and any content the
// domain.Run interface cannot express field-by-field (namely images) is
// preserved by construction rather than by an explicit copy that has to be
// kept in sync with the interface.
//
// This function modifies the paragraph in place via SetText/RemoveRun. If a
// mutation fails partway through, it stops immediately and returns that error
// rather than silently continuing with a partially consolidated paragraph.
//
// It is called automatically by MergeTemplate and FindPlaceholders.
func ConsolidateRuns(para domain.Paragraph) error {
	runs := para.Runs()
	if len(runs) <= 1 {
		return nil
	}

	// Build merge groups over the original run indices. Each group is a
	// sequence of adjacent text-only runs with identical formatting; leader is
	// the index of the run that survives and receives the combined text.
	type mergeGroup struct {
		text   string
		leader int
		count  int
	}

	groups := make([]mergeGroup, 0, len(runs))
	groups = append(groups, mergeGroup{text: runs[0].Text(), leader: 0, count: 1})

	for i := 1; i < len(runs); i++ {
		prev := runs[i-1]
		curr := runs[i]

		// Merge if both are text-only and have identical formatting
		if isTextOnly(prev) && isTextOnly(curr) && formatsEqual(prev, curr) {
			// Absorb this run's text into the current group
			groups[len(groups)-1].text += curr.Text()
			groups[len(groups)-1].count++
		} else {
			// Start a new group
			groups = append(groups, mergeGroup{text: curr.Text(), leader: i, count: 1})
		}
	}

	// If nothing was merged, there is nothing to rewrite.
	if len(groups) == len(runs) {
		return nil
	}

	// Write each multi-run group's combined text onto its leader. Groups of
	// one are left completely untouched — their text is already correct.
	absorbed := make([]bool, len(runs))
	for _, g := range groups {
		if g.count == 1 {
			continue
		}
		if err := runs[g.leader].SetText(g.text); err != nil {
			return consolidateErr(err)
		}
		for i := g.leader + 1; i < g.leader+g.count; i++ {
			absorbed[i] = true
		}
	}

	// Drop the absorbed runs, highest index first so earlier indices stay valid.
	for i := len(runs) - 1; i >= 0; i-- {
		if !absorbed[i] {
			continue
		}
		if err := para.RemoveRun(i); err != nil {
			return consolidateErr(err)
		}
	}

	return nil
}

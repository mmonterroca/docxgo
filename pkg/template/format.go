// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package template

import "github.com/mmonterroca/docxgo/v2/domain"

// formatsEqual returns true if two runs have identical visible formatting.
// It compares the 8 visual formatting attributes plus Language — not
// visible, but ConsolidateRuns must not merge two runs with different
// language overrides, since the merge keeps only the leader's formatting
// (see ConsolidateRuns) and silently discarding one run's Language would
// wrongly re-tag a foreign-language phrase (domain.Run.SetLanguage; e.g. a
// [bonjour]{lang=fr} span in otherwise French-unmarked prose) as the
// surrounding document's default language the moment MergeTemplate or
// ReplaceText happened to run over it.
func formatsEqual(a, b domain.Run) bool {
	return a.Font() == b.Font() &&
		a.Color() == b.Color() &&
		a.Size() == b.Size() &&
		a.Bold() == b.Bold() &&
		a.Italic() == b.Italic() &&
		a.Underline() == b.Underline() &&
		a.Strike() == b.Strike() &&
		a.Highlight() == b.Highlight() &&
		languagesEqual(a.Language(), b.Language())
}

// languagesEqual is a nil-safe comparison of two *domain.Language: both
// unset, or both set to the same Val/EastAsia/Bidi.
func languagesEqual(a, b *domain.Language) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// isTextOnly returns true if the run contains only text (no fields, breaks, or images).
// Runs with non-text content must not be merged during consolidation.
func isTextOnly(r domain.Run) bool {
	return len(r.Fields()) == 0 &&
		len(r.Breaks()) == 0 &&
		r.Image() == nil
}

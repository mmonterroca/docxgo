// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package template

import "github.com/mmonterroca/docxgo/v2/domain"

// formatsEqual returns true if two runs have identical visible formatting.
// It compares the 9 visual formatting attributes plus Language — not
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
		a.Caps() == b.Caps() &&
		capsSetOf(a) == capsSetOf(b) &&
		a.Highlight() == b.Highlight() &&
		languagesEqual(a.Language(), b.Language())
}

// capsExplicitSetter mirrors internal/serializer's capsSetter: it exposes
// whether a run's Caps was ever explicitly set (true or false), as opposed
// to defaulting to false because it was never touched. Implemented by the
// concrete run type in internal/core; degrades to "never explicitly set" for
// any other domain.Run implementation.
type capsExplicitSetter interface {
	CapsSet() bool
}

// capsSetOf compares alongside Caps() in formatsEqual so a run with an
// explicit false (which must serialize as <w:caps w:val="false"/> to
// override a style's own All Caps) never silently merges into a neighbor
// that shares the same false Caps() value but was never explicitly set --
// merging would drop the leader's un-set state onto the absorbed text and
// serialize the merged run with no <w:caps> at all, letting a style's All
// Caps apply where the source explicitly turned it off.
func capsSetOf(r domain.Run) bool {
	if cs, ok := r.(capsExplicitSetter); ok {
		return cs.CapsSet()
	}
	return false
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

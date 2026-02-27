package template

import "github.com/mmonterroca/docxgo/v2/domain"

// formatsEqual returns true if two runs have identical visible formatting.
// It compares all 8 formatting attributes that affect text appearance.
// This is used by ConsolidateRuns to decide if adjacent runs can be merged.
func formatsEqual(a, b domain.Run) bool {
	return a.Font() == b.Font() &&
		a.Color() == b.Color() &&
		a.Size() == b.Size() &&
		a.Bold() == b.Bold() &&
		a.Italic() == b.Italic() &&
		a.Underline() == b.Underline() &&
		a.Strike() == b.Strike() &&
		a.Highlight() == b.Highlight()
}

// isTextOnly returns true if the run contains only text (no fields, breaks, or images).
// Runs with non-text content must not be merged during consolidation.
func isTextOnly(r domain.Run) bool {
	return len(r.Fields()) == 0 &&
		len(r.Breaks()) == 0 &&
		r.Image() == nil
}

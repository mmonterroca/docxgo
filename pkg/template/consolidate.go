package template

import "github.com/mmonterroca/docxgo/v2/domain"

// ConsolidateRuns merges adjacent runs with identical formatting in a paragraph.
// This heals the "split placeholder" problem where Word fragments tokens like
// {{name}} across multiple <w:r> elements due to spell-check, proofing, or
// editing history boundaries.
//
// Only text-only runs (no fields, breaks, or images) are eligible for merging.
// The merged run retains the formatting of the first run in each merge group.
//
// This function modifies the paragraph in place via ClearRuns/RemoveRun.
// It is called automatically by MergeTemplate and FindPlaceholders.
func ConsolidateRuns(para domain.Paragraph) {
	runs := para.Runs()
	if len(runs) <= 1 {
		return
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
		return
	}

	// Rebuild the paragraph's runs
	para.ClearRuns()
	for _, m := range merged {
		r, err := para.AddRun()
		if err != nil {
			continue // should not happen in practice
		}
		// Copy formatting from the source run
		_ = r.SetText(m.text)
		_ = r.SetFont(m.src.Font())
		_ = r.SetColor(m.src.Color())
		_ = r.SetSize(m.src.Size())
		_ = r.SetBold(m.src.Bold())
		_ = r.SetItalic(m.src.Italic())
		_ = r.SetUnderline(m.src.Underline())
		_ = r.SetStrike(m.src.Strike())
		_ = r.SetHighlight(m.src.Highlight())

		// Re-add fields, breaks from the source run (for non-text runs)
		for _, f := range m.src.Fields() {
			_ = r.AddField(f)
		}
		for _, b := range m.src.Breaks() {
			_ = r.AddBreak(b)
		}
	}
}

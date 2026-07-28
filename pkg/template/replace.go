// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package template

import (
	"fmt"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
)

// ReplaceResult reports the outcome of a ReplaceText call.
type ReplaceResult struct {
	// Replaced is the number of occurrences that were replaced.
	Replaced int
	// Skipped is the number of occurrences that were found but left
	// untouched because replacing them would not survive serialization.
	// See ReplaceText for exactly which run contents cause a skip.
	Skipped int
}

// ReplaceText replaces every occurrence of find with replace across the
// document: body paragraphs, table cells, headers, and footers. The document
// is modified in place.
//
// Matching is literal and case-sensitive. Runs in each paragraph are first
// consolidated (see ConsolidateRuns), so text fragmented across runs with
// identical formatting is matched. A match that spans runs with different
// formatting is replaced too: the replacement takes the formatting of the
// first spanned run.
//
// A match touching a run that carries a field is always skipped and counted
// in ReplaceResult.Skipped — such a run's text is never serialized, so
// rewriting it would report a replacement Word does not make. A match
// spanning several runs is likewise skipped when any of them carries a break
// or an image, since blanking a spanned run would strand that content in the
// middle of the replacement. A match confined to a single run carrying a
// break or an image is replaced, and the break or image is preserved.
func ReplaceText(doc domain.Document, find, replace string) (ReplaceResult, error) {
	if find == "" {
		return ReplaceResult{}, fmt.Errorf("template: find text must not be empty")
	}

	var result ReplaceResult
	err := walkParagraphs(doc, func(para domain.Paragraph, _ paragraphContext) error {
		replaced, skipped, err := replaceInParagraph(para, find, replace)
		result.Replaced += replaced
		result.Skipped += skipped
		return err
	})
	return result, err
}

// runSpan maps a run to its byte-offset range [start, end) within the
// paragraph's concatenated text.
type runSpan struct {
	run        domain.Run
	start, end int
}

// replaceInParagraph replaces occurrences of find within a single paragraph.
// After each successful replacement the paragraph text is rebuilt and the
// scan resumes just past the inserted replacement, so a replacement that
// itself contains find is never re-matched.
func replaceInParagraph(para domain.Paragraph, find, replace string) (replaced, skipped int, err error) {
	// Consolidation never changes a paragraph's concatenated text, only run
	// boundaries — so a paragraph with no match anywhere must be left
	// untouched rather than restructured as a side effect of searching it.
	if _, full := paragraphSpans(para); !strings.Contains(full, find) {
		return 0, 0, nil
	}

	if err := ConsolidateRuns(para); err != nil {
		return 0, 0, err
	}

	cursor := 0
	for {
		spans, full := paragraphSpans(para)
		if cursor >= len(full) {
			break
		}
		rel := strings.Index(full[cursor:], find)
		if rel < 0 {
			break
		}
		start := cursor + rel
		end := start + len(find)

		ok, err := replaceSpan(spans, start, end, replace)
		if err != nil {
			return replaced, skipped, err
		}
		if ok {
			replaced++
			cursor = start + len(replace)
		} else {
			skipped++
			cursor = end
		}
	}
	return replaced, skipped, nil
}

// paragraphSpans returns the paragraph's runs with their byte-offset ranges
// and the concatenated paragraph text.
func paragraphSpans(para domain.Paragraph) ([]runSpan, string) {
	runs := para.Runs()
	spans := make([]runSpan, 0, len(runs))
	var sb strings.Builder
	off := 0
	for _, r := range runs {
		t := r.Text()
		spans = append(spans, runSpan{run: r, start: off, end: off + len(t)})
		sb.WriteString(t)
		off += len(t)
	}
	return spans, sb.String()
}

// replaceSpan writes replace over the byte range [start, end) of the
// paragraph text, subject to the run-content rules documented on
// ReplaceText. Returns false (and no error) when the match must be skipped.
func replaceSpan(spans []runSpan, start, end int, replace string) (bool, error) {
	// Runs that actually carry matched text. Zero-width runs contribute no
	// match text, so only these are ever rewritten below — but they are not
	// the whole story for deciding whether the match may be rewritten at
	// all, which is what the two scans further down widen.
	lo, hi := -1, -1
	var spanned []runSpan
	for i, s := range spans {
		if s.end > s.start && s.start < end && s.end > start {
			if lo < 0 {
				lo = i
			}
			hi = i
			spanned = append(spanned, s)
		}
	}
	if len(spanned) == 0 {
		return false, nil
	}

	// A field-bearing run's text is never serialized: the paragraph
	// serializer routes it through expandRunWithFields, which emits the
	// field's own result instead. Rewriting around it would report a
	// replacement Word never makes, so any field vetoes the match however
	// few runs it spans.
	//
	// The scan starts before the first matched run because Word stores a
	// MERGEFIELD as an empty field-bearing run immediately followed by the
	// run holding its display text (see replaceParagraph in merge.go): the
	// field carries no text of its own, so an overlap test alone never sees
	// it, yet replacing the display text would leave the field live beside
	// the replacement.
	fieldLo := lo
	for fieldLo > 0 && spans[fieldLo-1].end == spans[fieldLo-1].start {
		fieldLo--
	}
	for _, s := range spans[fieldLo : hi+1] {
		if len(s.run.Fields()) > 0 {
			return false, nil
		}
	}

	// Breaks and images are the opposite case: a run carrying one still has
	// its text serialized normally, and SetText leaves the break or image
	// alone, so a match confined to a single run is safely replaceable.
	// Across several runs it is not — the runs between the ends are blanked,
	// and a break or image among them would be stranded in the middle of the
	// replacement. That includes zero-width runs interleaved between the
	// matched ones, the shape the reader produces for a standalone <w:br/>.
	if len(spanned) > 1 {
		for _, s := range spans[lo : hi+1] {
			if !isTextOnly(s.run) {
				return false, nil
			}
		}
	}

	first := spanned[0]
	last := spanned[len(spanned)-1]
	firstText := first.run.Text()
	lastText := last.run.Text()

	if len(spanned) == 1 {
		newText := firstText[:start-first.start] + replace + firstText[end-first.start:]
		return true, first.run.SetText(newText)
	}

	if err := first.run.SetText(firstText[:start-first.start] + replace); err != nil {
		return false, err
	}
	for _, s := range spanned[1 : len(spanned)-1] {
		if err := s.run.SetText(""); err != nil {
			return false, err
		}
	}
	return true, last.run.SetText(lastText[end-last.start:])
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package template

import (
	"regexp"

	"github.com/mmonterroca/docxgo/v2/domain"
)

// LocationType describes the structural location where a placeholder was found.
type LocationType int

const (
	// LocationParagraph is a top-level document paragraph.
	LocationParagraph LocationType = iota
	// LocationTableCell is a paragraph inside a table cell.
	LocationTableCell
	// LocationHeader is a paragraph inside a header.
	LocationHeader
	// LocationFooter is a paragraph inside a footer.
	LocationFooter
)

// Location provides structural context for a placeholder occurrence.
//
// When RunIndex == EndRunIndex (the common case — the whole match sits in
// one run), StartOffset and EndOffset are both offsets into that single run,
// exactly as in releases before v2.9.0: runs[RunIndex].Text()[StartOffset:EndOffset]
// yields the full match text.
//
// When RunIndex != EndRunIndex, Word split the placeholder across multiple
// runs (a common artifact of spell-check/proofing) and it was found by
// scanning across them without merging — see FindPlaceholders. StartOffset
// and EndOffset are then offsets into two *different* runs (RunIndex and
// EndRunIndex respectively); slicing a single run with both offsets is
// incorrect. A caller that needs the full match text should use FullMatch,
// or MergeTemplate to replace it (which does merge split runs, since writing
// a single replacement value requires it).
type Location struct {
	// Type indicates where the placeholder was found.
	Type LocationType
	// ParagraphIndex is the paragraph index within its container.
	ParagraphIndex int
	// RunIndex is the index of the run containing the start of the
	// placeholder, in the paragraph's own (unmodified) run list.
	RunIndex int
	// EndRunIndex is the index of the run containing the end of the
	// placeholder. Equal to RunIndex unless Word split the placeholder across
	// multiple runs, in which case it is the last run the match spans.
	EndRunIndex int
	// StartOffset is the byte offset of the placeholder start within the run
	// at RunIndex.
	StartOffset int
	// EndOffset is the byte offset after the placeholder end within the run
	// at EndRunIndex.
	EndOffset int
	// TableIndex, RowIndex, CellIndex are set when Type == LocationTableCell,
	// and also when the match is inside a table cell within a header or
	// footer (Type == LocationHeader/LocationFooter) — in that case they're
	// set alongside SectionIndex/HeaderType/FooterType below.
	TableIndex int
	RowIndex   int
	CellIndex  int
	// SectionIndex is set when Type == LocationHeader or LocationFooter.
	SectionIndex int
	// HeaderType is set when Type == LocationHeader.
	HeaderType domain.HeaderType
	// FooterType is set when Type == LocationFooter.
	FooterType domain.FooterType
}

// Placeholder represents one occurrence of a template placeholder.
type Placeholder struct {
	// Name is the placeholder key without delimiters (e.g. "FirstName").
	Name string
	// FullMatch is the complete matched text (e.g. "{{FirstName}}").
	FullMatch string
	// Location provides structural context for this occurrence.
	Location Location
}

// defaultPattern matches {{key}} where key is one or more word characters or dots.
var defaultPattern = regexp.MustCompile(`\{\{\.?(\w+(?:\.\w+)*)\}\}`)

// FindPlaceholders scans the entire document and returns all placeholder
// occurrences. It does not modify the document: placeholders that Word split
// across multiple runs are found by scanning across the run group they would
// merge into (see runGroups), without actually merging it. Use MergeTemplate
// to replace placeholders, which does merge split runs since that is
// necessary to write the replacement text.
func FindPlaceholders(doc domain.Document) []Placeholder {
	return findPlaceholdersWithPattern(doc, defaultPattern)
}

// FindPlaceholdersCustom scans the document using a custom regex pattern.
// The pattern must have exactly one capturing group for the placeholder name.
func FindPlaceholdersCustom(doc domain.Document, pattern *regexp.Regexp) []Placeholder {
	return findPlaceholdersWithPattern(doc, pattern)
}

func findPlaceholdersWithPattern(doc domain.Document, pattern *regexp.Regexp) []Placeholder {
	var results []Placeholder

	_ = walkParagraphs(doc, func(para domain.Paragraph, ctx paragraphContext) error {
		found := scanParagraph(para, pattern, ctx)
		results = append(results, found...)
		return nil
	})

	return results
}

// scanParagraph searches a paragraph for placeholders without mutating it.
// It groups runs the same way ConsolidateRuns would (see runGroups) so a
// placeholder split across several runs by Word is still found, then matches
// against each group's concatenated text and translates offsets back to the
// paragraph's actual, unmodified runs.
func scanParagraph(para domain.Paragraph, pattern *regexp.Regexp, ctx paragraphContext) []Placeholder {
	var results []Placeholder
	runs := para.Runs()

	for _, g := range runGroups(runs) {
		matches := pattern.FindAllStringSubmatchIndex(g.text, -1)
		if len(matches) == 0 {
			continue
		}

		runLens := make([]int, g.count)
		for i := 0; i < g.count; i++ {
			runLens[i] = len(runs[g.leader+i].Text())
		}

		for _, match := range matches {
			// Ensure there is at least one capturing group before indexing.
			if len(match) < 4 {
				// Pattern does not provide the expected capturing group; skip this match.
				continue
			}
			// match[0:2] is full match, match[2:4] is first capture group
			fullMatch := g.text[match[0]:match[1]]
			name := g.text[match[2]:match[3]]

			startRun, startOffset := locateGroupOffset(g.leader, runLens, match[0], false)
			endRun, endOffset := locateGroupOffset(g.leader, runLens, match[1], true)

			loc := Location{
				Type:           ctx.locationType,
				ParagraphIndex: ctx.paraIdx,
				RunIndex:       startRun,
				EndRunIndex:    endRun,
				StartOffset:    startOffset,
				EndOffset:      endOffset,
				TableIndex:     ctx.tableIdx,
				RowIndex:       ctx.rowIdx,
				CellIndex:      ctx.cellIdx,
				SectionIndex:   ctx.sectionIdx,
				HeaderType:     ctx.headerType,
				FooterType:     ctx.footerType,
			}

			results = append(results, Placeholder{
				Name:      name,
				FullMatch: fullMatch,
				Location:  loc,
			})
		}
	}

	return results
}

// locateGroupOffset translates a byte offset into a run group's concatenated
// text back to the index (relative to the paragraph's own runs) and local
// offset of the run that contains it.
//
// isEnd distinguishes how a boundary offset — one landing exactly at the
// junction between two runs — resolves. A match's exclusive end offset
// belongs to the run it closes out (isEnd=true, "<="): the end of a match
// that consumes a run's last character lands at that run's own length, not
// offset 0 of a run that may not even exist. A match's start offset belongs
// to the run its first character actually starts in (isEnd=false, strict
// "<"): using "<=" here would misattribute a start landing exactly at a run
// boundary to the end of the earlier run instead of the start of the next
// one, which also skips past empty runs (length 0) to the first run that
// actually has content — the correct behavior for a leading empty run.
// An offset at or beyond the end of the group's text resolves to the end of
// its last run.
func locateGroupOffset(leader int, runLens []int, offset int, isEnd bool) (runIndex, runOffset int) {
	for i, l := range runLens {
		if offset < l || (isEnd && offset == l) {
			return leader + i, offset
		}
		offset -= l
	}
	last := len(runLens) - 1
	return leader + last, runLens[last]
}

// PlaceholderNames returns the deduplicated set of placeholder names found in the document.
func PlaceholderNames(doc domain.Document) []string {
	placeholders := FindPlaceholders(doc)
	seen := make(map[string]struct{})
	var names []string

	for _, p := range placeholders {
		if _, ok := seen[p.Name]; !ok {
			seen[p.Name] = struct{}{}
			names = append(names, p.Name)
		}
	}

	return names
}

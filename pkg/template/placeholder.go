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
type Location struct {
	// Type indicates where the placeholder was found.
	Type LocationType
	// ParagraphIndex is the paragraph index within its container.
	ParagraphIndex int
	// RunIndex is the run index within the paragraph.
	RunIndex int
	// StartOffset is the byte offset of the placeholder start within the run text.
	StartOffset int
	// EndOffset is the byte offset after the placeholder end within the run text.
	EndOffset int
	// TableIndex, RowIndex, CellIndex are set when Type == LocationTableCell.
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

// FindPlaceholders scans the entire document and returns all placeholder occurrences.
// It consolidates runs in each paragraph before scanning to heal split placeholders.
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
		// Consolidate runs to heal split placeholders
		ConsolidateRuns(para)

		found := scanParagraph(para, pattern, ctx)
		results = append(results, found...)
		return nil
	})

	return results
}

// scanParagraph searches all runs in a paragraph for placeholders.
func scanParagraph(para domain.Paragraph, pattern *regexp.Regexp, ctx paragraphContext) []Placeholder {
	var results []Placeholder
	runs := para.Runs()

	for ri, run := range runs {
		text := run.Text()
		matches := pattern.FindAllStringSubmatchIndex(text, -1)

		for _, match := range matches {
			// match[0:2] is full match, match[2:4] is first capture group
			fullMatch := text[match[0]:match[1]]
			name := text[match[2]:match[3]]

			loc := Location{
				Type:           ctx.locationType,
				ParagraphIndex: ctx.paraIdx,
				RunIndex:       ri,
				StartOffset:    match[0],
				EndOffset:      match[1],
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

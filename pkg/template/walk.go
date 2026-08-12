// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package template

import "github.com/mmonterroca/docxgo/v2/domain"

// paragraphContext provides metadata about where a paragraph was found.
type paragraphContext struct {
	locationType LocationType
	paraIdx      int
	// Table cell context. inTableCell is the discriminator: tableIdx/rowIdx/
	// cellIdx are only meaningful when it's true. LocationType alone can't
	// serve that role for a header/footer paragraph -- unlike the body,
	// where LocationTableCell vs LocationParagraph already says it -- because
	// walkHeaderFooterTables deliberately reuses LocationHeader/LocationFooter
	// for header/footer table cells too (see its doc comment), so a plain
	// header paragraph and a header table's cell (0, 0, 0) are otherwise
	// indistinguishable by Type and zero-valued indices alone.
	inTableCell bool
	tableIdx    int
	rowIdx      int
	cellIdx     int
	// Header/footer context
	sectionIdx int
	headerType domain.HeaderType
	footerType domain.FooterType
}

// walkParagraphs calls fn for every paragraph in the document,
// including those in table cells, headers, and footers.
func walkParagraphs(doc domain.Document, fn func(para domain.Paragraph, ctx paragraphContext) error) error {
	// 1. Walk document-level paragraphs
	for i, para := range doc.Paragraphs() {
		ctx := paragraphContext{
			locationType: LocationParagraph,
			paraIdx:      i,
		}
		if err := fn(para, ctx); err != nil {
			return err
		}
	}

	// 2. Walk table cells
	for ti, table := range doc.Tables() {
		for ri, row := range table.Rows() {
			for ci, cell := range row.Cells() {
				for pi, para := range cell.Paragraphs() {
					ctx := paragraphContext{
						locationType: LocationTableCell,
						paraIdx:      pi,
						inTableCell:  true,
						tableIdx:     ti,
						rowIdx:       ri,
						cellIdx:      ci,
					}
					if err := fn(para, ctx); err != nil {
						return err
					}
				}
			}
		}
	}

	// 3. Walk headers and footers
	// Use HeadersAll/FootersAll to read only existing headers/footers
	// without creating new ones (section.Header() auto-creates if missing).
	for si, section := range doc.Sections() {
		type sectionWithMaps interface {
			HeadersAll() map[domain.HeaderType]domain.Header
			FootersAll() map[domain.FooterType]domain.Footer
		}
		secMaps, ok := section.(sectionWithMaps)
		if !ok {
			continue
		}

		for ht, header := range secMaps.HeadersAll() {
			if header == nil {
				continue
			}
			for pi, para := range header.Paragraphs() {
				ctx := paragraphContext{
					locationType: LocationHeader,
					paraIdx:      pi,
					sectionIdx:   si,
					headerType:   ht,
				}
				if err := fn(para, ctx); err != nil {
					return err
				}
			}
			if err := walkHeaderFooterTables(header.Tables(), LocationHeader, si, ht, 0, fn); err != nil {
				return err
			}
		}

		for ft, footer := range secMaps.FootersAll() {
			if footer == nil {
				continue
			}
			for pi, para := range footer.Paragraphs() {
				ctx := paragraphContext{
					locationType: LocationFooter,
					paraIdx:      pi,
					sectionIdx:   si,
					footerType:   ft,
				}
				if err := fn(para, ctx); err != nil {
					return err
				}
			}
			if err := walkHeaderFooterTables(footer.Tables(), LocationFooter, si, 0, ft, fn); err != nil {
				return err
			}
		}
	}

	return nil
}

// walkHeaderFooterTables walks the cells of every top-level table in a
// header or footer, calling fn for each paragraph found. loc is
// LocationHeader or LocationFooter (never a distinct "table cell" type):
// that keeps the existing skipHeaderFooter check in ReplaceText/MergeTemplate
// — which only inspects ctx.locationType — correctly skipping these
// paragraphs on a document whose headers/footers were preserved verbatim.
// TableIndex/RowIndex/CellIndex are populated alongside SectionIndex/
// HeaderType/FooterType so a caller can tell exactly which cell a match
// came from.
func walkHeaderFooterTables(tables []domain.Table, loc LocationType, sectionIdx int, ht domain.HeaderType, ft domain.FooterType, fn func(para domain.Paragraph, ctx paragraphContext) error) error {
	for ti, table := range tables {
		for ri, row := range table.Rows() {
			for ci, cell := range row.Cells() {
				for pi, para := range cell.Paragraphs() {
					ctx := paragraphContext{
						locationType: loc,
						paraIdx:      pi,
						inTableCell:  true,
						tableIdx:     ti,
						rowIdx:       ri,
						cellIdx:      ci,
						sectionIdx:   sectionIdx,
						headerType:   ht,
						footerType:   ft,
					}
					if err := fn(para, ctx); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

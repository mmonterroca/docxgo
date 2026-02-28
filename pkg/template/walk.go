package template

import "github.com/mmonterroca/docxgo/v2/domain"

// paragraphContext provides metadata about where a paragraph was found.
type paragraphContext struct {
	locationType LocationType
	paraIdx      int
	// Table cell context
	tableIdx int
	rowIdx   int
	cellIdx  int
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
		}
	}

	return nil
}

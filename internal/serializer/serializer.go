package serializer

/*
   Copyright (c) 2025 SlideLang

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

import (
	"fmt"

	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/internal/xml"
	"github.com/mmonterroca/docxgo/pkg/color"
	"github.com/mmonterroca/docxgo/pkg/constants"
)

// RunSerializer converts a domain.Run to xml.Run
type RunSerializer struct{}

// NewRunSerializer creates a new RunSerializer.
func NewRunSerializer() *RunSerializer {
	return &RunSerializer{}
}

// Serialize converts a domain.Run to xml.Run.
func (s *RunSerializer) Serialize(run domain.Run) *xml.Run {
	xmlRun := &xml.Run{
		Properties: s.serializeProperties(run),
		Text:       s.serializeText(run),
	}
	return xmlRun
}

func (s *RunSerializer) serializeProperties(run domain.Run) *xml.RunProperties {
	props := &xml.RunProperties{}

	// Bold
	if run.Bold() {
		props.Bold = &xml.BoolValue{Val: boolPtr(true)}
	}

	// Italic
	if run.Italic() {
		props.Italic = &xml.BoolValue{Val: boolPtr(true)}
	}

	// Strike
	if run.Strike() {
		props.Strike = &xml.BoolValue{Val: boolPtr(true)}
	}

	// Underline
	if run.Underline() != domain.UnderlineNone {
		props.Underline = &xml.Underline{
			Val: s.underlineStyleToString(run.Underline()),
		}
	}

	// Color
	if run.Color() != domain.ColorBlack {
		props.Color = &xml.Color{
			Val: color.ToHex(run.Color()),
		}
	}

	// Font size
	if run.Size() != constants.DefaultFontSize {
		props.Size = &xml.HalfPt{Val: run.Size()}
		props.SizeCS = &xml.HalfPt{Val: run.Size()}
	}

	// Font
	font := run.Font()
	if font.Name != "" && font.Name != constants.DefaultFontName {
		props.Font = &xml.Font{
			ASCII:    font.Name,
			EastAsia: font.EastAsia,
			CS:       font.CS,
		}
	}

	// Highlight
	if run.Highlight() != domain.HighlightNone {
		props.Highlight = &xml.Highlight{
			Val: s.highlightColorToString(run.Highlight()),
		}
	}

	return props
}

func (s *RunSerializer) serializeText(run domain.Run) *xml.Text {
	text := run.Text()
	if text == "" {
		return nil
	}

	xmlText := &xml.Text{
		Content: text,
	}

	// Preserve spaces if text starts/ends with space
	if len(text) > 0 && (text[0] == ' ' || text[len(text)-1] == ' ') {
		xmlText.Space = "preserve"
	}

	return xmlText
}

func (s *RunSerializer) underlineStyleToString(style domain.UnderlineStyle) string {
	switch style {
	case domain.UnderlineNone:
		return constants.UnderlineValueNone
	case domain.UnderlineSingle:
		return constants.UnderlineValueSingle
	case domain.UnderlineDouble:
		return constants.UnderlineValueDouble
	case domain.UnderlineThick:
		return constants.UnderlineValueThick
	case domain.UnderlineDotted:
		return constants.UnderlineValueDotted
	case domain.UnderlineDashed:
		return constants.UnderlineValueDashed
	case domain.UnderlineWave:
		return constants.UnderlineValueWave
	default:
		return constants.UnderlineValueSingle
	}
}

func (s *RunSerializer) highlightColorToString(hlColor domain.HighlightColor) string {
	switch hlColor {
	case domain.HighlightNone:
		return constants.HighlightValueNone
	case domain.HighlightYellow:
		return constants.HighlightValueYellow
	case domain.HighlightGreen:
		return constants.HighlightValueGreen
	case domain.HighlightCyan:
		return constants.HighlightValueCyan
	case domain.HighlightMagenta:
		return constants.HighlightValueMagenta
	case domain.HighlightBlue:
		return constants.HighlightValueBlue
	case domain.HighlightRed:
		return constants.HighlightValueRed
	case domain.HighlightDarkBlue:
		return constants.HighlightValueDarkBlue
	case domain.HighlightDarkCyan:
		return constants.HighlightValueDarkCyan
	case domain.HighlightDarkGreen:
		return constants.HighlightValueDarkGreen
	case domain.HighlightDarkMagenta:
		return constants.HighlightValueDarkMagenta
	case domain.HighlightDarkRed:
		return constants.HighlightValueDarkRed
	case domain.HighlightDarkYellow:
		return constants.HighlightValueDarkYellow
	case domain.HighlightDarkGray:
		return constants.HighlightValueDarkGray
	case domain.HighlightLightGray:
		return constants.HighlightValueLightGray
	default:
		return constants.HighlightValueNone
	}
}

// ParagraphSerializer converts a domain.Paragraph to xml.Paragraph
type ParagraphSerializer struct {
	runSerializer *RunSerializer
}

// NewParagraphSerializer creates a new ParagraphSerializer.
func NewParagraphSerializer() *ParagraphSerializer {
	return &ParagraphSerializer{
		runSerializer: NewRunSerializer(),
	}
}

// Serialize converts a domain.Paragraph to xml.Paragraph.
func (s *ParagraphSerializer) Serialize(para domain.Paragraph) *xml.Paragraph {
	xmlPara := &xml.Paragraph{
		Properties: s.serializeProperties(para),
		Runs:       make([]*xml.Run, 0, len(para.Runs())),
	}

	// Serialize runs
	for _, run := range para.Runs() {
		xmlPara.Runs = append(xmlPara.Runs, s.runSerializer.Serialize(run))
	}

	return xmlPara
}

func (s *ParagraphSerializer) serializeProperties(para domain.Paragraph) *xml.ParagraphProperties {
	props := &xml.ParagraphProperties{}

	// Alignment
	if para.Alignment() != domain.AlignmentLeft {
		props.Justification = &xml.Justification{
			Val: s.alignmentToString(para.Alignment()),
		}
	}

	// Indentation
	indent := para.Indent()
	if indent.Left != 0 || indent.Right != 0 || indent.FirstLine != 0 || indent.Hanging != 0 {
		props.Indentation = &xml.Indentation{
			Left:      intPtrIfNotZero(indent.Left),
			Right:     intPtrIfNotZero(indent.Right),
			FirstLine: intPtrIfNotZero(indent.FirstLine),
			Hanging:   intPtrIfNotZero(indent.Hanging),
		}
	}

	// Spacing
	before := para.SpacingBefore()
	after := para.SpacingAfter()
	lineSpacing := para.LineSpacing()

	if before != 0 || after != 0 || lineSpacing.Value != constants.DefaultLineSpacing {
		props.Spacing = &xml.Spacing{
			Before:   intPtrIfNotZero(before),
			After:    intPtrIfNotZero(after),
			Line:     intPtrIfNotZero(lineSpacing.Value),
			LineRule: s.lineSpacingRuleToString(lineSpacing.Rule),
		}
	}

	return props
}

func (s *ParagraphSerializer) alignmentToString(align domain.Alignment) string {
	switch align {
	case domain.AlignmentLeft:
		return constants.AlignmentValueLeft
	case domain.AlignmentCenter:
		return constants.AlignmentValueCenter
	case domain.AlignmentRight:
		return constants.AlignmentValueRight
	case domain.AlignmentJustify:
		return constants.AlignmentValueJustify
	case domain.AlignmentDistribute:
		return constants.AlignmentValueDistribute
	default:
		return constants.AlignmentValueLeft
	}
}

func (s *ParagraphSerializer) lineSpacingRuleToString(rule domain.LineSpacingRule) *string {
	var val string
	switch rule {
	case domain.LineSpacingAuto:
		val = constants.LineSpacingRuleAuto
	case domain.LineSpacingExact:
		val = constants.LineSpacingRuleExact
	case domain.LineSpacingAtLeast:
		val = constants.LineSpacingRuleAtLeast
	default:
		val = constants.LineSpacingRuleAuto
	}
	return &val
}

// TableSerializer converts domain tables to XML
type TableSerializer struct {
	paraSerializer *ParagraphSerializer
}

// NewTableSerializer creates a new TableSerializer.
func NewTableSerializer() *TableSerializer {
	return &TableSerializer{
		paraSerializer: NewParagraphSerializer(),
	}
}

// Serialize converts a domain.Table to xml.Table.
func (s *TableSerializer) Serialize(table domain.Table) *xml.Table {
	xmlTable := &xml.Table{
		Properties: s.serializeTableProperties(table),
		Grid:       s.serializeGrid(table),
		Rows:       make([]*xml.TableRow, 0, table.RowCount()),
	}

	// Serialize rows
	for i := 0; i < table.RowCount(); i++ {
		row, _ := table.Row(i)
		xmlTable.Rows = append(xmlTable.Rows, s.serializeRow(row))
	}

	return xmlTable
}

func (s *TableSerializer) serializeTableProperties(table domain.Table) *xml.TableProperties {
	props := &xml.TableProperties{}

	// Width
	width := table.Width()
	props.Width = &xml.TableWidth{
		Type: s.widthTypeToString(width.Type),
		W:    width.Value,
	}

	// Alignment
	if table.Alignment() != domain.AlignmentLeft {
		props.Jc = &xml.Justification{
			Val: s.alignmentToString(table.Alignment()),
		}
	}

	return props
}

func (s *TableSerializer) serializeGrid(table domain.Table) *xml.TableGrid {
	grid := &xml.TableGrid{
		Cols: make([]*xml.GridCol, table.ColumnCount()),
	}

	for i := 0; i < table.ColumnCount(); i++ {
		grid.Cols[i] = &xml.GridCol{}
	}

	return grid
}

func (s *TableSerializer) serializeRow(row domain.TableRow) *xml.TableRow {
	xmlRow := &xml.TableRow{
		Cells: make([]*xml.TableCell, 0, len(row.Cells())),
	}

	// Height
	if row.Height() > 0 {
		xmlRow.Properties = &xml.TableRowProperties{
			Height: &xml.TableRowHeight{
				Val:  row.Height(),
				Rule: "atLeast",
			},
		}
	}

	// Serialize cells
	for _, cell := range row.Cells() {
		xmlRow.Cells = append(xmlRow.Cells, s.serializeCell(cell))
	}

	return xmlRow
}

func (s *TableSerializer) serializeCell(cell domain.TableCell) *xml.TableCell {
	xmlCell := &xml.TableCell{
		Properties: s.serializeCellProperties(cell),
		Paragraphs: make([]*xml.Paragraph, 0, len(cell.Paragraphs())),
		Tables:     make([]*xml.Table, 0, len(cell.Tables())),
	}

	// Serialize paragraphs
	for _, para := range cell.Paragraphs() {
		xmlCell.Paragraphs = append(xmlCell.Paragraphs, s.paraSerializer.Serialize(para))
	}

	// Serialize nested tables
	for _, table := range cell.Tables() {
		xmlCell.Tables = append(xmlCell.Tables, s.Serialize(table))
	}

	// Add empty paragraph if cell has no content
	if len(xmlCell.Paragraphs) == 0 && len(xmlCell.Tables) == 0 {
		xmlCell.Paragraphs = append(xmlCell.Paragraphs, &xml.Paragraph{})
	}

	return xmlCell
}

func (s *TableSerializer) serializeCellProperties(cell domain.TableCell) *xml.TableCellProperties {
	props := &xml.TableCellProperties{}

	// Width
	if cell.Width() > 0 {
		props.Width = &xml.TableWidth{
			Type: constants.WidthTypeDXA,
			W:    cell.Width(),
		}
	}

	// GridSpan (horizontal merge)
	if cell.GridSpan() > 1 {
		props.GridSpan = &xml.GridSpan{
			Val: cell.GridSpan(),
		}
	}

	// VMerge (vertical merge)
	if cell.VMerge() != domain.VMergeNone {
		vMerge := &xml.VMerge{}
		if cell.VMerge() == domain.VMergeRestart {
			vMerge.Val = "restart"
		}
		// VMergeContinue uses empty Val (omitted)
		props.VMerge = vMerge
	}

	// Vertical alignment
	if cell.VerticalAlignment() != domain.VerticalAlignTop {
		props.VAlign = &xml.VerticalAlign{
			Val: s.verticalAlignToString(cell.VerticalAlignment()),
		}
	}

	// Shading
	if cell.Shading() != domain.ColorWhite {
		props.Shading = &xml.Shading{
			Val:  "clear",
			Fill: color.ToHex(cell.Shading()),
		}
	}

	return props
}

func (s *TableSerializer) widthTypeToString(wType domain.WidthType) string {
	switch wType {
	case domain.WidthAuto:
		return constants.WidthTypeAuto
	case domain.WidthDXA:
		return constants.WidthTypeDXA
	case domain.WidthPct:
		return constants.WidthTypePct
	default:
		return constants.WidthTypeAuto
	}
}

func (s *TableSerializer) alignmentToString(align domain.Alignment) string {
	switch align {
	case domain.AlignmentLeft:
		return constants.AlignmentValueLeft
	case domain.AlignmentCenter:
		return constants.AlignmentValueCenter
	case domain.AlignmentRight:
		return constants.AlignmentValueRight
	case domain.AlignmentJustify:
		return constants.AlignmentValueJustify
	case domain.AlignmentDistribute:
		return constants.AlignmentValueDistribute
	default:
		return constants.AlignmentValueLeft
	}
}

func (s *TableSerializer) verticalAlignToString(align domain.VerticalAlignment) string {
	switch align {
	case domain.VerticalAlignTop:
		return constants.VerticalAlignmentValueTop
	case domain.VerticalAlignCenter:
		return constants.VerticalAlignmentValueCenter
	case domain.VerticalAlignBottom:
		return constants.VerticalAlignmentValueBottom
	default:
		return constants.VerticalAlignmentValueTop
	}
}

// Helper functions

func boolPtr(b bool) *bool {
	return &b
}

func intPtrIfNotZero(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

// DocumentSerializer converts a domain.Document to XML structures
type DocumentSerializer struct {
	paraSerializer  *ParagraphSerializer
	tableSerializer *TableSerializer
}

// NewDocumentSerializer creates a new DocumentSerializer.
func NewDocumentSerializer() *DocumentSerializer {
	return &DocumentSerializer{
		paraSerializer:  NewParagraphSerializer(),
		tableSerializer: NewTableSerializer(),
	}
}

// SerializeBody converts document content to xml.Body.
func (s *DocumentSerializer) SerializeBody(doc domain.Document) *xml.Body {
	body := &xml.Body{
		Paragraphs: make([]*xml.Paragraph, 0, len(doc.Paragraphs())),
		Tables:     make([]*xml.Table, 0, len(doc.Tables())),
	}

	// For now, serialize all paragraphs then all tables
	// TODO: Maintain insertion order
	for _, para := range doc.Paragraphs() {
		body.Paragraphs = append(body.Paragraphs, s.paraSerializer.Serialize(para))
	}

	for _, table := range doc.Tables() {
		body.Tables = append(body.Tables, s.tableSerializer.Serialize(table))
	}

	return body
}

// SerializeDocument creates the complete document XML structure.
func (s *DocumentSerializer) SerializeDocument(doc domain.Document) *xml.Document {
	return &xml.Document{
		XMLnsW: constants.NamespaceMain,
		XMLnsR: constants.NamespaceRelationships,
		Body:   s.SerializeBody(doc),
	}
}

// SerializeCoreProperties converts metadata to core properties.
func (s *DocumentSerializer) SerializeCoreProperties(meta *domain.Metadata) *xml.CoreProperties {
	props := &xml.CoreProperties{
		XMLnsCP:      constants.NamespaceCoreProperties,
		XMLnsDC:      constants.NamespaceDC,
		XMLnsDCTerms: constants.NamespaceDCTerms,
		XMLnsXSI:     "http://www.w3.org/2001/XMLSchema-instance",
		Title:        meta.Title,
		Subject:      meta.Subject,
		Creator:      meta.Creator,
		Description:  meta.Description,
	}

	// Keywords
	if len(meta.Keywords) > 0 {
		keywords := ""
		for i, kw := range meta.Keywords {
			if i > 0 {
				keywords += ", "
			}
			keywords += kw
		}
		props.Keywords = keywords
	}

	// Dates
	if meta.Created != "" {
		props.Created = &xml.DCDate{
			Type:  "dcterms:W3CDTF",
			Value: meta.Created,
		}
	}
	if meta.Modified != "" {
		props.Modified = &xml.DCDate{
			Type:  "dcterms:W3CDTF",
			Value: meta.Modified,
		}
	}

	return props
}

// SerializeAppProperties creates app.xml properties.
func (s *DocumentSerializer) SerializeAppProperties(doc domain.Document) *xml.AppProperties {
	return &xml.AppProperties{
		Xmlns:       constants.NamespaceExtendedProperties,
		Application: "go-docx/v2",
		DocSecurity: 0,
		Lines:       0,
		Paragraphs:  len(doc.Paragraphs()),
		Company:     "SlideLang",
	}
}

// Debug method for testing
func (s *DocumentSerializer) DebugPrint(doc domain.Document) {
	fmt.Printf("Document has %d paragraphs and %d tables\n",
		len(doc.Paragraphs()), len(doc.Tables()))
}

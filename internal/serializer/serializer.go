// Package serializer converts domain objects into XML structures for OOXML serialization.
// It provides serializers for documents, paragraphs, runs, tables, and other document elements.
package serializer

/*
   Copyright (c) 2025 Misael Monterroca

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
	"strings"

	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/internal/xml"
	"github.com/mmonterroca/docxgo/pkg/color"
	"github.com/mmonterroca/docxgo/pkg/constants"
)

type drawingIDProvider interface {
	NextDrawingID() int
}

// RunSerializer converts a domain.Run to xml.Run
type RunSerializer struct {
	idProvider drawingIDProvider
}

// NewRunSerializer creates a new RunSerializer.
func NewRunSerializer() *RunSerializer {
	return &RunSerializer{}
}

// SetDrawingIDProvider injects a provider for generating unique drawing IDs.
func (s *RunSerializer) SetDrawingIDProvider(provider drawingIDProvider) {
	s.idProvider = provider
}

// Serialize converts a domain.Run to xml.Run.
func (s *RunSerializer) Serialize(run domain.Run) *xml.Run {
	xmlRun := &xml.Run{
		Properties: s.serializeProperties(run),
		Text:       s.serializeText(run),
	}

	if imageProvider, ok := run.(interface{ Image() domain.Image }); ok {
		if img := imageProvider.Image(); img != nil {
			drawingID := 1
			if s.idProvider != nil {
				drawingID = s.idProvider.NextDrawingID()
			}
			xmlRun.Drawing = s.serializeDrawing(img, drawingID)
			// For image runs we don't include text content.
			xmlRun.Text = nil
		}
	}

	// Add breaks if any
	if breaks := run.(interface{ Breaks() []domain.BreakType }).Breaks(); breaks != nil {
		for _, br := range breaks {
			xmlRun.Break = s.serializeBreak(br)
		}
	}

	return xmlRun
}

func (s *RunSerializer) serializeDrawing(img domain.Image, drawingID int) *xml.Drawing {
	if img == nil {
		return nil
	}

	pos := img.Position()
	if pos.Type == domain.ImagePositionFloating {
		return xml.NewFloatingDrawing(img, drawingID)
	}
	return xml.NewInlineDrawing(img, drawingID)
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
			HAnsi:    font.Name,
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
	return s.serializeTextContent(run.Text())
}

func (s *RunSerializer) serializeTextContent(text string) *xml.Text {
	if text == "" {
		return nil
	}

	xmlText := &xml.Text{
		Content: text,
	}

	if len(text) > 0 && (text[0] == ' ' || text[len(text)-1] == ' ') {
		xmlText.Space = "preserve"
	}

	return xmlText
}

func (s *RunSerializer) serializeBreak(breakType domain.BreakType) *xml.Break {
	xmlBreak := &xml.Break{}

	switch breakType {
	case domain.BreakTypePage:
		xmlBreak.Type = "page"
	case domain.BreakTypeColumn:
		xmlBreak.Type = "column"
	case domain.BreakTypeLine:
		xmlBreak.Type = "textWrapping"
	default:
		xmlBreak.Type = "textWrapping"
	}

	return xmlBreak
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
		Elements:   make([]interface{}, 0, len(para.Runs())+2),
	}

	// Add bookmark if this paragraph has one (needed for TOC)
	if corePara, ok := para.(interface {
		BookmarkID() string
		BookmarkName() string
	}); ok {
		if bookmarkID := corePara.BookmarkID(); bookmarkID != "" {
			xmlPara.Elements = append(xmlPara.Elements, &xml.BookmarkStart{
				ID:   bookmarkID,
				Name: corePara.BookmarkName(),
			})
		}
	}

	// Serialize runs - expand runs with fields into multiple XML runs
	for _, run := range para.Runs() {
		// Check if run has fields
		if runWithFields, ok := run.(interface{ Fields() []domain.Field }); ok {
			fields := runWithFields.Fields()
			if len(fields) > 0 {
				// Expand run with fields into multiple XML runs
				xmlPara.Elements = append(xmlPara.Elements, s.expandRunWithFields(run, fields)...)
				continue
			}
		}

		if text := run.Text(); strings.Contains(text, "\n") {
			xmlPara.Elements = append(xmlPara.Elements, s.expandRunWithNewlines(run, text)...)
			continue
		}

		// Regular run without fields
		xmlPara.Elements = append(xmlPara.Elements, s.runSerializer.Serialize(run))
	}

	// Add bookmark end if this paragraph has a bookmark
	if corePara, ok := para.(interface{ BookmarkID() string }); ok {
		if bookmarkID := corePara.BookmarkID(); bookmarkID != "" {
			xmlPara.Elements = append(xmlPara.Elements, &xml.BookmarkEnd{ID: bookmarkID})
		}
	}

	return xmlPara
}

func (s *ParagraphSerializer) expandRunWithNewlines(run domain.Run, text string) []interface{} {
	parts := strings.Split(text, "\n")
	if len(parts) == 0 {
		return []interface{}{s.runSerializer.Serialize(run)}
	}

	result := make([]interface{}, 0, len(parts)*2-1)

	var (
		setter   func(string) error
		restore  func()
		canSet   bool
		original string
	)

	if s, ok := run.(interface{ SetText(string) error }); ok {
		canSet = true
		original = run.Text()
		setter = s.SetText
		restore = func() {
			_ = setter(original)
		}
	}

	for idx, part := range parts {
		var xmlRun *xml.Run

		if canSet {
			_ = setter(part)
			xmlRun = s.runSerializer.Serialize(run)
		} else {
			xmlRun = &xml.Run{
				Properties: s.runSerializer.serializeProperties(run),
				Text:       s.runSerializer.serializeTextContent(part),
			}
		}

		if idx < len(parts)-1 {
			xmlRun.Break = &xml.Break{}
		}

		result = append(result, xmlRun)
	}

	if canSet && restore != nil {
		restore()
	}

	return result
}

// expandRunWithFields expands a run containing fields into XML elements while preserving formatting.
// The returned slice may include runs, hyperlinks, and field components.
func (s *ParagraphSerializer) expandRunWithFields(run domain.Run, fields []domain.Field) []interface{} {
	elements := make([]interface{}, 0, len(fields)*5)

	for _, field := range fields {
		wasDirty := false
		if dirtyChecker, ok := field.(interface{ IsDirty() bool }); ok {
			wasDirty = dirtyChecker.IsDirty()
		}
		if updater, ok := field.(interface{ Update() error }); ok {
			_ = updater.Update()
			if wasDirty {
				if marker, ok := field.(interface{ MarkDirty() }); ok {
					marker.MarkDirty()
				}
			}
		}

		switch field.Type() { //nolint:exhaustive // Only Hyperlink needs special handling; others use standard field serialization below
		case domain.FieldTypeHyperlink:
			if accessor, ok := field.(interface {
				GetProperty(string) (string, bool)
			}); ok {
				relID, relOK := accessor.GetProperty("relationshipID")
				if relOK && relID != "" {
					display := field.Result()
					if display == "" {
						if disp, ok := accessor.GetProperty("display"); ok {
							display = disp
						}
					}

					var xmlRun *xml.Run
					if setter, ok := run.(interface{ SetText(string) error }); ok {
						origText := run.Text()
						_ = setter.SetText(display)
						xmlRun = s.runSerializer.Serialize(run)
						_ = setter.SetText(origText)
					} else {
						xmlRun = &xml.Run{
							Properties: s.runSerializer.serializeProperties(run),
							Text:       &xml.Text{Content: display},
						}
					}

					if xmlRun.Text == nil {
						xmlRun.Text = &xml.Text{Content: display}
					} else {
						xmlRun.Text.Content = display
					}

					xmlRun.FieldChar = nil
					xmlRun.InstrText = nil

					if xmlRun.Properties == nil {
						xmlRun.Properties = &xml.RunProperties{}
					}
					xmlRun.Properties.Style = &xml.RunStyle{Val: "Hyperlink"}

					hyperlink := &xml.Hyperlink{
						ID:   relID,
						Runs: []*xml.Run{xmlRun},
					}
					elements = append(elements, hyperlink)
					continue
				}
			}
		default:
			// Other field types (TOC, PageNumber, Date, etc.) use standard field serialization
			// which is handled below
		}

		beginRun := &xml.Run{FieldChar: xml.NewFieldBegin()}
		if dirtyField, ok := field.(interface{ IsDirty() bool }); ok {
			if dirtyField.IsDirty() {
				dirty := true
				beginRun.FieldChar.Dirty = &dirty
			}
		}
		elements = append(elements, beginRun)

		instrRun := &xml.Run{InstrText: xml.NewInstrText(field.Code())}
		elements = append(elements, instrRun)

		sepRun := &xml.Run{FieldChar: xml.NewFieldSeparate()}
		elements = append(elements, sepRun)

		resultText := field.Result()
		if resultText != "" {
			resultRun := &xml.Run{
				Properties: s.runSerializer.serializeProperties(run),
				Text:       &xml.Text{Content: resultText},
			}
			elements = append(elements, resultRun)
		}

		endRun := &xml.Run{FieldChar: xml.NewFieldEnd()}
		elements = append(elements, endRun)
	}

	if run.Text() != "" {
		elements = append(elements, s.runSerializer.Serialize(run))
	}

	return elements
}

func (s *ParagraphSerializer) serializeProperties(para domain.Paragraph) *xml.ParagraphProperties {
	props := &xml.ParagraphProperties{}

	// Style - access the internal styleName field via type assertion
	if corePara, ok := para.(interface{ StyleName() string }); ok {
		if styleName := corePara.StyleName(); styleName != "" {
			props.Style = &xml.ParagraphStyleRef{
				Val: styleName,
			}
		}
	}

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

	if ref, ok := para.Numbering(); ok {
		props.Numbering = &xml.NumberingProperties{
			Level: &xml.DecimalNumber{Val: ref.Level},
			NumID: &xml.DecimalNumber{Val: ref.ID},
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

	// Default look hints are expressed purely via w:val for strict OOXML compliance.
	props.Look = &xml.TableLook{Val: "04A0"}

	// Alignment
	if table.Alignment() != domain.AlignmentLeft {
		props.Jc = &xml.Justification{
			Val: s.alignmentToString(table.Alignment()),
		}
	}

	// Style
	if style := table.Style(); style.Name != "" {
		props.Style = &xml.TableStyle{
			Val: style.Name,
		}
	}

	return props
}

func (s *TableSerializer) serializeGrid(table domain.Table) *xml.TableGrid {
	grid := &xml.TableGrid{
		Cols: make([]*xml.GridCol, table.ColumnCount()),
	}

	for i := 0; i < table.ColumnCount(); i++ {
		grid.Cols[i] = &xml.GridCol{W: intPtr(0)}
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

	// Serialize cells, skipping horizontal merge continuations
	for _, cell := range row.Cells() {
		if cell.IsHorizontallyMergedContinuation() {
			continue
		}
		xmlRow.Cells = append(xmlRow.Cells, s.serializeCell(cell))
	}

	return xmlRow
}

func (s *TableSerializer) serializeCell(cell domain.TableCell) *xml.TableCell {
	paragraphs := cell.Paragraphs()
	tables := cell.Tables()

	content := make([]interface{}, 0, len(paragraphs)+len(tables)+1)

	for _, para := range paragraphs {
		content = append(content, s.paraSerializer.Serialize(para))
	}

	if len(tables) > 0 {
		// If the cell contains only nested tables, add a leading placeholder paragraph to anchor the table content.
		if len(paragraphs) == 0 {
			content = append(content, emptyParagraph())
		}

		for _, table := range tables {
			content = append(content, s.Serialize(table))
		}

		// Word expects a trailing empty paragraph after nested tables to keep the end-of-cell marker intact.
		content = append(content, emptyParagraph())
	}

	if len(content) == 0 {
		content = append(content, emptyParagraph())
	}

	return &xml.TableCell{
		Properties: s.serializeCellProperties(cell),
		Content:    content,
	}
}

func (s *TableSerializer) serializeCellProperties(cell domain.TableCell) *xml.TableCellProperties {
	props := &xml.TableCellProperties{}

	// Width (Word expects tcW even for auto width)
	widthType := constants.WidthTypeAuto
	widthValue := 0
	if cell.Width() > 0 {
		widthType = constants.WidthTypeDXA
		widthValue = cell.Width()
	}
	props.Width = &xml.TableWidth{
		Type: widthType,
		W:    widthValue,
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

func emptyParagraph() *xml.Paragraph {
	return &xml.Paragraph{
		Properties: &xml.ParagraphProperties{},
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

func intPtr(i int) *int {
	return &i
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
	drawingCounter  int
}

// NewDocumentSerializer creates a new DocumentSerializer.
func NewDocumentSerializer() *DocumentSerializer {
	paraSer := NewParagraphSerializer()
	tableSer := NewTableSerializer()
	serializer := &DocumentSerializer{
		paraSerializer:  paraSer,
		tableSerializer: tableSer,
	}

	paraSer.runSerializer.SetDrawingIDProvider(serializer)
	tableSer.paraSerializer.runSerializer.SetDrawingIDProvider(serializer)

	return serializer
}

// NextDrawingID returns a unique ID for drawing elements.
func (s *DocumentSerializer) NextDrawingID() int {
	s.drawingCounter++
	return s.drawingCounter
}

// SerializeBody converts document content to xml.Body while preserving insertion order.
func (s *DocumentSerializer) SerializeBody(doc domain.Document) *xml.Body {
	blocks := doc.Blocks()
	body := &xml.Body{
		Content: make([]interface{}, 0, len(blocks)),
	}

	for _, block := range blocks {
		switch {
		case block.Paragraph != nil:
			body.Content = append(body.Content, s.paraSerializer.Serialize(block.Paragraph))
		case block.Table != nil:
			body.Content = append(body.Content, s.tableSerializer.Serialize(block.Table))
		case block.SectionBreak != nil && block.SectionBreak.Section != nil:
			sectPr := s.serializeSectionProperties(block.SectionBreak.Section)
			if sectPr == nil {
				continue
			}

			if block.SectionBreak.Type >= domain.SectionBreakTypeNextPage &&
				block.SectionBreak.Type <= domain.SectionBreakTypeOddPage {
				sectPr.Type = &xml.SectionType{Val: s.sectionBreakTypeToString(block.SectionBreak.Type)}
			}

			para := &xml.Paragraph{
				Properties: &xml.ParagraphProperties{
					SectionProperties: sectPr,
				},
			}
			body.Content = append(body.Content, para)
		}
	}

	sections := doc.Sections()
	if len(sections) > 0 {
		if sectPr := s.serializeSectionProperties(sections[len(sections)-1]); sectPr != nil {
			body.SectPr = sectPr
		}
	}

	return body
}

// SerializeDocument creates the complete document XML structure.
func (s *DocumentSerializer) SerializeDocument(doc domain.Document) *xml.Document {
	return &xml.Document{
		XMLnsW:  constants.NamespaceMain,
		XMLnsR:  constants.NamespaceRelationships,
		XMLnsWP: constants.NamespaceWordprocessingDrawing,
		Body:    s.SerializeBody(doc),
	}
}

// SerializeSectionParts converts headers and footers into their own XML parts.
// The returned maps are keyed by the part filename (e.g., "header1.xml").
func (s *DocumentSerializer) SerializeSectionParts(doc domain.Document) (map[string]*xml.Header, map[string]*xml.Footer) {
	headers := make(map[string]*xml.Header)
	footers := make(map[string]*xml.Footer)

	sections := doc.Sections()
	for _, section := range sections {
		secWithMaps, ok := section.(interface {
			HeadersAll() map[domain.HeaderType]domain.Header
			FootersAll() map[domain.FooterType]domain.Footer
		})
		if !ok {
			continue
		}

		for _, header := range secWithMaps.HeadersAll() {
			headerMeta, ok := header.(interface {
				Paragraphs() []domain.Paragraph
				RelationshipID() string
				TargetPath() string
			})
			if !ok {
				continue
			}

			target := headerMeta.TargetPath()
			if target == "" || headerMeta.RelationshipID() == "" {
				continue
			}
			if _, exists := headers[target]; exists {
				continue
			}

			xmlHeader := xml.NewHeader()
			for _, para := range headerMeta.Paragraphs() {
				xmlHeader.AddParagraph(s.paraSerializer.Serialize(para))
			}
			headers[target] = xmlHeader
		}

		for _, footer := range secWithMaps.FootersAll() {
			footerMeta, ok := footer.(interface {
				Paragraphs() []domain.Paragraph
				RelationshipID() string
				TargetPath() string
			})
			if !ok {
				continue
			}

			target := footerMeta.TargetPath()
			if target == "" || footerMeta.RelationshipID() == "" {
				continue
			}
			if _, exists := footers[target]; exists {
				continue
			}

			xmlFooter := xml.NewFooter()
			for _, para := range footerMeta.Paragraphs() {
				xmlFooter.AddParagraph(s.paraSerializer.Serialize(para))
			}
			footers[target] = xmlFooter
		}
	}

	if len(headers) == 0 {
		headers = nil
	}
	if len(footers) == 0 {
		footers = nil
	}

	return headers, footers
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
		Company:     "Misael Monterroca",
	}
}

// DebugPrint outputs document statistics for testing and debugging purposes.
func (s *DocumentSerializer) DebugPrint(doc domain.Document) {
	fmt.Printf("Document has %d paragraphs and %d tables\n",
		len(doc.Paragraphs()), len(doc.Tables()))
}

// SerializeStyles converts a domain.StyleManager to xml.Styles.
func (s *DocumentSerializer) SerializeStyles(styleManager domain.StyleManager) *xml.Styles {
	xmlStyles := xml.NewStyles()

	// Set doc defaults
	// Include Word's latent style catalog to avoid auto-added styles during repair
	xmlStyles.LatentStyles = defaultLatentStyles

	// Serialize all styles from the style manager
	for _, style := range styleManager.ListStyles() {
		xmlStyle := s.serializeStyle(style)
		if xmlStyle != nil {
			xmlStyles.AddStyle(xmlStyle)
		}
	}

	return xmlStyles
}

func (s *DocumentSerializer) serializeStyle(style domain.Style) *xml.Style {
	if style == nil {
		return nil
	}

	xmlStyle := &xml.Style{
		Type:    s.styleTypeToString(style.Type()),
		StyleID: style.ID(),
		Name:    &xml.StyleName{Val: style.Name()},
	}

	// Set default flag
	if style.IsDefault() {
		defaultVal := true
		xmlStyle.Default = &defaultVal
	}

	// Set basedOn if applicable
	if style.BasedOn() != "" {
		xmlStyle.BasedOn = &xml.BasedOn{Val: style.BasedOn()}
	}

	// For Heading styles and Normal, add qFormat
	styleID := style.ID()
	if styleID == "Normal" {
		// Normal is the base quick format style
		xmlStyle.QFormat = &struct{}{}
		xmlStyle.UIPriority = &xml.UIPriority{Val: 0}
	} else if len(styleID) >= 7 && styleID[:7] == "Heading" {
		// Mark as quick format
		xmlStyle.QFormat = &struct{}{}
		// Next paragraph should be Normal
		xmlStyle.Next = &xml.Next{Val: "Normal"}
		// Set UI priority (Headings should have high priority)
		if len(styleID) == 8 { // Heading1-9
			priority := int(styleID[7] - '0')                        // Extract digit
			xmlStyle.UIPriority = &xml.UIPriority{Val: priority + 8} // Priority 9-17
		}
	}

	// Serialize properties based on style type
	switch style.Type() {
	case domain.StyleTypeParagraph:
		xmlStyle.ParaProps = s.serializeParagraphStyleProperties(style)
		xmlStyle.RunProps = s.serializeRunStyleProperties(style)
	case domain.StyleTypeCharacter:
		xmlStyle.RunProps = s.serializeRunStyleProperties(style)
	case domain.StyleTypeTable:
		// Table styles are handled differently, no props to serialize here
	case domain.StyleTypeNumbering:
		// Numbering styles are handled differently, no props to serialize here
	}

	return xmlStyle
}

func (s *DocumentSerializer) serializeParagraphStyleProperties(style domain.Style) *xml.StyleParagraphProperties {
	props := &xml.StyleParagraphProperties{}
	hasProps := false

	// Try to access paragraph-specific properties via interface
	if ps, ok := style.(interface{ OutlineLevel() int }); ok {
		level := ps.OutlineLevel()
		// Only include outline level for Heading styles (styleId starts with "Heading")
		// Heading styles should have outline levels 0-8
		styleID := style.ID()
		if len(styleID) >= 7 && styleID[:7] == "Heading" && level >= 0 && level <= 8 {
			props.OutlineLevel = &xml.OutlineLevel{Val: level}
			hasProps = true
		}
	}

	if ps, ok := style.(interface{ SpacingBefore() int }); ok {
		if spacing := ps.SpacingBefore(); spacing > 0 {
			if props.Spacing == nil {
				props.Spacing = &xml.StyleSpacing{}
			}
			props.Spacing.Before = &spacing
			hasProps = true
		}
	}

	if ps, ok := style.(interface{ SpacingAfter() int }); ok {
		if spacing := ps.SpacingAfter(); spacing > 0 {
			if props.Spacing == nil {
				props.Spacing = &xml.StyleSpacing{}
			}
			props.Spacing.After = &spacing
			hasProps = true
		}
	}

	if ps, ok := style.(interface{ KeepNext() bool }); ok {
		if ps.KeepNext() {
			props.KeepNext = &struct{}{}
			hasProps = true
		}
	}

	if ps, ok := style.(interface{ KeepLines() bool }); ok {
		if ps.KeepLines() {
			props.KeepLines = &struct{}{}
			hasProps = true
		}
	}

	if ps, ok := style.(interface{ Indentation() domain.Indentation }); ok {
		indent := ps.Indentation()
		if indent.Left != 0 || indent.Right != 0 || indent.FirstLine != 0 || indent.Hanging != 0 {
			props.Indentation = &xml.StyleIndentation{
				Left:      intPtrIfNotZero(indent.Left),
				Right:     intPtrIfNotZero(indent.Right),
				FirstLine: intPtrIfNotZero(indent.FirstLine),
				Hanging:   intPtrIfNotZero(indent.Hanging),
			}
			hasProps = true
		}
	}

	if !hasProps {
		return nil
	}
	return props
}

func (s *DocumentSerializer) serializeRunStyleProperties(style domain.Style) *xml.RunProperties {
	props := &xml.RunProperties{}
	hasProps := false

	// Font
	font := style.Font()
	if font.Name != "" && font.Name != constants.DefaultFontName {
		props.Font = &xml.Font{
			ASCII:    font.Name,
			HAnsi:    font.Name,
			EastAsia: font.EastAsia,
			CS:       font.CS,
		}
		hasProps = true
	}

	// Bold
	if rs, ok := style.(interface{ Bold() bool }); ok {
		if rs.Bold() {
			props.Bold = &xml.BoolValue{Val: boolPtr(true)}
			hasProps = true
		}
	}

	// Italic
	if rs, ok := style.(interface{ Italic() bool }); ok {
		if rs.Italic() {
			props.Italic = &xml.BoolValue{Val: boolPtr(true)}
			hasProps = true
		}
	}

	// Color
	if rs, ok := style.(interface{ Color() domain.Color }); ok {
		color := rs.Color()
		if color != domain.ColorBlack {
			props.Color = &xml.Color{
				Val: fmt.Sprintf("%02X%02X%02X", color.R, color.G, color.B),
			}
			hasProps = true
		}
	}

	// Size
	if rs, ok := style.(interface{ Size() int }); ok {
		if size := rs.Size(); size > 0 && size != constants.DefaultFontSize {
			props.Size = &xml.HalfPt{Val: size}
			props.SizeCS = &xml.HalfPt{Val: size}
			hasProps = true
		}
	}

	// Underline
	if rs, ok := style.(interface{ Underline() domain.UnderlineStyle }); ok {
		if underline := rs.Underline(); underline != domain.UnderlineNone {
			props.Underline = &xml.Underline{
				Val: s.underlineStyleToString(underline),
			}
			hasProps = true
		}
	}

	if !hasProps {
		return nil
	}
	return props
}

func (s *DocumentSerializer) styleTypeToString(t domain.StyleType) string {
	switch t {
	case domain.StyleTypeParagraph:
		return "paragraph"
	case domain.StyleTypeCharacter:
		return "character"
	case domain.StyleTypeTable:
		return "table"
	case domain.StyleTypeNumbering:
		return "numbering"
	default:
		return "paragraph"
	}
}

func (s *DocumentSerializer) serializeSectionProperties(section domain.Section) *xml.SectionProperties {
	if section == nil {
		return nil
	}

	sectPr := xml.NewSectionProperties()

	pageSize := section.PageSize()
	orient := section.Orientation()
	landscape := orient == domain.OrientationLandscape
	if pageSize.Width == 0 || pageSize.Height == 0 {
		pageSize = domain.PageSizeLetter
	}
	sectPr.SetPageSize(pageSize.Width, pageSize.Height, landscape)

	margins := section.Margins()
	if margins == (domain.Margins{}) {
		margins = domain.DefaultMargins
	}
	sectPr.SetPageMargins(margins.Top, margins.Right, margins.Bottom, margins.Left, margins.Header, margins.Footer)

	if cols := section.Columns(); cols > 1 {
		sectPr.SetColumns(cols)
	}

	if secWithMaps, ok := section.(interface {
		HeadersAll() map[domain.HeaderType]domain.Header
		FootersAll() map[domain.FooterType]domain.Footer
	}); ok {
		for headerType, header := range secWithMaps.HeadersAll() {
			headerMeta, ok := header.(interface {
				RelationshipID() string
			})
			if !ok {
				continue
			}
			if relID := headerMeta.RelationshipID(); relID != "" {
				sectPr.AddHeaderRef(s.headerTypeToString(headerType), relID)
			}
		}

		for footerType, footer := range secWithMaps.FootersAll() {
			footerMeta, ok := footer.(interface {
				RelationshipID() string
			})
			if !ok {
				continue
			}
			if relID := footerMeta.RelationshipID(); relID != "" {
				sectPr.AddFooterRef(s.footerTypeToString(footerType), relID)
			}
		}
	}

	return sectPr
}

func (s *DocumentSerializer) sectionBreakTypeToString(bt domain.SectionBreakType) string {
	switch bt {
	case domain.SectionBreakTypeNextPage:
		return "nextPage"
	case domain.SectionBreakTypeContinuous:
		return "continuous"
	case domain.SectionBreakTypeEvenPage:
		return "evenPage"
	case domain.SectionBreakTypeOddPage:
		return "oddPage"
	default:
		return "nextPage"
	}
}

func (s *DocumentSerializer) headerTypeToString(ht domain.HeaderType) string {
	switch ht {
	case domain.HeaderDefault:
		return "default"
	case domain.HeaderFirst:
		return "first"
	case domain.HeaderEven:
		return "even"
	default:
		return "default"
	}
}

func (s *DocumentSerializer) footerTypeToString(ft domain.FooterType) string {
	switch ft {
	case domain.FooterDefault:
		return "default"
	case domain.FooterFirst:
		return "first"
	case domain.FooterEven:
		return "even"
	default:
		return "default"
	}
}

func (s *DocumentSerializer) underlineStyleToString(style domain.UnderlineStyle) string {
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

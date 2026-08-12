// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

// Package serializer converts domain objects into XML structures for OOXML serialization.
// It provides serializers for documents, paragraphs, runs, tables, and other document elements.
package serializer

import (
	"fmt"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/xml"
	"github.com/mmonterroca/docxgo/v2/pkg/color"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
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

	// Caps -- emit an explicit false too, when the source (or caller) set one,
	// since that's the only way to override a run/paragraph style's own All
	// Caps. See capsSetter.
	if run.Caps() {
		props.Caps = &xml.BoolValue{Val: boolPtr(true)}
	} else if r, ok := run.(capsSetter); ok && r.CapsSet() {
		props.Caps = &xml.BoolValue{Val: boolPtr(false)}
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

	// Language override — falls back to the document default (SerializeStyles)
	// when unset, same Val/EastAsia/Bidi shape.
	if lang := run.Language(); lang != nil {
		props.Lang = &xml.Language{Val: lang.Val, EastAsia: lang.EastAsia, Bidi: lang.Bidi}
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

	// Set when a hyperlink branch below already serialized the run's text
	// into the <w:hyperlink>'s own <w:r> -- the trailing "leftover text"
	// block at the end of this function must not serialize it again.
	textConsumedByHyperlink := false

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
				display := field.Result()
				if display == "" {
					if disp, ok := accessor.GetProperty("display"); ok {
						display = disp
					}
				}

				// Check if this is an internal bookmark link (anchor)
				// Can be set via "anchor" property directly, or via URL starting with #
				var anchor string
				if anchorProp, anchorOK := accessor.GetProperty("anchor"); anchorOK && anchorProp != "" {
					anchor = anchorProp
				} else if url, urlOK := accessor.GetProperty("url"); urlOK && strings.HasPrefix(url, "#") {
					anchor = strings.TrimPrefix(url, "#")
				}

				if anchor != "" {
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
						Anchor:  anchor,
						History: "1",
						Runs:    []*xml.Run{xmlRun},
					}
					elements = append(elements, hyperlink)
					textConsumedByHyperlink = true
					continue
				}

				// Check for external hyperlink via relationship ID
				relID, relOK := accessor.GetProperty("relationshipID")
				if relOK && relID != "" {
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
					textConsumedByHyperlink = true
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

	if run.Text() != "" && !textConsumedByHyperlink {
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

	// Indentation. Each side is gated independently, the same way spacing is
	// below: an explicit 0 on just one side (e.g. SetIndent to zero out a
	// style's right indent while leaving left alone) must be emitted for that
	// side without stamping the other three sides as explicitly zero too.
	indent := para.Indent()
	leftSet, rightSet, firstLineSet, hangingSet := indentSetFlags(para)
	if indent.Left != 0 || indent.Right != 0 || indent.FirstLine != 0 || indent.Hanging != 0 ||
		leftSet || rightSet || firstLineSet || hangingSet {
		firstLine := explicitOrNonZero(indent.FirstLine, firstLineSet)
		hanging := explicitOrNonZero(indent.Hanging, hangingSet)
		// CT_Ind allows both attributes, but Word treats them as mutually
		// exclusive and only one can render correctly. SetIndent rejects
		// setting both together, but the per-side setters (SetIndentFirstLine,
		// SetIndentHanging) don't share that check — see their doc comments
		// — so a paragraph can still reach here with both set, most often via
		// the reader hydrating a source <w:ind> that already carries both.
		// Emit only one, matching what Word itself does: hanging wins.
		if firstLine != nil && hanging != nil {
			firstLine = nil
		}
		props.Indentation = &xml.Indentation{
			Left:      explicitOrNonZero(indent.Left, leftSet),
			Right:     explicitOrNonZero(indent.Right, rightSet),
			FirstLine: firstLine,
			Hanging:   hanging,
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
	beforeSet, afterSet, lineSet := spacingSetFlags(para)

	// Emit direct spacing when the paragraph departs from the document
	// defaults, OR when the caller explicitly set a value — even one that
	// happens to equal a default, like SetSpacingAfter(0) on a paragraph
	// whose style supplies non-zero spacing. Paragraph.SpacingBefore() int
	// can't distinguish "never set" from "explicitly set to 0" on its own;
	// beforeSet/afterSet/lineSet come from the concrete type's *Set() methods
	// (see spacingSetFlags) and degrade to false for any domain.Paragraph
	// implementation that doesn't expose them, falling back to the
	// zero-value gate below. The rule matters as much as the value for line
	// spacing: an exact or at-least rule is a departure even at the default
	// 240 twips, and without this check it would silently inherit the
	// lineRule="auto" written into w:pPrDefault.
	//
	// Before/After and Line/LineRule are gated independently within the same
	// element: a paragraph that only set spacingBefore/spacingAfter must not
	// also emit w:line/w:lineRule, since lineSpacingRuleToString never
	// returns nil and this fix's own beforeSet/afterSet would otherwise stamp
	// the default 240/auto onto every such paragraph, silently overriding a
	// style's own line spacing.
	lineDeparts := lineSet ||
		lineSpacing.Value != constants.DefaultLineSpacing ||
		lineSpacing.Rule != domain.LineSpacingAuto

	if before != 0 || after != 0 || beforeSet || afterSet || lineDeparts {
		spacing := &xml.Spacing{
			Before: explicitOrNonZero(before, beforeSet),
			After:  explicitOrNonZero(after, afterSet),
		}
		if lineDeparts {
			spacing.Line = explicitOrNonZero(lineSpacing.Value, lineSet)
			spacing.LineRule = s.lineSpacingRuleToString(lineSpacing.Rule)
		}
		props.Spacing = spacing
	}

	// Borders
	borders := para.Borders()
	if s.hasBorders(borders) {
		props.Borders = &xml.ParagraphBorders{
			Top:    s.serializeBorder(borders.Top),
			Bottom: s.serializeBorder(borders.Bottom),
			Left:   s.serializeBorder(borders.Left),
			Right:  s.serializeBorder(borders.Right),
		}
	}

	return props
}

func (s *ParagraphSerializer) hasBorders(borders domain.ParagraphBorders) bool {
	return borders.Top.Style != domain.BorderNone ||
		borders.Bottom.Style != domain.BorderNone ||
		borders.Left.Style != domain.BorderNone ||
		borders.Right.Style != domain.BorderNone
}

func (s *ParagraphSerializer) serializeBorder(border domain.BorderStyle) *xml.Border {
	if border.Style == domain.BorderNone {
		return nil
	}

	return &xml.Border{
		Val:   s.borderStyleToString(border.Style),
		Color: color.ToHex(border.Color),
		Sz:    border.Width,
	}
}

func (s *ParagraphSerializer) borderStyleToString(style domain.BorderLineStyle) string {
	switch style {
	case domain.BorderNone:
		return "none"
	case domain.BorderSingle:
		return "single"
	case domain.BorderDashed:
		return "dashed"
	case domain.BorderDotted:
		return "dotted"
	case domain.BorderDouble:
		return "double"
	case domain.BorderThick:
		return "thick"
	case domain.BorderTriple:
		return "triple"
	default:
		return "none"
	}
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
	colCount := table.ColumnCount()
	grid := &xml.TableGrid{
		Cols: make([]*xml.GridCol, colCount),
	}

	// Try to derive column widths from the first row's cell widths.
	// If cells have explicit widths, use them for the grid columns;
	// otherwise fall back to nil (omitted) so Word auto-calculates.
	if table.RowCount() > 0 {
		if firstRow, err := table.Row(0); err == nil {
			cells := firstRow.Cells()
			colIdx := 0
			for _, cell := range cells {
				if cell.IsHorizontallyMergedContinuation() {
					continue
				}
				w := cell.Width()
				span := cell.GridSpan()
				if span < 1 {
					span = 1
				}
				// Distribute the cell width evenly across spanned columns.
				perCol := 0
				if w > 0 && span > 0 {
					perCol = w / span
				}
				for j := 0; j < span && colIdx < colCount; j++ {
					if perCol > 0 {
						grid.Cols[colIdx] = &xml.GridCol{W: intPtr(perCol)}
					} else {
						grid.Cols[colIdx] = &xml.GridCol{}
					}
					colIdx++
				}
			}
			// Fill remaining columns (safety).
			for ; colIdx < colCount; colIdx++ {
				grid.Cols[colIdx] = &xml.GridCol{}
			}
			return grid
		}
	}

	// Fallback: no rows or error – omit widths so Word auto-calculates.
	for i := 0; i < colCount; i++ {
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

	// Borders
	borders := cell.Borders()
	if s.hasCellBorders(borders) {
		props.Borders = s.serializeCellBorders(borders)
	}

	return props
}

func (s *TableSerializer) hasCellBorders(borders domain.TableBorders) bool {
	return borders.Top.Style != domain.BorderNone ||
		borders.Bottom.Style != domain.BorderNone ||
		borders.Left.Style != domain.BorderNone ||
		borders.Right.Style != domain.BorderNone
}

func (s *TableSerializer) serializeCellBorders(borders domain.TableBorders) *xml.TableBorders {
	xmlBorders := &xml.TableBorders{}

	if borders.Top.Style != domain.BorderNone {
		xmlBorders.Top = s.serializeCellBorder(borders.Top)
	}
	if borders.Bottom.Style != domain.BorderNone {
		xmlBorders.Bottom = s.serializeCellBorder(borders.Bottom)
	}
	if borders.Left.Style != domain.BorderNone {
		xmlBorders.Left = s.serializeCellBorder(borders.Left)
	}
	if borders.Right.Style != domain.BorderNone {
		xmlBorders.Right = s.serializeCellBorder(borders.Right)
	}

	return xmlBorders
}

func (s *TableSerializer) serializeCellBorder(border domain.BorderStyle) *xml.Border {
	return &xml.Border{
		Val:   s.borderLineStyleToString(border.Style),
		Sz:    border.Width,
		Color: color.ToHex(border.Color),
	}
}

func (s *TableSerializer) borderLineStyleToString(style domain.BorderLineStyle) string {
	switch style {
	case domain.BorderNone:
		return "none"
	case domain.BorderSingle:
		return "single"
	case domain.BorderDashed:
		return "dashed"
	case domain.BorderDotted:
		return "dotted"
	case domain.BorderDouble:
		return "double"
	case domain.BorderThick:
		return "thick"
	case domain.BorderTriple:
		return "triple"
	default:
		return "none"
	}
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

// capsSetter exposes whether a run's SetCaps was ever called, letting the
// serializer emit an explicit <w:caps w:val="false"/> instead of treating a
// false Caps() as "unset" -- the only way to override a style's own All
// Caps. Implemented by the concrete run type in internal/core; degrades to
// "never explicitly set" for any other domain.Run implementation.
type capsSetter interface {
	CapsSet() bool
}

// spacingSetter exposes whether a paragraph's spacing setters were ever
// called, letting the serializer emit an explicit 0 instead of treating it as
// "unset". Implemented by the concrete paragraph type in internal/core; see
// serializeProperties.
type spacingSetter interface {
	SpacingBeforeSet() bool
	SpacingAfterSet() bool
	LineSpacingSet() bool
}

// spacingSetFlags returns whether para's spacing-before, spacing-after, and
// line-spacing were explicitly set, degrading to all-false for any
// domain.Paragraph implementation that doesn't expose spacingSetter.
func spacingSetFlags(para domain.Paragraph) (beforeSet, afterSet, lineSet bool) {
	if p, ok := para.(spacingSetter); ok {
		return p.SpacingBeforeSet(), p.SpacingAfterSet(), p.LineSpacingSet()
	}
	return false, false, false
}

// explicitOrNonZero returns a pointer to value for XML serialization: present
// whenever the caller explicitly set the property, even at zero, so it
// overrides a style's own value instead of being mistaken for "unset".
// Otherwise it falls back to omitting zero, letting docDefaults or the
// paragraph's style supply the value. Shared by spacing and indentation.
func explicitOrNonZero(value int, explicitlySet bool) *int {
	if explicitlySet {
		return intPtr(value)
	}
	return intPtrIfNotZero(value)
}

// indentSetter exposes whether a paragraph's individual indentation fields
// were ever set, letting the serializer emit an explicit 0 on just the sides
// the caller touched instead of either all four or none. Implemented by the
// concrete paragraph type in internal/core; see serializeProperties.
//
// SetIndent takes the whole domain.Indentation struct in one call, so it
// cannot itself express "only Left was set" -- the per-field setters this
// interface exposes are concrete-type-only (not part of domain.Paragraph)
// and exist specifically so the reader can mark only the attributes present
// in a source <w:ind> element, without SetIndent's all-or-nothing granularity
// forcing a re-serialized document to claim sides the source never set.
type indentSetter interface {
	IndentLeftSet() bool
	IndentRightSet() bool
	IndentFirstLineSet() bool
	IndentHangingSet() bool
}

// indentSetFlags returns whether para's four indentation fields were
// explicitly set, degrading to all-false for any domain.Paragraph
// implementation that doesn't expose indentSetter.
func indentSetFlags(para domain.Paragraph) (leftSet, rightSet, firstLineSet, hangingSet bool) {
	if p, ok := para.(indentSetter); ok {
		return p.IndentLeftSet(), p.IndentRightSet(), p.IndentFirstLineSet(), p.IndentHangingSet()
	}
	return false, false, false, false
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
	var background *xml.Background
	if colorValue, ok := doc.BackgroundColor(); ok {
		background = &xml.Background{Color: color.ToHex(colorValue)}
	}

	return &xml.Document{
		XMLnsW:     constants.NamespaceMain,
		XMLnsR:     constants.NamespaceRelationships,
		XMLnsWP:    constants.NamespaceWordprocessingDrawing,
		Background: background,
		Body:       s.SerializeBody(doc),
	}
}

// SerializeSectionParts converts headers and footers into their own XML parts.
// The returned maps are keyed by the part filename (e.g., "header1.xml").
// headerFooterPart is the subset of docxHeader/docxFooter that
// SerializeSectionParts needs: the block content to serialize, and the
// relationship metadata that decides whether (and where) to serialize it.
type headerFooterPart interface {
	Blocks() []domain.Block
	RelationshipID() string
	TargetPath() string
}

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
			headerMeta, ok := header.(headerFooterPart)
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
			xmlHeader.Content = s.serializeHeaderFooterContent(headerMeta.Blocks())
			headers[target] = xmlHeader
		}

		for _, footer := range secWithMaps.FootersAll() {
			footerMeta, ok := footer.(headerFooterPart)
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
			xmlFooter.Content = s.serializeHeaderFooterContent(footerMeta.Blocks())
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

// serializeHeaderFooterContent converts a header's or footer's top-level
// blocks to XML content in insertion order, mirroring SerializeBody minus
// the SectionBreak arm (w:hdr/w:ftr have no section properties of their
// own). Two adjacent w:tbl with no intervening w:p are coalesced by Word
// into a single table using the first table's grid, so a separator
// paragraph is inserted between consecutive tables and after a trailing
// table — the same policy serializeCell already applies to nested tables.
func (s *DocumentSerializer) serializeHeaderFooterContent(blocks []domain.Block) []interface{} {
	content := make([]interface{}, 0, len(blocks)+1)
	prevWasTable := false

	for _, block := range blocks {
		switch {
		case block.Paragraph != nil:
			content = append(content, s.paraSerializer.Serialize(block.Paragraph))
			prevWasTable = false
		case block.Table != nil:
			if prevWasTable {
				content = append(content, emptyParagraph())
			}
			content = append(content, s.tableSerializer.Serialize(block.Table))
			prevWasTable = true
		}
	}

	if prevWasTable {
		content = append(content, emptyParagraph())
	}

	return content
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
		Application: "docxgo",
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

// documentDefaultParagraphProperties returns the w:pPrDefault contents shared
// by every generated document: 0pt before/after spacing and single line
// spacing (240 twips, auto rule). Without an explicit default, an empty
// <w:pPr> (or a paragraph that never sets spacing) falls back to Word's own
// defaults — 8pt after, 1.15 line spacing — instead of the 0/240 the domain
// model assumes (constants.DefaultParagraphSpacing / DefaultLineSpacing).
func documentDefaultParagraphProperties() *xml.ParagraphDefaults {
	return &xml.ParagraphDefaults{
		Properties: &xml.StyleParagraphProperties{
			Spacing: &xml.StyleSpacing{
				Before:   intPtr(0),
				After:    intPtr(0),
				Line:     intPtr(constants.DefaultLineSpacing),
				LineRule: "auto",
			},
		},
	}
}

// SerializeStyles converts a domain.StyleManager to xml.Styles, writing lang,
// defaultFont, and defaultFontSize (any of which may be nil) into
// w:docDefaults/w:rPrDefault. docDefaults sits below the Normal style in the
// OOXML cascade, so a theme's own font/size on Normal still wins over these —
// they only take effect for styles that don't set their own.
func (s *DocumentSerializer) SerializeStyles(styleManager domain.StyleManager, lang *domain.Language, defaultFont *string, defaultFontSize *int) *xml.Styles {
	xmlStyles := xml.NewStyles()

	// Include Word's latent style catalog to avoid auto-added styles during repair
	xmlStyles.LatentStyles = defaultLatentStyles

	xmlStyles.DocDefaults = &xml.DocDefaults{
		ParaDefaults: documentDefaultParagraphProperties(),
	}
	if lang != nil || defaultFont != nil || defaultFontSize != nil {
		props := &xml.RunProperties{}
		if lang != nil {
			props.Lang = &xml.Language{Val: lang.Val, EastAsia: lang.EastAsia, Bidi: lang.Bidi}
		}
		if defaultFont != nil {
			props.Font = &xml.Font{ASCII: *defaultFont, HAnsi: *defaultFont}
		}
		if defaultFontSize != nil {
			props.Size = &xml.HalfPt{Val: *defaultFontSize}
			props.SizeCS = &xml.HalfPt{Val: *defaultFontSize}
		}
		xmlStyles.DocDefaults.RunDefaults = &xml.RunDefaults{Properties: props}
	}

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

	// Set link if applicable (for paragraph styles linked to character styles)
	if ps, ok := style.(interface{ Link() string }); ok {
		if link := ps.Link(); link != "" {
			xmlStyle.Link = &xml.Link{Val: link}
		}
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
		xmlStyle.TblPr = s.serializeTableStyleProperties(style)
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

	if ps, ok := style.(interface{ Alignment() domain.Alignment }); ok {
		if align := ps.Alignment(); align != domain.AlignmentLeft {
			props.Alignment = &xml.Alignment{Val: s.paraSerializer.alignmentToString(align)}
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

func (s *DocumentSerializer) serializeTableStyleProperties(style domain.Style) *xml.TableStyleProperties {
	tsd, ok := style.(domain.TableStyleDef)
	if !ok {
		return nil
	}

	props := &xml.TableStyleProperties{}
	hasProps := false

	// Table borders
	if tsd.HasTableBorders() {
		borders := tsd.TableBorders()
		xmlBorders := &xml.TableLevelBorders{}
		hasBorders := false

		if borders.Top.Style != domain.BorderNone {
			xmlBorders.Top = s.serializeStyleBorder(borders.Top)
			hasBorders = true
		}
		if borders.Left.Style != domain.BorderNone {
			xmlBorders.Left = s.serializeStyleBorder(borders.Left)
			hasBorders = true
		}
		if borders.Bottom.Style != domain.BorderNone {
			xmlBorders.Bottom = s.serializeStyleBorder(borders.Bottom)
			hasBorders = true
		}
		if borders.Right.Style != domain.BorderNone {
			xmlBorders.Right = s.serializeStyleBorder(borders.Right)
			hasBorders = true
		}
		if borders.InsideH.Style != domain.BorderNone {
			xmlBorders.InsideH = s.serializeStyleBorder(borders.InsideH)
			hasBorders = true
		}
		if borders.InsideV.Style != domain.BorderNone {
			xmlBorders.InsideV = s.serializeStyleBorder(borders.InsideV)
			hasBorders = true
		}

		if hasBorders {
			props.Borders = xmlBorders
			hasProps = true
		}
	}

	// Cell margins
	top, left, bottom, right := tsd.CellMargins()
	if top > 0 || left > 0 || bottom > 0 || right > 0 {
		props.CellMar = &xml.TableCellMargins{}
		if top > 0 {
			props.CellMar.Top = &xml.TableCellMargin{W: top, Type: constants.WidthTypeDXA}
		}
		if left > 0 {
			props.CellMar.Left = &xml.TableCellMargin{W: left, Type: constants.WidthTypeDXA}
		}
		if bottom > 0 {
			props.CellMar.Bottom = &xml.TableCellMargin{W: bottom, Type: constants.WidthTypeDXA}
		}
		if right > 0 {
			props.CellMar.Right = &xml.TableCellMargin{W: right, Type: constants.WidthTypeDXA}
		}
		hasProps = true
	}

	if !hasProps {
		return nil
	}
	return props
}

func (s *DocumentSerializer) serializeStyleBorder(border domain.BorderStyle) *xml.Border {
	b := &xml.Border{
		Val: s.borderLineStyleToString(border.Style),
		Sz:  border.Width,
	}
	// Use "auto" for zero-value color (black), otherwise convert to hex
	if border.Color == (domain.Color{}) {
		b.Color = "auto"
	} else {
		b.Color = color.ToHex(border.Color)
	}
	return b
}

func (s *DocumentSerializer) borderLineStyleToString(style domain.BorderLineStyle) string {
	switch style {
	case domain.BorderNone:
		return "none"
	case domain.BorderSingle:
		return "single"
	case domain.BorderDashed:
		return "dashed"
	case domain.BorderDotted:
		return "dotted"
	case domain.BorderDouble:
		return "double"
	case domain.BorderThick:
		return "thick"
	case domain.BorderTriple:
		return "triple"
	default:
		return "none"
	}
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

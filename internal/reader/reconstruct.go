// MIT License
//
// Copyright (c) 2025 Misael Monterroca
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package reader

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/internal/core"
	xmlstructs "github.com/mmonterroca/docxgo/internal/xml"
	pkgcolor "github.com/mmonterroca/docxgo/pkg/color"
	"github.com/mmonterroca/docxgo/pkg/constants"
	"github.com/mmonterroca/docxgo/pkg/errors"
)

const (
	opReconstructDocument     = "reader.ReconstructDocument"
	opHydrateParagraph        = "reader.hydrateParagraph"
	opApplyParagraphSpacing   = "reader.applyParagraphSpacing"
	opApplyParagraphAlignment = "reader.applyParagraphAlignment"
	opApplyParagraphIndent    = "reader.applyParagraphIndent"
	opApplyParagraphNumbering = "reader.applyParagraphNumbering"
	opHydrateRun              = "reader.hydrateRun"
	opApplyRunProperties      = "reader.applyRunProperties"
	opAttachFieldToRun        = "reader.attachFieldToRun"
	opHydrateHyperlink        = "reader.hydrateHyperlink"
	opHydrateSimpleField      = "reader.hydrateSimpleField"
	opHydrateDrawing          = "reader.hydrateDrawing"
	opBuildField              = "reader.buildFieldFromInstruction"
	opHydrateTable            = "reader.hydrateTable"
	opHydrateTableCell        = "reader.hydrateTableCell"
	opApplySectionProperties  = "reader.applySectionProperties"
	opHydrateSectionHeader    = "reader.hydrateSectionHeader"
	opHydrateSectionFooter    = "reader.hydrateSectionFooter"
)

type reconstructContext struct {
	relationships            map[string]*xmlstructs.Relationship
	media                    map[string]*MediaPart
	doc                      domain.Document
	parsed                   *ParsedPackage
	currentSection           domain.Section
	hydratedHeaders          map[domain.Section]map[domain.HeaderType]bool
	hydratedFooters          map[domain.Section]map[domain.FooterType]bool
	suppressSectionHydration int
}

type fieldState struct {
	ctx             *reconstructContext
	instruction     strings.Builder
	active          bool
	expectingResult bool
	pendingField    domain.Field
}

// ReconstructDocument converts a ParsedPackage into a domain.Document.
// This performs a minimal hydration pass that focuses on paragraph content
// and spacing so consumers can round-trip spacing metadata.
func ReconstructDocument(parsed *ParsedPackage) (domain.Document, error) {
	if parsed == nil {
		return nil, errors.Errorf(errors.ErrCodeInvalidState, opReconstructDocument, "parsed package cannot be nil")
	}
	if parsed.DocumentTree == nil {
		return nil, errors.Errorf(errors.ErrCodeInvalidState, opReconstructDocument, "document part is missing")
	}

	body := findChild(parsed.DocumentTree, "body")
	if body == nil {
		return nil, errors.Errorf(errors.ErrCodeInvalidState, opReconstructDocument, "document body is missing")
	}

	doc := core.NewDocument()
	defaultSection, err := doc.DefaultSection()
	if err != nil {
		return nil, errors.Wrap(err, opReconstructDocument)
	}

	ctx := newReconstructContext(doc, parsed, defaultSection)

	if registrar, ok := doc.(interface {
		RegisterExistingRelationship(string, string, string, string) error
	}); ok && parsed.DocumentRelationships != nil {
		for _, rel := range parsed.DocumentRelationships.Relationships {
			if rel == nil || rel.ID == "" {
				continue
			}
			if err := registrar.RegisterExistingRelationship(rel.ID, rel.Type, rel.Target, rel.TargetMode); err != nil {
				return nil, errors.Wrap(err, opReconstructDocument)
			}
			if rel.Type == constants.RelTypeNumbering && len(parsed.Numbering) > 0 {
				if setter, ok := doc.(interface{ SetNumberingPart([]byte, string) }); ok {
					setter.SetNumberingPart(parsed.Numbering, rel.Target)
				}
			}
		}
	} else if len(parsed.Numbering) > 0 {
		if setter, ok := doc.(interface{ SetNumberingPart([]byte, string) }); ok {
			setter.SetNumberingPart(parsed.Numbering, constants.PathNumbering)
		}
	}

	for _, child := range body.Children {
		if child == nil {
			continue
		}

		switch child.Name.Local {
		case "p":
			if err := hydrateParagraph(doc, child, ctx); err != nil {
				return nil, errors.Wrap(err, opReconstructDocument)
			}
		case "tbl":
			if err := hydrateTable(doc, child, ctx); err != nil {
				return nil, errors.Wrap(err, opReconstructDocument)
			}
		}
	}

	if sectPr := findChild(body, "sectPr"); sectPr != nil {
		if err := ctx.applySectionProperties(sectPr); err != nil {
			return nil, errors.Wrap(err, opReconstructDocument)
		}
	}

	return doc, nil
}

func hydrateParagraph(doc domain.Document, elem *Element, ctx *reconstructContext) error {
	para, err := doc.AddParagraph()
	if err != nil {
		return errors.Wrap(err, opHydrateParagraph)
	}

	return populateParagraph(para, elem, ctx)
}

func populateParagraph(para domain.Paragraph, elem *Element, ctx *reconstructContext) error {
	if para == nil || elem == nil {
		return nil
	}

	if props := findChild(elem, "pPr"); props != nil {
		if err := applyParagraphProperties(para, props); err != nil {
			return err
		}
	}

	state := newFieldState(ctx)

	for _, child := range elem.Children {
		if child == nil {
			continue
		}

		switch child.Name.Local {
		case "r":
			if _, err := hydrateRun(para, child, ctx, state, nil); err != nil {
				return err
			}
		case "hyperlink":
			if err := hydrateHyperlink(para, child, ctx, state); err != nil {
				return err
			}
		case "fldSimple":
			if err := hydrateSimpleField(para, child, ctx, state); err != nil {
				return err
			}
		}
	}

	state.reset()

	if ctx != nil {
		if props := findChild(elem, "pPr"); props != nil {
			if sectPr := findChild(props, "sectPr"); sectPr != nil {
				if err := ctx.applySectionProperties(sectPr); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func applyParagraphProperties(para domain.Paragraph, props *Element) error {
	// Apply paragraph style first (must be before other formatting)
	if err := applyParagraphStyle(para, props); err != nil {
		return err
	}
	if err := applyParagraphSpacing(para, props); err != nil {
		return err
	}
	if err := applyParagraphAlignment(para, props); err != nil {
		return err
	}
	if err := applyParagraphIndentation(para, props); err != nil {
		return err
	}
	if err := applyParagraphNumbering(para, props); err != nil {
		return err
	}
	return nil
}

func applyParagraphStyle(para domain.Paragraph, props *Element) error {
	pStyleElem := findChild(props, "pStyle")
	if pStyleElem == nil {
		return nil
	}

	if styleID, ok := getAttr(pStyleElem, "val"); ok && styleID != "" {
		if err := para.SetStyle(styleID); err != nil {
			return errors.WrapWithContext(err, "applyParagraphStyle", map[string]interface{}{"styleID": styleID})
		}
	}

	return nil
}

func applyParagraphSpacing(para domain.Paragraph, props *Element) error {
	spacingElem := findChild(props, "spacing")
	if spacingElem == nil {
		return nil
	}

	if val, ok := getAttr(spacingElem, "before"); ok && val != "" {
		twips, err := strconv.Atoi(val)
		if err != nil {
			return errors.WrapWithContext(err, opApplyParagraphSpacing, map[string]interface{}{"attr": "before", "value": val})
		}
		if err := para.SetSpacingBefore(twips); err != nil {
			return err
		}
	}

	if val, ok := getAttr(spacingElem, "after"); ok && val != "" {
		twips, err := strconv.Atoi(val)
		if err != nil {
			return errors.WrapWithContext(err, opApplyParagraphSpacing, map[string]interface{}{"attr": "after", "value": val})
		}
		if err := para.SetSpacingAfter(twips); err != nil {
			return err
		}
	}

	lineSpacing := para.LineSpacing()
	valueChanged := false
	ruleChanged := false

	if val, ok := getAttr(spacingElem, "line"); ok && val != "" {
		twips, err := strconv.Atoi(val)
		if err != nil {
			return errors.WrapWithContext(err, opApplyParagraphSpacing, map[string]interface{}{"attr": "line", "value": val})
		}
		lineSpacing.Value = twips
		valueChanged = true
	}

	if val, ok := getAttr(spacingElem, "lineRule"); ok && val != "" {
		lineSpacing.Rule = mapLineSpacingRule(val)
		ruleChanged = true
	}

	if valueChanged || ruleChanged {
		if err := para.SetLineSpacing(lineSpacing); err != nil {
			return err
		}
	}

	return nil
}

func applyParagraphAlignment(para domain.Paragraph, props *Element) error {
	jc := findChild(props, "jc")
	if jc == nil {
		return nil
	}

	if val, ok := getAttr(jc, "val"); ok && val != "" {
		if align, ok := mapAlignment(val); ok {
			if err := para.SetAlignment(align); err != nil {
				return errors.Wrap(err, opApplyParagraphAlignment)
			}
		}
	}

	return nil
}

func applyParagraphIndentation(para domain.Paragraph, props *Element) error {
	ind := findChild(props, "ind")
	if ind == nil {
		return nil
	}

	current := para.Indent()
	changed := false

	if val, ok := getAttr(ind, "left"); ok && val != "" {
		twips, err := strconv.Atoi(val)
		if err != nil {
			return errors.WrapWithContext(err, opApplyParagraphIndent, map[string]interface{}{"attr": "left", "value": val})
		}
		current.Left = twips
		changed = true
	}

	if val, ok := getAttr(ind, "right"); ok && val != "" {
		twips, err := strconv.Atoi(val)
		if err != nil {
			return errors.WrapWithContext(err, opApplyParagraphIndent, map[string]interface{}{"attr": "right", "value": val})
		}
		current.Right = twips
		changed = true
	}

	if val, ok := getAttr(ind, "firstLine"); ok && val != "" {
		twips, err := strconv.Atoi(val)
		if err != nil {
			return errors.WrapWithContext(err, opApplyParagraphIndent, map[string]interface{}{"attr": "firstLine", "value": val})
		}
		current.FirstLine = twips
		changed = true
	}

	if val, ok := getAttr(ind, "hanging"); ok && val != "" {
		twips, err := strconv.Atoi(val)
		if err != nil {
			return errors.WrapWithContext(err, opApplyParagraphIndent, map[string]interface{}{"attr": "hanging", "value": val})
		}
		current.Hanging = twips
		changed = true
	}

	if changed {
		if err := para.SetIndent(current); err != nil {
			return errors.Wrap(err, opApplyParagraphIndent)
		}
	}

	return nil
}

func applyParagraphNumbering(para domain.Paragraph, props *Element) error {
	if para == nil || props == nil {
		return nil
	}

	numPr := findChild(props, "numPr")
	if numPr == nil {
		para.ClearNumbering()
		return nil
	}

	ref := domain.NumberingReference{}
	foundID := false

	if numID := findChild(numPr, "numId"); numID != nil {
		if val, ok := getAttr(numID, "val"); ok && val != "" {
			id, err := strconv.Atoi(val)
			if err != nil {
				return errors.WrapWithContext(err, opApplyParagraphNumbering, map[string]interface{}{"attr": "numId", "value": val})
			}
			ref.ID = id
			foundID = true
		}
	}

	if !foundID {
		para.ClearNumbering()
		return nil
	}

	if ilvl := findChild(numPr, "ilvl"); ilvl != nil {
		if val, ok := getAttr(ilvl, "val"); ok && val != "" {
			lvl, err := strconv.Atoi(val)
			if err != nil {
				return errors.WrapWithContext(err, opApplyParagraphNumbering, map[string]interface{}{"attr": "ilvl", "value": val})
			}
			ref.Level = lvl
		}
	}

	if err := para.SetNumbering(ref); err != nil {
		return errors.Wrap(err, opApplyParagraphNumbering)
	}

	return nil
}

func hydrateRun(para domain.Paragraph, elem *Element, ctx *reconstructContext, state *fieldState, extraFields []domain.Field) (domain.Run, error) {
	if para == nil || elem == nil {
		return nil, nil
	}

	var (
		textBuilder strings.Builder
		breaks      []domain.BreakType
		props       *Element
		drawings    []*Element
	)

	for _, child := range elem.Children {
		if child == nil {
			continue
		}

		switch child.Name.Local {
		case "t":
			textBuilder.WriteString(child.Text)
		case "tab":
			textBuilder.WriteRune('\t')
		case "br":
			breaks = append(breaks, mapBreakType(child))
		case "fldChar":
			if state != nil {
				if err := state.handleFieldChar(child); err != nil {
					return nil, err
				}
			}
		case "instrText":
			if state != nil {
				state.appendInstruction(child.Text)
			}
		case "rPr":
			props = child
		case "drawing":
			drawings = append(drawings, child)
		}
	}

	createRun := textBuilder.Len() > 0 || len(breaks) > 0 || len(extraFields) > 0 || len(drawings) > 0
	if !createRun && state != nil && state.shouldForceRun() {
		createRun = true
	}

	if !createRun {
		return nil, nil
	}

	run, err := para.AddRun()
	if err != nil {
		return nil, errors.Wrap(err, opHydrateRun)
	}

	if textBuilder.Len() > 0 {
		if err := run.SetText(textBuilder.String()); err != nil {
			return nil, errors.Wrap(err, opHydrateRun)
		}
	}

	if props != nil {
		if err := applyRunProperties(run, props); err != nil {
			return nil, err
		}
	}

	for _, br := range breaks {
		if err := run.AddBreak(br); err != nil {
			return nil, errors.Wrap(err, opHydrateRun)
		}
	}

	if state != nil {
		if err := state.attachToRun(run); err != nil {
			return nil, err
		}
	}

	for _, field := range extraFields {
		if field == nil {
			continue
		}
		if setter, ok := field.(interface{ SetResult(string) }); ok {
			setter.SetResult(run.Text())
		}
		if accessor, ok := field.(interface{ SetProperty(string, string) }); ok {
			accessor.SetProperty("display", run.Text())
		}
		if err := run.AddField(field); err != nil {
			return nil, errors.Wrap(err, opHydrateRun)
		}
	}

	if len(drawings) > 0 {
		for _, drawing := range drawings {
			if err := hydrateDrawing(para, run, drawing, ctx); err != nil {
				return nil, err
			}
		}
	}

	return run, nil
}

func applyRunProperties(run domain.Run, props *Element) error {
	if run == nil || props == nil {
		return nil
	}

	if boldElem := findChild(props, "b"); boldElem != nil {
		if val, ok := parseOnOff(boldElem); ok {
			if err := run.SetBold(val); err != nil {
				return errors.Wrap(err, opApplyRunProperties)
			}
		}
	}

	if italicElem := findChild(props, "i"); italicElem != nil {
		if val, ok := parseOnOff(italicElem); ok {
			if err := run.SetItalic(val); err != nil {
				return errors.Wrap(err, opApplyRunProperties)
			}
		}
	}

	if strikeElem := findChild(props, "strike"); strikeElem != nil {
		if val, ok := parseOnOff(strikeElem); ok {
			if err := run.SetStrike(val); err != nil {
				return errors.Wrap(err, opApplyRunProperties)
			}
		}
	}

	if underlineElem := findChild(props, "u"); underlineElem != nil {
		underlineVal, ok := getAttr(underlineElem, "val")
		if !ok || underlineVal == "" {
			underlineVal = constants.UnderlineValueSingle
		}
		if style, mapped := mapUnderlineStyle(underlineVal); mapped && style != domain.UnderlineNone {
			if err := run.SetUnderline(style); err != nil {
				return errors.Wrap(err, opApplyRunProperties)
			}
		}
	}

	if colorElem := findChild(props, "color"); colorElem != nil {
		if val, ok := getAttr(colorElem, "val"); ok && val != "" && !strings.EqualFold(val, "auto") {
			clr, err := pkgcolor.FromHex(val)
			if err != nil {
				return errors.WrapWithContext(err, opApplyRunProperties, map[string]interface{}{"attr": "color", "value": val})
			}
			if err := run.SetColor(clr); err != nil {
				return errors.Wrap(err, opApplyRunProperties)
			}
		}
	}

	sizeVal := ""
	if szElem := findChild(props, "sz"); szElem != nil {
		if val, ok := getAttr(szElem, "val"); ok && val != "" {
			sizeVal = val
		}
	}
	if sizeVal == "" {
		if szCsElem := findChild(props, "szCs"); szCsElem != nil {
			if val, ok := getAttr(szCsElem, "val"); ok && val != "" {
				sizeVal = val
			}
		}
	}
	if sizeVal != "" {
		halfPoints, err := strconv.Atoi(sizeVal)
		if err != nil {
			return errors.WrapWithContext(err, opApplyRunProperties, map[string]interface{}{"attr": "sz", "value": sizeVal})
		}
		if err := run.SetSize(halfPoints); err != nil {
			return errors.Wrap(err, opApplyRunProperties)
		}
	}

	if fontElem := findChild(props, "rFonts"); fontElem != nil {
		current := run.Font()
		updated := current
		changed := false

		if val, ok := getAttr(fontElem, "ascii"); ok && val != "" {
			updated.Name = val
			changed = true
		} else if val, ok := getAttr(fontElem, "hAnsi"); ok && val != "" {
			updated.Name = val
			changed = true
		}

		if val, ok := getAttr(fontElem, "eastAsia"); ok && val != "" {
			updated.EastAsia = val
			changed = true
		}

		if val, ok := getAttr(fontElem, "cs"); ok && val != "" {
			updated.CS = val
			changed = true
		}

		if changed {
			if updated.Name == "" {
				updated.Name = current.Name
			}
			if err := run.SetFont(updated); err != nil {
				return errors.Wrap(err, opApplyRunProperties)
			}
		}
	}

	if highlightElem := findChild(props, "highlight"); highlightElem != nil {
		if val, ok := getAttr(highlightElem, "val"); ok && val != "" {
			if highlight, mapped := mapHighlightColor(val); mapped && highlight != domain.HighlightNone {
				if err := run.SetHighlight(highlight); err != nil {
					return errors.Wrap(err, opApplyRunProperties)
				}
			}
		}
	}

	return nil
}

func hydrateHyperlink(para domain.Paragraph, elem *Element, ctx *reconstructContext, state *fieldState) error {
	if para == nil || elem == nil {
		return nil
	}

	if state != nil {
		state.reset()
	}

	url := ""
	if relID, ok := getAttr(elem, "id"); ok {
		if target, found := ctx.resolveRelationshipTarget(relID); found {
			url = target
		}
	}

	if url == "" {
		if anchor, ok := getAttr(elem, "anchor"); ok && anchor != "" {
			if strings.HasPrefix(anchor, "#") {
				url = anchor
			} else {
				url = "#" + anchor
			}
		}
	}

	for _, child := range elem.Children {
		if child == nil || child.Name.Local != "r" {
			continue
		}

		var extraFields []domain.Field
		if url != "" {
			field := core.NewField(domain.FieldTypeHyperlink)
			if err := field.SetCode(fmt.Sprintf(`HYPERLINK "%s"`, url)); err != nil {
				return errors.Wrap(err, opHydrateHyperlink)
			}
			if accessor, ok := field.(interface{ SetProperty(string, string) }); ok {
				accessor.SetProperty("url", url)
			}
			extraFields = []domain.Field{field}
		}

		if _, err := hydrateRun(para, child, ctx, state, extraFields); err != nil {
			return errors.Wrap(err, opHydrateHyperlink)
		}
	}

	if state != nil {
		state.reset()
	}

	return nil
}

func hydrateSimpleField(para domain.Paragraph, elem *Element, ctx *reconstructContext, state *fieldState) error {
	if para == nil || elem == nil {
		return nil
	}

	if state != nil {
		state.reset()
	}

	instr, _ := getAttr(elem, "instr")

	for _, child := range elem.Children {
		if child == nil || child.Name.Local != "r" {
			continue
		}

		var extra []domain.Field
		if instr != "" {
			field, err := buildFieldFromInstruction(instr)
			if err != nil {
				return errors.Wrap(err, opHydrateSimpleField)
			}
			if field != nil {
				extra = []domain.Field{field}
			}
		}

		if _, err := hydrateRun(para, child, ctx, state, extra); err != nil {
			return errors.Wrap(err, opHydrateSimpleField)
		}
	}

	if state != nil {
		state.reset()
	}

	return nil
}

func hydrateDrawing(para domain.Paragraph, run domain.Run, elem *Element, ctx *reconstructContext) error {
	if para == nil || run == nil || elem == nil || ctx == nil {
		return nil
	}

	container := findChild(elem, "inline")
	floating := false
	if container == nil {
		container = findChild(elem, "anchor")
		floating = container != nil
	}
	if container == nil {
		return nil
	}

	relID := extractDrawingRelationshipID(container)
	if relID == "" {
		return nil
	}

	target, ok := ctx.resolveRelationshipTarget(relID)
	if !ok || target == "" {
		return errors.Errorf(errors.ErrCodeInvalidState, opHydrateDrawing, "relationship %s missing media target", relID)
	}

	mediaPart, mediaPath, found := ctx.mediaPartFor(target)
	if !found || mediaPart == nil {
		return errors.Errorf(errors.ErrCodeInvalidState, opHydrateDrawing, "unable to resolve media part for %s", mediaPath)
	}

	registerPath := mediaPart.Path
	if registerPath == "" {
		registerPath = mediaPath
	}

	img, err := core.NewImageFromPackage(registerPath, mediaPart.Data, mediaPart.ContentType)
	if err != nil {
		return errors.Wrap(err, opHydrateDrawing)
	}

	if setter, ok := img.(interface{ SetRelationshipID(string) }); ok {
		setter.SetRelationshipID(relID)
	}

	if desc := extractDrawingDescription(container); desc != "" {
		_ = img.SetDescription(desc)
	}

	if widthEMU, heightEMU := extractDrawingExtent(container); widthEMU > 0 && heightEMU > 0 {
		size := domain.ImageSize{
			WidthEMU:  widthEMU,
			HeightEMU: heightEMU,
			WidthPx:   emuToPixels(widthEMU),
			HeightPx:  emuToPixels(heightEMU),
		}
		if err := img.SetSize(size); err != nil {
			return errors.Wrap(err, opHydrateDrawing)
		}
	}

	if floating {
		position := buildFloatingPosition(container)
		if setter, ok := img.(interface {
			SetPosition(domain.ImagePosition) error
		}); ok {
			_ = setter.SetPosition(position)
		}
	}

	if attacher, ok := para.(interface {
		AttachHydratedImageToRun(domain.Run, domain.Image, string, string, []byte) error
	}); ok {
		if err := attacher.AttachHydratedImageToRun(run, img, registerPath, mediaPart.ContentType, mediaPart.Data); err != nil {
			return errors.Wrap(err, opHydrateDrawing)
		}
		return nil
	}

	if setter, ok := run.(interface{ setImage(domain.Image) }); ok {
		setter.setImage(img)
	}

	if registrar, ok := para.(interface {
		RegisterHydratedImage(domain.Image, string, string, []byte) error
	}); ok {
		if err := registrar.RegisterHydratedImage(img, registerPath, mediaPart.ContentType, mediaPart.Data); err != nil {
			return errors.Wrap(err, opHydrateDrawing)
		}
	}

	return nil
}

func extractDrawingRelationshipID(elem *Element) string {
	if elem == nil {
		return ""
	}

	graphic := findChild(elem, "graphic")
	if graphic == nil {
		return ""
	}

	data := findChild(graphic, "graphicData")
	if data == nil {
		return ""
	}

	pic := findChild(data, "pic")
	if pic == nil {
		return ""
	}

	blipFill := findChild(pic, "blipFill")
	if blipFill == nil {
		return ""
	}

	blip := findChild(blipFill, "blip")
	if blip == nil {
		return ""
	}

	if relID, ok := getAttr(blip, "embed"); ok {
		return relID
	}

	return ""
}

func extractDrawingExtent(elem *Element) (int, int) {
	if elem == nil {
		return 0, 0
	}

	if extent := findChild(elem, "extent"); extent != nil {
		width := attrToInt(extent, "cx")
		height := attrToInt(extent, "cy")
		if width > 0 && height > 0 {
			return width, height
		}
	}

	if ext := findDescendant(elem, "ext"); ext != nil {
		width := attrToInt(ext, "cx")
		height := attrToInt(ext, "cy")
		if width > 0 && height > 0 {
			return width, height
		}
	}

	return 0, 0
}

func extractDrawingDescription(elem *Element) string {
	if elem == nil {
		return ""
	}

	if docPr := findChild(elem, "docPr"); docPr != nil {
		if desc, ok := getAttr(docPr, "descr"); ok {
			return desc
		}
	}

	if cNvPr := findDescendant(elem, "cNvPr"); cNvPr != nil {
		if desc, ok := getAttr(cNvPr, "descr"); ok {
			return desc
		}
	}

	return ""
}

func buildFloatingPosition(elem *Element) domain.ImagePosition {
	pos := domain.DefaultImagePosition()
	pos.Type = domain.ImagePositionFloating

	if val, ok := getAttr(elem, "behindDoc"); ok {
		pos.BehindText = parseBoolAttr(val)
	}
	if val, ok := getAttr(elem, "relativeHeight"); ok {
		if n, err := strconv.Atoi(val); err == nil {
			pos.ZOrder = n
		}
	}

	if wrap := findWrapElement(elem); wrap != nil {
		if wrapText, ok := getAttr(wrap, "wrapText"); ok {
			if mapped, ok := mapWrapTextValue(wrapText); ok {
				pos.WrapText = mapped
			}
		}
	}

	if positionH := findChild(elem, "positionH"); positionH != nil {
		if align := findChild(positionH, "align"); align != nil {
			if mapped, ok := mapHorizontalAlignValue(strings.TrimSpace(align.Text)); ok {
				pos.HAlign = mapped
			}
		}
		if offset, ok := parseChildInt(positionH, "posOffset"); ok {
			pos.OffsetX = offset
		}
	}

	if positionV := findChild(elem, "positionV"); positionV != nil {
		if align := findChild(positionV, "align"); align != nil {
			if mapped, ok := mapVerticalAlignValue(strings.TrimSpace(align.Text)); ok {
				pos.VAlign = mapped
			}
		}
		if offset, ok := parseChildInt(positionV, "posOffset"); ok {
			pos.OffsetY = offset
		}
	}

	return pos
}

func findWrapElement(elem *Element) *Element {
	if elem == nil {
		return nil
	}

	candidates := []string{"wrapSquare", "wrapTight", "wrapThrough", "wrapTopAndBottom", "wrapNone"}
	for _, child := range elem.Children {
		if child == nil {
			continue
		}
		for _, name := range candidates {
			if child.Name.Local == name {
				return child
			}
		}
	}

	return nil
}

func mapWrapTextValue(value string) (domain.TextWrapType, bool) {
	switch strings.ToLower(value) {
	case "bothsides", "left", "right":
		return domain.WrapSquare, true
	case "tight":
		return domain.WrapTight, true
	case "through":
		return domain.WrapThrough, true
	case "topandbottom":
		return domain.WrapTopBottom, true
	case "none":
		return domain.WrapNone, true
	case "behind":
		return domain.WrapBehindText, true
	case "infront":
		return domain.WrapInFrontText, true
	default:
		return domain.WrapNone, false
	}
}

func mapHorizontalAlignValue(value string) (domain.HorizontalAlign, bool) {
	switch strings.ToLower(value) {
	case "left":
		return domain.HAlignLeft, true
	case "center":
		return domain.HAlignCenter, true
	case "right":
		return domain.HAlignRight, true
	case "inside":
		return domain.HAlignInside, true
	case "outside":
		return domain.HAlignOutside, true
	default:
		return domain.HAlignLeft, false
	}
}

func mapVerticalAlignValue(value string) (domain.VerticalAlign, bool) {
	switch strings.ToLower(value) {
	case "top":
		return domain.VAlignTop, true
	case "center":
		return domain.VAlignCenter, true
	case "bottom":
		return domain.VAlignBottom, true
	case "inside":
		return domain.VAlignInside, true
	case "outside":
		return domain.VAlignOutside, true
	default:
		return domain.VAlignTop, false
	}
}

func parseBoolAttr(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "on", "yes":
		return true
	default:
		return false
	}
}

func parseChildInt(elem *Element, local string) (int, bool) {
	if elem == nil {
		return 0, false
	}
	child := findChild(elem, local)
	if child == nil {
		return 0, false
	}
	text := strings.TrimSpace(child.Text)
	if text == "" {
		return 0, false
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return 0, false
	}
	return value, true
}

func attrToInt(elem *Element, name string) int {
	if elem == nil {
		return 0
	}
	if val, ok := getAttr(elem, name); ok && val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return 0
}

func emuToPixels(emu int) int {
	if emu <= 0 {
		return 0
	}
	const emuPerPixel = 9525
	return (emu + emuPerPixel/2) / emuPerPixel
}

func normalizeMediaPath(target string) string {
	path := strings.ReplaceAll(target, "\\", "/")
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "./")
	path = strings.TrimPrefix(path, "/")

	for strings.HasPrefix(path, "../") {
		path = strings.TrimPrefix(path, "../")
	}

	lower := strings.ToLower(path)
	if !strings.HasPrefix(lower, "word/") {
		path = "word/" + path
	}

	return path
}

func mapBreakType(elem *Element) domain.BreakType {
	if elem == nil {
		return domain.BreakTypeLine
	}

	if val, ok := getAttr(elem, "type"); ok {
		switch strings.ToLower(val) {
		case "page":
			return domain.BreakTypePage
		case "column":
			return domain.BreakTypeColumn
		case "textwrapping":
			return domain.BreakTypeLine
		}
	}

	return domain.BreakTypeLine
}

func newReconstructContext(doc domain.Document, parsed *ParsedPackage, defaultSection domain.Section) *reconstructContext {
	ctx := &reconstructContext{
		relationships:            make(map[string]*xmlstructs.Relationship),
		media:                    make(map[string]*MediaPart),
		doc:                      doc,
		parsed:                   parsed,
		currentSection:           defaultSection,
		hydratedHeaders:          make(map[domain.Section]map[domain.HeaderType]bool),
		hydratedFooters:          make(map[domain.Section]map[domain.FooterType]bool),
		suppressSectionHydration: 0,
	}

	if parsed != nil && parsed.DocumentRelationships != nil {
		for _, rel := range parsed.DocumentRelationships.Relationships {
			if rel == nil || rel.ID == "" {
				continue
			}
			ctx.relationships[rel.ID] = rel
		}
	}

	if parsed != nil && parsed.Package != nil {
		for path, part := range parsed.Package.Media {
			if part == nil || len(part.Data) == 0 {
				continue
			}
			normalized := normalizePartName(path)
			ctx.media[normalized] = part
		}
	}

	return ctx
}

func (ctx *reconstructContext) resolveRelationshipTarget(id string) (string, bool) {
	if ctx == nil || id == "" {
		return "", false
	}

	rel, ok := ctx.relationships[id]
	if !ok || rel == nil {
		return "", false
	}

	return rel.Target, true
}

func (ctx *reconstructContext) mediaPartFor(target string) (*MediaPart, string, bool) {
	if ctx == nil || target == "" {
		return nil, "", false
	}

	normalizedPath := normalizeMediaPath(target)
	part, ok := ctx.media[normalizePartName(normalizedPath)]
	if !ok || part == nil {
		return nil, normalizedPath, false
	}

	return part, normalizedPath, true
}

func newFieldState(ctx *reconstructContext) *fieldState {
	return &fieldState{ctx: ctx}
}

func (s *fieldState) reset() {
	if s == nil {
		return
	}

	s.active = false
	s.expectingResult = false
	s.pendingField = nil
	s.instruction.Reset()
}

func (s *fieldState) handleFieldChar(elem *Element) error {
	if s == nil || elem == nil {
		return nil
	}

	typ, _ := getAttr(elem, "fldCharType")
	switch strings.ToLower(typ) {
	case "begin":
		s.reset()
		s.active = true
	case "separate":
		if !s.active {
			return nil
		}
		field, err := buildFieldFromInstruction(s.instruction.String())
		if err != nil {
			return err
		}
		s.pendingField = field
		s.expectingResult = field != nil
		s.instruction.Reset()
	case "end":
		s.reset()
	}

	return nil
}

func (s *fieldState) appendInstruction(text string) {
	if s == nil || !s.active {
		return
	}

	s.instruction.WriteString(text)
}

func (s *fieldState) shouldForceRun() bool {
	return s != nil && s.expectingResult && s.pendingField != nil
}

func (s *fieldState) attachToRun(run domain.Run) error {
	if s == nil || run == nil || s.pendingField == nil || !s.expectingResult {
		return nil
	}

	if setter, ok := s.pendingField.(interface{ SetResult(string) }); ok {
		setter.SetResult(run.Text())
	}
	if accessor, ok := s.pendingField.(interface{ SetProperty(string, string) }); ok {
		accessor.SetProperty("display", run.Text())
	}
	if err := run.AddField(s.pendingField); err != nil {
		return errors.Wrap(err, opAttachFieldToRun)
	}

	s.expectingResult = false
	return nil
}

func buildFieldFromInstruction(instr string) (domain.Field, error) {
	trimmed := strings.TrimSpace(instr)
	if trimmed == "" {
		return nil, nil
	}

	upper := strings.ToUpper(trimmed)
	var field domain.Field

	switch {
	case strings.HasPrefix(upper, strings.ToUpper(constants.FieldCodePageNumber)):
		field = core.NewField(domain.FieldTypePageNumber)
	case strings.HasPrefix(upper, strings.ToUpper(constants.FieldCodeNumPages)):
		field = core.NewField(domain.FieldTypePageCount)
	case strings.HasPrefix(upper, strings.ToUpper(constants.FieldCodeTOC)):
		field = core.NewField(domain.FieldTypeTOC)
	case strings.HasPrefix(upper, strings.ToUpper(constants.FieldCodeDate)):
		field = core.NewField(domain.FieldTypeDate)
	case strings.HasPrefix(upper, strings.ToUpper(constants.FieldCodeTime)):
		field = core.NewField(domain.FieldTypeTime)
	case strings.HasPrefix(upper, strings.ToUpper(constants.FieldCodeStyleRef)):
		field = core.NewField(domain.FieldTypeStyleRef)
	case strings.HasPrefix(upper, strings.ToUpper(constants.FieldCodeSeq)):
		field = core.NewField(domain.FieldTypeSeq)
	case strings.HasPrefix(upper, strings.ToUpper(constants.FieldCodeRef)):
		field = core.NewField(domain.FieldTypeRef)
	case strings.HasPrefix(upper, "HYPERLINK"):
		field = core.NewField(domain.FieldTypeHyperlink)
		url, _ := parseHyperlinkInstruction(trimmed)
		if url == "" {
			return nil, nil
		}
		if accessor, ok := field.(interface{ SetProperty(string, string) }); ok {
			accessor.SetProperty("url", url)
		}
	default:
		field = core.NewField(domain.FieldTypeCustom)
	}

	if field == nil {
		return nil, nil
	}

	if err := field.SetCode(trimmed); err != nil {
		return nil, errors.Wrap(err, opBuildField)
	}

	return field, nil
}

func extractQuotedStrings(input string) []string {
	results := make([]string, 0, 2)
	inQuote := false
	start := 0

	for i, r := range input {
		switch r {
		case '"':
			if inQuote {
				results = append(results, input[start:i])
				inQuote = false
			} else {
				inQuote = true
				start = i + 1
			}
		}
	}

	return results
}

func parseHyperlinkInstruction(instr string) (string, bool) {
	quotes := extractQuotedStrings(instr)
	if len(quotes) == 0 {
		return "", false
	}

	url := quotes[0]
	isAnchor := false
	lower := strings.ToLower(instr)

	if strings.Contains(lower, "\\l") && !strings.Contains(strings.ToLower(url), "://") && !strings.HasPrefix(strings.ToLower(url), "mailto:") {
		if !strings.HasPrefix(url, "#") {
			url = "#" + url
		}
		isAnchor = true
	}

	return url, isAnchor
}

func (ctx *reconstructContext) ensureCurrentSection() (domain.Section, error) {
	if ctx == nil {
		return nil, nil
	}
	if ctx.currentSection != nil {
		return ctx.currentSection, nil
	}
	if ctx.doc == nil {
		return nil, nil
	}
	sec, err := ctx.doc.DefaultSection()
	if err != nil {
		return nil, err
	}
	ctx.currentSection = sec
	return sec, nil
}

func (ctx *reconstructContext) applySectionProperties(sectPr *Element) error {
	if ctx == nil || sectPr == nil || ctx.suppressSectionHydration > 0 {
		return nil
	}

	section, err := ctx.ensureCurrentSection()
	if err != nil {
		return errors.Wrap(err, opApplySectionProperties)
	}
	if section == nil {
		return nil
	}

	if err := ctx.applySectionLayout(section, sectPr); err != nil {
		return err
	}
	if err := ctx.applySectionHeaders(section, sectPr); err != nil {
		return err
	}
	if err := ctx.applySectionFooters(section, sectPr); err != nil {
		return err
	}

	if breakType, ok := extractSectionBreakType(sectPr); ok {
		if ctx.doc == nil {
			return nil
		}
		newSection, err := ctx.doc.AddSectionWithBreak(breakType)
		if err != nil {
			return errors.Wrap(err, opApplySectionProperties)
		}
		ctx.currentSection = newSection
	}

	return nil
}

func (ctx *reconstructContext) applySectionLayout(section domain.Section, sectPr *Element) error {
	if section == nil || sectPr == nil {
		return nil
	}

	if pgSz := findChild(sectPr, "pgSz"); pgSz != nil {
		width, hasWidth := parseIntAttr(pgSz, "w")
		height, hasHeight := parseIntAttr(pgSz, "h")

		if hasWidth && hasHeight {
			if err := section.SetPageSize(domain.PageSize{Width: width, Height: height}); err != nil {
				return errors.Wrap(err, opApplySectionProperties)
			}
		}

		if orientVal, ok := getAttr(pgSz, "orient"); ok && orientVal != "" {
			if orient, mapped := mapOrientation(orientVal); mapped {
				if err := section.SetOrientation(orient); err != nil {
					return errors.Wrap(err, opApplySectionProperties)
				}
			}
		} else if hasWidth && hasHeight {
			orient := domain.OrientationPortrait
			if width > height {
				orient = domain.OrientationLandscape
			}
			if err := section.SetOrientation(orient); err != nil {
				return errors.Wrap(err, opApplySectionProperties)
			}
		}
	}

	if pgMar := findChild(sectPr, "pgMar"); pgMar != nil {
		margins := section.Margins()
		if val, ok := parseIntAttr(pgMar, "top"); ok {
			margins.Top = val
		}
		if val, ok := parseIntAttr(pgMar, "right"); ok {
			margins.Right = val
		}
		if val, ok := parseIntAttr(pgMar, "bottom"); ok {
			margins.Bottom = val
		}
		if val, ok := parseIntAttr(pgMar, "left"); ok {
			margins.Left = val
		}
		if val, ok := parseIntAttr(pgMar, "header"); ok {
			margins.Header = val
		}
		if val, ok := parseIntAttr(pgMar, "footer"); ok {
			margins.Footer = val
		}

		if err := section.SetMargins(margins); err != nil {
			return errors.Wrap(err, opApplySectionProperties)
		}
	}

	if cols := findChild(sectPr, "cols"); cols != nil {
		if val, ok := parseIntAttr(cols, "num"); ok && val >= 1 {
			if err := section.SetColumns(val); err != nil {
				return errors.Wrap(err, opApplySectionProperties)
			}
		}
	}

	return nil
}

func (ctx *reconstructContext) applySectionHeaders(section domain.Section, sectPr *Element) error {
	if ctx == nil || section == nil || sectPr == nil {
		return nil
	}

	for _, child := range sectPr.Children {
		if child == nil || child.Name.Local != "headerReference" {
			continue
		}

		relID, ok := getAttr(child, "id")
		if !ok || relID == "" {
			continue
		}

		target, ok := ctx.resolveRelationshipTarget(relID)
		if !ok || target == "" {
			continue
		}

		headerTypeVal, _ := getAttr(child, "type")
		headerType := mapHeaderType(headerTypeVal)
		if !ctx.markHeaderHydrated(section, headerType) {
			continue
		}

		if err := ctx.hydrateHeader(section, headerType, relID, target); err != nil {
			return err
		}
	}

	return nil
}

func (ctx *reconstructContext) applySectionFooters(section domain.Section, sectPr *Element) error {
	if ctx == nil || section == nil || sectPr == nil {
		return nil
	}

	for _, child := range sectPr.Children {
		if child == nil || child.Name.Local != "footerReference" {
			continue
		}

		relID, ok := getAttr(child, "id")
		if !ok || relID == "" {
			continue
		}

		target, ok := ctx.resolveRelationshipTarget(relID)
		if !ok || target == "" {
			continue
		}

		footerTypeVal, _ := getAttr(child, "type")
		footerType := mapFooterType(footerTypeVal)
		if !ctx.markFooterHydrated(section, footerType) {
			continue
		}

		if err := ctx.hydrateFooter(section, footerType, relID, target); err != nil {
			return err
		}
	}

	return nil
}

func (ctx *reconstructContext) hydrateHeader(section domain.Section, headerType domain.HeaderType, relID, target string) error {
	if section == nil {
		return nil
	}

	header, err := section.Header(headerType)
	if err != nil {
		return errors.Wrap(err, opHydrateSectionHeader)
	}

	if setter, ok := header.(interface{ SetExistingRelationship(string, string) }); ok {
		setter.SetExistingRelationship(relID, target)
	}

	if ctx == nil || ctx.parsed == nil {
		return nil
	}

	tree := ctx.findPartTree(target, ctx.parsed.HeaderTrees)
	if tree == nil {
		return nil
	}

	return ctx.withSectionHydrationDisabled(func() error {
		for _, child := range tree.Children {
			if child == nil || child.Name.Local != "p" {
				continue
			}
			para, err := header.AddParagraph()
			if err != nil {
				return errors.Wrap(err, opHydrateSectionHeader)
			}
			if err := populateParagraph(para, child, ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

func (ctx *reconstructContext) hydrateFooter(section domain.Section, footerType domain.FooterType, relID, target string) error {
	if section == nil {
		return nil
	}

	footer, err := section.Footer(footerType)
	if err != nil {
		return errors.Wrap(err, opHydrateSectionFooter)
	}

	if setter, ok := footer.(interface{ SetExistingRelationship(string, string) }); ok {
		setter.SetExistingRelationship(relID, target)
	}

	if ctx == nil || ctx.parsed == nil {
		return nil
	}

	tree := ctx.findPartTree(target, ctx.parsed.FooterTrees)
	if tree == nil {
		return nil
	}

	return ctx.withSectionHydrationDisabled(func() error {
		for _, child := range tree.Children {
			if child == nil || child.Name.Local != "p" {
				continue
			}
			para, err := footer.AddParagraph()
			if err != nil {
				return errors.Wrap(err, opHydrateSectionFooter)
			}
			if err := populateParagraph(para, child, ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

func (ctx *reconstructContext) withSectionHydrationDisabled(fn func() error) error {
	if fn == nil {
		return nil
	}
	if ctx == nil {
		return fn()
	}
	ctx.suppressSectionHydration++
	defer func() { ctx.suppressSectionHydration-- }()
	return fn()
}

func (ctx *reconstructContext) markHeaderHydrated(section domain.Section, headerType domain.HeaderType) bool {
	if ctx.hydratedHeaders == nil {
		ctx.hydratedHeaders = make(map[domain.Section]map[domain.HeaderType]bool)
	}
	flags := ctx.hydratedHeaders[section]
	if flags == nil {
		flags = make(map[domain.HeaderType]bool)
		ctx.hydratedHeaders[section] = flags
	}
	if flags[headerType] {
		return false
	}
	flags[headerType] = true
	return true
}

func (ctx *reconstructContext) markFooterHydrated(section domain.Section, footerType domain.FooterType) bool {
	if ctx.hydratedFooters == nil {
		ctx.hydratedFooters = make(map[domain.Section]map[domain.FooterType]bool)
	}
	flags := ctx.hydratedFooters[section]
	if flags == nil {
		flags = make(map[domain.FooterType]bool)
		ctx.hydratedFooters[section] = flags
	}
	if flags[footerType] {
		return false
	}
	flags[footerType] = true
	return true
}

func (ctx *reconstructContext) findPartTree(target string, collection map[string]*Element) *Element {
	if target == "" || collection == nil {
		return nil
	}

	normalized := normalizePartName(normalizeMediaPath(target))
	for name, tree := range collection {
		if tree == nil {
			continue
		}
		if normalizePartName(name) == normalized {
			return tree
		}
	}

	return nil
}

func parseIntAttr(elem *Element, name string) (int, bool) {
	if elem == nil {
		return 0, false
	}
	val, ok := getAttr(elem, name)
	if !ok || val == "" {
		return 0, false
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, false
	}
	return n, true
}

func extractSectionBreakType(sectPr *Element) (domain.SectionBreakType, bool) {
	if sectPr == nil {
		return domain.SectionBreakTypeNextPage, false
	}
	typeElem := findChild(sectPr, "type")
	if typeElem == nil {
		return domain.SectionBreakTypeNextPage, false
	}
	val, ok := getAttr(typeElem, "val")
	if !ok || val == "" {
		return domain.SectionBreakTypeNextPage, false
	}
	return mapSectionBreakType(val)
}

func parseOnOff(elem *Element) (bool, bool) {
	if elem == nil {
		return false, false
	}

	if val, ok := getAttr(elem, "val"); ok {
		normalized := strings.ToLower(val)
		switch normalized {
		case "0", "false", "off":
			return false, true
		case "", "1", "true", "on":
			return true, true
		default:
			return true, true
		}
	}

	return true, true
}

func mapUnderlineStyle(value string) (domain.UnderlineStyle, bool) {
	switch strings.ToLower(value) {
	case constants.UnderlineValueNone:
		return domain.UnderlineNone, true
	case constants.UnderlineValueSingle:
		return domain.UnderlineSingle, true
	case constants.UnderlineValueDouble:
		return domain.UnderlineDouble, true
	case constants.UnderlineValueThick:
		return domain.UnderlineThick, true
	case constants.UnderlineValueDotted:
		return domain.UnderlineDotted, true
	case constants.UnderlineValueDashed:
		return domain.UnderlineDashed, true
	case constants.UnderlineValueWave:
		return domain.UnderlineWave, true
	default:
		return domain.UnderlineNone, false
	}
}

func mapHighlightColor(value string) (domain.HighlightColor, bool) {
	switch strings.ToLower(value) {
	case strings.ToLower(constants.HighlightValueNone):
		return domain.HighlightNone, true
	case strings.ToLower(constants.HighlightValueYellow):
		return domain.HighlightYellow, true
	case strings.ToLower(constants.HighlightValueGreen):
		return domain.HighlightGreen, true
	case strings.ToLower(constants.HighlightValueCyan):
		return domain.HighlightCyan, true
	case strings.ToLower(constants.HighlightValueMagenta):
		return domain.HighlightMagenta, true
	case strings.ToLower(constants.HighlightValueBlue):
		return domain.HighlightBlue, true
	case strings.ToLower(constants.HighlightValueRed):
		return domain.HighlightRed, true
	case strings.ToLower(constants.HighlightValueDarkBlue):
		return domain.HighlightDarkBlue, true
	case strings.ToLower(constants.HighlightValueDarkCyan):
		return domain.HighlightDarkCyan, true
	case strings.ToLower(constants.HighlightValueDarkGreen):
		return domain.HighlightDarkGreen, true
	case strings.ToLower(constants.HighlightValueDarkMagenta):
		return domain.HighlightDarkMagenta, true
	case strings.ToLower(constants.HighlightValueDarkRed):
		return domain.HighlightDarkRed, true
	case strings.ToLower(constants.HighlightValueDarkYellow):
		return domain.HighlightDarkYellow, true
	case strings.ToLower(constants.HighlightValueDarkGray):
		return domain.HighlightDarkGray, true
	case strings.ToLower(constants.HighlightValueLightGray):
		return domain.HighlightLightGray, true
	default:
		return domain.HighlightNone, false
	}
}

func findChild(parent *Element, local string) *Element {
	if parent == nil {
		return nil
	}
	for _, child := range parent.Children {
		if child == nil {
			continue
		}
		if child.Name.Local == local {
			return child
		}
	}
	return nil
}

func findDescendant(parent *Element, local string) *Element {
	if parent == nil {
		return nil
	}
	for _, child := range parent.Children {
		if child == nil {
			continue
		}
		if child.Name.Local == local {
			return child
		}
		if found := findDescendant(child, local); found != nil {
			return found
		}
	}
	return nil
}

func getAttr(elem *Element, local string) (string, bool) {
	if elem == nil {
		return "", false
	}
	for _, attr := range elem.Attr {
		if attr.Name.Local == local {
			return attr.Value, true
		}
	}
	return "", false
}

func mapLineSpacingRule(value string) domain.LineSpacingRule {
	switch value {
	case constants.LineSpacingRuleExact:
		return domain.LineSpacingExact
	case constants.LineSpacingRuleAtLeast:
		return domain.LineSpacingAtLeast
	default:
		return domain.LineSpacingAuto
	}
}

func mapAlignment(value string) (domain.Alignment, bool) {
	switch value {
	case constants.AlignmentValueLeft:
		return domain.AlignmentLeft, true
	case constants.AlignmentValueCenter:
		return domain.AlignmentCenter, true
	case constants.AlignmentValueRight:
		return domain.AlignmentRight, true
	case constants.AlignmentValueJustify:
		return domain.AlignmentJustify, true
	case constants.AlignmentValueDistribute:
		return domain.AlignmentDistribute, true
	default:
		return domain.AlignmentLeft, false
	}
}

func mapOrientation(value string) (domain.Orientation, bool) {
	switch strings.ToLower(value) {
	case "landscape":
		return domain.OrientationLandscape, true
	case "portrait":
		return domain.OrientationPortrait, true
	default:
		return domain.OrientationPortrait, false
	}
}

func mapSectionBreakType(value string) (domain.SectionBreakType, bool) {
	switch strings.ToLower(value) {
	case "nextpage":
		return domain.SectionBreakTypeNextPage, true
	case "continuous":
		return domain.SectionBreakTypeContinuous, true
	case "evenpage":
		return domain.SectionBreakTypeEvenPage, true
	case "oddpage":
		return domain.SectionBreakTypeOddPage, true
	default:
		return domain.SectionBreakTypeNextPage, false
	}
}

func mapHeaderType(value string) domain.HeaderType {
	switch strings.ToLower(value) {
	case "first":
		return domain.HeaderFirst
	case "even":
		return domain.HeaderEven
	default:
		return domain.HeaderDefault
	}
}

func mapFooterType(value string) domain.FooterType {
	switch strings.ToLower(value) {
	case "first":
		return domain.FooterFirst
	case "even":
		return domain.FooterEven
	default:
		return domain.FooterDefault
	}
}

func hydrateTable(doc domain.Document, elem *Element, ctx *reconstructContext) error {
	if doc == nil || elem == nil {
		return nil
	}

	rows := make([]*Element, 0, len(elem.Children))
	for _, child := range elem.Children {
		if child == nil || child.Name.Local != "tr" {
			continue
		}
		rows = append(rows, child)
	}

	if len(rows) == 0 {
		return nil
	}

	rowCells := make([][]*Element, len(rows))
	maxCols := 0
	for idx, row := range rows {
		cells := make([]*Element, 0, len(row.Children))
		for _, child := range row.Children {
			if child == nil || child.Name.Local != "tc" {
				continue
			}
			cells = append(cells, child)
		}
		if len(cells) > maxCols {
			maxCols = len(cells)
		}
		rowCells[idx] = cells
	}

	if maxCols == 0 {
		return nil
	}

	table, err := doc.AddTable(len(rows), maxCols)
	if err != nil {
		return errors.Wrap(err, opHydrateTable)
	}

	for i, cells := range rowCells {
		row, err := table.Row(i)
		if err != nil {
			return errors.Wrap(err, opHydrateTable)
		}

		for j, cellElem := range cells {
			if j >= table.ColumnCount() {
				continue
			}

			cell, err := row.Cell(j)
			if err != nil {
				return errors.Wrap(err, opHydrateTable)
			}

			if err := hydrateTableCell(cell, cellElem, ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

func hydrateTableCell(cell domain.TableCell, elem *Element, ctx *reconstructContext) error {
	if cell == nil || elem == nil {
		return nil
	}

	for _, child := range elem.Children {
		if child == nil || child.Name.Local != "p" {
			continue
		}

		para, err := cell.AddParagraph()
		if err != nil {
			return errors.Wrap(err, opHydrateTableCell)
		}

		if err := populateParagraph(para, child, ctx); err != nil {
			return err
		}
	}

	return nil
}

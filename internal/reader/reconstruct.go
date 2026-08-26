// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package reader

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/core"
	xmlstructs "github.com/mmonterroca/docxgo/v2/internal/xml"
	pkgcolor "github.com/mmonterroca/docxgo/v2/pkg/color"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
	"github.com/mmonterroca/docxgo/v2/pkg/errors"
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
	activeRelationships      map[string]*xmlstructs.Relationship
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

	body := findWordChild(parsed.DocumentTree, "body")
	if body == nil {
		return nil, errors.Errorf(errors.ErrCodeInvalidState, opReconstructDocument, "document body is missing")
	}

	doc := core.NewDocumentForReconstruction()
	defaultSection, err := doc.DefaultSection()
	if err != nil {
		return nil, errors.Wrap(err, opReconstructDocument)
	}

	ctx := newReconstructContext(doc, parsed, defaultSection)
	roundTripEntries := make([]core.RoundTripBodyEntrySource, 0, len(body.Children))
	mainNamespaces := documentNamespaces(parsed.DocumentTree, body)
	var mainDocumentXML []byte
	if parsed.Package != nil {
		mainDocumentXML = parsed.Package.MainDocument
	}
	rawCursor := body.ContentStartOffset
	if tracker, ok := doc.(interface{ ObserveHydratedBookmarkID(string) }); ok {
		observeSourceBookmarkIDs(parsed.DocumentTree, tracker)
	}
	maxDrawingID := highestSourceDrawingID(parsed.DocumentTree)

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

	// Preserve original styles.xml from the source document to ensure
	// custom styles are retained during round-trip (read-modify-write).
	if parsed.Package != nil && len(parsed.Package.Styles) > 0 {
		if setter, ok := doc.(interface{ SetPreservedStylesPart([]byte) }); ok {
			setter.SetPreservedStylesPart(parsed.Package.Styles)
		}
	}

	// Preserve all original parts for complete round-trip fidelity.
	if parsed.Package != nil {
		preserveOriginalParts(doc, parsed.Package)
	}

	// Hydrate metadata from docProps/core.xml so it survives round-trip.
	if parsed.Package != nil && len(parsed.Package.CoreProperties) > 0 {
		if meta, err := parseCoreProperties(parsed.Package.CoreProperties); err == nil && meta != nil {
			if err := doc.SetMetadata(meta); err != nil {
				return nil, errors.Wrap(err, opReconstructDocument)
			}
		}
	}

	// Hydrate the default proofing language from styles.xml's docDefaults so
	// Document.Language() reflects a document opened with one already set.
	// Uses SetLanguageRaw, not SetLanguage: preserveOriginalParts above has
	// already given this document preserved styles.xml/settings.xml bytes,
	// which would trip SetLanguage's round-trip guard.
	if parsed.Package != nil && len(parsed.Package.Styles) > 0 {
		if lang, err := parseStylesLanguage(parsed.Package.Styles); err == nil && lang != nil {
			if setter, ok := doc.(interface{ SetLanguageRaw(*domain.Language) }); ok {
				setter.SetLanguageRaw(lang)
			}
		}
	}

	for _, child := range body.Children {
		if child == nil {
			continue
		}
		if isWordElement(child, "sectPr") {
			continue
		}

		beforeBlocks := len(doc.Blocks())

		switch {
		case isWordElement(child, "p"):
			if err := hydrateParagraph(doc, child, ctx); err != nil {
				return nil, errors.Wrap(err, opReconstructDocument)
			}
		case isWordElement(child, "tbl"):
			if err := hydrateTable(doc, child, ctx); err != nil {
				return nil, errors.Wrap(err, opReconstructDocument)
			}
		}

		afterBlocks := doc.Blocks()
		entry := core.RoundTripBodyEntrySource{
			Prefix: rawRange(mainDocumentXML, rawCursor, child.StartOffset),
			Raw:    rawElementBytes(mainDocumentXML, child),
		}
		rawCursor = child.EndOffset
		if beforeBlocks < len(afterBlocks) {
			entry.Blocks = append([]domain.Block(nil), afterBlocks[beforeBlocks:]...)
		}
		if isWordElement(child, "tbl") && len(entry.Blocks) == 1 && entry.Blocks[0].Table != nil {
			entry.Table = roundTripTableSource(mainDocumentXML, child, entry.Blocks[0].Table, mainNamespaces)
		}
		if isWordElement(child, "p") {
			if props := findWordChild(child, "pPr"); props != nil {
				if sectPr := findWordChild(props, "sectPr"); sectPr != nil {
					entry.Section = roundTripSectionPropertiesSource(
						mainDocumentXML,
						sectPr,
						extendNamespaces(mainNamespaces, child, props, sectPr),
					)
				}
			}
		}
		roundTripEntries = append(roundTripEntries, entry)
	}

	if sectPr := findWordChild(body, "sectPr"); sectPr != nil {
		if err := ctx.applySectionProperties(sectPr, false); err != nil {
			return nil, errors.Wrap(err, opReconstructDocument)
		}
	}

	// Record what each header and footer looks like straight out of
	// hydration, so the writer can tell an untouched part (write its
	// preserved bytes back verbatim) from an edited one (regenerate it, or
	// the edit never reaches the saved file). Must run last: a header has to
	// be fully hydrated before it is snapshotted.
	if snapshotter, ok := doc.(interface{ SnapshotHeaderFooterParts() }); ok {
		snapshotter.SnapshotHeaderFooterParts()
	}

	if setter, ok := doc.(interface {
		SetRoundTripMainDocument(*core.RoundTripMainDocumentSource) error
	}); ok && len(mainDocumentXML) > 0 {
		mainXML := mainDocumentXML
		documentPrefix, documentSuffix := roundTripDocumentBoundaries(mainXML, body)
		source := &core.RoundTripMainDocumentSource{
			Original:        mainXML,
			Prefix:          documentPrefix,
			Suffix:          documentSuffix,
			MainNamespace:   parsed.DocumentTree.Name.Space,
			MaxDrawingID:    maxDrawingID,
			BodyStartOffset: body.StartOffset,
			BackgroundStart: -1,
			BackgroundEnd:   -1,
			Namespaces:      mainNamespaces,
			Entries:         roundTripEntries,
		}
		if background := findWordChild(parsed.DocumentTree, "background"); background != nil {
			source.BackgroundStart = background.StartOffset
			source.BackgroundEnd = background.EndOffset
		}
		if sectPr := findWordChild(body, "sectPr"); sectPr != nil {
			source.FinalSectionPrefix = rawRange(mainXML, rawCursor, sectPr.StartOffset)
			source.FinalSectionRaw = rawElementBytes(mainXML, sectPr)
			source.FinalSection = roundTripSectionPropertiesSource(
				mainXML,
				sectPr,
				extendNamespaces(mainNamespaces, sectPr),
			)
			source.BodyTail = rawRange(mainXML, sectPr.EndOffset, body.ContentEndOffset)
		} else {
			source.BodyTail = rawRange(mainXML, rawCursor, body.ContentEndOffset)
		}
		if err := setter.SetRoundTripMainDocument(source); err != nil {
			return nil, errors.Wrap(err, opReconstructDocument)
		}
	}

	return doc, nil
}

func documentNamespaces(elements ...*Element) map[string]string {
	result := make(map[string]string)
	for _, element := range elements {
		if element == nil {
			continue
		}
		for _, attr := range element.Attr {
			switch {
			case attr.Name.Space == "xmlns":
				result[attr.Name.Local] = attr.Value
			case attr.Name.Space == "" && attr.Name.Local == "xmlns":
				result[""] = attr.Value
			}
		}
	}
	return result
}

func extendNamespaces(base map[string]string, elements ...*Element) map[string]string {
	result := make(map[string]string, len(base))
	for prefix, namespace := range base {
		result[prefix] = namespace
	}
	for prefix, namespace := range documentNamespaces(elements...) {
		result[prefix] = namespace
	}
	return result
}

func roundTripSectionPropertiesSource(source []byte, elem *Element, namespaces map[string]string) *core.RoundTripSectionPropertiesSource {
	if elem == nil {
		return nil
	}
	return &core.RoundTripSectionPropertiesSource{
		Raw:        rawElementBytes(source, elem),
		Namespaces: namespaces,
	}
}

type bookmarkIDObserver interface {
	ObserveHydratedBookmarkID(string)
}

func observeSourceBookmarkIDs(elem *Element, observer bookmarkIDObserver) {
	if elem == nil || observer == nil {
		return
	}
	if isWordElement(elem, "bookmarkStart") {
		if id, ok := getWordAttr(elem, "id"); ok {
			observer.ObserveHydratedBookmarkID(id)
		}
	}
	for _, child := range elem.Children {
		observeSourceBookmarkIDs(child, observer)
	}
}

func highestSourceDrawingID(elem *Element) int {
	if elem == nil {
		return 0
	}
	maxID := 0
	if isWordprocessingDrawingElement(elem, "docPr") {
		if rawID, ok := getAttrQName(elem, "", "id"); ok {
			if id, err := strconv.Atoi(rawID); err == nil && id > maxID {
				maxID = id
			}
		}
	}
	for _, child := range elem.Children {
		if childMax := highestSourceDrawingID(child); childMax > maxID {
			maxID = childMax
		}
	}
	return maxID
}

func rawRange(source []byte, start, end int) []byte {
	if start < 0 || end < start || end > len(source) {
		return nil
	}
	out := make([]byte, end-start)
	copy(out, source[start:end])
	return out
}

func rawElementBytes(source []byte, elem *Element) []byte {
	if elem == nil {
		return nil
	}
	return rawRange(source, elem.StartOffset, elem.EndOffset)
}

func roundTripDocumentBoundaries(source []byte, body *Element) ([]byte, []byte) {
	if body == nil {
		return nil, nil
	}
	prefix := rawRange(source, 0, body.ContentStartOffset)
	suffix := rawRange(source, body.ContentEndOffset, len(source))
	open := rawRange(source, body.StartOffset, body.ContentStartOffset)
	slash := bytes.LastIndex(open, []byte("/>"))
	if slash < 0 || slash+2 != len(open) || body.ContentStartOffset != body.EndOffset {
		return prefix, suffix
	}

	nameStart := 1
	nameEnd := nameStart
	for nameEnd < len(open) && open[nameEnd] != ' ' && open[nameEnd] != '\t' && open[nameEnd] != '\r' && open[nameEnd] != '\n' && open[nameEnd] != '/' && open[nameEnd] != '>' {
		nameEnd++
	}
	if nameEnd == nameStart {
		return prefix, suffix
	}

	slashOffset := body.StartOffset + slash
	expandedPrefix := make([]byte, 0, len(prefix)-1)
	expandedPrefix = append(expandedPrefix, source[:slashOffset]...)
	expandedPrefix = append(expandedPrefix, source[slashOffset+1:body.ContentStartOffset]...)
	closeTag := append([]byte("</"), open[nameStart:nameEnd]...)
	closeTag = append(closeTag, '>')
	expandedSuffix := make([]byte, 0, len(closeTag)+len(source)-body.EndOffset)
	expandedSuffix = append(expandedSuffix, closeTag...)
	expandedSuffix = append(expandedSuffix, source[body.EndOffset:]...)
	return expandedPrefix, expandedSuffix
}

func roundTripTableSource(source []byte, elem *Element, table domain.Table, inheritedNamespaces map[string]string) *core.RoundTripTableSource {
	if elem == nil || table == nil {
		return nil
	}
	rowElems := make([]*Element, 0, table.RowCount())
	for _, child := range elem.Children {
		if isWordElement(child, "tr") {
			rowElems = append(rowElems, child)
		}
	}
	rows := table.Rows()
	if len(rowElems) == 0 || len(rowElems) != len(rows) {
		return nil
	}

	result := &core.RoundTripTableSource{
		Table:         table,
		Namespaces:    extendNamespaces(inheritedNamespaces, elem),
		Open:          rawRange(source, elem.StartOffset, elem.ContentStartOffset),
		ShellChildren: make([]core.RoundTripTableShellChildSource, 0),
		Suffix:        rawRange(source, rowElems[len(rowElems)-1].EndOffset, elem.EndOffset),
		Rows:          make([]core.RoundTripRowSource, 0, len(rows)),
	}
	shellCursor := elem.ContentStartOffset
	for _, child := range elem.Children {
		if child == nil || child.StartOffset >= rowElems[0].StartOffset {
			break
		}
		name := ""
		if isWordNamespace(child.Name.Space) {
			name = child.Name.Local
		}
		shellChild := core.RoundTripTableShellChildSource{
			Prefix: rawRange(source, shellCursor, child.StartOffset),
			Raw:    rawElementBytes(source, child),
			Name:   name,
		}
		result.ShellChildren = append(result.ShellChildren, shellChild)
		shellCursor = child.EndOffset
	}
	result.ShellTail = rawRange(source, shellCursor, rowElems[0].StartOffset)
	previousEnd := rowElems[0].StartOffset
	for i, row := range rows {
		result.Rows = append(result.Rows, core.RoundTripRowSource{
			Prefix: rawRange(source, previousEnd, rowElems[i].StartOffset),
			Raw:    rawElementBytes(source, rowElems[i]),
			Row:    row,
		})
		previousEnd = rowElems[i].EndOffset
	}
	return result
}

func hydrateParagraph(doc domain.Document, elem *Element, ctx *reconstructContext) error {
	// A paragraph whose only content is a mid-body section break -- pPr's
	// sole child is sectPr, no runs/hyperlinks/fields/anything else -- exists
	// in the source purely to carry that sectPr (ECMA-376 17.6.17: a sectPr
	// in a paragraph's pPr means "this paragraph ends a section"; Word always
	// puts one on its own dedicated, otherwise-empty paragraph). The writer
	// already synthesizes a paragraph like this for every SectionBreak block
	// (see DocumentSerializer.SerializeBody). Hydrating this source paragraph
	// as an ordinary domain.Paragraph *in addition to* the section break
	// would double it on the next write: the original (now content-less)
	// paragraph, immediately followed by the writer's own synthetic one.
	if sectPr, ok := bareSectionBreakSectPr(elem); ok {
		return ctx.applySectionProperties(sectPr, true)
	}

	para, err := doc.AddParagraph()
	if err != nil {
		return errors.Wrap(err, opHydrateParagraph)
	}

	return populateParagraph(para, elem, ctx)
}

// bareSectionBreakSectPr returns the sectPr and true if elem is a paragraph
// that exists solely to carry a mid-body section break: no children besides
// pPr, and pPr has no children besides sectPr.
func bareSectionBreakSectPr(elem *Element) (*Element, bool) {
	if elem == nil {
		return nil, false
	}

	var props *Element
	for _, child := range elem.Children {
		if child == nil {
			continue
		}
		if !isWordElement(child, "pPr") {
			return nil, false
		}
		props = child
	}
	if props == nil {
		return nil, false
	}

	var sectPr *Element
	for _, child := range props.Children {
		if child == nil {
			continue
		}
		if !isWordElement(child, "sectPr") {
			return nil, false
		}
		sectPr = child
	}

	if sectPr == nil {
		return nil, false
	}
	return sectPr, true
}

func populateParagraph(para domain.Paragraph, elem *Element, ctx *reconstructContext) error {
	if para == nil || elem == nil {
		return nil
	}

	if props := findWordChild(elem, "pPr"); props != nil {
		if err := applyParagraphProperties(para, props); err != nil {
			return err
		}
	}

	state := newFieldState(ctx)

	// Bookmarks are hydrated only when their w:bookmarkStart and
	// w:bookmarkEnd both fall in this same paragraph -- core.paragraph holds
	// one scalar (id, name) pair, so it cannot represent a bookmark spanning
	// several paragraphs or one hanging directly off w:body (where Word's own
	// _GoBack usually sits). Those are dropped, same as before this existed.
	// pendingBookmarks collects every start seen so far in this paragraph;
	// bookmarkHydrated fires on the first end that closes one of them, so a
	// paragraph with several (possibly nested) bookmarks keeps exactly one.
	//
	// A same-paragraph bookmark is further required to wrap the paragraph's
	// entire content, not just part of a run. core.paragraph re-serializes it
	// as w:bookmarkStart at the very start of the paragraph and w:bookmarkEnd
	// at the very end, so a bookmark that in the source wrapped only "target"
	// inside "prefix target suffix" would silently expand to cover the whole
	// paragraph on round-trip -- corrupting any REF field pointed at it.
	// Representing a partial-run bookmark correctly needs per-run position
	// tracking, which this single (id, name) pair can't hold; out of scope
	// here, so a partial bookmark is dropped instead of silently widened.
	// firstContentIdx/lastContentIdx locate the first and last content-bearing
	// child (a run, hyperlink, or simple field) in this paragraph up front, so
	// the position of each bookmarkStart/bookmarkEnd can be checked against
	// them below.
	firstContentIdx, lastContentIdx := -1, -1
	for i, child := range elem.Children {
		if child == nil {
			continue
		}
		switch {
		case isWordElement(child, "r"), isWordElement(child, "hyperlink"), isWordElement(child, "fldSimple"):
			if firstContentIdx == -1 {
				firstContentIdx = i
			}
			lastContentIdx = i
		}
	}

	pendingBookmarks := make(map[string]string)
	bookmarkHydrated := false

	for i, child := range elem.Children {
		if child == nil {
			continue
		}

		switch {
		case isWordElement(child, "r"):
			if err := hydrateRun(para, child, ctx, state, nil); err != nil {
				return err
			}
		case isWordElement(child, "hyperlink"):
			if err := hydrateHyperlink(para, child, ctx, state); err != nil {
				return err
			}
		case isWordElement(child, "fldSimple"):
			if err := hydrateSimpleField(para, child, ctx, state); err != nil {
				return err
			}
		case isWordElement(child, "bookmarkStart"):
			id, ok := getWordAttr(child, "id")
			if !ok || id == "" {
				continue
			}
			if firstContentIdx != -1 && i > firstContentIdx {
				// Content already appeared before this start: partial.
				continue
			}
			name, _ := getWordAttr(child, "name")
			pendingBookmarks[id] = name
		case isWordElement(child, "bookmarkEnd"):
			if bookmarkHydrated {
				continue
			}
			id, ok := getWordAttr(child, "id")
			if !ok {
				continue
			}
			name, started := pendingBookmarks[id]
			if !started {
				continue
			}
			if lastContentIdx != -1 && i < lastContentIdx {
				// More content follows this end: partial.
				continue
			}
			if hydrator, ok := para.(interface{ HydrateBookmark(string, string) }); ok {
				hydrator.HydrateBookmark(id, name)
				bookmarkHydrated = true
				if ctx != nil {
					if tracker, ok := ctx.doc.(interface{ ObserveHydratedBookmarkID(string) }); ok {
						tracker.ObserveHydratedBookmarkID(id)
					}
				}
			}
		}
	}

	state.reset()

	if ctx != nil {
		if props := findWordChild(elem, "pPr"); props != nil {
			if sectPr := findWordChild(props, "sectPr"); sectPr != nil {
				if err := ctx.applySectionProperties(sectPr, true); err != nil {
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
	pStyleElem := findWordChild(props, "pStyle")
	if pStyleElem == nil {
		return nil
	}

	if styleID, ok := getWordAttr(pStyleElem, "val"); ok && styleID != "" {
		if err := para.SetStyle(styleID); err != nil {
			return errors.WrapWithContext(err, "applyParagraphStyle", map[string]interface{}{"styleID": styleID})
		}
	}

	return nil
}

func applyParagraphSpacing(para domain.Paragraph, props *Element) error {
	spacingElem := findWordChild(props, "spacing")
	if spacingElem == nil {
		return nil
	}

	if val, ok := getWordAttr(spacingElem, "before"); ok && val != "" {
		twips, err := strconv.Atoi(val)
		if err != nil {
			return errors.WrapWithContext(err, opApplyParagraphSpacing, map[string]interface{}{"attr": "before", "value": val})
		}
		if err := para.SetSpacingBefore(twips); err != nil {
			return err
		}
	}

	if val, ok := getWordAttr(spacingElem, "after"); ok && val != "" {
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

	if val, ok := getWordAttr(spacingElem, "line"); ok && val != "" {
		twips, err := strconv.Atoi(val)
		if err != nil {
			return errors.WrapWithContext(err, opApplyParagraphSpacing, map[string]interface{}{"attr": "line", "value": val})
		}
		lineSpacing.Value = twips
		valueChanged = true
	}

	if val, ok := getWordAttr(spacingElem, "lineRule"); ok && val != "" {
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
	jc := findWordChild(props, "jc")
	if jc == nil {
		return nil
	}

	if val, ok := getWordAttr(jc, "val"); ok && val != "" {
		if align, ok := mapAlignment(val); ok {
			if err := para.SetAlignment(align); err != nil {
				return errors.Wrap(err, opApplyParagraphAlignment)
			}
		}
	}

	return nil
}

// applyParagraphIndentation applies each attribute present in a source
// <w:ind> element via its own SetIndentLeft/Right/FirstLine/Hanging call,
// one per attribute actually present. A single merged SetIndent call would
// mark all four sides as explicitly set (see SetIndent's doc comment) even
// for the three this element never mentioned, and on re-serialization that
// would emit e.g. right="0" for a paragraph whose source <w:ind> only ever
// specified left, clobbering a style's own right indentation on a side this
// element never touched.
func applyParagraphIndentation(para domain.Paragraph, props *Element) error {
	ind := findWordChild(props, "ind")
	if ind == nil {
		return nil
	}

	if val, ok := getWordAttr(ind, "left"); ok && val != "" {
		twips, err := strconv.Atoi(val)
		if err != nil {
			return errors.WrapWithContext(err, opApplyParagraphIndent, map[string]interface{}{"attr": "left", "value": val})
		}
		if err := para.SetIndentLeft(twips); err != nil {
			return errors.Wrap(err, opApplyParagraphIndent)
		}
	}

	if val, ok := getWordAttr(ind, "right"); ok && val != "" {
		twips, err := strconv.Atoi(val)
		if err != nil {
			return errors.WrapWithContext(err, opApplyParagraphIndent, map[string]interface{}{"attr": "right", "value": val})
		}
		if err := para.SetIndentRight(twips); err != nil {
			return errors.Wrap(err, opApplyParagraphIndent)
		}
	}

	if val, ok := getWordAttr(ind, "firstLine"); ok && val != "" {
		twips, err := strconv.Atoi(val)
		if err != nil {
			return errors.WrapWithContext(err, opApplyParagraphIndent, map[string]interface{}{"attr": "firstLine", "value": val})
		}
		if err := para.SetIndentFirstLine(twips); err != nil {
			return errors.Wrap(err, opApplyParagraphIndent)
		}
	}

	if val, ok := getWordAttr(ind, "hanging"); ok && val != "" {
		twips, err := strconv.Atoi(val)
		if err != nil {
			return errors.WrapWithContext(err, opApplyParagraphIndent, map[string]interface{}{"attr": "hanging", "value": val})
		}
		if err := para.SetIndentHanging(twips); err != nil {
			return errors.Wrap(err, opApplyParagraphIndent)
		}
	}

	return nil
}

func applyParagraphNumbering(para domain.Paragraph, props *Element) error {
	if para == nil || props == nil {
		return nil
	}

	numPr := findWordChild(props, "numPr")
	if numPr == nil {
		para.ClearNumbering()
		return nil
	}

	ref := domain.NumberingReference{}
	foundID := false

	if numID := findWordChild(numPr, "numId"); numID != nil {
		if val, ok := getWordAttr(numID, "val"); ok && val != "" {
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

	if ilvl := findWordChild(numPr, "ilvl"); ilvl != nil {
		if val, ok := getWordAttr(ilvl, "val"); ok && val != "" {
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

func hydrateRun(para domain.Paragraph, elem *Element, ctx *reconstructContext, state *fieldState, extraFields []domain.Field) error {
	if para == nil || elem == nil {
		return nil
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

		switch {
		case isWordElement(child, "t"):
			textBuilder.WriteString(child.Text)
		case isWordElement(child, "tab"):
			textBuilder.WriteRune('\t')
		case isWordElement(child, "br"):
			breaks = append(breaks, mapBreakType(child))
		case isWordElement(child, "fldChar"):
			if state != nil {
				if err := state.handleFieldChar(child); err != nil {
					return err
				}
			}
		case isWordElement(child, "instrText"):
			if state != nil {
				state.appendInstruction(child.Text)
			}
		case isWordElement(child, "rPr"):
			props = child
		case isWordElement(child, "drawing"):
			drawings = append(drawings, child)
		}
	}

	createRun := textBuilder.Len() > 0 || len(breaks) > 0 || len(extraFields) > 0 || len(drawings) > 0
	if !createRun && state != nil && state.shouldForceRun() {
		createRun = true
	}

	if !createRun {
		return nil
	}

	run, err := para.AddRun()
	if err != nil {
		return errors.Wrap(err, opHydrateRun)
	}

	if textBuilder.Len() > 0 {
		if err := run.SetText(textBuilder.String()); err != nil {
			return errors.Wrap(err, opHydrateRun)
		}
	}

	if props != nil {
		if err := applyRunProperties(run, props); err != nil {
			return err
		}
	}

	for _, br := range breaks {
		if err := run.AddBreak(br); err != nil {
			return errors.Wrap(err, opHydrateRun)
		}
	}

	if state != nil {
		if err := state.attachToRun(run); err != nil {
			return err
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
			return errors.Wrap(err, opHydrateRun)
		}
	}

	if len(drawings) > 0 {
		for _, drawing := range drawings {
			if err := hydrateDrawing(para, run, drawing, ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

func applyRunProperties(run domain.Run, props *Element) error {
	if run == nil || props == nil {
		return nil
	}

	if boldElem := findWordChild(props, "b"); boldElem != nil {
		if val, ok := parseOnOff(boldElem); ok {
			if err := run.SetBold(val); err != nil {
				return errors.Wrap(err, opApplyRunProperties)
			}
		}
	}

	if italicElem := findWordChild(props, "i"); italicElem != nil {
		if val, ok := parseOnOff(italicElem); ok {
			if err := run.SetItalic(val); err != nil {
				return errors.Wrap(err, opApplyRunProperties)
			}
		}
	}

	if capsElem := findWordChild(props, "caps"); capsElem != nil {
		if val, ok := parseOnOff(capsElem); ok {
			if err := run.SetCaps(val); err != nil {
				return errors.Wrap(err, opApplyRunProperties)
			}
		}
	}

	if strikeElem := findWordChild(props, "strike"); strikeElem != nil {
		if val, ok := parseOnOff(strikeElem); ok {
			if err := run.SetStrike(val); err != nil {
				return errors.Wrap(err, opApplyRunProperties)
			}
		}
	}

	if underlineElem := findWordChild(props, "u"); underlineElem != nil {
		underlineVal, ok := getWordAttr(underlineElem, "val")
		if !ok || underlineVal == "" {
			underlineVal = constants.UnderlineValueSingle
		}
		if style, mapped := mapUnderlineStyle(underlineVal); mapped && style != domain.UnderlineNone {
			if err := run.SetUnderline(style); err != nil {
				return errors.Wrap(err, opApplyRunProperties)
			}
		}
	}

	if colorElem := findWordChild(props, "color"); colorElem != nil {
		if val, ok := getWordAttr(colorElem, "val"); ok && val != "" && !strings.EqualFold(val, "auto") {
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
	if szElem := findWordChild(props, "sz"); szElem != nil {
		if val, ok := getWordAttr(szElem, "val"); ok && val != "" {
			sizeVal = val
		}
	}
	if sizeVal == "" {
		if szCsElem := findWordChild(props, "szCs"); szCsElem != nil {
			if val, ok := getWordAttr(szCsElem, "val"); ok && val != "" {
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

	if fontElem := findWordChild(props, "rFonts"); fontElem != nil {
		current := run.Font()
		updated := current
		changed := false

		if val, ok := getWordAttr(fontElem, "ascii"); ok && val != "" {
			updated.Name = val
			changed = true
		} else if val, ok := getWordAttr(fontElem, "hAnsi"); ok && val != "" {
			updated.Name = val
			changed = true
		}

		if val, ok := getWordAttr(fontElem, "eastAsia"); ok && val != "" {
			updated.EastAsia = val
			changed = true
		}

		if val, ok := getWordAttr(fontElem, "cs"); ok && val != "" {
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

	if highlightElem := findWordChild(props, "highlight"); highlightElem != nil {
		if val, ok := getWordAttr(highlightElem, "val"); ok && val != "" {
			if highlight, mapped := mapHighlightColor(val); mapped && highlight != domain.HighlightNone {
				if err := run.SetHighlight(highlight); err != nil {
					return errors.Wrap(err, opApplyRunProperties)
				}
			}
		}
	}

	if langElem := findWordChild(props, "lang"); langElem != nil {
		val, _ := getWordAttr(langElem, "val")
		eastAsia, _ := getWordAttr(langElem, "eastAsia")
		bidi, _ := getWordAttr(langElem, "bidi")
		if val != "" || eastAsia != "" || bidi != "" {
			if err := run.SetLanguage(&domain.Language{Val: val, EastAsia: eastAsia, Bidi: bidi}); err != nil {
				return errors.Wrap(err, opApplyRunProperties)
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
	originalRelID := "" // Preserve the original relationship ID for external hyperlinks
	if relID, ok := getRelationshipAttr(elem, "id"); ok {
		if target, found := ctx.resolveRelationshipTarget(relID); found {
			url = target
			originalRelID = relID // Store the original relationship ID
		}
	}

	if url == "" {
		if anchor, ok := getWordAttr(elem, "anchor"); ok && anchor != "" {
			if strings.HasPrefix(anchor, "#") {
				url = anchor
			} else {
				url = "#" + anchor
			}
		}
	}

	// w:history is ST_OnOff: "0" is a real, distinct value from the
	// attribute being absent, so track presence rather than defaulting.
	history, hasHistory := getWordAttr(elem, "history")

	for _, child := range elem.Children {
		if !isWordElement(child, "r") {
			continue
		}

		var extraFields []domain.Field
		if url != "" {
			field := core.NewField(domain.FieldTypeHyperlink)
			// SetCodeRaw, not SetCode: url comes from the source file's own
			// relationship target or w:anchor attribute, which is the
			// ground truth here, not this package's own validation — same
			// reasoning as buildFieldFromInstruction's SetCodeRaw call.
			if setter, ok := field.(interface{ SetCodeRaw(string) }); ok {
				setter.SetCodeRaw(fmt.Sprintf(`HYPERLINK "%s"`, url))
			} else if err := field.SetCode(fmt.Sprintf(`HYPERLINK "%s"`, url)); err != nil {
				return errors.Wrap(err, opHydrateHyperlink)
			}
			if accessor, ok := field.(interface{ SetProperty(string, string) }); ok {
				accessor.SetProperty("url", url)
				// Preserve the original relationship ID for external hyperlinks
				// so the serializer can reuse it instead of generating new IDs
				if originalRelID != "" && !strings.HasPrefix(url, "#") {
					accessor.SetProperty("relationshipID", originalRelID)
				}
				if hasHistory {
					accessor.SetProperty("history", history)
				}
			}
			extraFields = []domain.Field{field}
		}

		if err := hydrateRun(para, child, ctx, state, extraFields); err != nil {
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

	instr, _ := getWordAttr(elem, "instr")

	for _, child := range elem.Children {
		if !isWordElement(child, "r") {
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

		if err := hydrateRun(para, child, ctx, state, extra); err != nil {
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

	container := findWordprocessingDrawingChild(elem, "inline")
	floating := false
	if container == nil {
		container = findWordprocessingDrawingChild(elem, "anchor")
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

	graphic := findDrawingChild(elem, "graphic")
	if graphic == nil {
		return ""
	}

	data := findDrawingChild(graphic, "graphicData")
	if data == nil {
		return ""
	}

	pic := findPictureChild(data, "pic")
	if pic == nil {
		return ""
	}

	blipFill := findPictureChild(pic, "blipFill")
	if blipFill == nil {
		return ""
	}

	blip := findDrawingChild(blipFill, "blip")
	if blip == nil {
		return ""
	}

	if relID, ok := getRelationshipAttr(blip, "embed"); ok {
		return relID
	}

	return ""
}

func extractDrawingExtent(elem *Element) (int, int) {
	if elem == nil {
		return 0, 0
	}

	if extent := findWordprocessingDrawingChild(elem, "extent"); extent != nil {
		width := attrToInt(extent, "cx")
		height := attrToInt(extent, "cy")
		if width > 0 && height > 0 {
			return width, height
		}
	}

	if ext := findDrawingDescendant(elem, "ext"); ext != nil {
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

	if docPr := findWordprocessingDrawingChild(elem, "docPr"); docPr != nil {
		if desc, ok := getUnqualifiedAttr(docPr, "descr"); ok {
			return desc
		}
	}

	if cNvPr := findPictureDescendant(elem, "cNvPr"); cNvPr != nil {
		if desc, ok := getUnqualifiedAttr(cNvPr, "descr"); ok {
			return desc
		}
	}

	return ""
}

func buildFloatingPosition(elem *Element) domain.ImagePosition {
	pos := domain.DefaultImagePosition()
	pos.Type = domain.ImagePositionFloating

	if val, ok := getUnqualifiedAttr(elem, "behindDoc"); ok {
		pos.BehindText = parseBoolAttr(val)
	}
	if val, ok := getUnqualifiedAttr(elem, "relativeHeight"); ok {
		if n, err := strconv.Atoi(val); err == nil {
			pos.ZOrder = n
		}
	}

	if wrap := findWrapElement(elem); wrap != nil {
		if wrapText, ok := getUnqualifiedAttr(wrap, "wrapText"); ok {
			if mapped, ok := mapWrapTextValue(wrapText); ok {
				pos.WrapText = mapped
			}
		}
	}

	if positionH := findWordprocessingDrawingChild(elem, "positionH"); positionH != nil {
		if align := findWordprocessingDrawingChild(positionH, "align"); align != nil {
			if mapped, ok := mapHorizontalAlignValue(strings.TrimSpace(align.Text)); ok {
				pos.HAlign = mapped
			}
		}
		if offset, ok := parseChildInt(positionH, "posOffset"); ok {
			pos.OffsetX = offset
			pos.UseOffsetX = true // Mark that we should use offset, even if 0
		}
	}

	if positionV := findWordprocessingDrawingChild(elem, "positionV"); positionV != nil {
		if align := findWordprocessingDrawingChild(positionV, "align"); align != nil {
			if mapped, ok := mapVerticalAlignValue(strings.TrimSpace(align.Text)); ok {
				pos.VAlign = mapped
			}
		}
		if offset, ok := parseChildInt(positionV, "posOffset"); ok {
			pos.OffsetY = offset
			pos.UseOffsetY = true // Mark that we should use offset, even if 0
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
			if isWordprocessingDrawingElement(child, name) {
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
	child := findWordprocessingDrawingChild(elem, local)
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
	if val, ok := getUnqualifiedAttr(elem, name); ok && val != "" {
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

	if val, ok := getWordAttr(elem, "type"); ok {
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

	if ctx.activeRelationships != nil {
		if rel, ok := ctx.activeRelationships[id]; ok && rel != nil {
			return rel.Target, true
		}
	}

	rel, ok := ctx.relationships[id]
	if !ok || rel == nil {
		return "", false
	}

	return rel.Target, true
}

// partRelationshipMap builds an ID-keyed relationship map for the header/footer
// part whose archive path matches target, looked up in ctx.parsed.PartRelationships.
// Relationship IDs are scoped per-part in OOXML, so this must be used instead of
// the document-wide ctx.relationships when hydrating header/footer content.
func (ctx *reconstructContext) partRelationshipMap(target string) map[string]*xmlstructs.Relationship {
	if ctx == nil || ctx.parsed == nil || ctx.parsed.PartRelationships == nil {
		return nil
	}

	want := normalizePartName(normalizeMediaPath(target))
	for name, set := range ctx.parsed.PartRelationships {
		if set == nil || normalizePartName(name) != want {
			continue
		}
		out := make(map[string]*xmlstructs.Relationship, len(set.Relationships))
		for _, rel := range set.Relationships {
			if rel == nil || rel.ID == "" {
				continue
			}
			out[rel.ID] = rel
		}
		return out
	}

	return nil
}

// withPartRelationships temporarily scopes relationship-ID resolution to
// rels while fn runs, restoring the previous scope afterward.
func (ctx *reconstructContext) withPartRelationships(rels map[string]*xmlstructs.Relationship, fn func() error) error {
	if fn == nil {
		return nil
	}
	if ctx == nil {
		return fn()
	}

	prev := ctx.activeRelationships
	ctx.activeRelationships = rels
	defer func() { ctx.activeRelationships = prev }()
	return fn()
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

	typ, _ := getWordAttr(elem, "fldCharType")
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

	// SetCodeRaw, not SetCode: trimmed comes from the source file's own
	// <w:instrText>, which is the ground truth for what a field instruction
	// looks like, not this package's own validation. Bypassing SetCode's
	// control-character guard here means OpenDocument can't fail on a
	// perfectly normal .docx merely because some producer emitted a byte
	// SetCode wouldn't accept from a caller building a field from scratch.
	if setter, ok := field.(interface{ SetCodeRaw(string) }); ok {
		setter.SetCodeRaw(trimmed)
	} else if err := field.SetCode(trimmed); err != nil {
		return nil, errors.Wrap(err, opBuildField)
	}

	return field, nil
}

func extractQuotedStrings(input string) []string {
	results := make([]string, 0, 2)
	inQuote := false
	start := 0

	for i, r := range input {
		if r == '"' {
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

// applySectionProperties applies a <w:sectPr>'s layout, headers, and footers
// to the current section.
//
// embedded distinguishes the two places a <w:sectPr> can appear: as a
// paragraph's own pPr child, versus as the body's own last child. Per OOXML
// (ECMA-376 §17.6.17), a sectPr embedded in a paragraph's pPr always means
// "this paragraph ends a section" -- a new section always follows, even when
// the optional w:type child is absent (its schema default is "nextPage", not
// "no break"). The body-level sectPr describes the document's own last
// section and never starts another one, since nothing follows it.
func (ctx *reconstructContext) applySectionProperties(sectPr *Element, embedded bool) error {
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

	if embedded {
		if ctx.doc == nil {
			return nil
		}
		breakType, _ := extractSectionBreakType(sectPr)
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

	if pgSz := findWordChild(sectPr, "pgSz"); pgSz != nil {
		width, hasWidth := parseIntAttr(pgSz, "w")
		height, hasHeight := parseIntAttr(pgSz, "h")

		if hasWidth && hasHeight {
			if err := section.SetPageSize(domain.PageSize{Width: width, Height: height}); err != nil {
				return errors.Wrap(err, opApplySectionProperties)
			}
		}

		if orientVal, ok := getWordAttr(pgSz, "orient"); ok && orientVal != "" {
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

	if pgMar := findWordChild(sectPr, "pgMar"); pgMar != nil {
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

	if cols := findWordChild(sectPr, "cols"); cols != nil {
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
		if !isWordElement(child, "headerReference") {
			continue
		}

		relID, ok := getRelationshipAttr(child, "id")
		if !ok || relID == "" {
			continue
		}

		target, ok := ctx.resolveRelationshipTarget(relID)
		if !ok || target == "" {
			continue
		}

		headerTypeVal, _ := getWordAttr(child, "type")
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
		if !isWordElement(child, "footerReference") {
			continue
		}

		relID, ok := getRelationshipAttr(child, "id")
		if !ok || relID == "" {
			continue
		}

		target, ok := ctx.resolveRelationshipTarget(relID)
		if !ok || target == "" {
			continue
		}

		footerTypeVal, _ := getWordAttr(child, "type")
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

	return ctx.hydratePartBlocks(header, tree, target, opHydrateSectionHeader)
}

// partBlockContainer is satisfied by both domain.Header and domain.Footer,
// letting hydratePartBlocks drive either from the same code path.
type partBlockContainer interface {
	AddParagraph() (domain.Paragraph, error)
	AddTable(rows, cols int) (domain.Table, error)
}

// tableRollbacker is implemented by the concrete header/footer types in
// internal/core, exposing the undo half of AddTable for hydrateTable's
// best-effort error path -- see the "tbl" case in hydratePartBlocks.
type tableRollbacker interface {
	RemoveLastTable()
}

// hydratePartBlocks copies the paragraph and table children of a
// header/footer part tree into container, resolving relationship IDs
// against that part's own relationships (per-part scoping) with section
// hydration suppressed.
//
// A table that fails to hydrate (e.g. a malformed or oversized grid) is
// skipped rather than propagated, unlike the body path — an error here
// would fail OpenDocument entirely for a document whose header/footer table
// docxgo simply can't represent, and TestOpenDocument_MalformedHeaderRelsIsTolerated
// already establishes that header/footer parts are tolerant of malformed
// content in a way the body isn't. Do not "fix" this to match hydrateTable's
// body-path error propagation.
func (ctx *reconstructContext) hydratePartBlocks(container partBlockContainer, tree *Element, target, op string) error {
	partRels := ctx.partRelationshipMap(target)
	return ctx.withSectionHydrationDisabled(func() error {
		return ctx.withPartRelationships(partRels, func() error {
			for _, child := range tree.Children {
				if child == nil {
					continue
				}
				switch {
				case isWordElement(child, "p"):
					para, err := container.AddParagraph()
					if err != nil {
						return errors.Wrap(err, op)
					}
					if err := populateParagraph(para, child, ctx); err != nil {
						return err
					}
				case isWordElement(child, "tbl"):
					// Best-effort: skip a table docxgo can't hydrate rather
					// than failing the whole document (see doc comment above).
					// hydrateTable's own AddTable call already attached the
					// table to container before any row/cell hydration ran,
					// so a failure partway through must be rolled back too --
					// otherwise a table that only *partly* failed would stay
					// visible in Tables()/Blocks() instead of being skipped
					// as this comment promises. See tableRollbacker.
					if err := hydrateTable(container, child, ctx); err != nil {
						if remover, ok := container.(tableRollbacker); ok {
							remover.RemoveLastTable()
						}
					}
				}
			}
			return nil
		})
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

	return ctx.hydratePartBlocks(footer, tree, target, opHydrateSectionFooter)
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
	val, ok := getWordAttr(elem, name)
	if !ok || val == "" {
		return 0, false
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, false
	}
	return n, true
}

// parseMeasurementInt accepts both the integer spelling required by OOXML
// and the float-formatted measurements emitted by some real-world producers
// (for example "9360.0"). Measurements in the domain model are integers, so
// a fractional value is truncated toward zero. Non-finite and out-of-range
// values are rejected instead of relying on implementation-defined float to
// int conversion behavior.
func parseMeasurementInt(text string) (int, error) {
	if n, err := strconv.Atoi(text); err == nil {
		return n, nil
	}

	f, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, fmt.Errorf("measurement is not finite: %q", text)
	}

	// int's positive limit is 2^(bits-1)-1 and its negative limit is
	// -2^(bits-1). Checking against the power-of-two boundary avoids the
	// float64 rounding problem around MaxInt64.
	limit := math.Ldexp(1, strconv.IntSize-1)
	if f >= limit || f < -limit {
		return 0, fmt.Errorf("measurement is outside int range: %q", text)
	}

	return int(math.Trunc(f)), nil
}

func extractSectionBreakType(sectPr *Element) (domain.SectionBreakType, bool) {
	if sectPr == nil {
		return domain.SectionBreakTypeNextPage, false
	}
	typeElem := findWordChild(sectPr, "type")
	if typeElem == nil {
		return domain.SectionBreakTypeNextPage, false
	}
	val, ok := getWordAttr(typeElem, "val")
	if !ok || val == "" {
		return domain.SectionBreakTypeNextPage, false
	}
	return mapSectionBreakType(val)
}

func parseOnOff(elem *Element) (bool, bool) {
	if elem == nil {
		return false, false
	}

	if val, ok := getWordAttr(elem, "val"); ok {
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

func isWordNamespace(namespace string) bool {
	return namespace == constants.NamespaceMain || namespace == constants.NamespaceMainStrict
}

func isRelationshipNamespace(namespace string) bool {
	return namespace == constants.NamespaceRelationships || namespace == constants.NamespaceRelationshipsStrict
}

func isWordprocessingDrawingNamespace(namespace string) bool {
	return namespace == constants.NamespaceWordprocessingDrawing || namespace == constants.NamespaceWordprocessingDrawingStrict
}

func isDrawingNamespace(namespace string) bool {
	return namespace == constants.NamespaceDrawing || namespace == constants.NamespaceDrawingStrict
}

func isPictureNamespace(namespace string) bool {
	return namespace == constants.NamespacePicture || namespace == constants.NamespacePictureStrict
}

func isWordElement(elem *Element, local string) bool {
	return elem != nil && isWordNamespace(elem.Name.Space) && elem.Name.Local == local
}

func isWordprocessingDrawingElement(elem *Element, local string) bool {
	return elem != nil && isWordprocessingDrawingNamespace(elem.Name.Space) && elem.Name.Local == local
}

func isDrawingElement(elem *Element, local string) bool {
	return elem != nil && isDrawingNamespace(elem.Name.Space) && elem.Name.Local == local
}

func isPictureElement(elem *Element, local string) bool {
	return elem != nil && isPictureNamespace(elem.Name.Space) && elem.Name.Local == local
}

func findWordChild(parent *Element, local string) *Element {
	if parent == nil {
		return nil
	}
	for _, child := range parent.Children {
		if isWordElement(child, local) {
			return child
		}
	}
	return nil
}

func findWordprocessingDrawingChild(parent *Element, local string) *Element {
	if parent == nil {
		return nil
	}
	for _, child := range parent.Children {
		if isWordprocessingDrawingElement(child, local) {
			return child
		}
	}
	return nil
}

func findDrawingChild(parent *Element, local string) *Element {
	if parent == nil {
		return nil
	}
	for _, child := range parent.Children {
		if isDrawingElement(child, local) {
			return child
		}
	}
	return nil
}

func findPictureChild(parent *Element, local string) *Element {
	if parent == nil {
		return nil
	}
	for _, child := range parent.Children {
		if isPictureElement(child, local) {
			return child
		}
	}
	return nil
}

func findDrawingDescendant(parent *Element, local string) *Element {
	return findDescendantMatching(parent, func(elem *Element) bool {
		return isDrawingElement(elem, local)
	})
}

func findPictureDescendant(parent *Element, local string) *Element {
	return findDescendantMatching(parent, func(elem *Element) bool {
		return isPictureElement(elem, local)
	})
}

func findDescendantMatching(parent *Element, matches func(*Element) bool) *Element {
	if parent == nil {
		return nil
	}
	for _, child := range parent.Children {
		if child == nil {
			continue
		}
		if matches(child) {
			return child
		}
		if found := findDescendantMatching(child, matches); found != nil {
			return found
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

// getAttr returns the first attribute with the requested local name. Model
// hydration must use one of the namespace-aware helpers below; this remains
// for generic XML inspection in package-level tests.
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

func getAttrQName(elem *Element, namespace, local string) (string, bool) {
	if elem == nil {
		return "", false
	}
	for _, attr := range elem.Attr {
		if attr.Name.Space == namespace && attr.Name.Local == local {
			return attr.Value, true
		}
	}
	return "", false
}

func getWordAttr(elem *Element, local string) (string, bool) {
	if elem == nil {
		return "", false
	}
	for _, attr := range elem.Attr {
		if isWordNamespace(attr.Name.Space) && attr.Name.Local == local {
			return attr.Value, true
		}
	}
	return "", false
}

func getRelationshipAttr(elem *Element, local string) (string, bool) {
	if elem == nil {
		return "", false
	}
	for _, attr := range elem.Attr {
		if isRelationshipNamespace(attr.Name.Space) && attr.Name.Local == local {
			return attr.Value, true
		}
	}
	return "", false
}

func getUnqualifiedAttr(elem *Element, local string) (string, bool) {
	return getAttrQName(elem, "", local)
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

// tableAdder is the subset of domain.Document (and partBlockContainer) that
// hydrateTable needs — just enough to add a table and hydrate into it,
// regardless of whether it's being hydrated into the document body or a
// header/footer part.
type tableAdder interface {
	AddTable(rows, cols int) (domain.Table, error)
}

func hydrateTable(doc tableAdder, elem *Element, ctx *reconstructContext) error {
	if doc == nil || elem == nil {
		return nil
	}

	rows := make([]*Element, 0, len(elem.Children))
	for _, child := range elem.Children {
		if !isWordElement(child, "tr") {
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
			if !isWordElement(child, "tc") {
				continue
			}
			cells = append(cells, child)
		}
		// Sum gridSpan values to get the actual grid column count.
		gridCols := 0
		for _, c := range cells {
			gridCols += cellGridSpan(c)
		}
		if gridCols > maxCols {
			maxCols = gridCols
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

	if err := applyTableProperties(table, elem); err != nil {
		return errors.Wrap(err, opHydrateTable)
	}

	for i, cells := range rowCells {
		row, err := table.Row(i)
		if err != nil {
			return errors.Wrap(err, opHydrateTable)
		}

		if err := applyRowProperties(row, rows[i]); err != nil {
			return errors.Wrap(err, opHydrateTable)
		}

		colOffset := 0
		for _, cellElem := range cells {
			if colOffset >= table.ColumnCount() {
				break
			}

			cell, err := row.Cell(colOffset)
			if err != nil {
				return errors.Wrap(err, opHydrateTable)
			}

			if err := hydrateTableCell(cell, cellElem, ctx); err != nil {
				return err
			}

			colOffset += cellGridSpan(cellElem)
		}
	}

	return nil
}

// applyTableProperties hydrates the <w:tblPr> children the domain model can
// hold. The round-trip merger retains the remaining children in the source
// XML until an explicit model change conflicts with them.
func applyTableProperties(table domain.Table, elem *Element) error {
	if table == nil || elem == nil {
		return nil
	}
	tblPr := findWordChild(elem, "tblPr")
	if tblPr == nil {
		return nil
	}

	if err := applyTableStyle(table, tblPr); err != nil {
		return err
	}
	if err := applyTableWidth(table, tblPr); err != nil {
		return err
	}
	if err := applyTableAlignment(table, tblPr); err != nil {
		return err
	}
	return applyTableBorders(table, tblPr)
}

// applyTableStyle hydrates a table's <w:tblStyle> reference (e.g. "TableGrid").
// A table style commonly carries visible properties -- borders, shading,
// banding -- defined once in styles.xml and referenced by name rather than
// repeated on every table; dropping the reference orphans that rendering even
// though the style definition itself survives untouched in styles.xml.
func applyTableStyle(table domain.Table, tblPr *Element) error {
	styleElem := findWordChild(tblPr, "tblStyle")
	if styleElem == nil {
		return nil
	}
	val, ok := getWordAttr(styleElem, "val")
	if !ok || val == "" {
		return nil
	}
	return table.SetStyle(domain.TableStyle{Name: val})
}

// applyTableWidth hydrates <w:tblW>. Unlike a cell width, domain.TableWidth
// carries its own type, so pct and auto round-trip as themselves.
func applyTableWidth(table domain.Table, tblPr *Element) error {
	widthElem := findWordChild(tblPr, "tblW")
	if widthElem == nil {
		return nil
	}

	typeVal, _ := getWordAttr(widthElem, "type")
	widthType, ok := mapWidthType(typeVal)
	if !ok {
		return nil
	}
	if widthType == domain.WidthAuto {
		// w:w is ignored for auto, and auto is already the domain default.
		return nil
	}

	val, ok := getWordAttr(widthElem, "w")
	if !ok || val == "" {
		return nil
	}
	n, err := parseMeasurementInt(val)
	if err != nil {
		return errors.WrapWithContext(err, opHydrateTable, map[string]interface{}{"attr": "tblW/w", "value": val})
	}
	if n < 0 {
		return nil
	}

	return table.SetWidth(domain.TableWidth{Type: widthType, Value: n})
}

// applyTableAlignment hydrates <w:jc>, which positions the table itself within
// the text column -- not the text inside its cells.
func applyTableAlignment(table domain.Table, tblPr *Element) error {
	jcElem := findWordChild(tblPr, "jc")
	if jcElem == nil {
		return nil
	}
	val, ok := getWordAttr(jcElem, "val")
	if !ok || val == "" {
		return nil
	}
	align, mapped := mapAlignment(val)
	if !mapped {
		return nil
	}
	return table.SetAlignment(align)
}

// applyTableBorders hydrates <w:tblBorders>, the borders drawn around and
// inside the table as a whole. A table can carry these instead of (or on top
// of) per-cell w:tcBorders, so reading only the cell borders leaves a
// hand-bordered table looking borderless.
func applyTableBorders(table domain.Table, tblPr *Element) error {
	bordersElem := findWordChild(tblPr, "tblBorders")
	if bordersElem == nil {
		return nil
	}

	borders := domain.TableLevelBorders{
		Top:     parseBorder(findWordChild(bordersElem, "top")),
		Left:    parseBorder(findWordChild(bordersElem, "left")),
		Bottom:  parseBorder(findWordChild(bordersElem, "bottom")),
		Right:   parseBorder(findWordChild(bordersElem, "right")),
		InsideH: parseBorder(findWordChild(bordersElem, "insideH")),
		InsideV: parseBorder(findWordChild(bordersElem, "insideV")),
	}
	if borders == (domain.TableLevelBorders{}) {
		return nil
	}

	return table.SetBorders(borders)
}

// applyRowProperties hydrates the <w:trPr> children the domain model can hold.
// Row properties were not read at all before this: hydration collected a row's
// <w:tc> children and nothing else.
func applyRowProperties(row domain.TableRow, elem *Element) error {
	if row == nil || elem == nil {
		return nil
	}
	trPr := findWordChild(elem, "trPr")
	if trPr == nil {
		return nil
	}

	heightElem := findWordChild(trPr, "trHeight")
	if heightElem == nil {
		return nil
	}
	val, ok := getWordAttr(heightElem, "val")
	if !ok || val == "" {
		return nil
	}
	n, err := parseMeasurementInt(val)
	if err != nil {
		return errors.WrapWithContext(err, opHydrateTable, map[string]interface{}{"attr": "trHeight/val", "value": val})
	}
	if n <= 0 {
		return nil
	}

	// w:hRule is dropped: the domain holds a bare twip count and the
	// serializer always writes hRule="atLeast", so "exact" and "auto" both
	// come back as a minimum height.
	return row.SetHeight(n)
}

// parseBorder converts one w:top/w:left/... border element into the domain
// shape, returning the zero BorderStyle for anything with no representation.
func parseBorder(elem *Element) domain.BorderStyle {
	if elem == nil {
		return domain.BorderStyle{}
	}

	val, _ := getWordAttr(elem, "val")
	style, ok := mapBorderLineStyle(val)
	if !ok {
		return domain.BorderStyle{}
	}

	border := domain.BorderStyle{Style: style}
	if sz, ok := getWordAttr(elem, "sz"); ok && sz != "" {
		if n, err := strconv.Atoi(sz); err == nil && n > 0 {
			border.Width = n
		}
	}
	if c, ok := getWordAttr(elem, "color"); ok && c != "" && !strings.EqualFold(c, "auto") {
		if clr, err := pkgcolor.FromHex(c); err == nil {
			border.Color = clr
		}
	}

	return border
}

// mapBorderLineStyle maps ST_Border onto the seven line styles the domain
// models. The false return means "nothing to hydrate", not "unknown": an
// absent, "nil" or "none" border is indistinguishable from BorderNone (the
// zero value), so an explicit w:val="none" that suppresses a style-supplied
// border cannot be represented and is dropped.
func mapBorderLineStyle(value string) (domain.BorderLineStyle, bool) {
	switch strings.ToLower(value) {
	case "", "nil", "none":
		return domain.BorderNone, false
	case "single":
		return domain.BorderSingle, true
	case "dotted":
		return domain.BorderDotted, true
	case "dashed":
		return domain.BorderDashed, true
	case "double":
		return domain.BorderDouble, true
	case "triple":
		return domain.BorderTriple, true
	case "thick":
		return domain.BorderThick, true
	default:
		// ST_Border has on the order of 180 members. Approximating the rest
		// as a plain line keeps the border visible; dropping it would erase a
		// line the author drew.
		return domain.BorderSingle, true
	}
}

// mapWidthType maps ST_TblWidth onto domain.WidthType. "nil" and any unknown
// value report false -- there is no domain way to say "no width at all", and
// guessing one changes the layout.
func mapWidthType(value string) (domain.WidthType, bool) {
	switch strings.ToLower(value) {
	case "", "auto":
		return domain.WidthAuto, true
	case "dxa":
		return domain.WidthDXA, true
	case "pct":
		return domain.WidthPct, true
	default:
		return domain.WidthAuto, false
	}
}

// applyCellProperties hydrates the <w:tcPr> children that describe how a cell
// looks. Merge state (gridSpan, vMerge) is handled by the caller because it
// also drives how many grid columns the cell consumes.
func applyCellProperties(cell domain.TableCell, tcPr *Element) error {
	if err := applyCellWidth(cell, tcPr); err != nil {
		return err
	}
	if err := applyCellBorders(cell, tcPr); err != nil {
		return err
	}
	if err := applyCellShading(cell, tcPr); err != nil {
		return err
	}
	return applyCellVerticalAlignment(cell, tcPr)
}

// applyCellWidth hydrates <w:tcW>, but only when it is expressed in twips.
//
// domain.TableCell.SetWidth takes a bare twip count with no type, and the
// serializer writes w:type="dxa" for any positive width. Hydrating
// <w:tcW w:type="pct" w:w="2500"/> -- half the table -- would therefore write
// it back as 2500 twips, a fixed 1.7 inches. That is wrong, not merely lossy,
// so a non-dxa width is left alone: it degrades to auto, which still lays out
// sensibly.
func applyCellWidth(cell domain.TableCell, tcPr *Element) error {
	widthElem := findWordChild(tcPr, "tcW")
	if widthElem == nil {
		return nil
	}
	if typeVal, ok := getWordAttr(widthElem, "type"); ok && !strings.EqualFold(typeVal, constants.WidthTypeDXA) {
		return nil
	}

	val, ok := getWordAttr(widthElem, "w")
	if !ok || val == "" {
		return nil
	}
	n, err := parseMeasurementInt(val)
	if err != nil {
		return errors.WrapWithContext(err, opHydrateTableCell, map[string]interface{}{"attr": "tcW/w", "value": val})
	}
	if n <= 0 {
		return nil
	}

	if err := cell.SetWidth(n); err != nil {
		return errors.Wrap(err, opHydrateTableCell)
	}
	return nil
}

// applyCellBorders hydrates <w:tcBorders>. Only the four outer sides are read:
// domain.TableBorders has no insideH/insideV, and those are meaningless on a
// single cell anyway.
func applyCellBorders(cell domain.TableCell, tcPr *Element) error {
	bordersElem := findWordChild(tcPr, "tcBorders")
	if bordersElem == nil {
		return nil
	}

	borders := domain.TableBorders{
		Top:    parseBorder(findWordChild(bordersElem, "top")),
		Left:   parseBorder(findWordChild(bordersElem, "left")),
		Bottom: parseBorder(findWordChild(bordersElem, "bottom")),
		Right:  parseBorder(findWordChild(bordersElem, "right")),
	}
	if borders == (domain.TableBorders{}) {
		return nil
	}

	if err := cell.SetBorders(borders); err != nil {
		return errors.Wrap(err, opHydrateTableCell)
	}
	return nil
}

// applyCellShading hydrates <w:shd> when it resolves to one flat colour, which
// is all domain.TableCell.SetShading can hold. A pattern fill or an "auto"
// colour has no single colour to map onto, and inventing one paints a
// background the source never had. When the shading is also theme-linked, the
// link is captured too (see HydrateThemeFill) so the writer can keep it
// following the document's theme instead of freezing the cached colour.
func applyCellShading(cell domain.TableCell, tcPr *Element) error {
	shdElem := findWordChild(tcPr, "shd")
	if shdElem == nil {
		return nil
	}

	// w:shd paints a w:val pattern in w:color over a w:fill background, so
	// which attribute carries the colour a reader actually sees depends on the
	// pattern -- the two are not interchangeable. "clear" draws no pattern and
	// leaves the background showing, so w:fill is the visible colour. "solid"
	// is the opposite extreme, a 100% foreground fill that hides the
	// background entirely, so w:color is. Reading w:fill for both turns a
	// solid red-on-blue cell blue.
	val, _ := getWordAttr(shdElem, "val")
	var source string
	switch {
	case val == "" || strings.EqualFold(val, "clear"):
		source = "fill"
	case strings.EqualFold(val, "solid"):
		source = "color"
	default:
		// Any real pattern (pct25, thinDiagStripe, ...) blends the two
		// colours, and the domain holds one flat colour, so neither
		// attribute on its own is the right answer.
		return nil
	}

	// A cached concrete colour, if present, is set first -- SetShading clears
	// any theme link (see its doc comment), so running it before the theme
	// hydration below, not after, matters: reversed, it would wipe out the
	// very link this same element is about to hydrate.
	if raw, ok := getWordAttr(shdElem, source); ok && isHexRGB(raw) {
		if clr, err := pkgcolor.FromHex(raw); err == nil {
			if err := cell.SetShading(clr); err != nil {
				return errors.Wrap(err, opHydrateTableCell)
			}
		}
	}

	// A producer writes that cached colour as a fallback for a consumer that
	// doesn't resolve themes, but MS-OE376 is explicit that the theme
	// reference is the primary colour and the cached fallback is only used
	// when the reference is absent -- so a producer is free to write
	// w:themeFill (or w:themeColor) with no cached w:fill/w:color at all, and
	// that link must still be kept. This is why it isn't folded into an
	// early-return guard alongside the block above: it must run whether or
	// not a concrete colour was present there.
	themeAttr, tintAttr, shadeAttr := "themeFill", "themeFillTint", "themeFillShade"
	if source == "color" {
		themeAttr, tintAttr, shadeAttr = "themeColor", "themeTint", "themeShade"
	}
	if themeVal, ok := getWordAttr(shdElem, themeAttr); ok && themeVal != "" {
		if hydrator, ok := cell.(interface{ HydrateThemeFill(string, string, string) }); ok {
			tint, _ := getWordAttr(shdElem, tintAttr)
			shade, _ := getWordAttr(shdElem, shadeAttr)
			hydrator.HydrateThemeFill(themeVal, tint, shade)
		}
	}

	return nil
}

// applyCellVerticalAlignment hydrates <w:vAlign>. ST_VerticalJc's "both"
// (distribute vertically) has no domain equivalent and is dropped.
func applyCellVerticalAlignment(cell domain.TableCell, tcPr *Element) error {
	vAlignElem := findWordChild(tcPr, "vAlign")
	if vAlignElem == nil {
		return nil
	}
	val, ok := getWordAttr(vAlignElem, "val")
	if !ok || val == "" {
		return nil
	}

	var align domain.VerticalAlignment
	switch strings.ToLower(val) {
	case "top":
		align = domain.VerticalAlignTop
	case "center":
		align = domain.VerticalAlignCenter
	case "bottom":
		align = domain.VerticalAlignBottom
	default:
		return nil
	}

	if err := cell.SetVerticalAlignment(align); err != nil {
		return errors.Wrap(err, opHydrateTableCell)
	}
	return nil
}

// isHexRGB reports whether s is exactly six hex digits -- the only form Word
// writes for w:fill. pkgcolor.FromHex also accepts the three-digit shorthand,
// which in a .docx is far more likely to be garbage than a colour.
func isHexRGB(s string) bool {
	if len(s) != 6 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

func hydrateTableCell(cell domain.TableCell, elem *Element, ctx *reconstructContext) error {
	if cell == nil || elem == nil {
		return nil
	}

	// Parse cell properties (w:tcPr) for merge info and appearance.
	if tcPr := findWordChild(elem, "tcPr"); tcPr != nil {
		if err := applyCellProperties(cell, tcPr); err != nil {
			return err
		}
		if gs := findWordChild(tcPr, "gridSpan"); gs != nil {
			if val, ok := getWordAttr(gs, "val"); ok && val != "" {
				span, err := strconv.Atoi(val)
				if err != nil {
					return errors.WrapWithContext(err, opHydrateTableCell, map[string]interface{}{"attr": "gridSpan", "value": val})
				}
				if span > 1 {
					if err := cell.Merge(span, 1); err != nil {
						return errors.Wrap(err, opHydrateTableCell)
					}
				}
			}
		}
		if vm := findWordChild(tcPr, "vMerge"); vm != nil {
			val, _ := getWordAttr(vm, "val")
			switch val {
			case "restart":
				if err := cell.SetVMerge(domain.VMergeRestart); err != nil {
					return errors.Wrap(err, opHydrateTableCell)
				}
			default:
				// Empty or absent val means "continue"
				if err := cell.SetVMerge(domain.VMergeContinue); err != nil {
					return errors.Wrap(err, opHydrateTableCell)
				}
			}
		}
	}

	for _, child := range elem.Children {
		if !isWordElement(child, "p") {
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

// cellGridSpan returns the grid column span for a <w:tc> element (defaults to 1).
func cellGridSpan(tcElem *Element) int {
	if tcPr := findWordChild(tcElem, "tcPr"); tcPr != nil {
		if gs := findWordChild(tcPr, "gridSpan"); gs != nil {
			if val, ok := getWordAttr(gs, "val"); ok {
				if n, err := strconv.Atoi(val); err == nil && n > 1 {
					return n
				}
			}
		}
	}
	return 1
}

// preserveOriginalParts stores all original document parts for round-trip operations.
// This ensures that when a document is read and saved, all parts that we don't
// actively modify are preserved exactly as they were in the original.
func preserveOriginalParts(doc domain.Document, pkg *Package) {
	setter, ok := doc.(interface {
		SetPreservedParts(*core.PreservedParts)
	})
	if !ok {
		return
	}

	parts := &core.PreservedParts{
		Headers:    make(map[string][]byte, len(pkg.Headers)),
		Footers:    make(map[string][]byte, len(pkg.Footers)),
		HeaderRels: make(map[string][]byte, len(pkg.HeaderRels)),
		FooterRels: make(map[string][]byte, len(pkg.FooterRels)),
		Additional: make(map[string][]byte, len(pkg.AdditionalParts)),
		Themes:     make(map[string][]byte, len(pkg.ThemeParts)),
	}

	// Preserve headers
	for name, data := range pkg.Headers {
		parts.Headers[name] = data
	}

	// Preserve footers
	for name, data := range pkg.Footers {
		parts.Footers[name] = data
	}

	// Preserve header/footer relationship parts
	for name, data := range pkg.HeaderRels {
		parts.HeaderRels[name] = data
	}
	for name, data := range pkg.FooterRels {
		parts.FooterRels[name] = data
	}

	// Preserve document relationships
	if len(pkg.DocumentRelationships) > 0 {
		parts.DocRels = pkg.DocumentRelationships
	}

	// Preserve content types as raw XML
	if pkg.ContentTypes != nil {
		if rawCT, exists := pkg.RawParts["[Content_Types].xml"]; exists {
			parts.ContentTypes = rawCT
		}
	}

	// Preserve additional parts (comments, footnotes, customXml, etc.)
	for name, data := range pkg.AdditionalParts {
		parts.Additional[name] = data
	}

	// Preserve themes
	for name, data := range pkg.ThemeParts {
		parts.Themes[name] = data
	}

	// Preserve font table
	if len(pkg.FontTable) > 0 {
		parts.FontTable = pkg.FontTable
	}

	// Preserve settings
	if len(pkg.Settings) > 0 {
		parts.Settings = pkg.Settings
	}

	// Preserve web settings
	if len(pkg.WebSettings) > 0 {
		parts.WebSettings = pkg.WebSettings
	}

	// Preserve custom properties
	if len(pkg.CustomProperties) > 0 {
		parts.CustomProperties = pkg.CustomProperties
	}

	// Preserve root relationships (_rels/.rels)
	if len(pkg.RootRelationships) > 0 {
		parts.RootRels = pkg.RootRelationships
	}

	setter.SetPreservedParts(parts)
}

// corePropsRead is an unmarshal-friendly representation of docProps/core.xml.
// Go's encoding/xml resolves namespace prefixes to their URIs during decode,
// so the struct tags must use "namespace-URI element" format.
type corePropsRead struct {
	Title       string      `xml:"http://purl.org/dc/elements/1.1/ title"`
	Subject     string      `xml:"http://purl.org/dc/elements/1.1/ subject"`
	Creator     string      `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Description string      `xml:"http://purl.org/dc/elements/1.1/ description"`
	Keywords    string      `xml:"http://schemas.openxmlformats.org/package/2006/metadata/core-properties keywords"`
	Created     *dcDateRead `xml:"http://purl.org/dc/terms/ created"`
	Modified    *dcDateRead `xml:"http://purl.org/dc/terms/ modified"`
}

type dcDateRead struct {
	Value string `xml:",chardata"`
}

// parseCoreProperties unmarshals docProps/core.xml bytes into domain.Metadata.
func parseCoreProperties(data []byte) (*domain.Metadata, error) {
	var cp corePropsRead
	if err := xml.Unmarshal(data, &cp); err != nil {
		return nil, err
	}

	meta := &domain.Metadata{
		Title:       cp.Title,
		Subject:     cp.Subject,
		Creator:     cp.Creator,
		Description: cp.Description,
	}

	if cp.Keywords != "" {
		parts := strings.Split(cp.Keywords, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				meta.Keywords = append(meta.Keywords, p)
			}
		}
	}

	if cp.Created != nil {
		meta.Created = cp.Created.Value
	}
	if cp.Modified != nil {
		meta.Modified = cp.Modified.Value
	}

	return meta, nil
}

// stylesLanguageRead is a read-only mirror of styles.xml's
// docDefaults/rPrDefault/rPr/lang path, used only for Unmarshal. It can't
// reuse internal/xml's Styles/DocDefaults/RunDefaults/RunProperties/Language
// (marshal-only structs with literal "w:"-prefixed tags): Go's encoding/xml
// resolves namespace prefixes to their URIs during decode, so a struct meant
// for Unmarshal needs "namespace-URI element"/"namespace-URI attr,attr" tags
// instead of the literal "w:"-prefixed ones the write-side structs use.
type stylesLanguageRead struct {
	DocDefaults *docDefaultsRead `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main docDefaults"`
}

type docDefaultsRead struct {
	RunDefaults *rPrDefaultRead `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPrDefault"`
}

type rPrDefaultRead struct {
	Properties *rPrLangRead `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPr"`
}

type rPrLangRead struct {
	Lang *langRead `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lang"`
}

type langRead struct {
	Val      string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
	EastAsia string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main eastAsia,attr"`
	Bidi     string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bidi,attr"`
}

// parseStylesLanguage unmarshals word/styles.xml bytes and extracts the
// document's default proofing language from
// w:docDefaults/w:rPrDefault/w:rPr/w:lang, if present. Returns nil (with no
// error) when styles.xml doesn't declare one.
func parseStylesLanguage(data []byte) (*domain.Language, error) {
	var styles stylesLanguageRead
	if err := xml.Unmarshal(data, &styles); err != nil {
		return nil, err
	}

	if styles.DocDefaults == nil || styles.DocDefaults.RunDefaults == nil ||
		styles.DocDefaults.RunDefaults.Properties == nil {
		return nil, nil
	}

	lang := styles.DocDefaults.RunDefaults.Properties.Lang
	if lang == nil || (lang.Val == "" && lang.EastAsia == "" && lang.Bidi == "") {
		return nil, nil
	}

	return &domain.Language{Val: lang.Val, EastAsia: lang.EastAsia, Bidi: lang.Bidi}, nil
}

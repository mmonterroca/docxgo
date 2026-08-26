// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package core

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/ooxmlmerge"
	"github.com/mmonterroca/docxgo/v2/internal/serializer"
	xmlstructs "github.com/mmonterroca/docxgo/v2/internal/xml"
	"github.com/mmonterroca/docxgo/v2/pkg/color"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
)

// RoundTripMainDocumentSource describes the original word/document.xml and
// the model objects hydrated from each direct body child. It is constructed
// by internal/reader and consumed only by the concrete document type.
type RoundTripMainDocumentSource struct {
	Original           []byte
	Prefix             []byte
	Suffix             []byte
	MainNamespace      string
	MaxDrawingID       int
	BodyStartOffset    int
	BackgroundStart    int
	BackgroundEnd      int
	Namespaces         map[string]string
	Entries            []RoundTripBodyEntrySource
	FinalSectionPrefix []byte
	FinalSectionRaw    []byte
	FinalSection       *RoundTripSectionPropertiesSource
	BodyTail           []byte
}

type RoundTripSectionPropertiesSource struct {
	Raw        []byte
	Namespaces map[string]string
}

// RoundTripBodyEntrySource associates one original body child with the model
// blocks it produced. Unsupported children have no Blocks and remain opaque.
type RoundTripBodyEntrySource struct {
	Prefix  []byte
	Raw     []byte
	Blocks  []domain.Block
	Table   *RoundTripTableSource
	Section *RoundTripSectionPropertiesSource
}

// RoundTripTableSource splits an original table around its contiguous row
// sequence. Prefix retains tblPr/tblGrid and any unsupported table properties;
// Suffix retains content after the last row and the closing table tag.
type RoundTripTableSource struct {
	Table         domain.Table
	Namespaces    map[string]string
	Open          []byte
	ShellTail     []byte
	ShellChildren []RoundTripTableShellChildSource
	Suffix        []byte
	Rows          []RoundTripRowSource
}

type RoundTripTableShellChildSource struct {
	Prefix []byte
	Raw    []byte
	Name   string
}

// RoundTripRowSource associates one original w:tr fragment with its row.
type RoundTripRowSource struct {
	Prefix []byte
	Raw    []byte
	Row    domain.TableRow
}

type roundTripMainDocument struct {
	original                 []byte
	prefix                   []byte
	suffix                   []byte
	mainNamespace            string
	maxDrawingID             int
	bodyStartOffset          int
	backgroundStart          int
	backgroundEnd            int
	backgroundSnapshot       string
	entries                  []roundTripBodyEntry
	finalSectionPrefix       []byte
	finalSectionRaw          []byte
	finalSection             *roundTripSectionProperties
	bodyTail                 []byte
	finalSectionSnapshot     string
	finalSectionIdentity     domain.Section
	originalParagraphEntries map[domain.Paragraph]int
	namespaces               map[string]string
}

type roundTripBodyEntry struct {
	prefix               []byte
	raw                  []byte
	blocks               []domain.Block
	snapshot             string
	paragraphSnapshot    string
	sectionBreakSnapshot string
	table                *roundTripTable
	section              *roundTripSectionProperties
}

type roundTripTable struct {
	table              domain.Table
	namespaces         map[string]string
	mainNamespace      string
	open               []byte
	shellTail          []byte
	shellChildren      []roundTripTableShellChild
	suffix             []byte
	shellSnapshot      string
	propertiesSnapshot string
	gridSnapshot       string
	rows               []roundTripRow
}

type roundTripTableShellChild struct {
	prefix     []byte
	raw        []byte
	name       string
	properties *roundTripTableProperties
	grid       *roundTripTableGrid
}

type roundTripTableProperties struct {
	source     []byte
	namespaces map[string]string
}

type roundTripTableGrid struct {
	source     []byte
	namespaces map[string]string
}

type roundTripSectionProperties struct {
	source        []byte
	snapshot      []byte
	namespaces    map[string]string
	mainNamespace string
}

type roundTripRow struct {
	prefix   []byte
	raw      []byte
	row      domain.TableRow
	snapshot string
}

func cloneBytes(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	out := make([]byte, len(data))
	copy(out, data)
	return out
}

func cloneNamespaces(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for prefix, namespace := range source {
		result[prefix] = namespace
	}
	return result
}

func isWordprocessingNamespace(namespace string) bool {
	return namespace == constants.NamespaceMain || namespace == constants.NamespaceMainStrict
}

func newRoundTripSectionProperties(source *RoundTripSectionPropertiesSource, namespaces map[string]string, mainNamespace string) *roundTripSectionProperties {
	if source == nil {
		return nil
	}
	sectionNamespaces := source.Namespaces
	if len(sectionNamespaces) == 0 {
		sectionNamespaces = namespaces
	}
	return &roundTripSectionProperties{
		source:        cloneBytes(source.Raw),
		namespaces:    cloneNamespaces(sectionNamespaces),
		mainNamespace: mainNamespace,
	}
}

// SetRoundTripMainDocument snapshots the hydrated model and stores the raw
// fragments needed to preserve unsupported OOXML on a later write.
func (d *document) SetRoundTripMainDocument(source *RoundTripMainDocumentSource) error {
	if d == nil || source == nil || len(source.Original) == 0 {
		return nil
	}

	sourceNamespaces := cloneNamespaces(source.Namespaces)
	if len(sourceNamespaces) == 0 {
		sourceNamespaces = cloneNamespaces(roundTripNamespaces)
	}
	rt := &roundTripMainDocument{
		original:                 cloneBytes(source.Original),
		prefix:                   cloneBytes(source.Prefix),
		suffix:                   cloneBytes(source.Suffix),
		mainNamespace:            source.MainNamespace,
		maxDrawingID:             source.MaxDrawingID,
		bodyStartOffset:          source.BodyStartOffset,
		backgroundStart:          source.BackgroundStart,
		backgroundEnd:            source.BackgroundEnd,
		finalSectionPrefix:       cloneBytes(source.FinalSectionPrefix),
		finalSectionRaw:          cloneBytes(source.FinalSectionRaw),
		bodyTail:                 cloneBytes(source.BodyTail),
		originalParagraphEntries: make(map[domain.Paragraph]int),
		entries:                  make([]roundTripBodyEntry, 0, len(source.Entries)),
		namespaces:               sourceNamespaces,
	}
	if !isWordprocessingNamespace(rt.mainNamespace) {
		rt.mainNamespace = constants.NamespaceMain
	}
	rt.finalSection = newRoundTripSectionProperties(source.FinalSection, rt.namespaces, rt.mainNamespace)
	if len(d.sections) > 0 {
		rt.finalSectionIdentity = d.sections[len(d.sections)-1]
	}
	backgroundSnapshot, err := documentBackgroundFingerprint(d)
	if err != nil {
		return fmt.Errorf("snapshot document background: %w", err)
	}
	rt.backgroundSnapshot = backgroundSnapshot

	for _, entrySource := range source.Entries {
		entry := roundTripBodyEntry{
			prefix:  cloneBytes(entrySource.Prefix),
			raw:     cloneBytes(entrySource.Raw),
			blocks:  append([]domain.Block(nil), entrySource.Blocks...),
			section: newRoundTripSectionProperties(entrySource.Section, rt.namespaces, rt.mainNamespace),
		}
		if entrySource.Table != nil && entrySource.Table.Table != nil {
			tableNamespaces := entrySource.Table.Namespaces
			if len(tableNamespaces) == 0 {
				tableNamespaces = rt.namespaces
			}
			tableState := &roundTripTable{
				table:         entrySource.Table.Table,
				namespaces:    cloneNamespaces(tableNamespaces),
				mainNamespace: rt.mainNamespace,
				open:          cloneBytes(entrySource.Table.Open),
				shellTail:     cloneBytes(entrySource.Table.ShellTail),
				suffix:        cloneBytes(entrySource.Table.Suffix),
				shellChildren: make([]roundTripTableShellChild, 0, len(entrySource.Table.ShellChildren)),
				rows:          make([]roundTripRow, 0, len(entrySource.Table.Rows)),
			}
			props, grid := serializer.NewTableSerializer().SerializeShell(tableState.table)
			shell, err := tableShellFingerprintFromParts(props, grid)
			if err != nil {
				return fmt.Errorf("snapshot table shell: %w", err)
			}
			tableState.shellSnapshot = shell
			propsXML, err := marshalXML(props)
			if err != nil {
				return fmt.Errorf("snapshot table properties: %w", err)
			}
			gridXML, err := marshalXML(grid)
			if err != nil {
				return fmt.Errorf("snapshot table grid: %w", err)
			}
			tableState.propertiesSnapshot = string(propsXML)
			tableState.gridSnapshot = string(gridXML)
			for _, childSource := range entrySource.Table.ShellChildren {
				child := roundTripTableShellChild{
					prefix: cloneBytes(childSource.Prefix),
					raw:    cloneBytes(childSource.Raw),
					name:   childSource.Name,
				}
				if childSource.Name == "tblPr" {
					child.properties = &roundTripTableProperties{source: cloneBytes(childSource.Raw), namespaces: tableState.namespaces}
				}
				if childSource.Name == "tblGrid" {
					child.grid = &roundTripTableGrid{source: cloneBytes(childSource.Raw), namespaces: tableState.namespaces}
				}
				tableState.shellChildren = append(tableState.shellChildren, child)
			}
			for _, rowSource := range entrySource.Table.Rows {
				if rowSource.Row == nil {
					continue
				}
				fingerprint, err := tableRowFingerprint(rowSource.Row)
				if err != nil {
					return fmt.Errorf("snapshot table row: %w", err)
				}
				tableState.rows = append(tableState.rows, roundTripRow{
					prefix:   cloneBytes(rowSource.Prefix),
					raw:      cloneBytes(rowSource.Raw),
					row:      rowSource.Row,
					snapshot: fingerprint,
				})
			}
			entry.table = tableState
		} else if len(entry.blocks) > 0 {
			fingerprint, err := blockGroupFingerprint(entry.blocks)
			if err != nil {
				return fmt.Errorf("snapshot body block: %w", err)
			}
			entry.snapshot = fingerprint
			if entry.section != nil {
				for _, block := range entry.blocks {
					switch {
					case block.Paragraph != nil:
						entry.paragraphSnapshot, err = blockGroupFingerprint([]domain.Block{block})
						if err != nil {
							return fmt.Errorf("snapshot section carrier paragraph: %w", err)
						}
					case block.SectionBreak != nil:
						entry.sectionBreakSnapshot, err = blockGroupFingerprint([]domain.Block{block})
						if err != nil {
							return fmt.Errorf("snapshot embedded section break: %w", err)
						}
						sectionXML, findErr := firstNamedElement([]byte(entry.sectionBreakSnapshot), wordName("sectPr"), roundTripNamespaces)
						if findErr != nil {
							return fmt.Errorf("snapshot embedded section properties: %w", findErr)
						}
						entry.section.snapshot = cloneBytes(sectionXML)
					}
				}
			}
		}

		entryIndex := len(rt.entries)
		for _, block := range entry.blocks {
			if block.Paragraph != nil {
				rt.originalParagraphEntries[block.Paragraph] = entryIndex
			}
		}
		rt.entries = append(rt.entries, entry)
	}

	finalSnapshot, err := finalSectionFingerprint(d)
	if err != nil {
		return fmt.Errorf("snapshot final section: %w", err)
	}
	rt.finalSectionSnapshot = finalSnapshot
	if rt.finalSection != nil {
		rt.finalSection.snapshot = []byte(finalSnapshot)
	}
	d.roundTripMain = rt
	return nil
}

func marshalXML(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	return xml.Marshal(value)
}

func renderGeneratedFragment(data []byte, context map[string]string, mainNamespace string) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}
	nodes, err := ooxmlmerge.ParseAll(data, roundTripNamespaces)
	if err != nil {
		return nil, err
	}
	edits := make([]ooxmlmerge.Edit, 0, len(nodes))
	for _, node := range nodes {
		rendered, err := ooxmlmerge.RenderInContextWithMainNamespace(node, context, mainNamespace)
		if err != nil {
			return nil, err
		}
		edits = append(edits, ooxmlmerge.Edit{Start: node.Start, End: node.End, Replacement: rendered})
	}
	return ooxmlmerge.ApplyEdits(data, edits)
}

func ensureRequiredDocumentNamespaces(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	root, err := ooxmlmerge.Parse(data, nil)
	if err != nil || root == nil {
		return data
	}
	wordNamespace := constants.NamespaceMain
	relationshipsNamespace := constants.NamespaceRelationships
	wordDrawingNamespace := constants.NamespaceWordprocessingDrawing
	if root.Name.Namespace == constants.NamespaceMainStrict {
		wordNamespace = constants.NamespaceMainStrict
		relationshipsNamespace = constants.NamespaceRelationshipsStrict
		wordDrawingNamespace = constants.NamespaceWordprocessingDrawingStrict
	}
	namespaces := []struct {
		prefix string
		uri    string
	}{
		{prefix: "w", uri: wordNamespace},
		{prefix: "r", uri: relationshipsNamespace},
		{prefix: "wp", uri: wordDrawingNamespace},
	}
	var declarations bytes.Buffer
	for _, namespace := range namespaces {
		if !usesUnboundPrefix(root, namespace.prefix) {
			continue
		}
		fmt.Fprintf(&declarations, ` xmlns:%s="%s"`, namespace.prefix, namespace.uri)
	}
	if declarations.Len() == 0 {
		return data
	}

	insertAt := root.ContentStart - 1
	if insertAt > root.Start && data[insertAt-1] == '/' {
		insertAt--
	}
	result := append([]byte(nil), data[:insertAt]...)
	result = append(result, declarations.Bytes()...)
	result = append(result, data[insertAt:]...)
	return result
}

func usesUnboundPrefix(node *ooxmlmerge.Node, prefix string) bool {
	if node == nil {
		return false
	}
	if node.Name.Prefix == prefix && node.Name.Namespace == "" {
		return true
	}
	for _, attr := range node.Attributes {
		if attr.Name.Prefix == prefix && attr.Name.Namespace == "" {
			return true
		}
	}
	for _, child := range node.Children {
		if usesUnboundPrefix(child, prefix) {
			return true
		}
	}
	return false
}

func blockGroupBytesWithSerializer(ser *serializer.DocumentSerializer, blocks []domain.Block) ([]byte, error) {
	if ser == nil {
		return nil, fmt.Errorf("document serializer is nil")
	}
	var out bytes.Buffer
	for _, block := range blocks {
		elem := ser.SerializeBlock(block)
		if elem == nil {
			continue
		}
		data, err := marshalXML(elem)
		if err != nil {
			return nil, err
		}
		out.Write(data)
	}
	return out.Bytes(), nil
}

func blockGroupBytesWithSerializerInContext(ser *serializer.DocumentSerializer, blocks []domain.Block, context map[string]string, mainNamespace string) ([]byte, error) {
	data, err := blockGroupBytesWithSerializer(ser, blocks)
	if err != nil {
		return nil, err
	}
	return renderGeneratedFragment(data, context, mainNamespace)
}

func blockGroupBytes(blocks []domain.Block) ([]byte, error) {
	return blockGroupBytesWithSerializer(serializer.NewDocumentSerializer(), blocks)
}

func blockGroupFingerprint(blocks []domain.Block) (string, error) {
	data, err := blockGroupBytes(blocks)
	return string(data), err
}

func tableShellFingerprint(table domain.Table) (string, error) {
	ser := serializer.NewTableSerializer()
	props, grid := ser.SerializeShell(table)
	return tableShellFingerprintFromParts(props, grid)
}

func tableShellFingerprintFromParts(props *xmlstructs.TableProperties, grid *xmlstructs.TableGrid) (string, error) {
	propsXML, err := marshalXML(props)
	if err != nil {
		return "", err
	}
	gridXML, err := marshalXML(grid)
	if err != nil {
		return "", err
	}
	return string(propsXML) + string(gridXML), nil
}

var sectionPropertyRanks = map[string]int{
	"headerReference": 0,
	"footerReference": 1,
	"footnotePr":      2,
	"endnotePr":       3,
	"type":            4,
	"pgSz":            5,
	"pgMar":           6,
	"paperSrc":        7,
	"pgBorders":       8,
	"lnNumType":       9,
	"pgNumType":       10,
	"cols":            11,
	"formProt":        12,
	"vAlign":          13,
	"noEndnote":       14,
	"titlePg":         15,
	"textDirection":   16,
	"bidi":            17,
	"rtlGutter":       18,
	"docGrid":         19,
	"printerSettings": 20,
	"sectPrChange":    21,
}

func (p *roundTripSectionProperties) compose(currentXML []byte) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("preserved section properties are nil")
	}
	return mergeRoundTripFragment(p.source, p.snapshot, currentXML, sectionPropertiesPolicy(), p.namespaces)
}

var tablePropertyRanks = map[string]int{
	"tblStyle":            0,
	"tblpPr":              1,
	"tblOverlap":          2,
	"bidiVisual":          3,
	"tblStyleRowBandSize": 4,
	"tblStyleColBandSize": 5,
	"tblW":                6,
	"jc":                  7,
	"tblCellSpacing":      8,
	"tblInd":              9,
	"tblBorders":          10,
	"shd":                 11,
	"tblLayout":           12,
	"tblCellMar":          13,
	"tblLook":             14,
	"tblCaption":          15,
	"tblDescription":      16,
	"tblPrChange":         17,
}

func tableRowBytesWithSerializer(ser *serializer.DocumentSerializer, row domain.TableRow) ([]byte, error) {
	if ser == nil {
		return nil, fmt.Errorf("document serializer is nil")
	}
	data, err := marshalXML(ser.SerializeTableRow(row))
	return data, err
}

func tableRowBytesWithSerializerInContext(ser *serializer.DocumentSerializer, row domain.TableRow, context map[string]string, mainNamespace string) ([]byte, error) {
	data, err := tableRowBytesWithSerializer(ser, row)
	if err != nil {
		return nil, err
	}
	return renderGeneratedFragment(data, context, mainNamespace)
}

func tableRowBytes(row domain.TableRow) ([]byte, error) {
	return tableRowBytesWithSerializer(serializer.NewDocumentSerializer(), row)
}

func tableRowFingerprint(row domain.TableRow) (string, error) {
	data, err := tableRowBytes(row)
	return string(data), err
}

func finalSectionBytes(d domain.Document) ([]byte, error) {
	body := serializer.NewDocumentSerializer().SerializeBody(d)
	if body == nil || body.SectPr == nil {
		return nil, nil
	}
	data, err := marshalXML(body.SectPr)
	return data, err
}

func finalSectionFingerprint(d domain.Document) (string, error) {
	data, err := finalSectionBytes(d)
	return string(data), err
}

func documentBackgroundBytes(d domain.Document) ([]byte, error) {
	backgroundColor, ok := d.BackgroundColor()
	if !ok {
		return nil, nil
	}
	data, err := marshalXML(&xmlstructs.Background{Color: color.ToHex(backgroundColor)})
	return data, err
}

func documentBackgroundFingerprint(d domain.Document) (string, error) {
	data, err := documentBackgroundBytes(d)
	return string(data), err
}

func blockIdentity(block domain.Block) interface{} {
	switch {
	case block.Paragraph != nil:
		return block.Paragraph
	case block.Table != nil:
		return block.Table
	case block.SectionBreak != nil:
		return block.SectionBreak
	default:
		return nil
	}
}

func sameBlockOrder(a, b []domain.Block) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if blockIdentity(a[i]) != blockIdentity(b[i]) {
			return false
		}
	}
	return true
}

func (rt *roundTripMainDocument) sourceBlocks() []domain.Block {
	var blocks []domain.Block
	for _, entry := range rt.entries {
		blocks = append(blocks, entry.blocks...)
	}
	return blocks
}

func (rt *roundTripMainDocument) originalParagraphUnchanged(para domain.Paragraph) bool {
	if rt == nil || para == nil {
		return false
	}
	entryIndex, ok := rt.originalParagraphEntries[para]
	if !ok || entryIndex < 0 || entryIndex >= len(rt.entries) {
		return false
	}
	entry := rt.entries[entryIndex]
	if entry.section != nil && entry.paragraphSnapshot != "" {
		paragraphBlock, _ := entryParagraphAndSection(&entry)
		if paragraphBlock.Paragraph != para {
			return false
		}
		fingerprint, err := blockGroupFingerprint([]domain.Block{paragraphBlock})
		return err == nil && fingerprint == entry.paragraphSnapshot
	}
	fingerprint, err := blockGroupFingerprint(entry.blocks)
	return err == nil && fingerprint == entry.snapshot
}

func (t *roundTripTable) unchanged() (bool, error) {
	if t == nil || t.table == nil {
		return false, nil
	}
	shell, err := tableShellFingerprint(t.table)
	if err != nil {
		return false, err
	}
	if shell != t.shellSnapshot {
		return false, nil
	}
	currentRows := t.table.Rows()
	if len(currentRows) != len(t.rows) {
		return false, nil
	}
	for i, sourceRow := range t.rows {
		if currentRows[i] != sourceRow.row {
			return false, nil
		}
		fingerprint, err := tableRowFingerprint(currentRows[i])
		if err != nil {
			return false, err
		}
		if fingerprint != sourceRow.snapshot {
			return false, nil
		}
	}
	return true, nil
}

func (rt *roundTripMainDocument) unchanged(d *document) (bool, error) {
	if rt == nil || d == nil || !sameBlockOrder(rt.sourceBlocks(), d.blocks) {
		return false, nil
	}
	for _, entry := range rt.entries {
		if entry.table != nil {
			unchanged, err := entry.table.unchanged()
			if err != nil || !unchanged {
				return false, err
			}
			continue
		}
		if len(entry.blocks) == 0 {
			continue
		}
		fingerprint, err := blockGroupFingerprint(entry.blocks)
		if err != nil {
			return false, err
		}
		if fingerprint != entry.snapshot {
			return false, nil
		}
	}
	finalSection, err := finalSectionFingerprint(d)
	if err != nil {
		return false, err
	}
	if finalSection != rt.finalSectionSnapshot {
		return false, nil
	}
	background, err := documentBackgroundFingerprint(d)
	if err != nil {
		return false, err
	}
	return background == rt.backgroundSnapshot, nil
}

func (rt *roundTripMainDocument) prefixWithCurrentBackground(d domain.Document) ([]byte, error) {
	background, err := documentBackgroundBytes(d)
	if err != nil {
		return nil, err
	}
	background, err = renderGeneratedFragment(background, rt.namespaces, rt.mainNamespace)
	if err != nil {
		return nil, err
	}
	start, end := rt.backgroundStart, rt.backgroundEnd
	if start < 0 || end < start {
		start = rt.bodyStartOffset
		end = start
	}
	if start < 0 || end > len(rt.prefix) {
		return nil, fmt.Errorf("document background offsets are outside the preserved prefix")
	}
	result := append([]byte(nil), rt.prefix[:start]...)
	result = append(result, background...)
	result = append(result, rt.prefix[end:]...)
	return result, nil
}

func (p *roundTripTableProperties) compose(current, snapshot []byte) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("preserved table properties are nil")
	}
	return mergeRoundTripFragment(p.source, snapshot, current, tablePropertiesPolicy(), p.namespaces)
}

func (g *roundTripTableGrid) compose(current, snapshot []byte) ([]byte, error) {
	if g == nil {
		return nil, fmt.Errorf("preserved table grid is nil")
	}
	return mergeRoundTripFragment(g.source, snapshot, current, tableGridPolicy(), g.namespaces)
}

func (t *roundTripTable) composeShell(props *xmlstructs.TableProperties, grid *xmlstructs.TableGrid) ([]byte, error) {
	propsXML, err := marshalXML(props)
	if err != nil {
		return nil, err
	}
	gridXML, err := marshalXML(grid)
	if err != nil {
		return nil, err
	}
	propsChanged := string(propsXML) != t.propertiesSnapshot
	gridChanged := string(gridXML) != t.gridSnapshot
	hasProperties := false
	for _, child := range t.shellChildren {
		hasProperties = hasProperties || child.name == "tblPr"
	}

	var out bytes.Buffer
	out.Write(t.open)
	propertiesEmitted := false
	gridEmitted := false
	for _, child := range t.shellChildren {
		if child.name == "tblGrid" && !hasProperties && propsChanged && !propertiesEmitted {
			generated, err := renderGeneratedFragment(propsXML, t.namespaces, t.mainNamespace)
			if err != nil {
				return nil, err
			}
			out.Write(generated)
			propertiesEmitted = true
		}
		out.Write(child.prefix)
		switch child.name {
		case "tblPr":
			propertiesEmitted = true
			switch {
			case !propsChanged:
				out.Write(child.raw)
			case child.properties != nil:
				merged, err := child.properties.compose(propsXML, []byte(t.propertiesSnapshot))
				if err != nil {
					return nil, err
				}
				out.Write(merged)
			default:
				generated, err := renderGeneratedFragment(propsXML, t.namespaces, t.mainNamespace)
				if err != nil {
					return nil, err
				}
				out.Write(generated)
			}
		case "tblGrid":
			gridEmitted = true
			if gridChanged {
				if child.grid != nil {
					merged, err := child.grid.compose(gridXML, []byte(t.gridSnapshot))
					if err != nil {
						return nil, err
					}
					out.Write(merged)
				} else {
					generated, err := renderGeneratedFragment(gridXML, t.namespaces, t.mainNamespace)
					if err != nil {
						return nil, err
					}
					out.Write(generated)
				}
			} else {
				out.Write(child.raw)
			}
		default:
			out.Write(child.raw)
		}
	}
	if !propertiesEmitted && propsChanged {
		generated, err := renderGeneratedFragment(propsXML, t.namespaces, t.mainNamespace)
		if err != nil {
			return nil, err
		}
		out.Write(generated)
	}
	if !gridEmitted && gridChanged {
		generated, err := renderGeneratedFragment(gridXML, t.namespaces, t.mainNamespace)
		if err != nil {
			return nil, err
		}
		out.Write(generated)
	}
	out.Write(t.shellTail)
	return out.Bytes(), nil
}

func (t *roundTripTable) compose(ser *serializer.DocumentSerializer) ([]byte, error) {
	if t == nil || t.table == nil {
		return nil, fmt.Errorf("round-trip table is nil")
	}
	if ser == nil {
		return nil, fmt.Errorf("document serializer is nil")
	}
	props, grid := serializer.NewTableSerializer().SerializeShell(t.table)
	shell, err := tableShellFingerprintFromParts(props, grid)
	if err != nil {
		return nil, err
	}
	if len(t.open) == 0 || len(t.suffix) == 0 {
		data, err := marshalXML(ser.SerializeTable(t.table))
		if err != nil {
			return nil, err
		}
		return renderGeneratedFragment(data, t.namespaces, t.mainNamespace)
	}

	sourceRows := make(map[domain.TableRow]int, len(t.rows))
	for index, row := range t.rows {
		sourceRows[row.row] = index
	}

	var out bytes.Buffer
	if shell == t.shellSnapshot {
		out.Write(t.open)
		for _, child := range t.shellChildren {
			out.Write(child.prefix)
			out.Write(child.raw)
		}
		out.Write(t.shellTail)
	} else {
		shellXML, err := t.composeShell(props, grid)
		if err != nil {
			return nil, err
		}
		out.Write(shellXML)
	}
	nextPrefix := 0
	for _, row := range t.table.Rows() {
		if sourceIndex, ok := sourceRows[row]; ok {
			for nextPrefix <= sourceIndex {
				out.Write(t.rows[nextPrefix].prefix)
				nextPrefix++
			}
			source := t.rows[sourceIndex]
			fingerprint, err := tableRowFingerprint(row)
			if err != nil {
				return nil, err
			}
			if fingerprint == source.snapshot {
				out.Write(source.raw)
				continue
			}
		}
		data, err := tableRowBytesWithSerializerInContext(ser, row, t.namespaces, t.mainNamespace)
		if err != nil {
			return nil, err
		}
		out.Write(data)
	}
	for nextPrefix < len(t.rows) {
		out.Write(t.rows[nextPrefix].prefix)
		nextPrefix++
	}
	out.Write(t.suffix)
	return out.Bytes(), nil
}

func entryParagraphAndSection(entry *roundTripBodyEntry) (domain.Block, domain.Block) {
	var paragraph domain.Block
	var section domain.Block
	if entry == nil {
		return paragraph, section
	}
	for _, block := range entry.blocks {
		switch {
		case block.Paragraph != nil:
			paragraph = block
		case block.SectionBreak != nil:
			section = block
		}
	}
	return paragraph, section
}

func (rt *roundTripMainDocument) sectionPropertiesForIdentity(section domain.Section) *roundTripSectionProperties {
	if rt == nil || section == nil {
		return nil
	}
	if section == rt.finalSectionIdentity {
		return rt.finalSection
	}
	for index := range rt.entries {
		entry := &rt.entries[index]
		if entry.section == nil {
			continue
		}
		_, sectionBlock := entryParagraphAndSection(entry)
		if sectionBlock.SectionBreak != nil && sectionBlock.SectionBreak.Section == section {
			return entry.section
		}
	}
	return nil
}

func replaceNamedElement(data []byte, name ooxmlmerge.QName, replacement []byte, namespaces map[string]string) ([]byte, error) {
	root, err := ooxmlmerge.Parse(data, namespaces)
	if err != nil {
		return nil, err
	}
	target := ooxmlmerge.First(root, name)
	if target == nil {
		return nil, fmt.Errorf("XML element %s was not found", name.Local)
	}
	return ooxmlmerge.ApplyEdits(data, []ooxmlmerge.Edit{{
		Start:       target.Start,
		End:         target.End,
		Replacement: replacement,
	}})
}

func firstNamedElement(data []byte, name ooxmlmerge.QName, namespaces map[string]string) ([]byte, error) {
	roots, err := ooxmlmerge.ParseAll(data, namespaces)
	if err != nil {
		return nil, err
	}
	for _, root := range roots {
		if target := ooxmlmerge.First(root, name); target != nil {
			return cloneBytes(target.Bytes()), nil
		}
	}
	return nil, fmt.Errorf("XML element %s was not found", name.Local)
}

func attachSectionProperties(paragraphXML, sectionXML []byte, namespaces map[string]string) ([]byte, error) {
	root, err := ooxmlmerge.Parse(paragraphXML, namespaces)
	if err != nil {
		return nil, err
	}
	for _, child := range root.Children {
		if !isWordprocessingNamespace(child.Name.Namespace) || child.Name.Local != "pPr" {
			continue
		}
		properties, err := ooxmlmerge.AppendChild(child, sectionXML)
		if err != nil {
			return nil, err
		}
		return ooxmlmerge.ApplyEdits(paragraphXML, []ooxmlmerge.Edit{{
			Start:       child.Start,
			End:         child.End,
			Replacement: properties,
		}})
	}
	prefix := root.Name.Prefix
	if prefix != "" {
		prefix += ":"
	}
	properties := append([]byte("<"+prefix+"pPr>"), sectionXML...)
	properties = append(properties, []byte("</"+prefix+"pPr>")...)
	return ooxmlmerge.ApplyEdits(paragraphXML, []ooxmlmerge.Edit{{
		Start:       root.ContentStart,
		End:         root.ContentStart,
		Replacement: properties,
	}})
}

func (entry *roundTripBodyEntry) composeSectionCarrier(ser *serializer.DocumentSerializer, current map[interface{}]bool) ([]byte, error) {
	if entry == nil || entry.section == nil {
		return nil, fmt.Errorf("section carrier is missing preserved section properties")
	}
	paragraphBlock, sectionBlock := entryParagraphAndSection(entry)
	paragraphPresent := paragraphBlock.Paragraph != nil && current[blockIdentity(paragraphBlock)]
	sectionPresent := sectionBlock.SectionBreak != nil && current[blockIdentity(sectionBlock)]

	if !sectionPresent {
		if !paragraphPresent {
			return nil, nil
		}
		fingerprint, err := blockGroupFingerprint([]domain.Block{paragraphBlock})
		if err != nil {
			return nil, err
		}
		if fingerprint != entry.paragraphSnapshot {
			return blockGroupBytesWithSerializerInContext(ser, []domain.Block{paragraphBlock}, entry.section.namespaces, entry.section.mainNamespace)
		}
		return replaceNamedElement(entry.raw, wordName("sectPr"), nil, entry.section.namespaces)
	}

	sectionFingerprint, err := blockGroupFingerprint([]domain.Block{sectionBlock})
	if err != nil {
		return nil, err
	}
	sectionBytes, err := blockGroupBytesWithSerializer(ser, []domain.Block{sectionBlock})
	if err != nil {
		return nil, err
	}
	currentSection, err := firstNamedElement(sectionBytes, wordName("sectPr"), roundTripNamespaces)
	if err != nil {
		return nil, err
	}
	mergedSection, err := entry.section.compose(currentSection)
	if err != nil {
		return nil, err
	}

	if paragraphBlock.Paragraph == nil {
		if sectionFingerprint == entry.sectionBreakSnapshot {
			return cloneBytes(entry.raw), nil
		}
		return replaceNamedElement(entry.raw, wordName("sectPr"), mergedSection, entry.section.namespaces)
	}
	if !paragraphPresent {
		sectionCarrier, err := renderGeneratedFragment(sectionBytes, entry.section.namespaces, entry.section.mainNamespace)
		if err != nil {
			return nil, err
		}
		return replaceNamedElement(sectionCarrier, wordName("sectPr"), mergedSection, entry.section.namespaces)
	}
	paragraphFingerprint, err := blockGroupFingerprint([]domain.Block{paragraphBlock})
	if err != nil {
		return nil, err
	}
	if paragraphFingerprint == entry.paragraphSnapshot {
		if sectionFingerprint == entry.sectionBreakSnapshot {
			return cloneBytes(entry.raw), nil
		}
		return replaceNamedElement(entry.raw, wordName("sectPr"), mergedSection, entry.section.namespaces)
	}
	paragraphBytes, err := blockGroupBytesWithSerializerInContext(ser, []domain.Block{paragraphBlock}, entry.section.namespaces, entry.section.mainNamespace)
	if err != nil {
		return nil, err
	}
	return attachSectionProperties(paragraphBytes, mergedSection, entry.section.namespaces)
}

func (rt *roundTripMainDocument) composeMovedFinalSection(ser *serializer.DocumentSerializer, block domain.Block) ([]byte, error) {
	data, err := blockGroupBytesWithSerializer(ser, []domain.Block{block})
	if err != nil {
		return nil, err
	}
	if rt.finalSection == nil {
		return renderGeneratedFragment(data, rt.namespaces, rt.mainNamespace)
	}
	currentSection, err := firstNamedElement(data, wordName("sectPr"), roundTripNamespaces)
	if err != nil {
		return nil, err
	}
	mergedSection, err := rt.finalSection.compose(currentSection)
	if err != nil {
		return nil, err
	}
	rendered, err := renderGeneratedFragment(data, rt.finalSection.namespaces, rt.mainNamespace)
	if err != nil {
		return nil, err
	}
	return replaceNamedElement(rendered, wordName("sectPr"), mergedSection, rt.finalSection.namespaces)
}

// composeRoundTripMainDocument returns nil when the document did not come
// from a source package. Otherwise it produces word/document.xml by combining
// original fragments with serialized changed nodes.
func (d *document) composeRoundTripMainDocument() ([]byte, error) {
	if d == nil || d.roundTripMain == nil {
		return nil, nil
	}
	rt := d.roundTripMain
	unchanged, err := rt.unchanged(d)
	if err != nil {
		return nil, err
	}
	if unchanged {
		return cloneBytes(rt.original), nil
	}
	composeSerializer := serializer.NewDocumentSerializer()
	composeSerializer.EnsureDrawingIDAtLeast(rt.maxDrawingID)

	current := make(map[interface{}]bool, len(d.blocks))
	for _, block := range d.blocks {
		if key := blockIdentity(block); key != nil {
			current[key] = true
		}
	}
	consumed := make(map[interface{}]bool, len(current))
	originalFinalIsCurrent := rt.finalSectionIdentity != nil && len(d.sections) > 0 && d.sections[len(d.sections)-1] == rt.finalSectionIdentity
	finalSectionMoved := rt.finalSectionIdentity != nil && !originalFinalIsCurrent
	var currentFinalSection domain.Section
	if len(d.sections) > 0 {
		currentFinalSection = d.sections[len(d.sections)-1]
	}
	currentFinalSource := rt.sectionPropertiesForIdentity(currentFinalSection)
	finalSectionPrefixConsumed := false

	var body bytes.Buffer
	for _, entry := range rt.entries {
		body.Write(entry.prefix)
		if len(entry.blocks) == 0 {
			body.Write(entry.raw)
			continue
		}
		if entry.section != nil {
			data, err := entry.composeSectionCarrier(composeSerializer, current)
			if err != nil {
				return nil, fmt.Errorf("compose embedded section carrier: %w", err)
			}
			body.Write(data)
			for _, block := range entry.blocks {
				key := blockIdentity(block)
				if key != nil && current[key] {
					consumed[key] = true
				}
			}
			continue
		}

		allPresent := true
		for _, block := range entry.blocks {
			if !current[blockIdentity(block)] {
				allPresent = false
				break
			}
		}
		if !allPresent {
			continue
		}

		if entry.table != nil {
			data, err := entry.table.compose(composeSerializer)
			if err != nil {
				return nil, fmt.Errorf("compose table: %w", err)
			}
			body.Write(data)
		} else {
			fingerprint, err := blockGroupFingerprint(entry.blocks)
			if err != nil {
				return nil, fmt.Errorf("compare body block: %w", err)
			}
			if fingerprint == entry.snapshot {
				body.Write(entry.raw)
			} else {
				data, err := blockGroupBytesWithSerializerInContext(composeSerializer, entry.blocks, rt.namespaces, rt.mainNamespace)
				if err != nil {
					return nil, fmt.Errorf("serialize body block: %w", err)
				}
				body.Write(data)
			}
		}
		for _, block := range entry.blocks {
			consumed[blockIdentity(block)] = true
		}
	}

	for _, block := range d.blocks {
		key := blockIdentity(block)
		if key == nil || consumed[key] {
			continue
		}
		var data []byte
		var err error
		if finalSectionMoved && block.SectionBreak != nil && block.SectionBreak.Section == rt.finalSectionIdentity {
			body.Write(rt.finalSectionPrefix)
			finalSectionPrefixConsumed = true
			data, err = rt.composeMovedFinalSection(composeSerializer, block)
		} else {
			data, err = blockGroupBytesWithSerializerInContext(composeSerializer, []domain.Block{block}, rt.namespaces, rt.mainNamespace)
		}
		if err != nil {
			return nil, fmt.Errorf("serialize new body block: %w", err)
		}
		body.Write(data)
	}

	if !finalSectionPrefixConsumed {
		body.Write(rt.finalSectionPrefix)
	}
	finalSection, err := finalSectionFingerprint(d)
	if err != nil {
		return nil, fmt.Errorf("compare final section: %w", err)
	}
	if originalFinalIsCurrent && finalSection == rt.finalSectionSnapshot {
		body.Write(rt.finalSectionRaw)
	} else {
		data, err := finalSectionBytes(d)
		if err != nil {
			return nil, fmt.Errorf("serialize final section: %w", err)
		}
		if currentFinalSource != nil {
			data, err = currentFinalSource.compose(data)
			if err != nil {
				return nil, fmt.Errorf("merge final section: %w", err)
			}
		} else {
			data, err = renderGeneratedFragment(data, rt.namespaces, rt.mainNamespace)
			if err != nil {
				return nil, fmt.Errorf("render final section namespaces: %w", err)
			}
		}
		body.Write(data)
	}
	body.Write(rt.bodyTail)

	prefix := rt.prefix
	background, err := documentBackgroundFingerprint(d)
	if err != nil {
		return nil, fmt.Errorf("compare document background: %w", err)
	}
	if background != rt.backgroundSnapshot {
		prefix, err = rt.prefixWithCurrentBackground(d)
		if err != nil {
			return nil, fmt.Errorf("compose document background: %w", err)
		}
	}

	var out bytes.Buffer
	out.Write(prefix)
	out.Write(body.Bytes())
	out.Write(rt.suffix)
	balanced, err := removeUnbalancedRangeMarkers(rt.original, out.Bytes())
	if err != nil {
		return nil, err
	}
	return ensureRequiredDocumentNamespaces(balanced), nil
}

type rangeMarkerOccurrence struct {
	key   string
	side  int
	start int
	end   int
	name  xml.Name
}

func rangeMarker(start xml.StartElement) (string, int, bool) {
	if !isWordprocessingNamespace(start.Name.Space) {
		return "", 0, false
	}
	local := start.Name.Local
	side := 0
	base := ""
	switch {
	case strings.HasSuffix(local, "Start"):
		base = strings.TrimSuffix(local, "Start")
	case strings.HasSuffix(local, "End"):
		side = 1
		base = strings.TrimSuffix(local, "End")
	default:
		return "", 0, false
	}
	if base == "" {
		return "", 0, false
	}
	for _, attr := range start.Attr {
		if isWordprocessingNamespace(attr.Name.Space) && attr.Name.Local == "id" && attr.Value != "" {
			return base + "\x00" + attr.Value, side, true
		}
	}
	return "", 0, false
}

// scanRangeMarkers locates modeled range endpoints and counts each side by
// range type and ID.
func scanRangeMarkers(data []byte) ([]*rangeMarkerOccurrence, map[string][2]int, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var occurrences []*rangeMarkerOccurrence
	var open []*rangeMarkerOccurrence
	counts := make(map[string][2]int)

	for {
		tokenStart := int(dec.InputOffset())
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, err
		}
		switch token := tok.(type) {
		case xml.StartElement:
			key, side, ok := rangeMarker(token)
			if !ok {
				continue
			}
			occurrence := &rangeMarkerOccurrence{
				key:   key,
				side:  side,
				start: tokenStart,
				name:  token.Name,
			}
			occurrences = append(occurrences, occurrence)
			open = append(open, occurrence)
			count := counts[key]
			count[side]++
			counts[key] = count
		case xml.EndElement:
			if len(open) == 0 {
				continue
			}
			last := open[len(open)-1]
			if token.Name == last.name {
				last.end = int(dec.InputOffset())
				open = open[:len(open)-1]
			}
		}
	}
	return occurrences, counts, nil
}

// removeUnbalancedRangeMarkers prevents a localized rewrite from retaining
// only one endpoint of a range that was balanced in the source. A range that
// was already incomplete is opaque source XML and survives unrelated edits.
func removeUnbalancedRangeMarkers(original, data []byte) ([]byte, error) {
	_, originalCounts, err := scanRangeMarkers(original)
	if err != nil {
		return nil, err
	}
	occurrences, counts, err := scanRangeMarkers(data)
	if err != nil {
		return nil, err
	}

	var removals []*rangeMarkerOccurrence
	for _, occurrence := range occurrences {
		count := counts[occurrence.key]
		originalCount := originalCounts[occurrence.key]
		if originalCount[0] == originalCount[1] && count[0] != count[1] && occurrence.end > occurrence.start {
			removals = append(removals, occurrence)
		}
	}
	if len(removals) == 0 {
		return data, nil
	}
	sort.Slice(removals, func(i, j int) bool { return removals[i].start > removals[j].start })
	out := cloneBytes(data)
	for _, removal := range removals {
		out = append(out[:removal.start], out[removal.end:]...)
	}
	return out, nil
}

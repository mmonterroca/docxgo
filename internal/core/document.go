// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

// Package core provides concrete implementations of domain interfaces for docxgo v2.
//
// This package contains the core document model implementations including:
// - Document: The main document structure
// - Paragraph: Paragraph implementation with formatting
// - Run: Text run implementation with character formatting
// - Table: Table implementation with cells and rows
// - Section: Section implementation with page settings
// - Image: Image embedding and positioning
// - Field: Field implementation (TOC, page numbers, etc.)
//
// These implementations handle the business logic and coordinate with
// internal managers (ID generation, relationships, media, styles) and
// serialization to XML structures.
package core

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/manager"
	"github.com/mmonterroca/docxgo/v2/internal/serializer"
	"github.com/mmonterroca/docxgo/v2/internal/writer"
	xmlstructs "github.com/mmonterroca/docxgo/v2/internal/xml"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
	"github.com/mmonterroca/docxgo/v2/pkg/errors"
)

// document implements the domain.Document interface.
type document struct {
	paragraphs      []domain.Paragraph
	tables          []domain.Table
	sections        []domain.Section
	blocks          []domain.Block
	metadata        *domain.Metadata
	idGen           *manager.IDGenerator
	relManager      *manager.RelationshipManager
	mediaManager    *manager.MediaManager
	styleManager    domain.StyleManager
	headerCount     int
	footerCount     int
	activeSection   *docxSection
	numberingPart   []byte
	numberingTarget string
	backgroundColor *domain.Color
	language        *domain.Language
	defaultFont     *string
	defaultFontSize *int

	// Preserved parts for round-trip operations (read-modify-write).
	// When set, these parts are written verbatim to preserve original content.
	preservedStylesPart   []byte            // Original styles.xml
	preservedHeaders      map[string][]byte // Original headers (e.g., "word/header1.xml" -> bytes)
	preservedFooters      map[string][]byte // Original footers (e.g., "word/footer1.xml" -> bytes)
	preservedHeaderRels   map[string][]byte // Original word/_rels/headerN.xml.rels, keyed by archive path
	preservedFooterRels   map[string][]byte // Original word/_rels/footerN.xml.rels, keyed by archive path
	preservedDocRels      []byte            // Original word/_rels/document.xml.rels
	preservedContentTypes []byte            // Original [Content_Types].xml
	preservedAdditional   map[string][]byte // Additional parts (comments, footnotes, customXml, etc.)
	preservedThemes       map[string][]byte // Original theme parts
	preservedFontTable    []byte            // Original fontTable.xml
	preservedSettings     []byte            // Original settings.xml
	preservedWebSettings  []byte            // Original webSettings.xml
	preservedCustomProps  []byte            // Original docProps/custom.xml
	preservedRootRels     []byte            // Original _rels/.rels
}

// NewDocument creates a new Document.
func NewDocument() domain.Document {
	doc := newBareDocument()

	// Ensure core document relationships exist (styles, fonts, theme). Safe
	// to assign rId1.. here because nothing else has claimed an ID yet.
	doc.ensureDefaultRelationships()

	return doc
}

// NewDocumentForReconstruction creates a new Document for the reader to
// hydrate from an existing .docx package.
//
// Unlike NewDocument, it does NOT pre-populate the base relationships
// (styles, fontTable, theme, settings, webSettings) -- those must wait until
// after the source document's own relationships have been registered via
// RegisterExistingRelationship. RegisterExisting is a no-op when a
// relationship already exists under the given ID, so if NewDocument's rId1
// (styles) collided with the source file's own rId1 (which, in a real Word
// document, has no reason to be styles -- IDs are assigned in whatever order
// the authoring application added them), the source's real rId1 relationship
// would be silently dropped instead of registered, corrupting every part
// that references it by that ID (e.g. a header's w:headerReference). See
// issue #101's PR review.
//
// ensureDefaultRelationships still runs -- Document.WriteTo calls it again
// before serialization, by which point the source relationships are already
// registered under their real IDs, so its target-based deduplication
// correctly recognizes an already-present styles.xml/etc. relationship
// instead of creating a colliding one from scratch.
func NewDocumentForReconstruction() domain.Document {
	return newBareDocument()
}

// newBareDocument builds a *document with no relationships registered yet.
func newBareDocument() *document {
	idGen := manager.NewIDGenerator()
	relManager := manager.NewRelationshipManager(idGen)
	return &document{
		paragraphs:   make([]domain.Paragraph, 0, constants.DefaultParagraphCapacity),
		tables:       make([]domain.Table, 0, constants.DefaultTableCapacity),
		sections:     make([]domain.Section, 0, 1),
		blocks:       make([]domain.Block, 0, constants.DefaultParagraphCapacity),
		metadata:     &domain.Metadata{},
		idGen:        idGen,
		relManager:   relManager,
		mediaManager: manager.NewMediaManager(idGen),
		styleManager: manager.NewStyleManager(),
	}
}

// ensureActiveSection guarantees the document has a current section and returns it.
func (d *document) ensureActiveSection() (*docxSection, error) {
	if len(d.sections) == 0 {
		section := NewSection(d.relManager, d.idGen, d.mediaManager)
		coreSection, ok := section.(*docxSection)
		if !ok {
			return nil, errors.InvalidState("Document.ensureActiveSection", "unexpected section implementation type")
		}
		d.sections = append(d.sections, section)
		d.activeSection = coreSection
	}

	if d.activeSection == nil {
		last := d.sections[len(d.sections)-1]
		coreSection, ok := last.(*docxSection)
		if !ok {
			return nil, errors.InvalidState("Document.ensureActiveSection", "unexpected section implementation type")
		}
		d.activeSection = coreSection
	}

	return d.activeSection, nil
}

// AddParagraph adds a new paragraph to the document.
func (d *document) AddParagraph() (domain.Paragraph, error) {
	if _, err := d.ensureActiveSection(); err != nil {
		return nil, err
	}

	id := d.idGen.NextParagraphID()
	para := NewParagraph(id, d.idGen, d.relManager, d.mediaManager)
	d.paragraphs = append(d.paragraphs, para)
	d.blocks = append(d.blocks, domain.Block{Paragraph: para})
	return para, nil
}

// AddTable adds a new table with the specified dimensions.
func (d *document) AddTable(rows, cols int) (domain.Table, error) {
	if _, err := d.ensureActiveSection(); err != nil {
		return nil, err
	}

	if rows < constants.MinTableRows || rows > constants.MaxTableRows {
		return nil, errors.InvalidArgument("Document.AddTable", "rows", rows,
			"rows must be between 1 and 1000")
	}
	if cols < constants.MinTableCols || cols > constants.MaxTableCols {
		return nil, errors.InvalidArgument("Document.AddTable", "cols", cols,
			"columns must be between 1 and 63")
	}

	id := d.idGen.NextTableID()
	table := NewTable(id, rows, cols, d.idGen, d.relManager, d.mediaManager)
	d.tables = append(d.tables, table)
	d.blocks = append(d.blocks, domain.Block{Table: table})
	return table, nil
}

// AddSection adds a new section to the document using a next-page break.
func (d *document) AddSection() (domain.Section, error) {
	return d.AddSectionWithBreak(domain.SectionBreakTypeNextPage)
}

// AddSectionWithBreak adds a new section specifying the section break behavior.
func (d *document) AddSectionWithBreak(breakType domain.SectionBreakType) (domain.Section, error) {
	if breakType < domain.SectionBreakTypeNextPage || breakType > domain.SectionBreakTypeOddPage {
		return nil, errors.InvalidArgument("Document.AddSectionWithBreak", "breakType", breakType,
			"section break type must be between NextPage and OddPage")
	}

	currentSection, err := d.ensureActiveSection()
	if err != nil {
		return nil, err
	}

	d.blocks = append(d.blocks, domain.Block{
		SectionBreak: &domain.SectionBreak{
			Section: currentSection,
			Type:    breakType,
		},
	})

	newSection := NewSection(d.relManager, d.idGen, d.mediaManager)
	coreSection, ok := newSection.(*docxSection)
	if !ok {
		return nil, errors.InvalidState("Document.AddSectionWithBreak", "unexpected section implementation type")
	}

	d.sections = append(d.sections, newSection)
	d.activeSection = coreSection

	return newSection, nil
}

// RemoveLastSection undoes the most recent AddSectionWithBreak call, but
// only if it is still the very last thing added -- i.e. section is still
// d.activeSection and nothing else was appended to d.blocks since. It exists
// for callers (namely the CLI/RPC document.section.add handler) that build a
// section up across several steps (page size, margins, headers, footers,
// ...) against a document that's already shared with other in-flight state:
// AddSectionWithBreak attaches the new section immediately, so if a later
// step in that sequence fails, the caller needs a way to undo the whole
// section rather than leaving a half-configured one permanently in
// Sections(). Reports whether it actually removed anything, so a caller that
// can't be sure the section is still last-added can check before assuming
// the document is back to its prior state.
func (d *document) RemoveLastSection(section domain.Section) bool {
	if len(d.sections) == 0 || d.sections[len(d.sections)-1] != section {
		return false
	}
	if len(d.blocks) == 0 {
		return false
	}
	lastBlock := d.blocks[len(d.blocks)-1]
	if lastBlock.SectionBreak == nil || lastBlock.SectionBreak.Section == nil {
		return false
	}

	prevSection, ok := lastBlock.SectionBreak.Section.(*docxSection)
	if !ok {
		return false
	}

	d.sections = d.sections[:len(d.sections)-1]
	d.blocks = d.blocks[:len(d.blocks)-1]
	d.activeSection = prevSection
	return true
}

// AddPageBreak adds a page break to the document.
func (d *document) AddPageBreak() error {
	// Create a new paragraph
	para, err := d.AddParagraph()
	if err != nil {
		return err
	}

	// Add a run with a page break
	run, err := para.AddRun()
	if err != nil {
		return err
	}

	return run.AddBreak(domain.BreakTypePage)
}

// DefaultSection returns the default (first) section of the document.
func (d *document) DefaultSection() (domain.Section, error) {
	_, err := d.ensureActiveSection()
	if err != nil {
		return nil, err
	}

	// ensureActiveSection always keeps sections slice populated.
	return d.sections[0], nil
}

// Paragraphs returns all paragraphs in the document.
func (d *document) Paragraphs() []domain.Paragraph {
	// Return a copy to prevent external modification
	paras := make([]domain.Paragraph, len(d.paragraphs))
	copy(paras, d.paragraphs)
	return paras
}

// Tables returns all tables in the document.
func (d *document) Tables() []domain.Table {
	tables := make([]domain.Table, len(d.tables))
	copy(tables, d.tables)
	return tables
}

// Sections returns all sections in the document.
func (d *document) Sections() []domain.Section {
	sections := make([]domain.Section, len(d.sections))
	copy(sections, d.sections)
	return sections
}

// Blocks returns all top-level document content in insertion order.
func (d *document) Blocks() []domain.Block {
	blocks := make([]domain.Block, len(d.blocks))
	copy(blocks, d.blocks)
	return blocks
}

// generateHeadingBookmarks generates bookmarks for all headings in the document.
// This is required for Table of Contents (TOC) fields to work properly.
// Bookmarks are named _Toc{sequential_number} and only applied to paragraphs with Heading styles.
func (d *document) generateHeadingBookmarks() {
	bookmarkCounter := 0

	for _, para := range d.paragraphs {
		// Type assert to access internal paragraph methods
		if p, ok := para.(*paragraph); ok {
			styleName := p.StyleName()

			// Check if this paragraph has a Heading style
			if strings.HasPrefix(styleName, "Heading") {
				bookmarkID := fmt.Sprintf("%d", bookmarkCounter)
				bookmarkName := fmt.Sprintf("_Toc%d", bookmarkCounter)
				p.SetBookmark(bookmarkID, bookmarkName)
				bookmarkCounter++
			}
		}
	}
}

// prepareHeaderFooterRelationships ensures that every header/footer defined in the
// document has an associated relationship and target part name within the DOCX
// package. This must run before serialization so both section references and the
// document relationships list are consistent.
func (d *document) prepareHeaderFooterRelationships() {
	for _, sec := range d.sections {
		coreSection, ok := sec.(*docxSection)
		if !ok {
			continue
		}

		coreSection.mu.Lock()

		for _, header := range coreSection.headers {
			if header == nil {
				continue
			}

			if header.TargetPath() == "" {
				d.headerCount++
				target := fmt.Sprintf("header%d.xml", d.headerCount)
				header.setRelationship(header.RelationshipID(), target)
			}

			if header.RelationshipID() == "" {
				if relID, err := d.relManager.AddHeader(header.TargetPath()); err == nil {
					header.setRelationship(relID, header.TargetPath())
				}
			}
		}

		for _, footer := range coreSection.footers {
			if footer == nil {
				continue
			}

			if footer.TargetPath() == "" {
				d.footerCount++
				target := fmt.Sprintf("footer%d.xml", d.footerCount)
				footer.setRelationship(footer.RelationshipID(), target)
			}

			if footer.RelationshipID() == "" {
				if relID, err := d.relManager.AddFooter(footer.TargetPath()); err == nil {
					footer.setRelationship(relID, footer.TargetPath())
				}
			}
		}

		coreSection.mu.Unlock()
	}
}

// ensureDefaultRelationships guarantees that the DOCX package contains the
// required relationships for styles, fonts, and theme assets. Without these
// entries Word falls back to implicit defaults and style assignments appear as
// "Normal", which breaks features such as the Table of Contents.
func (d *document) ensureDefaultRelationships() {
	if d == nil || d.relManager == nil {
		return
	}

	// Track existing relationship targets to avoid duplicates when called
	// multiple times (e.g. SaveAs after WriteTo).
	existing := make(map[string]bool)
	for _, rel := range d.relManager.All() {
		existing[rel.Target] = true
	}

	baseRels := []struct {
		relType string
		target  string
	}{
		{constants.RelTypeStyles, "styles.xml"},
		{constants.RelTypeFontTable, "fontTable.xml"},
		{constants.RelTypeTheme, "theme/theme1.xml"},
		{constants.RelTypeSettings, "settings.xml"},
		{constants.RelTypeWebSettings, "webSettings.xml"},
	}

	for _, rel := range baseRels {
		if existing[rel.target] {
			continue
		}

		// Ignore the error because the inputs are fixed and validated. In the
		// unlikely event of a failure we still prefer to continue writing the
		// document instead of aborting.
		_, _ = d.relManager.Add(rel.relType, rel.target, "Internal")
	}
}

// docRelsNeedsRegeneration reports whether d.relManager holds a relationship
// ID that is not present in the preserved word/_rels/document.xml.rels bytes.
// Those bytes are written back verbatim on a round-tripped document (see
// WriteTo); if something was added to the relationship manager after the
// document was opened -- e.g. AddHyperlink -- writing the preserved bytes
// unchanged would leave a dangling r:id in document.xml. See issue #101.
func (d *document) docRelsNeedsRegeneration() bool {
	if len(d.preservedDocRels) == 0 {
		return false
	}

	var preserved xmlstructs.Relationships
	if err := xml.Unmarshal(d.preservedDocRels, &preserved); err != nil {
		// Malformed preserved bytes: keep the existing verbatim behavior
		// rather than risk mangling a part that couldn't be parsed.
		return false
	}

	known := make(map[string]bool, len(preserved.Relationships))
	for _, rel := range preserved.Relationships {
		if rel != nil {
			known[rel.ID] = true
		}
	}

	for _, rel := range d.relManager.All() {
		if rel != nil && !known[rel.ID] {
			return true
		}
	}

	return false
}

// WriteTo writes the document to the provided writer in .docx format.
func (d *document) WriteTo(w io.Writer) (int64, error) {
	if len(d.sections) == 0 {
		if _, err := d.DefaultSection(); err != nil {
			return 0, errors.Wrap(err, "Document.WriteTo")
		}
	}

	// Generate bookmarks for headings (needed for TOC)
	d.generateHeadingBookmarks()

	// Ensure headers and footers have relationships/targets before serialization
	d.prepareHeaderFooterRelationships()

	// Ensure required base relationships are present before serialization
	d.ensureDefaultRelationships()

	// Serialize domain objects to XML structures
	ser := serializer.NewDocumentSerializer()
	xmlDoc := ser.SerializeDocument(d)
	headers, footers := ser.SerializeSectionParts(d)

	// Create ZIP writer
	zipWriter := writer.NewZipWriter(w)
	zipWriter.SetLanguage(d.language)
	defer func() {
		if err := zipWriter.Close(); err != nil {
			// Log error but don't override return value as document may have been partially written
			_ = err
		}
	}()

	// Build relationships from relationship manager
	rels := d.relManager.ToXML()

	// Serialize metadata
	coreProps := ser.SerializeCoreProperties(d.metadata)
	appProps := ser.SerializeAppProperties(d)

	// Serialize styles (used only if no preserved styles are available)
	styles := ser.SerializeStyles(d.styleManager, d.language, d.defaultFont, d.defaultFontSize)

	mediaFiles := d.mediaManager.All()

	// Write document structure
	var numberingPart *writer.NumberingPart
	if len(d.numberingPart) > 0 {
		numberingPart = &writer.NumberingPart{
			Data:   d.numberingPart,
			Target: d.numberingTarget,
		}
	}

	// Build writer.PreservedParts from core.PreservedParts if available
	var writerPreserved *writer.PreservedParts
	if corePreserved := d.GetPreservedParts(); corePreserved != nil {
		docRels := corePreserved.DocRels
		if d.docRelsNeedsRegeneration() {
			// A relationship was added since this document was opened (e.g.
			// AddHyperlink). Leave DocRels empty so zip.go's non-round-trip
			// branch regenerates word/_rels/document.xml.rels from
			// d.relManager -- which was seeded with every original
			// relationship on open, so this is lossless for the unchanged
			// ones and adds the new one(s).
			docRels = nil
		}

		writerPreserved = &writer.PreservedParts{
			Headers:          corePreserved.Headers,
			Footers:          corePreserved.Footers,
			HeaderRels:       corePreserved.HeaderRels,
			FooterRels:       corePreserved.FooterRels,
			DocRels:          docRels,
			ContentTypes:     corePreserved.ContentTypes,
			Additional:       corePreserved.Additional,
			Themes:           corePreserved.Themes,
			FontTable:        corePreserved.FontTable,
			Settings:         corePreserved.Settings,
			WebSettings:      corePreserved.WebSettings,
			CustomProperties: corePreserved.CustomProperties,
			RootRels:         corePreserved.RootRels,
		}
	}

	// Use preserved styles and parts if available (from reading an existing document)
	if err := zipWriter.WriteDocument(xmlDoc, rels, coreProps, appProps, styles, mediaFiles, headers, footers, numberingPart, d.preservedStylesPart, writerPreserved); err != nil {
		return 0, errors.WrapWithCode(err, errors.ErrCodeIO, "Document.WriteTo")
	}

	// Get byte count from writer if available
	// For now, return 0 as ZipWriter doesn't track total bytes
	// This could be enhanced by wrapping the writer with a counting writer
	return 0, nil
}

// SaveAs saves the document to the specified file path.
func (d *document) SaveAs(path string) error {
	if path == "" {
		return errors.InvalidArgument("Document.SaveAs", "path", path, "path cannot be empty")
	}

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeIO, "Document.SaveAs")
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = errors.WrapWithCode(closeErr, errors.ErrCodeIO, "Document.SaveAs")
		}
	}()

	// Write document to file
	_, err = d.WriteTo(file)
	if err != nil {
		return errors.Wrap(err, "Document.SaveAs")
	}

	return nil
}

// Validate checks if the document structure is valid.
func (d *document) Validate() error {
	// Basic validation
	if len(d.paragraphs) == 0 && len(d.tables) == 0 {
		return errors.InvalidState("Document.Validate", "document is empty")
	}

	// Validate each paragraph
	for i, para := range d.paragraphs {
		if para == nil {
			return errors.InvalidState("Document.Validate",
				"paragraph at index "+string(rune(i))+" is nil")
		}
	}

	// Validate each table
	for i, table := range d.tables {
		if table == nil {
			return errors.InvalidState("Document.Validate",
				"table at index "+string(rune(i))+" is nil")
		}
	}

	return nil
}

// Metadata returns the document's metadata.
func (d *document) Metadata() *domain.Metadata {
	return d.metadata
}

// SetMetadata updates the document's metadata.
func (d *document) SetMetadata(meta *domain.Metadata) error {
	if meta == nil {
		return errors.InvalidArgument("Document.SetMetadata", "meta", meta, "metadata cannot be nil")
	}
	d.metadata = meta
	return nil
}

// SetBackgroundColor sets the document page background color.
func (d *document) SetBackgroundColor(color domain.Color) error {
	if d == nil {
		return errors.InvalidState("Document.SetBackgroundColor", "document is nil")
	}
	copyColor := color
	d.backgroundColor = &copyColor
	return nil
}

// BackgroundColor returns the configured page background color.
func (d *document) BackgroundColor() (domain.Color, bool) {
	if d == nil || d.backgroundColor == nil {
		return domain.Color{}, false
	}
	return *d.backgroundColor, true
}

// SetLanguage sets the document's default proofing language. It returns an
// error on a document opened via OpenDocument/OpenDocumentFromBytes/
// OpenDocumentFromReader whose styles.xml or settings.xml were preserved for
// round-trip fidelity: WriteTo writes those parts verbatim, so a language set
// here would silently never reach the saved file. Only documents created with
// NewDocument (including via NewDocumentBuilder) support SetLanguage.
func (d *document) SetLanguage(lang *domain.Language) error {
	if d == nil {
		return errors.InvalidState("Document.SetLanguage", "document is nil")
	}
	if len(d.preservedStylesPart) > 0 || len(d.preservedSettings) > 0 {
		return errors.InvalidState("Document.SetLanguage",
			"cannot set the proofing language on a document whose styles.xml/settings.xml are preserved from round-trip; only documents created with NewDocument support SetLanguage")
	}
	if lang != nil && lang.Val == "" && lang.EastAsia == "" && lang.Bidi == "" {
		return errors.InvalidArgument("Document.SetLanguage", "lang", lang, "at least one of Val, EastAsia, or Bidi must be set")
	}
	if lang == nil {
		d.language = nil
		return nil
	}
	langCopy := *lang
	d.language = &langCopy
	return nil
}

// Language returns a copy of the document's default proofing language, or
// nil if unset. Mutating the returned value has no effect on the document.
func (d *document) Language() *domain.Language {
	if d == nil || d.language == nil {
		return nil
	}
	langCopy := *d.language
	return &langCopy
}

// SetLanguageRaw hydrates the document's language field directly, bypassing
// SetLanguage's round-trip guard. Used exclusively by the reader package
// during OpenDocument to reflect a language already present in the source
// file's styles.xml — not part of domain.Document, reached only via the
// reader's own type-assertion (mirrors SetPreservedStylesPart/SetNumberingPart).
func (d *document) SetLanguageRaw(lang *domain.Language) {
	if d == nil {
		return
	}
	if lang == nil {
		d.language = nil
		return
	}
	langCopy := *lang
	d.language = &langCopy
}

// SetDefaultFont sets the document's default run font, written to
// styles.xml's w:docDefaults/w:rPrDefault. Not part of domain.Document —
// NewDocumentBuilder always starts from a fresh NewDocument(), which has no
// preserved styles.xml to conflict with, so there's no round-trip guard to
// write here the way SetLanguage needs one. Reached only via
// NewDocumentBuilder's own type-assertion (mirrors SetLanguageRaw).
func (d *document) SetDefaultFont(name string) error {
	if d == nil {
		return errors.InvalidState("Document.SetDefaultFont", "document is nil")
	}
	if name == "" {
		return errors.InvalidArgument("Document.SetDefaultFont", "name", name, "font name must not be empty")
	}
	d.defaultFont = &name
	return nil
}

// SetDefaultFontSize sets the document's default run font size in
// half-points, written to the same w:docDefaults/w:rPrDefault as
// SetDefaultFont. See SetDefaultFont regarding why this isn't part of
// domain.Document.
func (d *document) SetDefaultFontSize(halfPoints int) error {
	if d == nil {
		return errors.InvalidState("Document.SetDefaultFontSize", "document is nil")
	}
	if halfPoints <= 0 {
		return errors.InvalidArgument("Document.SetDefaultFontSize", "halfPoints", halfPoints,
			"font size must be a positive number of half-points")
	}
	d.defaultFontSize = &halfPoints
	return nil
}

// StyleManager returns the style manager for this document.
func (d *document) StyleManager() domain.StyleManager {
	return d.styleManager
}

func (d *document) RegisterExistingRelationship(id, relType, target, targetMode string) error {
	if d == nil || d.relManager == nil {
		return errors.InvalidState("Document.RegisterExistingRelationship", "relationship manager not initialized")
	}
	return d.relManager.RegisterExisting(id, relType, target, targetMode)
}

func (d *document) SetNumberingPart(data []byte, target string) {
	if d == nil || len(data) == 0 {
		return
	}
	copied := make([]byte, len(data))
	copy(copied, data)
	d.numberingPart = copied
	d.numberingTarget = normalizeNumberingTarget(target)
}

func (d *document) NumberingPartInfo() ([]byte, string) {
	if d == nil || len(d.numberingPart) == 0 {
		return nil, ""
	}
	copied := make([]byte, len(d.numberingPart))
	copy(copied, d.numberingPart)
	return copied, d.numberingTarget
}

// SetPreservedStylesPart stores the original styles.xml bytes from a read document.
// When set, these bytes are written verbatim to preserve the original document styles.
func (d *document) SetPreservedStylesPart(data []byte) {
	if d == nil || len(data) == 0 {
		return
	}
	copied := make([]byte, len(data))
	copy(copied, data)
	d.preservedStylesPart = copied
}

// PreservedStylesPartInfo returns the stored original styles.xml bytes, if any.
func (d *document) PreservedStylesPartInfo() []byte {
	if d == nil || len(d.preservedStylesPart) == 0 {
		return nil
	}
	copied := make([]byte, len(d.preservedStylesPart))
	copy(copied, d.preservedStylesPart)
	return copied
}

// PreservedParts holds all preserved parts from a read document for round-trip operations.
type PreservedParts struct {
	Headers          map[string][]byte
	Footers          map[string][]byte
	HeaderRels       map[string][]byte // word/_rels/headerN.xml.rels, keyed by archive path
	FooterRels       map[string][]byte // word/_rels/footerN.xml.rels, keyed by archive path
	DocRels          []byte
	ContentTypes     []byte
	Additional       map[string][]byte
	Themes           map[string][]byte
	FontTable        []byte
	Settings         []byte
	WebSettings      []byte
	CustomProperties []byte // docProps/custom.xml
	RootRels         []byte // _rels/.rels
}

// SetPreservedParts stores all preserved parts from a read document.
func (d *document) SetPreservedParts(parts *PreservedParts) {
	if d == nil || parts == nil {
		return
	}

	// Copy headers
	if len(parts.Headers) > 0 {
		d.preservedHeaders = make(map[string][]byte, len(parts.Headers))
		for k, v := range parts.Headers {
			copied := make([]byte, len(v))
			copy(copied, v)
			d.preservedHeaders[k] = copied
		}
	}

	// Copy footers
	if len(parts.Footers) > 0 {
		d.preservedFooters = make(map[string][]byte, len(parts.Footers))
		for k, v := range parts.Footers {
			copied := make([]byte, len(v))
			copy(copied, v)
			d.preservedFooters[k] = copied
		}
	}

	// Copy header relationships
	if len(parts.HeaderRels) > 0 {
		d.preservedHeaderRels = make(map[string][]byte, len(parts.HeaderRels))
		for k, v := range parts.HeaderRels {
			copied := make([]byte, len(v))
			copy(copied, v)
			d.preservedHeaderRels[k] = copied
		}
	}

	// Copy footer relationships
	if len(parts.FooterRels) > 0 {
		d.preservedFooterRels = make(map[string][]byte, len(parts.FooterRels))
		for k, v := range parts.FooterRels {
			copied := make([]byte, len(v))
			copy(copied, v)
			d.preservedFooterRels[k] = copied
		}
	}

	// Copy document relationships
	if len(parts.DocRels) > 0 {
		d.preservedDocRels = make([]byte, len(parts.DocRels))
		copy(d.preservedDocRels, parts.DocRels)
	}

	// Copy content types
	if len(parts.ContentTypes) > 0 {
		d.preservedContentTypes = make([]byte, len(parts.ContentTypes))
		copy(d.preservedContentTypes, parts.ContentTypes)
	}

	// Copy additional parts
	if len(parts.Additional) > 0 {
		d.preservedAdditional = make(map[string][]byte, len(parts.Additional))
		for k, v := range parts.Additional {
			copied := make([]byte, len(v))
			copy(copied, v)
			d.preservedAdditional[k] = copied
		}
	}

	// Copy themes
	if len(parts.Themes) > 0 {
		d.preservedThemes = make(map[string][]byte, len(parts.Themes))
		for k, v := range parts.Themes {
			copied := make([]byte, len(v))
			copy(copied, v)
			d.preservedThemes[k] = copied
		}
	}

	// Copy font table
	if len(parts.FontTable) > 0 {
		d.preservedFontTable = make([]byte, len(parts.FontTable))
		copy(d.preservedFontTable, parts.FontTable)
	}

	// Copy settings
	if len(parts.Settings) > 0 {
		d.preservedSettings = make([]byte, len(parts.Settings))
		copy(d.preservedSettings, parts.Settings)
	}

	// Copy web settings
	if len(parts.WebSettings) > 0 {
		d.preservedWebSettings = make([]byte, len(parts.WebSettings))
		copy(d.preservedWebSettings, parts.WebSettings)
	}

	// Copy custom properties
	if len(parts.CustomProperties) > 0 {
		d.preservedCustomProps = make([]byte, len(parts.CustomProperties))
		copy(d.preservedCustomProps, parts.CustomProperties)
	}

	// Copy root relationships
	if len(parts.RootRels) > 0 {
		d.preservedRootRels = make([]byte, len(parts.RootRels))
		copy(d.preservedRootRels, parts.RootRels)
	}
}

// GetPreservedParts returns all preserved parts for writing.
func (d *document) GetPreservedParts() *PreservedParts {
	if d == nil {
		return nil
	}

	// Only return if we have any preserved parts
	if len(d.preservedHeaders) == 0 && len(d.preservedFooters) == 0 &&
		len(d.preservedHeaderRels) == 0 && len(d.preservedFooterRels) == 0 &&
		len(d.preservedDocRels) == 0 && len(d.preservedContentTypes) == 0 &&
		len(d.preservedAdditional) == 0 && len(d.preservedThemes) == 0 &&
		len(d.preservedFontTable) == 0 && len(d.preservedSettings) == 0 &&
		len(d.preservedWebSettings) == 0 && len(d.preservedCustomProps) == 0 &&
		len(d.preservedRootRels) == 0 {
		return nil
	}

	return &PreservedParts{
		Headers:          d.preservedHeaders,
		Footers:          d.preservedFooters,
		HeaderRels:       d.preservedHeaderRels,
		FooterRels:       d.preservedFooterRels,
		DocRels:          d.preservedDocRels,
		ContentTypes:     d.preservedContentTypes,
		Additional:       d.preservedAdditional,
		Themes:           d.preservedThemes,
		FontTable:        d.preservedFontTable,
		Settings:         d.preservedSettings,
		WebSettings:      d.preservedWebSettings,
		CustomProperties: d.preservedCustomProps,
		RootRels:         d.preservedRootRels,
	}
}

// HasPreservedParts returns true if this document has preserved parts from reading.
func (d *document) HasPreservedParts() bool {
	return d != nil && d.GetPreservedParts() != nil
}

// HasPreservedHeadersOrFooters returns true if this document has preserved
// header or footer bytes from reading. WriteTo writes preserved headers and
// footers verbatim (see writer.PreservedParts), so any in-memory mutation to
// a header or footer paragraph on such a document never reaches the saved
// file.
func (d *document) HasPreservedHeadersOrFooters() bool {
	return d != nil && (len(d.preservedHeaders) > 0 || len(d.preservedFooters) > 0)
}

func normalizeNumberingTarget(target string) string {
	trimmed := strings.TrimSpace(target)
	trimmed = strings.TrimPrefix(trimmed, "./")
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	trimmed = strings.TrimPrefix(trimmed, "/")
	if strings.HasPrefix(strings.ToLower(trimmed), "word/") {
		trimmed = trimmed[5:]
	}
	for strings.HasPrefix(trimmed, "../") {
		trimmed = strings.TrimPrefix(trimmed, "../")
	}
	if trimmed == "" {
		trimmed = "numbering.xml"
	}
	return trimmed
}

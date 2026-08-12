// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

// Package writer handles writing DOCX files as ZIP archives containing XML documents.
// It provides the ZipWriter for creating properly structured Office Open XML packages.
package writer

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/manager"
	"github.com/mmonterroca/docxgo/v2/internal/serializer"
	xmlstructs "github.com/mmonterroca/docxgo/v2/internal/xml"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
)

// ZipWriter writes a .docx file to an io.Writer.
type ZipWriter struct {
	zipWriter  *zip.Writer
	serializer *serializer.DocumentSerializer
	language   *domain.Language
}

// NumberingPart represents numbering.xml data that should be preserved in the DOCX package.
type NumberingPart struct {
	Data   []byte
	Target string
}

// PreservedParts holds all parts that should be written verbatim from the original document.
// This enables complete round-trip fidelity when reading and saving documents.
type PreservedParts struct {
	Headers          map[string][]byte // Original headers, keyed by archive path (e.g. "word/header1.xml")
	Footers          map[string][]byte // Original footers, keyed by archive path (e.g. "word/footer1.xml")
	HeaderRels       map[string][]byte // Original word/_rels/headerN.xml.rels, keyed by archive path
	FooterRels       map[string][]byte // Original word/_rels/footerN.xml.rels, keyed by archive path
	DocRels          []byte            // Original word/_rels/document.xml.rels
	ContentTypes     []byte            // Original [Content_Types].xml
	Additional       map[string][]byte // Additional parts (comments, footnotes, customXml, etc.)
	Themes           map[string][]byte // Original theme parts
	FontTable        []byte            // Original fontTable.xml
	Settings         []byte            // Original settings.xml
	WebSettings      []byte            // Original webSettings.xml
	CustomProperties []byte            // Original docProps/custom.xml
	RootRels         []byte            // Original _rels/.rels
}

// NewZipWriter creates a new ZipWriter.
func NewZipWriter(w io.Writer) *ZipWriter {
	return &ZipWriter{
		zipWriter:  zip.NewWriter(w),
		serializer: serializer.NewDocumentSerializer(),
	}
}

// SetLanguage sets the document's default proofing language, written to
// word/settings.xml as w:themeFontLang when the default (non-preserved)
// settings part is generated.
func (zw *ZipWriter) SetLanguage(lang *domain.Language) {
	zw.language = lang
}

// WriteDocument writes a complete .docx document structure.
// If preservedStyles is provided (non-nil), it will be written verbatim instead of serializing styles.
// If preserved is provided, those parts will be written verbatim for complete round-trip fidelity.
//
// partRels holds the relationships owned by individual header/footer parts,
// keyed by the archive path of the .rels file itself (e.g.
// "word/_rels/header1.xml.rels"). It is the generated counterpart to
// PreservedParts.HeaderRels/FooterRels; an entry is written only when the
// part it belongs to is being regenerated, or when no preserved part claims
// that name.
//
// regenerated names the header/footer content parts (by archive path, e.g.
// "word/header1.xml") whose in-memory model has diverged from the file they
// were read from. Those are written from the model; every other preserved
// part goes back byte-for-byte. Empty for a document built from scratch,
// where there is nothing preserved to choose between.
func (zw *ZipWriter) WriteDocument(doc *xmlstructs.Document, rels *xmlstructs.Relationships, coreProps *xmlstructs.CoreProperties, appProps *xmlstructs.AppProperties, styles *xmlstructs.Styles, media []*manager.MediaFile, headers map[string]*xmlstructs.Header, footers map[string]*xmlstructs.Footer, partRels map[string]*xmlstructs.Relationships, regenerated map[string]bool, numbering *NumberingPart, preservedStyles []byte, preserved *PreservedParts) error {
	numberingPart := sanitizeNumberingPart(numbering)

	// Determine if we're in round-trip mode (have preserved parts)
	roundTrip := preserved != nil

	// Write [Content_Types].xml
	if roundTrip && len(preserved.ContentTypes) > 0 {
		if err := zw.writeRaw("[Content_Types].xml", amendContentTypes(preserved.ContentTypes, headers, footers, media)); err != nil {
			return fmt.Errorf("write preserved content types: %w", err)
		}
	} else {
		if err := zw.writeContentTypes(headers, footers, media, numberingPart); err != nil {
			return fmt.Errorf("write content types: %w", err)
		}
	}

	// Write _rels/.rels - use preserved if available
	if roundTrip && len(preserved.RootRels) > 0 {
		if err := zw.writeRaw("_rels/.rels", preserved.RootRels); err != nil {
			return fmt.Errorf("write preserved root rels: %w", err)
		}
	} else {
		if err := zw.writeRootRels(); err != nil {
			return fmt.Errorf("write root rels: %w", err)
		}
	}

	// Write word/document.xml
	if err := zw.writeMainDocument(doc); err != nil {
		return fmt.Errorf("write main document: %w", err)
	}

	// Write word/_rels/document.xml.rels
	if roundTrip && len(preserved.DocRels) > 0 {
		if err := zw.writeRaw("word/_rels/document.xml.rels", preserved.DocRels); err != nil {
			return fmt.Errorf("write preserved document rels: %w", err)
		}
	} else {
		if err := zw.writeDocumentRels(rels, numberingPart); err != nil {
			return fmt.Errorf("write document rels: %w", err)
		}
	}

	// Write docProps/core.xml
	if err := zw.writeCoreProperties(coreProps); err != nil {
		return fmt.Errorf("write core properties: %w", err)
	}

	// Write docProps/app.xml
	if err := zw.writeAppProperties(appProps); err != nil {
		return fmt.Errorf("write app properties: %w", err)
	}

	// Write word/styles.xml - use preserved bytes if available
	if len(preservedStyles) > 0 {
		if err := zw.writeRaw("word/styles.xml", preservedStyles); err != nil {
			return fmt.Errorf("write preserved styles: %w", err)
		}
	} else if err := zw.writeStyles(styles); err != nil {
		return fmt.Errorf("write styles: %w", err)
	}

	// Write word/fontTable.xml
	if roundTrip && len(preserved.FontTable) > 0 {
		if err := zw.writeRaw("word/fontTable.xml", preserved.FontTable); err != nil {
			return fmt.Errorf("write preserved font table: %w", err)
		}
	} else {
		if err := zw.writeDefaultFontTable(); err != nil {
			return fmt.Errorf("write font table: %w", err)
		}
	}

	// Write word/theme/theme1.xml
	if roundTrip && len(preserved.Themes) > 0 {
		for name, data := range preserved.Themes {
			if err := zw.writeRaw(name, data); err != nil {
				return fmt.Errorf("write preserved theme %s: %w", name, err)
			}
		}
	} else {
		if err := zw.writeDefaultTheme(); err != nil {
			return fmt.Errorf("write theme: %w", err)
		}
	}

	// Write word/settings.xml
	if roundTrip && len(preserved.Settings) > 0 {
		if err := zw.writeRaw("word/settings.xml", preserved.Settings); err != nil {
			return fmt.Errorf("write preserved settings: %w", err)
		}
	} else {
		if err := zw.writeDefaultSettings(); err != nil {
			return fmt.Errorf("write settings: %w", err)
		}
	}

	// Write word/webSettings.xml
	if roundTrip && len(preserved.WebSettings) > 0 {
		if err := zw.writeRaw("word/webSettings.xml", preserved.WebSettings); err != nil {
			return fmt.Errorf("write preserved web settings: %w", err)
		}
	} else {
		if err := zw.writeDefaultWebSettings(); err != nil {
			return fmt.Errorf("write web settings: %w", err)
		}
	}

	// Write media files to word/media
	if err := zw.writeMediaFiles(media); err != nil {
		return fmt.Errorf("write media: %w", err)
	}

	// Write headers and footers, merging generated and preserved per name.
	// See writeHeaderFooterParts for why this is not all-or-nothing.
	fromModel, err := zw.writeHeaderFooterParts(headers, footers, regenerated, preserved, roundTrip)
	if err != nil {
		return err
	}

	// Write the relationship parts each header/footer owns (an image or
	// hyperlink placed in one). These cannot live in
	// word/_rels/document.xml.rels: a header is its own OPC part and cannot
	// resolve an r:id declared there.
	//
	// The generated set wins for a part written from the model -- its r:ids
	// come from the model and will not match the preserved rels, and pairing
	// new content with old relationships is exactly the dangling-r:id package
	// #101 was about. For an untouched part the preserved bytes win.
	//
	// Whichever loses is skipped rather than also written: writeRaw is a bare
	// zip.Create and archive/zip accepts duplicate entry names silently,
	// leaving Word to pick one unpredictably.
	writtenRels := make(map[string]bool, len(partRels))
	for name, partRel := range partRels {
		if partRel == nil {
			continue
		}
		if roundTrip && !fromModel[relsOwnerPath(name)] &&
			(hasPart(preserved.HeaderRels, name) || hasPart(preserved.FooterRels, name)) {
			continue
		}
		if err := zw.writeXML(name, partRel); err != nil {
			return fmt.Errorf("write part rels %s: %w", name, err)
		}
		writtenRels[name] = true
	}

	if roundTrip {
		for _, preservedRels := range []map[string][]byte{preserved.HeaderRels, preserved.FooterRels} {
			for name, data := range preservedRels {
				// Skip a rels part already written from the model, and one
				// whose part was regenerated but owns no relationships at all
				// -- keeping the old rels there would leave r:ids the new
				// content never references.
				if writtenRels[name] || fromModel[relsOwnerPath(name)] {
					continue
				}
				if err := zw.writeRaw(name, data); err != nil {
					return fmt.Errorf("write preserved part rels %s: %w", name, err)
				}
				writtenRels[name] = true
			}
		}
	}

	// Write numbering part
	if numberingPart != nil {
		if err := zw.writeRaw(fmt.Sprintf("word/%s", numberingPart.Target), numberingPart.Data); err != nil {
			return fmt.Errorf("write numbering part: %w", err)
		}
	}

	// Write additional preserved parts (comments, footnotes, customXml, etc.)
	if roundTrip && len(preserved.Additional) > 0 {
		for name, data := range preserved.Additional {
			if err := zw.writeRaw(name, data); err != nil {
				return fmt.Errorf("write additional part %s: %w", name, err)
			}
		}
	}

	// Write custom properties if preserved
	if roundTrip && len(preserved.CustomProperties) > 0 {
		if err := zw.writeRaw("docProps/custom.xml", preserved.CustomProperties); err != nil {
			return fmt.Errorf("write custom properties: %w", err)
		}
	}

	return nil
}

// Close closes the ZIP writer.
func (zw *ZipWriter) Close() error {
	return zw.zipWriter.Close()
}

// writeContentTypes writes [Content_Types].xml
func (zw *ZipWriter) writeContentTypes(headers map[string]*xmlstructs.Header, footers map[string]*xmlstructs.Footer, media []*manager.MediaFile, numbering *NumberingPart) error {
	ct := &xmlstructs.ContentTypes{
		Xmlns: constants.NamespaceContentTypes,
		Defaults: []*xmlstructs.Default{
			{Extension: "rels", ContentType: constants.ContentTypeRelationships},
			{Extension: "xml", ContentType: "application/xml"},
		},
		Overrides: []*xmlstructs.Override{
			{PartName: "/word/document.xml", ContentType: constants.ContentTypeDocument},
			{PartName: "/word/styles.xml", ContentType: constants.ContentTypeStyles},
			{PartName: "/word/fontTable.xml", ContentType: constants.ContentTypeFontTable},
			{PartName: "/word/theme/theme1.xml", ContentType: constants.ContentTypeTheme},
			{PartName: "/word/settings.xml", ContentType: constants.ContentTypeSettings},
			{PartName: "/word/webSettings.xml", ContentType: constants.ContentTypeWebSettings},
			{PartName: "/docProps/core.xml", ContentType: constants.ContentTypeCoreProperties},
			{PartName: "/docProps/app.xml", ContentType: constants.ContentTypeExtendedProperties},
		},
	}

	addOverride := func(name, contentType string) {
		if name == "" {
			return
		}
		for _, existing := range ct.Overrides {
			if existing.PartName == name {
				return
			}
		}
		ct.Overrides = append(ct.Overrides, &xmlstructs.Override{PartName: name, ContentType: contentType})
	}

	for name := range headers {
		addOverride(fmt.Sprintf("/word/%s", name), constants.ContentTypeHeader)
	}

	// Include defaults for media content types
	addDefault := func(extension, contentType string) {
		if extension == "" || contentType == "" {
			return
		}
		ext := strings.ToLower(extension)
		for _, existing := range ct.Defaults {
			if existing != nil && strings.EqualFold(existing.Extension, ext) {
				return
			}
		}
		ct.Defaults = append(ct.Defaults, &xmlstructs.Default{
			Extension:   ext,
			ContentType: contentType,
		})
	}

	for _, file := range media {
		if file == nil || len(file.Data) == 0 {
			continue
		}
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(file.Name)), ".")
		addDefault(ext, file.ContentType)
	}
	for name := range footers {
		addOverride(fmt.Sprintf("/word/%s", name), constants.ContentTypeFooter)
	}

	if numbering != nil {
		addOverride(fmt.Sprintf("/word/%s", numbering.Target), constants.ContentTypeNumbering)
	}

	return zw.writeXML("[Content_Types].xml", ct)
}

// writeRootRels writes _rels/.rels
func (zw *ZipWriter) writeRootRels() error {
	rels := &xmlstructs.Relationships{
		Xmlns: constants.NamespacePackageRels,
		Relationships: []*xmlstructs.Relationship{
			{
				ID:     "rId1",
				Type:   constants.RelTypeDocument,
				Target: "word/document.xml",
			},
			{
				ID:     "rId2",
				Type:   "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties",
				Target: "docProps/core.xml",
			},
			{
				ID:     "rId3",
				Type:   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties",
				Target: "docProps/app.xml",
			},
		},
	}

	return zw.writeXML("_rels/.rels", rels)
}

// writeMainDocument writes word/document.xml
func (zw *ZipWriter) writeMainDocument(doc *xmlstructs.Document) error {
	return zw.writeXML("word/document.xml", doc)
}

// writeDocumentRels writes word/_rels/document.xml.rels
func (zw *ZipWriter) writeDocumentRels(rels *xmlstructs.Relationships, numbering *NumberingPart) error {
	if rels == nil {
		rels = &xmlstructs.Relationships{
			Xmlns:         constants.NamespacePackageRels,
			Relationships: []*xmlstructs.Relationship{},
		}
	}

	if rels.Xmlns == "" {
		rels.Xmlns = constants.NamespacePackageRels
	}

	nextRelID := func() string {
		maxID := 0
		for _, rel := range rels.Relationships {
			if rel == nil {
				continue
			}
			if strings.HasPrefix(rel.ID, "rId") {
				if n, err := strconv.Atoi(strings.TrimPrefix(rel.ID, "rId")); err == nil && n > maxID {
					maxID = n
				}
			}
		}
		return fmt.Sprintf("rId%d", maxID+1)
	}

	ensureRel := func(relType, target string) {
		if target == "" {
			return
		}
		for _, rel := range rels.Relationships {
			if rel != nil && rel.Target == target {
				return
			}
		}
		rels.Relationships = append(rels.Relationships, &xmlstructs.Relationship{
			ID:     nextRelID(),
			Type:   relType,
			Target: target,
		})
	}

	ensureRel(constants.RelTypeStyles, "styles.xml")
	ensureRel(constants.RelTypeFontTable, "fontTable.xml")
	ensureRel(constants.RelTypeTheme, "theme/theme1.xml")
	ensureRel(constants.RelTypeSettings, "settings.xml")
	ensureRel(constants.RelTypeWebSettings, "webSettings.xml")

	if numbering != nil {
		ensureRel(constants.RelTypeNumbering, numbering.Target)
	}

	return zw.writeXML("word/_rels/document.xml.rels", rels)
}

// writeCoreProperties writes docProps/core.xml
func (zw *ZipWriter) writeCoreProperties(props *xmlstructs.CoreProperties) error {
	if props == nil {
		now := time.Now()
		props = &xmlstructs.CoreProperties{
			XMLnsCP:      constants.NamespaceCoreProperties,
			XMLnsDC:      constants.NamespaceDC,
			XMLnsDCTerms: constants.NamespaceDCTerms,
			XMLnsXSI:     "http://www.w3.org/2001/XMLSchema-instance",
			Creator:      "docxgo",
			Created: &xmlstructs.DCDate{
				Type:  "dcterms:W3CDTF",
				Value: now.Format(time.RFC3339),
			},
			Modified: &xmlstructs.DCDate{
				Type:  "dcterms:W3CDTF",
				Value: now.Format(time.RFC3339),
			},
		}
	}
	return zw.writeXML("docProps/core.xml", props)
}

// writeAppProperties writes docProps/app.xml
func (zw *ZipWriter) writeAppProperties(props *xmlstructs.AppProperties) error {
	if props == nil {
		props = &xmlstructs.AppProperties{
			Xmlns:       constants.NamespaceExtendedProperties,
			Application: "docxgo",
			DocSecurity: 0,
		}
	}
	return zw.writeXML("docProps/app.xml", props)
}

// writeStyles writes word/styles.xml from serialized styles. styles must be
// non-nil: the only production caller (internal/core/document.go) always
// passes serializer.SerializeStyles's result, which is never nil, so a nil
// here indicates a caller bug rather than a case to degrade gracefully for.
func (zw *ZipWriter) writeStyles(styles *xmlstructs.Styles) error {
	if styles == nil {
		return fmt.Errorf("writer: writeStyles: styles must not be nil")
	}

	w, err := zw.zipWriter.Create("word/styles.xml")
	if err != nil {
		return err
	}

	// Write XML declaration
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return err
	}

	// Marshal and write styles
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	return encoder.Encode(styles)
}

// writeDefaultFontTable writes minimal word/fontTable.xml
func (zw *ZipWriter) writeDefaultFontTable() error {
	fontTable := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:fonts xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:font w:name="Calibri">
    <w:panose1 w:val="020F0502020204030204"/>
    <w:charset w:val="00"/>
    <w:family w:val="swiss"/>
    <w:pitch w:val="variable"/>
  </w:font>
</w:fonts>`
	return zw.writeRaw("word/fontTable.xml", []byte(fontTable))
}

// writeDefaultTheme writes minimal word/theme/theme1.xml
func (zw *ZipWriter) writeDefaultTheme() error {
	theme := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Office Theme">
	<a:themeElements>
		<a:clrScheme name="Office">
			<a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1>
			<a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1>
			<a:dk2><a:srgbClr val="44546A"/></a:dk2>
			<a:lt2><a:srgbClr val="E7E6E6"/></a:lt2>
			<a:accent1><a:srgbClr val="4472C4"/></a:accent1>
			<a:accent2><a:srgbClr val="ED7D31"/></a:accent2>
			<a:accent3><a:srgbClr val="A5A5A5"/></a:accent3>
			<a:accent4><a:srgbClr val="FFC000"/></a:accent4>
			<a:accent5><a:srgbClr val="5B9BD5"/></a:accent5>
			<a:accent6><a:srgbClr val="70AD47"/></a:accent6>
			<a:hlink><a:srgbClr val="0563C1"/></a:hlink>
			<a:folHlink><a:srgbClr val="954F72"/></a:folHlink>
		</a:clrScheme>
		<a:fontScheme name="Office">
			<a:majorFont>
				<a:latin typeface="Calibri Light"/>
				<a:ea typeface=""/>
				<a:cs typeface=""/>
			</a:majorFont>
			<a:minorFont>
				<a:latin typeface="Calibri"/>
				<a:ea typeface=""/>
				<a:cs typeface=""/>
			</a:minorFont>
		</a:fontScheme>
		<a:fmtScheme name="Office">
			<a:fillStyleLst>
				<a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
				<a:gradFill rotWithShape="1">
					<a:gsLst>
						<a:gs pos="0"><a:schemeClr val="phClr"><a:tint val="50000"/><a:satMod val="300000"/></a:schemeClr></a:gs>
						<a:gs pos="35000"><a:schemeClr val="phClr"><a:tint val="37000"/><a:satMod val="300000"/></a:schemeClr></a:gs>
						<a:gs pos="100000"><a:schemeClr val="phClr"><a:tint val="15000"/><a:satMod val="350000"/></a:schemeClr></a:gs>
					</a:gsLst>
					<a:lin ang="16200000" scaled="1"/>
				</a:gradFill>
				<a:gradFill rotWithShape="1">
					<a:gsLst>
						<a:gs pos="0"><a:schemeClr val="phClr"><a:shade val="51000"/><a:satMod val="130000"/></a:schemeClr></a:gs>
						<a:gs pos="80000"><a:schemeClr val="phClr"><a:shade val="93000"/><a:satMod val="130000"/></a:schemeClr></a:gs>
						<a:gs pos="100000"><a:schemeClr val="phClr"><a:shade val="94000"/><a:satMod val="350000"/></a:schemeClr></a:gs>
					</a:gsLst>
					<a:lin ang="16200000" scaled="1"/>
				</a:gradFill>
			</a:fillStyleLst>
			<a:lnStyleLst>
				<a:ln w="9525" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/><a:miter lim="800000"/></a:ln>
				<a:ln w="25400" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/><a:miter lim="800000"/></a:ln>
				<a:ln w="38100" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/><a:miter lim="800000"/></a:ln>
			</a:lnStyleLst>
			<a:effectStyleLst>
				<a:effectStyle><a:effectLst/></a:effectStyle>
				<a:effectStyle><a:effectLst/></a:effectStyle>
				<a:effectStyle>
					<a:effectLst>
						<a:outerShdw blurRad="57150" dist="19050" dir="5400000" algn="ctr" rotWithShape="0">
							<a:srgbClr val="000000"><a:alpha val="63000"/></a:srgbClr>
						</a:outerShdw>
					</a:effectLst>
				</a:effectStyle>
			</a:effectStyleLst>
			<a:bgFillStyleLst>
				<a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
				<a:solidFill><a:schemeClr val="phClr"><a:tint val="95000"/><a:satMod val="170000"/></a:schemeClr></a:solidFill>
				<a:gradFill rotWithShape="1">
					<a:gsLst>
						<a:gs pos="0"><a:schemeClr val="phClr"><a:tint val="93000"/><a:satMod val="150000"/><a:shade val="98000"/><a:lumMod val="102000"/></a:schemeClr></a:gs>
						<a:gs pos="50000"><a:schemeClr val="phClr"><a:tint val="98000"/><a:satMod val="130000"/><a:shade val="90000"/><a:lumMod val="103000"/></a:schemeClr></a:gs>
						<a:gs pos="100000"><a:schemeClr val="phClr"><a:shade val="63000"/><a:satMod val="120000"/></a:schemeClr></a:gs>
					</a:gsLst>
					<a:lin ang="16200000" scaled="1"/>
				</a:gradFill>
			</a:bgFillStyleLst>
		</a:fmtScheme>
	</a:themeElements>
	<a:objectDefaults/>
	<a:extraClrSchemeLst/>
</a:theme>`
	return zw.writeRaw("word/theme/theme1.xml", []byte(theme))
}

// writeDefaultSettings writes a baseline word/settings.xml part.
func (zw *ZipWriter) writeDefaultSettings() error {
	settings := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:settings xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
	<w:zoom w:percent="100"/>
	<w:defaultTabStop w:val="720"/>
	<w:characterSpacingControl w:val="doNotCompress"/>
	<w:compat>
		<w:compatSetting w:name="compatibilityMode" w:uri="http://schemas.microsoft.com/office/word" w:val="15"/>
	</w:compat>` + zw.themeFontLangXML() + `
</w:settings>`
	return zw.writeRaw("word/settings.xml", []byte(settings))
}

// themeFontLangXML returns the w:themeFontLang element reflecting the
// configured language, or an empty string when no language is set. It is
// placed immediately after w:compat, which is schema-valid since no elements
// between them (w:docVars, w:rsids, m:mathPr, w:attachedSchema) are emitted.
func (zw *ZipWriter) themeFontLangXML() string {
	if zw.language == nil ||
		(zw.language.Val == "" && zw.language.EastAsia == "" && zw.language.Bidi == "") {
		return ""
	}
	var attrs string
	if zw.language.Val != "" {
		attrs += ` w:val="` + xmlEscapeAttr(zw.language.Val) + `"`
	}
	if zw.language.EastAsia != "" {
		attrs += ` w:eastAsia="` + xmlEscapeAttr(zw.language.EastAsia) + `"`
	}
	if zw.language.Bidi != "" {
		attrs += ` w:bidi="` + xmlEscapeAttr(zw.language.Bidi) + `"`
	}
	return "\n\t<w:themeFontLang" + attrs + "/>"
}

// xmlEscapeAttr escapes a string for safe use as an XML attribute value.
func xmlEscapeAttr(s string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

// writeDefaultWebSettings writes a baseline word/webSettings.xml part.
func (zw *ZipWriter) writeDefaultWebSettings() error {
	webSettings := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:webSettings xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
	<w:allowPNG/>
</w:webSettings>`
	return zw.writeRaw("word/webSettings.xml", []byte(webSettings))
}

// writeXML marshals and writes an XML structure to the ZIP.
func (zw *ZipWriter) writeXML(path string, v interface{}) error {
	w, err := zw.zipWriter.Create(path)
	if err != nil {
		return err
	}

	// Write XML header
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return err
	}

	// Marshal and write XML
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(v); err != nil {
		return err
	}

	return nil
}

// hasPart reports whether parts already claims the given archive path. Used to
// keep a generated part from being written under a name a preserved part
// already occupies, which archive/zip would accept as a duplicate entry.
func hasPart(parts map[string][]byte, name string) bool {
	_, ok := parts[name]
	return ok
}

// amendContentTypes returns the preserved [Content_Types].xml with an Override
// added for any header or footer part that does not already have one, and a
// Default added for any media extension it does not already declare.
//
// A round-tripped package writes this part verbatim, which is correct while
// its part list cannot change -- but a part with no declared content type
// makes the whole package invalid, and two things can add one after the
// document was opened. A header *added* to an opened document needs an
// Override. An image added to a document whose source held none needs a
// Default for its extension: the media part is written either way, so without
// this the first image added to an image-free document produces a package Word
// offers to repair. That second one predates the header work here -- it is
// reachable on any released version through OpenDocument + AddImage -- but
// this is the function that owns the problem now.
//
// Only the missing entries are added; the preserved bytes are returned
// untouched when nothing is missing, so an unmodified document still
// round-trips byte-for-byte here.
//
// Malformed preserved bytes are returned unchanged rather than replaced: the
// same call the document-rels path makes, for the same reason -- better to
// keep a part that could not be parsed than to risk mangling it.
func amendContentTypes(preservedBytes []byte, headers map[string]*xmlstructs.Header, footers map[string]*xmlstructs.Footer, media []*manager.MediaFile) []byte {
	if len(headers) == 0 && len(footers) == 0 && len(media) == 0 {
		return preservedBytes
	}

	var types xmlstructs.ContentTypes
	if err := xml.Unmarshal(preservedBytes, &types); err != nil {
		return preservedBytes
	}

	declared := make(map[string]bool, len(types.Overrides))
	for _, override := range types.Overrides {
		if override != nil {
			declared[strings.ToLower(override.PartName)] = true
		}
	}

	// Sorted so the appended entries land in a stable order regardless of
	// map iteration.
	missing := make([]*xmlstructs.Override, 0, len(headers)+len(footers))
	collect := func(targets []string, contentType string) {
		sort.Strings(targets)
		for _, target := range targets {
			partName := "/" + PartArchivePath(target)
			if declared[strings.ToLower(partName)] {
				continue
			}
			declared[strings.ToLower(partName)] = true
			missing = append(missing, &xmlstructs.Override{PartName: partName, ContentType: contentType})
		}
	}
	collect(mapKeys(headers), constants.ContentTypeHeader)
	collect(mapKeys(footers), constants.ContentTypeFooter)

	declaredExt := make(map[string]bool, len(types.Defaults))
	for _, def := range types.Defaults {
		if def != nil {
			declaredExt[strings.ToLower(def.Extension)] = true
		}
	}

	// Sorted by part name so the appended defaults land in a stable order:
	// media arrives in whatever order the manager holds it, and two saves of
	// the same document should produce the same bytes.
	ordered := make([]*manager.MediaFile, 0, len(media))
	for _, file := range media {
		if file != nil && len(file.Data) > 0 && file.ContentType != "" {
			ordered = append(ordered, file)
		}
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Name < ordered[j].Name })

	missingDefaults := make([]*xmlstructs.Default, 0, len(ordered))
	for _, file := range ordered {
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(file.Name)), ".")
		if ext == "" || declaredExt[ext] {
			continue
		}
		declaredExt[ext] = true
		missingDefaults = append(missingDefaults, &xmlstructs.Default{Extension: ext, ContentType: file.ContentType})
	}

	if len(missing) == 0 && len(missingDefaults) == 0 {
		return preservedBytes
	}

	types.Defaults = append(types.Defaults, missingDefaults...)
	types.Overrides = append(types.Overrides, missing...)
	out, err := xml.MarshalIndent(&types, "", "  ")
	if err != nil {
		return preservedBytes
	}
	return append([]byte(xml.Header), out...)
}

// mapKeys returns m's keys, for the callers that need to iterate in a stable
// order rather than Go's randomized map order.
func mapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// PartArchivePath maps a header/footer relationship target -- which for an
// opened document is whatever the source file's rels said -- to the archive
// path the part is written under, the same shape PreservedParts.Headers is
// keyed by.
//
// A relationship target is either package-absolute ("/word/header1.xml"),
// which names the entry directly, or relative to the part that owns the
// relationship, which for a header or footer reference is word/document.xml
// and so resolves under "word/". Resolving rather than collapsing to the base
// name matters: a target naming a subdirectory ("headers/header1.xml") is
// legal, and treating it as "word/header1.xml" writes the part to a path
// nothing references while the real one is still written from its preserved
// bytes -- one junk part in the package, and the caller's edit lost.
//
// Exported because internal/core has to key its header/footer snapshots by
// exactly the same paths the writer uses; two copies of this that drifted
// apart would silently break the per-name merge.
func PartArchivePath(target string) string {
	t := strings.TrimSpace(strings.ReplaceAll(target, `\`, "/"))
	if t == "" {
		return ""
	}
	if strings.HasPrefix(t, "/") {
		return path.Clean(strings.TrimPrefix(t, "/"))
	}
	cleaned := path.Clean(t)
	// Tolerate a producer that spelled a relative target from the package
	// root anyway ("word/header1.xml"). Resolving that under "word/" a second
	// time would invent word/word/header1.xml; it is technically wrong of the
	// producer, but it round-tripped before and there is nothing to gain by
	// breaking it.
	if cleaned == "word" || strings.HasPrefix(cleaned, "word/") {
		return cleaned
	}
	return path.Clean("word/" + cleaned)
}

// relsOwnerPath maps a part's .rels archive path back to the part it belongs
// to: "word/_rels/header1.xml.rels" -> "word/header1.xml", and
// "word/headers/_rels/header1.xml.rels" -> "word/headers/header1.xml".
func relsOwnerPath(relsPath string) string {
	dir := path.Dir(relsPath) // ".../_rels"
	base := strings.TrimSuffix(path.Base(relsPath), ".rels")
	return path.Join(path.Dir(dir), base)
}

// writeHeaderFooterParts writes every header and footer, choosing per name
// between the model and the preserved original, and returns the archive paths
// it wrote from the model.
//
// This used to be all-or-nothing: a single preserved header discarded the
// entire generated map, so a header modified -- or newly added -- on an opened
// document was silently dropped. Merging per name is what lets an edit reach
// the file while every part the caller never touched still goes back
// byte-for-byte.
//
// A part is written from the model when regenerated says its model diverged
// from the file, or when nothing preserved claims that name (a header added
// after the document was opened). Otherwise the preserved bytes win.
func (zw *ZipWriter) writeHeaderFooterParts(headers map[string]*xmlstructs.Header, footers map[string]*xmlstructs.Footer, regenerated map[string]bool, preserved *PreservedParts, roundTrip bool) (map[string]bool, error) {
	fromModel := make(map[string]bool, len(headers)+len(footers))

	writeGenerated := func(target string, part interface{}, preservedParts map[string][]byte, kind string) error {
		name := PartArchivePath(target)
		if roundTrip && hasPart(preservedParts, name) && !regenerated[name] {
			return nil
		}
		if err := zw.writeXML(name, part); err != nil {
			return fmt.Errorf("write %s %s: %w", kind, name, err)
		}
		fromModel[name] = true
		return nil
	}

	for target, header := range headers {
		var preservedHeaders map[string][]byte
		if roundTrip {
			preservedHeaders = preserved.Headers
		}
		if err := writeGenerated(target, header, preservedHeaders, "header"); err != nil {
			return nil, err
		}
	}
	for target, footer := range footers {
		var preservedFooters map[string][]byte
		if roundTrip {
			preservedFooters = preserved.Footers
		}
		if err := writeGenerated(target, footer, preservedFooters, "footer"); err != nil {
			return nil, err
		}
	}

	if !roundTrip {
		return fromModel, nil
	}

	for _, preservedParts := range []map[string][]byte{preserved.Headers, preserved.Footers} {
		for name, data := range preservedParts {
			if fromModel[name] {
				continue
			}
			if err := zw.writeRaw(name, data); err != nil {
				return nil, fmt.Errorf("write preserved part %s: %w", name, err)
			}
		}
	}

	return fromModel, nil
}

// writeRaw writes raw bytes to the ZIP.
func (zw *ZipWriter) writeRaw(path string, data []byte) error {
	w, err := zw.zipWriter.Create(path)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// writeMediaFiles writes all media assets into the DOCX package.
func (zw *ZipWriter) writeMediaFiles(media []*manager.MediaFile) error {
	for _, file := range media {
		if file == nil || len(file.Data) == 0 || file.Path == "" {
			continue
		}
		if err := zw.writeRaw(file.Path, file.Data); err != nil {
			return err
		}
	}
	return nil
}

func sanitizeNumberingPart(part *NumberingPart) *NumberingPart {
	if part == nil || len(part.Data) == 0 {
		return nil
	}
	return &NumberingPart{
		Data:   part.Data,
		Target: sanitizeNumberingTarget(part.Target),
	}
}

func sanitizeNumberingTarget(target string) string {
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

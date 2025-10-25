package writer

/*
   Copyright (c) 2025 SlideLang

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"time"

	"github.com/SlideLang/go-docx/internal/serializer"
	xmlstructs "github.com/SlideLang/go-docx/internal/xml"
	"github.com/SlideLang/go-docx/pkg/constants"
)

// ZipWriter writes a .docx file to an io.Writer.
type ZipWriter struct {
	zipWriter  *zip.Writer
	serializer *serializer.DocumentSerializer
}

// NewZipWriter creates a new ZipWriter.
func NewZipWriter(w io.Writer) *ZipWriter {
	return &ZipWriter{
		zipWriter:  zip.NewWriter(w),
		serializer: serializer.NewDocumentSerializer(),
	}
}

// WriteDocument writes a complete .docx document structure.
func (zw *ZipWriter) WriteDocument(doc *xmlstructs.Document, rels *xmlstructs.Relationships, coreProps *xmlstructs.CoreProperties, appProps *xmlstructs.AppProperties) error {
	// Write [Content_Types].xml
	if err := zw.writeContentTypes(); err != nil {
		return fmt.Errorf("write content types: %w", err)
	}

	// Write _rels/.rels
	if err := zw.writeRootRels(); err != nil {
		return fmt.Errorf("write root rels: %w", err)
	}

	// Write word/document.xml
	if err := zw.writeMainDocument(doc); err != nil {
		return fmt.Errorf("write main document: %w", err)
	}

	// Write word/_rels/document.xml.rels
	if err := zw.writeDocumentRels(rels); err != nil {
		return fmt.Errorf("write document rels: %w", err)
	}

	// Write docProps/core.xml
	if err := zw.writeCoreProperties(coreProps); err != nil {
		return fmt.Errorf("write core properties: %w", err)
	}

	// Write docProps/app.xml
	if err := zw.writeAppProperties(appProps); err != nil {
		return fmt.Errorf("write app properties: %w", err)
	}

	// Write word/styles.xml (minimal default)
	if err := zw.writeDefaultStyles(); err != nil {
		return fmt.Errorf("write styles: %w", err)
	}

	// Write word/fontTable.xml (minimal default)
	if err := zw.writeDefaultFontTable(); err != nil {
		return fmt.Errorf("write font table: %w", err)
	}

	// Write word/theme/theme1.xml (minimal default)
	if err := zw.writeDefaultTheme(); err != nil {
		return fmt.Errorf("write theme: %w", err)
	}

	return nil
}

// Close closes the ZIP writer.
func (zw *ZipWriter) Close() error {
	return zw.zipWriter.Close()
}

// writeContentTypes writes [Content_Types].xml
func (zw *ZipWriter) writeContentTypes() error {
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
			{PartName: "/docProps/core.xml", ContentType: constants.ContentTypeCoreProperties},
			{PartName: "/docProps/app.xml", ContentType: constants.ContentTypeExtendedProperties},
		},
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
func (zw *ZipWriter) writeDocumentRels(rels *xmlstructs.Relationships) error {
	if rels == nil {
		rels = &xmlstructs.Relationships{
			Xmlns:         constants.NamespacePackageRels,
			Relationships: []*xmlstructs.Relationship{},
		}
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
			Creator:      "go-docx v2",
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
			Application: "go-docx v2.0.0",
			DocSecurity: 0,
		}
	}
	return zw.writeXML("docProps/app.xml", props)
}

// writeDefaultStyles writes minimal word/styles.xml
func (zw *ZipWriter) writeDefaultStyles() error {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:docDefaults>
    <w:rPrDefault>
      <w:rPr>
        <w:rFonts w:ascii="Calibri" w:hAnsi="Calibri"/>
        <w:sz w:val="22"/>
      </w:rPr>
    </w:rPrDefault>
    <w:pPrDefault/>
  </w:docDefaults>
</w:styles>`
	return zw.writeRaw("word/styles.xml", []byte(styles))
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
      <a:dk1><a:sysClr val="windowText"/></a:dk1>
      <a:lt1><a:sysClr val="window"/></a:lt1>
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
    <a:fontScheme name="Office"><a:majorFont/><a:minorFont/></a:fontScheme>
    <a:fmtScheme name="Office"><a:fillStyleLst/><a:lnStyleLst/><a:effectStyleLst/><a:bgFillStyleLst/></a:fmtScheme>
  </a:themeElements>
</a:theme>`
	return zw.writeRaw("word/theme/theme1.xml", []byte(theme))
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

// writeRaw writes raw bytes to the ZIP.
func (zw *ZipWriter) writeRaw(path string, data []byte) error {
	w, err := zw.zipWriter.Create(path)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

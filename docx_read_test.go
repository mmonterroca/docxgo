// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"image"
	"image/color"
	"image/png"
	"io"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
)

func TestOpenDocumentVariants(t *testing.T) {
	doc := NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("Opened document"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "sample.docx")
	if err := doc.SaveAs(path); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	opened, err := OpenDocument(path)
	if err != nil {
		t.Fatalf("OpenDocument: %v", err)
	}
	assertParagraphText(t, opened, "Opened document")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	openedBytes, err := OpenDocumentFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}
	assertParagraphText(t, openedBytes, "Opened document")

	openedReader, err := OpenDocumentFromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("OpenDocumentFromReader: %v", err)
	}
	assertParagraphText(t, openedReader, "Opened document")
}

func assertParagraphText(t *testing.T, doc domain.Document, expected string) {
	t.Helper()
	paras := doc.Paragraphs()
	if len(paras) == 0 {
		t.Fatalf("expected paragraphs in opened document")
	}
	if got := paras[0].Text(); got != expected {
		t.Fatalf("unexpected paragraph text: %q", got)
	}
}

func TestMetadataRoundTrip(t *testing.T) {
	doc := NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("Metadata test"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	if err := doc.SetMetadata(&domain.Metadata{
		Title:   "Test Title",
		Creator: "Test Author",
		Subject: "Test Subject",
	}); err != nil {
		t.Fatalf("SetMetadata: %v", err)
	}

	// Save to buffer
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	// Reopen from buffer
	doc2, err := OpenDocumentFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	meta := doc2.Metadata()
	if meta.Title != "Test Title" {
		t.Errorf("Title: got %q, want %q", meta.Title, "Test Title")
	}
	if meta.Creator != "Test Author" {
		t.Errorf("Creator: got %q, want %q", meta.Creator, "Test Author")
	}
	if meta.Subject != "Test Subject" {
		t.Errorf("Subject: got %q, want %q", meta.Subject, "Test Subject")
	}
}

// TestGridSpanPreservedAfterRoundTrip reproduces issue #25:
// horizontal cell merges (w:gridSpan) are lost after save+reopen.
func TestGridSpanPreservedAfterRoundTrip(t *testing.T) {
	doc := NewDocument()
	table, err := doc.AddTable(2, 3)
	if err != nil {
		t.Fatalf("AddTable: %v", err)
	}
	table.SetStyle(domain.TableStyleGrid)

	row0, err := table.Row(0)
	if err != nil {
		t.Fatalf("Row(0): %v", err)
	}
	cell, err := row0.Cell(0)
	if err != nil {
		t.Fatalf("Cell(0): %v", err)
	}
	if err := cell.Merge(3, 1); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	p, err := cell.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	r, err := p.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := r.AddText("Spans 3 columns"); err != nil {
		t.Fatalf("AddText: %v", err)
	}

	if got := cell.GridSpan(); got != 3 {
		t.Fatalf("before save: GridSpan = %d, want 3", got)
	}

	// Round-trip through bytes
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	doc2, err := OpenDocumentFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	tables := doc2.Tables()
	if len(tables) == 0 {
		t.Fatal("expected at least one table after reopen")
	}
	row0, err = tables[0].Row(0)
	if err != nil {
		t.Fatalf("Row(0) after reopen: %v", err)
	}
	cell, err = row0.Cell(0)
	if err != nil {
		t.Fatalf("Cell(0) after reopen: %v", err)
	}

	if got := cell.GridSpan(); got != 3 {
		t.Errorf("after reopen: GridSpan = %d, want 3", got)
	}

	// Verify continuation cells are marked so the serializer skips them.
	cell1, err := row0.Cell(1)
	if err != nil {
		t.Fatalf("Cell(1) after reopen: %v", err)
	}
	if !cell1.IsHorizontallyMergedContinuation() {
		t.Errorf("Cell(1) should be a horizontal merge continuation")
	}
	cell2, err := row0.Cell(2)
	if err != nil {
		t.Fatalf("Cell(2) after reopen: %v", err)
	}
	if !cell2.IsHorizontallyMergedContinuation() {
		t.Errorf("Cell(2) should be a horizontal merge continuation")
	}
}

// TestVMergePreservedAfterRoundTrip verifies vertical cell merges survive save+reopen.
func TestVMergePreservedAfterRoundTrip(t *testing.T) {
	doc := NewDocument()
	table, err := doc.AddTable(3, 2)
	if err != nil {
		t.Fatalf("AddTable: %v", err)
	}
	table.SetStyle(domain.TableStyleGrid)

	// Merge rows 0-2 in column 0 vertically (3 rows, 1 col).
	row0, err := table.Row(0)
	if err != nil {
		t.Fatalf("Row(0): %v", err)
	}
	cell, err := row0.Cell(0)
	if err != nil {
		t.Fatalf("Cell(0): %v", err)
	}
	if err := cell.Merge(1, 3); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	p, err := cell.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	r, err := p.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := r.AddText("Vertical span"); err != nil {
		t.Fatalf("AddText: %v", err)
	}

	if got := cell.VMerge(); got != domain.VMergeRestart {
		t.Fatalf("before save: VMerge = %v, want VMergeRestart", got)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	doc2, err := OpenDocumentFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	tables := doc2.Tables()
	if len(tables) == 0 {
		t.Fatal("expected at least one table after reopen")
	}

	// Row 0, Cell 0: should be VMergeRestart
	r0, err := tables[0].Row(0)
	if err != nil {
		t.Fatalf("Row(0) after reopen: %v", err)
	}
	c0, err := r0.Cell(0)
	if err != nil {
		t.Fatalf("Row(0).Cell(0) after reopen: %v", err)
	}
	if got := c0.VMerge(); got != domain.VMergeRestart {
		t.Errorf("Row(0).Cell(0) VMerge = %v, want VMergeRestart", got)
	}

	// Row 1, Cell 0: should be VMergeContinue
	r1, err := tables[0].Row(1)
	if err != nil {
		t.Fatalf("Row(1) after reopen: %v", err)
	}
	c1, err := r1.Cell(0)
	if err != nil {
		t.Fatalf("Row(1).Cell(0) after reopen: %v", err)
	}
	if got := c1.VMerge(); got != domain.VMergeContinue {
		t.Errorf("Row(1).Cell(0) VMerge = %v, want VMergeContinue", got)
	}
}

// TestOpenDocument_CollidingRelationshipIDsAcrossParts reproduces issue #37:
// relationship IDs are scoped per-part in OOXML, so a header's own
// "word/_rels/header1.xml.rels" may reuse an ID (e.g. rId1) that also exists
// in "word/_rels/document.xml.rels" for something unrelated (e.g. a customXml
// part). Opening such a document must resolve each drawing's r:embed against
// its own part's relationships, not the document's.
//
// No fixture .docx exists in this repo, and the library's own writer cannot
// produce a cross-part ID collision (relationship IDs are allocated from one
// shared counter, and header/footer .rels are not even emitted on write), so
// the minimal colliding package is hand-built here with archive/zip.
func TestOpenDocument_CollidingRelationshipIDsAcrossParts(t *testing.T) {
	pngBytes := encodeTestPNG(t)
	docxBytes := buildCollidingRelationshipDocx(t, pngBytes)

	doc, err := OpenDocumentFromBytes(docxBytes)
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	if err := doc.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	sections := doc.Sections()
	if len(sections) == 0 {
		t.Fatal("expected at least one section")
	}

	header, err := sections[0].Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header: %v", err)
	}

	var headerImage domain.Image
	for _, p := range header.Paragraphs() {
		if images := p.Images(); len(images) > 0 {
			headerImage = images[0]
			break
		}
	}

	if headerImage == nil {
		t.Fatal("expected header to contain a hydrated image")
	}

	if !bytes.Equal(headerImage.Data(), pngBytes) {
		t.Fatal("header image data does not match the header's own media part; " +
			"rId1 likely resolved against document.xml.rels instead of header1.xml.rels")
	}

	// Round-trip guard: the header's own .rels is preserved verbatim as an
	// opaque part, so saving and reopening must keep working too.
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "colliding-rel-ids.docx")
	if err := doc.SaveAs(path); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	reopened, err := OpenDocument(path)
	if err != nil {
		t.Fatalf("OpenDocument after round-trip: %v", err)
	}
	if err := reopened.Validate(); err != nil {
		t.Fatalf("Validate after round-trip: %v", err)
	}
}

// encodeTestPNG returns the bytes of a tiny valid PNG, used as the header's
// media part in TestOpenDocument_CollidingRelationshipIDsAcrossParts.
func encodeTestPNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.Set(x, y, color.RGBA{R: uint8(100 * x), G: uint8(100 * y), B: 50, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}
	return buf.Bytes()
}

// buildCollidingRelationshipDocx hand-assembles a minimal but valid .docx
// package where "rId1" is defined independently in both
// word/_rels/document.xml.rels (pointing at a customXml part) and
// word/_rels/header1.xml.rels (pointing at the header's image), mirroring
// the file attached to GitHub issue #37.
func buildCollidingRelationshipDocx(t *testing.T, pngBytes []byte) []byte {
	t.Helper()

	const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Default Extension="png" ContentType="image/png"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
<Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>
</Types>`

	const rootRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

	const documentXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"
            xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<w:body>
<w:p><w:r><w:t>Body paragraph</w:t></w:r></w:p>
<w:sectPr>
<w:headerReference w:type="default" r:id="rId7"/>
<w:pgSz w:w="11906" w:h="16838"/>
</w:sectPr>
</w:body>
</w:document>`

	// The collision: rId1 here targets a customXml part, while rId1 in
	// header1.xml.rels (below) targets the header's image. Per OOXML both
	// are valid since relationship IDs are scoped to their owning part.
	const documentRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/customXml" Target="../customXml/item1.xml"/>
<Relationship Id="rId7" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="header1.xml"/>
</Relationships>`

	const headerXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:hdr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"
       xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"
       xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
       xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
       xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture">
<w:p>
<w:r>
<w:drawing>
<wp:inline distT="0" distB="0" distL="0" distR="0">
<wp:extent cx="655320" cy="655320"/>
<wp:docPr id="1" name="Picture 1"/>
<a:graphic>
<a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture">
<pic:pic>
<pic:nvPicPr>
<pic:cNvPr id="1" name="Picture 1"/>
<pic:cNvPicPr>
<a:picLocks noChangeAspect="1" noChangeArrowheads="1"/>
</pic:cNvPicPr>
</pic:nvPicPr>
<pic:blipFill>
<a:blip r:embed="rId1"/>
<a:stretch>
<a:fillRect/>
</a:stretch>
</pic:blipFill>
<pic:spPr bwMode="auto">
<a:xfrm>
<a:off x="0" y="0"/>
<a:ext cx="655320" cy="655320"/>
</a:xfrm>
<a:prstGeom prst="rect">
<a:avLst/>
</a:prstGeom>
</pic:spPr>
</pic:pic>
</a:graphicData>
</a:graphic>
</wp:inline>
</w:drawing>
</w:r>
</w:p>
</w:hdr>`

	// header1.xml.rels has its OWN rId1, distinct from document.xml.rels's rId1.
	const headerRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/image1.png"/>
</Relationships>`

	const customXMLItem = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><root/>`

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	files := []struct {
		name string
		data []byte
	}{
		{"[Content_Types].xml", []byte(contentTypesXML)},
		{"_rels/.rels", []byte(rootRelsXML)},
		{"word/document.xml", []byte(documentXML)},
		{"word/_rels/document.xml.rels", []byte(documentRelsXML)},
		{"word/header1.xml", []byte(headerXML)},
		{"word/_rels/header1.xml.rels", []byte(headerRelsXML)},
		{"word/media/image1.png", pngBytes},
		{"customXml/item1.xml", []byte(customXMLItem)},
	}

	for _, f := range files {
		w, err := zw.Create(f.name)
		if err != nil {
			t.Fatalf("zip.Create(%s): %v", f.name, err)
		}
		if _, err := w.Write(f.data); err != nil {
			t.Fatalf("write %s: %v", f.name, err)
		}
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("zip.Close: %v", err)
	}

	return buf.Bytes()
}

// TestOpenDocument_PreservesHeaderRelsBytesOnUntouchedResave is a regression
// for the round-trip property that PR 2 (per-header/footer relationships,
// tracked as a follow-up to #101/#102) is about to put at risk: a header's
// own word/_rels/header1.xml.rels currently survives a save only
// incidentally, as an opaque entry in PreservedParts.Additional (it doesn't
// match the "word/header"/"word/footer" prefix isKnownPart uses to route
// content into PreservedParts.Headers/Footers, so it fell through to the
// generic bucket). Nothing previously asserted this as a property in its own
// right -- TestOpenDocument_CollidingRelationshipIDsAcrossParts, which uses
// the same fixture, only calls Validate() after a round-trip, which cannot
// catch either of the two ways this could silently break once headers start
// being regenerated instead of always preserved verbatim: (1) the rels part
// going missing or changing bytes, or (2) a duplicate zip entry under the
// same name (archive/zip does not dedupe; a corrupt/ambiguous package that
// most tools still open, silently preferring one of the two copies).
func TestOpenDocument_PreservesHeaderRelsBytesOnUntouchedResave(t *testing.T) {
	pngBytes := encodeTestPNG(t)
	docxBytes := buildCollidingRelationshipDocx(t, pngBytes)

	doc, err := OpenDocumentFromBytes(docxBytes)
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	resaved := buf.Bytes()

	origRels := zipPart(t, docxBytes, "word/_rels/header1.xml.rels")
	resavedRels := zipPart(t, resaved, "word/_rels/header1.xml.rels")
	if !bytes.Equal(origRels, resavedRels) {
		t.Errorf("word/_rels/header1.xml.rels changed on an untouched resave:\noriginal: %s\nresaved:  %s", origRels, resavedRels)
	}

	if got := countZipEntries(t, resaved, "word/_rels/header1.xml.rels"); got != 1 {
		t.Errorf("resaved package contains %d entries named word/_rels/header1.xml.rels, want exactly 1", got)
	}

	origCT := zipPart(t, docxBytes, "[Content_Types].xml")
	resavedCT := zipPart(t, resaved, "[Content_Types].xml")
	if !bytes.Equal(origCT, resavedCT) {
		t.Errorf("[Content_Types].xml changed on an untouched resave:\noriginal: %s\nresaved:  %s", origCT, resavedCT)
	}
}

// countZipEntries returns how many entries in a .docx archive are named exactly name.
func countZipEntries(t *testing.T, docxBytes []byte, name string) int {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	count := 0
	for _, f := range zr.File {
		if f.Name == name {
			count++
		}
	}
	return count
}

// TestOpenDocument_AddHyperlink_PreservesForeignRelationshipIDs pins a fix
// found in this PR's own review: NewDocument() -- used internally to build
// the domain.Document that ReconstructDocument hydrates into -- used to
// assign rId1..rId5 to the base relationships (styles, fontTable, theme,
// settings, webSettings) immediately at construction, before the source
// file's own relationships were ever registered. A real Word document has no
// reason to number its relationships in that order; its rId1 is whatever it
// happened to add first, which is routinely a header or a hyperlink, not
// styles.xml.
//
// RegisterExistingRelationship is a no-op when the ID is already taken (by
// design, so re-registering the same relationship twice is harmless), so the
// source file's real rId1 relationship was silently DROPPED instead of
// registered whenever it collided with one of NewDocument's freshly-minted
// IDs. That corruption was invisible as long as document.xml.rels was always
// written back verbatim on save -- but this PR's own fix for issue #101 adds
// a case where it is regenerated from the relationship manager instead (see
// docRelsNeedsRegeneration in internal/core/document.go), which exposed it:
// rId1 in the *regenerated* file no longer matched what word/document.xml's
// own w:headerReference expected, silently repointing the header at
// styles.xml.
//
// The fix is internal/core.NewDocumentForReconstruction, which defers
// assigning the base relationships until Document.WriteTo runs -- by which
// point the source file's relationships are already registered under their
// real IDs, so nothing collides.
func TestOpenDocument_AddHyperlink_PreservesForeignRelationshipIDs(t *testing.T) {
	const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
<Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>
</Types>`

	const rootRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

	// rId1 is the header here -- not styles, which is what a fresh
	// NewDocument() would hand out to rId1 if it ran before this file's own
	// relationships were registered.
	const documentRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="` + constants.RelTypeHeader + `" Target="header1.xml"/>
<Relationship Id="rId2" Type="` + constants.RelTypeStyles + `" Target="styles.xml"/>
</Relationships>`

	const mainDocumentXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"
            xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<w:body>
<w:p><w:r><w:t>Body paragraph</w:t></w:r></w:p>
<w:sectPr>
<w:headerReference w:type="default" r:id="rId1"/>
<w:pgSz w:w="11906" w:h="16838"/>
</w:sectPr>
</w:body>
</w:document>`

	const headerXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:hdr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:p><w:r><w:t>Header text</w:t></w:r></w:p>
</w:hdr>`

	const stylesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
</w:styles>`

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, f := range []struct {
		name string
		data string
	}{
		{"[Content_Types].xml", contentTypesXML},
		{"_rels/.rels", rootRelsXML},
		{"word/document.xml", mainDocumentXML},
		{"word/_rels/document.xml.rels", documentRelsXML},
		{"word/header1.xml", headerXML},
		{"word/styles.xml", stylesXML},
	} {
		w, err := zw.Create(f.name)
		if err != nil {
			t.Fatalf("zip.Create(%s): %v", f.name, err)
		}
		if _, err := w.Write([]byte(f.data)); err != nil {
			t.Fatalf("write %s: %v", f.name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip.Close: %v", err)
	}

	doc, err := OpenDocumentFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	paras := doc.Paragraphs()
	if len(paras) == 0 {
		t.Fatal("expected at least one paragraph")
	}
	if _, err := paras[0].AddHyperlink("https://example.com/added", "Added link"); err != nil {
		t.Fatalf("AddHyperlink: %v", err)
	}

	var resaved bytes.Buffer
	if _, err := doc.WriteTo(&resaved); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	rels := relsOf(t, zipPart(t, resaved.Bytes(), "word/_rels/document.xml.rels"))
	relByID := make(map[string]string, len(rels))
	for _, rel := range rels {
		relByID[rel.ID] = rel.Type
	}

	if got := relByID["rId1"]; got != constants.RelTypeHeader {
		t.Errorf("resaved rId1 Type = %q, want %q (the header relationship); "+
			"adding a hyperlink must not repoint a foreign document's rId1 at "+
			"docxgo's own default styles.xml relationship", got, constants.RelTypeHeader)
	}

	resavedDocumentXML := string(zipPart(t, resaved.Bytes(), "word/document.xml"))
	if !strings.Contains(resavedDocumentXML, `w:headerReference`) || !strings.Contains(resavedDocumentXML, `r:id="rId1"`) {
		t.Errorf("resaved document.xml should still reference the header via rId1:\n%s", resavedDocumentXML)
	}
}

// TestOpenDocument_MalformedHeaderRelsIsTolerated guards against a regression:
// a header/footer part's own .rels was previously never parsed, so a document
// with an unparseable "word/_rels/header1.xml.rels" still opened. Parsing those
// .rels for per-part relationship scoping (issue #37) must not make an
// otherwise-valid document unreadable just because one part's .rels is corrupt;
// the corrupt part's rels is skipped and parsing proceeds.
//
// The header is rewritten to text-only so the test isolates parse-time
// tolerance: a header that references media through its own (now unreadable)
// .rels is a separate, harder concern than "don't abort the whole open".
func TestOpenDocument_MalformedHeaderRelsIsTolerated(t *testing.T) {
	const textOnlyHeaderXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:hdr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:p><w:r><w:t>Header text</w:t></w:r></w:p>
</w:hdr>`

	docxBytes := buildCollidingRelationshipDocx(t, encodeTestPNG(t))
	docxBytes = rewriteZipEntry(t, docxBytes, "word/header1.xml", []byte(textOnlyHeaderXML))
	docxBytes = rewriteZipEntry(t, docxBytes, "word/_rels/header1.xml.rels",
		[]byte(`<?xml version="1.0"?><Relationships><unclosed`))

	doc, err := OpenDocumentFromBytes(docxBytes)
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes with malformed header .rels: %v", err)
	}
	if err := doc.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

// TestOpenDocument_HeaderTableSurvivesRoundTrip pins the pre-PR-2b behavior
// side by side with the new one: a table in a header now reaches the domain
// model (header.Tables() is no longer empty for a table-only header, unlike
// before this PR — see the plan for #101's follow-ups), AND an untouched
// resave still writes the header verbatim via preserved bytes (Known
// limitations: the model isn't consulted on resave for an *opened*
// document's header/footer yet, only for one built fresh via AddTable).
func TestOpenDocument_HeaderTableSurvivesRoundTrip(t *testing.T) {
	const tableHeaderXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:hdr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:p><w:r><w:t>Letterhead</w:t></w:r></w:p>
<w:tbl><w:tr><w:tc><w:p><w:r><w:t>Branded Cell</w:t></w:r></w:p></w:tc></w:tr></w:tbl>
</w:hdr>`

	docxBytes := buildCollidingRelationshipDocx(t, encodeTestPNG(t))
	docxBytes = rewriteZipEntry(t, docxBytes, "word/header1.xml", []byte(tableHeaderXML))

	doc, err := OpenDocumentFromBytes(docxBytes)
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	sections := doc.Sections()
	if len(sections) == 0 {
		t.Fatal("no sections")
	}
	header, err := sections[0].Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header(): %v", err)
	}
	tables := header.Tables()
	if len(tables) != 1 {
		t.Fatalf("len(header.Tables()) = %d, want 1 (table now reaches the domain model)", len(tables))
	}
	row, err := tables[0].Row(0)
	if err != nil {
		t.Fatalf("Row(0): %v", err)
	}
	cell, err := row.Cell(0)
	if err != nil {
		t.Fatalf("Cell(0): %v", err)
	}
	if paras := cell.Paragraphs(); len(paras) != 1 || paras[0].Text() != "Branded Cell" {
		t.Fatalf("cell text = %+v, want [\"Branded Cell\"]", paras)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	resavedHeaderXML := string(zipPart(t, buf.Bytes(), "word/header1.xml"))
	if !strings.Contains(resavedHeaderXML, "<w:tbl") || !strings.Contains(resavedHeaderXML, "Branded Cell") {
		t.Errorf("resaved header1.xml lost the table: %s", resavedHeaderXML)
	}
}

// rewriteZipEntry returns a copy of the zip archive in docxBytes with the entry
// named target replaced by newData, leaving every other entry byte-for-byte.
func rewriteZipEntry(t *testing.T, docxBytes []byte, target string, newData []byte) []byte {
	t.Helper()

	zr, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	var replaced bool
	for _, f := range zr.File {
		w, err := zw.Create(f.Name)
		if err != nil {
			t.Fatalf("zip.Create(%s): %v", f.Name, err)
		}

		data := newData
		if f.Name != target {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open %s: %v", f.Name, err)
			}
			original, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("read %s: %v", f.Name, err)
			}
			data = original
		} else {
			replaced = true
		}

		if _, err := w.Write(data); err != nil {
			t.Fatalf("write %s: %v", f.Name, err)
		}
	}

	if !replaced {
		t.Fatalf("rewriteZipEntry: entry %q not found in archive", target)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip.Close: %v", err)
	}

	return buf.Bytes()
}

// TestOpenDocument_PreservedStylesKeepDocumentDefaults pins a deliberate
// boundary of the document-default paragraph spacing: it applies to documents
// docxgo *generates*, not to ones opened from disk.
//
// OpenDocument keeps the source word/styles.xml verbatim for round-trip
// fidelity, and WriteTo writes those bytes back untouched. That is the whole
// point — rewriting a caller's styles.xml would change how their document
// renders, which is exactly what a fidelity-preserving round trip must not do.
// The same constraint is why SetLanguage refuses to run on an opened document
// rather than silently no-op'ing.
//
// The consequence worth pinning: a .docx produced by an older docxgo (whose
// styles.xml has no w:pPrDefault) does not acquire one by being opened and
// re-saved with a newer version. Such documents have to be regenerated, not
// round-tripped, to pick up the new defaults.
func TestOpenDocument_PreservedStylesKeepDocumentDefaults(t *testing.T) {
	// Build a document, then strip its docDefaults to stand in for one
	// generated by a docxgo release that predates the pPrDefault default.
	doc := NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("made by an older version"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	var generated bytes.Buffer
	if _, err := doc.WriteTo(&generated); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	// Sanity check: a freshly generated document does carry the defaults.
	if !bytes.Contains(zipPart(t, generated.Bytes(), "word/styles.xml"), []byte("<w:pPrDefault>")) {
		t.Fatal("setup: expected a freshly generated document to carry w:pPrDefault")
	}

	legacyStyles := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
</w:styles>`)
	legacy := rewriteZipEntry(t, generated.Bytes(), "word/styles.xml", legacyStyles)

	reopened, err := OpenDocumentFromBytes(legacy)
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	var resaved bytes.Buffer
	if _, err := reopened.WriteTo(&resaved); err != nil {
		t.Fatalf("WriteTo (resave): %v", err)
	}

	styles := zipPart(t, resaved.Bytes(), "word/styles.xml")
	if bytes.Contains(styles, []byte("<w:pPrDefault>")) {
		t.Errorf("opened document's preserved styles.xml was rewritten with new document defaults; "+
			"round-trip fidelity requires writing it back verbatim.\ngot: %s", styles)
	}
}

// relsOf unmarshals a word/_rels/*.rels part into a minimal, test-local view
// of its Relationship entries.
func relsOf(t *testing.T, relsXML []byte) []struct {
	ID     string `xml:"Id,attr"`
	Type   string `xml:"Type,attr"`
	Target string `xml:"Target,attr"`
} {
	t.Helper()
	var rels struct {
		Relationships []struct {
			ID     string `xml:"Id,attr"`
			Type   string `xml:"Type,attr"`
			Target string `xml:"Target,attr"`
		} `xml:"Relationship"`
	}
	if err := xml.Unmarshal(relsXML, &rels); err != nil {
		t.Fatalf("unmarshal .rels: %v", err)
	}
	return rels.Relationships
}

var hyperlinkElementRE = regexp.MustCompile(`<w:hyperlink[^>]*\br:id="([^"]+)"[^>]*>`)

// TestAddHyperlink_EmitsRealHyperlinkElement is the end-to-end regression for
// issue #101: Paragraph.AddHyperlink wrote a relationship into
// word/_rels/document.xml.rels that nothing in word/document.xml referenced
// -- Word rendered a plain blue, underlined run instead of a clickable link.
// This asserts the actual bytes docxgo produces, not just the in-memory
// domain.Field the fix attaches.
func TestAddHyperlink_EmitsRealHyperlinkElement(t *testing.T) {
	doc := NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	const url = "https://example.com/policy"
	if _, err := para.AddHyperlink(url, "See the policy"); err != nil {
		t.Fatalf("AddHyperlink: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	documentXML := string(zipPart(t, buf.Bytes(), "word/document.xml"))
	match := hyperlinkElementRE.FindStringSubmatch(documentXML)
	if match == nil {
		t.Fatalf("word/document.xml does not contain a <w:hyperlink r:id=...> element:\n%s", documentXML)
	}
	rID := match[1]

	rels := relsOf(t, zipPart(t, buf.Bytes(), "word/_rels/document.xml.rels"))

	var hyperlinkRelCount int
	var foundMatchingTarget bool
	for _, rel := range rels {
		if rel.Type != constants.RelTypeHyperlink {
			continue
		}
		hyperlinkRelCount++
		if rel.ID == rID && rel.Target == url {
			foundMatchingTarget = true
		}
	}

	if !foundMatchingTarget {
		t.Errorf("no hyperlink relationship with Id=%q Target=%q found in document.xml.rels; got %+v", rID, url, rels)
	}
	if hyperlinkRelCount != 1 {
		t.Errorf("hyperlink relationship count = %d, want 1 (AddHyperlink must not leave an orphaned relationship)", hyperlinkRelCount)
	}

	// Regression for a defect found while planning #101's follow-ups:
	// expandRunWithFields (internal/serializer/serializer.go) serialized the
	// run's display text into the <w:hyperlink>'s own <w:r>, then -- since
	// the run's in-memory text was never cleared -- serialized it a second
	// time as a plain trailing <w:r>. Every prior hyperlink assertion in this
	// repo counted <w:hyperlink> elements or inspected the in-memory model,
	// both blind to a duplicated rendered run, so this counts occurrences of
	// the display text in the actual bytes instead.
	if got := strings.Count(documentXML, "See the policy"); got != 1 {
		t.Errorf("display text %q appears %d times in word/document.xml, want 1 (duplicated trailing run)", "See the policy", got)
	}
}

// TestAddHyperlink_RoundTrip guards against the fix regressing on a document
// that is written, reopened, and written again -- the hyperlink and its
// relationship must survive both trips.
func TestAddHyperlink_RoundTrip(t *testing.T) {
	doc := NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	const url = "https://example.com/policy"
	if _, err := para.AddHyperlink(url, "See the policy"); err != nil {
		t.Fatalf("AddHyperlink: %v", err)
	}

	var first bytes.Buffer
	if _, err := doc.WriteTo(&first); err != nil {
		t.Fatalf("WriteTo (first): %v", err)
	}

	reopened, err := OpenDocumentFromBytes(first.Bytes())
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	var second bytes.Buffer
	if _, err := reopened.WriteTo(&second); err != nil {
		t.Fatalf("WriteTo (second): %v", err)
	}

	documentXML := string(zipPart(t, second.Bytes(), "word/document.xml"))
	match := hyperlinkElementRE.FindStringSubmatch(documentXML)
	if match == nil {
		t.Fatalf("word/document.xml (after round-trip) does not contain a <w:hyperlink r:id=...> element:\n%s", documentXML)
	}

	rels := relsOf(t, zipPart(t, second.Bytes(), "word/_rels/document.xml.rels"))
	var found bool
	for _, rel := range rels {
		if rel.ID == match[1] && rel.Type == constants.RelTypeHyperlink && rel.Target == url {
			found = true
		}
	}
	if !found {
		t.Errorf("round-tripped document.xml.rels missing hyperlink relationship Id=%q Target=%q; got %+v", match[1], url, rels)
	}
}

// TestOpenDocument_AddHyperlink_RegeneratesRelsWithoutLosingOriginal pins the
// issue #101 fix for the round-trip hazard: word/_rels/document.xml.rels is
// normally written back verbatim for a document opened with OpenDocument (see
// docRelsNeedsRegeneration in internal/core/document.go). Adding a hyperlink
// after open mints a new relationship id that the preserved bytes don't
// contain -- if they were still written verbatim, the saved document.xml
// would carry a dangling r:id. The fix regenerates the rels part from the
// relationship manager instead, which was fully seeded from the original
// rels on open, so nothing from the source document is lost.
//
// The "unchanged document" sub-test below only holds because the fixture is
// itself docxgo-authored and so already carries all five of docxgo's default
// relationships (styles, fontTable, theme, settings, webSettings). A source
// file missing one of those still regenerates the rels part on an otherwise
// untouched resave, since Document.WriteTo always ensures they're present
// (unrelated to this fix -- see ensureDefaultRelationships) and the newly
// minted relationship is, correctly, not in the preserved bytes either. See
// TestOpenDocument_AddHyperlink_PreservesForeignRelationshipIDs's fixture,
// which is missing all four non-styles defaults and demonstrates exactly
// that: it still round-trips correctly, just not byte-for-byte.
func TestOpenDocument_AddHyperlink_RegeneratesRelsWithoutLosingOriginal(t *testing.T) {
	doc := NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	const originalURL = "https://example.com/original"
	if _, err := para.AddHyperlink(originalURL, "Original link"); err != nil {
		t.Fatalf("AddHyperlink: %v", err)
	}

	var original bytes.Buffer
	if _, err := doc.WriteTo(&original); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	originalRels := relsOf(t, zipPart(t, original.Bytes(), "word/_rels/document.xml.rels"))

	t.Run("unchanged document keeps rels byte-identical", func(t *testing.T) {
		reopened, err := OpenDocumentFromBytes(original.Bytes())
		if err != nil {
			t.Fatalf("OpenDocumentFromBytes: %v", err)
		}
		var resaved bytes.Buffer
		if _, err := reopened.WriteTo(&resaved); err != nil {
			t.Fatalf("WriteTo: %v", err)
		}
		got := zipPart(t, resaved.Bytes(), "word/_rels/document.xml.rels")
		want := zipPart(t, original.Bytes(), "word/_rels/document.xml.rels")
		if !bytes.Equal(got, want) {
			t.Errorf("resaving an opened document with no new relationships changed document.xml.rels;\ngot:  %s\nwant: %s", got, want)
		}
	})

	t.Run("adding a hyperlink regenerates rels without dropping originals", func(t *testing.T) {
		reopened, err := OpenDocumentFromBytes(original.Bytes())
		if err != nil {
			t.Fatalf("OpenDocumentFromBytes: %v", err)
		}
		paras := reopened.Paragraphs()
		if len(paras) == 0 {
			t.Fatalf("expected at least one paragraph in the reopened document")
		}
		const newURL = "https://example.com/added-after-open"
		if _, err := paras[0].AddHyperlink(newURL, "Added after open"); err != nil {
			t.Fatalf("AddHyperlink (after open): %v", err)
		}

		var resaved bytes.Buffer
		if _, err := reopened.WriteTo(&resaved); err != nil {
			t.Fatalf("WriteTo: %v", err)
		}

		documentXML := string(zipPart(t, resaved.Bytes(), "word/document.xml"))
		matches := hyperlinkElementRE.FindAllStringSubmatch(documentXML, -1)
		if len(matches) != 2 {
			t.Fatalf("found %d <w:hyperlink r:id> elements in resaved document.xml, want 2:\n%s", len(matches), documentXML)
		}

		rels := relsOf(t, zipPart(t, resaved.Bytes(), "word/_rels/document.xml.rels"))
		relByID := make(map[string]string, len(rels))
		for _, rel := range rels {
			relByID[rel.ID] = rel.Target
		}

		var foundOriginal, foundNew bool
		for _, m := range matches {
			switch relByID[m[1]] {
			case originalURL:
				foundOriginal = true
			case newURL:
				foundNew = true
			}
		}
		if !foundOriginal {
			t.Errorf("resaved document.xml.rels lost the original relationship (Target=%q); got %+v", originalURL, rels)
		}
		if !foundNew {
			t.Errorf("resaved document.xml.rels missing the newly added relationship (Target=%q); got %+v", newURL, rels)
		}

		// Every non-hyperlink relationship from the original save must still
		// be present too (styles, fontTable, theme, settings, webSettings, ...).
		origByID := make(map[string]string, len(originalRels))
		for _, rel := range originalRels {
			origByID[rel.ID] = rel.Target
		}
		for id, target := range origByID {
			if got, ok := relByID[id]; !ok || got != target {
				t.Errorf("resaved rels lost or changed original relationship %s (Target=%q): got Target=%q, ok=%v", id, target, got, ok)
			}
		}
	})
}

// TestOpenDocument_PreservesRunCaps is the end-to-end regression for one of
// the three losses reported in issue #102: opening a real, Word-authored
// document whose title uses "All Caps" character formatting (<w:caps>, a
// display override -- the stored text is genuinely mixed-case) and resaving
// it dropped the formatting, so the title would render in its literal
// mixed-case text instead of forced uppercase.
//
// testdata/word/issue-102-input.docx is the reporter's own attachment. Only
// the w:caps loss is asserted here; the section-break and table-style losses
// reported in the same issue are fixed separately.
func TestOpenDocument_PreservesRunCaps(t *testing.T) {
	doc, err := OpenDocument(filepath.Join("internal", "reader", "testdata", "word", "issue-102-input.docx"))
	if err != nil {
		t.Fatalf("OpenDocument: %v", err)
	}

	var titleRun domain.Run
	for _, para := range doc.Paragraphs() {
		for _, run := range para.Runs() {
			if strings.Contains(run.Text(), "TiTlE") {
				titleRun = run
			}
		}
	}
	if titleRun == nil {
		t.Fatal("did not find the title run (looked for text containing \"TiTlE\")")
	}
	if !titleRun.Caps() {
		t.Error("title run Caps() = false, want true")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	written := string(zipPart(t, buf.Bytes(), "word/document.xml"))
	if !strings.Contains(written, "<w:caps") {
		t.Errorf("resaved document.xml lost the title's w:caps:\n%s", written)
	}
}

// TestOpenDocument_PreservesTableStyle is the end-to-end regression for one
// of the three losses reported in issue #102: opening a real, Word-authored
// document whose table borders come from a named table style (<w:tblStyle>,
// not an explicit <w:tblBorders> on the table itself) and resaving it
// dropped the style reference, orphaning the borders even though the style
// definition itself survived untouched in styles.xml.
//
// testdata/word/issue-102-input.docx is the reporter's own attachment. Only
// the table-style loss is asserted here; the section-break and w:caps losses
// reported in the same issue are separate, not fixed here.
func TestOpenDocument_PreservesTableStyle(t *testing.T) {
	doc, err := OpenDocument(filepath.Join("internal", "reader", "testdata", "word", "issue-102-input.docx"))
	if err != nil {
		t.Fatalf("OpenDocument: %v", err)
	}

	tables := doc.Tables()
	if len(tables) != 1 {
		t.Fatalf("len(Tables()) = %d, want 1", len(tables))
	}
	if got := tables[0].Style().Name; got != "TableGrid" {
		t.Errorf("Tables()[0].Style().Name = %q, want %q", got, "TableGrid")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	written := string(zipPart(t, buf.Bytes(), "word/document.xml"))
	if !strings.Contains(written, `<w:tblStyle w:val="TableGrid">`) {
		t.Errorf("resaved document.xml lost the table's tblStyle reference:\n%s", written)
	}
}

// TestOpenDocument_PreservesMidBodySectionBreak is the end-to-end regression
// for one of the three losses reported in issue #102: opening a real,
// Word-authored document with a section break in the middle of its body (a
// <w:sectPr> embedded in a paragraph's own pPr, not the body's last child)
// and resaving it dropped the break, collapsing two sections into one.
//
// testdata/word/issue-102-input.docx is the reporter's own attachment: a
// title page, a mid-body section break (with no explicit w:sectPr/w:type --
// its schema default is "nextPage"), and a table using a named style for its
// borders. Only the section-break loss is asserted here; the table style and
// w:caps losses reported in the same issue are separate, not yet fixed here.
func TestOpenDocument_PreservesMidBodySectionBreak(t *testing.T) {
	doc, err := OpenDocument(filepath.Join("internal", "reader", "testdata", "word", "issue-102-input.docx"))
	if err != nil {
		t.Fatalf("OpenDocument: %v", err)
	}

	if got := len(doc.Sections()); got != 2 {
		t.Fatalf("len(Sections()) = %d, want 2 (the source has a mid-body section break)", got)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	written := string(zipPart(t, buf.Bytes(), "word/document.xml"))
	if got := strings.Count(written, "<w:sectPr"); got != 2 {
		t.Errorf("resaved document.xml contains %d <w:sectPr> elements, want 2 (the mid-body break plus the document's final section)", got)
	}
}

// TestOpenDocument_PreservesTableCellWidths is the end-to-end regression for
// what was left of issue #102 after the table style, the mid-body section
// break and w:caps were fixed: the reporter's table lays its two columns out
// with explicit widths (918 and 1111 twips, in both <w:tcW> and <w:tblGrid>),
// and resaving collapsed every one of them to auto -- the table came back
// evenly split at whatever width Word chose.
//
// testdata/word/issue-102-input.docx is the reporter's own attachment.
func TestOpenDocument_PreservesTableCellWidths(t *testing.T) {
	doc, err := OpenDocument(filepath.Join("internal", "reader", "testdata", "word", "issue-102-input.docx"))
	if err != nil {
		t.Fatalf("OpenDocument: %v", err)
	}

	tables := doc.Tables()
	if len(tables) != 1 {
		t.Fatalf("len(Tables()) = %d, want 1", len(tables))
	}

	row, err := tables[0].Row(0)
	if err != nil {
		t.Fatalf("Row(0): %v", err)
	}
	for col, want := range []int{918, 1111} {
		cell, err := row.Cell(col)
		if err != nil {
			t.Fatalf("Cell(%d): %v", col, err)
		}
		if got := cell.Width(); got != want {
			t.Errorf("Row(0).Cell(%d).Width() = %d, want %d", col, got, want)
		}
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	written := string(zipPart(t, buf.Bytes(), "word/document.xml"))
	for _, want := range []string{
		`<w:tcW w:type="dxa" w:w="918">`,
		`<w:tcW w:type="dxa" w:w="1111">`,
		`<w:gridCol w:w="918">`,
		`<w:gridCol w:w="1111">`,
	} {
		if !strings.Contains(written, want) {
			t.Errorf("resaved document.xml lost %s:\n%s", want, written)
		}
	}
}

// TestOpenDocument_OnlyTheEditedHeaderIsRegenerated is the core of the
// per-name merge. WriteTo used to be all-or-nothing: one preserved header
// discarded the entire generated map, so a header edited on an opened
// document was silently dropped on save. Regenerating everything instead
// would be just as wrong in the other direction -- a regenerated part is
// rebuilt from what docxgo can model, so anything the reader does not
// understand is lost from it.
//
// Both halves are asserted here: the edited header carries the new text, and
// the untouched one is byte-for-byte what it was.
func TestOpenDocument_OnlyTheEditedHeaderIsRegenerated(t *testing.T) {
	built := docxWithTwoHeaders(t)

	doc, err := OpenDocumentFromBytes(built)
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	section := doc.Sections()[0]
	header, err := section.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header: %v", err)
	}
	paras := header.Paragraphs()
	if len(paras) != 1 {
		t.Fatalf("default header has %d paragraphs, want 1", len(paras))
	}
	runs := paras[0].Runs()
	if len(runs) != 1 {
		t.Fatalf("default header paragraph has %d runs, want 1", len(runs))
	}
	if err := runs[0].SetText("edited default header"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	resaved := buf.Bytes()

	edited := string(zipPart(t, resaved, "word/header1.xml"))
	if !strings.Contains(edited, "edited default header") {
		t.Errorf("header1.xml did not pick up the edit:\n%s", edited)
	}

	if before, after := zipPart(t, built, "word/header2.xml"), zipPart(t, resaved, "word/header2.xml"); !bytes.Equal(before, after) {
		t.Errorf("header2.xml was regenerated even though nothing touched it:\nbefore: %s\nafter:  %s", before, after)
	}

	for _, name := range []string{"word/header1.xml", "word/header2.xml"} {
		if got := countZipEntries(t, resaved, name); got != 1 {
			t.Errorf("resaved package contains %d entries named %s, want exactly 1", got, name)
		}
	}
}

// TestOpenDocument_HeaderAddedAfterOpenGetsContentTypeOverride covers the
// second mutation surface. [Content_Types].xml is written back verbatim on a
// round-trip, which is correct while the package's part list cannot change --
// but a header *added* to an opened document then has no Override, and a part
// with no declared content type makes the whole package invalid.
func TestOpenDocument_HeaderAddedAfterOpenGetsContentTypeOverride(t *testing.T) {
	built := docxWithTwoHeaders(t)

	doc, err := OpenDocumentFromBytes(built)
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	section := doc.Sections()[0]
	evenHeader, err := section.Header(domain.HeaderEven)
	if err != nil {
		t.Fatalf("Header(HeaderEven): %v", err)
	}
	para, err := evenHeader.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("brand new header"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	resaved := buf.Bytes()

	// The new part is header3.xml -- header1 and header2 are taken.
	added := string(zipPart(t, resaved, "word/header3.xml"))
	if !strings.Contains(added, "brand new header") {
		t.Errorf("header3.xml is missing the new content:\n%s", added)
	}

	contentTypes := string(zipPart(t, resaved, "[Content_Types].xml"))
	if !strings.Contains(contentTypes, `PartName="/word/header3.xml"`) {
		t.Errorf("[Content_Types].xml has no Override for the newly added header:\n%s", contentTypes)
	}
	for _, existing := range []string{"/word/header1.xml", "/word/header2.xml", "/word/document.xml"} {
		if !strings.Contains(contentTypes, `PartName="`+existing+`"`) {
			t.Errorf("[Content_Types].xml lost the Override for %s:\n%s", existing, contentTypes)
		}
	}
}

// TestOpenDocument_AbsoluteHeaderTargetIsNormalized covers a hazard that only
// became reachable once headers regenerate on opened documents. A header's
// target comes from the source file's own rels, where "/word/header1.xml" is
// as legal as "header1.xml"; the writer used to build the entry name by
// blindly prepending "word/", which for the absolute form yields
// "word//word/header1.xml" -- a part nothing references, and no header where
// one is expected.
func TestOpenDocument_AbsoluteHeaderTargetIsNormalized(t *testing.T) {
	built := docxWithHeaderTarget(t, "/word/header1.xml")

	doc, err := OpenDocumentFromBytes(built)
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	section := doc.Sections()[0]
	header, err := section.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header: %v", err)
	}
	paras := header.Paragraphs()
	if len(paras) != 1 {
		t.Fatalf("header has %d paragraphs, want 1 (was it hydrated at all?)", len(paras))
	}
	if err := paras[0].Runs()[0].SetText("edited"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	resaved := buf.Bytes()

	if got := countZipEntries(t, resaved, "word//word/header1.xml"); got != 0 {
		t.Errorf("resaved package contains %d entries named word//word/header1.xml -- the target was not normalized", got)
	}
	edited := string(zipPart(t, resaved, "word/header1.xml"))
	if !strings.Contains(edited, "edited") {
		t.Errorf("word/header1.xml did not pick up the edit:\n%s", edited)
	}
}

// TestOpenDocument_EditedHeaderKeepsItsImageAndRels is the relationship half
// of regeneration. Writing new header content next to the *preserved* rels
// would leave the r:ids the model minted pointing at nothing -- the same
// dangling-r:id package #101 was about, one part further down. The generated
// rels must win for a regenerated part, and win exclusively: writeRaw is a
// bare zip.Create, so writing both would leave two entries under one name and
// let Word pick.
func TestOpenDocument_EditedHeaderKeepsItsImageAndRels(t *testing.T) {
	pngBytes := encodeTestPNG(t)
	built := buildCollidingRelationshipDocx(t, pngBytes)

	doc, err := OpenDocumentFromBytes(built)
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	header, err := doc.Sections()[0].Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header: %v", err)
	}
	para, err := header.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("appended line"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	resaved := buf.Bytes()

	headerXML := string(zipPart(t, resaved, "word/header1.xml"))
	if !strings.Contains(headerXML, "appended line") {
		t.Errorf("header1.xml did not pick up the edit:\n%s", headerXML)
	}

	if got := countZipEntries(t, resaved, "word/_rels/header1.xml.rels"); got != 1 {
		t.Fatalf("resaved package contains %d entries named word/_rels/header1.xml.rels, want exactly 1", got)
	}
	headerRels := string(zipPart(t, resaved, "word/_rels/header1.xml.rels"))

	embeds := regexp.MustCompile(`r:embed="([^"]+)"`).FindAllStringSubmatch(headerXML, -1)
	if len(embeds) != 1 {
		t.Fatalf("regenerated header1.xml has %d r:embed references, want 1 (the image should survive regeneration):\n%s", len(embeds), headerXML)
	}
	for _, embed := range embeds {
		if !strings.Contains(headerRels, `Id="`+embed[1]+`"`) {
			t.Errorf("header1.xml references %s but word/_rels/header1.xml.rels does not declare it:\n%s", embed[1], headerRels)
		}
	}
	if !strings.Contains(headerRels, "media/image1.png") {
		t.Errorf("word/_rels/header1.xml.rels lost the image target:\n%s", headerRels)
	}
}

// TestOpenDocument_UntouchedHeaderIsNotDirtiedByABodyImage guards the
// dirty-detection itself rather than what it triggers. Regeneration is decided
// by comparing the model's serialization against a snapshot taken at
// hydration, so anything that makes the same untouched header serialize
// differently in those two passes silently regenerates it -- and a regenerated
// part loses whatever the reader could not model, which is the exact cost the
// per-name merge exists to avoid paying on parts nobody edited.
//
// The trap is shared serializer state: wp:docPr ids come from a counter on
// DocumentSerializer, and WriteTo serializes the body before the section
// parts while the snapshot serializes only the section parts. Give the body an
// image and the same header drawing gets a different id in each pass, so the
// header reads as edited when nothing touched it.
//
// The fixture is hand-authored on purpose. Built through docxgo's own writer
// the regenerated bytes could coincidentally equal the originals and the
// comparison would pass while still regenerating; hand-authored bytes never
// match docxgo's serialization, so a regeneration is always visible here.
func TestOpenDocument_UntouchedHeaderIsNotDirtiedByABodyImage(t *testing.T) {
	built := docxWithBodyAndHeaderImages(t)

	doc, err := OpenDocumentFromBytes(built)
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	before, after := zipPart(t, built, "word/header1.xml"), zipPart(t, buf.Bytes(), "word/header1.xml")
	if !bytes.Equal(before, after) {
		t.Errorf("header1.xml was regenerated on an untouched resave:\nbefore: %s\n\nafter:  %s", before, after)
	}
}

// TestOpenDocument_EditingTheBodyDoesNotDirtyAHeader is the case that decides
// how the divergence above has to be fixed. Making the snapshot mirror
// WriteTo's serialization order would satisfy the untouched-resave test but
// not this one: the caller is free to change the body between opening the
// document and saving it, and once that moves the shared drawing counter the
// two passes disagree again. Numbering each part's drawings from 1,
// independently of the body, is what holds for both.
//
// Editing the body is also the most ordinary thing a caller does to an opened
// document, so a header quietly losing content because of it would be the
// common case rather than an edge one.
func TestOpenDocument_EditingTheBodyDoesNotDirtyAHeader(t *testing.T) {
	built := docxWithBodyAndHeaderImages(t)

	doc, err := OpenDocumentFromBytes(built)
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	if _, err := para.AddImageFromBytes(encodeTestPNG(t), domain.ImageFormatPNG); err != nil {
		t.Fatalf("AddImageFromBytes: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	resaved := buf.Bytes()

	if body := string(zipPart(t, resaved, "word/document.xml")); strings.Count(body, "<w:drawing>") != 2 {
		t.Errorf("body should hold both the original and the added image, got:\n%s", body)
	}

	before, after := zipPart(t, built, "word/header1.xml"), zipPart(t, resaved, "word/header1.xml")
	if !bytes.Equal(before, after) {
		t.Errorf("adding an image to the body regenerated the untouched header1.xml:\nbefore: %s\n\nafter:  %s", before, after)
	}
}

// TestOpenDocument_EditedHeaderKeepsItsHyperlinkRels is the hyperlink twin of
// TestOpenDocument_EditedHeaderKeepsItsImageAndRels, and it did not follow for
// free. A hydrated image is attached through AttachHydratedImageToRun, which
// registers its relationship into the owning part's manager; a hydrated
// hyperlink keeps the source's r:id on the field and registers nothing.
//
// That was invisible while headers were never regenerated, and invisible in
// the body regardless, because the document's own manager is loaded with every
// source relationship at hydration. A header owns a separate, initially empty
// manager, so once the part is rebuilt it emits <w:hyperlink r:id="rId1">
// against a .rels that declares nothing -- the dangling r:id Word offers to
// repair, one part below the one #101 was about.
func TestOpenDocument_EditedHeaderKeepsItsHyperlinkRels(t *testing.T) {
	built := docxWithHeaderHyperlink(t)

	doc, err := OpenDocumentFromBytes(built)
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	header, err := doc.Sections()[0].Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header: %v", err)
	}
	para, err := header.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("appended line"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	resaved := buf.Bytes()

	headerXML := string(zipPart(t, resaved, "word/header1.xml"))
	if !strings.Contains(headerXML, "appended line") {
		t.Fatalf("header1.xml did not pick up the edit, so this test is not exercising regeneration:\n%s", headerXML)
	}

	refs := regexp.MustCompile(`<w:hyperlink[^>]*r:id="([^"]+)"`).FindAllStringSubmatch(headerXML, -1)
	if len(refs) != 1 {
		t.Fatalf("regenerated header1.xml has %d hyperlink r:id references, want 1:\n%s", len(refs), headerXML)
	}

	if got := countZipEntries(t, resaved, "word/_rels/header1.xml.rels"); got != 1 {
		t.Fatalf("resaved package contains %d entries named word/_rels/header1.xml.rels, want exactly 1", got)
	}
	headerRels := string(zipPart(t, resaved, "word/_rels/header1.xml.rels"))
	if !strings.Contains(headerRels, `Id="`+refs[0][1]+`"`) {
		t.Errorf("header1.xml references %s but word/_rels/header1.xml.rels does not declare it:\n%s", refs[0][1], headerRels)
	}
	if !strings.Contains(headerRels, "https://example.com/") {
		t.Errorf("word/_rels/header1.xml.rels lost the hyperlink target:\n%s", headerRels)
	}
}

// TestOpenDocument_HeaderInASubdirectoryIsNotDuplicated covers the other shape
// a relationship target can legally take. "headers/header1.xml" resolves under
// the owning part's directory, so the part lives at
// word/headers/header1.xml -- but collapsing the target to its base name calls
// it word/header1.xml, which is not where it is.
//
// The damage is not only a lost edit. Nothing preserved claims the collapsed
// name, so the writer treats the part as new and writes it there
// unconditionally, then writes the real part from its preserved bytes as well:
// two headers in the package, the referenced one still holding the old text.
// Both come from regenerating headers at all, so both are new here.
func TestOpenDocument_HeaderInASubdirectoryIsNotDuplicated(t *testing.T) {
	built := docxWithHeaderAt(t, "headers/header1.xml", "word/headers/header1.xml")

	doc, err := OpenDocumentFromBytes(built)
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}

	header, err := doc.Sections()[0].Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header: %v", err)
	}
	paras := header.Paragraphs()
	if len(paras) != 1 {
		t.Fatalf("header has %d paragraphs, want 1 (was it hydrated at all?)", len(paras))
	}
	if err := paras[0].Runs()[0].SetText("edited"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	resaved := buf.Bytes()

	if got := countZipEntries(t, resaved, "word/header1.xml"); got != 0 {
		t.Errorf("resaved package invented %d entries at word/header1.xml; the header lives in a subdirectory and nothing references that path", got)
	}
	if got := countZipEntries(t, resaved, "word/headers/header1.xml"); got != 1 {
		t.Fatalf("resaved package contains %d entries named word/headers/header1.xml, want exactly 1", got)
	}
	if got := string(zipPart(t, resaved, "word/headers/header1.xml")); !strings.Contains(got, "edited") {
		t.Errorf("the referenced header did not pick up the edit:\n%s", got)
	}

	// No name in the package may be written twice: writeRaw is a bare
	// zip.Create and archive/zip accepts a duplicate silently, leaving Word
	// to pick between them.
	assertNoDuplicateEntries(t, resaved)
}

// assertNoDuplicateEntries fails if any archive name appears more than once.
func assertNoDuplicateEntries(t *testing.T, docxBytes []byte) {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	seen := make(map[string]int, len(zr.File))
	for _, f := range zr.File {
		seen[f.Name]++
	}
	for name, n := range seen {
		if n > 1 {
			t.Errorf("archive holds %d entries named %s, want exactly 1", n, name)
		}
	}
}

// TestOpenDocument_FirstImageAddedGetsItsContentTypeDefault covers the other
// way an opened document's part list can outgrow its preserved
// [Content_Types].xml. A header added after opening needs an Override; an
// image added to a document whose source held none needs a Default for its
// extension. The media part is written regardless, so without the Default the
// package declares a part it cannot type, and Word offers to repair it.
//
// This one is not a regression from regenerating headers -- it reproduces on
// every released version through OpenDocument + AddImage -- but amendContentTypes
// is where the preserved part is now adjusted, so it is where the gap closes.
func TestOpenDocument_FirstImageAddedGetsItsContentTypeDefault(t *testing.T) {
	// A source document with no media at all, so its [Content_Types].xml has
	// no reason to declare png.
	base := NewDocument()
	para, err := base.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("body"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	var src bytes.Buffer
	if _, err := base.WriteTo(&src); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	if ct := string(zipPart(t, src.Bytes(), "[Content_Types].xml")); strings.Contains(ct, `Extension="png"`) {
		t.Fatalf("source already declares png, so this test proves nothing:\n%s", ct)
	}

	doc, err := OpenDocumentFromBytes(src.Bytes())
	if err != nil {
		t.Fatalf("OpenDocumentFromBytes: %v", err)
	}
	imgPara, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	if _, err := imgPara.AddImageFromBytes(encodeTestPNG(t), domain.ImageFormatPNG); err != nil {
		t.Fatalf("AddImageFromBytes: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	resaved := buf.Bytes()

	if got := zipPart(t, resaved, "word/media/image1.png"); len(got) == 0 {
		t.Fatal("no media part was written, so the content type is not the thing under test")
	}
	if ct := string(zipPart(t, resaved, "[Content_Types].xml")); !strings.Contains(ct, `Extension="png"`) {
		t.Errorf("resaved package writes word/media/image1.png but declares no content type for it:\n%s", ct)
	}
}

// docxWithHeaderHyperlink builds a hand-authored package whose default header
// holds one external hyperlink, declared in the header's own .rels.
func docxWithHeaderHyperlink(t *testing.T) []byte {
	t.Helper()

	const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
<Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>
</Types>`

	const rootRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

	const documentXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"
            xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<w:body>
<w:p><w:r><w:t>Body paragraph</w:t></w:r></w:p>
<w:sectPr>
<w:headerReference w:type="default" r:id="rId7"/>
<w:pgSz w:w="11906" w:h="16838"/>
</w:sectPr>
</w:body>
</w:document>`

	const documentRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId7" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="header1.xml"/>
</Relationships>`

	const headerXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:hdr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"
       xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<w:p>
<w:hyperlink r:id="rId1">
<w:r><w:rPr><w:rStyle w:val="Hyperlink"/></w:rPr><w:t>Our site</w:t></w:r>
</w:hyperlink>
</w:p>
</w:hdr>`

	// rId1 here is the header's own, unrelated to document.xml.rels's rId1.
	const headerRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink" Target="https://example.com/" TargetMode="External"/>
</Relationships>`

	return buildZip(t, map[string][]byte{
		"[Content_Types].xml":          []byte(contentTypesXML),
		"_rels/.rels":                  []byte(rootRelsXML),
		"word/document.xml":            []byte(documentXML),
		"word/_rels/document.xml.rels": []byte(documentRelsXML),
		"word/header1.xml":             []byte(headerXML),
		"word/_rels/header1.xml.rels":  []byte(headerRelsXML),
	})
}

// docxWithBodyAndHeaderImages builds a hand-authored package holding one image
// in the body and another in the default header, both pointing at the same
// media part through their own part-scoped relationship IDs.
func docxWithBodyAndHeaderImages(t *testing.T) []byte {
	t.Helper()

	// drawing renders a <w:drawing> referencing relID. Both parts embed the
	// same media file; only the r:id differs, since each part resolves it
	// through its own .rels.
	drawing := func(relID string) string {
		return `<w:drawing>
<wp:inline distT="0" distB="0" distL="0" distR="0">
<wp:extent cx="655320" cy="655320"/>
<wp:docPr id="1" name="Picture 1"/>
<a:graphic>
<a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture">
<pic:pic>
<pic:nvPicPr><pic:cNvPr id="1" name="Picture 1"/><pic:cNvPicPr/></pic:nvPicPr>
<pic:blipFill><a:blip r:embed="` + relID + `"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill>
<pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="655320" cy="655320"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr>
</pic:pic>
</a:graphicData>
</a:graphic>
</wp:inline>
</w:drawing>`
	}

	const drawingNS = ` xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"
 xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"
 xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
 xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
 xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"`

	contentTypesXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Default Extension="png" ContentType="image/png"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
<Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>
</Types>`

	rootRelsXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

	documentXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document` + drawingNS + `>
<w:body>
<w:p><w:r>` + drawing("rId2") + `</w:r></w:p>
<w:sectPr>
<w:headerReference w:type="default" r:id="rId7"/>
<w:pgSz w:w="11906" w:h="16838"/>
</w:sectPr>
</w:body>
</w:document>`

	documentRelsXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/image1.png"/>
<Relationship Id="rId7" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="header1.xml"/>
</Relationships>`

	headerXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:hdr` + drawingNS + `>
<w:p><w:r>` + drawing("rId1") + `</w:r></w:p>
</w:hdr>`

	headerRelsXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/image1.png"/>
</Relationships>`

	return buildZip(t, map[string][]byte{
		"[Content_Types].xml":          []byte(contentTypesXML),
		"_rels/.rels":                  []byte(rootRelsXML),
		"word/document.xml":            []byte(documentXML),
		"word/_rels/document.xml.rels": []byte(documentRelsXML),
		"word/header1.xml":             []byte(headerXML),
		"word/_rels/header1.xml.rels":  []byte(headerRelsXML),
		"word/media/image1.png":        encodeTestPNG(t),
	})
}

// buildZip writes the given entries into an in-memory zip, in sorted name
// order so the package bytes are reproducible across runs.
func buildZip(t *testing.T, files map[string][]byte) []byte {
	t.Helper()

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, name := range names {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip.Create(%s): %v", name, err)
		}
		if _, err := w.Write(files[name]); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip.Close: %v", err)
	}
	return buf.Bytes()
}

// docxWithTwoHeaders builds a package with a default and a first-page header,
// each holding one run, through docxgo's own writer -- the shape a caller
// gets from OpenDocument on a real two-header document.
func docxWithTwoHeaders(t *testing.T) []byte {
	t.Helper()

	doc := NewDocument()
	bodyPara, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	bodyRun, err := bodyPara.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := bodyRun.SetText("body"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	section, err := doc.DefaultSection()
	if err != nil {
		t.Fatalf("DefaultSection: %v", err)
	}
	for _, hdr := range []struct {
		kind domain.HeaderType
		text string
	}{
		{domain.HeaderDefault, "default header"},
		{domain.HeaderFirst, "first-page header"},
	} {
		header, err := section.Header(hdr.kind)
		if err != nil {
			t.Fatalf("Header(%v): %v", hdr.kind, err)
		}
		para, err := header.AddParagraph()
		if err != nil {
			t.Fatalf("AddParagraph: %v", err)
		}
		run, err := para.AddRun()
		if err != nil {
			t.Fatalf("AddRun: %v", err)
		}
		if err := run.SetText(hdr.text); err != nil {
			t.Fatalf("SetText: %v", err)
		}
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	return buf.Bytes()
}

// docxWithHeaderTarget hand-authors a minimal package whose header
// relationship uses the given Target verbatim, so a test can exercise the
// forms a real producer may write ("header1.xml", "word/header1.xml",
// "/word/header1.xml") rather than only the one docxgo emits.
func docxWithHeaderTarget(t *testing.T, target string) []byte {
	t.Helper()
	return docxWithHeaderAt(t, target, "word/header1.xml")
}

// docxWithHeaderAt is docxWithHeaderTarget with the header part's archive path
// spelled out separately, for the cases where the target does not resolve to
// word/header1.xml.
func docxWithHeaderAt(t *testing.T, target, partPath string) []byte {
	t.Helper()

	parts := map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
<Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>
</Types>`,
		"_rels/.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`,
		"word/document.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<w:body>
<w:p><w:r><w:t>body</w:t></w:r></w:p>
<w:sectPr><w:headerReference w:type="default" r:id="rId5"/><w:pgSz w:w="12240" w:h="15840"/></w:sectPr>
</w:body>
</w:document>`,
		"word/_rels/document.xml.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId5" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="` + target + `"/>
</Relationships>`,
		partPath: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:hdr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<w:p><w:r><w:t>original</w:t></w:r></w:p>
</w:hdr>`,
	}

	// Deliberately no .rels for the header. A subdirectory header that owns
	// relationships hits a separate, deeper problem: the reader claims parts
	// by literal prefix, and "word/headers/_rels/header1.xml.rels" matches
	// the header *content* prefix "word/header", so it is preserved as though
	// it were a header rather than as that header's relationships. Tightening
	// that taxonomy is a reader change well outside this branch; see the
	// known limitation in the CHANGELOG.

	// The Override has to name where the part actually is, not where a
	// base-name reading of the target would put it.
	parts["[Content_Types].xml"] = strings.Replace(
		parts["[Content_Types].xml"],
		`PartName="/word/header1.xml"`,
		`PartName="/`+partPath+`"`,
		1,
	)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range parts {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip.Create(%s): %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip.Close: %v", err)
	}
	return buf.Bytes()
}

// zipPart returns the named entry from a .docx archive.
func zipPart(t *testing.T, docxBytes []byte, name string) []byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", name, err)
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		return data
	}
	t.Fatalf("%s not found in archive", name)
	return nil
}

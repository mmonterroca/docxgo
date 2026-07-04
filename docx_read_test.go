/*
MIT License

Copyright (c) 2025 Misael Monterroca <misael@monterroca.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package docx

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"path/filepath"
	"testing"

	"github.com/mmonterroca/docxgo/v2/domain"
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

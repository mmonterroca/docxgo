// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package core

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/mmonterroca/docxgo/v2/domain"
)

func TestDocument_WriteTo(t *testing.T) {
	doc := NewDocument()

	// Add paragraph with text
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph failed: %v", err)
	}

	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun failed: %v", err)
	}

	run.SetText("Hello, World!")

	// Write to buffer
	var buf bytes.Buffer
	bytesWritten, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	// Verify something was written
	if buf.Len() == 0 {
		t.Error("No bytes written")
	}

	t.Logf("Written %d bytes (reported: %d)", buf.Len(), bytesWritten)

	// Verify it's a valid ZIP
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Not a valid ZIP: %v", err)
	}

	// Verify required files exist
	requiredFiles := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"word/document.xml",
		"word/_rels/document.xml.rels",
		"docProps/core.xml",
		"docProps/app.xml",
		"word/styles.xml",
		"word/fontTable.xml",
		"word/theme/theme1.xml",
		"word/settings.xml",
		"word/webSettings.xml",
	}

	fileMap := make(map[string]bool)
	for _, f := range zipReader.File {
		fileMap[f.Name] = true
	}

	for _, required := range requiredFiles {
		if !fileMap[required] {
			t.Errorf("Required file missing: %s", required)
		}
	}

	relsFile, err := zipReader.Open("word/_rels/document.xml.rels")
	if err != nil {
		t.Fatalf("Failed to open relationships file: %v", err)
	}
	defer relsFile.Close()

	relsContent, err := io.ReadAll(relsFile)
	if err != nil {
		t.Fatalf("Failed to read relationships file: %v", err)
	}

	relsXML := string(relsContent)
	for _, target := range []string{
		"Target=\"styles.xml\"",
		"Target=\"fontTable.xml\"",
		"Target=\"theme/theme1.xml\"",
		"Target=\"settings.xml\"",
		"Target=\"webSettings.xml\"",
	} {
		if !strings.Contains(relsXML, target) {
			t.Errorf("Relationship missing for %s", target)
		}
	}

	docFile, err := zipReader.Open("word/document.xml")
	if err != nil {
		t.Fatalf("Failed to open document.xml: %v", err)
	}
	defer docFile.Close()

	docContent, err := io.ReadAll(docFile)
	if err != nil {
		t.Fatalf("Failed to read document.xml: %v", err)
	}

	if !strings.Contains(string(docContent), "w:sectPr") {
		t.Error("Section properties not serialized")
	}
}

func TestDocument_HeaderFooterSerialization(t *testing.T) {
	doc := NewDocument()
	section, err := doc.DefaultSection()
	if err != nil {
		t.Fatalf("DefaultSection failed: %v", err)
	}

	header, err := section.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header failed: %v", err)
	}
	headPara, err := header.AddParagraph()
	if err != nil {
		t.Fatalf("Header.AddParagraph failed: %v", err)
	}
	headRun, err := headPara.AddRun()
	if err != nil {
		t.Fatalf("Header paragraph AddRun failed: %v", err)
	}
	headRun.SetText("Header Text")

	footer, err := section.Footer(domain.FooterDefault)
	if err != nil {
		t.Fatalf("Footer failed: %v", err)
	}
	footPara, err := footer.AddParagraph()
	if err != nil {
		t.Fatalf("Footer.AddParagraph failed: %v", err)
	}
	footRun, err := footPara.AddRun()
	if err != nil {
		t.Fatalf("Footer paragraph AddRun failed: %v", err)
	}
	footRun.SetText("Footer Text")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Not a valid ZIP: %v", err)
	}

	files := make(map[string]*zip.File)
	for _, f := range zipReader.File {
		files[f.Name] = f
	}

	if _, ok := files["word/header1.xml"]; !ok {
		t.Fatal("header1.xml not found in DOCX package")
	}
	if _, ok := files["word/footer1.xml"]; !ok {
		t.Fatal("footer1.xml not found in DOCX package")
	}

	docFile, err := files["word/document.xml"].Open()
	if err != nil {
		t.Fatalf("Failed to open document.xml: %v", err)
	}
	defer docFile.Close()

	docContent, err := io.ReadAll(docFile)
	if err != nil {
		t.Fatalf("Failed to read document.xml: %v", err)
	}

	docXML := string(docContent)
	if !strings.Contains(docXML, "w:headerReference") {
		t.Error("Document missing headerReference")
	}
	if !strings.Contains(docXML, "w:footerReference") {
		t.Error("Document missing footerReference")
	}

	relsFile, err := files["word/_rels/document.xml.rels"].Open()
	if err != nil {
		t.Fatalf("Failed to open document relations: %v", err)
	}
	defer relsFile.Close()

	relsContent, err := io.ReadAll(relsFile)
	if err != nil {
		t.Fatalf("Failed to read document relations: %v", err)
	}
	relsXML := string(relsContent)
	if !strings.Contains(relsXML, "header1.xml") {
		t.Error("Relationship for header1.xml missing")
	}
	if !strings.Contains(relsXML, "footer1.xml") {
		t.Error("Relationship for footer1.xml missing")
	}
}

// zipPartText reads a named part's content as a string from a resaved DOCX
// package, failing the test if the part is missing or unreadable.
func zipPartText(t *testing.T, docBytes []byte, name string) string {
	t.Helper()

	zipReader, err := zip.NewReader(bytes.NewReader(docBytes), int64(len(docBytes)))
	if err != nil {
		t.Fatalf("not a valid ZIP: %v", err)
	}
	for _, f := range zipReader.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("failed to open %s: %v", name, err)
		}
		defer rc.Close()
		content, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("failed to read %s: %v", name, err)
		}
		return string(content)
	}
	t.Fatalf("%s not found in DOCX package", name)
	return ""
}

// TestDocument_HeaderFooterTableSerialization pins the actual serialized
// XML for a header/footer containing an interleaved paragraph/table/
// paragraph sequence — unlike TestDocument_HeaderFooterSerialization above,
// which only asserts the part exists. Every failure mode in the storage
// (docxHeader/docxFooter.AddTable/Blocks) and serializer
// (serializeHeaderFooterContent) layers is invisible to that test, since it
// never adds a table or inspects the header/footer XML content at all.
func TestDocument_HeaderFooterTableSerialization(t *testing.T) {
	doc := NewDocument()
	section, err := doc.DefaultSection()
	if err != nil {
		t.Fatalf("DefaultSection failed: %v", err)
	}

	header, err := section.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header failed: %v", err)
	}
	addTextParagraph(t, header, "Before Table")
	headerTable, err := header.AddTable(1, 2)
	if err != nil {
		t.Fatalf("Header.AddTable failed: %v", err)
	}
	setCellText(t, headerTable, 0, 0, "H-Cell")
	addTextParagraph(t, header, "After Table")

	footer, err := section.Footer(domain.FooterDefault)
	if err != nil {
		t.Fatalf("Footer failed: %v", err)
	}
	footerTable, err := footer.AddTable(1, 1)
	if err != nil {
		t.Fatalf("Footer.AddTable failed: %v", err)
	}
	setCellText(t, footerTable, 0, 0, "F-Cell")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	headerXML := zipPartText(t, buf.Bytes(), "word/header1.xml")
	if !strings.Contains(headerXML, "<w:tbl") {
		t.Error("header1.xml missing <w:tbl>")
	}
	if !strings.Contains(headerXML, "H-Cell") {
		t.Error("header1.xml missing cell text \"H-Cell\"")
	}
	beforeIdx := strings.Index(headerXML, "Before Table")
	tblIdx := strings.Index(headerXML, "<w:tbl")
	afterIdx := strings.Index(headerXML, "After Table")
	if beforeIdx < 0 || tblIdx < 0 || afterIdx < 0 {
		t.Fatalf("header1.xml missing expected content: %s", headerXML)
	}
	if !(beforeIdx < tblIdx && tblIdx < afterIdx) {
		t.Errorf("header1.xml out of order: want \"Before Table\" < <w:tbl> < \"After Table\", got offsets %d, %d, %d", beforeIdx, tblIdx, afterIdx)
	}
	// A trailing <w:p> must follow the table so Word doesn't coalesce it with
	// whatever comes next; "After Table"'s own paragraph satisfies that here.

	footerXML := zipPartText(t, buf.Bytes(), "word/footer1.xml")
	if !strings.Contains(footerXML, "<w:tbl") {
		t.Error("footer1.xml missing <w:tbl>")
	}
	if !strings.Contains(footerXML, "F-Cell") {
		t.Error("footer1.xml missing cell text \"F-Cell\"")
	}
	// The table is the last block in the footer, so serializeHeaderFooterContent
	// must have appended a trailing empty paragraph after it.
	tblCloseIdx := strings.LastIndex(footerXML, "</w:tbl>")
	trailingPIdx := strings.Index(footerXML[tblCloseIdx:], "<w:p")
	if tblCloseIdx < 0 || trailingPIdx < 0 {
		t.Errorf("footer1.xml missing trailing <w:p> after </w:tbl>: %s", footerXML)
	}
}

// addTextParagraph adds a paragraph with a single run of text to a
// domain.Header or domain.Footer.
func addTextParagraph(t *testing.T, target interface {
	AddParagraph() (domain.Paragraph, error)
}, text string) {
	t.Helper()
	para, err := target.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph failed: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun failed: %v", err)
	}
	if err := run.SetText(text); err != nil {
		t.Fatalf("SetText failed: %v", err)
	}
}

// setCellText adds a paragraph with a single run of text to the table cell
// at [row, col].
func setCellText(t *testing.T, table domain.Table, row, col int, text string) {
	t.Helper()
	r, err := table.Row(row)
	if err != nil {
		t.Fatalf("Table.Row(%d) failed: %v", row, err)
	}
	c, err := r.Cell(col)
	if err != nil {
		t.Fatalf("Row.Cell(%d) failed: %v", col, err)
	}
	addTextParagraph(t, c, text)
}

func TestDocument_SaveAs(t *testing.T) {
	doc := NewDocument()

	para, _ := doc.AddParagraph()
	run, _ := para.AddRun()
	run.SetText("Test document")
	run.SetBold(true)

	// Save to temp file
	tmpFile := "/tmp/test_document.docx"
	defer os.Remove(tmpFile)

	err := doc.SaveAs(tmpFile)
	if err != nil {
		t.Fatalf("SaveAs failed: %v", err)
	}

	// Verify file exists
	stat, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("File not created: %v", err)
	}

	if stat.Size() == 0 {
		t.Error("File is empty")
	}

	t.Logf("File created: %d bytes", stat.Size())

	// Verify it's a valid .docx (ZIP file)
	f, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("Cannot open file: %v", err)
	}
	defer f.Close()

	_, err = zip.NewReader(f, stat.Size())
	if err != nil {
		t.Fatalf("Not a valid .docx file: %v", err)
	}
}

func TestDocument_ComplexDocument(t *testing.T) {
	doc := NewDocument()

	// Add multiple paragraphs
	for i := 0; i < 3; i++ {
		para, _ := doc.AddParagraph()
		run, _ := para.AddRun()
		run.SetText("Paragraph " + string(rune('1'+i)))

		if i == 0 {
			run.SetBold(true)
			run.SetSize(28) // 14pt
		}
	}

	// Add a table
	table, err := doc.AddTable(2, 3)
	if err != nil {
		t.Fatalf("AddTable failed: %v", err)
	}

	for r := 0; r < 2; r++ {
		row, _ := table.Row(r)
		for c := 0; c < 3; c++ {
			cell, _ := row.Cell(c)
			para, _ := cell.AddParagraph()
			run, _ := para.AddRun()
			run.SetText("Cell")
		}
	}

	// Write to buffer
	var buf bytes.Buffer
	_, err = doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("No bytes written")
	}

	t.Logf("Complex document: %d bytes", buf.Len())
}

package core

/*
   Copyright (c) 2025 Misael Monterroca

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

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

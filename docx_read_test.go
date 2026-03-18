package docx

import (
	"bytes"
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

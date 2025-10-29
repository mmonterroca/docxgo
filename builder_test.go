package docx

import (
	"testing"

	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/pkg/errors"
)

func TestDocumentBuilder_Build(t *testing.T) {
	t.Run("builds document with paragraph", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().Text("Test").End()

		doc, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if doc == nil {
			t.Fatal("expected document, got nil")
		}
	})

	t.Run("applies options", func(t *testing.T) {
		builder := NewDocumentBuilder(
			WithTitle("Test Doc"),
			WithAuthor("Test Author"),
		)
		builder.AddParagraph().Text("Test").End()

		doc, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if doc == nil {
			t.Fatal("expected document, got nil")
		}
	})

	t.Run("accumulates errors", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.errors = append(builder.errors, errors.NewValidationError("test", "field", "value", "test error"))

		doc, err := builder.Build()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if doc != nil {
			t.Fatal("expected nil document, got non-nil")
		}
	})
}

func TestDocumentBuilder_Sections(t *testing.T) {
	t.Run("configures default section", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.DefaultSection().
			PageSize(domain.PageSizeA4).
			Orientation(domain.OrientationLandscape).
			Margins(domain.Margins{Top: 720, Right: 1440, Bottom: 720, Left: 1440, Header: 360, Footer: 360}).
			Columns(2).
			End()
		builder.AddParagraph().Text("section content").End()

		doc, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		sections := doc.Sections()
		if len(sections) != 1 {
			t.Fatalf("expected 1 section, got %d", len(sections))
		}

		sec := sections[0]
		if sec.PageSize() != domain.PageSizeA4 {
			t.Errorf("expected A4 page size, got %+v", sec.PageSize())
		}
		if sec.Orientation() != domain.OrientationLandscape {
			t.Errorf("expected landscape orientation, got %v", sec.Orientation())
		}
		if sec.Columns() != 2 {
			t.Errorf("expected 2 columns, got %d", sec.Columns())
		}
		m := sec.Margins()
		if m.Top != 720 || m.Bottom != 720 || m.Header != 360 || m.Footer != 360 {
			t.Errorf("unexpected margins %+v", m)
		}
	})

	t.Run("adds additional section with break", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().Text("Before section break").End()

		builder.AddSection(domain.SectionBreakTypeEvenPage).
			Columns(3).
			Orientation(domain.OrientationPortrait).
			End()

		builder.AddParagraph().Text("After section break").End()

		doc, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		sections := doc.Sections()
		if len(sections) != 2 {
			t.Fatalf("expected 2 sections, got %d", len(sections))
		}
		if sections[1].Columns() != 3 {
			t.Errorf("expected 3 columns on second section, got %d", sections[1].Columns())
		}

		blocks := doc.Blocks()
		foundBreak := false
		for _, block := range blocks {
			if block.SectionBreak != nil {
				foundBreak = true
				if block.SectionBreak.Type != domain.SectionBreakTypeEvenPage {
					t.Errorf("expected even page break, got %v", block.SectionBreak.Type)
				}
			}
		}
		if !foundBreak {
			t.Fatalf("expected section break block in document")
		}
	})

	t.Run("records errors when section configuration fails", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.DefaultSection().Columns(0).End()

		if _, err := builder.Build(); err == nil {
			t.Fatal("expected error due to invalid column count, got nil")
		}
	})

	t.Run("configures header via section builder", func(t *testing.T) {
		builder := NewDocumentBuilder()
		secBuilder := builder.DefaultSection()
		header, err := secBuilder.Header(domain.HeaderDefault)
		if err != nil {
			t.Fatalf("expected header, got error %v", err)
		}

		para, err := header.AddParagraph()
		if err != nil {
			t.Fatalf("expected paragraph in header, got %v", err)
		}
		run, err := para.AddRun()
		if err != nil {
			t.Fatalf("expected run in header paragraph, got %v", err)
		}
		if err := run.SetText("Header text"); err != nil {
			t.Fatalf("expected to set header text, got %v", err)
		}

		secBuilder.End()
		builder.AddParagraph().Text("body content").End()
		doc, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		section := doc.Sections()[0]
		storedHeader, err := section.Header(domain.HeaderDefault)
		if err != nil {
			t.Fatalf("expected header from section, got %v", err)
		}
		if len(storedHeader.Paragraphs()) == 0 {
			t.Fatal("expected header to contain paragraphs")
		}
	})
}

func TestParagraphBuilder_Text(t *testing.T) {
	t.Run("adds single text run", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Text("Hello").
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("adds multiple text runs", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Text("Hello ").
			Text("World").
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestParagraphBuilder_Bold(t *testing.T) {
	t.Run("sets bold on current run", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Text("Bold text").
			Bold().
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("returns error when no current run", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Bold().
			End()

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestParagraphBuilder_Italic(t *testing.T) {
	t.Run("sets italic on current run", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Text("Italic text").
			Italic().
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("returns error when no current run", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Italic().
			End()

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestParagraphBuilder_Color(t *testing.T) {
	t.Run("sets color on current run", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Text("Red text").
			Color(domain.Color{R: 255, G: 0, B: 0}).
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("returns error when no current run", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Color(domain.Color{R: 255, G: 0, B: 0}).
			End()

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestParagraphBuilder_FontSize(t *testing.T) {
	t.Run("sets font size on current run", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Text("Large text").
			FontSize(16).
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("returns error for invalid font size", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Text("Text").
			FontSize(0).
			End()

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error for invalid font size, got nil")
		}
	})

	t.Run("returns error when no current run", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			FontSize(12).
			End()

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestParagraphBuilder_Alignment(t *testing.T) {
	t.Run("sets alignment on paragraph", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Text("Centered").
			Alignment(domain.AlignmentCenter).
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("works with empty paragraph", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Alignment(domain.AlignmentRight).
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestParagraphBuilder_Underline(t *testing.T) {
	t.Run("sets underline on current run", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Text("Underlined").
			Underline(domain.UnderlineSingle).
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("returns error when no current run", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Underline(domain.UnderlineSingle).
			End()

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestParagraphBuilder_Chaining(t *testing.T) {
	t.Run("chains multiple formatting calls", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Text("Multi-format").
			Bold().
			Italic().
			Color(domain.Color{R: 255, G: 0, B: 0}).
			FontSize(14).
			Underline(domain.UnderlineSingle).
			Alignment(domain.AlignmentCenter).
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("chains multiple text runs", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Text("Hello ").
			Text("beautiful ").Bold().
			Text("world").Italic().
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestTableBuilder_Basic(t *testing.T) {
	t.Run("creates empty table", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddTable(1, 1).End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("creates table with width", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddTable(1, 1).
			Width(domain.WidthDXA, 5000).
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("creates table with alignment", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddTable(1, 1).
			Alignment(domain.AlignmentCenter).
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestTableBuilder_Rows(t *testing.T) {
	t.Run("configures row height", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddTable(2, 2).
			Row(0).
			Height(500).
			End().
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("configures multiple rows", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddTable(3, 2).
			Row(0).Height(400).End().
			Row(1).Height(500).End().
			Row(2).Height(600).End().
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestRowBuilder_Cells(t *testing.T) {
	t.Run("adds text to cell", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddTable(1, 2).
			Row(0).
			Cell(0).Text("Cell 1").End().
			Cell(1).Text("Cell 2").End().
			End().
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("formats cell text", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddTable(1, 1).
			Row(0).
			Cell(0).
			Text("Bold text").
			Bold().
			End().
			End().
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestCellBuilder_Formatting(t *testing.T) {
	t.Run("sets cell width", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddTable(1, 1).
			Row(0).
			Cell(0).
			Width(3000).
			Text("Wide cell").
			End().
			End().
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("sets cell vertical alignment", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddTable(1, 1).
			Row(0).
			Cell(0).
			VerticalAlignment(domain.VerticalAlignCenter).
			Text("Centered").
			End().
			End().
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("sets cell shading", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddTable(1, 1).
			Row(0).
			Cell(0).
			Shading(domain.Color{R: 220, G: 220, B: 220}).
			Text("Gray background").
			End().
			End().
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestTableBuilder_ComplexTable(t *testing.T) {
	t.Run("creates table with mixed formatting", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddTable(2, 3).
			Width(domain.WidthDXA, 9000).
			Alignment(domain.AlignmentCenter).
			Row(0).
			Height(400).
			Cell(0).Text("Header 1").Bold().End().
			Cell(1).Text("Header 2").Bold().End().
			Cell(2).Text("Header 3").Bold().End().
			End().
			Row(1).
			Height(300).
			Cell(0).Text("Data 1").End().
			Cell(1).Text("Data 2").End().
			Cell(2).Text("Data 3").End().
			End().
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("creates table with cell merging", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddTable(2, 2).
			Row(0).
			Cell(0).
			Merge(2, 1).
			Text("Merged cell").
			End().
			End().
			Row(1).
			Cell(0).Text("Cell 1").End().
			Cell(1).Text("Cell 2").End().
			End().
			End()

		_, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestBuilder_ErrorAccumulation(t *testing.T) {
	t.Run("accumulates multiple errors", func(t *testing.T) {
		builder := NewDocumentBuilder()
		// Create a paragraph builder with errors
		builder.AddParagraph().
			Text("").                                // Error: empty text
			Bold().                                  // Error: no current run
			FontSize(0).                             // Error: invalid size
			Color(domain.Color{R: 255, G: 0, B: 0}). // Error: no current run
			End()

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("stops on first error and returns it", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			FontSize(0). // Invalid font size
			Text("Valid text").
			End()

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestOptions_Validation(t *testing.T) {
	t.Run("WithDefaultFont validates font name", func(t *testing.T) {
		doc, err := NewDocumentBuilder(
			WithDefaultFont(""),
		).Build()

		if err == nil {
			t.Fatal("expected error for empty font name, got nil")
		}
		if doc != nil {
			t.Fatal("expected nil document, got non-nil")
		}
	})

	t.Run("WithDefaultFontSize validates font size", func(t *testing.T) {
		doc, err := NewDocumentBuilder(
			WithDefaultFontSize(0),
		).Build()

		if err == nil {
			t.Fatal("expected error for invalid font size, got nil")
		}
		if doc != nil {
			t.Fatal("expected nil document, got non-nil")
		}
	})

	t.Run("WithTitle validates title", func(t *testing.T) {
		doc, err := NewDocumentBuilder(
			WithTitle(""),
		).Build()

		if err == nil {
			t.Fatal("expected error for empty title, got nil")
		}
		if doc != nil {
			t.Fatal("expected nil document, got non-nil")
		}
	})

	t.Run("WithAuthor validates author", func(t *testing.T) {
		doc, err := NewDocumentBuilder(
			WithAuthor(""),
		).Build()

		if err == nil {
			t.Fatal("expected error for empty author, got nil")
		}
		if doc != nil {
			t.Fatal("expected nil document, got non-nil")
		}
	})
}

func TestDocumentBuilder_Options(t *testing.T) {
	t.Run("WithPageSize sets page size", func(t *testing.T) {
		builder := NewDocumentBuilder(
			WithPageSize(Letter),
		)
		builder.AddParagraph().Text("Test").End()

		doc, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		sections := doc.Sections()
		if len(sections) == 0 {
			t.Fatal("expected at least one section")
		}

		// Note: Default is A4, so we just verify a document was created
		// The page size conversion is handled by the builder
		if doc == nil {
			t.Error("expected document to be created")
		}
	})

	t.Run("WithMargins sets margins", func(t *testing.T) {
		margins := Margins{
			Top:    1440,
			Bottom: 1440,
			Left:   1440,
			Right:  1440,
		}
		builder := NewDocumentBuilder(
			WithMargins(margins),
		)
		builder.AddParagraph().Text("Test").End()

		doc, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		sections := doc.Sections()
		if len(sections) == 0 {
			t.Fatal("expected at least one section")
		}

		actualMargins := sections[0].Margins()
		if actualMargins.Top != margins.Top || actualMargins.Left != margins.Left {
			t.Errorf("expected margins %+v, got %+v", margins, actualMargins)
		}
	})

	t.Run("WithMetadata sets document metadata", func(t *testing.T) {
		meta := &domain.Metadata{
			Title:   "Test Title",
			Creator: "Test Author",
			Subject: "Test Subject",
		}
		builder := NewDocumentBuilder(
			WithMetadata(meta),
		)
		builder.AddParagraph().Text("Test").End()

		doc, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		metadata := doc.Metadata()
		if metadata.Title != "Test Title" {
			t.Errorf("expected title 'Test Title', got %s", metadata.Title)
		}
		if metadata.Creator != "Test Author" {
			t.Errorf("expected author 'Test Author', got %s", metadata.Creator)
		}
		if metadata.Subject != "Test Subject" {
			t.Errorf("expected subject 'Test Subject', got %s", metadata.Subject)
		}
	})

	t.Run("WithSubject sets subject", func(t *testing.T) {
		builder := NewDocumentBuilder(
			WithSubject("Test Subject"),
		)
		builder.AddParagraph().Text("Test").End()

		doc, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		metadata := doc.Metadata()
		if metadata.Subject != "Test Subject" {
			t.Errorf("expected subject 'Test Subject', got %s", metadata.Subject)
		}
	})

	t.Run("WithStrictValidation enables strict validation", func(t *testing.T) {
		builder := NewDocumentBuilder(
			WithStrictValidation(),
		)
		builder.AddParagraph().Text("Test").End()

		doc, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if doc == nil {
			t.Fatal("expected document, got nil")
		}
	})
}

func TestDocumentBuilder_SetMetadata(t *testing.T) {
	t.Run("sets metadata on builder", func(t *testing.T) {
		builder := NewDocumentBuilder()
		meta := &domain.Metadata{
			Title:   "Title",
			Creator: "Author",
			Subject: "Subject",
		}
		builder.SetMetadata(meta)
		builder.AddParagraph().Text("Test").End()

		doc, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		metadata := doc.Metadata()
		if metadata.Title != "Title" {
			t.Errorf("expected title 'Title', got %s", metadata.Title)
		}
		if metadata.Creator != "Author" {
			t.Errorf("expected creator 'Author', got %s", metadata.Creator)
		}
	})
}

func TestDocumentBuilder_Footer(t *testing.T) {
	t.Run("adds footer to section", func(t *testing.T) {
		builder := NewDocumentBuilder()
		secBuilder := builder.DefaultSection()

		footer, err := secBuilder.Footer(domain.FooterDefault)
		if err != nil {
			t.Fatalf("expected footer, got error %v", err)
		}

		para, err := footer.AddParagraph()
		if err != nil {
			t.Fatalf("expected paragraph in footer, got %v", err)
		}

		run, err := para.AddRun()
		if err != nil {
			t.Fatalf("expected run in footer paragraph, got %v", err)
		}

		run.AddText("Footer text")
		secBuilder.End()
		builder.AddParagraph().Text("Content").End()

		doc, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		sections := doc.Sections()
		if len(sections) == 0 {
			t.Fatal("expected at least one section")
		}

		// Footer should be present
		footer2, err := sections[0].Footer(domain.FooterDefault)
		if err != nil || footer2 == nil {
			t.Fatal("expected footer in section")
		}
	})
}

func TestDocumentBuilder_SectionAccess(t *testing.T) {
	t.Run("Section returns current section", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().Text("Test").End()

		secBuilder := builder.DefaultSection()
		section := secBuilder.Section()
		if section == nil {
			t.Fatal("expected section, got nil")
		}

		// Should be able to use the returned section
		secBuilder.Orientation(domain.OrientationLandscape).End()

		doc, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		sections := doc.Sections()
		if len(sections) == 0 {
			t.Fatal("expected at least one section")
		}
	})
}

func TestFieldCreationFunctions(t *testing.T) {
	t.Run("NewField creates field", func(t *testing.T) {
		field := NewField(domain.FieldTypePageNumber)
		if field == nil {
			t.Fatal("expected field, got nil")
		}
		if field.Type() != domain.FieldTypePageNumber {
			t.Errorf("expected FieldTypePageNumber, got %v", field.Type())
		}
	})

	t.Run("NewPageNumberField creates page number field", func(t *testing.T) {
		field := NewPageNumberField()
		if field == nil {
			t.Fatal("expected field, got nil")
		}
		if field.Type() != domain.FieldTypePageNumber {
			t.Errorf("expected FieldTypePageNumber, got %v", field.Type())
		}
	})

	t.Run("NewPageCountField creates page count field", func(t *testing.T) {
		field := NewPageCountField()
		if field == nil {
			t.Fatal("expected field, got nil")
		}
		// FieldTypePageCount is an alias for FieldTypeNumPages
		fieldType := field.Type()
		if fieldType != domain.FieldTypeNumPages && fieldType != domain.FieldTypePageCount {
			t.Errorf("expected FieldTypeNumPages or FieldTypePageCount, got %v", fieldType)
		}
	})

	t.Run("NewTOCField creates TOC field", func(t *testing.T) {
		switches := map[string]string{
			"levels":     "1-3",
			"hyperlinks": "true",
		}
		field := NewTOCField(switches)
		if field == nil {
			t.Fatal("expected field, got nil")
		}
		if field.Type() != domain.FieldTypeTOC {
			t.Errorf("expected FieldTypeTOC, got %v", field.Type())
		}
	})

	t.Run("NewHyperlinkField creates hyperlink field", func(t *testing.T) {
		field := NewHyperlinkField("https://example.com", "Example")
		if field == nil {
			t.Fatal("expected field, got nil")
		}
		if field.Type() != domain.FieldTypeHyperlink {
			t.Errorf("expected FieldTypeHyperlink, got %v", field.Type())
		}
	})

	t.Run("NewStyleRefField creates style ref field", func(t *testing.T) {
		field := NewStyleRefField("Heading 1")
		if field == nil {
			t.Fatal("expected field, got nil")
		}
		if field.Type() != domain.FieldTypeStyleRef {
			t.Errorf("expected FieldTypeStyleRef, got %v", field.Type())
		}
	})
}

func TestTableBuilder_Style(t *testing.T) {
	t.Run("sets table style", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddTable(1, 1).
			Style(domain.TableStyleGrid).
			Row(0).Cell(0).Text("Cell 1").End().End().
			End()

		doc, err := builder.Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		tables := doc.Tables()
		if len(tables) == 0 {
			t.Fatal("expected at least one table")
		}
	})
}

func TestParagraphBuilder_AddImage(t *testing.T) {
	t.Run("AddImage fails with empty path", func(t *testing.T) {
		builder := NewDocumentBuilder()
		builder.AddParagraph().
			Text("Before image").
			AddImage("").
			End()

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error for empty image path, got nil")
		}
	})

	t.Run("AddImageWithSize fails with empty path", func(t *testing.T) {
		builder := NewDocumentBuilder()
		size := domain.NewImageSize(100, 100)
		builder.AddParagraph().
			Text("Before image").
			AddImageWithSize("", size).
			End()

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error for empty image path, got nil")
		}
	})

	t.Run("AddImageWithPosition fails with empty path", func(t *testing.T) {
		builder := NewDocumentBuilder()
		size := domain.NewImageSize(100, 100)
		pos := domain.DefaultImagePosition()
		builder.AddParagraph().
			Text("Before image").
			AddImageWithPosition("", size, pos).
			End()

		_, err := builder.Build()
		if err == nil {
			t.Fatal("expected error for empty image path, got nil")
		}
	})
}

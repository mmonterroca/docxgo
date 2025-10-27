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

package docx

import (
	"testing"

	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/pkg/errors"
)

func TestDocumentBuilder_Build(t *testing.T) {
	t.Run("builds empty document", func(t *testing.T) {
		doc, err := NewDocumentBuilder().Build()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if doc == nil {
			t.Fatal("expected document, got nil")
		}
	})

	t.Run("applies options", func(t *testing.T) {
		doc, err := NewDocumentBuilder(
			WithTitle("Test Doc"),
			WithAuthor("Test Author"),
		).Build()
		
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
	doc, _ := NewDocumentBuilder().Build()
	
	t.Run("adds single text run", func(t *testing.T) {
		err := doc.AddParagraph().
			Text("Hello").
			End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("adds multiple text runs", func(t *testing.T) {
		err := doc.AddParagraph().
			Text("Hello ").
			Text("World").
			End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("returns error for empty text", func(t *testing.T) {
		err := doc.AddParagraph().
			Text("").
			End()
		
		if err == nil {
			t.Fatal("expected error for empty text, got nil")
		}
	})
}

func TestParagraphBuilder_Bold(t *testing.T) {
	doc, _ := NewDocumentBuilder().Build()
	
	t.Run("sets bold on current run", func(t *testing.T) {
		err := doc.AddParagraph().
			Text("Bold text").
			Bold().
			End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("returns error when no current run", func(t *testing.T) {
		err := doc.AddParagraph().
			Bold().
			End()
		
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestParagraphBuilder_Italic(t *testing.T) {
	doc, _ := NewDocumentBuilder().Build()
	
	t.Run("sets italic on current run", func(t *testing.T) {
		err := doc.AddParagraph().
			Text("Italic text").
			Italic().
			End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("returns error when no current run", func(t *testing.T) {
		err := doc.AddParagraph().
			Italic().
			End()
		
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestParagraphBuilder_Color(t *testing.T) {
	doc, _ := NewDocumentBuilder().Build()
	
	t.Run("sets color on current run", func(t *testing.T) {
		err := doc.AddParagraph().
			Text("Red text").
			Color(0xFF0000).
			End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("returns error when no current run", func(t *testing.T) {
		err := doc.AddParagraph().
			Color(0xFF0000).
			End()
		
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestParagraphBuilder_FontSize(t *testing.T) {
	doc, _ := NewDocumentBuilder().Build()
	
	t.Run("sets font size on current run", func(t *testing.T) {
		err := doc.AddParagraph().
			Text("Large text").
			FontSize(16).
			End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("returns error for invalid font size", func(t *testing.T) {
		err := doc.AddParagraph().
			Text("Text").
			FontSize(0).
			End()
		
		if err == nil {
			t.Fatal("expected error for invalid font size, got nil")
		}
	})

	t.Run("returns error when no current run", func(t *testing.T) {
		err := doc.AddParagraph().
			FontSize(12).
			End()
		
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestParagraphBuilder_Alignment(t *testing.T) {
	doc, _ := NewDocumentBuilder().Build()
	
	t.Run("sets alignment on paragraph", func(t *testing.T) {
		err := doc.AddParagraph().
			Text("Centered").
			Alignment(domain.AlignmentCenter).
			End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("works with empty paragraph", func(t *testing.T) {
		err := doc.AddParagraph().
			Alignment(domain.AlignmentRight).
			End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestParagraphBuilder_Underline(t *testing.T) {
	doc, _ := NewDocumentBuilder().Build()
	
	t.Run("sets underline on current run", func(t *testing.T) {
		err := doc.AddParagraph().
			Text("Underlined").
			Underline(domain.UnderlineSingle).
			End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("returns error when no current run", func(t *testing.T) {
		err := doc.AddParagraph().
			Underline(domain.UnderlineSingle).
			End()
		
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestParagraphBuilder_Chaining(t *testing.T) {
	doc, _ := NewDocumentBuilder().Build()
	
	t.Run("chains multiple formatting calls", func(t *testing.T) {
		err := doc.AddParagraph().
			Text("Multi-format").
			Bold().
			Italic().
			Color(0xFF0000).
			FontSize(14).
			Underline(domain.UnderlineSingle).
			Alignment(domain.AlignmentCenter).
			End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("chains multiple text runs", func(t *testing.T) {
		err := doc.AddParagraph().
			Text("Hello ").
			Text("beautiful ").Bold().
			Text("world").Italic().
			End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestTableBuilder_Basic(t *testing.T) {
	doc, _ := NewDocumentBuilder().Build()
	
	t.Run("creates empty table", func(t *testing.T) {
		err := doc.AddTable().End()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("creates table with width", func(t *testing.T) {
		err := doc.AddTable().
			Width(5000).
			End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("creates table with alignment", func(t *testing.T) {
		err := doc.AddTable().
			Alignment(domain.AlignmentCenter).
			End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestTableBuilder_Rows(t *testing.T) {
	doc, _ := NewDocumentBuilder().Build()
	
	t.Run("adds single row", func(t *testing.T) {
		err := doc.AddTable().
			Row().
				Cell().Text("Cell 1").End().
			End().
		End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("adds multiple rows", func(t *testing.T) {
		err := doc.AddTable().
			Row().
				Cell().Text("R1C1").End().
				Cell().Text("R1C2").End().
			End().
			Row().
				Cell().Text("R2C1").End().
				Cell().Text("R2C2").End().
			End().
		End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestRowBuilder_Cells(t *testing.T) {
	doc, _ := NewDocumentBuilder().Build()
	
	t.Run("adds single cell", func(t *testing.T) {
		err := doc.AddTable().
			Row().
				Cell().Text("Data").End().
			End().
		End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("adds multiple cells", func(t *testing.T) {
		err := doc.AddTable().
			Row().
				Cell().Text("C1").End().
				Cell().Text("C2").End().
				Cell().Text("C3").End().
			End().
		End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("sets row height", func(t *testing.T) {
		err := doc.AddTable().
			Row().
				Height(500).
				Cell().Text("Tall row").End().
			End().
		End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestCellBuilder_Formatting(t *testing.T) {
	doc, _ := NewDocumentBuilder().Build()
	
	t.Run("sets text with formatting", func(t *testing.T) {
		err := doc.AddTable().
			Row().
				Cell().
					Text("Bold text").
					Bold().
				End().
			End().
		End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("sets cell width", func(t *testing.T) {
		err := doc.AddTable().
			Row().
				Cell().
					Text("Wide cell").
					Width(3000).
				End().
			End().
		End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("sets vertical alignment", func(t *testing.T) {
		err := doc.AddTable().
			Row().
				Cell().
					Text("Centered vertically").
					VerticalAlignment(domain.VerticalAlignmentCenter).
				End().
			End().
		End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("sets shading", func(t *testing.T) {
		err := doc.AddTable().
			Row().
				Cell().
					Text("Colored cell").
					Shading(0xE0E0E0).
				End().
			End().
		End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestTableBuilder_ComplexTable(t *testing.T) {
	doc, _ := NewDocumentBuilder().Build()
	
	t.Run("creates complex formatted table", func(t *testing.T) {
		err := doc.AddTable().
			Width(5000).
			Alignment(domain.AlignmentCenter).
			Row().
				Cell().Text("Header 1").Bold().Width(1500).End().
				Cell().Text("Header 2").Bold().Width(1500).End().
				Cell().Text("Header 3").Bold().Width(2000).End().
			End().
			Row().
				Height(400).
				Cell().Text("Data 1").End().
				Cell().Text("Data 2").Color(0xFF0000).End().
				Cell().Text("Data 3").Italic().End().
			End().
			Row().
				Cell().
					Text("Footer").
					Bold().
					Shading(0xE0E0E0).
					VerticalAlignment(domain.VerticalAlignmentCenter).
				End().
			End().
		End()
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestBuilder_ErrorAccumulation(t *testing.T) {
	doc, _ := NewDocumentBuilder().Build()
	
	t.Run("accumulates multiple errors", func(t *testing.T) {
		// Create a paragraph builder with errors
		err := doc.AddParagraph().
			Text("").           // Error: empty text
			Bold().             // Error: no current run
			FontSize(0).        // Error: invalid size
			Color(0xFF0000).    // Error: no current run
			End()
		
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("stops on first error and returns it", func(t *testing.T) {
		err := doc.AddParagraph().
			Text("").  // This should cause an error
			Text("Valid text").  // This shouldn't execute
			End()
		
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestOptions_Validation(t *testing.T) {
	t.Run("WithDefaultFont validates font name", func(t *testing.T) {
		doc, err := NewDocumentBuilder(
			WithDefaultFont("", 11),
		).Build()
		
		if err == nil {
			t.Fatal("expected error for empty font name, got nil")
		}
		if doc != nil {
			t.Fatal("expected nil document, got non-nil")
		}
	})

	t.Run("WithDefaultFont validates font size", func(t *testing.T) {
		doc, err := NewDocumentBuilder(
			WithDefaultFont("Arial", 0),
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

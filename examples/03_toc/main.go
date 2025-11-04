package main

import (
	"fmt"
	"log"
	"strings"

	docx "github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/domain"
)

func main() {
	doc := docx.NewDocument()

	if err := buildDocument(doc); err != nil {
		log.Fatalf("failed to build document: %v", err)
	}

	if err := doc.SaveAs("03_toc_demo.docx"); err != nil {
		log.Fatalf("failed to save document: %v", err)
	}

	fmt.Println("Table of contents demo created: 03_toc_demo.docx")
	fmt.Println("Open the file in Word and press F9 to refresh the TOC.")
}

func buildDocument(doc domain.Document) error {
	meta := &domain.Metadata{
		Title:   "go-docx Table of Contents Demo",
		Subject: "Automatic TOC generated from heading styles",
		Creator: "go-docx",
	}
	if err := doc.SetMetadata(meta); err != nil {
		return fmt.Errorf("set metadata: %w", err)
	}

	if err := addCover(doc); err != nil {
		return err
	}

	if err := addTableOfContents(doc); err != nil {
		return err
	}

	chapters := []struct {
		number   int
		title    string
		sections []string
	}{
		{
			number: 1,
			title:  "Getting Started",
			sections: []string{
				"Installing go-docx",
				"Creating your first document",
				"Applying heading styles",
			},
		},
		{
			number: 2,
			title:  "Building Content",
			sections: []string{
				"Paragraph formatting",
				"Working with lists",
				"Inserting images",
			},
		},
	}

	for _, chapter := range chapters {
		if err := addChapter(doc, chapter.number, chapter.title, chapter.sections); err != nil {
			return err
		}
	}

	if err := addAppendix(doc); err != nil {
		return err
	}

	return nil
}

func addCover(doc domain.Document) error {
	title, err := doc.AddParagraph()
	if err != nil {
		return fmt.Errorf("add title paragraph: %w", err)
	}
	if err := title.SetStyle(domain.StyleIDTitle); err != nil {
		return fmt.Errorf("set title style: %w", err)
	}
	if err := title.SetAlignment(domain.AlignmentCenter); err != nil {
		return fmt.Errorf("set title alignment: %w", err)
	}
	titleRun, err := title.AddRun()
	if err != nil {
		return fmt.Errorf("add title run: %w", err)
	}
	if err := titleRun.SetText("go-docx v2 Table of Contents Demo"); err != nil {
		return fmt.Errorf("set title text: %w", err)
	}

	subtitle, err := doc.AddParagraph()
	if err != nil {
		return fmt.Errorf("add subtitle paragraph: %w", err)
	}
	if err := subtitle.SetStyle(domain.StyleIDSubtitle); err != nil {
		return fmt.Errorf("set subtitle style: %w", err)
	}
	if err := subtitle.SetAlignment(domain.AlignmentCenter); err != nil {
		return fmt.Errorf("set subtitle alignment: %w", err)
	}
	subtitleRun, err := subtitle.AddRun()
	if err != nil {
		return fmt.Errorf("add subtitle run: %w", err)
	}
	if err := subtitleRun.SetText("The TOC is driven by Heading 1 and Heading 2 paragraphs"); err != nil {
		return fmt.Errorf("set subtitle text: %w", err)
	}
	if err := subtitle.SetSpacingAfter(240); err != nil {
		return fmt.Errorf("set subtitle spacing: %w", err)
	}

	intro, err := doc.AddParagraph()
	if err != nil {
		return fmt.Errorf("add intro paragraph: %w", err)
	}
	introRun, err := intro.AddRun()
	if err != nil {
		return fmt.Errorf("add intro run: %w", err)
	}
	if err := introRun.SetText("Use heading styles while authoring. Word rebuilds the TOC automatically when you press F9."); err != nil {
		return fmt.Errorf("set intro text: %w", err)
	}

	if err := doc.AddPageBreak(); err != nil {
		return fmt.Errorf("add page break after cover: %w", err)
	}

	return nil
}

func addTableOfContents(doc domain.Document) error {
	heading, err := doc.AddParagraph()
	if err != nil {
		return fmt.Errorf("add TOC heading: %w", err)
	}
	if err := heading.SetStyle(domain.StyleIDHeading1); err != nil {
		return fmt.Errorf("set TOC heading style: %w", err)
	}
	headingRun, err := heading.AddRun()
	if err != nil {
		return fmt.Errorf("add TOC heading run: %w", err)
	}
	if err := headingRun.SetText("Table of Contents"); err != nil {
		return fmt.Errorf("set TOC heading text: %w", err)
	}

	description, err := doc.AddParagraph()
	if err != nil {
		return fmt.Errorf("add TOC description: %w", err)
	}
	descRun, err := description.AddRun()
	if err != nil {
		return fmt.Errorf("add TOC description run: %w", err)
	}
	if err := descRun.SetText("This field stays in sync with Heading 1 and Heading 2 entries. Update it in Word when the outline changes."); err != nil {
		return fmt.Errorf("set TOC description text: %w", err)
	}
	if err := description.SetSpacingAfter(160); err != nil {
		return fmt.Errorf("set TOC description spacing: %w", err)
	}

	tocPara, err := doc.AddParagraph()
	if err != nil {
		return fmt.Errorf("add TOC paragraph: %w", err)
	}
	tocRun, err := tocPara.AddRun()
	if err != nil {
		return fmt.Errorf("add TOC run: %w", err)
	}

	tocOptions := map[string]string{
		"levels":     "1-2",
		"hyperlinks": "true",
	}
	tocField := docx.NewTOCField(tocOptions)

	placeholder := strings.TrimSpace(`
Getting Started ................................................. 3
    Installing go-docx ........................................ 4
    Creating your first document ............................... 5
    Applying heading styles .................................... 6
Building Content ................................................ 7
    Paragraph formatting .................................... 8
    Working with lists ....................................... 9
    Inserting images ......................................... 10
Appendix A (Resources) ....................................... 11
`)

	if setter, ok := tocField.(interface{ SetResult(string) }); ok {
		setter.SetResult(placeholder)
	}

	if err := tocRun.AddField(tocField); err != nil {
		return fmt.Errorf("add TOC field: %w", err)
	}

	note, err := doc.AddParagraph()
	if err != nil {
		return fmt.Errorf("add TOC note: %w", err)
	}
	if err := note.SetStyle(domain.StyleIDIntenseQuote); err != nil {
		return fmt.Errorf("set TOC note style: %w", err)
	}
	noteRun, err := note.AddRun()
	if err != nil {
		return fmt.Errorf("add TOC note run: %w", err)
	}
	if err := noteRun.SetText("Tip: Right-click inside the table and choose 'Update Field' whenever you add or remove sections."); err != nil {
		return fmt.Errorf("set TOC note text: %w", err)
	}
	if err := note.SetSpacingAfter(240); err != nil {
		return fmt.Errorf("set TOC note spacing: %w", err)
	}

	if err := doc.AddPageBreak(); err != nil {
		return fmt.Errorf("add page break after TOC: %w", err)
	}

	return nil
}

func addChapter(doc domain.Document, number int, title string, sections []string) error {
	heading, err := doc.AddParagraph()
	if err != nil {
		return fmt.Errorf("add chapter heading: %w", err)
	}
	if err := heading.SetStyle(domain.StyleIDHeading1); err != nil {
		return fmt.Errorf("set chapter heading style: %w", err)
	}
	headingRun, err := heading.AddRun()
	if err != nil {
		return fmt.Errorf("add chapter heading run: %w", err)
	}
	if err := headingRun.SetText(fmt.Sprintf("Chapter %d: %s", number, title)); err != nil {
		return fmt.Errorf("set chapter heading text: %w", err)
	}
	if err := heading.SetSpacingAfter(200); err != nil {
		return fmt.Errorf("set chapter heading spacing: %w", err)
	}

	for idx, section := range sections {
		subHeading, err := doc.AddParagraph()
		if err != nil {
			return fmt.Errorf("add subsection heading: %w", err)
		}
		if err := subHeading.SetStyle(domain.StyleIDHeading2); err != nil {
			return fmt.Errorf("set subsection style: %w", err)
		}
		subRun, err := subHeading.AddRun()
		if err != nil {
			return fmt.Errorf("add subsection run: %w", err)
		}
		if err := subRun.SetText(fmt.Sprintf("%d.%d %s", number, idx+1, section)); err != nil {
			return fmt.Errorf("set subsection text: %w", err)
		}

		body, err := doc.AddParagraph()
		if err != nil {
			return fmt.Errorf("add subsection body: %w", err)
		}
		bodyRun, err := body.AddRun()
		if err != nil {
			return fmt.Errorf("add subsection body run: %w", err)
		}
		if err := bodyRun.SetText("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Suspendisse potenti. Integer sit amet mauris sed mauris molestie aliquam."); err != nil {
			return fmt.Errorf("set subsection body text: %w", err)
		}
		if err := body.SetSpacingAfter(120); err != nil {
			return fmt.Errorf("set subsection body spacing: %w", err)
		}
	}

	if err := doc.AddPageBreak(); err != nil {
		return fmt.Errorf("add page break after chapter: %w", err)
	}

	return nil
}

func addAppendix(doc domain.Document) error {
	heading, err := doc.AddParagraph()
	if err != nil {
		return fmt.Errorf("add appendix heading: %w", err)
	}
	if err := heading.SetStyle(domain.StyleIDHeading1); err != nil {
		return fmt.Errorf("set appendix heading style: %w", err)
	}
	headingRun, err := heading.AddRun()
	if err != nil {
		return fmt.Errorf("add appendix heading run: %w", err)
	}
	if err := headingRun.SetText("Appendix A: Resources"); err != nil {
		return fmt.Errorf("set appendix heading text: %w", err)
	}
	if err := heading.SetSpacingAfter(160); err != nil {
		return fmt.Errorf("set appendix heading spacing: %w", err)
	}

	intro, err := doc.AddParagraph()
	if err != nil {
		return fmt.Errorf("add appendix intro: %w", err)
	}
	introRun, err := intro.AddRun()
	if err != nil {
		return fmt.Errorf("add appendix intro run: %w", err)
	}
	if err := introRun.SetText("Further reading and helpful links for building long-form documents with go-docx:"); err != nil {
		return fmt.Errorf("set appendix intro text: %w", err)
	}
	if err := intro.SetSpacingAfter(120); err != nil {
		return fmt.Errorf("set appendix intro spacing: %w", err)
	}

	resources := []string{
		"API reference: pkg.go.dev/github.com/mmonterroca/docxgo/v2",
		"Examples: github.com/mmonterroca/docxgo/examples",
		"Design document: docs/V2_DESIGN.md",
		"Fields deep dive: examples/04_fields",
	}

	for _, item := range resources {
		para, err := doc.AddParagraph()
		if err != nil {
			return fmt.Errorf("add resource paragraph: %w", err)
		}
		if err := para.SetStyle(domain.StyleIDListParagraph); err != nil {
			return fmt.Errorf("set resource style: %w", err)
		}
		run, err := para.AddRun()
		if err != nil {
			return fmt.Errorf("add resource run: %w", err)
		}
		if err := run.SetText("- " + item); err != nil {
			return fmt.Errorf("set resource text: %w", err)
		}
	}

	return nil
}

package main

import (
	"fmt"
	"log"

	docx "github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/domain"
)

func main() {
	doc := docx.NewDocument()

	// Configure the default section (executive summary)
	defaultSection, err := doc.DefaultSection()
	if err != nil {
		log.Fatalf("failed to obtain default section: %v", err)
	}
	_ = defaultSection.SetPageSize(domain.PageSizeLetter)
	_ = defaultSection.SetMargins(domain.Margins{
		Top:    1080,
		Right:  1080,
		Bottom: 1080,
		Left:   1440,
		Header: 720,
		Footer: 720,
	})

	// Header for first section
	header, err := defaultSection.Header(domain.HeaderDefault)
	if err != nil {
		log.Fatalf("failed to create header: %v", err)
	}
	headerPara, err := header.AddParagraph()
	if err != nil {
		log.Fatalf("failed to add header paragraph: %v", err)
	}
	headerPara.SetAlignment(domain.AlignmentRight)
	headerRun, err := headerPara.AddRun()
	if err != nil {
		log.Fatalf("failed to add header run: %v", err)
	}
	headerRun.AddText("Quarterly Report")
	headerRun.SetBold(true)

	// Footer with dynamic page numbers
	footer, err := defaultSection.Footer(domain.FooterDefault)
	if err != nil {
		log.Fatalf("failed to create footer: %v", err)
	}
	footerPara, err := footer.AddParagraph()
	if err != nil {
		log.Fatalf("failed to add footer paragraph: %v", err)
	}
	footerPara.SetAlignment(domain.AlignmentCenter)

	footerRun, err := footerPara.AddRun()
	if err != nil {
		log.Fatalf("failed to add footer run: %v", err)
	}
	footerRun.AddText("Page ")

	pageFieldRun, err := footerPara.AddRun()
	if err != nil {
		log.Fatalf("failed to add footer field run: %v", err)
	}
	pageFieldRun.AddField(docx.NewPageNumberField())

	footerRun2, err := footerPara.AddRun()
	if err != nil {
		log.Fatalf("failed to add footer text run: %v", err)
	}
	footerRun2.AddText(" of ")

	totalFieldRun, err := footerPara.AddRun()
	if err != nil {
		log.Fatalf("failed to add footer total run: %v", err)
	}
	totalFieldRun.AddField(docx.NewPageCountField())

	// Executive summary content
	title, _ := doc.AddParagraph()
	title.SetStyle(domain.StyleIDTitle)
	titleRun, _ := title.AddRun()
	titleRun.AddText("Q3 Business Review")

	intro, _ := doc.AddParagraph()
	intro.SetStyle(domain.StyleIDNormal)
	introRun, _ := intro.AddRun()
	introRun.AddText("This report highlights performance across business units and showcases the new multi-section support in go-docx v2.")

	summaryHeading, _ := doc.AddParagraph()
	summaryHeading.SetStyle(domain.StyleIDHeading1)
	summaryHeadingRun, _ := summaryHeading.AddRun()
	summaryHeadingRun.AddText("1. Executive Summary")

	bulletPoints := []string{
		"Revenue up 12% quarter-over-quarter",
		"New customer acquisitions increased by 18%",
		"Operational costs reduced by 6% through automation",
	}
	for _, text := range bulletPoints {
		item, _ := doc.AddParagraph()
		item.SetStyle(domain.StyleIDListParagraph)
		run, _ := item.AddRun()
		run.AddText(fmt.Sprintf("• %s", text))
	}

	doc.AddParagraph() // spacer

	// Create a new landscape section for wide content
	landscapeSection, err := doc.AddSectionWithBreak(domain.SectionBreakTypeNextPage)
	if err != nil {
		log.Fatalf("failed to add landscape section: %v", err)
	}
	_ = landscapeSection.SetOrientation(domain.OrientationLandscape)
	_ = landscapeSection.SetPageSize(domain.PageSizeLetter)
	_ = landscapeSection.SetColumns(2)
	_ = landscapeSection.SetMargins(domain.Margins{
		Top:    720,
		Right:  720,
		Bottom: 720,
		Left:   720,
		Header: 540,
		Footer: 540,
	})

	landscapeHeader, err := landscapeSection.Header(domain.HeaderDefault)
	if err != nil {
		log.Fatalf("failed to add landscape header: %v", err)
	}
	landscapeHeaderPara, _ := landscapeHeader.AddParagraph()
	landscapeHeaderPara.SetAlignment(domain.AlignmentCenter)
	landscapeHeaderRun, _ := landscapeHeaderPara.AddRun()
	landscapeHeaderRun.AddText("Operational Dashboards")
	landscapeHeaderRun.SetBold(true)

	landscapeHeading, _ := doc.AddParagraph()
	landscapeHeading.SetStyle(domain.StyleIDHeading1)
	landscapeHeadingRun, _ := landscapeHeading.AddRun()
	landscapeHeadingRun.AddText("2. Metrics Overview (Landscape)")

	metrics := []string{
		"Sales pipeline coverage: 3.5x",
		"Average deal cycle: 42 days",
		"Customer satisfaction (NPS): 67",
		"Support ticket resolution SLA: 94%",
		"Cloud infrastructure cost savings: 9%",
		"Website conversion rate: 3.1%",
		"Outbound response time: 3.4 hours",
		"Renewal retention: 91%",
		"Engineering release cadence: bi-weekly",
		"Escalation backlog: 6 open",
		"Regional quota attainment: 108%",
		"Marketing influenced pipeline: 54%",
	}

	for _, metric := range metrics {
		para, _ := doc.AddParagraph()
		para.SetStyle(domain.StyleIDListParagraph)
		run, _ := para.AddRun()
		run.AddText(fmt.Sprintf("• %s", metric))
	}

	landscapeFooter, err := landscapeSection.Footer(domain.FooterDefault)
	if err != nil {
		log.Fatalf("failed to add landscape footer: %v", err)
	}
	landscapeFooterPara, _ := landscapeFooter.AddParagraph()
	landscapeFooterPara.SetAlignment(domain.AlignmentCenter)
	landscapeFooterRun, _ := landscapeFooterPara.AddRun()
	landscapeFooterRun.AddText("Landscape metrics • Page ")
	landscapeFooterField, _ := landscapeFooterPara.AddRun()
	landscapeFooterField.AddField(docx.NewPageNumberField())

	doc.AddParagraph() // spacer

	// Final section returning to portrait for closing notes
	closingSection, err := doc.AddSectionWithBreak(domain.SectionBreakTypeContinuous)
	if err != nil {
		log.Fatalf("failed to add closing section: %v", err)
	}
	_ = closingSection.SetOrientation(domain.OrientationPortrait)
	_ = closingSection.SetColumns(1)
	_ = closingSection.SetMargins(domain.Margins{
		Top:    1440,
		Right:  1440,
		Bottom: 1440,
		Left:   1080,
		Header: 540,
		Footer: 720,
	})

	closingFooter, err := closingSection.Footer(domain.FooterDefault)
	if err != nil {
		log.Fatalf("failed to add closing footer: %v", err)
	}
	closingFooterPara, _ := closingFooter.AddParagraph()
	closingFooterPara.SetAlignment(domain.AlignmentRight)
	closingFooterRun, _ := closingFooterPara.AddRun()
	closingFooterRun.AddText("Next Steps • Page ")
	closingFooterField, _ := closingFooterPara.AddRun()
	closingFooterField.AddField(docx.NewPageNumberField())

	closingHeader, err := closingSection.Header(domain.HeaderDefault)
	if err != nil {
		log.Fatalf("failed to add closing header: %v", err)
	}
	closingHeaderPara, _ := closingHeader.AddParagraph()
	closingHeaderRun, _ := closingHeaderPara.AddRun()
	closingHeaderRun.AddText("Action Items")
	closingHeaderRun.SetBold(true)

	closingHeading, _ := doc.AddParagraph()
	closingHeading.SetStyle(domain.StyleIDHeading1)
	closingHeadingRun, _ := closingHeading.AddRun()
	closingHeadingRun.AddText("3. Next Steps")

	actions := []string{
		"Finalize FY roadmap with updated revenue targets",
		"Expand automation program to customer success workflows",
		"Pilot advanced analytics dashboard rollout",
	}
	for _, action := range actions {
		item, _ := doc.AddParagraph()
		item.SetStyle(domain.StyleIDListParagraph)
		run, _ := item.AddRun()
		run.AddText(fmt.Sprintf("• %s", action))
	}

	conclusion, _ := doc.AddParagraph()
	conclusion.SetStyle(domain.StyleIDNormal)
	conclusionRun, _ := conclusion.AddRun()
	conclusionRun.AddText("The new multi-section capabilities allow individual sections to own their headers, footers, margins, and layouts, enabling rich document scenarios without leaving go-docx.")

	if err := doc.SaveAs("11_multi_section_demo.docx"); err != nil {
		log.Fatalf("failed to save document: %v", err)
	}

	fmt.Println("Document created successfully: 11_multi_section_demo.docx")
	fmt.Println("\nKey highlights:")
	fmt.Println("- Portrait default section with executive summary and dynamic footer")
	fmt.Println("- Next-page break into landscape section with two-column layout")
	fmt.Println("- Continuous break back to portrait for closing action items")
	fmt.Println("- Per-section headers to showcase unique content per layout")
}

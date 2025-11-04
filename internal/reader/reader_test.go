/*
MIT License

Copyright (c) 2025 Misael Monterroca

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

package reader

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/core"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
)

func TestLoadPackageFromBytes(t *testing.T) {
	doc := core.NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("Hello, reader!"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pkg, err := LoadPackageFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}

	if len(pkg.MainDocument) == 0 {
		t.Fatalf("expected main document data")
	}
	if pkg.ContentTypes == nil {
		t.Fatalf("expected content types")
	}
	if pkg.RootRelationships == nil {
		t.Fatalf("expected root relationships part")
	}
	if pkg.DocumentRelationships == nil {
		t.Fatalf("expected document relationships part")
	}
	if got := pkg.contentTypeFor("word/document.xml"); got == "" {
		t.Fatalf("expected content type for main document")
	}
	if pkg.PackageSize == 0 {
		t.Fatalf("expected package size to be recorded")
	}
}

func TestParsePackage(t *testing.T) {
	doc := core.NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("Parse! "); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	if err := run.SetBold(true); err != nil {
		t.Fatalf("SetBold: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pkg, err := LoadPackageFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}
	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	if parsed.DocumentTree == nil {
		t.Fatalf("expected parsed document tree")
	}
	if parsed.DocumentTree.Name.Local != "document" {
		t.Fatalf("unexpected document local name: %s", parsed.DocumentTree.Name.Local)
	}
	if parsed.DocumentTree.Name.Space != constants.NamespaceMain {
		t.Fatalf("unexpected document namespace: %s", parsed.DocumentTree.Name.Space)
	}
	if parsed.DocumentRelationships == nil {
		t.Fatalf("expected document relationships")
	}
	if parsed.RootRelationships == nil {
		t.Fatalf("expected root relationships")
	}
	if parsed.Package == nil {
		t.Fatalf("expected parsed package to retain raw package reference")
	}
	if parsed.CorePropertiesTree == nil {
		t.Fatalf("expected core properties to be parsed")
	}
	if parsed.AppPropertiesTree == nil {
		t.Fatalf("expected app properties to be parsed")
	}
}

func TestParsePackageNil(t *testing.T) {
	if _, err := ParsePackage(nil); err == nil {
		t.Fatalf("expected error when parsing nil package")
	}
}

func TestReconstructDocumentFormatting(t *testing.T) {
	doc := core.NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	if err := para.SetSpacingBefore(360); err != nil {
		t.Fatalf("SetSpacingBefore: %v", err)
	}
	if err := para.SetSpacingAfter(120); err != nil {
		t.Fatalf("SetSpacingAfter: %v", err)
	}
	if err := para.SetAlignment(domain.AlignmentCenter); err != nil {
		t.Fatalf("SetAlignment: %v", err)
	}
	indent := domain.Indentation{Left: 720, Right: 360, FirstLine: 240}
	if err := para.SetIndent(indent); err != nil {
		t.Fatalf("SetIndent: %v", err)
	}
	if err := para.SetLineSpacing(domain.LineSpacing{Rule: domain.LineSpacingExact, Value: 480}); err != nil {
		t.Fatalf("SetLineSpacing: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("Spacing sample"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	if err := run.SetBold(true); err != nil {
		t.Fatalf("SetBold: %v", err)
	}
	if err := run.SetItalic(true); err != nil {
		t.Fatalf("SetItalic: %v", err)
	}
	if err := run.SetUnderline(domain.UnderlineDouble); err != nil {
		t.Fatalf("SetUnderline: %v", err)
	}
	if err := run.SetStrike(true); err != nil {
		t.Fatalf("SetStrike: %v", err)
	}
	if err := run.SetHighlight(domain.HighlightGreen); err != nil {
		t.Fatalf("SetHighlight: %v", err)
	}
	expectedColor := domain.Color{R: 0x12, G: 0x34, B: 0x56}
	if err := run.SetColor(expectedColor); err != nil {
		t.Fatalf("SetColor: %v", err)
	}
	if err := run.SetSize(48); err != nil {
		t.Fatalf("SetSize: %v", err)
	}
	if err := run.SetFont(domain.Font{Name: "Times New Roman", EastAsia: "SimSun", CS: "Arial"}); err != nil {
		t.Fatalf("SetFont: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pkg, err := LoadPackageFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}

	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	reconstructed, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}

	paras := reconstructed.Paragraphs()
	if len(paras) != 1 {
		t.Fatalf("expected 1 paragraph, got %d", len(paras))
	}

	got := paras[0]
	if got.SpacingBefore() != 360 {
		t.Fatalf("SpacingBefore mismatch: got %d", got.SpacingBefore())
	}
	if got.SpacingAfter() != 120 {
		t.Fatalf("SpacingAfter mismatch: got %d", got.SpacingAfter())
	}
	if got.Alignment() != domain.AlignmentCenter {
		t.Fatalf("Alignment mismatch: got %v", got.Alignment())
	}
	recoveredIndent := got.Indent()
	if recoveredIndent.Left != 720 || recoveredIndent.Right != 360 || recoveredIndent.FirstLine != 240 {
		t.Fatalf("Indent mismatch: %+v", recoveredIndent)
	}
	lineSpacing := got.LineSpacing()
	if lineSpacing.Rule != domain.LineSpacingExact {
		t.Fatalf("unexpected line spacing rule: %v", lineSpacing.Rule)
	}
	if lineSpacing.Value != 480 {
		t.Fatalf("unexpected line spacing value: %d", lineSpacing.Value)
	}
	if got.Text() != "Spacing sample" {
		t.Fatalf("unexpected paragraph text: %q", got.Text())
	}
	runs := got.Runs()
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	recoveredRun := runs[0]
	if !recoveredRun.Bold() {
		t.Fatalf("expected run to be bold")
	}
	if !recoveredRun.Italic() {
		t.Fatalf("expected run to be italic")
	}
	if !recoveredRun.Strike() {
		t.Fatalf("expected run to be strike")
	}
	if recoveredRun.Underline() != domain.UnderlineDouble {
		t.Fatalf("unexpected underline: %v", recoveredRun.Underline())
	}
	if recoveredRun.Size() != 48 {
		t.Fatalf("unexpected size: %d", recoveredRun.Size())
	}
	if recoveredRun.Color() != expectedColor {
		t.Fatalf("unexpected color: %+v", recoveredRun.Color())
	}
	if recoveredRun.Highlight() != domain.HighlightGreen {
		t.Fatalf("unexpected highlight: %v", recoveredRun.Highlight())
	}
	font := recoveredRun.Font()
	if font.Name != "Times New Roman" {
		t.Fatalf("unexpected font name: %s", font.Name)
	}
	if font.EastAsia != "SimSun" {
		t.Fatalf("unexpected east asia font: %s", font.EastAsia)
	}
	if font.CS != "Arial" {
		t.Fatalf("unexpected complex script font: %s", font.CS)
	}
}

func TestReconstructRunContentExtensions(t *testing.T) {
	doc := core.NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}

	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("Intro"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	if err := run.AddText("\tTab"); err != nil {
		t.Fatalf("AddText: %v", err)
	}
	if err := run.AddBreak(domain.BreakTypeLine); err != nil {
		t.Fatalf("AddBreak: %v", err)
	}

	fieldRun, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun field: %v", err)
	}
	pageField := core.NewField(domain.FieldTypePageNumber)
	if err := fieldRun.AddField(pageField); err != nil {
		t.Fatalf("AddField: %v", err)
	}
	if err := fieldRun.SetText("1"); err != nil {
		t.Fatalf("SetText field: %v", err)
	}

	hyperlinkRun, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun hyperlink: %v", err)
	}
	hyperlinkField := core.NewField(domain.FieldTypeHyperlink)
	if accessor, ok := hyperlinkField.(interface{ SetProperty(string, string) }); ok {
		accessor.SetProperty("url", "https://example.com")
	}
	if err := hyperlinkField.SetCode(`HYPERLINK "https://example.com"`); err != nil {
		t.Fatalf("SetCode hyperlink: %v", err)
	}
	if err := hyperlinkRun.AddField(hyperlinkField); err != nil {
		t.Fatalf("AddField hyperlink: %v", err)
	}
	if err := hyperlinkRun.SetText("Example"); err != nil {
		t.Fatalf("SetText hyperlink: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pkg, err := LoadPackageFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}

	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	reconstructed, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}

	paras := reconstructed.Paragraphs()
	if len(paras) == 0 {
		t.Fatalf("expected paragraphs")
	}

	reconstructedRuns := paras[0].Runs()
	if len(reconstructedRuns) < 3 {
		t.Fatalf("expected at least 3 runs, got %d", len(reconstructedRuns))
	}

	firstRun := reconstructedRuns[0]
	if !strings.Contains(firstRun.Text(), "\t") {
		t.Fatalf("expected tab character in first run text: %q", firstRun.Text())
	}
	if breakAccessor, ok := firstRun.(interface{ Breaks() []domain.BreakType }); ok {
		if len(breakAccessor.Breaks()) == 0 {
			t.Fatalf("expected line break in first run")
		}
	}

	var (
		pageFieldFound bool
		hyperFieldURL  string
	)

	for _, candidate := range reconstructedRuns {
		runFields, ok := candidate.(interface{ Fields() []domain.Field })
		if !ok {
			continue
		}

		for _, f := range runFields.Fields() {
			switch f.Type() {
			case domain.FieldTypePageNumber:
				pageFieldFound = true
			case domain.FieldTypeHyperlink:
				if accessor, ok := f.(interface{ GetProperty(string) (string, bool) }); ok {
					if url, ok := accessor.GetProperty("url"); ok {
						hyperFieldURL = url
					}
				}
			}
		}
	}

	if !pageFieldFound {
		t.Fatalf("expected page number field to round-trip")
	}
	if hyperFieldURL != "https://example.com" {
		t.Fatalf("unexpected hyperlink url: %q", hyperFieldURL)
	}
}

func TestReconstructImageRuns(t *testing.T) {
	imgPath := createTestPNG(t)

	doc := core.NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}

	originalImg, err := para.AddImage(imgPath)
	if err != nil {
		t.Fatalf("AddImage: %v", err)
	}

	if err := originalImg.SetDescription("Sample image"); err != nil {
		t.Fatalf("SetDescription: %v", err)
	}

	originalData := originalImg.Data()
	originalTarget := originalImg.Target()
	originalSize := originalImg.Size()

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pkg, err := LoadPackageFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}

	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	reconstructed, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}

	paras := reconstructed.Paragraphs()
	if len(paras) == 0 {
		t.Fatalf("expected paragraphs")
	}

	var (
		foundImage   domain.Image
		imageRun     domain.Run
		runSummaries []string
	)

	for pi, p := range paras {
		for ri, run := range p.Runs() {
			provider, ok := run.(interface{ Image() domain.Image })
			hasImage := false
			if ok {
				if img := provider.Image(); img != nil {
					hasImage = true
				}
			}
			runSummaries = append(runSummaries, fmt.Sprintf("para[%d] run[%d]: text=%q, image=%t", pi, ri, run.Text(), hasImage))
			if !ok {
				continue
			}
			if img := provider.Image(); img != nil {
				foundImage = img
				imageRun = run
				break
			}
		}
		if foundImage != nil {
			break
		}
	}

	if foundImage == nil {
		t.Fatalf("expected hydrated image run; got runs: %s", strings.Join(runSummaries, "; "))
	}

	if imageRun.Text() != "" {
		t.Fatalf("expected image run text to be empty, got %q", imageRun.Text())
	}

	if foundImage.Description() != "Sample image" {
		t.Fatalf("unexpected image description: %q", foundImage.Description())
	}

	if foundImage.Target() != originalTarget {
		t.Fatalf("expected image target %q, got %q", originalTarget, foundImage.Target())
	}

	if size := foundImage.Size(); size.WidthEMU != originalSize.WidthEMU || size.HeightEMU != originalSize.HeightEMU {
		t.Fatalf("unexpected hydrated image size: %+v", size)
	}

	if gotData := foundImage.Data(); len(gotData) == 0 {
		t.Fatalf("expected image data")
	} else if !bytes.Equal(gotData, originalData) {
		t.Fatalf("hydrated image data mismatch")
	}

	if len(paras[0].Images()) == 0 {
		t.Fatalf("expected paragraph to register hydrated image")
	}

	var roundTrip bytes.Buffer
	if _, err := reconstructed.WriteTo(&roundTrip); err != nil {
		t.Fatalf("WriteTo after hydration: %v", err)
	}
}

func TestReconstructParagraphNumbering(t *testing.T) {
	numberingXML := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="1">
    <w:lvl w:ilvl="0">
      <w:start w:val="1"/>
      <w:numFmt w:val="bullet"/>
      <w:lvlText w:val="•"/>
      <w:lvlJc w:val="left"/>
    </w:lvl>
  </w:abstractNum>
  <w:num w:numId="1">
    <w:abstractNumId w:val="1"/>
  </w:num>
</w:numbering>`) // minimal numbering definition

	doc := core.NewDocument()
	para, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph: %v", err)
	}
	if err := para.SetNumbering(domain.NumberingReference{ID: 1, Level: 0}); err != nil {
		t.Fatalf("SetNumbering: %v", err)
	}
	run, err := para.AddRun()
	if err != nil {
		t.Fatalf("AddRun: %v", err)
	}
	if err := run.SetText("List item 1"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	config, ok := doc.(interface {
		SetNumberingPart([]byte, string)
		RegisterExistingRelationship(string, string, string, string) error
	})
	if !ok {
		t.Fatalf("document does not expose numbering configuration hooks")
	}
	config.SetNumberingPart(numberingXML, "numbering.xml")
	if err := config.RegisterExistingRelationship("rId50", constants.RelTypeNumbering, "numbering.xml", "Internal"); err != nil {
		t.Fatalf("registerExistingRelationship: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pkg, err := LoadPackageFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}
	if len(pkg.Numbering) == 0 {
		t.Fatalf("expected numbering part to be written")
	}

	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	reconstructed, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}

	paras := reconstructed.Paragraphs()
	if len(paras) != 1 {
		t.Fatalf("expected 1 paragraph, got %d", len(paras))
	}

	ref, ok := paras[0].Numbering()
	if !ok {
		t.Fatalf("expected numbering reference to be hydrated")
	}
	if ref.ID != 1 || ref.Level != 0 {
		t.Fatalf("unexpected numbering ref: %+v", ref)
	}

	if accessor, ok := reconstructed.(interface{ NumberingPartInfo() ([]byte, string) }); ok {
		data, target := accessor.NumberingPartInfo()
		if target != "numbering.xml" {
			t.Fatalf("unexpected numbering target: %q", target)
		}
		if !bytes.Equal(data, numberingXML) {
			t.Fatalf("numbering part mismatch")
		}
	}

	var roundTrip bytes.Buffer
	if _, err := reconstructed.WriteTo(&roundTrip); err != nil {
		t.Fatalf("WriteTo (roundtrip): %v", err)
	}

	roundPkg, err := LoadPackageFromBytes(roundTrip.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes roundtrip: %v", err)
	}
	if len(roundPkg.Numbering) == 0 {
		t.Fatalf("expected numbering part to persist after roundtrip")
	}
	if !bytes.Equal(roundPkg.Numbering, numberingXML) {
		t.Fatalf("roundtrip numbering mismatch")
	}
}

func TestReconstructDocumentSections(t *testing.T) {
	doc := core.NewDocument()

	defaultSection, err := doc.DefaultSection()
	if err != nil {
		t.Fatalf("DefaultSection: %v", err)
	}

	landscapeSize := domain.PageSize{Width: domain.PageSizeA4.Height, Height: domain.PageSizeA4.Width}
	if err := defaultSection.SetPageSize(landscapeSize); err != nil {
		t.Fatalf("SetPageSize default: %v", err)
	}
	if err := defaultSection.SetOrientation(domain.OrientationLandscape); err != nil {
		t.Fatalf("SetOrientation default: %v", err)
	}
	marginsDefault := domain.Margins{Top: 720, Right: 900, Bottom: 720, Left: 1440, Header: 480, Footer: 600}
	if err := defaultSection.SetMargins(marginsDefault); err != nil {
		t.Fatalf("SetMargins default: %v", err)
	}
	if err := defaultSection.SetColumns(2); err != nil {
		t.Fatalf("SetColumns default: %v", err)
	}

	headDefault, err := defaultSection.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header default: %v", err)
	}
	headPara, err := headDefault.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph header default: %v", err)
	}
	headRun, err := headPara.AddRun()
	if err != nil {
		t.Fatalf("AddRun header default: %v", err)
	}
	if err := headRun.SetText("Section 1 Header"); err != nil {
		t.Fatalf("SetText header default: %v", err)
	}

	bodyPara1, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph body default: %v", err)
	}
	bodyRun1, err := bodyPara1.AddRun()
	if err != nil {
		t.Fatalf("AddRun body default: %v", err)
	}
	if err := bodyRun1.SetText("Default section content"); err != nil {
		t.Fatalf("SetText body default: %v", err)
	}

	secondSection, err := doc.AddSectionWithBreak(domain.SectionBreakTypeEvenPage)
	if err != nil {
		t.Fatalf("AddSectionWithBreak: %v", err)
	}
	if err := secondSection.SetPageSize(domain.PageSizeLetter); err != nil {
		t.Fatalf("SetPageSize second: %v", err)
	}
	if err := secondSection.SetOrientation(domain.OrientationPortrait); err != nil {
		t.Fatalf("SetOrientation second: %v", err)
	}
	marginsSecond := domain.Margins{Top: 1440, Right: 1440, Bottom: 1440, Left: 1440, Header: 720, Footer: 960}
	if err := secondSection.SetMargins(marginsSecond); err != nil {
		t.Fatalf("SetMargins second: %v", err)
	}
	if err := secondSection.SetColumns(3); err != nil {
		t.Fatalf("SetColumns second: %v", err)
	}

	footDefault, err := secondSection.Footer(domain.FooterDefault)
	if err != nil {
		t.Fatalf("Footer second: %v", err)
	}
	footPara, err := footDefault.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph footer: %v", err)
	}
	footRun, err := footPara.AddRun()
	if err != nil {
		t.Fatalf("AddRun footer: %v", err)
	}
	if err := footRun.SetText("Section 2 Footer"); err != nil {
		t.Fatalf("SetText footer: %v", err)
	}

	bodyPara2, err := doc.AddParagraph()
	if err != nil {
		t.Fatalf("AddParagraph body second: %v", err)
	}
	bodyRun2, err := bodyPara2.AddRun()
	if err != nil {
		t.Fatalf("AddRun body second: %v", err)
	}
	if err := bodyRun2.SetText("Second section content"); err != nil {
		t.Fatalf("SetText body second: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pkg, err := LoadPackageFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}

	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	reconstructed, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}

	sections := reconstructed.Sections()
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}

	rehydratedDefault := sections[0]
	if rehydratedDefault.Orientation() != domain.OrientationLandscape {
		t.Fatalf("expected default section landscape orientation, got %v", rehydratedDefault.Orientation())
	}
	if size := rehydratedDefault.PageSize(); size.Width <= size.Height {
		t.Fatalf("expected width > height for landscape page size, got %+v", size)
	}
	if cols := rehydratedDefault.Columns(); cols != 2 {
		t.Fatalf("expected default section columns=2, got %d", cols)
	}
	if gotMargins := rehydratedDefault.Margins(); gotMargins.Left != marginsDefault.Left || gotMargins.Right != marginsDefault.Right || gotMargins.Header != marginsDefault.Header {
		t.Fatalf("unexpected default section margins: %+v", gotMargins)
	}

	defaultHeader, err := rehydratedDefault.Header(domain.HeaderDefault)
	if err != nil {
		t.Fatalf("Header(default) after hydration: %v", err)
	}
	defaultHeaderParas := defaultHeader.Paragraphs()
	if len(defaultHeaderParas) == 0 {
		t.Fatalf("expected default section header paragraphs")
	}
	if text := defaultHeaderParas[0].Text(); text != "Section 1 Header" {
		t.Fatalf("unexpected default header text: %q", text)
	}

	rehydratedSecond := sections[1]
	if rehydratedSecond.Orientation() != domain.OrientationPortrait {
		t.Fatalf("expected second section portrait orientation, got %v", rehydratedSecond.Orientation())
	}
	if cols := rehydratedSecond.Columns(); cols != 3 {
		t.Fatalf("expected second section columns=3, got %d", cols)
	}
	if gotMargins := rehydratedSecond.Margins(); gotMargins.Footer != marginsSecond.Footer {
		t.Fatalf("unexpected second section footer margin: %+v", gotMargins)
	}

	secondFooter, err := rehydratedSecond.Footer(domain.FooterDefault)
	if err != nil {
		t.Fatalf("Footer(default) after hydration: %v", err)
	}
	footerParas := secondFooter.Paragraphs()
	if len(footerParas) == 0 {
		t.Fatalf("expected second section footer paragraphs")
	}
	if text := footerParas[0].Text(); text != "Section 2 Footer" {
		t.Fatalf("unexpected second footer text: %q", text)
	}

	foundBreak := false
	for _, block := range reconstructed.Blocks() {
		if block.SectionBreak == nil {
			continue
		}
		if block.SectionBreak.Type == domain.SectionBreakTypeEvenPage {
			foundBreak = true
			break
		}
	}
	if !foundBreak {
		t.Fatalf("expected section break block to be preserved")
	}
}

func createTestPNG(t *testing.T) string {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: uint8(20 * x), G: uint8(20 * y), B: 200, A: 255})
		}
	}

	file, err := os.CreateTemp(t.TempDir(), "img-*.png")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}

	return file.Name()
}

func TestReconstructDocumentTable(t *testing.T) {
	doc := core.NewDocument()
	table, err := doc.AddTable(1, 2)
	if err != nil {
		t.Fatalf("AddTable: %v", err)
	}

	row, err := table.Row(0)
	if err != nil {
		t.Fatalf("Row: %v", err)
	}

	for idx, text := range []string{"Cell 1", "Cell 2"} {
		cell, err := row.Cell(idx)
		if err != nil {
			t.Fatalf("Cell(%d): %v", idx, err)
		}
		para, err := cell.AddParagraph()
		if err != nil {
			t.Fatalf("AddParagraph cell %d: %v", idx, err)
		}
		run, err := para.AddRun()
		if err != nil {
			t.Fatalf("AddRun cell %d: %v", idx, err)
		}
		if err := run.SetText(text); err != nil {
			t.Fatalf("SetText cell %d: %v", idx, err)
		}
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	pkg, err := LoadPackageFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadPackageFromBytes: %v", err)
	}

	parsed, err := ParsePackage(pkg)
	if err != nil {
		t.Fatalf("ParsePackage: %v", err)
	}

	reconstructed, err := ReconstructDocument(parsed)
	if err != nil {
		t.Fatalf("ReconstructDocument: %v", err)
	}

	tables := reconstructed.Tables()
	if len(tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(tables))
	}

	r, err := tables[0].Row(0)
	if err != nil {
		t.Fatalf("Row(0): %v", err)
	}

	for idx, expected := range []string{"Cell 1", "Cell 2"} {
		cell, err := r.Cell(idx)
		if err != nil {
			t.Fatalf("Cell(%d): %v", idx, err)
		}
		paras := cell.Paragraphs()
		if len(paras) == 0 {
			t.Fatalf("expected paragraphs in cell %d", idx)
		}
		if got := paras[0].Text(); got != expected {
			t.Fatalf("unexpected text in cell %d: %q", idx, got)
		}
	}
}

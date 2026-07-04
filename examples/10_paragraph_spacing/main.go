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

package main

import (
	"fmt"
	"log"

	docx "github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/domain"
)

func main() {
	doc := docx.NewDocument()

	intro, err := doc.AddParagraph()
	if err != nil {
		log.Fatalf("AddParagraph: %v", err)
	}
	if err := intro.SetSpacingAfter(120); err != nil {
		log.Fatalf("SetSpacingAfter: %v", err)
	}
	introRun, err := intro.AddRun()
	if err != nil {
		log.Fatalf("AddRun: %v", err)
	}
	if err := introRun.SetText("Custom paragraph spacing makes long-form content easier to scan."); err != nil {
		log.Fatalf("SetText: %v", err)
	}

	spacious, err := doc.AddParagraph()
	if err != nil {
		log.Fatalf("AddParagraph: %v", err)
	}
	if err := spacious.SetSpacingBefore(360); err != nil {
		log.Fatalf("SetSpacingBefore: %v", err)
	}
	if err := spacious.SetSpacingAfter(180); err != nil {
		log.Fatalf("SetSpacingAfter: %v", err)
	}
	if err := spacious.SetLineSpacing(domain.LineSpacing{Rule: domain.LineSpacingExact, Value: 480}); err != nil {
		log.Fatalf("SetLineSpacing: %v", err)
	}
	spaciousRun, err := spacious.AddRun()
	if err != nil {
		log.Fatalf("AddRun: %v", err)
	}
	if err := spaciousRun.SetText("This paragraph uses exact 24pt line spacing with extra padding above and below."); err != nil {
		log.Fatalf("SetText: %v", err)
	}

	tight, err := doc.AddParagraph()
	if err != nil {
		log.Fatalf("AddParagraph: %v", err)
	}
	if err := tight.SetLineSpacing(domain.LineSpacing{Rule: domain.LineSpacingAtLeast, Value: 200}); err != nil {
		log.Fatalf("SetLineSpacing: %v", err)
	}
	tightRun, err := tight.AddRun()
	if err != nil {
		log.Fatalf("AddRun: %v", err)
	}
	if err := tightRun.SetText("This block shows tighter line spacing using the \"at least\" rule for compact text."); err != nil {
		log.Fatalf("SetText: %v", err)
	}

	filename := "10_paragraph_spacing_demo.docx"
	if err := doc.SaveAs(filename); err != nil {
		log.Fatalf("SaveAs: %v", err)
	}

	fmt.Printf("Generated %s\n", filename)
}

/*
MIT License

Copyright (c) 2025 Misael Monterroca
Copyright (c) 2020-2023 fumiama (original go-docx)

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

	"github.com/mmonterroca/docxgo"
)

func main() {
	// Create a new document
	doc := docx.NewDocument()

	// Add a title paragraph
	title, err := doc.AddParagraph()
	if err != nil {
		log.Fatalf("Failed to add title paragraph: %v", err)
	}

	titleRun, err := title.AddRun()
	if err != nil {
		log.Fatalf("Failed to add title run: %v", err)
	}

	titleRun.SetText("Hello from go-docx v2!")
	titleRun.SetBold(true)
	titleRun.SetSize(32) // 16pt

	// Add a normal paragraph
	para, err := doc.AddParagraph()
	if err != nil {
		log.Fatalf("Failed to add paragraph: %v", err)
	}

	run, err := para.AddRun()
	if err != nil {
		log.Fatalf("Failed to add run: %v", err)
	}

	run.SetText("This is a simple document created with the new v2 API.")

	// Add another paragraph with formatting
	para2, err := doc.AddParagraph()
	if err != nil {
		log.Fatalf("Failed to add paragraph 2: %v", err)
	}

	run2, err := para2.AddRun()
	if err != nil {
		log.Fatalf("Failed to add run 2: %v", err)
	}

	run2.SetText("This text is ")
	
	run3, err := para2.AddRun()
	if err != nil {
		log.Fatalf("Failed to add run 3: %v", err)
	}
	
	run3.SetText("bold and italic")
	run3.SetBold(true)
	run3.SetItalic(true)

	run4, err := para2.AddRun()
	if err != nil {
		log.Fatalf("Failed to add run 4: %v", err)
	}
	
	run4.SetText("!")

	// Save the document
	filename := "hello_v2.docx"
	if err := doc.SaveAs(filename); err != nil {
		log.Fatalf("Failed to save document: %v", err)
	}

	fmt.Printf("✅ Document created successfully: %s\n", filename)
}

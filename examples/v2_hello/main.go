/*
   Copyright (c) 2025 SlideLang

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

package main

import (
	"fmt"
	"log"

	"github.com/SlideLang/go-docx/v2"
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

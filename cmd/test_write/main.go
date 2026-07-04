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

// Command test_write demonstrates basic document creation and writing using go-docx.
package main

import (
	"bytes"
	"fmt"

	"github.com/mmonterroca/docxgo/v2/internal/core"
)

func main() {
	// Create document
	doc := core.NewDocument()

	para, _ := doc.AddParagraph()
	run, _ := para.AddRun()
	if err := run.SetText("Hello, World!"); err != nil {
		fmt.Printf("Error setting text: %v\n", err)
		return
	}

	// Write to buffer
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Generated .docx file: %d bytes\n", buf.Len())
	fmt.Println("Success! File written to memory")
}

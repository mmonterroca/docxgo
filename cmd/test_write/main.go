// Command test_write demonstrates basic document creation and writing using go-docx.
package main

import (
	"bytes"
	"fmt"

	"github.com/mmonterroca/docxgo/internal/core"
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

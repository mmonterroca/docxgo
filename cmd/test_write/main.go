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
	run.SetText("Hello, World!")

	// Write to buffer
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Generated .docx file: %d bytes\n", buf.Len())
	fmt.Println("Success! File written to memory")
}

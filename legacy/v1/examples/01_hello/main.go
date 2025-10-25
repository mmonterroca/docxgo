package main

import (
"fmt"
"os"

"github.com/fumiama/go-docx"
)

// Example 1: Hello World - The simplest possible document
func main() {
fmt.Println("Creating Hello World document...")

// Create a new document
doc := docx.New()

// Add a paragraph with text
para := doc.AddParagraph()
para.AddText("Hello, World!")

// Set page size (important: do this at the END)
doc.WithA4Page()

// Save the document
f, err := os.Create("01_hello_world.docx")
if err != nil {
panic(err)
}
defer f.Close()

_, err = doc.WriteTo(f)
if err != nil {
panic(err)
}

fmt.Println("✅ Document created successfully: 01_hello_world.docx")
fmt.Println("   Open it in Microsoft Word to view")
}

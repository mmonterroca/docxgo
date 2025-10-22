package main

import (
"fmt"
"os"

"github.com/fumiama/go-docx"
)

// Example 2: Formatted Text - Demonstrates various text formatting options
func main() {
fmt.Println("Creating formatted text document...")

doc := docx.New().WithDefaultTheme()

// Title
title := doc.AddParagraph()
title.AddText("Text Formatting Examples").Bold().Size("32").Color("2E75B6")
title.Justification("center")

doc.AddParagraph() // Empty line

// Bold text
p1 := doc.AddParagraph()
p1.AddText("This is bold text").Bold()

// Italic text
p2 := doc.AddParagraph()
p2.AddText("This is italic text").Italic()

// Colored text
p3 := doc.AddParagraph()
p3.AddText("This is red text").Color("FF0000")

// Combined formatting
p4 := doc.AddParagraph()
p4.AddText("This is bold, italic, and blue").Bold().Italic().Color("0000FF")

// Large text
p5 := doc.AddParagraph()
p5.AddText("This is large text").Size("32") // 16pt

doc.WithA4Page()

f, err := os.Create("02_formatted_text.docx")
if err != nil {
panic(err)
}
defer f.Close()

doc.WriteTo(f)

fmt.Println("✅ Document created successfully: 02_formatted_text.docx")
}

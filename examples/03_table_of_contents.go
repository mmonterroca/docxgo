package main

import (
"fmt"
"os"

"github.com/fumiama/go-docx"
)

// Example 3: Table of Contents
func main() {
fmt.Println("Creating document with Table of Contents...")

doc := docx.New().WithDefaultTheme()

title := doc.AddParagraph()
title.AddText("Professional Technical Document").Bold().Size("32").Color("2E75B6")
title.Justification("center")

doc.AddParagraph()

opts := docx.DefaultTOCOptions()
opts.Title = "Table of Contents"
opts.Depth = 3
doc.AddTOC(opts)

doc.AddParagraph().AddPageBreaks()

h1 := doc.AddHeadingWithTOC("1. Introduction", 1, 1)
h1.Style("Heading1")

p1 := doc.AddParagraph()
p1.AddText("This demonstrates TOC features.")

h11 := doc.AddHeadingWithTOC("1.1 Background", 2, 2)
h11.Style("Heading2")

p11 := doc.AddParagraph()
p11.AddText("Background information.")

h2 := doc.AddHeadingWithTOC("2. Main Content", 1, 3)
h2.Style("Heading1")

p2 := doc.AddParagraph()
p2.AddText("Main content section.")

footer := doc.AddParagraph()
footer.AddText("Page ")
footer.AddPageField()
footer.AddText(" of ")
footer.AddNumPagesField()
footer.Justification("center")

doc.WithA4Page()

f, _ := os.Create("03_table_of_contents.docx")
defer f.Close()
doc.WriteTo(f)

fmt.Println("✅ Document created: 03_table_of_contents.docx")
fmt.Println("   Open in Word and press Ctrl+A then F9 to update TOC")
}

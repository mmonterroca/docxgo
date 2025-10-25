/*
   Copyright (c) 2020 gingfrederik
   Copyright (c) 2021 Gonzalo Fernandez-Victorio
   Copyright (c) 2021 Basement Crowd Ltd (https://www.basementcrowd.com)
   Copyright (c) 2023 Fumiama Minamoto (源文雨)

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

package docx

// ensureParagraphProperties ensures ParagraphProperties is initialized
func (p *Paragraph) ensureParagraphProperties() {
	if p.Properties == nil {
		p.Properties = &ParagraphProperties{}
	}
}

// ensureRunProperties ensures RunProperties within ParagraphProperties is initialized
func (p *Paragraph) ensureRunProperties() {
	p.ensureParagraphProperties()
	if p.Properties.RunProperties == nil {
		p.Properties.RunProperties = &RunProperties{}
	}
}

// AddParagraph adds a new paragraph
func (f *Docx) AddParagraph() *Paragraph {
	p := &Paragraph{
		Children: make([]interface{}, 0, DefaultParagraphCapacity),
		file:     f,
	}
	f.Document.Body.Items = append(f.Document.Body.Items, p)
	return p
}

// AddParagraph adds a new paragraph
func (c *WTableCell) AddParagraph() *Paragraph {
	c.Paragraphs = append(c.Paragraphs, &Paragraph{
		Children: make([]interface{}, 0, DefaultParagraphCapacity),
		file:     c.file,
	})

	return c.Paragraphs[len(c.Paragraphs)-1]
}

// Justification allows to set para's horizonal alignment
//
//	w:jc 属性的取值可以是以下之一：
//		start：左对齐。
//		center：居中对齐。
//		end：右对齐。
//		both：两端对齐。
//		distribute：分散对齐。
func (p *Paragraph) Justification(val string) *Paragraph {
	p.ensureParagraphProperties()
	p.Properties.Justification = &Justification{Val: val}
	return p
}

// AddPageBreaks adds a pagebreaks
func (p *Paragraph) AddPageBreaks() *Run {
	c := make([]interface{}, 1, DefaultParagraphCapacity)
	c[0] = &BarterRabbet{
		Type: PageBreakType,
	}
	run := &Run{
		RunProperties: &RunProperties{},
		Children:      c,
	}
	p.Children = append(p.Children, run)
	return run
}

// Style name
func (p *Paragraph) Style(val string) *Paragraph {
	p.ensureParagraphProperties()
	p.Properties.Style = &Style{Val: val}
	return p
}

// NumPr number properties
func (p *Paragraph) NumPr(numID, ilvl string) *Paragraph {
	p.ensureParagraphProperties()
	p.ensureRunProperties()
	p.Properties.NumProperties = &NumProperties{
		NumID: &NumID{
			Val: numID,
		},
		Ilvl: &Ilevel{
			Val: ilvl,
		},
	}
	return p
}

// NumFont sets the font for numbering
func (p *Paragraph) NumFont(ascii, eastAsia, hansi, hint string) *Paragraph {
	p.ensureParagraphProperties()
	p.ensureRunProperties()
	p.Properties.RunProperties.Fonts = &RunFonts{
		ASCII:    ascii,
		EastAsia: eastAsia,
		HAnsi:    hansi,
		Hint:     hint,
	}
	return p
}

// NumSize sets the size for numbering
func (p *Paragraph) NumSize(size string) *Paragraph {
	p.ensureParagraphProperties()
	p.ensureRunProperties()
	p.Properties.RunProperties.Size = &Size{Val: size}
	return p
}

// Indent sets paragraph indentation
// left: left indentation in twips (1440 = 1 inch, 720 = 0.5 inch)
// firstLine: first line indentation in twips (optional, use 0 for none)
// hanging: hanging indentation in twips (optional, use 0 for none)
//
// Note: You cannot specify both firstLine and hanging indents simultaneously.
// Valid range: -31680 to 31680 twips (-22 to 22 inches)
func (p *Paragraph) Indent(left, firstLine, hanging int) *Paragraph {
	p.ensureParagraphProperties()
	ind := &Ind{
		Left: left,
	}
	if firstLine > 0 {
		ind.FirstLine = firstLine
	}
	if hanging > 0 {
		ind.Hanging = hanging
	}
	p.Properties.Ind = ind
	return p
}

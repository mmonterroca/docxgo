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

package xml

import "encoding/xml"

// Styles represents the styles.xml document.
type Styles struct {
	XMLName      xml.Name      `xml:"w:styles"`
	Xmlns        string        `xml:"xmlns:w,attr"`
	LatentStyles *LatentStyles `xml:"w:latentStyles,omitempty"`
	DocDefaults  *DocDefaults  `xml:"w:docDefaults,omitempty"`
	Styles       []*Style      `xml:"w:style"`
}

// DocDefaults represents w:docDefaults element.
type DocDefaults struct {
	XMLName      xml.Name           `xml:"w:docDefaults"`
	RunDefaults  *RunDefaults       `xml:"w:rPrDefault,omitempty"`
	ParaDefaults *ParagraphDefaults `xml:"w:pPrDefault,omitempty"`
}

// RunDefaults represents default run properties.
type RunDefaults struct {
	XMLName    xml.Name       `xml:"w:rPrDefault"`
	Properties *RunProperties `xml:"w:rPr,omitempty"`
}

// ParagraphDefaults represents default paragraph properties.
type ParagraphDefaults struct {
	XMLName    xml.Name                  `xml:"w:pPrDefault"`
	Properties *StyleParagraphProperties `xml:"w:pPr,omitempty"`
}

// Style represents w:style element.
type Style struct {
	XMLName     xml.Name                  `xml:"w:style"`
	Type        string                    `xml:"w:type,attr"` // paragraph, character, table, numbering
	StyleID     string                    `xml:"w:styleId,attr"`
	Default     *bool                     `xml:"w:default,attr,omitempty"`
	CustomStyle *bool                     `xml:"w:customStyle,attr,omitempty"`
	Name        *StyleName                `xml:"w:name,omitempty"`
	BasedOn     *BasedOn                  `xml:"w:basedOn,omitempty"`
	Next        *Next                     `xml:"w:next,omitempty"`
	Link        *Link                     `xml:"w:link,omitempty"`
	UIPriority  *UIPriority               `xml:"w:uiPriority,omitempty"`
	QFormat     *struct{}                 `xml:"w:qFormat,omitempty"`
	ParaProps   *StyleParagraphProperties `xml:"w:pPr,omitempty"` // Must come before rPr per OOXML spec
	RunProps    *RunProperties            `xml:"w:rPr,omitempty"`
}

// StyleName represents w:name element.
type StyleName struct {
	XMLName xml.Name `xml:"w:name"`
	Val     string   `xml:"w:val,attr"`
}

// BasedOn represents w:basedOn element.
type BasedOn struct {
	XMLName xml.Name `xml:"w:basedOn"`
	Val     string   `xml:"w:val,attr"`
}

// Next represents w:next element (next paragraph style).
type Next struct {
	XMLName xml.Name `xml:"w:next"`
	Val     string   `xml:"w:val,attr"`
}

// Link represents w:link element (linked character style).
type Link struct {
	XMLName xml.Name `xml:"w:link"`
	Val     string   `xml:"w:val,attr"`
}

// UIPriority represents w:uiPriority element.
type UIPriority struct {
	XMLName xml.Name `xml:"w:uiPriority"`
	Val     int      `xml:"w:val,attr"`
}

// StyleParagraphProperties represents w:pPr element (paragraph properties in styles).
type StyleParagraphProperties struct {
	XMLName         xml.Name          `xml:"w:pPr"`
	KeepNext        *struct{}         `xml:"w:keepNext,omitempty"`
	KeepLines       *struct{}         `xml:"w:keepLines,omitempty"`
	PageBreakBefore *struct{}         `xml:"w:pageBreakBefore,omitempty"`
	Spacing         *StyleSpacing     `xml:"w:spacing,omitempty"`
	Indentation     *StyleIndentation `xml:"w:ind,omitempty"`
	Alignment       *Alignment        `xml:"w:jc,omitempty"`
	OutlineLevel    *OutlineLevel     `xml:"w:outlineLvl,omitempty"`
}

// Alignment represents w:jc element (justification/alignment).
type Alignment struct {
	XMLName xml.Name `xml:"w:jc"`
	Val     string   `xml:"w:val,attr"` // left, center, right, both (justified), distribute
}

// StyleSpacing represents w:spacing element (paragraph spacing in styles).
type StyleSpacing struct {
	XMLName  xml.Name `xml:"w:spacing"`
	Before   *int     `xml:"w:before,attr,omitempty"`   // Space before in twips
	After    *int     `xml:"w:after,attr,omitempty"`    // Space after in twips
	Line     *int     `xml:"w:line,attr,omitempty"`     // Line spacing
	LineRule string   `xml:"w:lineRule,attr,omitempty"` // auto, exact, atLeast
}

// StyleIndentation represents w:ind element (paragraph indentation in styles).
type StyleIndentation struct {
	XMLName   xml.Name `xml:"w:ind"`
	Left      *int     `xml:"w:left,attr,omitempty"`      // Left indent in twips
	Right     *int     `xml:"w:right,attr,omitempty"`     // Right indent in twips
	FirstLine *int     `xml:"w:firstLine,attr,omitempty"` // First line indent
	Hanging   *int     `xml:"w:hanging,attr,omitempty"`   // Hanging indent
}

// OutlineLevel represents w:outlineLvl element (outline level 0-9).
type OutlineLevel struct {
	XMLName xml.Name `xml:"w:outlineLvl"`
	Val     int      `xml:"w:val,attr"`
}

// LatentStyles represents w:latentStyles element (latent style metadata).
type LatentStyles struct {
	XMLName           xml.Name                `xml:"w:latentStyles"`
	DefLockedState    string                  `xml:"w:defLockedState,attr,omitempty"`
	DefUIPriority     string                  `xml:"w:defUIPriority,attr,omitempty"`
	DefSemiHidden     string                  `xml:"w:defSemiHidden,attr,omitempty"`
	DefUnhideWhenUsed string                  `xml:"w:defUnhideWhenUsed,attr,omitempty"`
	DefQFormat        string                  `xml:"w:defQFormat,attr,omitempty"`
	Count             string                  `xml:"w:count,attr,omitempty"`
	Exceptions        []*LatentStyleException `xml:"w:lsdException,omitempty"`
}

// LatentStyleException represents w:lsdException element (latent style exception definition).
type LatentStyleException struct {
	XMLName        xml.Name `xml:"w:lsdException"`
	Name           string   `xml:"w:name,attr"`
	UIPriority     string   `xml:"w:uiPriority,attr,omitempty"`
	QFormat        string   `xml:"w:qFormat,attr,omitempty"`
	SemiHidden     string   `xml:"w:semiHidden,attr,omitempty"`
	UnhideWhenUsed string   `xml:"w:unhideWhenUsed,attr,omitempty"`
	Locked         string   `xml:"w:locked,attr,omitempty"`
}

// NewStyles creates a new styles document.
func NewStyles() *Styles {
	return &Styles{
		Xmlns:  "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
		Styles: make([]*Style, 0),
	}
}

// AddStyle adds a style to the styles collection.
func (s *Styles) AddStyle(style *Style) {
	s.Styles = append(s.Styles, style)
}

// NewParagraphStyle creates a new paragraph style.
func NewParagraphStyle(styleID, name string, isDefault bool) *Style {
	style := &Style{
		Type:    "paragraph",
		StyleID: styleID,
		Name:    &StyleName{Val: name},
	}
	if isDefault {
		defaultVal := true
		style.Default = &defaultVal
	}
	return style
}

// NewCharacterStyle creates a new character style.
func NewCharacterStyle(styleID, name string, isDefault bool) *Style {
	style := &Style{
		Type:    "character",
		StyleID: styleID,
		Name:    &StyleName{Val: name},
	}
	if isDefault {
		defaultVal := true
		style.Default = &defaultVal
	}
	return style
}

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

package xml

import "encoding/xml"

// Styles represents the styles.xml document.
type Styles struct {
	XMLName  xml.Name `xml:"w:styles"`
	Xmlns    string   `xml:"xmlns:w,attr"`
	DocDefaults *DocDefaults `xml:"w:docDefaults,omitempty"`
	Styles   []*Style `xml:"w:style"`
}

// DocDefaults represents w:docDefaults element.
type DocDefaults struct {
	XMLName         xml.Name               `xml:"w:docDefaults"`
	RunDefaults     *RunDefaults           `xml:"w:rPrDefault,omitempty"`
	ParaDefaults    *ParagraphDefaults     `xml:"w:pPrDefault,omitempty"`
}

// RunDefaults represents default run properties.
type RunDefaults struct {
	XMLName    xml.Name       `xml:"w:rPrDefault"`
	Properties *RunProperties `xml:"w:rPr,omitempty"`
}

// ParagraphDefaults represents default paragraph properties.
type ParagraphDefaults struct {
	XMLName    xml.Name              `xml:"w:pPrDefault"`
	Properties *ParagraphProperties  `xml:"w:pPr,omitempty"`
}

// Style represents w:style element.
type Style struct {
	XMLName       xml.Name              `xml:"w:style"`
	Type          string                `xml:"w:type,attr"`           // paragraph, character, table, numbering
	StyleID       string                `xml:"w:styleId,attr"`
	Default       *bool                 `xml:"w:default,attr,omitempty"`
	CustomStyle   *bool                 `xml:"w:customStyle,attr,omitempty"`
	Name          *StyleName            `xml:"w:name,omitempty"`
	BasedOn       *BasedOn              `xml:"w:basedOn,omitempty"`
	Next          *Next                 `xml:"w:next,omitempty"`
	UIPriority    *UIPriority           `xml:"w:uiPriority,omitempty"`
	QFormat       *struct{}             `xml:"w:qFormat,omitempty"`
	RunProps      *RunProperties        `xml:"w:rPr,omitempty"`
	ParaProps     *ParagraphProperties  `xml:"w:pPr,omitempty"`
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

// UIPriority represents w:uiPriority element.
type UIPriority struct {
	XMLName xml.Name `xml:"w:uiPriority"`
	Val     int      `xml:"w:val,attr"`
}

// ParagraphProperties represents w:pPr element (paragraph properties in styles).
type ParagraphProperties struct {
	XMLName         xml.Name         `xml:"w:pPr"`
	Alignment       *Alignment       `xml:"w:jc,omitempty"`
	SpacingBefore   *Spacing         `xml:"w:spacing,omitempty"`
	Indentation     *Indentation     `xml:"w:ind,omitempty"`
	KeepNext        *struct{}        `xml:"w:keepNext,omitempty"`
	KeepLines       *struct{}        `xml:"w:keepLines,omitempty"`
	PageBreakBefore *struct{}        `xml:"w:pageBreakBefore,omitempty"`
	OutlineLevel    *OutlineLevel    `xml:"w:outlineLvl,omitempty"`
}

// Alignment represents w:jc element (justification/alignment).
type Alignment struct {
	XMLName xml.Name `xml:"w:jc"`
	Val     string   `xml:"w:val,attr"` // left, center, right, both (justified), distribute
}

// Spacing represents w:spacing element (paragraph spacing).
type Spacing struct {
	XMLName     xml.Name `xml:"w:spacing"`
	Before      *int     `xml:"w:before,attr,omitempty"`      // Space before in twips
	After       *int     `xml:"w:after,attr,omitempty"`       // Space after in twips
	Line        *int     `xml:"w:line,attr,omitempty"`        // Line spacing
	LineRule    string   `xml:"w:lineRule,attr,omitempty"`    // auto, exact, atLeast
}

// Indentation represents w:ind element (paragraph indentation).
type Indentation struct {
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

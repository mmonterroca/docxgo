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

// Color allows to set run color
func (r *Run) Color(color string) *Run {
	if r.RunProperties == nil {
		r.RunProperties = &RunProperties{}
	}
	r.RunProperties.Color = &Color{
		Val: color,
	}
	return r
}

// Size allows to set run size
func (r *Run) Size(size string) *Run {
	if r.RunProperties == nil {
		r.RunProperties = &RunProperties{}
	}
	r.RunProperties.Size = &Size{
		Val: size,
	}
	return r
}

// SizeCs allows to set run sizecs
func (r *Run) SizeCs(size string) *Run {
	if r.RunProperties == nil {
		r.RunProperties = &RunProperties{}
	}
	r.RunProperties.SizeCs = &SizeCs{
		Val: size,
	}
	return r
}

// Shade allows to set run shade
func (r *Run) Shade(val, color, fill string) *Run {
	if r.RunProperties == nil {
		r.RunProperties = &RunProperties{}
	}
	r.RunProperties.Shade = &Shade{
		Val:   val,
		Color: color,
		Fill:  fill,
	}
	return r
}

// Spacing allows to set run spacing
func (r *Run) Spacing(line int) *Run {
	if r.RunProperties == nil {
		r.RunProperties = &RunProperties{}
	}
	r.RunProperties.Spacing = &Spacing{
		Line: line,
	}
	return r
}

// Bold ...
func (r *Run) Bold() *Run {
	if r.RunProperties == nil {
		r.RunProperties = &RunProperties{}
	}
	r.RunProperties.Bold = &Bold{}
	return r
}

// Italic ...
func (r *Run) Italic() *Run {
	if r.RunProperties == nil {
		r.RunProperties = &RunProperties{}
	}
	r.RunProperties.Italic = &Italic{}
	return r
}

// Underline has several possible values including
//
//	none: Specifies that no underline should be applied.
//	single: Specifies a single underline.
//	words: Specifies that only words within the text should be underlined.
//	double: Specifies a double underline.
//	thick: Specifies a thick underline.
//	dotted: Specifies a dotted underline.
//	dash: Specifies a dash underline.
//	dotDash: Specifies an alternating dot-dash underline.
//	dotDotDash: Specifies an alternating dot-dot-dash underline.
//	wave: Specifies a wavy underline.
//	dashLong: Specifies a long dash underline.
//	wavyDouble: Specifies a double wavy underline.
func (r *Run) Underline(val string) *Run {
	if r.RunProperties == nil {
		r.RunProperties = &RunProperties{}
	}
	r.RunProperties.Underline = &Underline{Val: val}
	return r
}

// Highlight ...
func (r *Run) Highlight(val string) *Run {
	if r.RunProperties == nil {
		r.RunProperties = &RunProperties{}
	}
	r.RunProperties.Highlight = &Highlight{Val: val}
	return r
}

// Strike ...
func (r *Run) Strike(val bool) *Run {
	if r.RunProperties == nil {
		r.RunProperties = &RunProperties{}
	}
	trueFalseStr := "false"
	if val {
		trueFalseStr = "true"
	}
	r.RunProperties.Strike = &Strike{Val: trueFalseStr}
	return r
}

// AddTab add a tab in front of the run
func (r *Run) AddTab() *Run {
	r.Children = append(r.Children, &Tab{})
	return r
}

// Font sets the font of the run
func (r *Run) Font(ascii, eastAsia, hansi, hint string) *Run {
	if r.RunProperties == nil {
		r.RunProperties = &RunProperties{}
	}
	r.RunProperties.Fonts = &RunFonts{
		ASCII:    ascii,
		EastAsia: eastAsia,
		HAnsi:    hansi,
		Hint:     hint,
	}
	return r
}

// AddText adds text directly to a run
func (r *Run) AddText(text string) *Run {
	// NOTE: Do NOT initialize RunProperties if empty
	// An empty <w:rPr/> element can cause Word to reject the document
	// RunProperties should only be created when actually setting properties
	r.Children = append(r.Children, &Text{Text: text})
	return r
}

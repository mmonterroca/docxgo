/*
   Copyright (c) 2020 gingfrederik
   Copyright (c) 2021 Gonzalo Fernandez-Victorio
   Copyright (c) 2021 Basement Crowd Ltd (https://www.basementcrowd.com)
   Copyright (c) 2023 Fumiama Minamoto (源文雨)
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

package docx

import "fmt"

// ValidateIndent validates indent parameters
func ValidateIndent(left, firstLine, hanging int) error {
	// Validate left indent
	if left < MinIndentTwips || left > MaxIndentTwips {
		return newValidationError("left", left,
			fmt.Sprintf("must be between %d and %d twips", MinIndentTwips, MaxIndentTwips))
	}

	// Validate firstLine indent
	if firstLine < 0 || firstLine > MaxIndentTwips {
		return newValidationError("firstLine", firstLine,
			fmt.Sprintf("must be between 0 and %d twips", MaxIndentTwips))
	}

	// Validate hanging indent
	if hanging < 0 || hanging > MaxIndentTwips {
		return newValidationError("hanging", hanging,
			fmt.Sprintf("must be between 0 and %d twips", MaxIndentTwips))
	}

	// Cannot specify both firstLine and hanging
	if firstLine > 0 && hanging > 0 {
		return ErrConflictingIndent
	}

	return nil
}

// ValidateJustification validates justification values
func ValidateJustification(val string) error {
	validJustifications := map[string]bool{
		AlignLeft:       true,
		AlignCenter:     true,
		AlignRight:      true,
		AlignBoth:       true,
		AlignDistribute: true,
	}

	if !validJustifications[val] {
		return newValidationError("justification", val,
			"must be one of: start, center, end, both, distribute")
	}

	return nil
}

// ValidateUnderline validates underline style values
func ValidateUnderline(val string) error {
	validUnderlines := map[string]bool{
		UnderlineNone:       true,
		UnderlineSingle:     true,
		UnderlineWords:      true,
		UnderlineDouble:     true,
		UnderlineThick:      true,
		UnderlineDotted:     true,
		UnderlineDash:       true,
		UnderlineDotDash:    true,
		UnderlineDotDotDash: true,
		UnderlineWave:       true,
		UnderlineDashLong:   true,
		UnderlineWavyDouble: true,
	}

	if !validUnderlines[val] {
		return newValidationError("underline", val,
			"must be a valid underline style (see constants)")
	}

	return nil
}

// ValidateColor validates hex color values
func ValidateColor(color string) error {
	if len(color) == 0 {
		return newValidationError("color", color, "cannot be empty")
	}

	// Basic validation for hex colors (should start with # and be 7 chars)
	if color[0] == '#' && len(color) != 7 {
		return newValidationError("color", color, "hex color must be in format #RRGGBB")
	}

	return nil
}

// ValidateSize validates font size values
func ValidateSize(size string) error {
	if len(size) == 0 {
		return newValidationError("size", size, "cannot be empty")
	}

	// Size should be a numeric string (half-points)
	// We don't validate the actual number here to allow flexibility
	return nil
}

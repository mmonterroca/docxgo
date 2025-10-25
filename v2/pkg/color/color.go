package color
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

// Package color provides color utilities for go-docx v2.
package color

import (
	"fmt"
	"strconv"

	"github.com/SlideLang/go-docx/v2/domain"
	"github.com/SlideLang/go-docx/v2/pkg/errors"
)

// ToHex converts a Color to a hex string (e.g., "FF0000" for red).
func ToHex(c domain.Color) string {
	return fmt.Sprintf("%02X%02X%02X", c.R, c.G, c.B)
}

// FromHex creates a Color from a hex string.
// Accepts formats: "RGB", "RRGGBB", "#RGB", "#RRGGBB"
func FromHex(hex string) (domain.Color, error) {
	// Remove # if present
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}

	var r, g, b uint8

	switch len(hex) {
	case 3:
		// Short form: RGB -> RRGGBB
		rv, err := strconv.ParseUint(string(hex[0]), 16, 8)
		if err != nil {
			return domain.Color{}, errors.InvalidArgument("FromHex", "hex", hex, "invalid red component")
		}
		gv, err := strconv.ParseUint(string(hex[1]), 16, 8)
		if err != nil {
			return domain.Color{}, errors.InvalidArgument("FromHex", "hex", hex, "invalid green component")
		}
		bv, err := strconv.ParseUint(string(hex[2]), 16, 8)
		if err != nil {
			return domain.Color{}, errors.InvalidArgument("FromHex", "hex", hex, "invalid blue component")
		}
		r = uint8(rv*16 + rv)
		g = uint8(gv*16 + gv)
		b = uint8(bv*16 + bv)

	case 6:
		// Full form: RRGGBB
		rv, err := strconv.ParseUint(hex[0:2], 16, 8)
		if err != nil {
			return domain.Color{}, errors.InvalidArgument("FromHex", "hex", hex, "invalid red component")
		}
		gv, err := strconv.ParseUint(hex[2:4], 16, 8)
		if err != nil {
			return domain.Color{}, errors.InvalidArgument("FromHex", "hex", hex, "invalid green component")
		}
		bv, err := strconv.ParseUint(hex[4:6], 16, 8)
		if err != nil {
			return domain.Color{}, errors.InvalidArgument("FromHex", "hex", hex, "invalid blue component")
		}
		r = uint8(rv)
		g = uint8(gv)
		b = uint8(bv)

	default:
		return domain.Color{}, errors.InvalidArgument("FromHex", "hex", hex,
			"hex color must be 3 or 6 characters (optionally prefixed with #)")
	}

	return domain.Color{R: r, G: g, B: b}, nil
}

// Validate checks if a color is valid (all components in range 0-255).
func Validate(c domain.Color) error {
	// uint8 automatically ensures 0-255 range, so this is always valid
	return nil
}

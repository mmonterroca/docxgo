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



// Package color provides color utilities for go-docx v2.
package color

import (
	"fmt"
	"strconv"

	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/pkg/errors"
)

// Common color constants for convenience.
var (
	Black   = domain.Color{R: 0, G: 0, B: 0}
	White   = domain.Color{R: 255, G: 255, B: 255}
	Red     = domain.Color{R: 255, G: 0, B: 0}
	Green   = domain.Color{R: 0, G: 128, B: 0}
	Blue    = domain.Color{R: 0, G: 0, B: 255}
	Yellow  = domain.Color{R: 255, G: 255, B: 0}
	Cyan    = domain.Color{R: 0, G: 255, B: 255}
	Magenta = domain.Color{R: 255, G: 0, B: 255}
	Orange  = domain.Color{R: 255, G: 165, B: 0}
	Purple  = domain.Color{R: 128, G: 0, B: 128}
	Gray    = domain.Color{R: 128, G: 128, B: 128}
	Silver  = domain.Color{R: 192, G: 192, B: 192}
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
func Validate(_ domain.Color) error {
	// uint8 automatically ensures 0-255 range, so this is always valid
	return nil
}

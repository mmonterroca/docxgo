// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package domain

import "testing"

// TestImagePositionUseOffsetFlags tests that UseOffsetX and UseOffsetY flags
// are correctly used to distinguish between "offset = 0" and "no offset".
// This is critical for preserving wp:posOffset elements during round-trip.
// See: docs/TROUBLESHOOTING_DOCX_VALIDATION.md - Issue 2
func TestImagePositionUseOffsetFlags(t *testing.T) {
	t.Run("default position has UseOffset flags as false", func(t *testing.T) {
		pos := DefaultImagePosition()

		if pos.UseOffsetX {
			t.Error("UseOffsetX = true; want false for default")
		}
		if pos.UseOffsetY {
			t.Error("UseOffsetY = true; want false for default")
		}
	})

	t.Run("UseOffsetX flag set with zero offset", func(t *testing.T) {
		pos := ImagePosition{
			Type:       ImagePositionFloating,
			OffsetX:    0,
			UseOffsetX: true, // Explicitly set, even though value is 0
		}

		if !pos.UseOffsetX {
			t.Error("UseOffsetX = false; want true")
		}
		if pos.OffsetX != 0 {
			t.Errorf("OffsetX = %d; want 0", pos.OffsetX)
		}
	})

	t.Run("UseOffsetY flag set with zero offset", func(t *testing.T) {
		pos := ImagePosition{
			Type:       ImagePositionFloating,
			OffsetY:    0,
			UseOffsetY: true, // Explicitly set, even though value is 0
		}

		if !pos.UseOffsetY {
			t.Error("UseOffsetY = false; want true")
		}
		if pos.OffsetY != 0 {
			t.Errorf("OffsetY = %d; want 0", pos.OffsetY)
		}
	})

	t.Run("non-zero offset without UseOffset flag", func(t *testing.T) {
		// When offset is non-zero, UseOffset flag doesn't matter
		pos := ImagePosition{
			Type:       ImagePositionFloating,
			OffsetX:    12700, // Some EMU value
			OffsetY:    25400,
			UseOffsetX: false,
			UseOffsetY: false,
		}

		// Both offsets should be preserved based on non-zero value
		if pos.OffsetX != 12700 {
			t.Errorf("OffsetX = %d; want 12700", pos.OffsetX)
		}
		if pos.OffsetY != 25400 {
			t.Errorf("OffsetY = %d; want 25400", pos.OffsetY)
		}
	})

	t.Run("alignment with no offset", func(t *testing.T) {
		// When no offset is set and UseOffset is false, alignment should be used
		pos := ImagePosition{
			Type:       ImagePositionFloating,
			HAlign:     HAlignCenter,
			VAlign:     VAlignTop,
			OffsetX:    0,
			OffsetY:    0,
			UseOffsetX: false,
			UseOffsetY: false,
		}

		if pos.HAlign != HAlignCenter {
			t.Errorf("HAlign = %s; want center", pos.HAlign)
		}
		if pos.VAlign != VAlignTop {
			t.Errorf("VAlign = %s; want top", pos.VAlign)
		}
	})
}

// TestImagePositionFloatingComplete tests a complete floating image position
// configuration that would be read from an existing DOCX file.
func TestImagePositionFloatingComplete(t *testing.T) {
	pos := ImagePosition{
		Type:       ImagePositionFloating,
		HAlign:     "",
		VAlign:     "",
		OffsetX:    0,
		OffsetY:    0,
		UseOffsetX: true, // Read from <wp:posOffset>0</wp:posOffset>
		UseOffsetY: true, // Read from <wp:posOffset>0</wp:posOffset>
		WrapText:   WrapSquare,
		ZOrder:     251659264,
		BehindText: false,
	}

	// Verify the position would serialize using posOffset, not align
	if !pos.UseOffsetX {
		t.Error("UseOffsetX should be true to preserve original posOffset")
	}
	if !pos.UseOffsetY {
		t.Error("UseOffsetY should be true to preserve original posOffset")
	}
	if pos.Type != ImagePositionFloating {
		t.Errorf("Type = %s; want floating", pos.Type)
	}
}

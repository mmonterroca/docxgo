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

// Capacity constants for pre-allocation
const (
	// DefaultParagraphCapacity is the default capacity for paragraph children
	DefaultParagraphCapacity = 64

	// DefaultFileMapCapacity is the default capacity for file maps
	DefaultFileMapCapacity = 64

	// DefaultMediaIndexCapacity is the default capacity for media index maps
	DefaultMediaIndexCapacity = 64

	// DefaultSlowIDCapacity is the default capacity for slow ID maps
	DefaultSlowIDCapacity = 64
)

// Measurement constants (twips)
// A twip is 1/20 of a point, or 1/1440 of an inch
const (
	// TwipsPerInch represents the number of twips in one inch
	TwipsPerInch = 1440

	// TwipsPerHalfInch represents the number of twips in half an inch
	TwipsPerHalfInch = 720

	// TwipsPerQuarterInch represents the number of twips in a quarter inch
	TwipsPerQuarterInch = 360

	// TwipsPerPoint represents the number of twips in one point
	TwipsPerPoint = 20

	// MaxIndentTwips is the maximum allowed indentation (22 inches)
	MaxIndentTwips = 31680

	// MinIndentTwips is the minimum allowed indentation (-22 inches)
	MinIndentTwips = -31680
)

// ID type constants for internal identification
const (
	// IDTypeImage is the identifier type for images
	IDTypeImage = "image"

	// IDTypeDrawing is the identifier type for drawings
	IDTypeDrawing = "drawing"

	// IDTypeBookmark is the identifier type for bookmarks
	IDTypeBookmark = "bookmark"

	// IDTypeShape is the identifier type for shapes
	IDTypeShape = "shape"

	// IDTypeCanvas is the identifier type for canvas
	IDTypeCanvas = "canvas"

	// IDTypeTable is the identifier type for tables
	IDTypeTable = "table"
)

// Alignment constants for paragraph justification
const (
	// AlignLeft aligns content to the left
	AlignLeft = "start"

	// AlignCenter centers content
	AlignCenter = "center"

	// AlignRight aligns content to the right
	AlignRight = "end"

	// AlignBoth justifies content (both left and right alignment)
	AlignBoth = "both"

	// AlignDistribute distributes content evenly
	AlignDistribute = "distribute"
)

// Underline style constants
const (
	// UnderlineNone means no underline
	UnderlineNone = "none"

	// UnderlineSingle is a single underline
	UnderlineSingle = "single"

	// UnderlineWords underlines only words
	UnderlineWords = "words"

	// UnderlineDouble is a double underline
	UnderlineDouble = "double"

	// UnderlineThick is a thick underline
	UnderlineThick = "thick"

	// UnderlineDotted is a dotted underline
	UnderlineDotted = "dotted"

	// UnderlineDash is a dashed underline
	UnderlineDash = "dash"

	// UnderlineDotDash is an alternating dot-dash underline
	UnderlineDotDash = "dotDash"

	// UnderlineDotDotDash is an alternating dot-dot-dash underline
	UnderlineDotDotDash = "dotDotDash"

	// UnderlineWave is a wavy underline
	UnderlineWave = "wave"

	// UnderlineDashLong is a long dash underline
	UnderlineDashLong = "dashLong"

	// UnderlineWavyDouble is a double wavy underline
	UnderlineWavyDouble = "wavyDouble"
)

// Page break constants
const (
	// PageBreakType is the type attribute for page breaks
	PageBreakType = "page"
)

// Default relationship IDs
const (
	// DefaultRelationshipIDStart is the starting ID for relationships
	DefaultRelationshipIDStart = 3
)

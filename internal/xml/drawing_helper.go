// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package xml

import "github.com/mmonterroca/docxgo/v2/domain"

// NewInlineDrawing creates an inline drawing (flows with text).
func NewInlineDrawing(img domain.Image, drawingID int) *Drawing {
	size := img.Size()

	return &Drawing{
		Inline: &Inline{
			DistT: 0,
			DistB: 0,
			DistL: 0,
			DistR: 0,
			Extent: &Extent{
				Cx: size.WidthEMU,
				Cy: size.HeightEMU,
			},
			EffectExtent: &EffectExtent{
				L: 0,
				T: 0,
				R: 0,
				B: 0,
			},
			DocPr: &DocPr{
				ID:    drawingID,
				Name:  "Picture " + img.ID(),
				Descr: img.Description(),
			},
			Graphic: newGraphic(img, size),
		},
	}
}

// NewFloatingDrawing creates a floating drawing (absolute positioning).
func NewFloatingDrawing(img domain.Image, drawingID int) *Drawing {
	size := img.Size()
	pos := img.Position()

	// Convert position to anchor
	anchor := &Anchor{
		DistT:          114300, // Default distances (0.125 inch)
		DistB:          114300,
		DistL:          114300,
		DistR:          114300,
		SimplePosAttr:  false,
		RelativeHeight: pos.ZOrder,
		BehindDoc:      pos.BehindText,
		Locked:         false,
		LayoutInCell:   true,
		AllowOverlap:   true,
		SimplePos: &SimplePos{
			X: 0,
			Y: 0,
		},
		Extent: &Extent{
			Cx: size.WidthEMU,
			Cy: size.HeightEMU,
		},
		EffectExtent: &EffectExtent{
			L: 0,
			T: 0,
			R: 0,
			B: 0,
		},
		DocPr: &DocPr{
			ID:    drawingID,
			Name:  "Picture " + img.ID(),
			Descr: img.Description(),
		},
		Graphic: newGraphic(img, size),
	}

	// Set horizontal position
	anchor.PositionH = &PositionH{
		RelativeFrom: convertHAlign(pos.HAlign),
	}
	// Use offset if explicitly set (even if 0), otherwise use alignment if provided
	if pos.UseOffsetX || pos.OffsetX != 0 {
		offset := pos.OffsetX
		anchor.PositionH.PosOffset = &offset
	} else if pos.HAlign != "" {
		// Only set Align if we have a non-empty value
		align := string(pos.HAlign)
		anchor.PositionH.Align = &align
	}

	// Set vertical position
	anchor.PositionV = &PositionV{
		RelativeFrom: convertVAlign(pos.VAlign),
	}
	// Use offset if explicitly set (even if 0), otherwise use alignment if provided
	if pos.UseOffsetY || pos.OffsetY != 0 {
		offset := pos.OffsetY
		anchor.PositionV.PosOffset = &offset
	} else if pos.VAlign != "" {
		// Only set Align if we have a non-empty value
		align := string(pos.VAlign)
		anchor.PositionV.Align = &align
	}

	// Set wrap type
	if pos.WrapText != domain.WrapNone {
		anchor.WrapType = &WrapType{
			WrapText: "bothSides",
		}
	}

	return &Drawing{
		Anchor: anchor,
	}
}

// newGraphic creates the graphic content for an image.
func newGraphic(img domain.Image, size domain.ImageSize) *Graphic {
	return &Graphic{
		Xmlns: "http://schemas.openxmlformats.org/drawingml/2006/main",
		GraphicData: &GraphicData{
			URI: "http://schemas.openxmlformats.org/drawingml/2006/picture",
			Pic: &Pic{
				Xmlns: "http://schemas.openxmlformats.org/drawingml/2006/picture",
				NvPicPr: &NvPicPr{
					CNvPr: &CNvPr{
						ID:    0,
						Name:  "Picture " + img.ID(),
						Descr: img.Description(),
					},
					CNvPicPr: &CNvPicPr{
						PicLocks: &PicLocks{
							NoChangeAspect: true,
						},
					},
				},
				BlipFill: &BlipFill{
					Blip: &Blip{
						Xmlns: "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
						Embed: img.RelationshipID(),
					},
					Stretch: &Stretch{
						FillRect: &FillRect{},
					},
				},
				SpPr: &SpPr{
					Xfrm: &Xfrm{
						Off: &Off{X: 0, Y: 0},
						Ext: &Ext{
							Cx: size.WidthEMU,
							Cy: size.HeightEMU,
						},
					},
					PrstGeom: &PrstGeom{
						Prst:  "rect",
						AvLst: &AvLst{},
					},
				},
			},
		},
	}
}

// convertHAlign converts domain horizontal alignment to XML relative from.
func convertHAlign(align domain.HorizontalAlign) string {
	switch align {
	case domain.HAlignLeft:
		return "column"
	case domain.HAlignCenter:
		return "column"
	case domain.HAlignRight:
		return "column"
	case domain.HAlignInside:
		return "margin"
	case domain.HAlignOutside:
		return "margin"
	default:
		return "column"
	}
}

// convertVAlign converts domain vertical alignment to XML relative from.
func convertVAlign(align domain.VerticalAlign) string {
	switch align {
	case domain.VAlignTop:
		return "paragraph"
	case domain.VAlignCenter:
		return "paragraph"
	case domain.VAlignBottom:
		return "paragraph"
	case domain.VAlignInside:
		return "margin"
	case domain.VAlignOutside:
		return "margin"
	default:
		return "paragraph"
	}
}

/*
MIT License

Copyright (c) 2025 Misael Montero

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

import "github.com/mmonterroca/docxgo/domain"

// NewInlineDrawing creates an inline drawing (flows with text).
func NewInlineDrawing(img domain.Image, drawingID int) *Drawing {
	size := img.Size()
	
	return &Drawing{
		Inline: &Inline{
			DistT:  0,
			DistB:  0,
			DistL:  0,
			DistR:  0,
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
	if pos.OffsetX != 0 {
		offset := pos.OffsetX
		anchor.PositionH.PosOffset = &offset
	} else {
		align := string(pos.HAlign)
		anchor.PositionH.Align = &align
	}

	// Set vertical position
	anchor.PositionV = &PositionV{
		RelativeFrom: convertVAlign(pos.VAlign),
	}
	if pos.OffsetY != 0 {
		offset := pos.OffsetY
		anchor.PositionV.PosOffset = &offset
	} else {
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

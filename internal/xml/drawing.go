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

import "encoding/xml"

// Drawing represents a w:drawing element (container for images/shapes).
type Drawing struct {
	XMLName xml.Name `xml:"w:drawing"`
	Inline  *Inline  `xml:"wp:inline,omitempty"`
	Anchor  *Anchor  `xml:"wp:anchor,omitempty"`
}

// Inline represents an inline drawing (flows with text).
type Inline struct {
	XMLName           xml.Name          `xml:"wp:inline"`
	DistT             int               `xml:"distT,attr,omitempty"`   // Distance from text (top)
	DistB             int               `xml:"distB,attr,omitempty"`   // Distance from text (bottom)
	DistL             int               `xml:"distL,attr,omitempty"`   // Distance from text (left)
	DistR             int               `xml:"distR,attr,omitempty"`   // Distance from text (right)
	Extent            *Extent           `xml:"wp:extent"`              // Drawing size
	EffectExtent      *EffectExtent     `xml:"wp:effectExtent,omitempty"`
	DocPr             *DocPr            `xml:"wp:docPr"`               // Non-visual properties
	CNvGraphicFramePr *CNvGraphicFramePr `xml:"wp:cNvGraphicFramePr,omitempty"`
	Graphic           *Graphic          `xml:"a:graphic"`              // Graphic content
}

// Anchor represents a floating drawing (absolute positioning).
type Anchor struct {
	XMLName            xml.Name           `xml:"wp:anchor"`
	DistT              int                `xml:"distT,attr,omitempty"`
	DistB              int                `xml:"distB,attr,omitempty"`
	DistL              int                `xml:"distL,attr,omitempty"`
	DistR              int                `xml:"distR,attr,omitempty"`
	SimplePos          *SimplePos         `xml:"wp:simplePos"`
	PositionH          *PositionH         `xml:"wp:positionH"`
	PositionV          *PositionV         `xml:"wp:positionV"`
	Extent             *Extent            `xml:"wp:extent"`
	EffectExtent       *EffectExtent      `xml:"wp:effectExtent,omitempty"`
	WrapType           *WrapType          `xml:"wp:wrapSquare,omitempty"` // or wrapTight, wrapThrough, etc.
	DocPr              *DocPr             `xml:"wp:docPr"`
	CNvGraphicFramePr  *CNvGraphicFramePr `xml:"wp:cNvGraphicFramePr,omitempty"`
	Graphic            *Graphic           `xml:"a:graphic"`
	RelativeHeight     int                `xml:"relativeHeight,attr,omitempty"`
	BehindDoc          bool               `xml:"behindDoc,attr,omitempty"`
	Locked             bool               `xml:"locked,attr,omitempty"`
	LayoutInCell       bool               `xml:"layoutInCell,attr,omitempty"`
	AllowOverlap       bool               `xml:"allowOverlap,attr,omitempty"`
}

// Extent represents the size of a drawing in EMUs.
type Extent struct {
	XMLName xml.Name `xml:"wp:extent"`
	Cx      int      `xml:"cx,attr"` // Width in EMUs
	Cy      int      `xml:"cy,attr"` // Height in EMUs
}

// EffectExtent represents additional space for effects.
type EffectExtent struct {
	XMLName xml.Name `xml:"wp:effectExtent"`
	L       int      `xml:"l,attr"` // Left
	T       int      `xml:"t,attr"` // Top
	R       int      `xml:"r,attr"` // Right
	B       int      `xml:"b,attr"` // Bottom
}

// DocPr represents non-visual drawing properties.
type DocPr struct {
	XMLName xml.Name `xml:"wp:docPr"`
	ID      int      `xml:"id,attr"`
	Name    string   `xml:"name,attr"`
	Descr   string   `xml:"descr,attr,omitempty"` // Alt text
}

// CNvGraphicFramePr represents non-visual graphic frame properties.
type CNvGraphicFramePr struct {
	XMLName           xml.Name          `xml:"wp:cNvGraphicFramePr"`
	GraphicFrameLocks *GraphicFrameLocks `xml:"a:graphicFrameLocks,omitempty"`
}

// GraphicFrameLocks represents graphic frame lock settings.
type GraphicFrameLocks struct {
	XMLName         xml.Name `xml:"a:graphicFrameLocks"`
	NoChangeAspect  bool     `xml:"noChangeAspect,attr,omitempty"`
	NoMove          bool     `xml:"noMove,attr,omitempty"`
	NoResize        bool     `xml:"noResize,attr,omitempty"`
}

// SimplePos represents simple positioning.
type SimplePos struct {
	XMLName xml.Name `xml:"wp:simplePos"`
	X       int      `xml:"x,attr"`
	Y       int      `xml:"y,attr"`
}

// PositionH represents horizontal positioning.
type PositionH struct {
	XMLName    xml.Name `xml:"wp:positionH"`
	RelativeFrom string   `xml:"relativeFrom,attr"` // page, column, character, etc.
	PosOffset  *int     `xml:"wp:posOffset,omitempty"` // Offset in EMUs
	Align      *string  `xml:"wp:align,omitempty"`     // left, center, right, etc.
}

// PositionV represents vertical positioning.
type PositionV struct {
	XMLName    xml.Name `xml:"wp:positionV"`
	RelativeFrom string   `xml:"relativeFrom,attr"` // page, paragraph, line, etc.
	PosOffset  *int     `xml:"wp:posOffset,omitempty"`
	Align      *string  `xml:"wp:align,omitempty"`
}

// WrapType represents text wrapping (multiple types available).
type WrapType struct {
	XMLName  xml.Name `xml:"wp:wrapSquare"` // or wrapTight, wrapThrough, etc.
	WrapText string   `xml:"wrapText,attr"` // bothSides, left, right, largest
}

// Graphic represents the graphic content.
type Graphic struct {
	XMLName     xml.Name    `xml:"a:graphic"`
	Xmlns       string      `xml:"xmlns:a,attr"`
	GraphicData *GraphicData `xml:"a:graphicData"`
}

// GraphicData represents the graphic data container.
type GraphicData struct {
	XMLName xml.Name `xml:"a:graphicData"`
	URI     string   `xml:"uri,attr"` // Namespace URI
	Pic     *Pic     `xml:"pic:pic,omitempty"`
}

// Pic represents a picture element.
type Pic struct {
	XMLName   xml.Name  `xml:"pic:pic"`
	Xmlns     string    `xml:"xmlns:pic,attr"`
	NvPicPr   *NvPicPr  `xml:"pic:nvPicPr"`  // Non-visual picture properties
	BlipFill  *BlipFill `xml:"pic:blipFill"` // Image fill
	SpPr      *SpPr     `xml:"pic:spPr"`     // Shape properties
}

// NvPicPr represents non-visual picture properties.
type NvPicPr struct {
	XMLName  xml.Name  `xml:"pic:nvPicPr"`
	CNvPr    *CNvPr    `xml:"pic:cNvPr"`
	CNvPicPr *CNvPicPr `xml:"pic:cNvPicPr"`
}

// CNvPr represents non-visual drawing properties.
type CNvPr struct {
	XMLName xml.Name `xml:"pic:cNvPr"`
	ID      int      `xml:"id,attr"`
	Name    string   `xml:"name,attr"`
	Descr   string   `xml:"descr,attr,omitempty"`
}

// CNvPicPr represents non-visual picture drawing properties.
type CNvPicPr struct {
	XMLName    xml.Name    `xml:"pic:cNvPicPr"`
	PicLocks   *PicLocks   `xml:"a:picLocks,omitempty"`
}

// PicLocks represents picture lock settings.
type PicLocks struct {
	XMLName        xml.Name `xml:"a:picLocks"`
	NoChangeAspect bool     `xml:"noChangeAspect,attr,omitempty"`
	NoChangeArrowheads bool `xml:"noChangeArrowheads,attr,omitempty"`
}

// BlipFill represents the image fill.
type BlipFill struct {
	XMLName xml.Name `xml:"pic:blipFill"`
	Blip    *Blip    `xml:"a:blip"`
	Stretch *Stretch `xml:"a:stretch,omitempty"`
}

// Blip represents the embedded or linked image.
type Blip struct {
	XMLName xml.Name `xml:"a:blip"`
	Xmlns   string   `xml:"xmlns:r,attr"`
	Embed   string   `xml:"r:embed,attr"` // Relationship ID
}

// Stretch represents stretch fill mode.
type Stretch struct {
	XMLName  xml.Name  `xml:"a:stretch"`
	FillRect *FillRect `xml:"a:fillRect,omitempty"`
}

// FillRect represents the fill rectangle.
type FillRect struct {
	XMLName xml.Name `xml:"a:fillRect"`
}

// SpPr represents shape properties.
type SpPr struct {
	XMLName xml.Name `xml:"pic:spPr"`
	Xfrm    *Xfrm    `xml:"a:xfrm"`
	PrstGeom *PrstGeom `xml:"a:prstGeom"`
}

// Xfrm represents 2D transform.
type Xfrm struct {
	XMLName xml.Name `xml:"a:xfrm"`
	Off     *Off     `xml:"a:off,omitempty"`
	Ext     *Ext     `xml:"a:ext"`
}

// Off represents offset (position).
type Off struct {
	XMLName xml.Name `xml:"a:off"`
	X       int      `xml:"x,attr"`
	Y       int      `xml:"y,attr"`
}

// Ext represents extent (size).
type Ext struct {
	XMLName xml.Name `xml:"a:ext"`
	Cx      int      `xml:"cx,attr"` // Width in EMUs
	Cy      int      `xml:"cy,attr"` // Height in EMUs
}

// PrstGeom represents preset geometry.
type PrstGeom struct {
	XMLName xml.Name `xml:"a:prstGeom"`
	Prst    string   `xml:"prst,attr"` // rect, ellipse, etc.
	AvLst   *AvLst   `xml:"a:avLst,omitempty"`
}

// AvLst represents adjust value list.
type AvLst struct {
	XMLName xml.Name `xml:"a:avLst"`
}

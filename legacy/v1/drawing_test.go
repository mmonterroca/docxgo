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

import (
	"testing"
)

func TestWPInlineSize(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	run, err := para.AddInlineDrawingFrom("testdata/fumiama.JPG")
	if err != nil {
		t.Fatalf("Failed to add inline drawing: %v", err)
	}

	drawing := run.Children[0].(*Drawing)
	if drawing.Inline == nil {
		t.Fatal("Expected inline drawing")
	}

	// Set new size
	newWidth := int64(1000000)
	newHeight := int64(2000000)
	drawing.Inline.Size(newWidth, newHeight)

	if drawing.Inline.Extent.CX != newWidth {
		t.Errorf("Expected Extent.CX = %d, got %d", newWidth, drawing.Inline.Extent.CX)
	}
	if drawing.Inline.Extent.CY != newHeight {
		t.Errorf("Expected Extent.CY = %d, got %d", newHeight, drawing.Inline.Extent.CY)
	}

	if drawing.Inline.Graphic.GraphicData.Pic.SpPr.Xfrm.Ext.CX != newWidth {
		t.Errorf("Expected Xfrm.Ext.CX = %d, got %d", newWidth, drawing.Inline.Graphic.GraphicData.Pic.SpPr.Xfrm.Ext.CX)
	}
	if drawing.Inline.Graphic.GraphicData.Pic.SpPr.Xfrm.Ext.CY != newHeight {
		t.Errorf("Expected Xfrm.Ext.CY = %d, got %d", newHeight, drawing.Inline.Graphic.GraphicData.Pic.SpPr.Xfrm.Ext.CY)
	}
}

func TestWPAnchorSize(t *testing.T) {
	doc := New().WithDefaultTheme()
	para := doc.AddParagraph()
	run, err := para.AddAnchorDrawingFrom("testdata/fumiama.JPG")
	if err != nil {
		t.Fatalf("Failed to add anchor drawing: %v", err)
	}

	drawing := run.Children[0].(*Drawing)
	if drawing.Anchor == nil {
		t.Fatal("Expected anchor drawing")
	}

	// Set new size
	newWidth := int64(3000000)
	newHeight := int64(4000000)
	drawing.Anchor.Size(newWidth, newHeight)

	if drawing.Anchor.Extent.CX != newWidth {
		t.Errorf("Expected Extent.CX = %d, got %d", newWidth, drawing.Anchor.Extent.CX)
	}
	if drawing.Anchor.Extent.CY != newHeight {
		t.Errorf("Expected Extent.CY = %d, got %d", newHeight, drawing.Anchor.Extent.CY)
	}

	if drawing.Anchor.Graphic.GraphicData.Pic.SpPr.Xfrm.Ext.CX != newWidth {
		t.Errorf("Expected Xfrm.Ext.CX = %d, got %d", newWidth, drawing.Anchor.Graphic.GraphicData.Pic.SpPr.Xfrm.Ext.CX)
	}
	if drawing.Anchor.Graphic.GraphicData.Pic.SpPr.Xfrm.Ext.CY != newHeight {
		t.Errorf("Expected Xfrm.Ext.CY = %d, got %d", newHeight, drawing.Anchor.Graphic.GraphicData.Pic.SpPr.Xfrm.Ext.CY)
	}
}

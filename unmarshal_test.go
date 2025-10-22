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
	"encoding/xml"
	"testing"
)

func TestWordprocessingGroupUnmarshalXML(t *testing.T) {
	xmlData := `<wpg:wgp>
		<wpg:cNvGrpSpPr>
			<a:grpSpLocks/>
		</wpg:cNvGrpSpPr>
		<wpg:grpSpPr>
			<a:xfrm>
				<a:off x="0" y="0"/>
				<a:ext cx="1000" cy="1000"/>
			</a:xfrm>
		</wpg:grpSpPr>
	</wpg:wgp>`

	var gs WordprocessingGroup
	err := xml.Unmarshal(StringToBytes(xmlData), &gs)
	if err != nil {
		t.Fatalf("Failed to unmarshal WordprocessingGroup: %v", err)
	}
}

func TestWordprocessingCanvasUnmarshalXML(t *testing.T) {
	xmlData := `<wpc:wpc>
		<wpc:bg>
			<a:solidFill>
				<a:srgbClr val="FFFFFF"/>
			</a:solidFill>
		</wpc:bg>
		<wpc:whole>
			<a:ln w="12700">
				<a:solidFill>
					<a:srgbClr val="000000"/>
				</a:solidFill>
			</a:ln>
		</wpc:whole>
	</wpc:wpc>`

	var canvas WordprocessingCanvas
	err := xml.Unmarshal(StringToBytes(xmlData), &canvas)
	if err != nil {
		t.Fatalf("Failed to unmarshal WordprocessingCanvas: %v", err)
	}
}

func TestNumPropertiesUnmarshalXML(t *testing.T) {
	xmlData := `<w:numPr>
		<w:ilvl w:val="0"/>
		<w:numId w:val="1"/>
	</w:numPr>`

	var np NumProperties
	err := xml.Unmarshal(StringToBytes(xmlData), &np)
	if err != nil {
		t.Fatalf("Failed to unmarshal NumProperties: %v", err)
	}
}

func TestWTableBorderUnmarshalXML(t *testing.T) {
	xmlData := `<w:top w:val="single" w:sz="4" w:space="0" w:color="000000"/>`

	var border WTableBorder
	err := xml.Unmarshal(StringToBytes(xmlData), &border)
	if err != nil {
		t.Fatalf("Failed to unmarshal WTableBorder: %v", err)
	}
	
	if border.Val != "single" {
		t.Errorf("Expected Val=single, got %s", border.Val)
	}
}

func TestWTableCellWidthUnmarshalXML(t *testing.T) {
	xmlData := `<w:tblW w:w="5000" w:type="dxa"/>`

	var width WTableWidth
	err := xml.Unmarshal(StringToBytes(xmlData), &width)
	if err != nil {
		t.Fatalf("Failed to unmarshal WTableWidth: %v", err)
	}

	if width.W != 5000 {
		t.Errorf("Expected W=5000, got %d", width.W)
	}
}

func TestWordprocessingShapeUnmarshalXML(t *testing.T) {
	xmlData := `<wps:wsp>
		<wps:cNvSpPr>
			<a:spLocks noChangeArrowheads="1"/>
		</wps:cNvSpPr>
		<wps:spPr>
			<a:xfrm>
				<a:off x="0" y="0"/>
				<a:ext cx="1000" cy="1000"/>
			</a:xfrm>
			<a:prstGeom prst="rect">
				<a:avLst/>
			</a:prstGeom>
		</wps:spPr>
	</wps:wsp>`

	var shape WordprocessingShape
	err := xml.Unmarshal(StringToBytes(xmlData), &shape)
	if err != nil {
		t.Fatalf("Failed to unmarshal WordprocessingShape: %v", err)
	}
}

func TestWPGGroupShapeUnmarshalXML(t *testing.T) {
	xmlData := `<wpg:grpSpPr>
		<a:xfrm>
			<a:off x="0" y="0"/>
			<a:ext cx="1000" cy="1000"/>
			<a:chOff x="0" y="0"/>
			<a:chExt cx="1000" cy="1000"/>
		</a:xfrm>
	</wpg:grpSpPr>`

	var gs WPGGroupShape
	err := xml.Unmarshal(StringToBytes(xmlData), &gs)
	if err != nil {
		t.Fatalf("Failed to unmarshal WPGGroupShape: %v", err)
	}
}

func TestWPExtentUnmarshalXML(t *testing.T) {
	xmlData := `<wp:extent cx="1000000" cy="2000000"/>`

	var extent WPExtent
	err := xml.Unmarshal(StringToBytes(xmlData), &extent)
	if err != nil {
		t.Fatalf("Failed to unmarshal WPExtent: %v", err)
	}

	if extent.CX != 1000000 {
		t.Errorf("Expected CX=1000000, got %d", extent.CX)
	}
	if extent.CY != 2000000 {
		t.Errorf("Expected CY=2000000, got %d", extent.CY)
	}
}

func TestWTableWidthUnmarshalXML(t *testing.T) {
	xmlData := `<w:tblW w:w="5000" w:type="dxa"/>`

	var width WTableWidth
	err := xml.Unmarshal(StringToBytes(xmlData), &width)
	if err != nil {
		t.Fatalf("Failed to unmarshal WTableWidth: %v", err)
	}

	if width.W != 5000 {
		t.Errorf("Expected W=5000, got %d", width.W)
	}
	if width.Type != "dxa" {
		t.Errorf("Expected Type=dxa, got %s", width.Type)
	}
}

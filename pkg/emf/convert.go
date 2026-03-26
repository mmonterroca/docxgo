package emf

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
)

const emrStretchDIBits = 81

// ConvertToPNG extracts the embedded bitmap from an EMF that contains a
// single EMR_STRETCHDIBITS record and encodes it as PNG.
// Office-generated EMFs are almost always simple bitmap wrappers using this
// record type, not complex vector drawings.
func ConvertToPNG(emfData []byte, w io.Writer) error {
	if len(emfData) < 116 {
		return fmt.Errorf("emf: data too short (%d bytes)", len(emfData))
	}

	headerSize := binary.LittleEndian.Uint32(emfData[4:8])

	recOff := int(headerSize)
	if recOff+80 > len(emfData) {
		return fmt.Errorf("emf: no room for StretchDIBits record")
	}

	recType := binary.LittleEndian.Uint32(emfData[recOff : recOff+4])
	if recType != emrStretchDIBits {
		return fmt.Errorf("emf: expected EMR_STRETCHDIBITS (%d), got %d", emrStretchDIBits, recType)
	}

	offBmiSrc := int(binary.LittleEndian.Uint32(emfData[recOff+40 : recOff+44]))
	offBitsSrc := int(binary.LittleEndian.Uint32(emfData[recOff+48 : recOff+52]))
	cbBitsSrc := int(binary.LittleEndian.Uint32(emfData[recOff+52 : recOff+56]))

	bmiStart := recOff + offBmiSrc
	bitsStart := recOff + offBitsSrc
	if bmiStart+20 > len(emfData) || bitsStart+cbBitsSrc > len(emfData) {
		return fmt.Errorf("emf: bitmap offsets out of range")
	}

	biWidth := int(int32(binary.LittleEndian.Uint32(emfData[bmiStart+4 : bmiStart+8])))
	biHeight := int(int32(binary.LittleEndian.Uint32(emfData[bmiStart+8 : bmiStart+12])))
	biBitCount := binary.LittleEndian.Uint16(emfData[bmiStart+14 : bmiStart+16])
	if biBitCount != 32 {
		return fmt.Errorf("emf: unsupported bit depth %d (only 32-bit supported)", biBitCount)
	}

	bottomUp := biHeight > 0
	if biHeight < 0 {
		biHeight = -biHeight
	}

	stride := biWidth * 4
	if cbBitsSrc < biHeight*stride {
		return fmt.Errorf("emf: pixel buffer too small (%d bytes for %dx%d)", cbBitsSrc, biWidth, biHeight)
	}

	img := image.NewNRGBA(image.Rect(0, 0, biWidth, biHeight))
	pixels := emfData[bitsStart : bitsStart+cbBitsSrc]

	for y := 0; y < biHeight; y++ {
		srcY := y
		if bottomUp {
			srcY = biHeight - 1 - y
		}
		srcRow := pixels[srcY*stride : srcY*stride+stride]
		for x := 0; x < biWidth; x++ {
			off := x * 4
			img.SetNRGBA(x, y, color.NRGBA{
				R: srcRow[off+2],
				G: srcRow[off+1],
				B: srcRow[off+0],
				A: srcRow[off+3],
			})
		}
	}

	return png.Encode(w, img)
}

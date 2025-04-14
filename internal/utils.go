package internal

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"image"
	"strings"

	"golang.org/x/image/draw"
)

// Compressed stream with buffer reuse
func CompressStream(data []byte) []byte {
	var b bytes.Buffer
	zw := zlib.NewWriter(&b)
	_, _ = zw.Write(data)
	_ = zw.Close()
	return b.Bytes()
}

// Generates PDF /XObject resource dictionary
func NewResourceDictionary(imageIDs []int) string {
	var sb strings.Builder
	sb.WriteString("<< /ProcSet [/PDF /ImageC /Text] /Font << /F1 2 0 R >> /XObject <<")
	for _, objID := range imageIDs {
		sb.WriteString(fmt.Sprintf(" /Im%d %d 0 R", objID, objID))
	}
	sb.WriteString(" >> >>")
	return sb.String()
}

func EscapePDFString(s string) string {
	// Escaping parentheses and backslashes
	sBytes := bytes.ReplaceAll([]byte(s), []byte("\\"), []byte("\\\\"))
	sBytes = bytes.ReplaceAll(sBytes, []byte("("), []byte("\\("))
	sBytes = bytes.ReplaceAll(sBytes, []byte(")"), []byte("\\)"))
	return string(sBytes)
}
func ResizeImage(img image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}
func FlattenToWhite(img image.Image) image.Image {
	bounds := img.Bounds()
	whiteBG := image.NewRGBA(bounds)

	// Fill with white
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			whiteBG.Set(x, y, image.White)
		}
	}

	// Draw original image over white
	draw.Draw(whiteBG, bounds, img, bounds.Min, draw.Over)
	return whiteBG
}
